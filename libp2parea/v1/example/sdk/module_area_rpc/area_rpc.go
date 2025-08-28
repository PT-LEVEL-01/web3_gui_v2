package module_area_rpc

import (
	"web3_gui/libp2parea/v1"
	rpc "web3_gui/libp2parea/v1/sdk/jsonrpc2"
	"web3_gui/libp2parea/v1/sdk/web/routers"
)

var Area *libp2parea.Area

func Start(area *libp2parea.Area) {
	Area = area
	RegisterLibp2pareaRPC()
	routers.RegisterRpc()
}

func RegisterLibp2pareaRPC() {
	// p2p 管理RPC
	rpc.RegisterRPC("waitautonomyfinish", WaitAutonomyFinish)
	rpc.RegisterRPC("waitautonomyfinishvnode", WaitAutonomyFinishVnode)
	rpc.RegisterRPC("getnetid", GetNetId)
	rpc.RegisterRPC("getvnodeid", GetVnodeId)
	rpc.RegisterRPC("getidinfo", GetIdInfo)
	rpc.RegisterRPC("getnodeself", GetNodeSelf)
	rpc.RegisterRPC("closenet", CloseNet)
	rpc.RegisterRPC("reconnectnet", ReconnectNet)
	rpc.RegisterRPC("checkonline", CheckOnline)
	rpc.RegisterRPC("addwhitelist", AddWhiteList)
	rpc.RegisterRPC("removewhitelist", RemoveWhiteList)
	rpc.RegisterRPC("addconnect", AddConnect)
	rpc.RegisterRPC("searchnetaddr", SearchNetAddr)
	rpc.RegisterRPC("searchnetaddrvnode", SearchNetAddrVnode)
	rpc.RegisterRPC("networkinfolist", GetNetworkInfo)
	rpc.RegisterRPC("addaddrwhitelist", AddAddrWhiteList)
	rpc.RegisterRPC("setareagodaddr", SetAreaGodAddr)
	rpc.RegisterRPC("searchnetaddrproxy", SearchNetAddrProxy)
	rpc.RegisterRPC("findnearvnodessearchvnode", FindNearVnodesSearchVnode)
	rpc.RegisterRPC("getallnodes", GetAllNodes)
	rpc.RegisterRPC("getmachineid", GetMachineID)
	rpc.RegisterRPC("getallconnectnode", GetAllConnectNode)

	// p2p 消息RPC
	rpc.RegisterRPC("sendmulticastmsg", SendMulticastMsg)
	rpc.RegisterRPC("sendsearchsupermsg", SendSearchSuperMsg)
	rpc.RegisterRPC("sendsearchsupermsgwaitrequest", SendSearchSuperMsgWaitRequest)
	rpc.RegisterRPC("sendp2pmsg", SendP2pMsg)
	rpc.RegisterRPC("sendp2pmsgwaitrequest", SendP2pMsgWaitRequest)
	rpc.RegisterRPC("sendp2pmsghe", SendP2pMsgHE)
	rpc.RegisterRPC("sendp2pmsghewaitrequest", SendP2pMsgHEWaitRequest)
	rpc.RegisterRPC("searchvnodeid", SearchVnodeId)
	rpc.RegisterRPC("sendsearchsupermsgproxy", SendSearchSuperMsgProxy)
	rpc.RegisterRPC("sendsearchsupermsgproxywaitrequest", SendSearchSuperMsgProxyWaitRequest)
	rpc.RegisterRPC("sendp2pmsgproxy", SendP2pMsgProxy)
	rpc.RegisterRPC("sendp2pmsgproxywaitrequest", SendP2pMsgProxyWaitRequest)
	rpc.RegisterRPC("sendp2pmsgheproxy", SendP2pMsgHEProxy)
	rpc.RegisterRPC("sendp2pmsgheproxywaitrequest", SendP2pMsgHEProxyWaitRequest)

	// 消息测试，输出日志
	RegisterTestMsg()
}
