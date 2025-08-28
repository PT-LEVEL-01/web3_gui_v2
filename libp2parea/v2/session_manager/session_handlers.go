package session_manager

import (
	"bytes"
	"time"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
注册消息
*/
func (this *SessionManager) RegisterHanders() {
	this.sessionEngine.RegisterMsg(config.MSGID_exchange_nodeinfo, this.exchangeNodeInfo) //连上后互相交换节点信息
	this.sessionEngine.RegisterMsg(config.MSGID_register_node, this.registerNode)         //首次连接注册节点
	this.sessionEngine.RegisterMsg(config.MSGID_searchAreaIP, this.searchAreaIP)          //查找自己的域网络
	this.sessionEngine.RegisterMsg(config.MSGID_getLogicAreaIP, this.getLogicArea)        //获取自己的逻辑域节点地址
	this.sessionEngine.RegisterMsg(config.MSGID_ask_close_conn, this.askCloseNode)        //询问关闭连接
	this.sessionEngine.RegisterMsg(config.MSGID_getNearSuperIP, this.findSuperID)         //从邻居节点得到自己的逻辑节点
	this.sessionEngine.RegisterMsg(config.MSGID_send_new_node, this.sendNewNode)          //向邻居节点推送新节点
}

/*
节点之间互相交换信息
*/
func (this *SessionManager) exchangeNodeInfo(msg *engine.Packet) {
	//var ERR utils.ERROR
	//this.Log.Info().Str("交换节点信息", "11111111").Send()
	np, err := utils.ParseNetParams(msg.Data)
	if err != nil {
		time.Sleep(time.Second * 10)
		this.Log.Error().Str("解析参数错误", err.Error()).Send()
		ERR := utils.NewErrorBus(config.ERROR_code_params_fail, "")
		//ERR := utils.NewErrorBus(config.ERROR_code_nodeinfo_illegal, "")
		this.replySession(msg, nil, ERR)
		return
	}
	//解析数据流
	nodeRemote, err := nodeStore.ParseNodeProto(np.Data)
	if err != nil {
		time.Sleep(time.Second * 10)
		this.Log.Error().Str("解析参数错误", err.Error()).Send()
		ERR := utils.NewErrorBus(config.ERROR_code_params_fail, "")
		//ERR := utils.NewErrorBus(config.ERROR_code_nodeinfo_illegal, "")
		this.replySession(msg, nil, ERR)
		return
	}
	//验证节点合法性
	ok, ERR := nodeRemote.Validate()
	if ERR.CheckFail() {
		time.Sleep(time.Second * 10)
		this.Log.Error().Str("节点信息不合法", "").Send()
		ERR = utils.NewErrorBus(config.ERROR_code_nodeinfo_illegal, "")
		this.replySession(msg, nil, ERR)
		return
	}
	if !ok {
		//不合法
		time.Sleep(time.Second * 10)
		//ERR = utils.NewErrorBus(config.ERROR_code_params_fail, "")
		this.Log.Error().Str("节点信息不合法", "").Send()
		ERR = utils.NewErrorBus(config.ERROR_code_nodeinfo_illegal, "")
		this.replySession(msg, nil, ERR)
		return
	}
	//this.Log.Info().Interface("registerNode", nodeRemote).Send()
	_, ERR = engine.CheckAddr(nodeRemote.RemoteMultiaddr)
	if ERR.CheckFail() {
		time.Sleep(time.Second * 10)
		this.Log.Error().Str("节点信息不合法", "").Send()
		ERR = utils.NewErrorBus(config.ERROR_code_nodeinfo_illegal, "")
		this.replySession(msg, nil, ERR)
		return
	}
	//this.Log.Info().Str("交换节点信息", "11111111").Send()

	//节点信息放入session
	SetNodeInfo(msg.Session, nodeRemote)

	//this.Log.Info().Str("交换节点信息", "11111111").Send()
	//判断是否自己连接自己
	connSelf := bytes.Equal(this.nodeManager.NodeSelf.AreaName, nodeRemote.AreaName) &&
		bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), nodeRemote.IdInfo.Id.Data()) &&
		bytes.Equal(this.nodeManager.NodeSelf.MachineID, nodeRemote.MachineID)
	if !connSelf {
		//检查并添加自己的公开端口
		this.checkSelfOpenAddr(nodeRemote)

		//不是自己连接自己，才能反向连接
		//检查对方是否可以连接
		connectCan, _, addrInfo, ERR := this.checkNodeCanConnect(msg.Session, nodeRemote)
		if ERR.CheckSuccess() && connectCan {
			SetSessionCanConnPort(msg.Session, addrInfo)
		}

	}

	//this.Log.Info().Str("交换节点信息", "11111111").Send()

	//给对方回复自己的节点信息
	bs, err := this.nodeManager.NodeSelf.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		ERR = utils.NewErrorBus(config.ERROR_code_params_fail, "")
		return
	}
	ERR = utils.NewErrorSuccess()
	this.replySession(msg, bs, ERR)
}

