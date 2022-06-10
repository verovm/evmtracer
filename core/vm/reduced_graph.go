package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"sort"
	"sync"
)

type RNode struct {
	op   OpCode
	val  uint256.Int
	deps []*RNode
	id   int //  always unique, for sorting deps
	// hash string
}

func (this RNode) toString() string {
	return fmt.Sprintf("%s_%d", this.op.String(), this.id)
}

func (this RNode) hash() string {
	ret := fmt.Sprintf("%s_%s", this.op.String(), this.val.Hex())
	for _, dep := range this.deps {
		ret += fmt.Sprintf("_%d", dep.id)
	}
	return ret
}

type RTuple [2]uint64
type RedundantCountTable [256]RTuple

type ReducedGraph struct {
	// Nodes  []*RNode
	Nodes map[string]*RNode
	// id               int
	RTable           RedundantCountTable
	NumSloads        int
	NumSstores       int
	NumSloadsCached  int
	NumSstoresCached int
	evm              *EVM
}

func (this *ReducedGraph) getNodeId() int {
	defer func() { this.evm.NodeId++ }()
	return this.evm.NodeId
}

func (this *ReducedGraph) Visualization() string {
	var (
		graph   = ""
		visited = make(map[int]bool)
	)

	for _, node := range this.Nodes {
		if _, ok := visited[node.id]; ok {
			continue
		}
		visited[node.id] = true
		for _, dep := range node.deps {
			graph += fmt.Sprintf("%s -> %s\n", dep.toString(), node.toString())
		}
	}
	fmt.Printf(graph)
	return graph
}

func (this *ReducedGraph) recordRedundancy(op OpCode, gas uint64) {
	this.RTable[op][0] += 1
	this.RTable[op][1] += gas
}

func NewReducedGraph(blockNum int64, evm *EVM) *ReducedGraph {
	return &ReducedGraph{
		// id:               0,
		NumSloads:        0,
		NumSstores:       0,
		NumSstoresCached: 0,
		NumSloadsCached:  0,
		Nodes:            make(map[string]*RNode),
		evm:              evm,
	}
}

func (this *ReducedGraph) AddReducedGraph(other ReducedGraph) {
	for i := range this.RTable {
		this.RTable[i][0] += other.RTable[i][0]
		this.RTable[i][1] += other.RTable[i][1]
		// if (i == int(SSTORE)) {
		//     fmt.Printf("This %d  other %d\n", this.RTable[i][0], other.RTable[i][1])
		// }
	}
	this.NumSloads += other.NumSloads
	this.NumSstores += other.NumSstores
	this.NumSloadsCached += other.NumSloadsCached
	this.NumSstoresCached += other.NumSstoresCached
}

// Does not perform check, add a new node with a new id
func (this *ReducedGraph) addNewNode(newNode *RNode) *RNode {
	newNode.id = this.getNodeId()
	this.Nodes[newNode.hash()] = newNode
	return newNode
}

func reusable(op OpCode) bool {
	return op != CREATE && op != CREATE2 &&
		op != CALL && op != CALLCODE &&
		op != DELEGATECALL && op != STATICCALL
}

// Add a node into graph
// This method add new node only if there is no node in the graph that can be reused
func (this *ReducedGraph) tryAddNode(newNode *RNode) (*RNode, bool) {
	if val, ok := this.Nodes[newNode.hash()]; ok {
		return val, true
	}
	return this.addNewNode(newNode), false
}

var ReducedStackPool = sync.Pool{
	New: func() interface{} {
		return &ReducedStack{data: make([]*RNode, 0, 16)}
	},
}

type ReducedStack struct {
	data []*RNode
}

func NewReducedStack() *ReducedStack {
	return ReducedStackPool.Get().(*ReducedStack)
}

func order(deps *[]*RNode) {
	// sort deps by address
	sort.Slice(*deps, func(i, j int) bool {
		return (*deps)[i].id < (*deps)[j].id
	})
}

