// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package token20

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/abi/bind"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/libp2parea/adapter/engine"
	"math/big"
	"strings"
	"time"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
)

// Token20MetaData contains all meta data concerning the Token20 contract.
var Token20MetaData = &bind.MetaData{
	ABI: "[{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"freeze\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"unfreeze\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdrawEther\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"initialSupply\",\"type\":\"uint256\"},{\"name\":\"tokenName\",\"type\":\"string\"},{\"name\":\"decimalUnits\",\"type\":\"uint8\"},{\"name\":\"tokenSymbol\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Freeze\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Unfreeze\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"freezeOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b506040516200174a3803806200174a8339810180604052810190808051906020019092919080518201929190602001805190602001909291908051820192919050505083600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550836003819055508260009080519060200190620000b792919062000137565b508060019080519060200190620000d092919062000137565b5081600260006101000a81548160ff021916908360ff16021790555033600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050505050620001e6565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106200017a57805160ff1916838001178555620001ab565b82800160010185558215620001ab579182015b82811115620001aa5782518255916020019190600101906200018d565b5b509050620001ba9190620001be565b5090565b620001e391905b80821115620001df576000816000905550600101620001c5565b5090565b90565b61155480620001f66000396000f3006080604052600436106100db576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806306fdde03146100dd578063095ea7b31461016d57806318160ddd146101d257806323b872dd146101fd578063313ce567146102825780633bed33ce146102b357806342966c68146102e05780636623fc461461032557806370a082311461036a5780638da5cb5b146103c157806395d89b4114610418578063a9059cbb146104a8578063cd4217c1146104f5578063d7a78db81461054c578063dd62ed3e14610591575b005b3480156100e957600080fd5b506100f2610608565b6040518080602001828103825283818151815260200191508051906020019080838360005b83811015610132578082015181840152602081019050610117565b50505050905090810190601f16801561015f5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561017957600080fd5b506101b8600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506106a6565b604051808215151515815260200191505060405180910390f35b3480156101de57600080fd5b506101e7610741565b6040518082815260200191505060405180910390f35b34801561020957600080fd5b50610268600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610747565b604051808215151515815260200191505060405180910390f35b34801561028e57600080fd5b50610297610b6b565b604051808260ff1660ff16815260200191505060405180910390f35b3480156102bf57600080fd5b506102de60048036038101908080359060200190929190505050610b7e565b005b3480156102ec57600080fd5b5061030b60048036038101908080359060200190929190505050610c46565b604051808215151515815260200191505060405180910390f35b34801561033157600080fd5b5061035060048036038101908080359060200190929190505050610d98565b604051808215151515815260200191505060405180910390f35b34801561037657600080fd5b506103ab600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610f64565b6040518082815260200191505060405180910390f35b3480156103cd57600080fd5b506103d6610f7c565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b34801561042457600080fd5b5061042d610fa2565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561046d578082015181840152602081019050610452565b50505050905090810190601f16801561049a5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156104b457600080fd5b506104f3600480360381019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050611040565b005b34801561050157600080fd5b50610536600480360381019080803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506112cd565b6040518082815260200191505060405180910390f35b34801561055857600080fd5b50610577600480360381019080803590602001909291905050506112e5565b604051808215151515815260200191505060405180910390f35b34801561059d57600080fd5b506105f2600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506114b1565b6040518082815260200191505060405180910390f35b60008054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561069e5780601f106106735761010080835404028352916020019161069e565b820191906000526020600020905b81548152906001019060200180831161068157829003601f168201915b505050505081565b600080821115156106b657600080fd5b81600760003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506001905092915050565b60035481565b6000808373ffffffffffffffffffffffffffffffffffffffff16141561076c57600080fd5b60008211151561077b57600080fd5b81600560008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410156107c757600080fd5b600560008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205482600560008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205401101561085457600080fd5b600760008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020548211156108dd57600080fd5b610926600560008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114d6565b600560008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506109b2600560008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114ef565b600560008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550610a7b600760008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114d6565b600760008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a3600190509392505050565b600260009054906101000a900460ff1681565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16141515610bda57600080fd5b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610c42573d6000803e3d6000fd5b5050565b600081600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541015610c9457600080fd5b600082111515610ca357600080fd5b610cec600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114d6565b600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550610d3b600354836114d6565b6003819055503373ffffffffffffffffffffffffffffffffffffffff167fcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5836040518082815260200191505060405180910390a260019050919050565b600081600660003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541015610de657600080fd5b600082111515610df557600080fd5b610e3e600660003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114d6565b600660003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550610eca600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114ef565b600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055503373ffffffffffffffffffffffffffffffffffffffff167f2cfce4af01bcb9d6cf6c84ee1b7c491100b8695368264146a94d71e10a63083f836040518082815260200191505060405180910390a260019050919050565b60056020528060005260406000206000915090505481565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60018054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156110385780601f1061100d57610100808354040283529160200191611038565b820191906000526020600020905b81548152906001019060200180831161101b57829003601f168201915b505050505081565b60008273ffffffffffffffffffffffffffffffffffffffff16141561106457600080fd5b60008111151561107357600080fd5b80600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410156110bf57600080fd5b600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205481600560008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205401101561114c57600080fd5b611195600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054826114d6565b600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550611221600560008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054826114ef565b600560008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef836040518082815260200191505060405180910390a35050565b60066020528060005260406000206000915090505481565b600081600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054101561133357600080fd5b60008211151561134257600080fd5b61138b600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114d6565b600560003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550611417600660003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054836114ef565b600660003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055503373ffffffffffffffffffffffffffffffffffffffff167ff97a274face0b5517365ad396b1fdba6f68bd3135ef603e44272adba3af5a1e0836040518082815260200191505060405180910390a260019050919050565b6007602052816000526040600020602052806000526040600020600091509150505481565b60006114e483831115611519565b818303905092915050565b600080828401905061150f84821015801561150a5750838210155b611519565b8091505092915050565b80151561152557600080fd5b505600a165627a7a72305820923750b424c610a9f8d481a6ce9c66c50d0b0e5f777fab747182d56e849524220029",
}

