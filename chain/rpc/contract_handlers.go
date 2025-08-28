package rpc

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"net/http"
	"strings"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/vm/opcodes"

	"github.com/golang/protobuf/proto"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/chain/evm/precompiled/sol"
	"web3_gui/chain/mining"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

// 合约部署交易
func CreateContractByTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.CheckOutOfMemory() {
		return model.Errcode(model.Timeout)
	}

	config.GetRpcRate(rj.Method, true)

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}
	//部署合约时，value为0
	amount = 0
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	//var gasLimit = config.Wallet_tx_gas_min
	//if config.EVM_GAS_MAX > config.Wallet_tx_gas_min {
	//	gasLimit = config.EVM_GAS_MAX
	//}
	//if gas > gasLimit {
	//	res, err = model.Errcode(GasTooBig, "gas is too big")
	//	return
	//}
	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}

	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
	if strings.HasPrefix(comment, "0x") || strings.HasPrefix(comment, "0X") {
		res, err = model.Errcode(model.Nomarl, "comment should be del 0x prefix")
		return
	}
	source := ""
	sourceStr, ok := rj.Get("source")
	if ok && rj.VerifyType("source", "string") {
		source = sourceStr.(string)
	}

	abi := ""
	abiStr, ok := rj.Get("abi")
	if ok && rj.VerifyType("abi", "string") {
		abi = abiStr.(string)
	}

	defaultClass := config.NORMAL_CONTRACT
	classStr, ok := rj.Get("class")
	if ok {
		defaultClass = config.ContractClass(toUint64(classStr.(float64)))
		if !defaultClass.IsLegal() {
			res, err = model.Errcode(model.Nomarl, "contract class is illegality")
			return
		}
	}
	//runeLength := len([]rune(comment))
	////if runeLength > 1024 {
	////	res, err = model.Errcode(CommentOverLengthMax, "comment")
	////	return
	////}
	//temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(100000000))
	//temp = new(big.Int).Div(temp, big.NewInt(1024))
	//if gas < temp.Uint64() {
	//	res, err = model.Errcode(GasTooLittle, "gas")
	//	return
	//}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, nil, amount, gas, frozenHeight, pwd, comment, source, uint64(defaultClass), gasPrice, abi)

	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	result["contract_address"] = txpay.Vout[0].Address.B58String()
	res, err = model.Tojson(result)

	return res, err
}

// 合约调用交易
func CallContractByTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	//var gasLimit = config.Wallet_tx_gas_min
	//if config.EVM_GAS_MAX > config.Wallet_tx_gas_min {
	//	gasLimit = config.EVM_GAS_MAX
	//}
	//if gas > gasLimit {
	//	res, err = model.Errcode(GasTooBig, "gas is too big")
	//	return
	//}

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	//runeLength := len([]rune(comment))
	//if runeLength > 1024 {
	//	res, err = model.Errcode(CommentOverLengthMax, "comment")
	//	return
	//}
	//temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(100000000))
	//temp = new(big.Int).Div(temp, big.NewInt(1024))
	//if gas < temp.Uint64() {
	//	res, err = model.Errcode(GasTooLittle, "gas")
	//	return
	//}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 合约只读类型方法调用
func CallContractStack(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//rpc接口限速
	config.GetRpcRate(rj.Method, true)

	//直接在本地evm中执行即可
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}
	from := crypto.AddressFromB58String(addrItr.(string))
	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		//res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		//return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}
	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
	input := common.Hex2Bytes(comment)
	vmRun := evm.NewVmRun(from, contractAddr, []byte("0x1"), nil)
	result, _, evmErr := vmRun.Run(input, gas, config.DEFAULT_GAS_PRICE, 0, false)
	if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
		if result.ExitOpCode == opcodes.REVERT {
			unpackedMsg, unpackErr := abi.UnpackRevert(result.ResultData)
			if unpackErr != nil {
				res, err = model.Errcode(model.Nomarl, "执行合约失败,退出码:"+result.ExitOpCode.String()+unpackErr.Error())
				return
			}
			res, err = model.Errcode(model.Nomarl, "执行合约失败,退出码:"+result.ExitOpCode.String()+",失败原因:"+unpackedMsg)
			return
		}
		if evmErr != nil {
			res, err = model.Errcode(model.Nomarl, "执行合约失败,退出码:"+result.ExitOpCode.String()+evmErr.Error())
			return
		}
		res, err = model.Errcode(model.Nomarl, "执行合约失败")
		return
	}

	resHex := common.Bytes2Hex(result.ResultData)
	res, err = model.Tojson(resHex)
	return
}
func GetContractSourceV2(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}
	zipBytes, err := mining.GetContractSource(contractAddr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	var sourceBytes []byte
	if len(zipBytes) > 0 {
		sourceBytes, err = common.UnzipBytes(zipBytes)
		if err != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
	}
	data := make(map[string]interface{})
	data["source"] = string(sourceBytes)
	res, err = model.Tojson(data)
	return
}

