package precompiled

import (
	"encoding/binary"
	"errors"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"strings"
	"web3_gui/chain/db"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
)

const ERC20_ABI = `[
	{
		"constant": false,
		"inputs": [
			{
				"name": "_spender",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "approve",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "burn",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "freeze",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_to",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "transfer",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_from",
				"type": "address"
			},
			{
				"name": "_to",
				"type": "address"
			},
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "transferFrom",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "_value",
				"type": "uint256"
			}
		],
		"name": "unfreeze",
		"outputs": [
			{
				"name": "success",
				"type": "bool"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "withdrawEther",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"name": "initialSupply",
				"type": "uint256"
			},
			{
				"name": "tokenName",
				"type": "string"
			},
			{
				"name": "decimalUnits",
				"type": "uint8"
			},
			{
				"name": "tokenSymbol",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"payable": true,
		"stateMutability": "payable",
		"type": "fallback"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "from",
				"type": "address"
			},
			{
				"indexed": true,
				"name": "to",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Transfer",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "from",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Burn",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "from",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Freeze",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"name": "from",
				"type": "address"
			},
			{
				"indexed": false,
				"name": "value",
				"type": "uint256"
			}
		],
		"name": "Unfreeze",
		"type": "event"
	},
	{
		"constant": true,
		"inputs": [
			{
				"name": "",
				"type": "address"
			},
			{
				"name": "",
				"type": "address"
			}
		],
		"name": "allowance",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"name": "",
				"type": "address"
			}
		],
		"name": "balanceOf",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [
			{
				"name": "",
				"type": "uint8"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [
			{
				"name": "",
				"type": "address"
			}
		],
		"name": "freezeOf",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "name",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "owner",
		"outputs": [
			{
				"name": "",
				"type": "address"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "totalSupply",
		"outputs": [
			{
				"name": "",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	}
]`

const ERC721_SupportsInterface = "0x01ffc9a780ac58cd00000000000000000000000000000000000000000000000000000000"
const ERC721_ABI = `[
{
"constant": true,
"inputs": [
{
"name": "interfaceID",
"type": "bytes4"
}
],
"name": "supportsInterface",
"outputs": [
{
"name": "",
"type": "bool"
}
],
"payable": false,
"stateMutability": "view",
"type": "function"
}
]`

const ERC1155_SupportsInterface = "0x01ffc9a7d9b67a2600000000000000000000000000000000000000000000000000000000"
const ERC1155_ABI = `[
{
"constant": true,
"inputs": [
{
"name": "interfaceID",
"type": "bytes4"
}
],
"name": "supportsInterface",
"outputs": [
{
"name": "",
"type": "bool"
}
],
"payable": false,
"stateMutability": "view",
"type": "function"
}
]`

// 获取转账编码后的参数
func BuildErc20TransferInput(to string, amount uint64) []byte {
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	_to := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(to)))
	input, _ := abiObj.Pack("transfer", _to, big.NewInt(int64(amount)))
	return input
}
func BuildErc20TransferBigInput(to string, amount *big.Int) []byte {
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	_to := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(to)))
	input, _ := abiObj.Pack("transfer", _to, amount)
	return input
}

// 查询代币余额
func GetBalance(from crypto.AddressCoin, addr, contractAddr string) uint64 {
	_addr := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(addr)))
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("balanceOf", _addr)
	vmRun := evm.NewVmRun(from, crypto.AddressFromB58String(contractAddr), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return 0
	}
	out, err := abiObj.Unpack("balanceOf", result.ResultData)
	if err != nil {
		return 0
	}
	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	return out0.Uint64()
}

// 查询代币余额
func GetBigBalance(from crypto.AddressCoin, addr, contractAddr string) *big.Int {
	_addr := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(addr)))
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("balanceOf", _addr)
	vmRun := evm.NewVmRun(from, crypto.AddressFromB58String(contractAddr), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return big.NewInt(0)
	}
	out, err := abiObj.Unpack("balanceOf", result.ResultData)
	if err != nil {
		return big.NewInt(0)
	}
	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	return out0
}