func (this *SessionManager) replySession(msg *engine.Packet, resultBs *[]byte, ERR utils.ERROR) {
	bs := &[]byte{}
	if resultBs != nil {
		bs = resultBs
	}
	//ERR = utils.NewErrorSuccess()
	nr := utils.NewNetResult(config.Version_1, ERR.Code, ERR.Msg, *bs)
	bs, err := nr.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		ERR = utils.NewErrorBus(config.ERROR_code_params_fail, "")
		return
	}
	ERR = msg.Session.Reply(msg, bs, 0)
	if ERR.CheckFail() {
		//this.Log.Error().Str("ERR", ERR.String()).Send()
	}
}

/*
注册节点
*/
func (this *SessionManager) registerNode(msg *engine.Packet) {
	nodeRemote := GetNodeInfo(msg.Session)
	//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()
	//判断是否自己连接自己
	if bytes.Equal(this.nodeManager.NodeSelf.AreaName, nodeRemote.AreaName) &&
		bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), nodeRemote.IdInfo.Id.Data()) &&
		bytes.Equal(this.nodeManager.NodeSelf.MachineID, nodeRemote.MachineID) {
		//不允许自己注册自己
		//this.FreeSession.Store(utils.Bytes2string(msg.Session.GetId()), msg.Session)
		//this.nodeManager.AddNodeInfoFree(nodeRemote)
		return
	}
	//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()
	//this.CloseConnLock.Lock()
	//defer this.CloseConnLock.Unlock()
	var ERR utils.ERROR
	defer func() {
		//注册后，无论如何都要从游客列表中删除
		//delete(this.guestSession, utils.Bytes2string(msg.Session.GetId()))
		this.guestSession.Delete(utils.Bytes2string(msg.Session.GetId()))
		if ERR.CheckSuccess() {
			return
		}
		//当注册失败会关闭这个会话
		//this.Log.Info().Str("关闭连接", ERR.String()).Send()
		msg.Session.Close()
	}()
	//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()
	_, err := utils.ParseNetParams(msg.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		ERR = utils.NewErrorBus(config.ERROR_code_params_fail, "")
		return
	}

	//远端是通过内网还是外网发起的连接
	addrInfoSelf, ERR := engine.CheckAddr(nodeRemote.RemoteMultiaddr)
	if ERR.CheckFail() {
		return
	}
	isOnlyIp := addrInfoSelf.IsOnlyIp()
	//对方是否能反向连接
	connectCan := false
	//远端端口信息
	addrInfoRemoate := GetSessionCanConnPort(msg.Session)
	if addrInfoRemoate != nil {
		connectCan = true
	}

	ERR = utils.NewErrorSuccess()
	//session := msg.Session
	//检查对方是否可以连接
	//connectCan, isOnlyIp, addrInfo, ERR := this.checkNodeCanConnect(session, nodeRemote)
	//if ERR.CheckSuccess() {
	if connectCan {
		//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()
		//检查是否有重复连接
		exist := false
		exist, ERR = this.ConnStatus.LoadOrStore(addrInfoRemoate.Multiaddr, msg.Session.GetId())
		if ERR.CheckSuccess() {
			//重复的连接
			if exist {
				//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()
				ERR = utils.NewErrorBus(config.ERROR_code_Repeated_connections, "")
			}
		}
	}
	//成功或者不是重复的连接
	if ERR.CheckSuccess() || ERR.Code != config.ERROR_code_Repeated_connections {
		this.ClassificationSession(msg.Session, nodeRemote, connectCan, isOnlyIp, addrInfoRemoate)
	}
	//}

	//节点信息合法，注册成功

	//this.ClassificationSession(msg.Session, nodeRemote)

	//this.Log.Info().Str("注册节点", "11111111").Send()
	//this.Log.Info().Str("registerNode放入", "11111111").Send()

	//给对方回复自己的节点信息
	//bs, err := this.nodeManager.NodeSelf.Proto()
	//if err != nil {
	//	this.Log.Error().Err(err).Send()
	//	ERR = utils.NewErrorBus(config.ERROR_code_params_fail, "")
	//	return
	//}
	//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()

	nr := utils.NewNetResult(config.Version_1, ERR.Code, ERR.Msg, nil)
	bs, err := nr.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		ERR = utils.NewErrorBus(config.ERROR_code_params_fail, "")
		return
	}
	//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()
	ERR = msg.Session.Reply(msg, bs, 0)
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	ERR = utils.NewErrorSuccess()
	//this.Log.Info().Str("注册节点", nodeRemote.IdInfo.Id.B58String()).Send()
	//this.Log.Info().Str("注册节点", "11111111").Send()
}

