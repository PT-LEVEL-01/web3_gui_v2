// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package full_node

import (
	"errors"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// TrxUsdtTokenMetaData contains all meta data concerning the TrxUsdtToken contract.
var TrxUsdtTokenMetaData = &bind.MetaData{
	ABI: "[{\"constant\":true,\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_upgradedAddress\",\"type\":\"address\"}],\"name\":\"deprecate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"deprecated\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_evilUser\",\"type\":\"address\"}],\"name\":\"addBlackList\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_from\",\"type\":\"address\"},{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"upgradedAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maximumFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"_totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"unpause\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_maker\",\"type\":\"address\"}],\"name\":\"getBlackListStatus\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseApproval\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"who\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"calcFee\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"pause\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"who\",\"type\":\"address\"}],\"name\":\"oldBalanceOf\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newBasisPoints\",\"type\":\"uint256\"},{\"name\":\"newMaxFee\",\"type\":\"uint256\"}],\"name\":\"setParams\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"issue\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_spender\",\"type\":\"address\"},{\"name\":\"_addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseApproval\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"redeem\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"_spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"name\":\"remaining\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"basisPointsRate\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"isBlackListed\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_clearedUser\",\"type\":\"address\"}],\"name\":\"removeBlackList\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"MAX_UINT\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_blackListedUser\",\"type\":\"address\"}],\"name\":\"destroyBlackFunds\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_initialSupply\",\"type\":\"uint256\"},{\"name\":\"_name\",\"type\":\"string\"},{\"name\":\"_symbol\",\"type\":\"string\"},{\"name\":\"_decimals\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_blackListedUser\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"_balance\",\"type\":\"uint256\"}],\"name\":\"DestroyedBlackFunds\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Issue\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Redeem\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"Deprecate\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"AddedBlackList\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"_user\",\"type\":\"address\"}],\"name\":\"RemovedBlackList\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"feeBasisPoints\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"maxFee\",\"type\":\"uint256\"}],\"name\":\"Params\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Pause\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"Unpause\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"}]",
}

// TrxUsdtTokenABI is the input ABI used to generate the binding from.
// Deprecated: Use TrxUsdtTokenMetaData.ABI instead.
var TrxUsdtTokenABI = TrxUsdtTokenMetaData.ABI

// TrxUsdtToken is an auto generated Go binding around an Ethereum contract.
type TrxUsdtToken struct {
	TrxUsdtTokenCaller     // Read-only binding to the contract
	TrxUsdtTokenTransactor // Write-only binding to the contract
	TrxUsdtTokenFilterer   // Log filterer for contract events
}

// TrxUsdtTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type TrxUsdtTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TrxUsdtTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TrxUsdtTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TrxUsdtTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TrxUsdtTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TrxUsdtTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TrxUsdtTokenSession struct {
	Contract     *TrxUsdtToken     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TrxUsdtTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TrxUsdtTokenCallerSession struct {
	Contract *TrxUsdtTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// TrxUsdtTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TrxUsdtTokenTransactorSession struct {
	Contract     *TrxUsdtTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// TrxUsdtTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type TrxUsdtTokenRaw struct {
	Contract *TrxUsdtToken // Generic contract binding to access the raw methods on
}

// TrxUsdtTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TrxUsdtTokenCallerRaw struct {
	Contract *TrxUsdtTokenCaller // Generic read-only contract binding to access the raw methods on
}

// TrxUsdtTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TrxUsdtTokenTransactorRaw struct {
	Contract *TrxUsdtTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTrxUsdtToken creates a new instance of TrxUsdtToken, bound to a specific deployed contract.
func NewTrxUsdtToken(address common.Address, backend bind.ContractBackend) (*TrxUsdtToken, error) {
	contract, err := bindTrxUsdtToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtToken{TrxUsdtTokenCaller: TrxUsdtTokenCaller{contract: contract}, TrxUsdtTokenTransactor: TrxUsdtTokenTransactor{contract: contract}, TrxUsdtTokenFilterer: TrxUsdtTokenFilterer{contract: contract}}, nil
}

// NewTrxUsdtTokenCaller creates a new read-only instance of TrxUsdtToken, bound to a specific deployed contract.
func NewTrxUsdtTokenCaller(address common.Address, caller bind.ContractCaller) (*TrxUsdtTokenCaller, error) {
	contract, err := bindTrxUsdtToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenCaller{contract: contract}, nil
}

// NewTrxUsdtTokenTransactor creates a new write-only instance of TrxUsdtToken, bound to a specific deployed contract.
func NewTrxUsdtTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*TrxUsdtTokenTransactor, error) {
	contract, err := bindTrxUsdtToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenTransactor{contract: contract}, nil
}

// NewTrxUsdtTokenFilterer creates a new log filterer instance of TrxUsdtToken, bound to a specific deployed contract.
func NewTrxUsdtTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*TrxUsdtTokenFilterer, error) {
	contract, err := bindTrxUsdtToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenFilterer{contract: contract}, nil
}

// bindTrxUsdtToken binds a generic wrapper to an already deployed contract.
func bindTrxUsdtToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TrxUsdtTokenABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TrxUsdtToken *TrxUsdtTokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TrxUsdtToken.Contract.TrxUsdtTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TrxUsdtToken *TrxUsdtTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.TrxUsdtTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TrxUsdtToken *TrxUsdtTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.TrxUsdtTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_TrxUsdtToken *TrxUsdtTokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _TrxUsdtToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_TrxUsdtToken *TrxUsdtTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_TrxUsdtToken *TrxUsdtTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.contract.Transact(opts, method, params...)
}

