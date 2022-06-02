package vm

import (
	"container/heap"
	// "math"
	"fmt"
)

type DepType int

const (
	RAW DepType = iota
	WAR
	WAW
	ART  // artifical edge
	CTRL // between basic block
)

var depTypeToString = map[DepType]string{
	RAW:  "RAW",
	WAR:  "WAR",
	WAW:  "WAW",
	ART:  "ART",
	CTRL: "CTRL",
}

func (op DepType) String() string {
	str := depTypeToString[op]
	return str
}

// Dependency graph
//
type DepGraph struct {
	Edges            map[SNode]map[SNode]DepType // dependency edge
	Back_edges       map[SNode]map[SNode]DepType // back edge for analysis
	Nodes            []SNode                     // Nodes in a topological order
	Root             SNode
	Dest             SNode
	EstimatedRuntime *RuntimeTable
	Total            int
}

func NewDepGraph(blockNum int64) *DepGraph {
	return &DepGraph{
		Edges:            make(map[SNode]map[SNode]DepType),
		Back_edges:       make(map[SNode]map[SNode]DepType),
		Root:             SNode{ROOT, -1},
		Dest:             SNode{DEST, -1},
		EstimatedRuntime: getRuntimeTable(blockNum),
		Total:            0,
	}
}

func (this *DepGraph) addNode(node SNode) {
	if _, ok := this.Edges[node]; !ok {
		this.Edges[node] = make(map[SNode]DepType)
		this.Back_edges[node] = make(map[SNode]DepType)
	}
}

// Add edge, dest depends on src
func (this *DepGraph) addEdge(src SNode, dest SNode, dep DepType) {
	if src == dest {
		return
	}
	this.addNode(src)
	this.addNode(dest)

	this.Edges[src][dest] = dep
	this.Back_edges[dest][src] = dep
}

func (this *DepGraph) addEdges(srcs []SNode, dest SNode, dep DepType) {
	for _, src := range srcs {
		this.addEdge(src, dest, dep)
	}
}

func (this *DepGraph) recordTopologicalOrder(node SNode) {
	this.Nodes = append(this.Nodes, node)
}

func (this *DepGraph) getRuntime(node SNode) int {
	return (*this.EstimatedRuntime)[node.op]
}

func (this *DepGraph) Dump() string {
	ret := ""
	for src, dests := range this.Edges {
		for dest, typ := range dests {
			ret += fmt.Sprintf("%s -> %s (%s)\n", src.GetName(), dest.GetName(), typ.String())
		}
	}
	return ret
}

// This should be called once after dep info has been collected
// Create
//  link between jump operations (constraint parallelsim within basic block)
//  artifical link from root to all root nodes,
//  artifical link from all leaves to a single dest node
//  count node
func (this *DepGraph) Init() {
    if (len(this.Edges) <= 0) {
        panic("Client should not init zero length graph")
    }
	
	var (
		basicBlocks [][]SNode // each basic block ends with a jump
		block       []SNode
	)

	for _, node := range this.Nodes {
		block = append(block, node)
		if node.op == JUMP || node.op == JUMPI {
			basicBlocks = append(basicBlocks, block)
			block = nil
		}
	}
	if len(block) != 0 {
		basicBlocks = append(basicBlocks, block)
	}

	for i, block := range basicBlocks {
		if i < len(basicBlocks)-1 {
			nextBlock := basicBlocks[i+1]
			jmp := block[len(block)-1]
			if jmp.op != JUMP && jmp.op != JUMPI {
				panic("This is not right")
			}
			for _, n := range nextBlock {
				if _, ok := this.Edges[jmp][n]; !ok {
					this.addEdge(jmp, n, CTRL)
				}
			}
		}
	}

	var (
		roots []SNode
		dests []SNode
	)
	for node := range this.Edges {
		if len(this.Back_edges[node]) == 0 {
			roots = append(roots, node)
		}
	}
	for _, node := range roots {
		this.addEdge(this.Root, node, ART)
	}

	for node := range this.Edges {
		if len(this.Edges[node]) == 0 {
			dests = append(dests, node)
		}
	}
	for _, node := range dests {
		this.addEdge(node, this.Dest, ART)
	}
	this.Nodes = append([]SNode{this.Root}, this.Nodes...)
	this.Nodes = append(this.Nodes, this.Dest)

	for _, node := range this.Nodes {
		this.Total += this.getRuntime(node)
	}
}

