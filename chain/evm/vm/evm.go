/*
 * Copyright (c) 2021.  BAEC.ORG.CN All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package vm

import (
	"errors"
	"fmt"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/vm/environment"
	"web3_gui/chain/evm/vm/instructions"
	"web3_gui/chain/evm/vm/memory"
	"web3_gui/chain/evm/vm/opcodes"
	"web3_gui/chain/evm/vm/precompiledContracts"
	"web3_gui/chain/evm/vm/stack"
	"web3_gui/chain/evm/vm/storage"
	"web3_gui/chain/evm/vm/utils"
)

const (
	ContractSenderOrgIdParam  = "__sender_org_id__"
	ContractSenderRoleParam   = "__sender_role__"
	ContractSenderPkParam     = "__sender_pk__"
	ContractCreatorOrgIdParam = "__creator_org_id__"
	ContractCreatorRoleParam  = "__creator_role__"
	ContractCreatorPkParam    = "__creator_pk__"
)

type EVMResultCallback func(result ExecuteResult, err error)
type EVMParam struct {
	MaxStackDepth  int
	ExternalStore  storage.IExternalStorage
	ResultCallback EVMResultCallback
	Context        *environment.Context
}

type EVM struct {
	stack        *stack.Stack
	memory       *memory.Memory
	storage      *storage.Storage //持久化存储
	context      *environment.Context
	instructions instructions.IInstructions //解释器
	resultNotify EVMResultCallback
}

type ExecuteResult struct {
	ResultData   []byte
	GasLeft      uint64
	StorageCache storage.ResultCache
	ExitOpCode   opcodes.OpCode
	ByteCodeHead []byte
	ByteCodeBody []byte
}

func init() {
	Load()
}
func Load() {
	instructions.Load()
}

func New(param EVMParam) *EVM {
	//先屏蔽对于区块gasLimit的判断
	//if param.Context.Block.GasLimit.Cmp(param.Context.Transaction.GasLimit.Int) < 0 {
	//	param.Context.Transaction.GasLimit = evmutils.FromBigInt(param.Context.Block.GasLimit.Int)
	//}

	evm := &EVM{
		stack:        stack.New(param.MaxStackDepth),
		memory:       memory.New(),
		storage:      storage.New(param.ExternalStore),
		context:      param.Context,
		instructions: nil,
		resultNotify: param.ResultCallback,
	}

	evm.instructions = instructions.New(evm, evm.stack, evm.memory, evm.storage, evm.context, nil, closure)

	return evm
}

// default Evm结果回调函数
// getCurrentBlockVersion是长安链用来兼容版本的，
func (e *EVM) subResult(result ExecuteResult, err error) {
	if err == nil && result.ExitOpCode != opcodes.REVERT {
		if e.storage.GetCurrentBlockVersion() < 2211 {
			storage.MergeResultCache(&result.StorageCache, &e.storage.ResultCache)
		} else {
			storage.MergeResultCache2211(&result.StorageCache, &e.storage.ResultCache)
		}
	}
}

func (e *EVM) executePreCompiled(addr uint64, input []byte) (ExecuteResult, error) {
	contract := precompiledContracts.Contracts[addr]
	switch addr {
	case 10:
		input = []byte(e.context.Parameters[ContractSenderOrgIdParam])
		//contract.SetValue(e.context.Parameters[protocol.ContractSenderOrgIdParam])
	case 11:
		input = []byte(e.context.Parameters[ContractSenderRoleParam])
		//contract.SetValue(e.context.Parameters[protocol.ContractSenderRoleParam])
	case 12:
		input = []byte(e.context.Parameters[ContractSenderPkParam])
		//contract.SetValue(e.context.Parameters[protocol.ContractSenderPkParam])
	case 13:
		input = []byte(e.context.Parameters[ContractCreatorOrgIdParam])
		//contract.SetValue(e.context.Parameters[protocol.ContractCreatorOrgIdParam])
	case 14:
		input = []byte(e.context.Parameters[ContractCreatorRoleParam])
		//contract.SetValue(e.context.Parameters[protocol.ContractCreatorRoleParam])
	case 15:
		input = []byte(e.context.Parameters[ContractCreatorPkParam])
		//contract.SetValue(e.context.Parameters[protocol.ContractCreatorPkParam])
	default:
		if addr < 1 || addr > 15 {
			return ExecuteResult{}, errors.New("not existed precompiled contract")
		}
	}
	gasCost := contract.GasCost(input)
	gasLeft := e.instructions.GetGasLeft()

	result := ExecuteResult{
		ResultData:   nil,
		GasLeft:      gasLeft,
		StorageCache: e.storage.ResultCache,
	}

	if gasLeft < gasCost {
		return result, utils.ErrOutOfGas
	}

	execRet, err := contract.Execute(input)
	gasLeft -= gasCost
	e.instructions.SetGasLimit(gasLeft)
	result.ResultData = execRet
	return result, err
}

func (e *EVM) ExecuteContract(isCreate bool, opCode opcodes.OpCode) (ExecuteResult, error) {
	//fmt.Println("--------------------- contract address", e.context.Contract.Address.B58String())
	//合约地址用来判断是否是内置地址
	//contractAddr := e.context.Contract.Address
	//来源evm.Context.Transaction的gasLimit
	gasLeft := e.instructions.GetGasLeft()

	result := ExecuteResult{
		ResultData:   nil,
		GasLeft:      gasLeft,
		StorageCache: e.storage.ResultCache,
	}
	if opCode != opcodes.DELEGATECALL && opCode != opcodes.STATICCALL {
		if e.context.Message.Value != nil && !e.context.Message.Value.IsZero() {
			if !e.storage.CanTransfer(e.context.Message.Caller, e.context.Contract.Address, e.context.Message.Value) {
				return result, errors.New("余额不足")
			}
			//这里执行转账
			e.storage.Transfer(e.context.Message.Caller, e.context.Contract.Address, e.context.Message.Value)
			if s, _ := e.storage.GetCodeSize(e.context.Message.Caller); s.Sign() > 0 {
				e.storage.InternalTx(e.context.Message.Caller, e.context.Transaction.TxHash, e.context.Message.Caller, e.context.Contract.Address, e.context.Message.Value)
			}
		}
	}

	//return e.executePreCompiled(1, e.context.Message.Data)
	//判断是否是内置合约
	//if contractAddr != nil {
	//	if contractAddr.IsUint64() {
	//		addr := contractAddr.Uint64()
	//		if addr < precompiledContracts.ContractsMaxAddress {
	//			return e.executePreCompiled(addr, e.context.Message.Data)
	//		}
	//	}
	//}

	//合约给普通账户转账也会走此逻辑
	if len(e.context.Contract.Code) == 0 {
		if e.resultNotify != nil {
			e.resultNotify(result, nil)
		}
		return result, nil
	}

	//解释器执行
	execRet, gasLeft, byteCodeHead, byteCodeBody, err := e.instructions.ExecuteContract(isCreate)
	result.ResultData = execRet
	result.GasLeft = gasLeft
	result.ExitOpCode = e.instructions.ExitOpCode()
	result.ByteCodeBody = byteCodeBody
	result.ByteCodeHead = byteCodeHead

	if e.resultNotify != nil {
		e.resultNotify(result, err)
	}
	return result, err
}

func (e *EVM) GetPcCountAndTimeUsed() (uint64, int64) {
	return e.instructions.GetPcCountAndTimeUsed()
}

func (e *EVM) getClosureDefaultEVM(param instructions.ClosureParam) *EVM {
	newEVM := New(EVMParam{
		MaxStackDepth:  1024,
		ExternalStore:  e.storage.ExternalStorage,
		ResultCallback: e.subResult,
		Context: &environment.Context{
			Block:       e.context.Block,
			Transaction: e.context.Transaction,
			Message: environment.Message{
				Data: param.CallData,
			},
			Parameters: e.context.Parameters,
		},
	})
	newEVM.instructions.SetGasLimit(param.GasRemaining.Uint64())
	newEVM.context.Contract = environment.Contract{
		Address: param.ContractAddress,
		Code:    param.ContractCode,
		Hash:    param.ContractHash,
	}

	return newEVM
}

/*
*
合约中的transfer等操作通过CALL指令实现
*/
func (e *EVM) commonCall(param instructions.ClosureParam) ([]byte, error) {
	newEVM := e.getClosureDefaultEVM(param)
	//set storage address and call value
	switch param.OpCode {
	case opcodes.CALLCODE:
		newEVM.context.Contract.Address = e.context.Contract.Address
		newEVM.context.Message.Value = param.CallValue
		newEVM.context.Message.Caller = e.context.Contract.Address

	case opcodes.DELEGATECALL:
		newEVM.context.Contract.Address = e.context.Contract.Address
		newEVM.context.Message.Value = e.context.Message.Value
		//newEVM.context.Message.Value = evmutils.New(0)
		newEVM.context.Message.Caller = e.context.Message.Caller
		//newEVM.context.Message.Caller = e.context.Contract.Address
	case opcodes.CALL:
		newEVM.context.Contract.Address = param.ContractAddress
		newEVM.context.Message.Value = param.CallValue
		newEVM.context.Message.Caller = e.context.Contract.Address
		//if e.storage.GetCurrentBlockVersion() <= 2212 {
		//	newEVM.context.Message.Caller = e.context.Message.Caller
		//} else {
		//	newEVM.context.Message.Caller = e.context.Contract.Address
		//}

	}
	//if param.OpCode == opcodes.STATICCALL || e.instructions.IsReadOnly() {
	//修复DELEGATECALL调用
	//STATICCALL -> DELEGATECALL 时，会因e.instructions.IsReadOnly()修改调用上下文
	if param.OpCode == opcodes.STATICCALL {
		newEVM.context.Contract.Address = param.ContractAddress
		//newEVM.context.Message.Value = e.context.Message.Value
		newEVM.context.Message.Value = evmutils.New(0)
		newEVM.context.Message.Caller = e.context.Contract.Address
		//if e.storage.GetCurrentBlockVersion() > 2212 {
		//	newEVM.context.Contract.Address = param.ContractAddress
		//	newEVM.context.Message.Value = e.context.Message.Value
		//	newEVM.context.Message.Caller = e.context.Contract.Address
		//}
		newEVM.instructions.SetReadOnly()
	}

	//engine.Log.Error("commonCall", newEVM.context.Message.Caller.B58String())

	ret, err := newEVM.ExecuteContract(false, param.OpCode)
	if err != nil {
		return nil, err
	}
	if ret.ExitOpCode == opcodes.REVERT {
		msg, err := abi.UnpackRevert(ret.ResultData)
		if err != nil {
			return ret.ResultData, err
		}
		return ret.ResultData, fmt.Errorf("%s", msg)
	}
	//ret, err := newEVM.ExecuteContract(opcodes.CALL == param.OpCode)

	e.instructions.SetGasLimit(ret.GasLeft)
	return ret.ResultData, err
}