// MAXUINT is a free data retrieval call binding the contract method 0xe5b5019a.
//
// Solidity: function MAX_UINT() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) MAXUINT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "MAX_UINT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXUINT is a free data retrieval call binding the contract method 0xe5b5019a.
//
// Solidity: function MAX_UINT() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) MAXUINT() (*big.Int, error) {
	return _TrxUsdtToken.Contract.MAXUINT(&_TrxUsdtToken.CallOpts)
}

// MAXUINT is a free data retrieval call binding the contract method 0xe5b5019a.
//
// Solidity: function MAX_UINT() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) MAXUINT() (*big.Int, error) {
	return _TrxUsdtToken.Contract.MAXUINT(&_TrxUsdtToken.CallOpts)
}

// TotalSupply1 is a free data retrieval call binding the contract method 0x3eaaf86b.
//
// Solidity: function _totalSupply() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) TotalSupply1(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "_totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply1 is a free data retrieval call binding the contract method 0x3eaaf86b.
//
// Solidity: function _totalSupply() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) TotalSupply1() (*big.Int, error) {
	return _TrxUsdtToken.Contract.TotalSupply1(&_TrxUsdtToken.CallOpts)
}

// TotalSupply1 is a free data retrieval call binding the contract method 0x3eaaf86b.
//
// Solidity: function _totalSupply() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) TotalSupply1() (*big.Int, error) {
	return _TrxUsdtToken.Contract.TotalSupply1(&_TrxUsdtToken.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_TrxUsdtToken *TrxUsdtTokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, _spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "allowance", _owner, _spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_TrxUsdtToken *TrxUsdtTokenSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _TrxUsdtToken.Contract.Allowance(&_TrxUsdtToken.CallOpts, _owner, _spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address _spender) view returns(uint256 remaining)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) Allowance(_owner common.Address, _spender common.Address) (*big.Int, error) {
	return _TrxUsdtToken.Contract.Allowance(&_TrxUsdtToken.CallOpts, _owner, _spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address who) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) BalanceOf(opts *bind.CallOpts, who common.Address) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "balanceOf", who)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address who) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) BalanceOf(who common.Address) (*big.Int, error) {
	return _TrxUsdtToken.Contract.BalanceOf(&_TrxUsdtToken.CallOpts, who)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address who) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) BalanceOf(who common.Address) (*big.Int, error) {
	return _TrxUsdtToken.Contract.BalanceOf(&_TrxUsdtToken.CallOpts, who)
}

// BasisPointsRate is a free data retrieval call binding the contract method 0xdd644f72.
//
// Solidity: function basisPointsRate() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) BasisPointsRate(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "basisPointsRate")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BasisPointsRate is a free data retrieval call binding the contract method 0xdd644f72.
//
// Solidity: function basisPointsRate() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) BasisPointsRate() (*big.Int, error) {
	return _TrxUsdtToken.Contract.BasisPointsRate(&_TrxUsdtToken.CallOpts)
}

// BasisPointsRate is a free data retrieval call binding the contract method 0xdd644f72.
//
// Solidity: function basisPointsRate() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) BasisPointsRate() (*big.Int, error) {
	return _TrxUsdtToken.Contract.BasisPointsRate(&_TrxUsdtToken.CallOpts)
}

// CalcFee is a free data retrieval call binding the contract method 0x75dc7d8c.
//
// Solidity: function calcFee(uint256 _value) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) CalcFee(opts *bind.CallOpts, _value *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "calcFee", _value)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalcFee is a free data retrieval call binding the contract method 0x75dc7d8c.
//
// Solidity: function calcFee(uint256 _value) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) CalcFee(_value *big.Int) (*big.Int, error) {
	return _TrxUsdtToken.Contract.CalcFee(&_TrxUsdtToken.CallOpts, _value)
}

// CalcFee is a free data retrieval call binding the contract method 0x75dc7d8c.
//
// Solidity: function calcFee(uint256 _value) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) CalcFee(_value *big.Int) (*big.Int, error) {
	return _TrxUsdtToken.Contract.CalcFee(&_TrxUsdtToken.CallOpts, _value)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TrxUsdtToken *TrxUsdtTokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TrxUsdtToken *TrxUsdtTokenSession) Decimals() (uint8, error) {
	return _TrxUsdtToken.Contract.Decimals(&_TrxUsdtToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) Decimals() (uint8, error) {
	return _TrxUsdtToken.Contract.Decimals(&_TrxUsdtToken.CallOpts)
}

// Deprecated is a free data retrieval call binding the contract method 0x0e136b19.
//
// Solidity: function deprecated() view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCaller) Deprecated(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "deprecated")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Deprecated is a free data retrieval call binding the contract method 0x0e136b19.
//
// Solidity: function deprecated() view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) Deprecated() (bool, error) {
	return _TrxUsdtToken.Contract.Deprecated(&_TrxUsdtToken.CallOpts)
}

// Deprecated is a free data retrieval call binding the contract method 0x0e136b19.
//
// Solidity: function deprecated() view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) Deprecated() (bool, error) {
	return _TrxUsdtToken.Contract.Deprecated(&_TrxUsdtToken.CallOpts)
}

// GetBlackListStatus is a free data retrieval call binding the contract method 0x59bf1abe.
//
// Solidity: function getBlackListStatus(address _maker) view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCaller) GetBlackListStatus(opts *bind.CallOpts, _maker common.Address) (bool, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "getBlackListStatus", _maker)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetBlackListStatus is a free data retrieval call binding the contract method 0x59bf1abe.
//
// Solidity: function getBlackListStatus(address _maker) view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) GetBlackListStatus(_maker common.Address) (bool, error) {
	return _TrxUsdtToken.Contract.GetBlackListStatus(&_TrxUsdtToken.CallOpts, _maker)
}

