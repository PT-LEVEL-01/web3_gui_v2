package im

import (
	"encoding/hex"
	"reflect"
	"web3_gui/chain_orders"
	"web3_gui/config"
	"web3_gui/im/db"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func RegisterRPC(node *libp2parea.Node) {
	//utils.Log.Info().Str("注册rpc接口", "").Send()
	RegisterRPC_Panic(node, 101, config.RPC_Method_Sharebox_GetFileOrder, Sharebox_GetFileOrder, "获取文件下载订单",
		engine.NewParamValid_Mast_CustomFun_Panic("addr", reflect.String, VlidAddressNet, "address", "远端地址"),
		engine.NewParamValid_Mast_Panic("fileHash16", reflect.String, "文件hash 16进制"),
		engine.NewParamValid_Mast_Panic("price", reflect.Uint64, "价格"))

	//RegisterRPC_Panic(node, 102, config.RPC_Method_Sharebox_PayOrder, Sharebox_PayOrder, "支付订单",
	//	engine.NewParamValid_Mast_CustomFun_Panic("serverAddr", reflect.String, VlidAddress, "serverAddr1", "远端地址"),
	//	engine.NewParamValid_Mast_Panic("orderId16", reflect.String, "订单id 16进制"),
	//	engine.NewParamValid_Mast_Panic("amount", reflect.Uint64, "支付总额"),
	//	engine.NewParamValid_Mast_Panic("pwd", reflect.String, "钱包支付密码"))
}

func RegisterRPC_Panic(node *libp2parea.Node, index int, rpcName string, handler any, desc string, pvs ...engine.ParamValid) {
	ERR := node.RegisterRPC(index, rpcName, handler, desc, pvs...)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
}

/*
验证地址的合法性
*/
func VlidAddressNet(params interface{}) (any, utils.ERROR) {
	addrStr := params.(string)
	if addrStr == "" {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	//utils.Log.Info().Str("地址", addrStr).Send()
	addrNet := nodeStore.AddressFromB58String(addrStr)
	if !keystore.ValidAddrNet(addrNet.GetPre(), addrNet.GetAddr()) {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	return addrNet, utils.NewErrorSuccess()
}

/*
获取订单
*/
func Sharebox_GetFileOrder(params *map[string]interface{}, addr, fileHash16 string, price uint64) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["address"]
	remoteAddr := addrItr.(nodeStore.AddressNet)
	fileHash, err := hex.DecodeString(fileHash16)
	if err != nil {
		ERR := utils.NewErrorSuccess()
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	order, ERR := GetOrder(remoteAddr, fileHash, price)
	if ERR.CheckFail() {
		utils.Log.Info().Interface("获取的订单信息", order).Send()
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	ERR = db.Sharebox_Client_SaveOrderUnpaid(order)
	if ERR.CheckFail() {
		utils.Log.Info().Str("保存订单失败", ERR.String()).Send()
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	ERR = chain_orders.OrderClientStatic.RegOrder(order)
	if ERR.CheckFail() {
		utils.Log.Info().Str("获取的订单信息", ERR.String()).Send()
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["data"] = order.ConverVO()
	return *pr
}

/*
支付订单
*/
//func Sharebox_PayOrder(params *map[string]interface{}, serverAddr, orderId16 string, amount uint64, pwd string) engine.PostResult {
//	pr := engine.NewPostResult()
//	addrItr := (*params)["serverAddr1"]
//	remoteAddr := addrItr.(nodeStore.AddressNet)
//	fileHash, err := hex.DecodeString(orderId16)
//	if err != nil {
//		ERR := utils.NewErrorSuccess()
//		pr.Code = ERR.Code
//		pr.Msg = ERR.Msg
//		return *pr
//	}
//	txPay, ERR := PayOrder(remoteAddr.GetAddr(), fileHash, amount, pwd)
//	if ERR.CheckFail() {
//		pr.Code = ERR.Code
//		pr.Msg = ERR.Msg
//		return *pr
//	}
//	utils.Log.Info().Interface("获取的订单信息", txPay).Send()
//	pr.Code = utils.ERROR_CODE_success
//	pr.Data["data"] = txPay
//	return *pr
//}
