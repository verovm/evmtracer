// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
    // "fmt"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/core/state"
)

func opAdd(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.Add(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)
	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opSub(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.Sub(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opMul(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.Mul(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opDiv(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.Div(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opSdiv(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.SDiv(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opMod(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.Mod(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opSmod(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.SMod(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opExp(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	base, exponent := scope.Stack.pop(), scope.Stack.peek()
	exponent.Exp(&base, exponent)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	back, num := scope.Stack.pop(), scope.Stack.peek()
	num.ExtendSign(num, &back)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opNot(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x := scope.Stack.peek()
	x.Not(x)

	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(1)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opLt(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	if x.Lt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opGt(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	if x.Gt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opSlt(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	if x.Slt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opSgt(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	if x.Sgt(y) {
		y.SetOne()
	} else {
		y.Clear()
	}

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opEq(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	if x.Eq(y) {
		y.SetOne()
	} else {
		y.Clear()
	}

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opIszero(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x := scope.Stack.peek()
	if x.IsZero() {
		x.SetOne()
	} else {
		x.Clear()
	}

	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(1)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opAnd(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.And(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opOr(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.Or(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opXor(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y := scope.Stack.pop(), scope.Stack.peek()
	y.Xor(&x, y)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opByte(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	th, val := scope.Stack.pop(), scope.Stack.peek()
	val.Byte(&th)

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opAddmod(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y, z := scope.Stack.pop(), scope.Stack.pop(), scope.Stack.peek()
	if z.IsZero() {
		z.Clear()
	} else {
		z.AddMod(&x, &y, z)
	}

	// scope.sstack.ConsumeN(3, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(3)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opMulmod(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x, y, z := scope.Stack.pop(), scope.Stack.pop(), scope.Stack.peek()
	z.MulMod(&x, &y, z)

	// scope.sstack.ConsumeN(3, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(3)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

// opSHL implements Shift Left
// The SHL instruction (shift left) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the left by arg1 number of bits.
func opSHL(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := scope.Stack.pop(), scope.Stack.peek()
	if shift.LtUint64(256) {
		value.Lsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

// opSHR implements Logical Shift Right
// The SHR instruction (logical shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with zero fill.
func opSHR(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := scope.Stack.pop(), scope.Stack.peek()
	if shift.LtUint64(256) {
		value.Rsh(value, uint(shift.Uint64()))
	} else {
		value.Clear()
	}

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

// opSAR implements Arithmetic Shift Right
// The SAR instruction (arithmetic shift right) pops 2 values from the stack, first arg1 and then arg2,
// and pushes on the stack arg2 shifted to the right by arg1 number of bits with sign extension.
func opSAR(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)

	shift, value := scope.Stack.pop(), scope.Stack.peek()
	if shift.GtUint64(256) {
		if value.Sign() >= 0 {
			value.Clear()
		} else {
			// Max negative shift: all bits set
			value.SetAllOne()
		}
		return nil, nil
	}
	n := uint(shift.Uint64())
	value.SRsh(value, n)
	return nil, nil
}

func opKeccak256(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	offset, size := scope.Stack.pop(), scope.Stack.peek()
	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)

	data := scope.Memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	// scope.smemory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()), scope.destSNode, scope.graph)

	// scope.sstack.Push(&scope.destSNode)
	dataDep := scope.rdstack.consumeN(2)
	dataDep2 := scope.rmemory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	scope.destRNode.deps = append(dataDep, dataDep2...)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	if interpreter.hasher == nil {
		interpreter.hasher = sha3.NewLegacyKeccak256().(keccakState)
	} else {
		interpreter.hasher.Reset()
	}
	interpreter.hasher.Write(data)
	interpreter.hasher.Read(interpreter.hasherBuf[:])

	evm := interpreter.evm
	if evm.Config.EnablePreimageRecording {
		evm.StateDB.AddPreimage(interpreter.hasherBuf, data)
	}

	size.SetBytes(interpreter.hasherBuf[:])
	return nil, nil
}
func opAddress(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetBytes(scope.Contract.Address().Bytes()))
	// scope.sstack.Push(&scope.destSNode)

	// Address depends on the contract address
	scope.destRNode.val = *scope.Stack.peek()
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opBalance(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	slot := scope.Stack.peek()
	address := common.Address(slot.Bytes20())
	slot.SetFromBig(interpreter.evm.StateDB.GetBalance(address))

	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	// Balance is based on the actual balance value
	scope.destRNode.deps = scope.rdstack.consumeN(1)
	scope.destRNode.val = *scope.Stack.peek()
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetBytes(interpreter.evm.Origin.Bytes()))

	// scope.sstack.Push(&scope.destSNode)
	// Origin is constant within a transaction
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}
func opCaller(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetBytes(scope.Contract.Caller().Bytes()))

	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.val = *scope.Stack.peek()
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opCallValue(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	v, _ := uint256.FromBig(scope.Contract.value)
	scope.Stack.push(v)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.val = *scope.Stack.peek()
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opCallDataLoad(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	x := scope.Stack.peek()
	x_copy := x
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	if offset, overflow := x.Uint64WithOverflow(); !overflow {
		data := getData(scope.Contract.Input, offset, 32)
		x.SetBytes(data)
	} else {
		x.Clear()
	}
	// scope.sstack.Push(&scope.destSNode)

	// CallDataload is based on x
	scope.destRNode.deps = scope.rdstack.consumeN(1)
	scope.destRNode.val = *x_copy
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetUint64(uint64(len(scope.Contract.Input))))

	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.val = *scope.Stack.peek()
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opCallDataCopy(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		memOffset  = scope.Stack.pop()
		dataOffset = scope.Stack.pop()
		length     = scope.Stack.pop()
	)
	// scope.sstack.ConsumeN(3, scope.destSNode, scope.graph)
	dataOffset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		dataOffset64 = 0xffffffffffffffff
	}
	// These values are checked for overflow during gas cost calculation
	memOffset64 := memOffset.Uint64()
	length64 := length.Uint64()
	scope.Memory.Set(memOffset64, length64, getData(scope.Contract.Input, dataOffset64, length64))
	scope.mmemory.Set(memOffset64, length64, getData(scope.Contract.Input, dataOffset64, length64))
	// scope.smemory.Set(memOffset64, length64, scope.destSNode, scope.graph)

	scope.destRNode.deps = scope.rdstack.consumeN(3)
	order(&scope.destRNode.deps)
	rnode, _ := scope.rgraph.tryAddNode(scope.destRNode)
	reused := scope.rmemory.Set(memOffset64, length64, rnode)
	if reused {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetUint64(uint64(len(interpreter.returnData))))

	// scope.sstack.Push(&scope.destSNode)
	scope.destRNode.val = *(new(uint256.Int).SetUint64(uint64(len(interpreter.returnData))))
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		memOffset  = scope.Stack.pop()
		dataOffset = scope.Stack.pop()
		length     = scope.Stack.pop()
	)
	// scope.sstack.ConsumeN(3, scope.destSNode, scope.graph)
	scope.destRNode.deps = scope.rdstack.consumeN(3)
	order(&scope.destRNode.deps)

	offset64, overflow := dataOffset.Uint64WithOverflow()
	if overflow {
		return nil, ErrReturnDataOutOfBounds
	}
	// we can reuse dataOffset now (aliasing it for clarity)
	var end = dataOffset
	end.Add(&dataOffset, &length)
	end64, overflow := end.Uint64WithOverflow()
	if overflow || uint64(len(interpreter.returnData)) < end64 {
		return nil, ErrReturnDataOutOfBounds
	}
	scope.Memory.Set(memOffset.Uint64(), length.Uint64(), interpreter.returnData[offset64:end64])
	scope.mmemory.Set(memOffset.Uint64(), length.Uint64(), interpreter.returnData[offset64:end64])
	// scope.smemory.Set(memOffset.Uint64(), length.Uint64(), scope.destSNode, scope.graph)
	rnode, _ := scope.rgraph.tryAddNode(scope.destRNode)
	reused := scope.rmemory.Set(memOffset.Uint64(), length.Uint64(), rnode)
	if reused {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	return nil, nil
}

func opExtCodeSize(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	slot := scope.Stack.peek()
	slot.SetUint64(uint64(interpreter.evm.StateDB.GetCodeSize(slot.Bytes20())))
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(1)
	scope.destRNode.val = *slot
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opCodeSize(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	l := new(uint256.Int)
	l.SetUint64(uint64(len(scope.Contract.Code)))
	scope.Stack.push(l)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.val = *scope.Stack.peek()
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opCodeCopy(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		memOffset  = scope.Stack.pop()
		codeOffset = scope.Stack.pop()
		length     = scope.Stack.pop()
	)
	// scope.sstack.ConsumeN(3, scope.destSNode, scope.graph)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	codeCopy := getData(scope.Contract.Code, uint64CodeOffset, length.Uint64())
	scope.Memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)
	scope.mmemory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)
	// scope.smemory.Set(memOffset.Uint64(), length.Uint64(), scope.destSNode, scope.graph)

	scope.destRNode.deps = scope.rdstack.consumeN(3)
	order(&scope.destRNode.deps)
	rnode, _ := scope.rgraph.tryAddNode(scope.destRNode)
	reused := scope.rmemory.Set(memOffset.Uint64(), length.Uint64(), rnode)
	if reused {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}

	return nil, nil
}

func opExtCodeCopy(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		stack      = scope.Stack
		a          = stack.pop()
		memOffset  = stack.pop()
		codeOffset = stack.pop()
		length     = stack.pop()
	)
	// scope.sstack.ConsumeN(4, scope.destSNode, scope.graph)
	uint64CodeOffset, overflow := codeOffset.Uint64WithOverflow()
	if overflow {
		uint64CodeOffset = 0xffffffffffffffff
	}
	addr := common.Address(a.Bytes20())
	codeCopy := getData(interpreter.evm.StateDB.GetCode(addr), uint64CodeOffset, length.Uint64())
	scope.Memory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)
	scope.mmemory.Set(memOffset.Uint64(), length.Uint64(), codeCopy)
	// scope.smemory.Set(memOffset.Uint64(), length.Uint64(), scope.destSNode, scope.graph)

	scope.destRNode.deps = scope.rdstack.consumeN(4)
	order(&scope.destRNode.deps)
	rnode, _ := scope.rgraph.tryAddNode(scope.destRNode)
	reused := scope.rmemory.Set(memOffset.Uint64(), length.Uint64(), rnode)
	if reused {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	return nil, nil
}

// opExtCodeHash returns the code hash of a specified account.
// There are several cases when the function is called, while we can relay everything
// to `state.GetCodeHash` function to ensure the correctness.
//   (1) Caller tries to get the code hash of a normal contract account, state
// should return the relative code hash and set it as the result.
//
//   (2) Caller tries to get the code hash of a non-existent account, state should
// return common.Hash{} and zero will be set as the result.
//
//   (3) Caller tries to get the code hash for an account without contract code,
// state should return emptyCodeHash(0xc5d246...) as the result.
//
//   (4) Caller tries to get the code hash of a precompiled account, the result
// should be zero or emptyCodeHash.
//
// It is worth noting that in order to avoid unnecessary create and clean,
// all precompile accounts on mainnet have been transferred 1 wei, so the return
// here should be emptyCodeHash.
// If the precompile account is not transferred any amount on a private or
// customized chain, the return value will be zero.
//
//   (5) Caller tries to get the code hash for an account which is marked as suicided
// in the current transaction, the code hash of this account should be returned.
//
//   (6) Caller tries to get the code hash for an account which is marked as deleted,
// this account should be regarded as a non-existent account and zero should be returned.
func opExtCodeHash(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	slot := scope.Stack.peek()
	address := common.Address(slot.Bytes20())
	if interpreter.evm.StateDB.Empty(address) {
		slot.Clear()
	} else {
		slot.SetBytes(interpreter.evm.StateDB.GetCodeHash(address).Bytes())
	}
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(1)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opGasprice(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.GasPrice)
	scope.Stack.push(v)

	// Gasprice is a constant within transaction
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	// scope.sstack.Push(&scope.destSNode)
	return nil, nil
}

func opBlockhash(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	num := scope.Stack.peek()
	num64, overflow := num.Uint64WithOverflow()

	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)
	scope.destRNode.deps = scope.rdstack.consumeN(1)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	// record-replay: convert vm.StateDB to state.StateDB and save block hash
	defer func() {
		statedb, ok := interpreter.evm.StateDB.(*state.StateDB)
		if ok {
			statedb.ResearchBlockHashes[num64] = common.BytesToHash(num.Bytes())
		}
	}()

	if overflow {
		num.Clear()
		return nil, nil
	}
	var upper, lower uint64
	upper = interpreter.evm.Context.BlockNumber.Uint64()
	if upper < 257 {
		lower = 0
	} else {
		lower = upper - 256
	}
	if num64 >= lower && num64 < upper {
		num.SetBytes(interpreter.evm.Context.GetHash(num64).Bytes())
	} else {
		num.Clear()
	}
	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetBytes(interpreter.evm.Context.Coinbase.Bytes()))
	// scope.sstack.Push(&scope.destSNode)

	// constant within a block
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opTimestamp(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Context.Time)
	scope.Stack.push(v)
	// scope.sstack.Push(&scope.destSNode)

	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opNumber(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Context.BlockNumber)
	scope.Stack.push(v)
	// scope.sstack.Push(&scope.destSNode)

	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	v, _ := uint256.FromBig(interpreter.evm.Context.Difficulty)
	scope.Stack.push(v)
	// scope.sstack.Push(&scope.destSNode)

	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opGasLimit(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetUint64(interpreter.evm.Context.GasLimit))
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.val = *(new(uint256.Int).SetUint64(interpreter.evm.Context.GasLimit))
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opPop(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.pop()
	scope.rdstack.pop()
	if scope.Stack.len() == 0 {
		// If after pop the stack is empty, this is a redudant pop
		// Hence we only trace the dependency before the pop, but not after
		// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	} else {
		// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
		// scope.sstack.Push(&scope.destSNode)
	}
	return nil, nil
}

func opMload(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
    scope.rgraph.NumMloads += 1
	v := scope.Stack.peek()
	offset := int64(v.Uint64())
    expect := scope.Memory.GetPtr(offset, 32)
	v.SetBytes(expect)

	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)
	// scope.smemory.GetPtr(offset, 32, scope.destSNode, scope.graph)

	dataDep := scope.rdstack.consumeN(1)
	stateDep := scope.rmemory.GetPtr(offset, 32)
	scope.destRNode.deps = append(dataDep, stateDep...)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
    cached := scope.mmemory.GetPtr(offset, 32, expect)
    if cached {
        scope.rgraph.NumMloadsCached += 1
    }
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
    if r && !cached {
        panic("Mload reused not cached")
        // fmt.Println("Mload reused not cached")
    }
	scope.rdstack.push(rnode)
	return nil, nil
}

func opMstore(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
    scope.rgraph.NumMstores += 1
	// pop value of the stack
	mStart, val := scope.Stack.pop(), scope.Stack.pop()
	scope.Memory.Set32(mStart.Uint64(), &val)
    cached := scope.mmemory.Set32(mStart.Uint64(), &val)
    if cached {
        scope.rgraph.NumMstoresCached += 1
    }

	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.smemory.Set32(mStart.Uint64(), scope.destSNode, scope.graph)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, _ := scope.rgraph.tryAddNode(scope.destRNode)
	reused := scope.rmemory.Set32(mStart.Uint64(), rnode)
	if reused {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
    if reused && !cached {
        panic("Mstore reused but not cached")
        // fmt.Println("Mstore reused but not cached")
    }
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
    scope.rgraph.NumMstore8s += 1
	off, val := scope.Stack.pop(), scope.Stack.pop()
	scope.Memory.store[off.Uint64()] = byte(val.Uint64())
    cached := false
    if scope.mmemory.store[off.Uint64()] == byte(val.Uint64()) {
        cached = true
    }
    scope.mmemory.store[off.Uint64()] = byte(val.Uint64())
    if (cached) {
        scope.rgraph.NumMstore8sCached += 1
    }


	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.smemory.SetOffSet(off.Uint64(), scope.destSNode)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
	order(&scope.destRNode.deps)
	rnode, _ := scope.rgraph.tryAddNode(scope.destRNode)
	if rnode == scope.rmemory.store[off.Uint64()] {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
        if !cached {
            panic("Mstore8 reused but not cached")
            // fmt.Println("Mstore8 reused but not cached")
        }
	}
	scope.rmemory.store[off.Uint64()] = rnode
	return nil, nil
}

func opSload(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.rgraph.NumSloads += 1
	loc := scope.Stack.peek()
	hash := common.Hash(loc.Bytes32())
	val := interpreter.evm.StateDB.GetState(scope.Contract.Address(), hash)
	loc.SetBytes(val.Bytes())
	cached := interpreter.evm.MemDB.GetStateMem(scope.Contract.Address(), hash, val)
	if cached {
		scope.rgraph.NumSloadsCached += 1
	}

	// scope.sdb.GetState(scope.Contract.Address(), hash, scope.destSNode, scope.graph)
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	// scope.sstack.Push(&scope.destSNode)

	deps := interpreter.evm.ReducedDB.GetState(scope.Contract.Address(), hash, scope.rgraph)
	scope.destRNode.deps = append(deps, scope.rdstack.consumeN(1)...)
	order(&scope.destRNode.deps)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
    if r && !cached {
        panic("Fatal Sload reused but not cached!\n")
    }
	scope.rdstack.push(rnode)
	return nil, nil
}

func opSstore(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	scope.rgraph.NumSstores += 1
	loc := scope.Stack.pop()
	val := scope.Stack.pop()
	interpreter.evm.StateDB.SetState(scope.Contract.Address(),
		loc.Bytes32(), val.Bytes32())
	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	// scope.sdb.SetState(scope.Contract.Address(),
	// common.Hash(loc.Bytes32()), scope.destSNode, scope.graph)

	scope.destRNode.deps = scope.rdstack.consumeN(2)
    order(&scope.destRNode.deps)
	rnode, _ := scope.rgraph.tryAddNode(scope.destRNode)
	interpreter.evm.StateDB.SetState(scope.Contract.Address(),
		common.Hash(loc.Bytes32()), common.Hash(val.Bytes32()))
	// fmt.Printf("SSTORE %s (%d)\n", common.Hash(loc.Bytes32()).String(), interpreter.evm.depth)
	// scope.sdb.SetState(scope.contract.Address(),
	// common.Hash(loc.Bytes32()), scope.destSNode, scope.graph)
	reused := interpreter.evm.ReducedDB.SetState(scope.Contract.Address(),
		common.Hash(loc.Bytes32()), rnode)
	cached := interpreter.evm.MemDB.SetStateMem(scope.Contract.Address(),
		common.Hash(loc.Bytes32()), common.Hash(val.Bytes32()))
	if cached {
		scope.rgraph.NumSstoresCached += 1
	}
	if reused {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
    if reused && !cached {
        panic("Fatal Sstore reused but not cached!\n")
    }
	return nil, nil
}

func opJump(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if atomic.LoadInt32(&interpreter.evm.abort) != 0 {
		return nil, errStopToken
	}
	pos := scope.Stack.pop()
	scope.rdstack.consumeN(1)
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	if !scope.Contract.validJumpdest(&pos) {
		return nil, ErrInvalidJump
	}
	*pc = pos.Uint64() - 1 // pc will be increased by the interpreter loop

	return nil, nil
}

func opJumpi(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if atomic.LoadInt32(&interpreter.evm.abort) != 0 {
		return nil, errStopToken
	}
	pos, cond := scope.Stack.pop(), scope.Stack.pop()
	scope.rdstack.consumeN(2)
	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
	if !cond.IsZero() {
		if !scope.Contract.validJumpdest(&pos) {
			return nil, ErrInvalidJump
		}
		*pc = pos.Uint64() - 1 // pc will be increased by the interpreter loop
	}
	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	return nil, nil
}

func opPc(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetUint64(*pc))
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.val = *(new(uint256.Int).SetUint64(*pc))
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opMsize(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetUint64(uint64(scope.Memory.Len())))
	// scope.sstack.Push(&scope.destSNode)
	// Msize depdends on the last resize
	// if !(scope.smemory.last_resize == SNode{NOP, -1}) {
	// scope.graph.addEdge(scope.smemory.last_resize, scope.destSNode, RAW)
	// }

	// Can be reused as long as the size is the same
	scope.destRNode.val = *(new(uint256.Int).SetUint64(uint64(scope.Memory.Len())))
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

func opGas(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	scope.Stack.push(new(uint256.Int).SetUint64(scope.Contract.Gas))
	// scope.sstack.Push(&scope.destSNode)

	scope.destRNode.val = *(new(uint256.Int).SetUint64(scope.Contract.Gas))
	rnode := scope.rgraph.addNewNode(scope.destRNode)
	scope.rdstack.push(rnode)
	return nil, nil
}

func opCreate(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	var (
		value        = scope.Stack.pop()
		offset, size = scope.Stack.pop(), scope.Stack.pop()
		input        = scope.Memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = scope.Contract.Gas
	)
	// scope.sstack.ConsumeN(3, scope.destSNode, scope.graph)
	// scope.smemory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()), scope.destSNode, scope.graph)
	// Create is not reusable
	scope.destRNode.deps = scope.rdstack.consumeN(3)
	order(&scope.destRNode.deps)

	if interpreter.evm.chainRules.IsEIP150 {
		gas -= gas / 64
	}
	// reuse size int for stackvalue
	stackvalue := size

	scope.Contract.UseGas(gas)
	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		bigVal = value.ToBig()
	}

	res, addr, returnGas, suberr := interpreter.evm.Create(scope.Contract, input, gas, bigVal)
	// Push item on the stack based on the returned error. If the ruleset is
	// homestead we must check for CodeStoreOutOfGasError (homestead only
	// rule) and treat as an error, if the ruleset is frontier we must
	// ignore this error and pretend the operation was successful.
	if interpreter.evm.chainRules.IsHomestead && suberr == ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else if suberr != nil && suberr != ErrCodeStoreOutOfGas {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	scope.Stack.push(&stackvalue)
	// scope.sstack.Push(&scope.destSNode)
	scope.Contract.Gas += returnGas
	rnode := scope.rgraph.addNewNode(scope.destRNode)
	scope.rdstack.push(rnode)

	if suberr == ErrExecutionReverted {
		interpreter.returnData = res // set REVERT data to return data buffer
		return res, nil
	}
	interpreter.returnData = nil // clear dirty return data buffer
	return nil, nil
}

func opCreate2(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	var (
		endowment    = scope.Stack.pop()
		offset, size = scope.Stack.pop(), scope.Stack.pop()
		salt         = scope.Stack.pop()
		input        = scope.Memory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()))
		gas          = scope.Contract.Gas
	)
	// scope.sstack.ConsumeN(4, scope.destSNode, scope.graph)
	// scope.smemory.GetCopy(int64(offset.Uint64()), int64(size.Uint64()), scope.destSNode, scope.graph)

	// Create is not reusable
	scope.destRNode.deps = scope.rdstack.consumeN(4)
	order(&scope.destRNode.deps)

	// Apply EIP150
	gas -= gas / 64
	scope.Contract.UseGas(gas)
	// reuse size int for stackvalue
	stackvalue := size
	//TODO: use uint256.Int instead of converting with toBig()
	bigEndowment := big0
	if !endowment.IsZero() {
		bigEndowment = endowment.ToBig()
	}
	res, addr, returnGas, suberr := interpreter.evm.Create2(scope.Contract, input, gas,
		bigEndowment, &salt)
	// Push item on the stack based on the returned error.
	if suberr != nil {
		stackvalue.Clear()
	} else {
		stackvalue.SetBytes(addr.Bytes())
	}
	scope.Stack.push(&stackvalue)
	// scope.sstack.Push(&scope.destSNode)
	scope.Contract.Gas += returnGas
	rnode := scope.rgraph.addNewNode(scope.destRNode)
	scope.rdstack.push(rnode)

	if suberr == ErrExecutionReverted {
		interpreter.returnData = res // set REVERT data to return data buffer
		return res, nil
	}
	interpreter.returnData = nil // clear dirty return data buffer
	return nil, nil
}

func opCall(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	stack := scope.Stack
	// Pop gas. The actual gas in interpreter.evm.callGasTemp.
	// We can use this as a temporary value
	temp := stack.pop()
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(1)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	// scope.sstack.ConsumeN(6, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(6)
	toAddr := common.Address(addr.Bytes20())
	// Get the arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	// scope.smemory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()), scope.destSNode, scope.graph)
	scope.rmemory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	if interpreter.readOnly && !value.IsZero() {
		return nil, ErrWriteProtection
	}
	var bigVal = big0
	//TODO: use uint256.Int instead of converting with toBig()
	// By using big0 here, we save an alloc for the most common case (non-ether-transferring contract calls),
	// but it would make more sense to extend the usage of uint256.Int
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.Call(scope.Contract, toAddr, args, gas, bigVal)

	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	// scope.sstack.Push(&scope.destSNode)
	rnode := scope.rgraph.addNewNode(scope.destRNode)
	scope.rdstack.push(rnode)
	if err == nil || err == ErrExecutionReverted {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		// scope.smemory.Set(retOffset.Uint64(), retSize.Uint64(), scope.destSNode, scope.graph)
		scope.rmemory.Set(retOffset.Uint64(), retSize.Uint64(), rnode)
		scope.mmemory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
	}
	scope.Contract.Gas += returnGas

	interpreter.returnData = ret
	return ret, nil
}

func opCallCode(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := scope.Stack
	// We use it as a temporary value
	temp := stack.pop()
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(1)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	// scope.sstack.ConsumeN(6, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(6)
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	// scope.smemory.GetCopy(int64(inOffset.Uint64()), int64(inSize.Uint64()), scope.destSNode, scope.graph)
	scope.rmemory.GetCopy(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	//TODO: use uint256.Int instead of converting with toBig()
	var bigVal = big0
	if !value.IsZero() {
		gas += params.CallStipend
		bigVal = value.ToBig()
	}

	ret, returnGas, err := interpreter.evm.CallCode(scope.Contract, toAddr, args, gas, bigVal)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	// scope.sstack.Push(&scope.destSNode)
	rnode := scope.rgraph.addNewNode(scope.destRNode)
	scope.rdstack.push(rnode)
	if err == nil || err == ErrExecutionReverted {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		// scope.smemory.Set(retOffset.Uint64(), retSize.Uint64(), scope.destSNode, scope.graph)
		scope.rmemory.Set(retOffset.Uint64(), retSize.Uint64(), rnode)
	}
	scope.Contract.Gas += returnGas

	interpreter.returnData = ret
	return ret, nil
}

func opDelegateCall(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	stack := scope.Stack
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	// We use it as a temporary value
	temp := stack.pop()
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(1)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	// scope.sstack.ConsumeN(5, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(5)
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	// scope.smemory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()), scope.destSNode, scope.graph)
	scope.rmemory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.DelegateCall(scope.Contract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	// scope.sstack.Push(&scope.destSNode)
	rnode := scope.rgraph.addNewNode(scope.destRNode)
	scope.rdstack.push(rnode)
	if err == nil || err == ErrExecutionReverted {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		scope.mmemory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		// scope.smemory.Set(retOffset.Uint64(), retSize.Uint64(), scope.destSNode, scope.graph)
		scope.rmemory.Set(retOffset.Uint64(), retSize.Uint64(), rnode)
	}
	scope.Contract.Gas += returnGas

	interpreter.returnData = ret
	return ret, nil
}

func opStaticCall(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.
	stack := scope.Stack
	// We use it as a temporary value
	temp := stack.pop()
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(1)
	gas := interpreter.evm.callGasTemp
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	// scope.sstack.ConsumeN(5, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(5)
	toAddr := common.Address(addr.Bytes20())
	// Get arguments from the memory.
	args := scope.Memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))
	// scope.smemory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()), scope.destSNode, scope.graph)
	scope.rmemory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))

	ret, returnGas, err := interpreter.evm.StaticCall(scope.Contract, toAddr, args, gas)
	if err != nil {
		temp.Clear()
	} else {
		temp.SetOne()
	}
	stack.push(&temp)
	// scope.sstack.Push(&scope.destSNode)
	rnode := scope.rgraph.addNewNode(scope.destRNode)
	scope.rdstack.push(rnode)
	if err == nil || err == ErrExecutionReverted {
		ret = common.CopyBytes(ret)
		scope.Memory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		scope.mmemory.Set(retOffset.Uint64(), retSize.Uint64(), ret)
		// scope.smemory.Set(retOffset.Uint64(), retSize.Uint64(), scope.destSNode, scope.graph)
		scope.rmemory.Set(retOffset.Uint64(), retSize.Uint64(), rnode)
	}
	scope.Contract.Gas += returnGas

	interpreter.returnData = ret
	return ret, nil
}

func opReturn(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	offset, size := scope.Stack.pop(), scope.Stack.pop()
	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)

	ret := scope.Memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	// scope.smemory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()), scope.destSNode, scope.graph)

	// return does not produce result
	scope.rdstack.consumeN(2)
	return ret, errStopToken
}

func opRevert(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	offset, size := scope.Stack.pop(), scope.Stack.pop()
	ret := scope.Memory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()))
	// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)

	// revert does not produce result
	scope.rdstack.consumeN(2)
	interpreter.returnData = ret
	// scope.smemory.GetPtr(int64(offset.Uint64()), int64(size.Uint64()), scope.destSNode, scope.graph)

	return ret, ErrExecutionReverted
}

func opUndefined(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	return nil, &ErrInvalidOpCode{opcode: OpCode(scope.Contract.Code[*pc])}
}

func opStop(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	return nil, errStopToken
}

func opSelfdestruct(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	if interpreter.readOnly {
		return nil, ErrWriteProtection
	}
	beneficiary := scope.Stack.pop()
	// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
	scope.rdstack.consumeN(1)

	balance := interpreter.evm.StateDB.GetBalance(scope.Contract.Address())
	interpreter.evm.StateDB.AddBalance(beneficiary.Bytes20(), balance)
	interpreter.evm.StateDB.Suicide(scope.Contract.Address())
	if interpreter.cfg.Debug {
		interpreter.cfg.Tracer.CaptureEnter(SELFDESTRUCT, scope.Contract.Address(), beneficiary.Bytes20(), []byte{}, 0, balance)
		interpreter.cfg.Tracer.CaptureExit([]byte{}, 0, nil)
	}
	return nil, errStopToken
}

// following functions are used by the instruction jump  table

// make log instruction function
func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		if interpreter.readOnly {
			return nil, ErrWriteProtection
		}
		topics := make([]common.Hash, size)
		stack := scope.Stack
		mStart, mSize := stack.pop(), stack.pop()
		// scope.sstack.ConsumeN(2, scope.destSNode, scope.graph)
		deps := scope.rdstack.consumeN(2)
		for i := 0; i < size; i++ {
			addr := stack.pop()
			// scope.sstack.ConsumeN(1, scope.destSNode, scope.graph)
			deps = append(deps, scope.rdstack.consumeN(1)...)
			topics[i] = addr.Bytes32()
		}

		d := scope.Memory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()))
		// scope.smemory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()), scope.destSNode, scope.graph)
		deps2 := scope.rmemory.GetCopy(int64(mStart.Uint64()), int64(mSize.Uint64()))
		scope.destRNode.deps = append(deps, deps2...)
		order(&scope.destRNode.deps)
		scope.rgraph.tryAddNode(scope.destRNode)

		interpreter.evm.StateDB.AddLog(&types.Log{
			Address: scope.Contract.Address(),
			Topics:  topics,
			Data:    d,
			// This is a non-consensus field, but assigned here because
			// core/state doesn't know the current block number.
			BlockNumber: interpreter.evm.Context.BlockNumber.Uint64(),
		})

		return nil, nil
	}
}

// opPush1 is a specialized version of pushN
func opPush1(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	var (
		codeLen = uint64(len(scope.Contract.Code))
		integer = new(uint256.Int)
	)
	*pc += 1
	if *pc < codeLen {
		scope.Stack.push(integer.SetUint64(uint64(scope.Contract.Code[*pc])))
		scope.destRNode.val = *(integer.SetUint64(uint64(scope.Contract.Code[*pc])))
	} else {
		scope.Stack.push(integer.Clear())
		scope.destRNode.val = *(integer.Clear())
	}
	// scope.sstack.Push(&scope.destSNode)
	rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
	if r {
		scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
	}
	scope.rdstack.push(rnode)
	return nil, nil
}

// make push instruction function
func makePush(size uint64, pushByteSize int) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		codeLen := len(scope.Contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}

		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}

		integer := new(uint256.Int)
		scope.Stack.push(integer.SetBytes(common.RightPadBytes(
			scope.Contract.Code[startMin:endMin], pushByteSize)))
		// scope.sstack.Push(&scope.destSNode)
		scope.destRNode.val = *(integer.SetBytes(
			common.RightPadBytes(scope.Contract.Code[startMin:endMin], pushByteSize)))
		rnode, r := scope.rgraph.tryAddNode(scope.destRNode)
		if r {
			scope.rgraph.recordRedundancy(scope.destRNode.op, scope.rgasCost)
		}
		scope.rdstack.push(rnode)
		*pc += size
		return nil, nil
	}
}

// make dup instruction function
func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		scope.Stack.dup(int(size))
		// scope.sstack.Dup(int(size), scope.destSNode, scope.graph)
		scope.rdstack.dup(int(size))
		return nil, nil
	}
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
		scope.Stack.swap(int(size))
		// scope.sstack.Swap(int(size), scope.destSNode, scope.graph)
		scope.rdstack.swap(int(size))
		return nil, nil
	}
}