// GetBlackListStatus is a free data retrieval call binding the contract method 0x59bf1abe.
//
// Solidity: function getBlackListStatus(address _maker) view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) GetBlackListStatus(_maker common.Address) (bool, error) {
	return _TrxUsdtToken.Contract.GetBlackListStatus(&_TrxUsdtToken.CallOpts, _maker)
}

// IsBlackListed is a free data retrieval call binding the contract method 0xe47d6060.
//
// Solidity: function isBlackListed(address ) view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCaller) IsBlackListed(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "isBlackListed", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBlackListed is a free data retrieval call binding the contract method 0xe47d6060.
//
// Solidity: function isBlackListed(address ) view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) IsBlackListed(arg0 common.Address) (bool, error) {
	return _TrxUsdtToken.Contract.IsBlackListed(&_TrxUsdtToken.CallOpts, arg0)
}

// IsBlackListed is a free data retrieval call binding the contract method 0xe47d6060.
//
// Solidity: function isBlackListed(address ) view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) IsBlackListed(arg0 common.Address) (bool, error) {
	return _TrxUsdtToken.Contract.IsBlackListed(&_TrxUsdtToken.CallOpts, arg0)
}

// MaximumFee is a free data retrieval call binding the contract method 0x35390714.
//
// Solidity: function maximumFee() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) MaximumFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "maximumFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaximumFee is a free data retrieval call binding the contract method 0x35390714.
//
// Solidity: function maximumFee() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) MaximumFee() (*big.Int, error) {
	return _TrxUsdtToken.Contract.MaximumFee(&_TrxUsdtToken.CallOpts)
}

// MaximumFee is a free data retrieval call binding the contract method 0x35390714.
//
// Solidity: function maximumFee() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) MaximumFee() (*big.Int, error) {
	return _TrxUsdtToken.Contract.MaximumFee(&_TrxUsdtToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TrxUsdtToken *TrxUsdtTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TrxUsdtToken *TrxUsdtTokenSession) Name() (string, error) {
	return _TrxUsdtToken.Contract.Name(&_TrxUsdtToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) Name() (string, error) {
	return _TrxUsdtToken.Contract.Name(&_TrxUsdtToken.CallOpts)
}

// OldBalanceOf is a free data retrieval call binding the contract method 0xb7a3446c.
//
// Solidity: function oldBalanceOf(address who) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) OldBalanceOf(opts *bind.CallOpts, who common.Address) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "oldBalanceOf", who)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OldBalanceOf is a free data retrieval call binding the contract method 0xb7a3446c.
//
// Solidity: function oldBalanceOf(address who) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) OldBalanceOf(who common.Address) (*big.Int, error) {
	return _TrxUsdtToken.Contract.OldBalanceOf(&_TrxUsdtToken.CallOpts, who)
}

// OldBalanceOf is a free data retrieval call binding the contract method 0xb7a3446c.
//
// Solidity: function oldBalanceOf(address who) view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) OldBalanceOf(who common.Address) (*big.Int, error) {
	return _TrxUsdtToken.Contract.OldBalanceOf(&_TrxUsdtToken.CallOpts, who)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TrxUsdtToken *TrxUsdtTokenCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TrxUsdtToken *TrxUsdtTokenSession) Owner() (common.Address, error) {
	return _TrxUsdtToken.Contract.Owner(&_TrxUsdtToken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) Owner() (common.Address, error) {
	return _TrxUsdtToken.Contract.Owner(&_TrxUsdtToken.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) Paused() (bool, error) {
	return _TrxUsdtToken.Contract.Paused(&_TrxUsdtToken.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) Paused() (bool, error) {
	return _TrxUsdtToken.Contract.Paused(&_TrxUsdtToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TrxUsdtToken *TrxUsdtTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TrxUsdtToken *TrxUsdtTokenSession) Symbol() (string, error) {
	return _TrxUsdtToken.Contract.Symbol(&_TrxUsdtToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) Symbol() (string, error) {
	return _TrxUsdtToken.Contract.Symbol(&_TrxUsdtToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenSession) TotalSupply() (*big.Int, error) {
	return _TrxUsdtToken.Contract.TotalSupply(&_TrxUsdtToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) TotalSupply() (*big.Int, error) {
	return _TrxUsdtToken.Contract.TotalSupply(&_TrxUsdtToken.CallOpts)
}

// UpgradedAddress is a free data retrieval call binding the contract method 0x26976e3f.
//
// Solidity: function upgradedAddress() view returns(address)
func (_TrxUsdtToken *TrxUsdtTokenCaller) UpgradedAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _TrxUsdtToken.contract.Call(opts, &out, "upgradedAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// UpgradedAddress is a free data retrieval call binding the contract method 0x26976e3f.
//
// Solidity: function upgradedAddress() view returns(address)
func (_TrxUsdtToken *TrxUsdtTokenSession) UpgradedAddress() (common.Address, error) {
	return _TrxUsdtToken.Contract.UpgradedAddress(&_TrxUsdtToken.CallOpts)
}

// UpgradedAddress is a free data retrieval call binding the contract method 0x26976e3f.
//
// Solidity: function upgradedAddress() view returns(address)
func (_TrxUsdtToken *TrxUsdtTokenCallerSession) UpgradedAddress() (common.Address, error) {
	return _TrxUsdtToken.Contract.UpgradedAddress(&_TrxUsdtToken.CallOpts)
}

// AddBlackList is a paid mutator transaction binding the contract method 0x0ecb93c0.
//
// Solidity: function addBlackList(address _evilUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) AddBlackList(opts *bind.TransactOpts, _evilUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "addBlackList", _evilUser)
}

// AddBlackList is a paid mutator transaction binding the contract method 0x0ecb93c0.
//
// Solidity: function addBlackList(address _evilUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) AddBlackList(_evilUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.AddBlackList(&_TrxUsdtToken.TransactOpts, _evilUser)
}

// AddBlackList is a paid mutator transaction binding the contract method 0x0ecb93c0.
//
// Solidity: function addBlackList(address _evilUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) AddBlackList(_evilUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.AddBlackList(&_TrxUsdtToken.TransactOpts, _evilUser)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactor) Approve(opts *bind.TransactOpts, _spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "approve", _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Approve(&_TrxUsdtToken.TransactOpts, _spender, _value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address _spender, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) Approve(_spender common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Approve(&_TrxUsdtToken.TransactOpts, _spender, _value)
}

// DecreaseApproval is a paid mutator transaction binding the contract method 0x66188463.
//
// Solidity: function decreaseApproval(address _spender, uint256 _subtractedValue) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactor) DecreaseApproval(opts *bind.TransactOpts, _spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "decreaseApproval", _spender, _subtractedValue)
}

// DecreaseApproval is a paid mutator transaction binding the contract method 0x66188463.
//
// Solidity: function decreaseApproval(address _spender, uint256 _subtractedValue) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) DecreaseApproval(_spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.DecreaseApproval(&_TrxUsdtToken.TransactOpts, _spender, _subtractedValue)
}

// DecreaseApproval is a paid mutator transaction binding the contract method 0x66188463.
//
// Solidity: function decreaseApproval(address _spender, uint256 _subtractedValue) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) DecreaseApproval(_spender common.Address, _subtractedValue *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.DecreaseApproval(&_TrxUsdtToken.TransactOpts, _spender, _subtractedValue)
}

// Deprecate is a paid mutator transaction binding the contract method 0x0753c30c.
//
// Solidity: function deprecate(address _upgradedAddress) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) Deprecate(opts *bind.TransactOpts, _upgradedAddress common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "deprecate", _upgradedAddress)
}

// Deprecate is a paid mutator transaction binding the contract method 0x0753c30c.
//
// Solidity: function deprecate(address _upgradedAddress) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) Deprecate(_upgradedAddress common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Deprecate(&_TrxUsdtToken.TransactOpts, _upgradedAddress)
}

