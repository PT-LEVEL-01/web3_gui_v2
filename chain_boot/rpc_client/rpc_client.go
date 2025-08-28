package rpc_client

import (
	"bytes"
	"encoding/base64"
	jsoniter "github.com/json-iterator/go"
	"reflect"
	"web3_gui/chain/mining"
	"web3_gui/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
查询区块高度
*/
func GetInfo(ipPort, rpcUser, prcPwd string) (*Info, utils.ERROR) {
	//{"method":"getinfo"}
	params := map[string]interface{}{}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_Info, params)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//fmt.Printf("Info接口返回:%+v\n", result.Data)
	info := new(Info)
	infoItr, ok := result.Data["info"]
	if !ok {
		return info, ERR
	}
	bs, err := json.Marshal(infoItr)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(info)
	return info, ERR
}

type Info struct {
	Netid          []byte `json:"netid"`          //网络版本号
	TotalAmount    uint64 `json:"TotalAmount"`    //发行总量
	Balance        uint64 `json:"balance"`        //可用余额
	BalanceFrozen  uint64 `json:"BalanceFrozen"`  //冻结的余额
	Testnet        bool   `json:"testnet"`        //是否是测试网络
	Blocks         uint64 `json:"blocks"`         //已经同步到的区块高度
	Group          uint64 `json:"group"`          //区块组高度
	StartingBlock  uint64 `json:"StartingBlock"`  //区块开始高度
	HighestBlock   uint64 `json:"HighestBlock"`   //所链接的节点的最高高度
	CurrentBlock   uint64 `json:"CurrentBlock"`   //已经同步到的区块高度
	PulledStates   uint64 `json:"PulledStates"`   //正在同步的区块高度
	BlockTime      uint64 `json:"BlockTime"`      //出块时间
	LightNode      uint64 `json:"LightNode"`      //轻节点押金数量
	CommunityNode  uint64 `json:"CommunityNode"`  //社区节点押金数量
	WitnessNode    uint64 `json:"WitnessNode"`    //见证人押金数量
	NameDepositMin uint64 `json:"NameDepositMin"` //域名押金最少金额
	AddrPre        string `json:"AddrPre"`        //地址前缀
}

/*
获取指定高度的区块及交易
*/
func BlockProto64ByHeight(ipPort, rpcUser, prcPwd string, height uint64) (*mining.BlockHeadVO, utils.ERROR) {
	params := map[string]interface{}{
		"height": height,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_BlockProto64ByHeight, params)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//fmt.Printf("BlockProto64ByHeight 接口返回:%+v\n", result.Data)

	blockProto64, ok := result.Data["BlockProto64"]
	if !ok {
		return nil, ERR
	}
	bs, err := base64.StdEncoding.DecodeString(blockProto64.(string))
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	bhvo, err := mining.ParseBlockHeadVOProto(&bs)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return bhvo, ERR
}

/*
判断地址数量，解析地址及余额
*/
func AddressList(ipPort, rpcUser, prcPwd string, startAddr string, total int) ([]Address, utils.ERROR) {
	params := map[string]interface{}{
		"startAddr": startAddr,
		"total":     total,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_AddressList, params)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//fmt.Printf("BlockProto64ByHeight 接口返回:%+v\n", result.Data)
	listJsonStrItr, ok := result.Data["data"]
	if !ok {
		return nil, ERR
	}
	addrs := make([]Address, 0)
	for _, one := range listJsonStrItr.([]interface{}) {
		bs, err := json.Marshal(one)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		addr := new(Address)
		buf := bytes.NewBuffer(bs)
		decoder := json.NewDecoder(buf)
		decoder.UseNumber()
		err = decoder.Decode(addr)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		addrs = append(addrs, *addr)
	}
	return addrs, ERR
}

type Address struct {
	Index       int
	AddrCoin    string
	Value       uint64
	ValueFrozen uint64
	ValueLockup uint64
	Type        int
}

/*
创建新地址
*/
func AddressCreate(ipPort, rpcUser, prcPwd, nickname, seedPassword, addrPassword string) (string, utils.ERROR) {
	params := map[string]interface{}{
		"nickname":     nickname,
		"seedPassword": seedPassword,
		"addrPassword": addrPassword,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_AddressCreate, params)
	if err != nil {
		return "", utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return "", ERR
	}

	//fmt.Printf("CreatNewAddr 接口返回:%+v\n", result.Data)
	addrStrItr, ok := result.Data["addr"]
	if !ok {
		return "", ERR
	}
	return addrStrItr.(string), ERR
}

/*
转账
*/
func SendToAddress(ipPort, rpcUser, prcPwd, srcAddress, address string, amount, gas uint64, pwd string) (*mining.Tx_Pay, utils.ERROR) {
	params := map[string]interface{}{
		"srcaddress": srcAddress,
		"address":    address,
		"amount":     amount,
		"gas":        gas,
		"pwd":        pwd,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_SendToAddress, params)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return nil, ERR
	}
	txJsonStrItr, ok := result.Data["data"]
	if !ok {
		return nil, ERR
	}
	bs, err := json.Marshal(txJsonStrItr)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	txPay := new(mining.Tx_Pay)
	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(txPay)
	return txPay, ERR
}

