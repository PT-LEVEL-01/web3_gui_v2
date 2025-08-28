package storage

import (
	"web3_gui/config"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func RegisterRPC(node *libp2parea.Node) {
	//utils.Log.Info().Str("注册rpc接口", "").Send()
	//RegisterRPC_Panic(node, 101, config.RPC_Method_Sharebox_GetFileOrder, Sharebox_GetFileOrder, "获取文件下载订单",
	//	engine.NewParamValid_Mast_CustomFun_UnsupportedTypePanic("addr", reflect.String, VlidAddress, "address", "远端地址"),
	//	engine.NewParamValid_Mast_UnsupportedTypePanic("fileHash16", reflect.String, "文件hash 16进制"),
	//	engine.NewParamValid_Mast_UnsupportedTypePanic("price", reflect.Int64, "价格"))
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
	addrNet := nodeStore.AddressFromB58String(addrStr)
	if !keystore.ValidAddrNet(addrNet.GetPre(), addrNet.GetAddr()) {
		return nil, utils.NewErrorBus(config.ERROR_CODE_params_format, "")
	}
	return addrNet, utils.NewErrorSuccess()
}