func (e *EVM) commonCreate(param instructions.ClosureParam) ([]byte, error) {
	//var addr *utils.Int
	//if opcodes.CREATE == param.OpCode {
	//	addr = e.storage.ExternalStorage.CreateAddress(e.context.Message.Caller, e.context.Transaction)
	//} else {
	//	addr = e.storage.ExternalStorage.CreateFixedAddress(e.context.Message.Caller, param.CreateSalt, e.context.Transaction)
	//}

	newEVM := e.getClosureDefaultEVM(param)

	//newEVM.context.Contract.Address = addr
	newEVM.context.Message.Value = param.CallValue
	newEVM.context.Message.Caller = e.context.Contract.Address

	//engine.Log.Error("evm commonCreate %s", newEVM.context.Message.Caller.B58String())

	ret, err := newEVM.ExecuteContract(true, param.OpCode)
	if err != nil {
		return nil, err
	}
	if ret.ExitOpCode == opcodes.REVERT {
		msg, err := abi.UnpackRevert(ret.ResultData)
		if err != nil {
			return ret.ResultData, err
		}
		return ret.ResultData, fmt.Errorf("%s", msg)
	}
	e.instructions.SetGasLimit(ret.GasLeft)
	return ret.ResultData, err
}

func closure(param instructions.ClosureParam) ([]byte, error) {
	evm, ok := param.VM.(*EVM)
	if !ok {
		return nil, utils.ErrInvalidEVMInstance
	}
	switch param.OpCode {
	case opcodes.CALL, opcodes.CALLCODE, opcodes.DELEGATECALL, opcodes.STATICCALL:
		return evm.commonCall(param)
	case opcodes.CREATE, opcodes.CREATE2:
		return evm.commonCreate(param)
	}

	return nil, nil
}