// Deprecate is a paid mutator transaction binding the contract method 0x0753c30c.
//
// Solidity: function deprecate(address _upgradedAddress) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) Deprecate(_upgradedAddress common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Deprecate(&_TrxUsdtToken.TransactOpts, _upgradedAddress)
}

// DestroyBlackFunds is a paid mutator transaction binding the contract method 0xf3bdc228.
//
// Solidity: function destroyBlackFunds(address _blackListedUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) DestroyBlackFunds(opts *bind.TransactOpts, _blackListedUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "destroyBlackFunds", _blackListedUser)
}

// DestroyBlackFunds is a paid mutator transaction binding the contract method 0xf3bdc228.
//
// Solidity: function destroyBlackFunds(address _blackListedUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) DestroyBlackFunds(_blackListedUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.DestroyBlackFunds(&_TrxUsdtToken.TransactOpts, _blackListedUser)
}

// DestroyBlackFunds is a paid mutator transaction binding the contract method 0xf3bdc228.
//
// Solidity: function destroyBlackFunds(address _blackListedUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) DestroyBlackFunds(_blackListedUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.DestroyBlackFunds(&_TrxUsdtToken.TransactOpts, _blackListedUser)
}

// IncreaseApproval is a paid mutator transaction binding the contract method 0xd73dd623.
//
// Solidity: function increaseApproval(address _spender, uint256 _addedValue) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactor) IncreaseApproval(opts *bind.TransactOpts, _spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "increaseApproval", _spender, _addedValue)
}

// IncreaseApproval is a paid mutator transaction binding the contract method 0xd73dd623.
//
// Solidity: function increaseApproval(address _spender, uint256 _addedValue) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) IncreaseApproval(_spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.IncreaseApproval(&_TrxUsdtToken.TransactOpts, _spender, _addedValue)
}

// IncreaseApproval is a paid mutator transaction binding the contract method 0xd73dd623.
//
// Solidity: function increaseApproval(address _spender, uint256 _addedValue) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) IncreaseApproval(_spender common.Address, _addedValue *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.IncreaseApproval(&_TrxUsdtToken.TransactOpts, _spender, _addedValue)
}

// Issue is a paid mutator transaction binding the contract method 0xcc872b66.
//
// Solidity: function issue(uint256 amount) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) Issue(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "issue", amount)
}

// Issue is a paid mutator transaction binding the contract method 0xcc872b66.
//
// Solidity: function issue(uint256 amount) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) Issue(amount *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Issue(&_TrxUsdtToken.TransactOpts, amount)
}

// Issue is a paid mutator transaction binding the contract method 0xcc872b66.
//
// Solidity: function issue(uint256 amount) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) Issue(amount *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Issue(&_TrxUsdtToken.TransactOpts, amount)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) Pause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "pause")
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) Pause() (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Pause(&_TrxUsdtToken.TransactOpts)
}

// Pause is a paid mutator transaction binding the contract method 0x8456cb59.
//
// Solidity: function pause() returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) Pause() (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Pause(&_TrxUsdtToken.TransactOpts)
}

// Redeem is a paid mutator transaction binding the contract method 0xdb006a75.
//
// Solidity: function redeem(uint256 amount) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) Redeem(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "redeem", amount)
}

// Redeem is a paid mutator transaction binding the contract method 0xdb006a75.
//
// Solidity: function redeem(uint256 amount) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) Redeem(amount *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Redeem(&_TrxUsdtToken.TransactOpts, amount)
}