// Token20ABI is the input ABI used to generate the binding from.
// Deprecated: Use Token20MetaData.ABI instead.
var Token20ABI = Token20MetaData.ABI

// Token20Bin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use Token20MetaData.Bin instead.
var Token20Bin = Token20MetaData.Bin

func (_Token20 *Token20) DeployToken20(opts *bind.TransactionOpts, initialSupply *big.Int, tokenName string, decimalUnits uint8, tokenSymbol string) (string, string, error) {

	return _Token20.contract.DeployContract(opts, Token20Bin, initialSupply, tokenName, decimalUnits, tokenSymbol)
}

type Token20 struct {
	contract *bind.BoundContract
}

// NewToken20 creates a new instance of Token20, bound to a specific deployed contract.
func NewToken20(api string) (*Token20, error) {
	contract, err := bindToken20(api)
	if err != nil {
		return nil, err
	}
	return &Token20{contract: contract}, nil
}

// bindToken20 binds a generic wrapper to an already deployed contract.
func bindToken20(api string) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(Token20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(api, parsed), nil
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_Token20 *Token20) Allowance(opts *bind.TransactionOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "allowance", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_Token20 *Token20) BalanceOf(opts *bind.TransactionOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Token20 *Token20) Decimals(opts *bind.TransactionOpts) (uint8, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// FreezeOf is a free data retrieval call binding the contract method 0xcd4217c1.
//
// Solidity: function freezeOf(address ) view returns(uint256)
func (_Token20 *Token20) FreezeOf(opts *bind.TransactionOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "freezeOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Token20 *Token20) Name(opts *bind.TransactionOpts) (string, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Token20 *Token20) Owner(opts *bind.TransactionOpts) (common.Address, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Token20 *Token20) Symbol(opts *bind.TransactionOpts) (string, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Token20 *Token20) TotalSupply(opts *bind.TransactionOpts) (*big.Int, error) {
	var out []interface{}
	err := _Token20.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Approve is a paid mutator sendRequest binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool success)
func (_Token20 *Token20) Approve(opts *bind.TransactionOpts, _spender common.Address, _value *big.Int) (string, error) {
	return _Token20.contract.Transfer(opts, "approve", _spender, _value)
}

// Burn is a paid mutator sendRequest binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 _value) returns(bool success)
func (_Token20 *Token20) Burn(opts *bind.TransactionOpts, _value *big.Int) (string, error) {
	return _Token20.contract.Transfer(opts, "burn", _value)
}

// Freeze is a paid mutator sendRequest binding the contract method 0xd7a78db8.
//
// Solidity: function freeze(uint256 _value) returns(bool success)
func (_Token20 *Token20) Freeze(opts *bind.TransactionOpts, _value *big.Int) (string, error) {
	return _Token20.contract.Transfer(opts, "freeze", _value)
}

// Transfer is a paid mutator sendRequest binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns()
func (_Token20 *Token20) Transfer(opts *bind.TransactionOpts, _to common.Address, _value *big.Int) (string, error) {
	return _Token20.contract.Transfer(opts, "transfer", _to, _value)
}

// TransferFrom is a paid mutator sendRequest binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool success)
func (_Token20 *Token20) TransferFrom(opts *bind.TransactionOpts, _from common.Address, _to common.Address, _value *big.Int) (string, error) {
	return _Token20.contract.Transfer(opts, "transferFrom", _from, _to, _value)
}

// Unfreeze is a paid mutator sendRequest binding the contract method 0x6623fc46.
//
// Solidity: function unfreeze(uint256 _value) returns(bool success)
func (_Token20 *Token20) Unfreeze(opts *bind.TransactionOpts, _value *big.Int) (string, error) {
	return _Token20.contract.Transfer(opts, "unfreeze", _value)
}

// WithdrawEther is a paid mutator sendRequest binding the contract method 0x3bed33ce.
//
// Solidity: function withdrawEther(uint256 amount) returns()
func (_Token20 *Token20) WithdrawEther(opts *bind.TransactionOpts, amount *big.Int) (string, error) {
	return _Token20.contract.Transfer(opts, "withdrawEther", amount)
}

// 合约事件部分
type Token20Subscribe struct {
	client go_protos.SubscriberClient
	abi    abi.ABI
}

func NewToken20Subscribe(ctx context.Context, endpoins string) (*Token20Subscribe, error) {
	conn, err := grpc.DialContext(ctx, endpoins, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := go_protos.NewSubscriberClient(conn)
	parsed, err := abi.JSON(strings.NewReader(Token20ABI))
	if err != nil {
		return nil, err
	}
	return &Token20Subscribe{
		client: client,
		abi:    parsed,
	}, nil
}
func (_Token20 *Token20Subscribe) UnpackLog(out interface{}, event string, contractEventInfo *go_protos.ContractEventInfo) error {
	if contractEventInfo.Topic != common.Bytes2Hex(_Token20.abi.Events[event].ID.Bytes()) {
		return errors.New("事件签名不符")
	}
	eventData := contractEventInfo.EventData
	logData := eventData[len(eventData)-1]
	//if log.Topics[0] != _Token20.abi.Events[event].ID {
	//	return fmt.Errorf("event signature mismatch")
	//}
	if logData != "" {
		if err := _Token20.abi.UnpackIntoInterface(out, event, common.Hex2Bytes(logData)); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range _Token20.abi.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	//去除掉第一个主题
	topics := []common.Hash{}
	for i := 0; i < len(eventData)-1; i++ {
		topics = append(topics, common.BytesToHash(common.Hex2Bytes(eventData[i])))
	}
	return abi.ParseTopics(out, indexed, topics)
}

// 含有event事件
type EventQuery struct {
	StartBlock   int64
	EndBlock     int64
	ContractAddr string
	Topic        string
}

// 生成请求体
func createPayload(query *EventQuery) ([]byte, error) {
	payload := &go_protos.SubscribeContractEventPayload{
		Topic:           query.Topic,
		ContractAddress: query.ContractAddr,
		StartBlock:      query.StartBlock,
		EndBlock:        query.EndBlock,
	}
	bt, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return bt, nil
}
func createReq(txId string, reqType go_protos.ReqType, payload []byte) *go_protos.SubscriberReq {
	header := &go_protos.ReqHeader{ReqType: reqType, TxId: txId, Timestamp: config.TimeNow().Unix(), ExpTime: 0}
	req := &go_protos.SubscriberReq{
		Header:  header,
		Payload: payload,
	}
	return req
}

// Token20Burn represents a Burn event raised by the Token20 contract.
type LogEventToken20Burn struct {
	From  common.Address
	Value *big.Int
}

func (_Token20 *Token20Subscribe) SubscribeEventBurn(ctx context.Context, query *EventQuery) (<-chan LogEventToken20Burn, error) {
	//执行订阅
	payload, _ := createPayload(query)
	txid := uuid.New().String()
	req := createReq(txid, go_protos.ReqType_SUBSCRIBE_CONTRACT_EVENT_INFO, payload)
	resp, err := _Token20.client.EventSub(ctx, req)
	if err != nil {
		return nil, err
	}
	c := make(chan LogEventToken20Burn)
	go func() {
		defer close(c)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := resp.Recv()
				if err != nil {
					engine.Log.Error("")
					return
				}
				//解析数据
				events := &go_protos.ContractEventInfoList{}
				err = proto.Unmarshal(result.Data, events)
				if err != nil {
					engine.Log.Error("事件解析失败")
					return
				}
				for _, event := range events.ContractEvents {
					//
					log := _Token20.UnPackBurnLog(event)
					c <- log
				}
				continue
			}
		}
	}()
	return c, nil
}

// 解析日志
func (_Token20 *Token20Subscribe) UnPackBurnLog(contractEventInfo *go_protos.ContractEventInfo) LogEventToken20Burn {

	var log LogEventToken20Burn
	_Token20.UnpackLog(&log, "Burn", contractEventInfo)
	return log
}

// Token20Freeze represents a Freeze event raised by the Token20 contract.
type LogEventToken20Freeze struct {
	From  common.Address
	Value *big.Int
}

func (_Token20 *Token20Subscribe) SubscribeEventFreeze(ctx context.Context, query *EventQuery) (<-chan LogEventToken20Freeze, error) {
	//执行订阅
	payload, _ := createPayload(query)
	txid := uuid.New().String()
	req := createReq(txid, go_protos.ReqType_SUBSCRIBE_CONTRACT_EVENT_INFO, payload)
	resp, err := _Token20.client.EventSub(ctx, req)
	if err != nil {
		return nil, err
	}
	c := make(chan LogEventToken20Freeze)
	go func() {
		defer close(c)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := resp.Recv()
				if err != nil {
					engine.Log.Error("")
					return
				}
				//解析数据
				events := &go_protos.ContractEventInfoList{}
				err = proto.Unmarshal(result.Data, events)
				if err != nil {
					engine.Log.Error("事件解析失败")
					return
				}
				for _, event := range events.ContractEvents {
					//
					log := _Token20.UnPackFreezeLog(event)
					c <- log
				}
				continue
			}
		}
	}()
	return c, nil
}

