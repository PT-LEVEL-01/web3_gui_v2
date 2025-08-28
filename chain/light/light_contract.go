package light

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"math/big"
	"strings"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/sol"
	"web3_gui/chain/mining"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/chain/rpc"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

func RegisterContractMsg() {
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX, CreateContractByTx) //创建合约
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX, CallContractByTx)     //调用合约
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_CALLCONTRACTSTACK, CallContractStack)   //本地模拟调用
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCONTRACTCOUNT, GetContractCount)     //查询有效合约的数量
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCONTRACTSOURCE, GetContractSource)   //获取合约源代码
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCONTRACTINFO, GetContractInfo)       //获取合约状态
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETSPECIALCONTRACT, GetSpecialContract) //获取特定类型合约地址
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_CHECKCONTRACT, CheckContract)           //验证合约
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_WITHDRAWREWARD, WithDrawReward)         //提现奖励
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETREWARDHISTORY, GetRewardHistory)     //获取奖励历史
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_RECEIVETCOIN, ReceiveTCoin)             //领取测试币
}

/*
创建合约
*/
func CreateContractByTx(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	if mining.CheckOutOfMemory() {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.Timeout, ""))
	}

	config.GetRpcRate(rj.Method, true)

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
		}
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.NoField, "amount"))
		return
	}
	amount := uint64(amountItr.(float64))
	if amount < 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(rpc.AmountIsZero, "amount"))
		return
	}
	//部署合约时，value为0
	amount = 0
	gasItr, ok := rj.Get("gas")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.NoField, "gas"))
		return
	}
	gas := uint64(gasItr.(float64))
	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = uint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.Nomarl, "gas_price is too low"))
			return
		}
	}

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.NoField, "pwd"))
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
	if strings.HasPrefix(comment, "0x") || strings.HasPrefix(comment, "0X") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.Nomarl, "comment should be del 0x prefix"))
		return
	}
	source := ""
	sourceStr, ok := rj.Get("source")
	if ok && rj.VerifyType("source", "string") {
		source = sourceStr.(string)
	}
	defaultClass := config.NORMAL_CONTRACT
	classStr, ok := rj.Get("class")
	if ok {
		defaultClass = config.ContractClass(uint64(classStr.(float64)))
		if !defaultClass.IsLegal() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.Nomarl, "contract class is illegality"))
			return
		}
	}
	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(rpc.BalanceNotEnough, ""))
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, nil, amount, gas, frozenHeight, pwd, comment, source, uint64(defaultClass), gasPrice, "")

	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.FailPwd, ""))
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(rpc.AmountIsZero, "amount"))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.Nomarl, err.Error()))
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	result["contract_address"] = txpay.Vout[0].Address.B58String()
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CREATECONTRACTBYTX_REV, pkg(model.Success, result))
}

/*
调用合约
*/
func CallContractByTx(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(rpc.ContentIncorrectFormat, "srcaddress"))
				return
			}
		}
	}

	addrItr, ok = rj.Get("contractaddress")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.NoField, "contractaddress"))
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(rpc.ContentIncorrectFormat, "contractaddress"))
		return
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.NoField, "amount"))
		return
	}
	amount := uint64(amountItr.(float64))
	if amount < 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(rpc.AmountIsZero, "amount"))
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.NoField, "gas"))
		return
	}
	gas := uint64(gasItr.(float64))
	gasPrice := uint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = uint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.Nomarl, "gas_price is too low"))
			return
		}
	}
	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.NoField, "pwd"))
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}
	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas*gasPrice)
	if total < amount+gas*gasPrice {
		//资金不够
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(rpc.BalanceNotEnough, ""))
		return
	}
	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, amount, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.FailPwd, ""))
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(rpc.AmountIsZero, "amount"))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.Nomarl, err.Error()))
		return
	}

	result, err := utils.ChangeMap(txpay)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTBYTX_REV, pkg(model.Success, result))
}

/*
本地模拟调用
*/
func CallContractStack(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTSTACK_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	comment, _ := rj.Get("comment")
	from, _ := rj.Get("addrItr")
	contractAddr, _ := rj.Get("contractAddr")
	gas, _ := rj.Get("gas")
	input := common.Hex2Bytes(comment.(string))

	engine.Log.Error("from:", crypto.AddressFromB58String(from.(string)))
	engine.Log.Error("contractAddr:", crypto.AddressFromB58String(contractAddr.(string)))

	vmRun := evm.NewVmRun(crypto.AddressFromB58String(from.(string)), crypto.AddressFromB58String(contractAddr.(string)), []byte("0x1"), nil)
	result, _, err := vmRun.Run(input, uint64(gas.(float64)), config.DEFAULT_GAS_PRICE, 0, false)

	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTSTACK_REV, pkg(model.Nomarl, "合约调用失败:"+err.Error()))
		return
	}
	resHex := common.Bytes2Hex(result.ResultData)
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CALLCONTRACTSTACK_REV, pkg(model.Success, resHex))
}