// Redeem is a paid mutator transaction binding the contract method 0xdb006a75.
//
// Solidity: function redeem(uint256 amount) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) Redeem(amount *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Redeem(&_TrxUsdtToken.TransactOpts, amount)
}

// RemoveBlackList is a paid mutator transaction binding the contract method 0xe4997dc5.
//
// Solidity: function removeBlackList(address _clearedUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) RemoveBlackList(opts *bind.TransactOpts, _clearedUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "removeBlackList", _clearedUser)
}

// RemoveBlackList is a paid mutator transaction binding the contract method 0xe4997dc5.
//
// Solidity: function removeBlackList(address _clearedUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) RemoveBlackList(_clearedUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.RemoveBlackList(&_TrxUsdtToken.TransactOpts, _clearedUser)
}

// RemoveBlackList is a paid mutator transaction binding the contract method 0xe4997dc5.
//
// Solidity: function removeBlackList(address _clearedUser) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) RemoveBlackList(_clearedUser common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.RemoveBlackList(&_TrxUsdtToken.TransactOpts, _clearedUser)
}

// SetParams is a paid mutator transaction binding the contract method 0xc0324c77.
//
// Solidity: function setParams(uint256 newBasisPoints, uint256 newMaxFee) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) SetParams(opts *bind.TransactOpts, newBasisPoints *big.Int, newMaxFee *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "setParams", newBasisPoints, newMaxFee)
}

// SetParams is a paid mutator transaction binding the contract method 0xc0324c77.
//
// Solidity: function setParams(uint256 newBasisPoints, uint256 newMaxFee) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) SetParams(newBasisPoints *big.Int, newMaxFee *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.SetParams(&_TrxUsdtToken.TransactOpts, newBasisPoints, newMaxFee)
}