func GetContractSource(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}
	zipBytes, err := mining.GetContractSource(contractAddr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	var sourceBytes []byte
	if len(zipBytes) > 0 {
		sourceBytes, err = common.UnzipBytes(zipBytes)
		if err != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
	}
	data := make(map[string]interface{})
	data["source"] = hex.EncodeToString(sourceBytes)
	res, err = model.Tojson(data)
	return
}

// 获取ERC20合约信息
func GetErc20Info(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	info := db.GetErc20Info(addr)
	if info.Address == "" {
		res, err = model.Errcode(model.NotExist, "address")
		return
	}

	type erc20Info struct {
		Address     string `json:"address"`
		From        string `json:"from"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Decimals    uint8  `json:"decimals"`
		TotalSupply uint64 `json:"totalSupply"`
	}

	data := erc20Info{
		Address:     info.Address,
		From:        info.From,
		Name:        info.Name,
		Symbol:      info.Symbol,
		Decimals:    info.Decimals,
		TotalSupply: info.TotalSupply.Uint64(),
	}

	res, err = model.Tojson(data)
	return
}

// 获取合约信息
func GetContractInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}
	//获取合约状态,0无效,1正常，2自毁合约
	contract := struct {
		IsContract      bool   `json:"is_contract"`
		ContractStatus  uint   `json:"contract_status"`
		ContractCode    string `json:"contract_code"`
		Name            string `json:"name"`
		CompilerVersion string `json:"compiler_version"`
		Abi             string `json:"abi"`
		Source          string `json:"source"`
		Bin             string `json:"bin"`
	}{}
	objectKey := append([]byte(config.DBKEY_CONTRACT_OBJECT), contractAddr...)
	objectValue, err := db.LevelTempDB.Find(objectKey)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if len(*objectValue) == 0 {
		contract.IsContract = false
		res, err = model.Tojson(contract)
		return
	}
	obj := new(go_protos.StateObject)
	proto.Unmarshal(*objectValue, obj)
	if len(obj.Code) == 0 {
		contract.IsContract = false
		res, err = model.Tojson(contract)
		return
	}
	contract.IsContract = true
	contract.ContractStatus = 1
	contract.ContractCode = hex.EncodeToString(obj.Code)
	if obj.Suicide {
		contract.ContractStatus = 2
	}
	if bytes.Equal(contractAddr, precompiled.RewardContract) {
		contract.Name = "blockReward"
		contract.CompilerVersion = "0.4.25+commit.59dbf8f1"
		contract.Source = sol.SOURCE_SOL
		contract.Abi = precompiled.REWARD_RATE_ABI
		contract.Bin = precompiled.REWARD_RATE_BIN
	} else {
		// 获取source
		zipBytes, err1 := mining.GetContractSource(contractAddr)
		if err1 != nil {
			res, err1 = model.Errcode(model.Nomarl, err.Error())
			return
		}
		var sourceBytes []byte
		if len(zipBytes) > 0 {
			sourceBytes, err = common.UnzipBytes(zipBytes)
			if err != nil {
				res, err = model.Errcode(model.Nomarl, err.Error())
				return
			}
		}
		contract.Source = hex.EncodeToString(sourceBytes)

		// 获取abi
		zipAbiBytes, errAbi := mining.GetContractAbi(contractAddr)
		if errAbi != nil {
			res, errAbi = model.Errcode(model.Nomarl, errAbi.Error())
			return
		}
		var abiBytes []byte
		if len(zipAbiBytes) > 0 {
			abiBytes, errAbi = common.UnzipBytes(zipAbiBytes)
			if errAbi != nil {
				res, errAbi = model.Errcode(model.Nomarl, err.Error())
				return
			}
		}
		contract.Abi = string(abiBytes)

		// 获取bin
		zipBinBytes, errBin := mining.GetContractBin(contractAddr)
		if errBin != nil {
			res, errBin = model.Errcode(model.Nomarl, errBin.Error())
			return
		}
		var binBytes string
		if len(zipBinBytes) > 0 {
			binBytes = common.Bytes2Hex(zipBinBytes)
		}
		contract.Bin = binBytes

	}
	res, err = model.Tojson(contract)
	return
}

// 获取地址特征,(1=主链币地址，2=合约地址，3=ERC20地址)
func GetAddrFeature(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addrStr := addrItr.(string)

	out := struct {
		Name    string `json:"name,omitempty"`    //未知地址 主链地址 合约地址 代币地址
		Feature int    `json:"feature,omitempty"` //0=未知 1=主链地址 2=合约地址 3=代币地币
	}{
		Name:    "未知地址",
		Feature: 0,
	}

	addr := crypto.AddressFromB58String(addrStr)
	if !crypto.ValidAddr(config.AddrPre, addr) {
		res, err = model.Tojson(out)
		return
	}

	res, err = model.Tojson(out)
	objectKey := append([]byte(config.DBKEY_CONTRACT_OBJECT), addr...)
	objectValue, err := db.LevelTempDB.Find(objectKey)
	if err == nil && len(*objectValue) != 0 {
		obj := new(go_protos.StateObject)
		proto.Unmarshal(*objectValue, obj)
		if len(obj.Code) != 0 {
			out.Name = "合约地址"
			out.Feature = 2
			res, err = model.Tojson(out)
			return
		}
	}

	////是否erc20
	//erc20Info := db.GetErc20Info(addrStr)
	//if erc20Info.Address != "" {
	//	out.Name = "ERC20地址"
	//	out.Feature = 3
	//	res, err = model.Tojson(out)
	//	return
	//}

	//其它主链币地址
	out.Name = "主链币地址"
	out.Feature = 1
	res, err = model.Tojson(out)
	return
}

func GetSpecialContract(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	address := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, address) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}
	//获取类型字段
	classItr, ok := rj.Get("class")
	if !ok {
		res, err = model.Errcode(model.NoField, "class")
		return
	}
	class := toUint64(classItr.(float64))
	value, err := mining.GetSpecialContract(address, class)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["contract_addr"] = value.B58String()
	res, err = model.Tojson(data)
	return
}

func CheckContract(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "address")
		return
	}
	objectKey := append([]byte(config.DBKEY_CONTRACT_OBJECT), contractAddr...)
	objectValue, err := db.LevelTempDB.Find(objectKey)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	var checkRes = false
	if len(*objectValue) == 0 {
		//contract.IsContract = false
		res, err = model.Tojson("address is not contract")
		return
	}
	obj := new(go_protos.StateObject)
	proto.Unmarshal(*objectValue, obj)
	code := hex.EncodeToString(obj.Code)
	bin := ""
	binItr, ok := rj.Get("bin")
	if ok && rj.VerifyType("bin", "string") {
		bin = binItr.(string)
	}
	input := common.Hex2Bytes(bin)
	vmRun := evm.NewVmRun(config.Area.Keystore.GetCoinbase().Addr, contractAddr, []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, 90000000, config.DEFAULT_GAS_PRICE, 0, true)
	//result, err := evm.Run(from, contractAddr, []byte("0x1"), input, gas, 0, nil, false)

	if err != nil {
		res, err = model.Errcode(model.NoField, "合约调用失败:"+err.Error())
		return
	}
	resHex := common.Bytes2Hex(result.ResultData)
	if code == resHex {
		checkRes = true
	}
	res, err = model.Tojson(checkRes)
	return
}

// 设置域名管理员为某注册器合约
func SetDomainManger(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	//节点名称
	nameItr, ok := rj.Get("name")
	if !ok {
		res, err = model.Errcode(model.NoField, "name")
		return
	}
	name := nameItr.(string)
	//注册器地址
	registarItr, ok := rj.Get("registar")
	if !ok {
		res, err = model.Errcode(model.NoField, "registar")
		return
	}
	registar := registarItr.(string)
	comment := ""
	//comment = ens.BuildNodeOwnerInput(name, registar)
	//comment = ens.BuildSubNodeOwnerInput("", name, registar)
	if name == "" {
		comment = ens.BuildNodeOwnerInput(name, registar)
	} else {
		comment = ens.BuildSubNodeRecordInput("", name, registar)
	}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 注册域名
func RegisterDomain(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	//节点名称
	nameItr, ok := rj.Get("name")
	if !ok {
		res, err = model.Errcode(model.NoField, "name")
		return
	}
	name := nameItr.(string)
	forever := false
	foreverItr, ok := rj.Get("forever")
	if !ok {
		res, err = model.Errcode(model.NoField, "forever")
		return
	}
	forever = foreverItr.(bool)

	//如果合约地址是平台基础注册器，则域名为根域名注册，默认为永久注册
	baseRegistarAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.BASE_REGISTAR_ADDR).Bytes())
	if addr == baseRegistarAddr.B58String() {
		forever = true
	}

	duration := toUint64(0)
	durationItr, ok := rj.Get("duration")
	if !ok {
		res, err = model.Errcode(model.NoField, "duration")
		return
	}
	duration = toUint64(durationItr.(float64))
	if duration < 365*24*60*60 {
		res, err = model.Errcode(model.Nomarl, "min duration is 31536000")
		return
	}
	if !forever && duration%31536000 != 0 {
		res, err = model.Errcode(model.Nomarl, "duration must multiple 31536000")
		return
	}
	//如果是永久，时间设置为1万年
	if forever {
		duration = 31536000 * 10000
	}
	//secret := ""
	//secretItr, ok := rj.Get("secret")
	//if !ok {
	//	res, err = model.Errcode(model.NoField, "secret")
	//	return
	//}
	//secret = secretItr.(string)
	//secretBt := common.Hex2Bytes(secret)
	//if len(secretBt) != 32 {
	//	res, err = model.Errcode(model.Nomarl, "secret must be 32 byte")
	//	return
	//}
	//secret32 := [32]byte{}
	//for index, v := range secretBt {
	//	secret32[index] = v
	//}
	//持有人
	ownerItr, ok := rj.Get("owner")
	if !ok {
		res, err = model.Errcode(model.NoField, "owner")
		return
	}
	owner := ownerItr.(string)
	comment := ens.BuildRegisterInput(name, owner, big.NewInt(int64(duration)))

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 根域名持有人和平台持有人提现
func DomainWithDraw(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	//amountItr, ok := rj.Get("amount")
	//if !ok {
	//	res, err = model.Errcode(model.NoField, "amount")
	//	return
	//}
	//amount := toUint64(amountItr.(float64))

	//comment := ens.BuildWithDrawInput(new(big.Int).SetUint64(amount))
	comment := ens.BuildWithDrawInput()

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 根域名持有人和平台持有人提现
func DomainTransfer(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)
	toItr, ok := rj.Get("to")
	if !ok {
		res, err = model.Errcode(model.NoField, "to")
		return
	}
	to := toItr.(string)
	comment := ens.BuildTransferInput(to)

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

// 续费域名
func ReNewDomain(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	frozenHeight := toUint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = toUint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	//节点名称
	nameItr, ok := rj.Get("name")
	if !ok {
		res, err = model.Errcode(model.NoField, "name")
		return
	}
	name := nameItr.(string)
	forever := false
	foreverItr, ok := rj.Get("forever")
	if !ok {
		res, err = model.Errcode(model.NoField, "forever")
		return
	}
	forever = foreverItr.(bool)
	duration := toUint64(0)
	durationItr, ok := rj.Get("duration")
	if !ok {
		res, err = model.Errcode(model.NoField, "duration")
		return
	}
	duration = toUint64(durationItr.(float64))
	if duration < 365*24*60*60 {
		res, err = model.Errcode(model.Nomarl, "min duration is 31536000")
		return
	}
	if !forever && duration%31536000 != 0 {
		res, err = model.Errcode(model.Nomarl, "duration must multiple 31536000")
		return
	}
	//如果是永久，时间设置为1万年
	if forever {
		duration = 31536000 * 10000
	}
	////检测注册
	//moneyRes, nameRes := ens.CheckRegister(src.B58String(), addr, name, big.NewInt(int64(amount)), big.NewInt(int64(duration)))
	//if !moneyRes {
	//	res, err = model.Errcode(model.Nomarl, "amount is less")
	//	return
	//}
	//if !nameRes {
	//	res, err = model.Errcode(model.Nomarl, "name is not valid")
	//	return
	//}

	comment := ens.BuildReNewInput(name, big.NewInt(int64(duration)))

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		res, err = model.Errcode(BalanceNotEnough)
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}

func GetRewardContract(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	res, err = model.Tojson(precompiled.RewardContract.B58String())
	return res, err
}

/*
合约预执行
*/
func PreCallContract(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//直接在本地evm中执行即可
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}
	from := crypto.AddressFromB58String(addrItr.(string))
	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		res, err = model.Errcode(model.NoField, "contractaddress")
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if len(contractAddr) > 0 && !crypto.ValidAddr(config.AddrPre, contractAddr) {
		res, err = model.Errcode(ContentIncorrectFormat, "contractaddress")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}
	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
	input := common.Hex2Bytes(comment)
	vmRun := evm.NewVmRun(from, contractAddr, []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, gas, config.DEFAULT_GAS_PRICE, amount, len(contractAddr) == 0)
	//result, err := evm.Run(from, contractAddr, []byte("0x1"), input, gas, 0, nil, false)

	if result.ExitOpCode == opcodes.REVERT {
		msg, err := abi.UnpackRevert(result.ResultData)
		if err != nil {
			return model.Tojson(map[string]interface{}{"gasUsed": 0, "errMsg": err.Error()})
		}
		return model.Tojson(map[string]interface{}{"gasUsed": 0, "errMsg": msg})
	}
	if err != nil {
		return model.Tojson(map[string]interface{}{"gasUsed": 0, "errMsg": err.Error()})
	}

	return model.Tojson(map[string]interface{}{"gasUsed": gas - result.GasLeft, "errMsg": ""})
}