// 解析日志
func (_Token20 *Token20Subscribe) UnPackFreezeLog(contractEventInfo *go_protos.ContractEventInfo) LogEventToken20Freeze {

	var log LogEventToken20Freeze
	_Token20.UnpackLog(&log, "Freeze", contractEventInfo)
	return log
}

// Token20Transfer represents a Transfer event raised by the Token20 contract.
type LogEventToken20Transfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
}

func (_Token20 *Token20Subscribe) SubscribeEventTransfer(ctx context.Context, query *EventQuery) (<-chan LogEventToken20Transfer, error) {
	//执行订阅
	payload, _ := createPayload(query)
	txid := uuid.New().String()
	req := createReq(txid, go_protos.ReqType_SUBSCRIBE_CONTRACT_EVENT_INFO, payload)
	resp, err := _Token20.client.EventSub(ctx, req)
	if err != nil {
		return nil, err
	}
	c := make(chan LogEventToken20Transfer)
	go func() {
		defer close(c)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := resp.Recv()
				if err != nil {
					engine.Log.Error("")
					return
				}
				//解析数据
				events := &go_protos.ContractEventInfoList{}
				err = proto.Unmarshal(result.Data, events)
				if err != nil {
					engine.Log.Error("事件解析失败")
					return
				}
				for _, event := range events.ContractEvents {
					//
					log := _Token20.UnPackTransferLog(event)
					c <- log
				}
				continue
			}
		}
	}()
	return c, nil
}