// 查询代币余额, 等同GetBigBalance,参数类型不同
func GetBigBalanceWithCryptoAddress(from, addr, contractAddr crypto.AddressCoin) *big.Int {
	_addr := common.Address(evmutils.AddressCoinToAddress(addr))
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("balanceOf", _addr)
	vmRun := evm.NewVmRun(from, contractAddr, []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return big.NewInt(0)
	}
	out, err := abiObj.Unpack("balanceOf", result.ResultData)
	if err != nil {
		return big.NewInt(0)
	}
	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	return out0
}

// 查询代币余额
func balanceOf(from, contractAddr string) (uint64, error) {
	_addr := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(from)))
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("balanceOf", _addr)
	vmRun := evm.NewVmRun(crypto.AddressFromB58String(from), crypto.AddressFromB58String(contractAddr), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return 0, err
	}
	out, err := abiObj.Unpack("balanceOf", result.ResultData)
	if err != nil {
		return 0, err
	}
	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	return out0.Uint64(), nil
}

// 查询代币名称
func getName(from, contractAddr string) (string, error) {
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("name")
	vmRun := evm.NewVmRun(crypto.AddressFromB58String(from), crypto.AddressFromB58String(contractAddr), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return "", err
	}
	out, err := abiObj.Unpack("name", result.ResultData)
	if err != nil {
		return "", err
	}
	out0 := *abi.ConvertType(out[0], new(string)).(*string)
	return out0, nil
}

// 查询代币标识
func getSymbol(from, contractAddr string) (string, error) {
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("symbol")
	vmRun := evm.NewVmRun(crypto.AddressFromB58String(from), crypto.AddressFromB58String(contractAddr), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return "", err
	}
	out, err := abiObj.Unpack("symbol", result.ResultData)
	if err != nil {
		return "", err
	}
	out0 := *abi.ConvertType(out[0], new(string)).(*string)
	return out0, nil
}

// 查询代币精度
func getDecimals(from, contractAddr string) (uint8, error) {
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("decimals")
	vmRun := evm.NewVmRun(crypto.AddressFromB58String(from), crypto.AddressFromB58String(contractAddr), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return 0, err
	}
	out, err := abiObj.Unpack("decimals", result.ResultData)
	if err != nil {
		return 0, err
	}
	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)
	return out0, nil
}

// 查询代币精度
func GetDecimals(from, contractAddr string) uint8 {
	d, _ := getDecimals(from, contractAddr)
	return d
}

// 获取发行量
func getTotalSupply(from, contractAddr string) (*big.Int, error) {
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	input, _ := abiObj.Pack("totalSupply")
	vmRun := evm.NewVmRun(crypto.AddressFromB58String(from), crypto.AddressFromB58String(contractAddr), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 9000000, 1, 0, false)
	if err != nil {
		return big.NewInt(0), err
	}
	out, err := abiObj.Unpack("totalSupply", result.ResultData)
	if err != nil {
		return big.NewInt(0), err
	}
	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	return out0, nil
}

// 检查是否erc20
func IsErc20(from, contractAddr string) (string, string, bool) {
	_, err := balanceOf(from, contractAddr)
	if err != nil {
		return "", "", false
	}
	name, err := getName(from, contractAddr)
	if err != nil {
		return "", "", false
	}
	symbol, err := getSymbol(from, contractAddr)
	if err != nil {
		return "", "", false
	}
	_, err = getDecimals(from, contractAddr)
	if err != nil {
		return "", "", false
	}
	_, err = getTotalSupply(from, contractAddr)
	if err != nil {
		return "", "", false
	}
	return name, symbol, true
}

func Uint64ToString(u uint64, d uint8) string {
	return new(big.Float).Quo(new(big.Float).SetUint64(u), big.NewFloat(math.Pow10(int(d)))).Text('f', -1)
}