/*
查找自己的域网络
*/
func (this *SessionManager) searchAreaIP(msg *engine.Packet) {
	_, err := utils.ParseNetParams(msg.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	//缓存中查找节点信息
	nodeInfo := GetNodeInfo(msg.Session)
	if nodeInfo == nil {
		return
	}
	//先查找超级节点的地址，查找指定域节点
	nodeInfosWAN := this.nodeManager.AreaLogicAddrWAN.FindNearAreaAddr(nodeInfo.AreaName)
	if len(nodeInfosWAN) == 0 {
		nodeInfosWAN = this.nodeManager.AreaLogicAddrWAN.FindNearAreaAddr(nodeInfo.AreaName)
	}
	nodeWANBs, err := nodeStore.NewNodeInfoList(nodeInfosWAN).Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}

	//再查找局域网节点的地址，查找指定域节点
	nodeInfosLAN := this.nodeManager.AreaLogicAddrLAN.FindNearAreaAddr(nodeInfo.AreaName)
	if len(nodeInfosLAN) == 0 {
		nodeInfosLAN = this.nodeManager.AreaLogicAddrLAN.FindNearAreaAddr(nodeInfo.AreaName)
	}
	nodeLANBs, err := nodeStore.NewNodeInfoList(nodeInfosLAN).Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	listBs := [][]byte{*nodeWANBs, *nodeLANBs}
	bs, err := config.ByteListProto([][][]byte{listBs})
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	nr := utils.NewNetResult(config.Version_1, utils.ERROR_CODE_success, "", bs)
	bs2, err := nr.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	//this.Log.Info().Int("发送的返回大小", len(*bs2)).Send()
	ERR := msg.Session.Reply(msg, bs2, 0)
	if ERR.CheckFail() {
		//this.Log.Error().Str("ERR", ERR.String()).Send()
	}

	//_, err := utils.ParseNetParams(msg.Data)
	//if err != nil {
	//	this.Log.Error().Err(err).Send()
	//	return
	//}
	////缓存中查找节点信息
	//nodeInfo := GetNodeInfo(msg.Session)
	//if nodeInfo == nil {
	//	return
	//}
	////本节点中查找指定域节点
	//nodeInfos := this.nodeManager.AreaLogicAddrWAN.FindNearAreaAddr(nodeInfo.AreaName)
	//nodeInfoBs := make([][]byte, 0)
	//for _, one := range nodeInfos {
	//	bs, err := one.Proto()
	//	if err != nil {
	//		continue
	//	}
	//	nodeInfoBs = append(nodeInfoBs, *bs)
	//}
	//bs, err := config.ByteListProto([][][]byte{nodeInfoBs})
	//if err != nil {
	//	this.Log.Error().Str("解析参数错误", err.Error()).Send()
	//	return
	//}
	//ERR := utils.NewErrorSuccess()
	//nr := utils.NewNetResult(config.Version_1, ERR.Code, ERR.Msg, bs)
	//bs2, err := nr.Proto()
	//if err != nil {
	//	this.Log.Error().Err(err).Send()
	//	return
	//}
	//ERR = msg.Session.Reply(msg, bs2, 0)
	//if ERR.CheckFail() {
	//	//this.Log.Error().Str("ERR", ERR.String()).Send()
	//}
	//return
}