// --- =====  Analysis  ==== ---

// a) Infinite processors
// b) Unified runtime
func (this *DepGraph) NaiveParallelisationLevel() (maxLevel int, frac float64, required_processors int) {
	lvls := make(map[SNode]int) // record node's lvl
	width := make(map[int]int)  // record number of nodes in a lvl
	for _, node := range this.Nodes {
		max := int(0)
		for bk, _ := range this.Back_edges[node] {
			if val, ok := lvls[bk]; !ok {
				for _, n := range this.Nodes {
					fmt.Printf("%s\n", n.GetName())
				}
				fmt.Printf("%s\n", this.Dump())
				panic(fmt.Sprintf("Not a topological order: %s is not set before %s", bk.GetName(), node.GetName()))
			} else if val > max {
				max = val
			}
		}
		lvl := max + 1
		lvls[node] = lvl
		if _, ok := width[lvl]; !ok {
			width[lvl] = 0
		} else {
			width[lvl] += 1
		}
	}
	required_processors = 0
	for _, w := range width {
		if w > required_processors {
			required_processors = w
		}
	}
	return lvls[this.Dest], float64(lvls[this.Dest]) / float64(len(this.Nodes)), required_processors
}

// a) Infinite processors
// b) Emprical runtime
func (this *DepGraph) NaiveParallelisationEmprical() (maxRuntime int, speedup float64) {
	runtime := make(map[SNode]int)
	total := 0 // record runtime of all nodes (with no parallelsim)
	for _, node := range this.Nodes {
		max := 0
		for bk, _ := range this.Back_edges[node] {
			if val, ok := runtime[bk]; !ok {
				panic("Not a topological order")
			} else if val > max {
				max = val
			}
		}
		runtime[node] = max + this.getRuntime(node)
		total += this.getRuntime(node)
	}
	return runtime[this.Dest], float64(runtime[this.Dest]) / float64(total)
}

// estimiated runtime with X-processors using CP list scheduling
func (this *DepGraph) CPListScheduling(numProcessors int) int {
	var (
		priority = make(map[SNode]Pval)
	)
	priority[this.Dest] = Pval{0, 0, this.Dest}
	// initialize priority
	for i := len(this.Nodes) - 1; i >= 0; i-- {
		var (
			node = this.Nodes[i]
			max  = 0
		)
		for fwd, _ := range this.Edges[node] {
			if p, ok := priority[fwd]; !ok {
				// if fwd.op == NOP || fwd.op == ROOT || fwd.op == DEST {
				// 	continue
				// }
				panic("Not a topological order")
			} else if p.priority > max { // TODO: am I sure?
				max = p.priority
			}
		}
		priority[node] = Pval{
			priority: max + this.getRuntime(node),
			deps:     len(this.Back_edges[node]),
			node:     node,
		}
	}

	var (
		pipeline  = TaskPipeline{priority: &priority}
		scheduler = newScheduler(numProcessors, this.EstimatedRuntime)
		total     = 0
	)
	// start simulating
	heap.Push(&pipeline.Pq, &Item{this.Root, priority[this.Root].priority, 0})
	for pipeline.hasJobs() || scheduler.hasTask() {
		// add new job as long as there are job avaiable and idle processor
		if i := scheduler.numFreeSlot(); i > 0 {
			if nextJobs := pipeline.nextAvaiableTask(i); nextJobs != nil {
				scheduler.addNewJobs(&nextJobs)
			}
		}
		// run until next job finish
		if scheduler.hasTask() {
			time, node := scheduler.run()
			total += time
			pipeline.updateDep(this, node)
		} else if pipeline.hasJobs() {
			for key, value := range priority {
				fmt.Printf("%s %d %d\n", key.GetName(), value.deps, value.priority)
			}
			fmt.Printf("----\n")
			fmt.Printf("%s", this.Dump())
			panic("CPU idle but no job was added")
		}
		// fmt.Printf("Remaining Node %d\n", len(priority))
	}
	return total
}

func (this *DepGraph) ReportEdgeDistribution() (raw int, war int, waw int) {
	raw = 0
	war = 0
	waw = 0
	for _, edges := range this.Edges {
		for _, typ := range edges {
			if typ == RAW {
				raw += 1
			} else if typ == WAR {
				war += 1
			} else if typ == WAW {
				waw += 1
			}
		}
	}
	return
}
