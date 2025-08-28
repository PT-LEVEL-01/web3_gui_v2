package chain

import "C"
import (
	"encoding/json"
	"strconv"
	"web3_gui/chain/mining"
	"web3_gui/chain/utils"
)

/*
创建离线交易
*/
func CreateOfflineTx(keyStorePath, srcaddress, address, pwd, comment string, amount, gas, frozenHeight, nonce, currentHeight string, domain string, domainType string) string {
	res, _, err := mining.CreateOfflineTx(keyStorePath, srcaddress, address, pwd, comment, StringToUint64(amount), StringToUint64(gas), StringToUint64(frozenHeight), StringToUint64(nonce), StringToUint64(currentHeight), domain, StringToUint64(domainType))
	if err != nil {
		return utils.Out(500, err.Error())
	}

	//b, err := json.Marshal(res)
	//if err != nil {
	//	return utils.Out(500, err.Error())
	//}
	return utils.Out(200, res)

}

/*
创建离线合约交易
*/
func CreateOfflineContractTx(keyStorePath, srcaddress, address, pwd, comment string, amount, gas, frozenHeight, gasPrice, nonce, currentHeight string, domain string, domainType, abi, source string) string {
	res, hash, addressContract, err := mining.CreateOfflineContractTx(keyStorePath, srcaddress, address, pwd, comment, StringToUint64(amount), StringToUint64(gas), StringToUint64(frozenHeight), StringToUint64(gasPrice), StringToUint64(nonce), StringToUint64(currentHeight), domain, StringToUint64(domainType), abi, source)
	if err != nil {
		return utils.Out(500, err.Error())
	}

	var txData struct {
		Hash    string `json:"hash"`
		Tx      string `json:"tx"`
		Address string `json:"address"`
	}
	txData.Hash = hash
	txData.Tx = res
	txData.Address = addressContract

	//engine.Log.Error("keyStorePath:", keyStorePath)
	//engine.Log.Error("srcaddress:", srcaddress)
	//engine.Log.Error("address:", address)
	//engine.Log.Error("pwd:", pwd)
	//engine.Log.Error("comment:", comment)
	//engine.Log.Error("amount:", StringToUint64(amount))
	//engine.Log.Error("gas:", StringToUint64(gas))
	//engine.Log.Error("frozenHeight:", StringToUint64(frozenHeight))
	//engine.Log.Error("gasPrice:", StringToUint64(gasPrice))
	//engine.Log.Error("nonce:", StringToUint64(nonce))
	//engine.Log.Error("currentHeight:", StringToUint64(currentHeight))
	//engine.Log.Error("domain:", domain)
	//engine.Log.Error("domainType:", StringToUint64(domainType))

	//b, err := json.Marshal(res)
	//if err != nil {
	//	return utils.Out(500, err.Error()+"result:"+string(b))
	//}
	return utils.Out(200, txData)

}

func StringToUint64(intStr string) uint64 {
	intNum, _ := strconv.Atoi(intStr)
	return uint64(intNum)
}

/*
获取comment
*/
func GetComment(tag, jsonDataItrs string) string {

	bs := []byte(jsonDataItrs)
	jsonData := make(map[string]interface{})
	err := json.Unmarshal(bs, &jsonData)
	if err != nil {
		return utils.Out(500, "Umarshal failed")
	}

	res, err := mining.GetComment(tag, jsonData)
	if err != nil {
		return utils.Out(500, err.Error())
	}

	//b, err := json.Marshal(res)
	//if err != nil {
	//	return utils.Out(500, err.Error())
	//}
	return utils.Out(200, res)

}

/*
合并处理comment、构建离线交易和push
*/
func MultDeal(tag, jsonDataItrs, keyStorePath, srcaddress, address, pwd, comment string, amount, gas, frozenHeight, gasPrice, nonce, currentHeight uint64, domain string, domainType uint64) string {

	bs := []byte(jsonDataItrs)
	jsonData := make(map[string]interface{})
	err := json.Unmarshal(bs, &jsonData)
	if err != nil {
		return utils.Out(500, "Umarshal failed")

	}

	hash, err := mining.MultDeal(tag, jsonData, keyStorePath, srcaddress, address, pwd, amount, gas, frozenHeight, gasPrice, nonce, currentHeight, domain, domainType)
	if err != nil {
		return utils.Out(500, err.Error())
	}

	//b, _ := json.Marshal(hash)
	return utils.Out(200, hash)

}