/*
获取其他域节点
*/
func (this *SessionManager) getLogicArea(msg *engine.Packet) {
	nodeInfoRemote := GetNodeInfo(msg.Session)
	if nodeInfoRemote == nil {
		return
	}
	//先计算WAN网络中的域名称
	nodeInfos := this.nodeManager.AreaLogicAddrWAN.FindNearAreaAddr(nodeInfoRemote.AreaName)
	//nodeInfosWAN := this.nodeManager.GetLogicNodesWAN()
	kad := nodeStore.NewKademlia(nodeInfoRemote.AreaName, nodeStore.NodeIdLevel, this.Log)
	for _, one := range nodeInfos {
		if bytes.Equal(one.IdInfo.Id.Data(), nodeInfoRemote.IdInfo.Id.Data()) {
			continue
		}
		kad.AddId(one.AreaName)
	}
	//游离的超级节点放进来计算
	nodeInfosFree := this.nodeManager.GetFreeNodes()
	for _, one := range nodeInfosFree {
		if bytes.Equal(one.IdInfo.Id.Data(), nodeInfoRemote.IdInfo.Id.Data()) {
			continue
		}
		kad.AddId(one.AreaName)
	}
	areaNamesWAN := kad.GetIds()
	//
	nodeInfosWAN := make([]nodeStore.NodeInfo, 0)
	for _, one := range areaNamesWAN {
		nodeInfos := this.nodeManager.AreaLogicAddrWAN.FindNearAreaAddr(one)
		nodeInfosWAN = append(nodeInfosWAN, nodeInfos...)
	}
	//序列化
	nodeWANBs, err := nodeStore.NewNodeInfoList(nodeInfosWAN).Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}

	//再计算局域网节点地址
	nodeInfos = this.nodeManager.AreaLogicAddrLAN.FindNearAreaAddr(nodeInfoRemote.AreaName)
	//nodeInfosLAN := this.nodeManager.GetLogicNodesLAN()
	kad = nodeStore.NewKademlia(nodeInfoRemote.AreaName, nodeStore.NodeIdLevel, this.Log)
	for _, one := range nodeInfos {
		if bytes.Equal(one.IdInfo.Id.Data(), nodeInfoRemote.IdInfo.Id.Data()) {
			continue
		}
		kad.AddId(one.AreaName)
	}
	areaNamesLAN := kad.GetIds()
	//
	nodeInfosLAN := make([]nodeStore.NodeInfo, 0)
	for _, one := range areaNamesLAN {
		nodeInfos := this.nodeManager.AreaLogicAddrLAN.FindNearAreaAddr(one)
		nodeInfosLAN = append(nodeInfosLAN, nodeInfos...)
	}
	//序列化
	nodeLANBs, err := nodeStore.NewNodeInfoList(nodeInfosLAN).Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	listBs := [][]byte{*nodeWANBs, *nodeLANBs}
	bs, err := config.ByteListProto([][][]byte{listBs})
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	nr := utils.NewNetResult(config.Version_1, utils.ERROR_CODE_success, "", bs)
	bs2, err := nr.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	//this.Log.Info().Int("发送的返回大小", len(*bs2)).Send()
	ERR := msg.Session.Reply(msg, bs2, 0)
	if ERR.CheckFail() {
		//this.Log.Error().Str("ERR", ERR.String()).Send()
	}
	return
	//
	//_, err := utils.ParseNetParams(msg.Data)
	//if err != nil {
	//	this.Log.Error().Err(err).Send()
	//	return
	//}
	////缓存中查找节点信息
	//nodeInfo := GetNodeInfo(msg.Session)
	//if nodeInfo == nil {
	//	return
	//}
	//nodeInfos := this.nodeManager.AreaLogicAddrWAN.FindNearAreaAddr(nodeInfo.AreaName)
	//
	//nodeStore.NewNodeInfoList(nodeInfos)
	//
	//nodeInfoBs := make([][]byte, 0)
	//for _, one := range nodeInfos {
	//	bs, err := one.Proto()
	//	if err != nil {
	//		continue
	//	}
	//	nodeInfoBs = append(nodeInfoBs, *bs)
	//}
	//bs, err := config.ByteListProto([][][]byte{nodeInfoBs})
	//if err != nil {
	//	this.Log.Error().Str("解析参数错误", err.Error()).Send()
	//	return
	//}
	//ERR := utils.NewErrorSuccess()
	//nr := utils.NewNetResult(config.Version_1, ERR.Code, ERR.Msg, bs)
	//bs2, err := nr.Proto()
	//if err != nil {
	//	this.Log.Error().Err(err).Send()
	//	return
	//}
	//ERR = msg.Session.Reply(msg, bs2, 0)
	//if ERR.CheckFail() {
	//	//this.Log.Error().Str("ERR", ERR.String()).Send()
	//}
	//return
}

