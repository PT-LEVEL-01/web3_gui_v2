package example

import (
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/message_center"
)

func RegisterMsg(node *libp2parea.Node) {
	hs := handlers{node}
	node.Register_p2p(MSGID_hello, hs.hello_h)
	node.Register_p2pHE(MSGID_hello_HE, hs.helloHE_h)
}

type handlers struct {
	node *libp2parea.Node
}

func (this *handlers) hello_h(message *message_center.MessageBase) {
	this.node.Log.Info().Str("对方发送的消息内容", string(message.Content)).Send()
	ERR := this.node.SendP2pReplyMsg(message, &message.Content)
	if ERR.CheckFail() {
		this.node.Log.Error().Str("回复消息失败", ERR.String()).Send()
	}
}

func (this *handlers) helloHE_h(message *message_center.MessageBase) {
	this.node.Log.Info().Str("对方发送的加密消息内容", string(message.Content)).Send()
	ERR := this.node.SendP2pReplyMsgHE(message, &message.Content)
	if ERR.CheckFail() {
		this.node.Log.Error().Str("回复加密消息失败", ERR.String()).Send()
	}
}
