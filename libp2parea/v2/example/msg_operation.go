package example

import (
	"time"
	"web3_gui/libp2parea/v2"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func hello_o(node *libp2parea.Node, recvNodeId *nodeStore.AddressNet) utils.ERROR {
	//node.Log.Info().Str("发送消息给", recvNodeId.B58String()).Send()
	bs := []byte("hello")
	resultBs, ERR := node.SendP2pMsgWaitRequest(MSGID_hello, recvNodeId, &bs, time.Second*10)
	if ERR.CheckFail() {
		node.Log.Error().Str("发送消息失败", ERR.String()).Send()
		return ERR
	}
	node.Log.Info().Str("对方返回消息内容", string(*resultBs)).Send()
	return utils.NewErrorSuccess()
}

/*
发送一条加密消息
*/
func HelloHE_o(node *libp2parea.Node, recvNodeId *nodeStore.AddressNet) utils.ERROR {
	node.Log.Info().Str("发送加密消息给", recvNodeId.B58String()).Send()
	bs := []byte("hello")
	resultBs, ERR := node.SendP2pMsgHEWaitRequest(MSGID_hello_HE, recvNodeId, &bs, time.Second*10)
	if ERR.CheckFail() {
		node.Log.Error().Str("发送消息失败", ERR.String()).Send()
		return ERR
	}
	node.Log.Info().Str("对方返回加密消息内容", string(*resultBs)).Send()
	return utils.NewErrorSuccess()
}
