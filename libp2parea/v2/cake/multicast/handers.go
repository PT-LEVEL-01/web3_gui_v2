package multicast

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/libp2parea/v2/engine"
	mc "web3_gui/libp2parea/v2/message_center"
	"web3_gui/libp2parea/v2/message_center/flood"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/libp2parea/v2/protobuf/go_protobuf"
	"web3_gui/utils"
)

// 注册消息
func (gm *MulticastMsg) registerMsg() {
	if gm == nil || gm.Area == nil || !gm.initialized {
		return
	}

	gm.Area.Register_p2p(MSGID_P2P_SEND_MULTICAST_MSG, gm.sendMulticastMsg)
	gm.Area.Register_p2p(MSGID_P2P_SEND_MULTICAST_MSG_RECV, gm.sendMulticastMsg_recv)
}

// 发送组播消息
func (gm *MulticastMsg) sendMulticastMsg(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// 1. 判断消息内容
	if message == nil || message.Body == nil || message.Body.Content == nil {
		return
	}

	// 2. 解析组播消息
	multicastMsgInfo := new(go_protobuf.MulticastMsg)
	err := proto.Unmarshal(*message.Body.Content, multicastMsgInfo)
	if err != nil {
		utils.Log.Error().Msgf("解析组播消息失败 err:%s", err)
		return
	}
	// 2.1 验证解析的组播消息参数信息
	if multicastMsgInfo.MsgID == 0 || len(multicastMsgInfo.Nodes) == 0 {
		utils.Log.Error().Msgf("组播消息参数有误 msgID:%d nodeLen:%d", multicastMsgInfo.MsgID, len(multicastMsgInfo.Nodes))
		return
	}

	// 3. 依次发送消息
	var content []byte
	if multicastMsgInfo.Content != "" {
		content = []byte(multicastMsgInfo.Content)
	}
	for i := range multicastMsgInfo.Nodes {
		nodeId := nodeStore.AddressNet(multicastMsgInfo.Nodes[i].Id)
		exist, proxyInfoes := gm.Area.ProxyData.GetNodeIdProxy2(&nodeId, multicastMsgInfo.Nodes[i].MachineID)
		if !exist || len(proxyInfoes) == 0 {
			// 没有代理, 直接发送消息
			utils.Log.Error().Msgf("tId:%s tmid:%s 没有代理", nodeId.B58String(), multicastMsgInfo.Nodes[i].MachineID)
			gm.Area.SendP2pMsgProxyMachineID(multicastMsgInfo.MsgID, &nodeId, nil, gm.Area.GodID, multicastMsgInfo.Nodes[i].MachineID, &content)
			continue
		}

		// 根据代理信息依次发送消息
		for ii := range proxyInfoes {
			utils.Log.Error().Msgf("tId:%s tmid:%s 代理: %s", nodeId.B58String(), proxyInfoes[ii].MachineId, proxyInfoes[ii].ProxyId.B58String())
			gm.Area.SendP2pMsgProxyMachineID(multicastMsgInfo.MsgID, &nodeId, proxyInfoes[ii].ProxyId, gm.Area.GodID, proxyInfoes[ii].MachineId, &content)
		}
	}

	// 4. 回复消息
	gm.Area.SendP2pReplyMsg(message, MSGID_P2P_SEND_MULTICAST_MSG_RECV, nil)
}

// 发送组播消息的返回
func (gm *MulticastMsg) sendMulticastMsg_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}