// Consume N Node from the reducedstack
// construct deps for the RNode
func (st *ReducedStack) consumeN(n int64) []*RNode {
	var deps []*RNode
	for i := int64(0); i < n; i++ {
		// This is always a RAW operation (read something from the stack)
		deps = append(deps, st.pop())
	}
	return deps
}

func (st *ReducedStack) push(d *RNode) {
	// NOTE push limit (1024) is checked in baseCheck
	st.data = append(st.data, d)
}

func (st *ReducedStack) pushN(ds ...*RNode) {
	// FIXME: Is there a way to pass args by pointers.
	st.data = append(st.data, ds...)
}

func (st *ReducedStack) pop() (ret *RNode) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *ReducedStack) len() int {
	return len(st.data)
}

// swap is just like regular swap
func (st *ReducedStack) swap(n int) {
	st.data[st.len()-n], st.data[st.len()-1] = st.data[st.len()-1], st.data[st.len()-n]
}

// Dup is just like regular dup
func (st *ReducedStack) dup(n int) {
	st.push(st.data[st.len()-n])
}

type ReducedMemory struct {
	store       []*RNode
	last_resize *RNode
}

func NewReducedMemory() *ReducedMemory {
	return &ReducedMemory{
		last_resize: nil,
	}
}

// return true if the dest location are the same computatonal result of dest
// False otherwise, meanwhile, change the last modifier on mem
func (m *ReducedMemory) Set(offset, size uint64, dest *RNode) bool {
	if size > 0 {
		if offset+size > uint64(len(m.store)) {
			panic("invalid memory: store empty")
		}
		var reused = true
		for i := offset; i < offset+size; i++ {
			if m.store[i] != dest {
				reused = false
			}
			m.store[i] = dest
		}
		return reused
	}
	return true
}

func (m *ReducedMemory) Set32(offset uint64, dest *RNode) bool {
	if offset+32 > uint64(len(m.store)) {
		panic("invalid memory: store empty")
	}
	var reused = true
	for i := offset; i < offset+32; i++ {
		if m.store[i] != dest {
			reused = false
		}
		m.store[i] = dest
		return reused
	}
	return true
}

// Resize resizes the memory to size
func (m *ReducedMemory) Resize(size uint64, destNode *RNode) {
	i := uint64(m.Len())
	if uint64(m.Len()) < size {
		m.last_resize = destNode
		m.store = append(m.store, make([]*RNode, size-uint64(m.Len()))...)
		for ; i < size; i++ {
			m.store[i] = destNode
		}
	}
}

// GetCopy is a read operation, so the data is part of the computational deps (but stateful)
// This method simply returns the deps
func (m *ReducedMemory) GetCopy(offset, size int64) []*RNode {
	if size == 0 {
		return nil
	}

	if len(m.store) > int(offset) {
		var (
			visited = make(map[*RNode]bool)
			deps    []*RNode
		)
		for _, node := range m.store[offset : offset+size] {
			if _, ok := visited[node]; !ok {
				deps = append(deps, node)
				visited[node] = true
			}
		}
		return deps
	}
	return nil
}

// This method simply returns the deps
func (m *ReducedMemory) GetPtr(offset, size int64) []*RNode {
	if size == 0 {
		return nil
	}

	if len(m.store) > int(offset) {
		var (
			visited = make(map[*RNode]bool)
			deps    []*RNode
		)
		for _, node := range m.store[offset : offset+size] {
			if _, ok := visited[node]; !ok {
				deps = append(deps, node)
				visited[node] = true
			}
		}
		return deps
	}
	return []*RNode{}
}

// Len returns the length of the backing slice
func (m *ReducedMemory) Len() int {
	return len(m.store)
}

// Data returns the backing slice
func (m *ReducedMemory) Data() []*RNode {
	return m.store
}

type ReducedDB struct {
	state map[common.Address]map[common.Hash]*RNode
}

type MemDB struct {
	cache map[common.Address]map[common.Hash]common.Hash
}