/*
给多人转账
*/
func SendToAddressMore(ipPort, rpcUser, prcPwd, srcAddress string, payNumber []PayNumber, gas uint64, pwd, comment string) (*mining.Tx_Pay, utils.ERROR) {
	if len(payNumber) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	addressMore := make([]string, 0, len(payNumber))
	amountMore := make([]uint64, 0, len(payNumber))
	frozenHeightMore := make([]uint64, 0, len(payNumber))
	domainMore := make([]string, 0, len(payNumber))
	domainTypeMore := make([]int, 0, len(payNumber))
	for _, one := range payNumber {
		addressMore = append(addressMore, one.Address)
		amountMore = append(amountMore, one.Amount)
		frozenHeightMore = append(frozenHeightMore, one.FrozenHeight)
		domainMore = append(domainMore, one.Domain)
		domainTypeMore = append(domainTypeMore, one.DomainType)
	}
	params := map[string]interface{}{
		"srcaddress":       srcAddress,
		"addressMore":      addressMore,
		"amountMore":       amountMore,
		"frozenHeightMore": frozenHeightMore,
		"domainMore":       domainMore,
		"domainTypeMore":   domainTypeMore,
		"gas":              gas,
		"pwd":              pwd,
		"comment":          comment,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_SendToAddressMore, params)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return nil, ERR
	}
	txJsonStrItr, ok := result.Data["data"]
	if !ok {
		return nil, ERR
	}
	bs, err := json.Marshal(txJsonStrItr)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	txPay := new(mining.Tx_Pay)
	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(txPay)
	return txPay, ERR
}

/*
多人转账
*/
type PayNumber struct {
	Address      string `json:"address"` //转账地址
	Amount       uint64 `json:"amount"`  //转账金额
	FrozenHeight uint64 //
	Domain       string //
	DomainType   int    //
}

/*
查询全网多个地址的余额
*/
func AddressBalanceMore(ipPort, rpcUser, prcPwd string, addressMore []string) ([]uint64, []uint64, utils.ERROR) {
	if len(addressMore) == 0 {
		return nil, nil, utils.NewErrorSuccess()
	}
	params := map[string]interface{}{
		"addressMore": addressMore,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_AddressBalanceMore, params)
	if err != nil {
		return nil, nil, utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return nil, nil, ERR
	}

	//fmt.Printf("result: %v\n", result)

	valuesItr, ok := result.Data["BalanceNotSpend"]
	if !ok {
		return nil, nil, ERR
	}
	valuesFrozenItr, ok := result.Data["BalanceFrozen"]
	if !ok {
		return nil, nil, ERR
	}

	sliseItr, ok := valuesItr.([]interface{})
	if !ok {
		return nil, nil, utils.NewErrorBus(engine.ERROR_code_rpc_result_param_type_fail, "")
	}
	results := make([]uint64, 0, len(sliseItr))
	//整理类型
	for _, oneValue := range sliseItr {
		itr, ERR := engine.ConverNeedType(oneValue, reflect.Uint64)
		if ERR.CheckFail() {
			return nil, nil, ERR
		}
		v, ok := itr.(uint64)
		if !ok {
			return nil, nil, utils.NewErrorBus(engine.ERROR_code_rpc_result_param_type_fail, "")
		}
		results = append(results, v)
	}

	sliseItr, ok = valuesFrozenItr.([]interface{})
	if !ok {
		return nil, nil, utils.NewErrorBus(engine.ERROR_code_rpc_result_param_type_fail, "")
	}
	resultsFrozen := make([]uint64, 0, len(sliseItr))
	//整理类型
	for _, oneValue := range sliseItr {
		itr, ERR := engine.ConverNeedType(oneValue, reflect.Uint64)
		if ERR.CheckFail() {
			return nil, nil, ERR
		}
		v, ok := itr.(uint64)
		if !ok {
			return nil, nil, utils.NewErrorBus(engine.ERROR_code_rpc_result_param_type_fail, "")
		}
		resultsFrozen = append(resultsFrozen, v)
	}
	return results, resultsFrozen, ERR
}

/*
查询全网多个地址的余额
*/
func PushTxProto64(ipPort, rpcUser, prcPwd string, base64StdStr string, checkBalance bool) (map[string]interface{}, utils.ERROR) {
	params := map[string]interface{}{
		"base64StdStr": base64StdStr,
		"checkBalance": checkBalance,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_PushTxProto64, params)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return nil, ERR
	}
	//fmt.Printf("result: %v\n", result)
	txInfoItr, ok := result.Data["info"]
	if !ok {
		return nil, ERR
	}
	txInfo := txInfoItr.(map[string]interface{})
	return txInfo, ERR
}

/*
创建一个离线签名交易
*/
func OfflineTxCreate(ipPort, rpcUser, prcPwd, key_store_path, srcaddress, address string, amount, nonce, currentHeight,
	frozen_height, gas uint64, pwd, comment string) (string, string, utils.ERROR) {
	params := map[string]interface{}{
		"key_store_path": key_store_path,
		"srcaddress":     srcaddress,
		"address":        address,
		"amount":         amount,
		"nonce":          nonce,
		"currentHeight":  currentHeight,
		"frozen_height":  frozen_height,
		"gas":            gas,
		"pwd":            pwd,
		"comment":        comment,
	}
	result, err := engine.Post(ipPort, rpcUser, prcPwd, config.RPC_Method_chain_OfflineTxCreate, params)
	if err != nil {
		utils.Log.Info().Interface("返回结果", result).Send()
		return "", "", utils.NewErrorSysSelf(err)
	}
	ERR := result.ConverERR()
	if ERR.CheckFail() {
		return "", "", ERR
	}
	txhashItr, ok := result.Data["hash"]
	if !ok {
		return "", "", ERR
	}
	txhash := txhashItr.(string)
	txStrB64Itr, ok := result.Data["tx"]
	if !ok {
		return "", "", ERR
	}
	txStrB64 := txStrB64Itr.(string)
	return txhash, txStrB64, ERR
}
