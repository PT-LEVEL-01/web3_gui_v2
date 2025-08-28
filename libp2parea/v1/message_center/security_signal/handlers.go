package security_signal

// import (
// 	// "web3_gui/libp2parea/v1/config"
// 	"web3_gui/libp2parea/v1/engine"

// 	// mc "web3_gui/libp2parea/v1/message_center"
// 	"web3_gui/libp2parea/v1/message_center/flood"
// 	// "web3_gui/libp2parea/v1/nodeStore"
// 	// "encoding/json"
// 	"fmt"
// )

// func Init() {

// 	// mc.Register_p2p(config.MSGID_SearchAddr, SearchAddress)
// 	// mc.Register_p2p(config.MSGID_SearchAddr_recv, SearchAddress_recv)

// 	// engine.RegisterMsg(MSGID_SearchAddr, SearchAddress)
// 	// engine.RegisterMsg(MSGID_SearchAddr_recv, SearchAddress_recv)

// 	// message_search.Init()
// }

// /*
// 	获取节点地址和身份公钥
// */
// func SearchAddress(c engine.Controller, msg engine.Packet, message *Message) {

// 	if !message.CheckSendhash() {
// 		return
// 	}

// 	fmt.Println("收到Hello消息", string(*message.Body.Content))

// 	//回复消息
// 	// data, _ := json.Marshal(nodeStore.NodeSelf)
// 	SendP2pReplyMsg(message, MSGID_SearchAddr_recv, nil)

// }

// /*
// 	获取节点地址和身份公钥_返回
// */
// func SearchAddress_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {

// 	if !message.CheckSendhash() {
// 		return
// 	}

// 	fmt.Println("收到Hello消息", string(*message.Body.Content))

// 	message.Head.RecvId

// 	flood.ResponseWait(CLASS_Hello, message.Body.Hash.B58String(), message.Body.Content)

// }
