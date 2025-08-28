package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"strings"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/sol"
	"web3_gui/chain/evm/vm/opcodes"
	"web3_gui/chain/mining"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/chain_boot/chain_plus"
	chainbootconfig "web3_gui/chain_boot/config"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	keyconfig "web3_gui/keystore/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func ContractCreate(params *map[string]interface{}, dataStr, abi, source, srcaddress string, class, gas, gasPrice uint64,
	pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()

	pr := engine.NewPostResult()

	if mining.CheckOutOfMemory() {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_Memory_not_enough, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//config.GetRpcRate(rj.Method, true)

	var src crypto.AddressCoin
	if srcaddress != "" {
		srcAddrItr := (*params)["srcaddress1"]
		src = crypto.AddressCoin(srcAddrItr.(coin_address.AddressCoin))
	}

	if gasPrice < config.DEFAULT_GAS_PRICE {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_too_low, "gasPrice")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	if strings.HasPrefix(dataStr, "0x") || strings.HasPrefix(dataStr, "0X") {
		ERR := utils.NewErrorBus(engine.ERROR_code_rpc_method_fail, "dataStr")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	defaultClass := config.NORMAL_CONTRACT
	if _, ok := (*params)["class"]; ok {
		defaultClass = config.ContractClass(class)
	}
	if !defaultClass.IsLegal() {
		ERR := utils.NewErrorBus(engine.ERROR_code_rpc_param_value_overstep, "class")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//判断矿工费是否足够
	//if gas < chainconfig.Wallet_tx_gas_min {
	//	ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	//	pr.Code = ERR.Code
	//	pr.Msg = ERR.Msg
	//	return *pr
	//}

	total, _ := chain.Balance.BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	/*------------------------*/
	txpay, err := mining.ContractTx(&src, nil, 0, gas, 0, pwd, dataStr, source, uint64(defaultClass), gasPrice, abi)

	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			ERR := utils.NewErrorBus(keyconfig.ERROR_code_coinAddr_password_fail, "")
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	pr.Data["hash"] = hex.EncodeToString(*txpay.GetHash())
	pr.Data["contract_address"] = txpay.Vout[0].Address.B58String()
	pr.Data["tx"] = txpay
	return *pr
}

func ContractPushTxProto64(params *map[string]interface{}, base64StdStr string, checkBalance bool) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	txItr, ERR := chain_plus.CheckTxBase64(base64StdStr, checkBalance)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	tx := txItr.(*mining.Tx_Contract)
	if len(tx.Payload) < 4 {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_contract_Payload_size_fail, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	fee := tx.GetSpend()
	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(tx.Vin[0].GetPukToAddr(), fee)
	if total < fee {
		//资金不够
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//验证合约
	gasUsed, err := tx.PreExecNew()
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	tx.GasUsed = gasUsed
	tx.BuildHash()
	err = mining.AddTx(txItr)
	if err != nil {
		pr.Code = utils.ERROR_CODE_system_error_self
		pr.Msg = err.Error()
		return *pr
	}
	return *pr
}

func ContractInfo(params *map[string]interface{}, address string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	utils.Log.Info().Str("address", address).Send()
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)
	contractAddr := crypto.AddressCoin(addr)
	utils.Log.Info().Str("address", contractAddr.B58String()).Send()

	countBig := db.GetContractCount()
	total := countBig.String()
	utils.Log.Info().Str("total", total).Send()

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
	utils.Log.Info().Hex("保存合约地址", contractAddr).Hex("保存合约", objectKey).Send()
	objectValue, err := db.LevelTempDB.Find(objectKey)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	if len(*objectValue) == 0 {
		utils.Log.Info().Str("address", address).Send()
		contract.IsContract = false
		pr.Data["contract"] = contract
		return *pr
	}
	obj := new(go_protos.StateObject)
	proto.Unmarshal(*objectValue, obj)
	if len(obj.Code) == 0 {
		utils.Log.Info().Str("address", address).Send()
		contract.IsContract = false
		pr.Data["contract"] = contract
		return *pr
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
			ERR := utils.NewErrorSysSelf(err1)
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		var sourceBytes []byte
		if len(zipBytes) > 0 {
			sourceBytes, err = common.UnzipBytes(zipBytes)
			if err != nil {
				ERR := utils.NewErrorSysSelf(err)
				pr.Code = ERR.Code
				pr.Msg = ERR.Msg
				return *pr
			}
		}
		contract.Source = hex.EncodeToString(sourceBytes)

		// 获取abi
		zipAbiBytes, errAbi := mining.GetContractAbi(contractAddr)
		if errAbi != nil {
			ERR := utils.NewErrorSysSelf(errAbi)
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		var abiBytes []byte
		if len(zipAbiBytes) > 0 {
			abiBytes, errAbi = common.UnzipBytes(zipAbiBytes)
			if errAbi != nil {
				ERR := utils.NewErrorSysSelf(errAbi)
				pr.Code = ERR.Code
				pr.Msg = ERR.Msg
				return *pr
			}
		}
		contract.Abi = string(abiBytes)

		// 获取bin
		zipBinBytes, errBin := mining.GetContractBin(contractAddr)
		if errBin != nil {
			ERR := utils.NewErrorSysSelf(errBin)
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		var binBytes string
		if len(zipBinBytes) > 0 {
			binBytes = common.Bytes2Hex(zipBinBytes)
		}
		contract.Bin = binBytes

	}
	pr.Data["contract"] = contract
	return *pr
}

func ContractCall(params *map[string]interface{}, srcaddress, contractAddress, dataStr string, gas, gasPrice uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	var src crypto.AddressCoin
	if srcaddress != "" {
		srcAddrItr := (*params)["srcaddress1"]
		src = crypto.AddressCoin(srcAddrItr.(coin_address.AddressCoin))
	}

	srcAddrItr := (*params)["contractAddress1"]
	contractAddr := crypto.AddressCoin(srcAddrItr.(coin_address.AddressCoin))

	if gasPrice < config.DEFAULT_GAS_PRICE {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_too_low, "gasPrice")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//判断矿工费是否足够
	//if gas < chainconfig.Wallet_tx_gas_min {
	//	//return nil, utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	//	ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_GasTooLittle, "least:"+strconv.Itoa(int(chainconfig.Wallet_tx_gas_min)))
	//	pr.Code = ERR.Code
	//	pr.Msg = ERR.Msg
	//	return *pr
	//}

	total, _ := chain.Balance.BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	utils.Log.Info().Str("调用参数 dataStr", dataStr).Send()

	/*------------------------*/
	txpay, err := mining.ContractTx(&src, &contractAddr, 0, gas, 0, pwd, dataStr, "", 0, gasPrice, "")
	// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			ERR := utils.NewErrorBus(keyconfig.ERROR_code_coinAddr_password_fail, "")
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	pr.Data["hash"] = hex.EncodeToString(*txpay.GetHash())
	pr.Data["tx"] = txpay
	return *pr
}

func ContractCallStack(params *map[string]interface{}, srcaddress, contractAddress, dataStr string, gas, gasPrice uint64) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	var src crypto.AddressCoin
	srcAddrItr := (*params)["srcaddress1"]
	src = crypto.AddressCoin(srcAddrItr.(coin_address.AddressCoin))

	srcAddrItr = (*params)["contractAddress1"]
	contractAddr := crypto.AddressCoin(srcAddrItr.(coin_address.AddressCoin))

	if gasPrice < config.DEFAULT_GAS_PRICE {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_too_low, "gasPrice")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	total, _ := chain.Balance.BuildPayVinNew(&src, gas*gasPrice)
	if total < gas*gasPrice {
		//资金不够
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_balance_not_enough, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	input := common.Hex2Bytes(dataStr)
	vmRun := evm.NewVmRun(src, contractAddr, []byte("0x1"), nil)
	result, _, evmErr := vmRun.Run(input, gas, gasPrice, 0, false)
	if evmErr != nil {
		ERR := utils.NewErrorSysSelf(evmErr)
		pr.Code = ERR.Code
		pr.Msg = "Contract execution failed.error code:" + result.ExitOpCode.String() + "," + evmErr.Error()
		return *pr
		//res, err = model.Errcode(model.Nomarl, "执行合约失败,退出码:"+result.ExitOpCode.String()+evmErr.Error())
		//return
	}
	if result.ExitOpCode == opcodes.REVERT {
		unpackedMsg, unpackErr := abi.UnpackRevert(result.ResultData)
		if unpackErr != nil {
			ERR := utils.NewErrorSysSelf(unpackErr)
			pr.Code = ERR.Code
			pr.Msg = "Contract execution failed.error code:" + result.ExitOpCode.String() + "," + unpackErr.Error()
			return *pr
			//res, err = model.Errcode(model.Nomarl, "执行合约失败,退出码:"+result.ExitOpCode.String()+unpackErr.Error())
			//return
		}
		ERR := utils.NewErrorSysSelf(unpackErr)
		pr.Code = ERR.Code
		pr.Msg = "Contract execution failed.error code:" + result.ExitOpCode.String() + "," + unpackedMsg
		return *pr
		//res, err = model.Errcode(model.Nomarl, "执行合约失败,退出码:"+result.ExitOpCode.String()+",失败原因:"+unpackedMsg)
		//return
	}
	resHex := common.Bytes2Hex(result.ResultData)
	//res, err = model.Tojson(resHex)
	pr.Data["data"] = resHex
	return *pr
}

func ContractEvent(params *map[string]interface{}, customTxEvent bool, txHash16 string, height uint64) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	var cte = false
	if _, ok := (*params)["customTxEvent"]; ok {
		cte = customTxEvent
	}
	var txHash []byte
	if txHashItr, ok := (*params)["txHash161"]; ok {
		txHash = txHashItr.([]byte)
	}

	//if !config.EVM_Reward_Enable && customTxEvent {
	if cte {
		//内存模式的事件信息
		event, ERR := GetCustomTxEvents(txHash)
		if ERR.CheckFail() {
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		pr.Data["data"] = event
		return *pr
	}

	list := &go_protos.ContractEventInfoList{}

	var intHeight []byte
	if height == 0 {
		_, _, _, blockHeight, _ := mining.FindTxJsonVo(txHash)
		intHeight = evmutils.New(0).SetUint64(blockHeight).Bytes()
	} else {
		intHeight = evmutils.New(int64(height)).Bytes()
	}

	key := append([]byte(config.DBKEY_BLOCK_EVENT), intHeight...)
	body, err := db.LevelDB.GetDB().HGet(key, txHash)
	if err != nil {
		//engine.Log.Error("获取保存合约事件日志失败%s", err.Error())
		pr.Data["data"] = list
		return *pr
	}
	if len(body) == 0 {
		pr.Data["data"] = list
		return *pr
	}

	err = proto.Unmarshal(body, list)
	if err != nil {
		pr.Data["data"] = list
		return *pr
	}
	pr.Data["data"] = list
	return *pr
}

// 获取自定义交易事件
func GetCustomTxEvents(txHhash []byte) (*go_protos.CustomTxEvents, utils.ERROR) {
	e := go_protos.CustomTxEvents{
		CustomTxEvents: []*go_protos.CustomTxEvent{},
		RewardPools:    []*go_protos.CustomTxEvent{},
	}
	key := config.BuildCustomTxEvent(txHhash)
	body, err := db.LevelTempDB.Find(key)
	if err != nil || len(*body) == 0 {
		return nil, utils.NewErrorSysSelf(err)
	}
	if err = e.Unmarshal(*body); err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return &e, utils.NewErrorSuccess()
}