func StringToUint64(s string, d uint8) (uint64, bool) {
	f, ok := new(big.Float).SetString(s)
	if !ok {
		return 0, false
	}

	r, _ := new(big.Float).Mul(f, big.NewFloat(math.Pow10(int(d)))).Uint64()
	return r, true
}

func ValueToString(v *big.Int, d uint8) string {
	return decimal.NewFromBigInt(v, 0-int32(d)).String()
}

func StringToValue(s string, d uint8) *big.Int {
	fs, err := decimal.NewFromString(s)
	if err != nil {
		return big.NewInt(0)
	}

	fs = fs.Mul(decimal.NewFromFloat(math.Pow10(int(d))))
	return fs.BigInt()
}

var errDecode = errors.New("token decode fail")

var tokenSep byte = ':'

/*
*
编码k,v
*/
func EncodeToken(name, symbol []byte) []byte {
	buf := make([]byte, len(name)+len(symbol)+2+1)
	pos := 0
	binary.BigEndian.PutUint16(buf[pos:], uint16(len(name)))
	pos += 2

	copy(buf[pos:], name)
	pos += len(name)

	buf[pos] = tokenSep
	pos++

	copy(buf[pos:], symbol)

	return buf
}

/*
*
解码k,v
*/
func DecodeToken(buf []byte) ([]byte, []byte, error) {
	pos := 0
	if pos+2 > len(buf) {
		return nil, nil, errDecode
	}

	prefixLen := int(binary.BigEndian.Uint16(buf[pos:]))
	pos += 2

	if prefixLen+pos > len(buf) {
		return nil, nil, errDecode
	}
	hash := buf[pos : pos+prefixLen]
	pos += prefixLen

	if buf[pos] != tokenSep {
		return nil, nil, errDecode
	}
	pos++
	sign := buf[pos:]
	return hash, sign, nil
}

func CheckErc20(vmRun *evm.VmRun, from, contractAddr crypto.AddressCoin) (db.Erc20Info, bool) {
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))

	gas := uint64(9000000)

	ms := make([]MethodCall, 0)
	ms = append(ms, MethodCall{"name", gas, true, nil})
	ms = append(ms, MethodCall{"symbol", gas, true, nil})
	ms = append(ms, MethodCall{"decimals", gas, true, nil})
	ms = append(ms, MethodCall{"totalSupply", gas, true, nil})
	_addr := common.Address(evmutils.AddressCoinToAddress(from))
	ms = append(ms, MethodCall{"balanceOf", gas, false, []interface{}{_addr}})

	ks := make(map[string]int, len(ms))
	for k, v := range ms {
		ks[v.method] = k
	}

	r, allok := multiCallFunc(vmRun, &abiObj, ms, false)
	if !allok {
		return db.Erc20Info{}, false
	}

	return db.Erc20Info{
		Address:     contractAddr.B58String(),
		From:        from.B58String(),
		Name:        r[ks["name"]][0].(string),
		Symbol:      r[ks["symbol"]][0].(string),
		Decimals:    r[ks["decimals"]][0].(uint8),
		TotalSupply: r[ks["totalSupply"]][0].(*big.Int),
	}, true
}

