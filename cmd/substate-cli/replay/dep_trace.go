package replay

import (
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/research"
	cli "gopkg.in/urfave/cli.v1"
)

// Some flags
var (
	OutputPath = cli.StringFlag{
		Name:  "output-dir",
		Usage: "Filename",
	}
)

// dependency trace
var DependencyTraceCommand = cli.Command{
	Action:    depTrace,
	Name:      "dependency-trace",
	Usage:     "Collect dependency stats in each txs",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		research.WorkersFlag,
		research.SkipTransferTxsFlag,
		research.SkipCallTxsFlag,
		research.SkipCreateTxsFlag,
		research.SubstateDirFlag,
		OutputPath,
	},
	Description: `
dependency-trace TODO
    `,
}

type DepTraceWorkerResult struct {
	BlockId uint64
	TxId    int
	Result  string
}

func DepTraceWorkerAction(block uint64, tx int, substate *research.Substate) (ret research.WorkerResult, err error) {
	var result DepTraceWorkerResult
	result.BlockId = block
	result.TxId = tx
	result.Result = ""

	inputAlloc := substate.InputAlloc
	inputEnv := substate.Env
	inputMessage := substate.Message

	var (
		vmConfig    vm.Config
		chainConfig *params.ChainConfig
		getTracerFn func(txIndex int, txHash common.Hash) (tracer vm.EVMLogger, err error)
	)

	vmConfig = vm.Config{}

	chainConfig = &params.ChainConfig{}
	*chainConfig = *params.MainnetChainConfig
	// disable DAOForkSupport, otherwise account states will be overwritten
	chainConfig.DAOForkSupport = false

	getTracerFn = func(txIndex int, txHash common.Hash) (tracer vm.EVMLogger, err error) {
		return nil, nil
	}

	var hashError error
	getHash := func(num uint64) common.Hash {
		if inputEnv.BlockHashes == nil {
			hashError = fmt.Errorf("getHash(%d) invoked, no blockhashes provided", num)
			return common.Hash{}
		}
		h, ok := inputEnv.BlockHashes[num]
		if !ok {
			hashError = fmt.Errorf("getHash(%d) invoked, blockhash for that block not provided", num)
		}
		return h
	}

	// Apply Message
	var (
		statedb = MakeOffTheChainStateDB(inputAlloc)
		gaspool = new(core.GasPool)
		txHash  = common.Hash{0x02}
		txIndex = tx
	)

	gaspool.AddGas(inputEnv.GasLimit)
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    inputEnv.Coinbase,
		BlockNumber: new(big.Int).SetUint64(inputEnv.Number),
		Time:        new(big.Int).SetUint64(inputEnv.Timestamp),
		Difficulty:  inputEnv.Difficulty,
		GasLimit:    inputEnv.GasLimit,
		GetHash:     getHash,
	}
	// If currentBaseFee is defined, add it to the vmContext.
	if inputEnv.BaseFee != nil {
		blockCtx.BaseFee = new(big.Int).Set(inputEnv.BaseFee)
	}

	msg := inputMessage.AsMessage()

	tracer, err := getTracerFn(txIndex, txHash)
	if err != nil {
		return result, err
	}
	vmConfig.Tracer = tracer
	vmConfig.Debug = (tracer != nil)
	statedb.Prepare(txHash, txIndex)

	txCtx := vm.TxContext{
		GasPrice: msg.GasPrice(),
		Origin:   msg.From(),
	}

	evm := vm.NewEVM(blockCtx, txCtx, statedb, chainConfig, vmConfig)
	snapshot := statedb.Snapshot()
	_, err = core.ApplyMessage(evm, msg, gaspool)

	if err != nil {
		statedb.RevertToSnapshot(snapshot)
		return result, err
	}

	if hashError != nil {
		return result, hashError
	}

	if chainConfig.IsByzantium(blockCtx.BlockNumber) {
		statedb.Finalise(true)
	} else {
		statedb.IntermediateRoot(chainConfig.IsEIP158(blockCtx.BlockNumber))
	}
	for i, graph := range evm.Graphs {
		if len(graph.Edges) == 0 {
			continue
		}
		graph.Init()
		// fmt.Printf("Number of edges: %d\n", len((*graph).Edges))
		// edges := (*graph).Edges
		// for src, dests := range edges {
		// 	for dest, typ := range dests {
		// 		fmt.Printf("%s   ---   %s (%s)\n", src.GetName(), dest.GetName(), typ.String())
		// 	}
		// }
		maxLevel, frac, required_processors := graph.NaiveParallelisationLevel()
		maxRuntime, speedup := graph.NaiveParallelisationEmprical()
		cpu2 := graph.CPListScheduling(2)
		cpu4 := graph.CPListScheduling(4)
		cpu8 := graph.CPListScheduling(8)
		formatted := fmt.Sprintf("%d,%d,%d,%d,%f,%d,%d,%f,%d,%d,%d\n",
			block, tx, i,
			maxLevel, frac, required_processors,
			maxRuntime, speedup,
			cpu2, cpu4, cpu8)
		result.Result += formatted
	}
	return result, nil
}

func depTrace(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli replay command requires exactly 2 arguments")
	}

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("substate-cli replay: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("substate-cli replay: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("substate-cli replay: error: first block has larger number than last block")
	}

	research.SetSubstateFlags(ctx)
	research.OpenSubstateDBReadOnly()
	defer research.CloseSubstateDB()

	path := ctx.String(OutputPath.Name)
	collectorAction := func(result research.BlockResult, prev *research.CollectorResult) error {
        filename := fmt.Sprintf("%s/%d", path, result.BlockId)
		file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(err)
		}
		for _, txResult := range result.Results {
			file.WriteString(txResult.(DepTraceWorkerResult).Result)
		}
		file.Close()
		return nil
	}

	taskPool := research.NewSubstateTaskPool(
		"substate-cli dependecy trace",
		DepTraceWorkerAction, collectorAction, research.VanillaCollectorInit,
		uint64(first), uint64(last), ctx)
	_, err = taskPool.Execute()
	return err
}
