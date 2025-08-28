package light

import (
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
)

func RegisterMulticastMsg() {
	Area.Register_multicast(111111, GetMulticastMsg) //获得候选见证人列表
}

func GetMulticastMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//if !Area.NodeManager.NodeSelf.IsApp {
	//	content := string(*message.Body.Content)
	//	engine.Log.Info("接收到广播消息：%s", content)
	//	engine.Log.Info("我的地址：%s", Area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	//}
}