func CheckErc721(vmRun *evm.VmRun, from, contractAddr crypto.AddressCoin) (db.Erc721Info, bool) {
	gas := uint64(9000000)
	abiObj, _ := abi.JSON(strings.NewReader(ERC721_ABI))
	input := common.Hex2Bytes(ERC721_SupportsInterface)
	result, _, err := evm.Run(input, gas, 1, 0, false, vmRun)
	if err != nil {
		return db.Erc721Info{}, false
	}

	out, err := abiObj.Unpack("supportsInterface", result.ResultData)
	if err != nil || len(out) == 0 {
		return db.Erc721Info{}, false
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	if out0 {
		return db.Erc721Info{
			Name:        "",
			From:        from.B58String(),
			Address:     contractAddr.B58String(),
			TotalSupply: nil,
		}, true
	}

	return db.Erc721Info{}, false
}

func CheckErc1155(vmRun *evm.VmRun, from, contractAddr crypto.AddressCoin) (db.Erc1155Info, bool) {
	gas := uint64(9000000)
	abiObj, _ := abi.JSON(strings.NewReader(ERC1155_ABI))
	input := common.Hex2Bytes(ERC1155_SupportsInterface)
	result, _, err := evm.Run(input, gas, 1, 0, false, vmRun)
	if err != nil {
		return db.Erc1155Info{}, false
	}

	out, err := abiObj.Unpack("supportsInterface", result.ResultData)
	if err != nil || len(out) == 0 {
		return db.Erc1155Info{}, false
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	if out0 {
		return db.Erc1155Info{
			Name:        "",
			From:        from.B58String(),
			Address:     contractAddr.B58String(),
			TotalSupply: nil,
		}, true
	}

	return db.Erc1155Info{}, false
}

type MethodCall struct {
	method   string
	gas      uint64
	isResult bool
	args     []interface{}
}

func multiCallFunc(vmRun *evm.VmRun, abiObj *abi.ABI, ms []MethodCall, isAll bool) ([][]interface{}, bool) {
	r := make([][]interface{}, len(ms))
	allok := true
	for k, v := range ms {
		r[k] = nil
		input, _ := abiObj.Pack(v.method, v.args...)
		result, _, err := evm.Run(input, v.gas, 1, 0, false, vmRun)
		if err != nil {
			if !isAll {
				return nil, false
			}
			allok = false
			//continue
		}

		out, err := abiObj.Unpack(v.method, result.ResultData)
		if err != nil {
			if !isAll {
				return nil, false
			}
			allok = false
			//continue
		}

		if v.isResult {
			r[k] = out
		}
	}

	return r, allok
}

// 获取多个地址的总余额
func GetMultiAddrBalance(contractAddr crypto.AddressCoin, addrs []*keystore.AddressInfo) *big.Int {
	if len(addrs) == 0 {
		return big.NewInt(0)
	}

	vmRun := evm.NewVmRun(addrs[0].Addr, contractAddr, []byte("0x1"), nil)
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))

	gas := uint64(9000000)

	value := new(big.Int)
	for _, v := range addrs {
		_addr := common.Address(evmutils.AddressCoinToAddress(v.Addr))
		input, _ := abiObj.Pack("balanceOf", _addr)
		result, _, err := vmRun.Run(input, gas, 1, 0, false)
		if err != nil {
			continue
		}
		out, err := abiObj.Unpack("balanceOf", result.ResultData)
		if err != nil {
			continue
		}
		out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
		value.Add(value, out0)
	}

	return value
}

// 查询代币余额
func GetMultiBigBalance(contractAddr crypto.AddressCoin, addrs []crypto.AddressCoin) []*big.Int {
	if len(addrs) == 0 {
		return nil
	}

	vmRun := evm.NewVmRun(addrs[0], contractAddr, []byte("0x1"), nil)
	vmRun.SetStorage(nil)
	abiObj, _ := abi.JSON(strings.NewReader(ERC20_ABI))
	gas := uint64(9000000)

	r := make([]*big.Int, len(addrs))
	for k, v := range addrs {
		_addr := common.Address(evmutils.AddressCoinToAddress(v))
		input, _ := abiObj.Pack("balanceOf", _addr)
		result, _, err := evm.Run(input, gas, 1, 0, false, vmRun)
		if err != nil {
			r[k] = new(big.Int)
			continue
		}
		out, err := abiObj.Unpack("balanceOf", result.ResultData)
		if err != nil {
			r[k] = new(big.Int)
			continue
		}
		out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

		r[k] = out0
	}
	return r
}