func NewMemDB() *MemDB {
	return &MemDB{
		cache: make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func NewReducedDB() *ReducedDB {
	return &ReducedDB{
		state: make(map[common.Address]map[common.Hash]*RNode),
	}
}

func (this *MemDB) GetStateMem(addr common.Address, hash common.Hash, expected common.Hash) bool {
	// fmt.Printf("Get %s %s\n", addr.Hex(), hash.Hex())
	cached := false
	if ret, ok := this.cache[addr]; ok {
		if value, ok := ret[hash]; ok {
			this.cache[addr][hash] = expected
			cached = value == expected
			// if (!cached) {
			//     fmt.Printf("Get cannot cached, expected %s get %s\n", expected.Hex(), value.Hex())
			// }
		} else {
			this.cache[addr][hash] = expected
			cached = false
			// if (!cached) {
			//     fmt.Printf("Get cannot cached, no Entry\n")
			// }
		}
	} else {
		this.cache[addr] = make(map[common.Hash]common.Hash)
		this.cache[addr][hash] = expected
		cached = false
		// if (!cached) {
		//     fmt.Printf("Get cannot cached, no Entry\n")
		// }
	}
	return cached
}

// Read operation collect the deps
func (this *ReducedDB) GetState(addr common.Address, hash common.Hash, graph *ReducedGraph) []*RNode {
	if ret, ok := this.state[addr]; ok {
		if node, ok := ret[hash]; ok {
			// fmt.Printf("sload0 %s %s\n", addr.Hex(), hash.Hex())
			// fmt.Printf("sload0 deps: %s\n", node.hash())
			return []*RNode{node}
		} else {
			// fmt.Printf("init dependency with %d get %s\n", this.id, uint256.NewInt(this.id).Hex())
			this.state[addr][hash] = &RNode{op: NOP, deps: nil, id: graph.getNodeId()}
			// fmt.Printf("sload1 deps: %s\n", this.state[addr][hash].hash())
		}
	} else {
		this.state[addr] = make(map[common.Hash]*RNode)
		this.state[addr][hash] = &RNode{op: NOP, deps: nil, id: graph.getNodeId()}
		// fmt.Printf("sload2 deps: %s\n", this.state[addr][hash].hash())
	}
	// fmt.Printf("sload1 %s %s\n", addr.Hex(), hash.Hex())
	return []*RNode{this.state[addr][hash]}
}

func (this *MemDB) SetStateMem(addr common.Address, hash common.Hash, newValue common.Hash) bool {
	// fmt.Printf("Set %s %s to %s\n", addr.Hex(), hash.Hex(), newValue.Hex())
	if ret, ok := this.cache[addr]; ok {
		if value, ok := ret[hash]; ok {
			this.cache[addr][hash] = newValue
			// if value != newValue {
			// 	fmt.Printf("Cannot cache: Old %s new %s\n", value.Hex(), newValue.Hex())
			// }
			return value == newValue
		} else {
			this.cache[addr][hash] = newValue
			return false
		}
	} else {
		this.cache[addr] = make(map[common.Hash]common.Hash)
		this.cache[addr][hash] = newValue
		return false
	}
}

// Set operation return true if the last modifier is the same as newNode
// update the last modifier
func (this *ReducedDB) SetState(addr common.Address, hash common.Hash, newNode *RNode) bool {
	// fmt.Printf("Set %s %s to %s\n", addr.Hex(), hash.Hex(), newNode.hash())
	var reused = false
	if ret, ok := this.state[addr]; ok {
		if node, ok := ret[hash]; ok {
			if node == newNode {
				reused = true
				// fmt.Printf("old node: %s\n", node.hash())
				// fmt.Printf("new node: %s\n", node.hash())
			}
			// else {
			// fmt.Printf("Set dep: %s\n", newNode.hash())
			// }
			ret[hash] = newNode
		} else {
			this.state[addr][hash] = newNode
		}
	} else {
		this.state[addr] = make(map[common.Hash]*RNode)
		this.state[addr][hash] = newNode
	}
	return reused
}
