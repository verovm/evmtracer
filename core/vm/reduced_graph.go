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

var do_debug = false

func Debug(s string) {
	if do_debug {
		fmt.Print(s)
	}
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
	RTable            RedundantCountTable
	NumSloads         int
	NumSstores        int
	NumSloadsCached   int
	NumSstoresCached  int
	NumMloads         int
	NumMstores        int
	NumMstore8s       int
	NumMloadsCached   int
	NumMstoresCached  int
	NumMstore8sCached int
	evm               *EVM
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
		NumSloads:         0,
		NumSstores:        0,
		NumSstoresCached:  0,
		NumSloadsCached:   0,
		NumMloads:         0,
		NumMstores:        0,
		NumMstore8s:       0,
		NumMloadsCached:   0,
		NumMstoresCached:  0,
		NumMstore8sCached: 0,
		Nodes:             make(map[string]*RNode),
		evm:               evm,
	}
}

func (this *ReducedGraph) AddReducedGraph(other ReducedGraph) {
	for i := range this.RTable {
		this.RTable[i][0] += other.RTable[i][0]
		this.RTable[i][1] += other.RTable[i][1]
	}
	this.NumSloads += other.NumSloads
	this.NumSstores += other.NumSstores
	this.NumSloadsCached += other.NumSloadsCached
	this.NumSstoresCached += other.NumSstoresCached
	this.NumMloads += other.NumMloads
	this.NumMstores += other.NumMstores
	this.NumMloadsCached += other.NumMloadsCached
	this.NumMstoresCached += other.NumMstoresCached
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

func NewMemMemory() *MemMemory {
	return &MemMemory{}
}

func NewReducedMemory() *ReducedMemory {
	return &ReducedMemory{
		last_resize: nil,
	}
}

type MemMemory struct {
	store []byte
}

func (m *MemMemory) Set(offset, size uint64, expect []byte) bool {
	if size <= 0 {
		return false
	}
	Debug(fmt.Sprintf("M: Set %d to %s\n", offset, string(expect)))
	var i uint64
	var length uint64
	if size > uint64(len(expect)) {
		length = uint64(len(expect))
	} else {
		length = size
	}
	for i = 0; i < length; i++ {
		if m.store[offset+i] != expect[i] {
			copy(m.store[offset:offset+size], expect)
			return false
		}
	}
	return true
}

// return true if the dest location are the same computatonal result of dest
// False otherwise, meanwhile, change the last modifier on mem
func (m *ReducedMemory) Set(offset, size uint64, dest *RNode) bool {
	if size > 0 {
		if offset+size > uint64(len(m.store)) {
			panic("invalid memory: store empty")
		}
		var reused = true
		Debug(fmt.Sprintf("R: Set %d to %s\n", offset, dest.hash()))
		for i := offset; i < offset+size; i++ {
			if m.store[i] != dest {
				reused = false
			}
			m.store[i] = dest
		}
		return reused
	}
	return false
}

func (m *MemMemory) Set32(offset uint64, val *uint256.Int) bool {
	Debug(fmt.Sprintf("M: Set32 %d to %s\n", offset, val.Hex()))
	var tmp = make([]byte, 32)
	copy(tmp, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	val.WriteToSlice(tmp[0:32])

	var i uint64
	for i = 0; i < 32; i++ {
		if m.store[offset+i] != tmp[i] {
			Debug(fmt.Sprintf("M: Set32 %d byte cannot reused\n", i))
			copy(m.store[offset:offset+32], tmp)
			return false
		}
	}
	return true
}

func (m *ReducedMemory) Set32(offset uint64, dest *RNode) bool {
	Debug(fmt.Sprintf("R: Set32 %d to %s\n", offset, dest.hash()))
	if offset+32 > uint64(len(m.store)) {
		panic("invalid memory: store empty")
	}
	var reused = true
	for i := offset; i < offset+32; i++ {
		if m.store[i] != dest {
			reused = false
		}
		m.store[i] = dest
	}
	return reused
}

func (m *MemMemory) Len() int {
	return len(m.store)
}

func (m *MemMemory) Resize(size uint64) {
	Debug(fmt.Sprintf("M: Resize to %d\n", size))
	if uint64(m.Len()) < size {
		m.store = append(m.store, make([]byte, size-uint64(m.Len()))...)
	}
}

// Resize resizes the memory to size
func (m *ReducedMemory) Resize(size uint64, destNode *RNode, graph *ReducedGraph) {
	Debug(fmt.Sprintf("R: Resize to %d\n", size))
	i := uint64(m.Len())
	if uint64(m.Len()) < size {
		m.last_resize = destNode

		m.store = append(m.store, make([]*RNode, size-uint64(m.Len()))...)
		for ; i < size; i++ {
			m.store[i] = &RNode{op: NOP, deps: nil, id: graph.getNodeId()}
		}
	}
}

func (m *MemMemory) GetCopy(offset, size int64, expect []byte) bool {
	if size == 0 {
		return false
	}

	if len(m.store) > int(offset) {
		var i int64
		for i = 0; i < size; i++ {
			if m.store[offset+i] != expect[i] {
				return false
			}
		}
		return false
	}

	return true
}

// GetCopy is a read operation, so the data is part of the computational deps (but stateful)
// This method simply returns the deps
func (m *ReducedMemory) GetCopy(offset, size int64) []*RNode {
	if size == 0 {
		return nil
	}

	if len(m.store) > int(offset) {
		var (
			// visited = make(map[*RNode]bool)
			deps []*RNode
		)
		for _, node := range m.store[offset : offset+size] {
			// if _, ok := visited[node]; !ok {
			deps = append(deps, node)
			// visited[node] = true
			// }
		}
		return deps
	}
	return nil
}

func (m *MemMemory) GetPtr(offset, size int64, expect []byte) bool {
	Debug(fmt.Sprintf("GetPtr %d, Old (%s), expect (%s)\n", offset, string(size), string(expect)))
	if size == 0 {
		return false
	}

	if len(m.store) > int(offset) {
		var i int64
		for i = 0; i < size; i++ {
			if m.store[offset+i] != expect[i] {
				copy(m.store[offset:offset+size], expect)
				return false
			}
		}
	}

	return true
}

// This method simply returns the deps
func (m *ReducedMemory) GetPtr(offset, size int64) []*RNode {
	if size == 0 {
		return nil
	}

	if len(m.store) > int(offset) {
		var (
			deps []*RNode
		)
		for _, node := range m.store[offset : offset+size] {
			deps = append(deps, node)
		}
		Debug(fmt.Sprintf("%d Dep: %s\n", offset, deps[0].hash()))
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
	cache map[common.Address]map[common.Hash]bool
}

func NewMemDB() *MemDB {
	return &MemDB{
		cache: make(map[common.Address]map[common.Hash]bool),
	}
}

func NewReducedDB() *ReducedDB {
	return &ReducedDB{
		state: make(map[common.Address]map[common.Hash]*RNode),
	}
}

func (this *MemDB) GetStateMem(addr common.Address, hash common.Hash) bool {
	fmt.Printf("MemGet %s_%s\n", addr.Hex(), hash.Hex())
	if ret, ok := this.cache[addr]; ok {
		if value, ok := ret[hash]; ok {
			this.cache[addr][hash] = true
            if value {
	            fmt.Printf("Cached\n")
            } else {
	            fmt.Printf("Not Cached\n")
            }
			return value
		} else {
			this.cache[addr][hash] = true
	        fmt.Printf("Not Cached\n")
			return false
		}
	} else {
		this.cache[addr] = make(map[common.Hash]bool)
		this.cache[addr][hash] = true
	    fmt.Printf("Not Cached\n")
		return false
	}
}

// Read operation collect the deps
func (this *ReducedDB) GetState(addr common.Address, hash common.Hash, graph *ReducedGraph) []*RNode {
	fmt.Printf("Get %s_%s\n", addr.Hex(), hash.Hex())
	if ret, ok := this.state[addr]; ok {
		if node, ok := ret[hash]; ok {
			return []*RNode{node}
		} else {
			this.state[addr][hash] = &RNode{op: NOP, deps: nil, id: graph.getNodeId()}
		}
	} else {
		this.state[addr] = make(map[common.Hash]*RNode)
		this.state[addr][hash] = &RNode{op: NOP, deps: nil, id: graph.getNodeId()}
	}
	return []*RNode{this.state[addr][hash]}
}

func (this *MemDB) SetStateMem(addr common.Address, hash common.Hash) {
	fmt.Printf("Mem Overwrite %s_%s\n", addr.Hex(), hash.Hex())
	if _, ok := this.cache[addr]; ok {
			this.cache[addr][hash] = false
	} else {
		this.cache[addr] = make(map[common.Hash]bool)
		this.cache[addr][hash] = false
	}
}

// Set operation return true if the last modifier is the same as newNode
// update the last modifier
func (this *ReducedDB) SetState(addr common.Address, hash common.Hash, newNode *RNode) bool {
	var reused = false
	if ret, ok := this.state[addr]; ok {
		if node, ok := ret[hash]; ok {
			if node == newNode {
				reused = true
			}
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
