# EVM Tracer
To build the program, run `make all`.

To start the program and collect the statistics:
```
./substate-cli redundancy-trace --workers=<N> --skip-transfer-txs --substatedir=<path-to-substate-database> --output-dir=<path-to-output-directory> <block_start> <block_end>
```

In order to use the tracer, you need to first record the substates, see the
following section on how to run the replayer.

In order to minimize the disk usage, raw data is not stored on disk. Instead,
analysis is performed on-the-fly and get dumped in the output folder. 

The main branch implement the redundancy analysis, to use the parallel level
analysis, check out the parallel branch.

# Ethereum Substate Recorder/Replayer
Ethereum substate recorder/replayer based on the paper:

**Yeonsoo Kim, Seongho Jeong, Kamil Jezek, Bernd Burgstaller, and Bernhard Scholz**: _An Off-The-Chain Execution Environment for Scalable Testing and Profiling of Smart Contracts_,  USENIX ATC'21

To build all programs, simply run `make all`.
You can find all executables including `geth` and our `substate-cli` in `build/bin/` directory.

## Record transaction substates
Use `geth import` to save transaction substates in the argument of `--substatedir`
(default: `substate.ethereum`).

There are 5 data structures stored in a substate DB:
1. `SubstateAccount`: account information (nonce, balance, code, storage)
2. `SubstateAlloc`: mapping of account address and `SubstateAccount`
3. `SubstateEnv`: information from block headers (block gasLimit, number, timestamp, hashes)
4. `SubstateMessage`: message for transaction execution
5. `SubstateResult`: result of transaction execution

5 values are required to replay transactions and validate results:
1. `InputAlloc`: alloc that read during transaction execution
2. `Env`: block information required for transaction execution
3. `Message`: array with exactly 1 transaction
4. `OutputAlloc`: alloc that generated by transaction execution
5. `Result`: execution result and receipt array with exactly 1 receipt

The First 2 bytes of a key in a substate DB represent different data types as follows:
1. `1s`: Substate, a key is `"1s"+N+T` with transaction index `T` at block `N`.
`T` and `N` are encoded in a big-endian 64-bit binary.
2. `1c`: EVM bytecode, a key is `"1c"+codeHash` where `codeHash` is Keccak256 hash of the bytecode.

## Replay trasnactions
`substate-cli replay` executes transaction substates in a given block range.
If `substate-cli replay` finds an execution result that is not equivalent to the recorded result,
it returns an error immediately.

For example, if you want to replay transactions from block 1,000,001 to block 2,000,000 in `substate.ethereum`:
```bash
./substate-cli replay 1000001 2000000
```

Here are command line options for `substate-cli replay`:
```
replay [command options] <blockNumFirst> <blockNumLast>

The substate-cli replay command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.

OPTIONS:
   --workers value      Number of worker threads that execute in parallel (default: 4)
   --skip-transfer-txs  Skip executing transactions that only transfer ETH
   --skip-call-txs      Skip executing CALL transactions to accounts with contract bytecode
   --skip-create-txs    Skip executing CREATE transactions
   --substatedir value  Data directory for substate recorder/replayer (default: "substate.ethereum")
```

For example, if you want 32 workers to replay transactions except CREATE transactions:
```bash
./substate-cli replay 1000001 2000000 --workers 32
```

If you want to replay only CALL transactions and skip the other types of transactions:
```bash
./substate-cli replay 1000001 2000000 --skip-transfer-txs --skip-create-txs
```

If you want to use a substate DB other than `substate.ethereum` (e.g. `/path/to/substate_db`):
```bash
./substate-cli replay 1000001 2000000 --substatedir /path/to/substate_db
```

### Hard-fork assessment
To assess hard-forks with prior transactions, use `substate-cli replay-fork` command. Run `./substate-cli replay-fork --help` for more details:

```
replay-fork [command options] <blockNumFirst> <blockNumLast>

The replay-fork command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.

--hard-fork parameter is recommended for this command.

OPTIONS:
   --workers value      Number of worker threads that execute in parallel (default: 4)
   --skip-transfer-txs  Skip executing transactions that only transfer ETH
   --skip-call-txs      Skip executing CALL transactions to accounts with contract bytecode
   --skip-create-txs    Skip executing CREATE transactions
   --hard-fork value    Hard-fork block number, won't change block number in Env for NUMBER instruction
                          1: Frontier
                          1150000: Homestead
                          2463000: Tangerine Whistle
                          2675000: Spurious Dragon
                          4370000: Byzantium
                          7280000: Constantinople + Petersburg
                          9069000: Istanbul
                          12244000: Berlin
                          12965000: London (default: 12965000)
   --substatedir value  Data directory for substate recorder/replayer (default: "substate.ethereum")
```

## Substate DB manipulation
`substate-cli db` is an additional command to directly manipulate substate DBs.

### `upgrade`
`substate-cli db upgrade` command convert the old DB layout (`stage1-substate`) used for the USENIX ATC'21 paper to the latest DB layout (`substate.ethereum`).
```
./substate-cli db upgrade stage1-substate substate.ethereum
```

### `clone`
`substate-cli db clone` command reads substates of a given block range and copies them in a substate DB clone.
```
./substate-cli db clone srcdb dstdb 46147 50000
```

### `compact`
`substate-cli db compact` command compacts any LevelDB instance including the substate DB.
```
./substate-cli db compact substate.ethereum
```
