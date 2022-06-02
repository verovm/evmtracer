package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

type SNode struct {
	op OpCode
	id int64
}

func NewSNode(op OpCode, id int64) SNode {
	return SNode{
		op: op,
		id: id,
	}
}

func (this *SNode) GetName() string {
	return fmt.Sprintf("%s_%d", this.op.String(), this.id)
}

var shadowStackPool = sync.Pool{
	New: func() interface{} {
		return &ShadowStack{data: make([]SNode, 0, 16)}
	},
}

type ShadowStack struct {
	data []SNode
}

func newShadowStack() *ShadowStack {
	return shadowStackPool.Get().(*ShadowStack)
}

func (st *ShadowStack) ConsumeN(n int64, node SNode, graph *DepGraph) {
	for i := int64(0); i < n; i++ {
		// This is always a RAW operation (read something from the stack)
		graph.addEdge(st.pop(), node, RAW)
	}
}

func (st *ShadowStack) Push(d *SNode) {
	// NOTE push limit (1024) is checked in baseCheck
	st.data = append(st.data, *d)
}

func (st *ShadowStack) pushN(ds ...SNode) {
	// FIXME: Is there a way to pass args by pointers.
	st.data = append(st.data, ds...)
}

func (st *ShadowStack) pop() (ret SNode) {
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *ShadowStack) len() int {
	return len(st.data)
}

func (st *ShadowStack) Swap(n int, dest SNode, graph *DepGraph) {
    if (st.len() - n < 0) {
        fmt.Printf("Swap %d, stack size: %d\n", n, len(st.data));
    }
	src1, src2 := st.data[st.len()-1], st.data[st.len()-n]
	graph.addEdge(src1, dest, WAW)
	for node, typ := range graph.Edges[src1] {
		if typ == RAW {
			graph.addEdge(node, dest, WAR)
		}
	}
	graph.addEdge(src2, dest, WAW)
	for node, typ := range graph.Edges[src2] {
		if typ == RAW {
			graph.addEdge(node, dest, WAR)
		}
	}
	st.data[st.len()-1] = dest
	st.data[st.len()-n] = dest
}

func (st *ShadowStack) Dup(n int, dest SNode, graph *DepGraph) {
	src := st.data[st.len()-n]
	graph.addEdge(src, dest, RAW)
	st.Push(&dest)
}

type ShadowMemory struct {
	store       []SNode
	last_resize SNode
}

func NewShadowMemory() *ShadowMemory {
	return &ShadowMemory{
		last_resize: SNode{NOP, -1},
	}
}

func (m *ShadowMemory) SetOffSet(offset uint64, dest SNode) {
    m.store[offset] = dest;
}

func (m *ShadowMemory) Set(offset, size uint64, dest SNode, graph *DepGraph) {
	// It's possible the offset is greater than 0 and size equals 0. This is because
	// the calcMemSize (common.go) could potentially return 0 when size is zero (NO-OP)
	if size > 0 {
		// length of store may never be less than offset + size.
		// The store should be resized PRIOR to setting the memory
		if offset+size > uint64(len(m.store)) {
			panic("invalid memory: store empty")
		}
		//copy(m.store[offset:offset+size], value)
		for i := offset; i < offset+size; i++ {
			if m.store[i].op != NOP {
				graph.addEdge(m.store[i], dest, WAW)
				// Detect any potential WAR
				for node, typ := range graph.Edges[m.store[i]] {
					if typ == RAW {
						graph.addEdge(node, dest, WAR)
					}
				}
			}
			m.store[i] = dest
		}
	}
}

func (m *ShadowMemory) Set32(offset uint64, dest SNode, graph *DepGraph) {
	// length of store may never be less than offset + size.
	// The store should be resized PRIOR to setting the memory
	if offset+32 > uint64(len(m.store)) {
		panic("invalid memory: store empty")
	}
	// Zero the memory area
	// copy(m.store[offset:offset+32], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	// Fill in relevant bits
	// val.WriteToSlice(m.store[offset:])
	for i := offset; i < offset+32; i++ {
		if m.store[i].op != NOP {
			graph.addEdge(m.store[i], dest, WAW)
			// Detect any potential WAR
			for node, typ := range graph.Edges[m.store[i]] {
				if typ == RAW {
					graph.addEdge(node, dest, WAR)
				}
			}
		}
		m.store[i] = dest
	}
}

// Resize resizes the memory to size
func (m *ShadowMemory) Resize(size uint64, destNode SNode) {
	i := uint64(m.Len())
	if uint64(m.Len()) < size {
		m.last_resize = destNode
		m.store = append(m.store, make([]SNode, size-uint64(m.Len()))...)
		for ; i < size; i++ {
			m.store[i] = destNode
		}
	}
}

// Get returns offset + size as a new slice
func (m *ShadowMemory) GetCopy(offset, size int64, dest SNode, graph *DepGraph) {
	if size == 0 {
		return
	}

	if len(m.store) > int(offset) {
		graph.addEdges(m.store[offset:offset+size], dest, RAW)
		return
	}

	return
}

// GetPtr returns the offset + size
func (m *ShadowMemory) GetPtr(offset, size int64, dest SNode, graph *DepGraph) {
	if size == 0 {
		return
	}

	if len(m.store) > int(offset) {
		graph.addEdges(m.store[offset:offset+size], dest, RAW)
		return
	}
}

// Len returns the length of the backing slice
func (m *ShadowMemory) Len() int {
	return len(m.store)
}

// Data returns the backing slice
func (m *ShadowMemory) Data() []SNode {
	return m.store
}

type ShadowDB struct {
	state map[common.Address]map[common.Hash]SNode
}

func NewShadowDB() *ShadowDB {
	return &ShadowDB{
		state: make(map[common.Address]map[common.Hash]SNode),
	}
}

func (this *ShadowDB) GetState(addr common.Address, hash common.Hash, dest SNode, graph *DepGraph) {
	if ret, ok := this.state[addr]; ok {
		if node, ok := ret[hash]; ok {
			graph.addEdge(node, dest, RAW)
			return
		}
	}
	// Reading something from another trace
	graph.addEdge(graph.Root, dest, RAW)
	return
}

func (this *ShadowDB) SetState(addr common.Address, hash common.Hash, dest SNode, graph *DepGraph) {
	if ret, ok := this.state[addr]; ok {
		if node, ok := ret[hash]; ok {
			graph.addEdge(node, dest, WAW)
			for src, typ := range graph.Edges[node] {
				if typ == RAW {
					graph.addEdge(src, dest, WAR)
				}
			}
			return
		} else {
			this.state[addr][hash] = dest
		}
	} else {
		this.state[addr] = make(map[common.Hash]SNode)
		this.state[addr][hash] = dest
		return
	}
}
