package module_area_rpc

import (
	"time"

	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/utils"
)

// 用于测试
const (
	DEFAULT_TIMEOUT      = 6 * time.Second
	MSGID_TEST_MULTICAST = 70000 + iota
	MSGID_TEST_SEARCH_SUPER
	MSGID_TEST_SEARCH_SUPER_WAIT
	MSGID_TEST_SEARCH_SUPER_WAIT_RECV
	MSGID_TEST_P2P
	MSGID_TEST_P2P_WAIT
	MSGID_TEST_P2P_WAIT_RECV
	MSGID_TEST_P2P_HE
	MSGID_TEST_P2P_HE_WAIT
	MSGID_TEST_P2P_HE_WAIT_RECV
	MSGID_TEST_P2P_PROXY
	MSGID_TEST_P2P_PROXY_WAIT
	MSGID_TEST_P2P_PROXY_WAIT_RECV
	MSGID_TEST_P2P_HE_PROXY
	MSGID_TEST_P2P_HE_PROXY_WAIT
	MSGID_TEST_P2P_HE_PROXY_WAIT_RECV
	MSGID_TEST_SEARCH_SUPER_PROXY
	MSGID_TEST_SEARCH_SUPER_PROXY_WAIT
	MSGID_TEST_SEARCH_SUPER_PROXY_WAIT_RECV
)

var msglog *engine.LogQueue
var MSGIDStr = make(map[uint64]string)

func init() {
	// 打印消息到文件
	msglog = engine.NewLog(engine.LOG_file, "msg_test.log")

	MSGIDStr[MSGID_TEST_MULTICAST] = "sendmulticastmsg"
	MSGIDStr[MSGID_TEST_SEARCH_SUPER] = "sendsearchsupermsg"
	MSGIDStr[MSGID_TEST_P2P] = "sendp2pmsg"
	MSGIDStr[MSGID_TEST_P2P_HE] = "sendp2pmsghe"
	MSGIDStr[MSGID_TEST_P2P_PROXY] = "sendp2pmsgproxy"
	MSGIDStr[MSGID_TEST_P2P_HE_PROXY] = "sendp2pmsgheproxy"
	MSGIDStr[MSGID_TEST_SEARCH_SUPER_PROXY] = "sendsearchsupermsgproxy"
}

// 注册测试 MSGID
func RegisterTestMsg() {
	Area.Register_multicast(MSGID_TEST_MULTICAST, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		printLog(MSGID_TEST_MULTICAST, message)
	})

	Area.Register_search_super(MSGID_TEST_SEARCH_SUPER, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		printLog(MSGID_TEST_SEARCH_SUPER, message)
	})
	Area.Register_search_super(MSGID_TEST_SEARCH_SUPER_WAIT, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		Area.MessageCenter.SendSearchSuperReplyMsg(message, MSGID_TEST_SEARCH_SUPER_WAIT_RECV, message.Body.Content)
	})
	Area.Register_search_super(MSGID_TEST_SEARCH_SUPER_WAIT_RECV, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	})

	Area.Register_p2p(MSGID_TEST_P2P, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		printLog(MSGID_TEST_P2P, message)
	})
	Area.Register_p2p(MSGID_TEST_P2P_WAIT, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		Area.MessageCenter.SendP2pReplyMsg(message, MSGID_TEST_P2P_WAIT_RECV, message.Body.Content)
	})
	Area.Register_p2p(MSGID_TEST_P2P_WAIT_RECV, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	})

	Area.Register_p2pHE(MSGID_TEST_P2P_HE, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		printLog(MSGID_TEST_P2P_HE, message)
	})
	Area.Register_p2pHE(MSGID_TEST_P2P_HE_WAIT, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		Area.MessageCenter.SendP2pReplyMsgHE(message, MSGID_TEST_P2P_HE_WAIT_RECV, message.Body.Content)
	})
	Area.Register_p2pHE(MSGID_TEST_P2P_HE_WAIT_RECV, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	})

	Area.Register_search_super(MSGID_TEST_SEARCH_SUPER_PROXY, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		printLog(MSGID_TEST_P2P_HE, message)
	})
	Area.Register_search_super(MSGID_TEST_SEARCH_SUPER_PROXY_WAIT, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		Area.MessageCenter.SendSearchSuperReplyMsg(message, MSGID_TEST_SEARCH_SUPER_PROXY_WAIT_RECV, message.Body.Content)
	})
	Area.Register_search_super(MSGID_TEST_SEARCH_SUPER_PROXY_WAIT_RECV, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	})

	Area.Register_p2p(MSGID_TEST_P2P_PROXY, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		printLog(MSGID_TEST_P2P_PROXY, message)
	})
	Area.Register_p2p(MSGID_TEST_P2P_PROXY_WAIT, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		Area.MessageCenter.SendP2pReplyMsg(message, MSGID_TEST_P2P_PROXY_WAIT_RECV, message.Body.Content)
	})
	Area.Register_p2p(MSGID_TEST_P2P_PROXY_WAIT_RECV, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	})

	Area.Register_p2pHE(MSGID_TEST_P2P_HE_PROXY, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		printLog(MSGID_TEST_P2P_HE_PROXY, message)
	})
	Area.Register_p2pHE(MSGID_TEST_P2P_HE_PROXY_WAIT, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		Area.MessageCenter.SendP2pReplyMsgHE(message, MSGID_TEST_P2P_HE_PROXY_WAIT_RECV, message.Body.Content)
	})
	Area.Register_p2pHE(MSGID_TEST_P2P_HE_PROXY_WAIT_RECV, func(c engine.Controller, msg engine.Packet, message *message_center.Message) {
		flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	})
}

func printLog(msgid uint64, message *message_center.Message) {
	msgIdStr, ok := MSGIDStr[msgid]
	if ok {
		msglog.Info(0, "[%s] %v", msgIdStr, string(*message.Body.Content))
		return
	}

	msglog.Info(0, "[%d] %v", msgid, string(*message.Body.Content))
}
