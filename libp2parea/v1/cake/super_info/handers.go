package superinfo

import (
	"web3_gui/libp2parea/v1/engine"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

// 注册消息
func (si *SuperInfo) registerMsg() {
	if si == nil || si.Area == nil || !si.initialized {
		return
	}

	si.Area.Register_p2p(MSGID_SUPER_INFO_MULTICAST_MSG, si.getSuperOnlineHeartMulti)
}

// 接收上线广播信息
func (si *SuperInfo) getSuperOnlineHeartMulti(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// 1. 判断消息内容
	if message == nil || message.Body == nil || message.Body.Content == nil {
		return
	}

	// 2. 解析节点信息
	newNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	if err != nil {
		return
	}

	// 3. 处理上线信息
	// utils.Log.Error().Msgf("接收到节点上线广播 nodeId:%s", newNode.IdInfo.Id.B58String())
	if !newNode.IsApp && newNode.GetIsSuper() {
		// 保存超级节点上线信息
		newNode.FlashOnlineTime()
		strKey := utils.Bytes2string(newNode.IdInfo.Id)
		// utils.Log.Error().Msgf("更新节点在线时间 nodeID:%s", newNode.IdInfo.Id.B58String())
		_, exist := si.SuperNodes.Load(strKey)
		si.SuperNodes.Store(strKey, newNode)

		// 新的节点上线, 调用上线回调
		if !exist {
			for _, h := range si.superNodeOnlineCallbackFunc {
				go h(newNode.IdInfo.Id, newNode.MachineIDStr)
			}
		}
	}
}

// 新连接回调
func (si *SuperInfo) newConnCallbackFunc(addr nodeStore.AddressNet, machineID string) {
	// 1. superinfo有效性判断
	if si == nil || !si.initialized || si.Area == nil {
		return
	}
	// 1.1 判断是不是超级节点
	if !si.Area.NodeManager.NodeSelf.GetIsSuper() {
		return
	}

	// 2. 获取节点信息
	node := si.Area.NodeManager.FindNode(&addr)
	if node == nil {
		return
	}

	// 3. 判断节点是否为超级节点
	if !node.GetIsSuper() {
		return
	}

	// 4. 保存超级节点上线信息
	node.FlashOnlineTime()
	strKey := utils.Bytes2string(node.IdInfo.Id)
	// utils.Log.Error().Msgf("更新节点在线时间 nodeID:%s", newNode.IdInfo.Id.B58String())
	_, exist := si.SuperNodes.Load(strKey)
	si.SuperNodes.Store(strKey, node)

	// 新的节点上线, 调用上线回调
	if !exist {
		for _, h := range si.superNodeOnlineCallbackFunc {
			go h(node.IdInfo.Id, node.MachineIDStr)
		}
	}
}
