package research

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	cli "gopkg.in/urfave/cli.v1"
)

var (
	WorkersFlag = cli.IntFlag{
		Name:  "workers",
		Usage: "Number of worker threads that execute in parallel",
		Value: 4,
	}
	SkipTransferTxsFlag = cli.BoolFlag{
		Name:  "skip-transfer-txs",
		Usage: "Skip executing transactions that only transfer ETH",
	}
	SkipCallTxsFlag = cli.BoolFlag{
		Name:  "skip-call-txs",
		Usage: "Skip executing CALL transactions to accounts with contract bytecode",
	}
	SkipCreateTxsFlag = cli.BoolFlag{
		Name:  "skip-create-txs",
		Usage: "Skip executing CREATE transactions",
	}
)

// Abstract the behvaior of replay
// An abstract replayer should take:
// 1. <start> and <end> block range
// 2. An action on each tx (workers)
// 3. A summary action for each result from the worker (collector)
// 4. As well as an init value for the summary action to start with
// Effectively, foldr(init, summary_func, fmap(action, <start>..<end>))
type WorkloadConfig struct {
	First uint64
	Last  uint64
}

type WorkerResult interface{}

// A BlockResult is simply a collection of results from all txs in a block
type BlockResult struct {
	Results []WorkerResult
	BlockId uint64
}

type CollectorResult interface{}

type WorkerAction func(block uint64, tx int, substate *Substate) (ret WorkerResult, err error)
type CollectorAction func(result BlockResult, prev *CollectorResult) error

// Return an initial value for the collector result
type CollectorInit func() CollectorResult

//
// A set of Vanilla worker and collector actions
//

type VanillaWorkerResult struct {
	BlockId uint64
	TxId    int
}

type VanillaCollectorResult struct{}

func VanillaCollectorInit() CollectorResult {
	return &VanillaCollectorResult{}
}
func VanillaCollectorAction(result BlockResult, prev *CollectorResult) error {
	return nil
}

type SubstateTaskPool struct {
	Name            string
	WorkerAction    WorkerAction
	CollectorAction CollectorAction
	CollectorInit   CollectorInit

	First uint64
	Last  uint64

	Workers         int
	SkipTransferTxs bool
	SkipCallTxs     bool
	SkipCreateTxs   bool

	Ctx *cli.Context // CLI context required to read additional flags

	DB *SubstateDB
}

func NewSubstateTaskPool(name string,
	workerAction WorkerAction,
	collectorAction CollectorAction, collectorInit CollectorInit,
	first, last uint64,
	ctx *cli.Context) *SubstateTaskPool {
	return &SubstateTaskPool{
		Name:            name,
		WorkerAction:    workerAction,
		CollectorAction: collectorAction,
		CollectorInit:   collectorInit,

		First: first,
		Last:  last,

		Workers:         ctx.Int(WorkersFlag.Name),
		SkipTransferTxs: ctx.Bool(SkipTransferTxsFlag.Name),
		SkipCallTxs:     ctx.Bool(SkipCallTxsFlag.Name),
		SkipCreateTxs:   ctx.Bool(SkipCreateTxsFlag.Name),

		Ctx: ctx,

		DB: staticSubstateDB,
	}
}

// ExecuteBlock function iterates on substates of a given block call TaskFunc
func (pool *SubstateTaskPool) ExecuteBlock(block uint64) (results BlockResult, err error) {
	results.BlockId = block
	var res WorkerResult
	for tx, substate := range pool.DB.GetBlockSubstates(block) {
		alloc := substate.InputAlloc
		msg := substate.Message

		to := msg.To
		if pool.SkipTransferTxs && to != nil {
			// skip regular transactions (ETH transfer)
			if account, exist := alloc[*to]; !exist || len(account.Code) == 0 {
				continue
			}
		}
		if pool.SkipCallTxs && to != nil {
			// skip CALL trasnactions with contract bytecode
			if account, exist := alloc[*to]; exist && len(account.Code) > 0 {
				continue
			}
		}
		if pool.SkipCreateTxs && to == nil {
			// skip CREATE transactions
			continue
		}

		res, err = pool.WorkerAction(block, tx, substate)
		if err != nil {
			return results, fmt.Errorf("%s: %v_%v: %v", pool.Name, block, tx, err)
		}
		results.Results = append(results.Results, res)
	}

	return results, nil
}

