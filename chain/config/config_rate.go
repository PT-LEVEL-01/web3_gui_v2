package config

import (
	"time"
	"web3_gui/utils"
)

/*
rpc请求速率
*/
var RpcRateMap = map[string]time.Duration{
	"sendtoaddress":      time.Second / 5000, //转帐限速每秒5000
	"MulticastTx":        time.Second / 5000,
	"pushtx":             time.Second / 5000,
	"createContract":     time.Second / 300,
	"callContract":       time.Second / 300,
	"staticCallContract": time.Second / 300,
	"pushContractTx":     time.Second / 300,
}

// 注册rpc限速器
func registerRpcRate() {
	for k, v := range RpcRateMap {
		utils.SetTimeToken(k, v)
	}
}

/*
交易接收速率
*/
var TxRateMap = map[uint64]*TxRateConf{
	Wallet_tx_type_pay:      {"TIMETOKEN_Tx", time.Second / 5000},
	Wallet_tx_type_contract: {"TIMETOKEN_ContractTx_Create", time.Second / 300},
}

type TxRateConf struct {
	N string
	R time.Duration
}

// 注册交易限速器
func registerTxRate() {
	for _, v := range TxRateMap {
		utils.SetTimeToken(v.N, v.R)
	}
}

// 注册限速器
func RegisterRateHandle() {
	registerRpcRate()
	registerTxRate()
}

func GetRpcRate(m string, w bool) {
	if _, ok := RpcRateMap[m]; !ok {
		return
	}

	utils.GetTimeToken(m, w)
}

func GetTxRate(tx uint64, w bool) {
	conf, ok := TxRateMap[tx]
	if !ok {
		return
	}

	utils.GetTimeToken(conf.N, w)
}