// SetParams is a paid mutator transaction binding the contract method 0xc0324c77.
//
// Solidity: function setParams(uint256 newBasisPoints, uint256 newMaxFee) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) SetParams(newBasisPoints *big.Int, newMaxFee *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.SetParams(&_TrxUsdtToken.TransactOpts, newBasisPoints, newMaxFee)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "transfer", _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Transfer(&_TrxUsdtToken.TransactOpts, _to, _value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address _to, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) Transfer(_to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Transfer(&_TrxUsdtToken.TransactOpts, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactor) TransferFrom(opts *bind.TransactOpts, _from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "transferFrom", _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.TransferFrom(&_TrxUsdtToken.TransactOpts, _from, _to, _value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address _from, address _to, uint256 _value) returns(bool)
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) TransferFrom(_from common.Address, _to common.Address, _value *big.Int) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.TransferFrom(&_TrxUsdtToken.TransactOpts, _from, _to, _value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.TransferOwnership(&_TrxUsdtToken.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.TransferOwnership(&_TrxUsdtToken.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactor) Unpause(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _TrxUsdtToken.contract.Transact(opts, "unpause")
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_TrxUsdtToken *TrxUsdtTokenSession) Unpause() (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Unpause(&_TrxUsdtToken.TransactOpts)
}

// Unpause is a paid mutator transaction binding the contract method 0x3f4ba83a.
//
// Solidity: function unpause() returns()
func (_TrxUsdtToken *TrxUsdtTokenTransactorSession) Unpause() (*types.Transaction, error) {
	return _TrxUsdtToken.Contract.Unpause(&_TrxUsdtToken.TransactOpts)
}

// TrxUsdtTokenAddedBlackListIterator is returned from FilterAddedBlackList and is used to iterate over the raw logs and unpacked data for AddedBlackList events raised by the TrxUsdtToken contract.
type TrxUsdtTokenAddedBlackListIterator struct {
	Event *TrxUsdtTokenAddedBlackList // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenAddedBlackListIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenAddedBlackList)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenAddedBlackList)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenAddedBlackListIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenAddedBlackListIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenAddedBlackList represents a AddedBlackList event raised by the TrxUsdtToken contract.
type TrxUsdtTokenAddedBlackList struct {
	User common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAddedBlackList is a free log retrieval operation binding the contract event 0x42e160154868087d6bfdc0ca23d96a1c1cfa32f1b72ba9ba27b69b98a0d819dc.
//
// Solidity: event AddedBlackList(address indexed _user)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterAddedBlackList(opts *bind.FilterOpts, _user []common.Address) (*TrxUsdtTokenAddedBlackListIterator, error) {

	var _userRule []interface{}
	for _, _userItem := range _user {
		_userRule = append(_userRule, _userItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "AddedBlackList", _userRule)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenAddedBlackListIterator{contract: _TrxUsdtToken.contract, event: "AddedBlackList", logs: logs, sub: sub}, nil
}

// WatchAddedBlackList is a free log subscription operation binding the contract event 0x42e160154868087d6bfdc0ca23d96a1c1cfa32f1b72ba9ba27b69b98a0d819dc.
//
// Solidity: event AddedBlackList(address indexed _user)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchAddedBlackList(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenAddedBlackList, _user []common.Address) (event.Subscription, error) {

	var _userRule []interface{}
	for _, _userItem := range _user {
		_userRule = append(_userRule, _userItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "AddedBlackList", _userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenAddedBlackList)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "AddedBlackList", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAddedBlackList is a log parse operation binding the contract event 0x42e160154868087d6bfdc0ca23d96a1c1cfa32f1b72ba9ba27b69b98a0d819dc.
//
// Solidity: event AddedBlackList(address indexed _user)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseAddedBlackList(log types.Log) (*TrxUsdtTokenAddedBlackList, error) {
	event := new(TrxUsdtTokenAddedBlackList)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "AddedBlackList", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the TrxUsdtToken contract.
type TrxUsdtTokenApprovalIterator struct {
	Event *TrxUsdtTokenApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenApproval represents a Approval event raised by the TrxUsdtToken contract.
type TrxUsdtTokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*TrxUsdtTokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenApprovalIterator{contract: _TrxUsdtToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenApproval)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseApproval(log types.Log) (*TrxUsdtTokenApproval, error) {
	event := new(TrxUsdtTokenApproval)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenDeprecateIterator is returned from FilterDeprecate and is used to iterate over the raw logs and unpacked data for Deprecate events raised by the TrxUsdtToken contract.
type TrxUsdtTokenDeprecateIterator struct {
	Event *TrxUsdtTokenDeprecate // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenDeprecateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenDeprecate)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenDeprecate)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenDeprecateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenDeprecateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenDeprecate represents a Deprecate event raised by the TrxUsdtToken contract.
type TrxUsdtTokenDeprecate struct {
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterDeprecate is a free log retrieval operation binding the contract event 0xcc358699805e9a8b7f77b522628c7cb9abd07d9efb86b6fb616af1609036a99e.
//
// Solidity: event Deprecate(address newAddress)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterDeprecate(opts *bind.FilterOpts) (*TrxUsdtTokenDeprecateIterator, error) {

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Deprecate")
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenDeprecateIterator{contract: _TrxUsdtToken.contract, event: "Deprecate", logs: logs, sub: sub}, nil
}

// WatchDeprecate is a free log subscription operation binding the contract event 0xcc358699805e9a8b7f77b522628c7cb9abd07d9efb86b6fb616af1609036a99e.
//
// Solidity: event Deprecate(address newAddress)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchDeprecate(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenDeprecate) (event.Subscription, error) {

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Deprecate")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenDeprecate)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Deprecate", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeprecate is a log parse operation binding the contract event 0xcc358699805e9a8b7f77b522628c7cb9abd07d9efb86b6fb616af1609036a99e.
//
// Solidity: event Deprecate(address newAddress)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseDeprecate(log types.Log) (*TrxUsdtTokenDeprecate, error) {
	event := new(TrxUsdtTokenDeprecate)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Deprecate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenDestroyedBlackFundsIterator is returned from FilterDestroyedBlackFunds and is used to iterate over the raw logs and unpacked data for DestroyedBlackFunds events raised by the TrxUsdtToken contract.
type TrxUsdtTokenDestroyedBlackFundsIterator struct {
	Event *TrxUsdtTokenDestroyedBlackFunds // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenDestroyedBlackFundsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenDestroyedBlackFunds)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenDestroyedBlackFunds)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenDestroyedBlackFundsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenDestroyedBlackFundsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenDestroyedBlackFunds represents a DestroyedBlackFunds event raised by the TrxUsdtToken contract.
type TrxUsdtTokenDestroyedBlackFunds struct {
	BlackListedUser common.Address
	Balance         *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterDestroyedBlackFunds is a free log retrieval operation binding the contract event 0x61e6e66b0d6339b2980aecc6ccc0039736791f0ccde9ed512e789a7fbdd698c6.
//
// Solidity: event DestroyedBlackFunds(address indexed _blackListedUser, uint256 _balance)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterDestroyedBlackFunds(opts *bind.FilterOpts, _blackListedUser []common.Address) (*TrxUsdtTokenDestroyedBlackFundsIterator, error) {

	var _blackListedUserRule []interface{}
	for _, _blackListedUserItem := range _blackListedUser {
		_blackListedUserRule = append(_blackListedUserRule, _blackListedUserItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "DestroyedBlackFunds", _blackListedUserRule)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenDestroyedBlackFundsIterator{contract: _TrxUsdtToken.contract, event: "DestroyedBlackFunds", logs: logs, sub: sub}, nil
}

// WatchDestroyedBlackFunds is a free log subscription operation binding the contract event 0x61e6e66b0d6339b2980aecc6ccc0039736791f0ccde9ed512e789a7fbdd698c6.
//
// Solidity: event DestroyedBlackFunds(address indexed _blackListedUser, uint256 _balance)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchDestroyedBlackFunds(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenDestroyedBlackFunds, _blackListedUser []common.Address) (event.Subscription, error) {

	var _blackListedUserRule []interface{}
	for _, _blackListedUserItem := range _blackListedUser {
		_blackListedUserRule = append(_blackListedUserRule, _blackListedUserItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "DestroyedBlackFunds", _blackListedUserRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenDestroyedBlackFunds)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "DestroyedBlackFunds", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDestroyedBlackFunds is a log parse operation binding the contract event 0x61e6e66b0d6339b2980aecc6ccc0039736791f0ccde9ed512e789a7fbdd698c6.
//
// Solidity: event DestroyedBlackFunds(address indexed _blackListedUser, uint256 _balance)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseDestroyedBlackFunds(log types.Log) (*TrxUsdtTokenDestroyedBlackFunds, error) {
	event := new(TrxUsdtTokenDestroyedBlackFunds)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "DestroyedBlackFunds", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenIssueIterator is returned from FilterIssue and is used to iterate over the raw logs and unpacked data for Issue events raised by the TrxUsdtToken contract.
type TrxUsdtTokenIssueIterator struct {
	Event *TrxUsdtTokenIssue // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenIssueIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenIssue)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenIssue)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenIssueIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenIssueIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenIssue represents a Issue event raised by the TrxUsdtToken contract.
type TrxUsdtTokenIssue struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterIssue is a free log retrieval operation binding the contract event 0xcb8241adb0c3fdb35b70c24ce35c5eb0c17af7431c99f827d44a445ca624176a.
//
// Solidity: event Issue(uint256 amount)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterIssue(opts *bind.FilterOpts) (*TrxUsdtTokenIssueIterator, error) {

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Issue")
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenIssueIterator{contract: _TrxUsdtToken.contract, event: "Issue", logs: logs, sub: sub}, nil
}

// WatchIssue is a free log subscription operation binding the contract event 0xcb8241adb0c3fdb35b70c24ce35c5eb0c17af7431c99f827d44a445ca624176a.
//
// Solidity: event Issue(uint256 amount)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchIssue(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenIssue) (event.Subscription, error) {

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Issue")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenIssue)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Issue", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseIssue is a log parse operation binding the contract event 0xcb8241adb0c3fdb35b70c24ce35c5eb0c17af7431c99f827d44a445ca624176a.
//
// Solidity: event Issue(uint256 amount)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseIssue(log types.Log) (*TrxUsdtTokenIssue, error) {
	event := new(TrxUsdtTokenIssue)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Issue", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the TrxUsdtToken contract.
type TrxUsdtTokenOwnershipTransferredIterator struct {
	Event *TrxUsdtTokenOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenOwnershipTransferred represents a OwnershipTransferred event raised by the TrxUsdtToken contract.
type TrxUsdtTokenOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*TrxUsdtTokenOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenOwnershipTransferredIterator{contract: _TrxUsdtToken.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenOwnershipTransferred)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseOwnershipTransferred(log types.Log) (*TrxUsdtTokenOwnershipTransferred, error) {
	event := new(TrxUsdtTokenOwnershipTransferred)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenParamsIterator is returned from FilterParams and is used to iterate over the raw logs and unpacked data for Params events raised by the TrxUsdtToken contract.
type TrxUsdtTokenParamsIterator struct {
	Event *TrxUsdtTokenParams // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenParamsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenParams)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenParams)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenParamsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenParamsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenParams represents a Params event raised by the TrxUsdtToken contract.
type TrxUsdtTokenParams struct {
	FeeBasisPoints *big.Int
	MaxFee         *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterParams is a free log retrieval operation binding the contract event 0xb044a1e409eac5c48e5af22d4af52670dd1a99059537a78b31b48c6500a6354e.
//
// Solidity: event Params(uint256 feeBasisPoints, uint256 maxFee)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterParams(opts *bind.FilterOpts) (*TrxUsdtTokenParamsIterator, error) {

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Params")
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenParamsIterator{contract: _TrxUsdtToken.contract, event: "Params", logs: logs, sub: sub}, nil
}

// WatchParams is a free log subscription operation binding the contract event 0xb044a1e409eac5c48e5af22d4af52670dd1a99059537a78b31b48c6500a6354e.
//
// Solidity: event Params(uint256 feeBasisPoints, uint256 maxFee)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchParams(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenParams) (event.Subscription, error) {

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Params")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenParams)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Params", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseParams is a log parse operation binding the contract event 0xb044a1e409eac5c48e5af22d4af52670dd1a99059537a78b31b48c6500a6354e.
//
// Solidity: event Params(uint256 feeBasisPoints, uint256 maxFee)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseParams(log types.Log) (*TrxUsdtTokenParams, error) {
	event := new(TrxUsdtTokenParams)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Params", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenPauseIterator is returned from FilterPause and is used to iterate over the raw logs and unpacked data for Pause events raised by the TrxUsdtToken contract.
type TrxUsdtTokenPauseIterator struct {
	Event *TrxUsdtTokenPause // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenPauseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenPause)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenPause)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenPauseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenPauseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenPause represents a Pause event raised by the TrxUsdtToken contract.
type TrxUsdtTokenPause struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterPause is a free log retrieval operation binding the contract event 0x6985a02210a168e66602d3235cb6db0e70f92b3ba4d376a33c0f3d9434bff625.
//
// Solidity: event Pause()
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterPause(opts *bind.FilterOpts) (*TrxUsdtTokenPauseIterator, error) {

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Pause")
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenPauseIterator{contract: _TrxUsdtToken.contract, event: "Pause", logs: logs, sub: sub}, nil
}

// WatchPause is a free log subscription operation binding the contract event 0x6985a02210a168e66602d3235cb6db0e70f92b3ba4d376a33c0f3d9434bff625.
//
// Solidity: event Pause()
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchPause(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenPause) (event.Subscription, error) {

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Pause")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenPause)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Pause", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePause is a log parse operation binding the contract event 0x6985a02210a168e66602d3235cb6db0e70f92b3ba4d376a33c0f3d9434bff625.
//
// Solidity: event Pause()
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParsePause(log types.Log) (*TrxUsdtTokenPause, error) {
	event := new(TrxUsdtTokenPause)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Pause", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenRedeemIterator is returned from FilterRedeem and is used to iterate over the raw logs and unpacked data for Redeem events raised by the TrxUsdtToken contract.
type TrxUsdtTokenRedeemIterator struct {
	Event *TrxUsdtTokenRedeem // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenRedeemIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenRedeem)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenRedeem)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenRedeemIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenRedeemIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenRedeem represents a Redeem event raised by the TrxUsdtToken contract.
type TrxUsdtTokenRedeem struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRedeem is a free log retrieval operation binding the contract event 0x702d5967f45f6513a38ffc42d6ba9bf230bd40e8f53b16363c7eb4fd2deb9a44.
//
// Solidity: event Redeem(uint256 amount)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterRedeem(opts *bind.FilterOpts) (*TrxUsdtTokenRedeemIterator, error) {

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Redeem")
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenRedeemIterator{contract: _TrxUsdtToken.contract, event: "Redeem", logs: logs, sub: sub}, nil
}

// WatchRedeem is a free log subscription operation binding the contract event 0x702d5967f45f6513a38ffc42d6ba9bf230bd40e8f53b16363c7eb4fd2deb9a44.
//
// Solidity: event Redeem(uint256 amount)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchRedeem(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenRedeem) (event.Subscription, error) {

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Redeem")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenRedeem)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Redeem", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRedeem is a log parse operation binding the contract event 0x702d5967f45f6513a38ffc42d6ba9bf230bd40e8f53b16363c7eb4fd2deb9a44.
//
// Solidity: event Redeem(uint256 amount)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseRedeem(log types.Log) (*TrxUsdtTokenRedeem, error) {
	event := new(TrxUsdtTokenRedeem)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Redeem", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenRemovedBlackListIterator is returned from FilterRemovedBlackList and is used to iterate over the raw logs and unpacked data for RemovedBlackList events raised by the TrxUsdtToken contract.
type TrxUsdtTokenRemovedBlackListIterator struct {
	Event *TrxUsdtTokenRemovedBlackList // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenRemovedBlackListIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenRemovedBlackList)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenRemovedBlackList)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenRemovedBlackListIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenRemovedBlackListIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenRemovedBlackList represents a RemovedBlackList event raised by the TrxUsdtToken contract.
type TrxUsdtTokenRemovedBlackList struct {
	User common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRemovedBlackList is a free log retrieval operation binding the contract event 0xd7e9ec6e6ecd65492dce6bf513cd6867560d49544421d0783ddf06e76c24470c.
//
// Solidity: event RemovedBlackList(address indexed _user)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterRemovedBlackList(opts *bind.FilterOpts, _user []common.Address) (*TrxUsdtTokenRemovedBlackListIterator, error) {

	var _userRule []interface{}
	for _, _userItem := range _user {
		_userRule = append(_userRule, _userItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "RemovedBlackList", _userRule)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenRemovedBlackListIterator{contract: _TrxUsdtToken.contract, event: "RemovedBlackList", logs: logs, sub: sub}, nil
}

// WatchRemovedBlackList is a free log subscription operation binding the contract event 0xd7e9ec6e6ecd65492dce6bf513cd6867560d49544421d0783ddf06e76c24470c.
//
// Solidity: event RemovedBlackList(address indexed _user)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchRemovedBlackList(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenRemovedBlackList, _user []common.Address) (event.Subscription, error) {

	var _userRule []interface{}
	for _, _userItem := range _user {
		_userRule = append(_userRule, _userItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "RemovedBlackList", _userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenRemovedBlackList)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "RemovedBlackList", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRemovedBlackList is a log parse operation binding the contract event 0xd7e9ec6e6ecd65492dce6bf513cd6867560d49544421d0783ddf06e76c24470c.
//
// Solidity: event RemovedBlackList(address indexed _user)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseRemovedBlackList(log types.Log) (*TrxUsdtTokenRemovedBlackList, error) {
	event := new(TrxUsdtTokenRemovedBlackList)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "RemovedBlackList", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the TrxUsdtToken contract.
type TrxUsdtTokenTransferIterator struct {
	Event *TrxUsdtTokenTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenTransfer represents a Transfer event raised by the TrxUsdtToken contract.
type TrxUsdtTokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*TrxUsdtTokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenTransferIterator{contract: _TrxUsdtToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenTransfer)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseTransfer(log types.Log) (*TrxUsdtTokenTransfer, error) {
	event := new(TrxUsdtTokenTransfer)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TrxUsdtTokenUnpauseIterator is returned from FilterUnpause and is used to iterate over the raw logs and unpacked data for Unpause events raised by the TrxUsdtToken contract.
type TrxUsdtTokenUnpauseIterator struct {
	Event *TrxUsdtTokenUnpause // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TrxUsdtTokenUnpauseIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TrxUsdtTokenUnpause)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TrxUsdtTokenUnpause)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TrxUsdtTokenUnpauseIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TrxUsdtTokenUnpauseIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TrxUsdtTokenUnpause represents a Unpause event raised by the TrxUsdtToken contract.
type TrxUsdtTokenUnpause struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterUnpause is a free log retrieval operation binding the contract event 0x7805862f689e2f13df9f062ff482ad3ad112aca9e0847911ed832e158c525b33.
//
// Solidity: event Unpause()
func (_TrxUsdtToken *TrxUsdtTokenFilterer) FilterUnpause(opts *bind.FilterOpts) (*TrxUsdtTokenUnpauseIterator, error) {

	logs, sub, err := _TrxUsdtToken.contract.FilterLogs(opts, "Unpause")
	if err != nil {
		return nil, err
	}
	return &TrxUsdtTokenUnpauseIterator{contract: _TrxUsdtToken.contract, event: "Unpause", logs: logs, sub: sub}, nil
}

// WatchUnpause is a free log subscription operation binding the contract event 0x7805862f689e2f13df9f062ff482ad3ad112aca9e0847911ed832e158c525b33.
//
// Solidity: event Unpause()
func (_TrxUsdtToken *TrxUsdtTokenFilterer) WatchUnpause(opts *bind.WatchOpts, sink chan<- *TrxUsdtTokenUnpause) (event.Subscription, error) {

	logs, sub, err := _TrxUsdtToken.contract.WatchLogs(opts, "Unpause")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TrxUsdtTokenUnpause)
				if err := _TrxUsdtToken.contract.UnpackLog(event, "Unpause", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUnpause is a log parse operation binding the contract event 0x7805862f689e2f13df9f062ff482ad3ad112aca9e0847911ed832e158c525b33.
//
// Solidity: event Unpause()
func (_TrxUsdtToken *TrxUsdtTokenFilterer) ParseUnpause(log types.Log) (*TrxUsdtTokenUnpause, error) {
	event := new(TrxUsdtTokenUnpause)
	if err := _TrxUsdtToken.contract.UnpackLog(event, "Unpause", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