// Execute function spawns worker goroutines and schedule tasks.
func (pool *SubstateTaskPool) Execute() (res CollectorResult, err error) {
	start := time.Now()

	var totalNumBlock, totalNumTx int64
	defer func() {
		duration := time.Since(start) + 1*time.Nanosecond
		sec := duration.Seconds()

		nb, nt := atomic.LoadInt64(&totalNumBlock), atomic.LoadInt64(&totalNumTx)
		blkPerSec := float64(nb) / sec
		txPerSec := float64(nt) / sec
		fmt.Printf("%s: block range = %v %v\n", pool.Name, pool.First, pool.Last)
		fmt.Printf("%s: total #block = %v\n", pool.Name, nb)
		fmt.Printf("%s: total #tx    = %v\n", pool.Name, nt)
		fmt.Printf("%s: %.2f blk/s, %.2f tx/s\n", pool.Name, blkPerSec, txPerSec)
		fmt.Printf("%s done in %v\n", pool.Name, duration.Round(1*time.Millisecond))
	}()

	// numProcs = numWorker + work producer (1) + main thread (1)
	numProcs := pool.Workers + 2
	if goMaxProcs := runtime.GOMAXPROCS(0); goMaxProcs < numProcs {
		runtime.GOMAXPROCS(numProcs)
	}

	fmt.Printf("%s: block range = %v %v\n", pool.Name, pool.First, pool.Last)
	fmt.Printf("%s: #CPU = %v, #worker = %v\n", pool.Name, runtime.NumCPU(), pool.Workers)

	workChan := make(chan uint64, pool.Workers*10)
	doneChan := make(chan interface{}, pool.Workers*10)
	stopChan := make(chan struct{}, pool.Workers)
	wg := sync.WaitGroup{}
	defer func() {
		// stop all workers
		for i := 0; i < pool.Workers; i++ {
			stopChan <- struct{}{}
		}
		// stop work producer (1)
		stopChan <- struct{}{}

		wg.Wait()
		close(workChan)
		close(doneChan)
	}()
	// dynamically schedule one block per worker
	for i := 0; i < pool.Workers; i++ {
		wg.Add(1)
		// worker goroutine
		go func() {
			defer wg.Done()

			for {
				select {

				case block := <-workChan:
					results, err := pool.ExecuteBlock(block)
					if err != nil {
						doneChan <- err
					} else {
						doneChan <- results
					}

				case <-stopChan:
					return

				}
			}
		}()
	}

	// wait until all workers finish all tasks
	wg.Add(1)
	go func() {
		defer wg.Done()

		for block := pool.First; block <= pool.Last; block++ {
			select {

			case workChan <- block:
				continue

			case <-stopChan:
				return

			}
		}
	}()

	// Count finished blocks in order and report execution speed
	var lastSec float64
	var lastNumBlock, lastNumTx int64
	waitMap := make(map[uint64]struct{})
	collectorResult := pool.CollectorInit()
	for block := pool.First; block <= pool.Last; {

		// Count finshed blocks from waitMap in order
		if _, ok := waitMap[block]; ok {
			delete(waitMap, block)

			block++
			continue
		}

		duration := time.Since(start) + 1*time.Nanosecond
		sec := duration.Seconds()
		if block == pool.Last ||
			(block%10000 == 0 && sec > lastSec+5) ||
			(block%1000 == 0 && sec > lastSec+10) ||
			(block%100 == 0 && sec > lastSec+20) ||
			(block%10 == 0 && sec > lastSec+40) ||
			(sec > lastSec+60) {
			nb, nt := atomic.LoadInt64(&totalNumBlock), atomic.LoadInt64(&totalNumTx)
			blkPerSec := float64(nb-lastNumBlock) / (sec - lastSec)
			txPerSec := float64(nt-lastNumTx) / (sec - lastSec)
			fmt.Printf("%s: elapsed time: %v, number = %v\n", pool.Name, duration.Round(1*time.Millisecond), block)
			fmt.Printf("%s: %.2f blk/s, %.2f tx/s\n", pool.Name, blkPerSec, txPerSec)

			lastSec, lastNumBlock, lastNumTx = sec, nb, nt
		}

		data := <-doneChan
		switch t := data.(type) {

		case BlockResult:
			var err error
			blockResult := BlockResult(data.(BlockResult))
			blockId := blockResult.BlockId
			nt := int64(len(blockResult.Results))
			waitMap[blockId] = struct{}{}
			err = pool.CollectorAction(blockResult, &collectorResult)
			if err != nil {
				panic(err)
			}
			totalNumTx += nt
			totalNumBlock += 1
		case error:
			err := data.(error)
			return res, err

		default:
			panic(fmt.Errorf("%s: unknown type %T value from doneChan", pool.Name, t))

		}
	}

	return res, nil
}