/*
查询有效合约的数量
*/
func GetContractCount(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTCOUNT_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	count := db.GetContractCount()
	data := make(map[string]interface{})
	data["count"] = count.String()
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTCOUNT_REV, pkg(model.Success, data))
}

/*
获取合约源代码
*/
func GetContractSource(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTSOURCE_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrItr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTSOURCE_REV, pkg(model.NoField, "address"))
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTSOURCE_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}
	zipBytes, err := mining.GetContractSource(contractAddr)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTSOURCE_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	var sourceBytes []byte
	if len(zipBytes) > 0 {
		sourceBytes, err = common.UnzipBytes(zipBytes)
		if err != nil {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTSOURCE_REV, pkg(model.Nomarl, err.Error()))
			return
		}
	}
	data := make(map[string]interface{})
	data["source"] = string(sourceBytes)
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTSOURCE_REV, pkg(model.Success, data))
}

/*
获取合约状态
*/
func GetContractInfo(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrItr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.NoField, "address"))
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(rpc.ContentIncorrectFormat, "address"))
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
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	if len(*objectValue) == 0 {
		contract.IsContract = false
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Success, contract))
		return
	}
	obj := new(go_protos.StateObject)
	proto.Unmarshal(*objectValue, obj)
	if len(obj.Code) == 0 {
		contract.IsContract = false
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Success, contract))
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
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Nomarl, err.Error()))
			return
		}
		var sourceBytes []byte
		if len(zipBytes) > 0 {
			sourceBytes, err = common.UnzipBytes(zipBytes)
			if err != nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Nomarl, err.Error()))
				return
			}
		}
		contract.Source = string(sourceBytes)

		// 获取abi
		zipAbiBytes, errAbi := mining.GetContractAbi(contractAddr)
		if errAbi != nil {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Nomarl, errAbi.Error()))
			return
		}
		var abiBytes []byte
		if len(zipAbiBytes) > 0 {
			abiBytes, errAbi = common.UnzipBytes(zipAbiBytes)
			if errAbi != nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Nomarl, err.Error()))
				return
			}
		}
		contract.Abi = string(abiBytes)

		// 获取bin
		zipBinBytes, errBin := mining.GetContractBin(contractAddr)
		if errBin != nil {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Nomarl, errBin.Error()))
			return
		}
		var binBytes string
		if len(zipBinBytes) > 0 {
			binBytes = common.Bytes2Hex(zipBinBytes)
		}
		contract.Bin = binBytes

	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCONTRACTINFO_REV, pkg(model.Success, contract))
}

/*
获取特定类型合约地址
*/
func GetSpecialContract(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSPECIALCONTRACT_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrItr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSPECIALCONTRACT_REV, pkg(model.NoField, "address"))
		return
	}
	addr := addrItr.(string)

	address := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, address) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSPECIALCONTRACT_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}
	//获取类型字段
	classItr, ok := rj.Get("class")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSPECIALCONTRACT_REV, pkg(model.NoField, "class"))
		return
	}
	class := uint64(classItr.(float64))
	value, err := mining.GetSpecialContract(address, class)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSPECIALCONTRACT_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	data := make(map[string]interface{})
	data["contract_addr"] = value.B58String()
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETSPECIALCONTRACT_REV, pkg(model.Success, data))
}

/*
验证合约
*/
func CheckContract(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECKCONTRACT_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrItr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECKCONTRACT_REV, pkg(model.NoField, "address"))
		return
	}
	addr := addrItr.(string)

	contractAddr := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, contractAddr) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECKCONTRACT_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}
	objectKey := append([]byte(config.DBKEY_CONTRACT_OBJECT), contractAddr...)
	objectValue, err := db.LevelTempDB.Find(objectKey)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECKCONTRACT_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	var checkRes = false
	if len(*objectValue) == 0 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECKCONTRACT_REV, pkg(model.Nomarl, "address is not contract"))
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

	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECKCONTRACT_REV, pkg(model.NoField, "合约调用失败:"+err.Error()))
		return
	}
	resHex := common.Bytes2Hex(result.ResultData)
	if code == resHex {
		checkRes = true
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_CHECKCONTRACT_REV, pkg(model.Success, checkRes))
}

/*
提现奖励
*/

