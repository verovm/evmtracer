package vm

import (
	"container/heap"
	"fmt"
	// "sort"
)

type Pval struct {
	priority int // schedule priority
	deps     int // deps left, only schedule when it is zero
	node     SNode
}

type Scheduler struct {
	numProcessors int
	tasks         PriorityQueue
	runtime       *RuntimeTable
}

func newScheduler(numProcessors int, runtime *RuntimeTable) *Scheduler {
	return &Scheduler{
		numProcessors: numProcessors,
		runtime:       runtime,
	}
}

func (this *Scheduler) numFreeSlot() int {
	return this.numProcessors - this.tasks.Len()
}

func (this *Scheduler) hasTask() bool {
	return this.tasks.Len() > 0
}

// Finish the minimum job, update time, return time spent
// Assuming there is at least one job
func (this *Scheduler) run() (int, SNode) {
	ret := heap.Pop(&this.tasks).(*Item)
	this.tasks.updateAll(ret.priority)
	return ret.priority, ret.ele
}

func (this *Scheduler) addNewJobs(jobs *[]SNode) {
	if i := this.numFreeSlot(); i < len(*jobs) {
		panic("Cannot add that many job")
	}
	for _, job := range *jobs {
		heap.Push(&this.tasks, &Item{job, (*this.runtime)[job.op], 0})
	}
	// sort.Slice(this.tasks, func(i, j int) bool {
	// 	return this.tasks[i].runtime < this.tasks[j].runtime
	// })
}

type TaskPipeline struct {
	priority *map[SNode]Pval
	Pq       PriorityQueue
}

// return next n (or less) jobs, (zero deps and highest n runtime)
func (this *TaskPipeline) nextAvaiableTask(n int) []SNode {
	var (
		nextReadies []SNode
	)
	// for _, pval := range *this.priority {
	// 	if pval.deps <= 0 {
	// 		// nextReadies = append(nextReadies, pval)
	// 		heap.Push(&this.pq, &Item{pval.node, pval.priority, 0})
	// 		// fmt.Printf("add heap %s\n", pval.node.GetName())
	// 	}
	// }

	// sort.Slice(nextReadies, func(i, j int) bool {
	// 	return nextReadies[i].priority > nextReadies[j].priority
	// })

	// if len(nextReadies) > n {
	// 	nextReadies = nextReadies[:n]
	// }
	for i := 0; this.Pq.Len() > 0 && i < n; {
		r := heap.Pop(&this.Pq).(*Item)
		nextReadies = append(nextReadies, r.ele)
		i += 1
	}

	for _, job := range nextReadies {
		delete(*this.priority, job)
	}

	return nextReadies
}

func (this *TaskPipeline) updateDep(graph *DepGraph, node SNode) {
	for fwd, _ := range graph.Edges[node] {
		if next, ok := (*this.priority)[fwd]; !ok {
			panic(fmt.Sprintf("(%s) %s is already finished\n", node.GetName(), fwd.GetName()))
		} else {
			next.deps -= 1
			(*this.priority)[fwd] = next
			if next.deps <= 0 {
				// fmt.Printf("Push %s\n", next.node.GetName())
				heap.Push(&this.Pq, &Item{next.node, next.priority, 0})
			}
		}
	}
}

func (this *TaskPipeline) hasJobs() bool {
	return this.Pq.Len() > 0
}
