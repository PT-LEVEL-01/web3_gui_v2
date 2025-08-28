package multicast

import (
	"bytes"
	"time"

	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/cake/country"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/protobuf/go_protobuf"
)

// 组播消息
type MulticastMsg struct {
	Area        *libp2parea.Area     // 节点信息
	Country     *country.AreaCountry // 大区信息
	initialized bool                 // 是否初始化过
}

/*
 * 创建组播消息
 *
 * @param	area	*Area			所属区域
 * @param	country	*AreaCountry	大区信息
 * @return	mm		*MulticastMsg	群消息
 */
func NewMulticastMsg(area *libp2parea.Area, country *country.AreaCountry) (gm *MulticastMsg) {
	// 判断参数的合法性
	if area == nil {
		utils.Log.Error().Msgf("节点信息为空, 请传入有效的节点信息!!!!!!")
		return nil
	}
	if area.NodeManager.NodeSelf.GetIsSuper() {
		if country == nil {
			utils.Log.Error().Msgf("大区信息为空, 请传入有效的大区信息!!!!!!")
			return nil
		}

		// 启动大区, 确保大区一定启动
		country.Start()
	}

	// 构建群消息
	gm = &MulticastMsg{Area: area, Country: country}

	// 初始化群消息
	gm.init()

	return
}

// 初始化
func (gm *MulticastMsg) init() {
	// 判断是否初始化过
	if gm.initialized {
		return
	}
	gm.initialized = true

	// 注册消息
	gm.registerMsg()

	// 获取到大区中保存的节点地址
	if gm.Area.NodeManager.NodeSelf.GetIsSuper() {
		// 只有超级节点才会执行该操作
		go gm.getAreaNodeIds()
	}
}

/*
 * 获取大区中记录的
 */
func (gm *MulticastMsg) getAreaNodeIds() {
	if gm == nil {
		return
	}

	// 构建定时器
	timer := time.NewTicker(time.Second * 5)
	defer timer.Stop()

	// 定时获取一次存储的节点地址, 直到获取到为止
	for range timer.C {
		// 如果还没初始化, 或大区信息为空, 则等待下一循环
		if !gm.initialized || gm.Country == nil {
			continue
		}

		// 获取到大区保存的节点列表
		saveIds, err := gm.Country.GetAreaSaveNodeIds()
		if err != nil || len(saveIds) == 0 {
			continue
		}

		// 注册代理存储节点白名单
		if !gm.Area.ProxyDetail.AddWhiteNodes(saveIds) {
			continue
		}

		// 退出循环
		break
	}
}

/*
 * 发送组播消息
 *
 * @param	msgId		uint64			消息号
 * @param	nodeIds		[]*AddressNet	接收端的地址列表
 * @param	content		string			消息内容
 */
func (gm *MulticastMsg) SendMulticastMsg(msgId uint64, nodeIds []*NodeMachineInfo, content string, timeout time.Duration) error {
	// 1. 检查是否初始化
	if gm == nil || !gm.initialized {
		return Err_MULTICAST_MSG_NOT_INIT
	}
	// 2. 检查参数
	if len(nodeIds) == 0 {
		return Err_MULTICAST_NODES_EMPTY
	}
	if msgId <= 0 {
		return Err_MULTICAST_MSGID_INVALID
	}

	// 3. 构建请求体
	multicastMsgInfo := new(go_protobuf.MulticastMsg)
	multicastMsgInfo.MsgID = msgId
	multicastMsgInfo.Content = content
	for i := range nodeIds {
		if len(nodeIds[i].NodeId) == 0 {
			continue
		}

		// 构建节点信息
		var node go_protobuf.NodeInfo
		node.Id = nodeIds[i].NodeId
		node.MachineID = nodeIds[i].MachineID

		multicastMsgInfo.Nodes = append(multicastMsgInfo.Nodes, &node)
	}

	// 4. 把内容转换成字节数组
	bs, err := multicastMsgInfo.Marshal()
	if err != nil {
		return err
	}

	// 6. 获取大区对应的节点列表
	areaNodeIds, err := gm.Country.GetAreaSaveNodeIds()
	if err != nil {
		return err
	}
	if len(areaNodeIds) == 0 {
		return Err_MULTICAST_COUNTRY_NODE_EMPTY
	}

	// 7. 选择一个大区地址
	recvId := areaNodeIds[0]
	if gm.Area.GodID != nil {
		for i := range areaNodeIds {
			if !bytes.Equal(*gm.Area.GodID, areaNodeIds[i]) {
				continue
			}

			recvId = *gm.Area.GodID
			break
		}
	} else if gm.Area.NodeManager.NodeSelf.GetIsSuper() {
		for i := range areaNodeIds {
			if !bytes.Equal(gm.Area.GetNetId(), areaNodeIds[i]) {
				continue
			}

			recvId = gm.Area.GetNetId()
			break
		}
	}

	// 8. 判断接收地址是否合法
	if len(recvId) == 0 {
		return Err_MULTICAST_COUNTRY_NODE_EMPTY
	}

	// 9. 发送消息
	_, _, _, err = gm.Area.SendP2pMsgProxyWaitRequest(MSGID_P2P_SEND_MULTICAST_MSG, &recvId, nil, gm.Area.GodID, &bs, timeout)

	return err
}