/*
询问关闭节点
*/
func (this *SessionManager) askCloseNode(msg *engine.Packet) {
	nodeInfoRemote := GetNodeInfo(msg.Session)
	if nodeInfoRemote == nil {
		msg.Session.Close()
		return
	}
	//this.Log.Info().Str("询问关闭", "1111111111").Send()
	//nodeRemote := GetNodeInfo(msg.Session)
	//if nodeRemote != nil {
	//	this.Log.Info().Str("handler 询问关闭 self", this.nodeManager.NodeSelf.IdInfo.Id.B58String()).
	//		Str("询问关闭 remote", nodeRemote.IdInfo.Id.B58String()).Send()
	//	//this.Log.Info().Str("询问关闭", nodeRemote.IdInfo.Id.B58String()).Hex("sid", message.GetPacket().Session.GetId()).Send()
	//}
	_, err := utils.ParseNetParams(msg.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	nodeInfo := this.nodeManager.FindNodeInfoProxyByAddr(nodeInfoRemote.IdInfo.Id)
	if nodeInfo != nil {
		for _, one := range nodeInfo.GetSessions() {
			if one != nil && bytes.Equal(one.GetId(), msg.Session.GetId()) {
				//this.Log.Info().Str("关闭连接，被代理的节点连接可以关闭", nodeRemote.IdInfo.Id.B58String()).Send()
				msg.Session.Close()
				return
			}
		}
	}
	//this.Log.Info().Str("询问关闭", nodeRemote.IdInfo.Id.B58String()).Hex("sid", message.GetPacket().Session.GetId()).Send()
	nodeInfo = this.nodeManager.FindNodeInfoFreeByAddr(nodeInfoRemote.IdInfo.Id)
	if nodeInfo != nil {
		for _, one := range nodeInfo.GetSessions() {
			if one != nil && bytes.Equal(one.GetId(), msg.Session.GetId()) {
				//this.Log.Info().Str("关闭连接，游离的节点连接可以关闭", nodeRemote.IdInfo.Id.B58String()).Send()
				msg.Session.Close()
				return
			}
		}
	}
	//this.Log.Info().Str("询问关闭", nodeRemote.IdInfo.Id.B58String()).Hex("sid", message.GetPacket().Session.GetId()).Send()
}

/*
查询发送者地址的逻辑节点地址
*/
func (this *SessionManager) findSuperID(msg *engine.Packet) {
	nodeInfoRemote := GetNodeInfo(msg.Session)
	if nodeInfoRemote == nil {
		return
	}
	//先计算超级节点的地址
	nodeInfosWAN := this.nodeManager.GetLogicNodesWAN()
	kad := nodeStore.NewKademlia(nodeInfoRemote.IdInfo.Id.Data(), nodeStore.NodeIdLevel, this.Log)
	for _, one := range nodeInfosWAN {
		if bytes.Equal(one.IdInfo.Id.Data(), nodeInfoRemote.IdInfo.Id.Data()) {
			continue
		}
		kad.AddId(one.IdInfo.Id.Data())
	}
	nodeInfosFree := this.nodeManager.GetFreeNodes()
	for _, one := range nodeInfosFree {
		if bytes.Equal(one.IdInfo.Id.Data(), nodeInfoRemote.IdInfo.Id.Data()) {
			continue
		}
		kad.AddId(one.IdInfo.Id.Data())
	}
	superAddrs := kad.GetIds()

	nodeInfoMap := make(map[string]*nodeStore.NodeInfo)
	for _, one := range append(nodeInfosWAN, nodeInfosFree...) {
		nodeInfoMap[utils.Bytes2string(one.IdInfo.Id.Data())] = &one
	}
	nodeInfosWAN = make([]nodeStore.NodeInfo, 0, len(superAddrs))
	for _, one := range superAddrs {
		nodeInfo, ok := nodeInfoMap[utils.Bytes2string(one)]
		if ok {
			nodeInfosWAN = append(nodeInfosWAN, *nodeInfo)
		}
	}
	nodeWANBs, err := nodeStore.NewNodeInfoList(nodeInfosWAN).Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}

	//再计算局域网节点地址
	nodeInfosLAN := this.nodeManager.GetLogicNodesLAN()
	kad = nodeStore.NewKademlia(nodeInfoRemote.IdInfo.Id.Data(), nodeStore.NodeIdLevel, this.Log)
	for _, one := range nodeInfosLAN {
		if bytes.Equal(one.IdInfo.Id.Data(), nodeInfoRemote.IdInfo.Id.Data()) {
			continue
		}
		kad.AddId(one.IdInfo.Id.Data())
	}
	lanAddrs := kad.GetIds()
	nodeInfoMap = make(map[string]*nodeStore.NodeInfo)
	for _, one := range nodeInfosLAN {
		nodeInfoMap[utils.Bytes2string(one.IdInfo.Id.Data())] = &one
	}
	nodeInfosLAN = make([]nodeStore.NodeInfo, 0, len(lanAddrs))
	for _, one := range lanAddrs {
		nodeInfo, ok := nodeInfoMap[utils.Bytes2string(one)]
		if ok {
			nodeInfosLAN = append(nodeInfosLAN, *nodeInfo)
		}
	}
	nodeLANBs, err := nodeStore.NewNodeInfoList(nodeInfosLAN).Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	listBs := [][]byte{*nodeWANBs, *nodeLANBs}
	bs, err := config.ByteListProto([][][]byte{listBs})
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	nr := utils.NewNetResult(config.Version_1, utils.ERROR_CODE_success, "", bs)
	bs2, err := nr.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	//this.Log.Info().Int("发送的返回大小", len(*bs2)).Send()
	ERR := msg.Session.Reply(msg, bs2, 0)
	if ERR.CheckFail() {
		//this.Log.Error().Str("ERR", ERR.String()).Send()
	}
}