func WithDrawReward(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	comment := precompiled.BuildCommunityDistributeInput()
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_WITHDRAWREWARD_REV, pkg(model.Success, hex.EncodeToString(comment)))
}

/*
获取奖励历史
*/
func GetRewardHistory(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(model.NoField, "startHeight"))
		return
	}
	startHeight := uint64(startHeightItr.(float64))
	if startHeight <= 1 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(model.Nomarl, "起始高度至少大于1"))
		return
	}
	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(model.NoField, "endHeight"))
		return
	}
	endHeight := uint64(endHeightItr.(float64))

	if endHeight < startHeight {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(model.NoField, "endHeight"))
		return
	}
	page, ok := rj.Get("page")
	if !ok {
		page = float64(1)
	}
	pageInt := int(page.(float64))

	pageSize, ok := rj.Get("page_size")
	if !ok {
		pageSize = float64(0)
	}
	pageSizeInt := int(pageSize.(float64))
	rewardLogs := []precompiled.LogRewardHistoryV0{}
	//待返回的区块
	for i := startHeight; i <= endHeight; i++ {

		bh := mining.LoadBlockHeadByHeight(i)
		if bh == nil {
			break
		}
		if len(bh.Tx) > 0 {
			rewardTx, e := mining.LoadTxBase(bh.Tx[0])
			if e != nil {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(model.Nomarl, e.Error()))
				return
			}
			if rewardTx.Class() == config.Wallet_tx_type_mining {
				logs := precompiled.GetRewardHistoryLog(i, *rewardTx.GetHash())
				rewardLogs = append(rewardLogs, logs...)
			}

		}

	}
	if pageSizeInt == 0 {
		pageSizeInt = len(rewardLogs)
	}
	data := make(map[string]interface{})
	start := (pageInt - 1) * pageSizeInt
	end := start + pageSizeInt
	if start > len(rewardLogs) {
		data["total"] = 0
		data["list"] = []precompiled.LogRewardHistoryV0{}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(model.Success, data))
		return
	}
	if end > len(rewardLogs) {
		end = len(rewardLogs)
	}
	newList := []precompiled.LogRewardHistoryV0{}
	for _, v := range rewardLogs[start:end] {
		if v.From == "" {

			addrCoin := crypto.AddressFromB58String(v.Into)
			witnessName := mining.FindWitnessName(addrCoin)
			v.Name = witnessName
		}
		newList = append(newList, v)
	}
	data["total"] = len(rewardLogs)
	data["list"] = newList
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETREWARDHISTORY_REV, pkg(model.Success, data))
}

/*
领取测试币
*/
func ReceiveTCoin(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//Area.NodeManager.NodeSelf.IdInfo.Id.B58String()
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	if config.AddrPre != "TEST" {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.NotTestChain, "the rpc interface only call in test chain"))
		return
	}
	src := config.Area.Keystore.GetCoinbase().Addr

	gas := uint64(config.Wallet_tx_gas_min)
	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}
	addrItr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(model.NoField, "address"))
		return
	}
	addr := addrItr.(string)
	dst := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}
	pwd := config.Wallet_keystore_default_pwd
	comment := "发送测试币"

	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.CommentOverLengthMax, "comment"))
		return
	}

	temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(int64(config.Wallet_tx_gas_min)))
	temp = new(big.Int).Div(temp, big.NewInt(1024))
	if gas < temp.Uint64() {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.GasTooLittle, "gas"))
		return
	}
	amount := uint64(config.FAUCET_COIN)

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas)
	if total < amount+gas {
		//资金不够
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.BalanceNotEnough, ""))
		return
	}
	t1 := mining.GetFaucetTime(addr)
	if !config.TimeNow().After(time.Unix(t1+24*60*60, 0)) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.TestCoinLimit, "lock time has not expired. Please try again later"))
		return
	}
	// 获取domain
	domain := ""
	domainItr, ok := rj.Get("domain")
	if ok && rj.VerifyType("domain", "string") {
		domain = domainItr.(string)
	}

	// 获取domainType
	domainType := uint64(0)
	domainTypeItr, ok := rj.Get("domain_type")
	if ok {
		domainType = uint64(domainTypeItr.(float64))
	}

	txpay, err := mining.SendToAddress(&src, &dst, amount, gas, frozenHeight, pwd, comment, domain, domainType)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(model.FailPwd, ""))
			return
		}
		if err.Error() == config.ERROR_amount_zero.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(rpc.AmountIsZero, "amount"))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	err = mining.SetFaucetTime(addr)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	result, err := utils.ChangeMap(txpay)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_RECEIVETCOIN_REV, pkg(model.Success, result))
}