// 解析日志
func (_Token20 *Token20Subscribe) UnPackTransferLog(contractEventInfo *go_protos.ContractEventInfo) LogEventToken20Transfer {

	var log LogEventToken20Transfer
	_Token20.UnpackLog(&log, "Transfer", contractEventInfo)
	return log
}

// Token20Unfreeze represents a Unfreeze event raised by the Token20 contract.
type LogEventToken20Unfreeze struct {
	From  common.Address
	Value *big.Int
}

func (_Token20 *Token20Subscribe) SubscribeEventUnfreeze(ctx context.Context, query *EventQuery) (<-chan LogEventToken20Unfreeze, error) {
	//执行订阅
	payload, _ := createPayload(query)
	txid := uuid.New().String()
	req := createReq(txid, go_protos.ReqType_SUBSCRIBE_CONTRACT_EVENT_INFO, payload)
	resp, err := _Token20.client.EventSub(ctx, req)
	if err != nil {
		return nil, err
	}
	c := make(chan LogEventToken20Unfreeze)
	go func() {
		defer close(c)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := resp.Recv()
				if err != nil {
					engine.Log.Error("")
					return
				}
				//解析数据
				events := &go_protos.ContractEventInfoList{}
				err = proto.Unmarshal(result.Data, events)
				if err != nil {
					engine.Log.Error("事件解析失败")
					return
				}
				for _, event := range events.ContractEvents {
					//
					log := _Token20.UnPackUnfreezeLog(event)
					c <- log
				}
				continue
			}
		}
	}()
	return c, nil
}

// 解析日志
func (_Token20 *Token20Subscribe) UnPackUnfreezeLog(contractEventInfo *go_protos.ContractEventInfo) LogEventToken20Unfreeze {

	var log LogEventToken20Unfreeze
	_Token20.UnpackLog(&log, "Unfreeze", contractEventInfo)
	return log
}

// Fallback is a paid mutator sendRequest binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Token20 *Token20) Fallback(opts *bind.TransactionOpts, calldata []byte) {
	_Token20.contract.RawTransact(opts, calldata)
}