/*
邻居节点推送过来的新节点
*/
func (this *SessionManager) sendNewNode(msg *engine.Packet) {
	np, err := utils.ParseNetParams(msg.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	nodes, err := nodeStore.ParseNodesProto(&np.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	for _, one := range nodes {
		//推送过来的是不是相同域网络
		if bytes.Equal(one.AreaName, this.nodeManager.NodeSelf.AreaName) {
			//是相同域
			mlan := one.GetMultiaddrLAN()
			if len(mlan) > 0 {
				need, _, _ := this.nodeManager.CheckNeedNodeLan(&one, false)
				if need {
					for _, one := range mlan {
						this.Log.Info().Str("邻居节点推送过来的节点 LAN", one.Multiaddr.String()).Send()
						go this.connectNet(one)
					}
				}
			} else {
				//this.Log.Info().Str("推送的节点信息中没有地址", "LAN").Send()
			}
			mwan := one.GetMultiaddrWAN()
			if len(mwan) > 0 {
				//this.Log.Info().Str("邻居节点推送过来的节点 WAN", one.IdInfo.Id.B58String()).Send()
				need, _, _ := this.nodeManager.CheckNeedNodeWan(&one, false)
				if need {
					for _, one := range mwan {
						this.Log.Info().Str("邻居节点推送过来的节点 WAN", one.Multiaddr.String()).Send()
						go this.connectNet(one)
					}
				}
			} else {
				//this.Log.Info().Str("推送的节点信息中没有地址", "WAN").Send()
			}
		} else {
			//不是相同域
			mlan := one.GetMultiaddrLAN()
			if len(mlan) > 0 {
				need, _, _ := this.nodeManager.CheckNeedAreaLan(&one, false)
				if need {
					for _, one := range mlan {
						this.Log.Info().Str("邻居节点推送过来的节点 LAN", one.Multiaddr.String()).Send()
						go this.connectNet(one)
					}
				}
			} else {
				//this.Log.Info().Str("推送的节点信息中没有地址", "LAN").Send()
			}
			mwan := one.GetMultiaddrWAN()
			if len(mwan) > 0 {
				//this.Log.Info().Str("邻居节点推送过来的节点 WAN", one.IdInfo.Id.B58String()).Send()
				need, _, _ := this.nodeManager.CheckNeedAreaWan(&one, false)
				if need {
					for _, one := range mwan {
						this.Log.Info().Str("邻居节点推送过来的节点 WAN", one.Multiaddr.String()).Send()
						go this.connectNet(one)
					}
				}
			} else {
				//this.Log.Info().Str("推送的节点信息中没有地址", "WAN").Send()
			}

		}

	}
}
