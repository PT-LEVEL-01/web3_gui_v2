package message_center

import (
	"bytes"
	"time"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
检查该消息是否是自己的
不是自己的则自动转发出去
@return    bool    是否自己接收。true=是自己接收;false=不是自己接收;
@return    error   发送错误信息
*/
func (this *MessageCenter) forward(message *MessageBase, first bool, timeout time.Duration) (bool, utils.ERROR) {
	//this.Log.Info().Interface("转发消息", message).Send()
	//this.Log.Info().Str("转发消息", "111111").Send()
	if this.checkSelf(message) {
		//this.Log.Info().Str("转发消息", "111111").Send()
		return true, utils.NewErrorSuccess()
	}
	//this.Log.Info().Str("转发消息", message.GetBase().RecvAddr.B58String()).Send()
	//在本节点的所有连接中查找，是否有目标地址的连接
	nodeInfo := this.nodeManager.FindNodeInfoAllByAddr(message.GetBase().RecvAddr)
	if nodeInfo != nil {
		//this.Log.Info().Int("找到目标地址，则转发", len(nodeInfo.GetSessions())).Send()
		//找到目标地址，则转发
		this.sendToSessions(nodeInfo.GetSessions(), message, timeout)
		return false, utils.NewErrorSuccess()
	}
	//this.Log.Info().Str("转发消息", "111111").Send()
	//未找到直接连接的目标节点，则在逻辑节点中查找最近的节点
	var from *nodeStore.AddressNet
	if message.GetPacket() == nil {
		//packet为空时，是自己发送出去的消息，不是转发的消息，没有from
	} else {
		fromNodeInfo := GetNodeInfo(message.GetPacket().Session)
		from = fromNodeInfo.IdInfo.Id
	}
	//from := this.SessionManager.FindAddrBySessionID(message.GetPacket().Session.GetId())
	//var targetId *nodeStore.AddressNet
	targetId, ERR := this.nodeManager.FindNearInWAN(message.GetBase().RecvAddr, from, !first)
	if ERR.CheckFail() {
		return false, ERR
	}
	if targetId == nil || bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), targetId.Data()) {
		//在局域网中查找节点
		targetId, ERR = this.nodeManager.FindNearInLAN(message.GetBase().RecvAddr, from, !first)
		if ERR.CheckFail() {
			return false, ERR
		}
		if targetId == nil {
			//查询不到其他逻辑节点，则代表自己不在线
			return false, utils.NewErrorBus(config.ERROR_code_offline, "")
		}
	}
	//if targetId != nil {
	//	this.Log.Info().Str("转发给节点", targetId.B58String()).Send()
	//}
	//if targetId != nil && bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id, *targetId) {
	//	//离此节点最近的节点就是自己，然而在本节点的代理节点中未找到目标节点，则认为此节点下线，此数据路由终止。
	//	this.Log.Info().Str("此消息丢失", "---------").Send()
	//	return false, utils.NewErrorSuccess()
	//}
	if targetId != nil && bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), targetId.Data()) {
		//离此节点最近的节点就是自己，然而在本节点的代理节点中未找到目标节点，则认为此节点下线，此数据路由终止。
		this.Log.Info().Str("此消息丢失", targetId.B58String()).Send()
		return false, utils.NewErrorSuccess()
	}

	//this.Log.Info().Str("找到目标地址，则转发", targetId.B58String()).Send()
	//转发给目标节点
	//utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
	nodeInfo = this.nodeManager.FindNodeInfoAllByAddr(targetId)
	//找到目标地址，则转发
	//this.forward(ss, message)
	if nodeInfo == nil {
		//this.Log.Info().Str("节点不在线", "").Send()
		return false, utils.NewErrorSuccess()
	}
	//this.Log.Info().Str("找到目标地址，则转发", targetId.B58String()).Send()
	this.sendToSessions(nodeInfo.GetSessions(), message, timeout)
	return false, utils.NewErrorSuccess()
}

/*
不是自己的则自动转发出去
@return    bool    是否自己接收。true=是自己接收;false=不是自己接收;
@return    error   发送错误信息
*/
//func (this *MessageCenter) p2pMessageForward(message message_bean.MessageItr) (bool, utils.ERROR) {
//	if this.checkSelf(message) {
//		return
//	}
//}

/*
检查消息是否自己接收
@return    bool    是否自己接收。true=是自己接收;false=不是自己接收;
*/
func (this *MessageCenter) checkSelf(message *MessageBase) bool {
	if bytes.Equal(message.GetBase().RecvAddr.Data(), this.nodeManager.NodeSelf.IdInfo.Id.Data()) {
		return true
	}
	return false
}

/*
路由中查询消息id，并执行对应handler
*/
func (this *MessageCenter) handlerProcess(message *MessageBase) {
	//this.Log.Info().Str("收到加密等待消息", "").Send()
	if message == nil {
		return
	}
	defer utils.PrintPanicStack(this.Log)
	if len(message.ReplyID) > 0 {
		//回复消息
		engine.ResponseByteKey(config.Wait_major_p2p_sys_msg, message.SendID, &message.Content)
		return
	}
	h, ok := this.router.GetHandler(message.MsgID)
	if !ok {
		this.Log.Info().Msgf("This P2P message is not registered,msgId:%d", message.MsgID)
		return
	}
	//this.Log.Info().Str("收到加密等待消息", "").Send()
	h(message)
}

/*
转发给多个连接，其中一个发送成功，则成功
*/
func (this *MessageCenter) sendToSessions(ss []engine.Session, message *MessageBase, timeout time.Duration) utils.ERROR {
	var content *[]byte
	if message.GetPacket() == nil {
		var err error
		content, err = message.Proto()
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
	} else {
		content = &message.GetPacket().Data
	}
	//this.Log.Info().Int("会话数量", len(ss)).Send()
	//this.Log.Info().Int("消息长度", len(*bs)).Hex("发送的消息", *bs).Send()
	for _, sessionOne := range ss {
		go func() {
			//this.Log.Info().Msgf("发送给会话:%d", message.MsgEngineID)
			sessionOne.Send(message.MsgEngineID, content, timeout)
		}()
	}
	return utils.NewErrorSuccess()
}

/*
给多个会话发送同一消息并等待回复
其中一个回复则返回
*/
//func (this *MessageCenter) sendToSessionsWaitResponse(ss []engine.Session, message message_bean.MessageItr) utils.ERROR {
//	content := message.GetBase().Content
//	if message.GetPacket() == nil {
//		bs, err := message.Proto()
//		if err != nil {
//			return utils.NewErrorSysSelf(err)
//		}
//		content = *bs
//	}
//	this.Log.Info().Int("会话数量", len(ss)).Send()
//
//	for _, sessionOne := range ss {
//
//		go func() {
//			message.GetBase().SenderAddr
//
//			this.Log.Info().Interface("会话", sessionOne).Send()
//			sessionOne.SendWait()
//			sessionOne.Send(message.GetBase().MsgEngineID, &content, time.Second*30)
//		}()
//	}
//	return utils.NewErrorSuccess()
//}

//
///*
//路由搜索超级节点消息
//@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
//@return    *nodeStore.AddressNet  选出的发送/中转地址
//@return    error                  发送错误信息
//@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
//*/
//func (this *Message) routerP2P(version uint64, nodeManager *nodeStore.NodeManager,
//	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration,
//	blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
//	//p2p收消息人就是自己
//	if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
//		//		utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
//		return false, nil, nil, true
//	}
//	if this.Head.RecvProxyId != nil && len(*this.Head.RecvProxyId) != 0 {
//		// 指定了接收者的代理节点
//		if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvProxyId) {
//			// 自己是接收者代理节点，因此需要确认本地是否持有接收者的连接信息
//			// 接收者肯定不是自己，上面已做判断
//			//utils.Log.Info().Msgf("qlw----我就是接收方的代理节点: self:%s, recv:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvId.B58String())
//			_, ok := sessionEngine.GetSession(nodeManager.AreaNameSelf, utils.Bytes2string(*this.Head.RecvId))
//			if ok {
//				// 存在接收方的连接信息
//				if blockAddr[utils.Bytes2string(*this.Head.RecvId)] <= engine.MaxRetryCnt {
//					return true, this.Head.RecvId, nil, true
//				}
//			}
//
//			// 没有接收方的连接信息，则报错，丢失消息
//			return true, nil, errors.New("don't have recv node connection"), true
//		}
//		// 自己不是接收者的代理节点，查看自己是否有和代理节点直接连接的信息
//		_, ok := sessionEngine.GetSession(nodeManager.AreaNameSelf, utils.Bytes2string(*this.Head.RecvProxyId))
//		if ok {
//			//utils.Log.Info().Msgf("p2p消息找到目标代理节点并发送出去了self:%s recv:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvProxyId.B58String())
//			if blockAddr[utils.Bytes2string(*this.Head.RecvProxyId)] <= engine.MaxRetryCnt {
//				return true, this.Head.RecvProxyId, nil, true
//			}
//		}
//	} else {
//		_, ok := sessionEngine.GetSession(nodeManager.AreaNameSelf, utils.Bytes2string(*this.Head.RecvId))
//		if ok {
//			// utils.Log.Info().Msgf("p2p消息找到目标节点并发送出去了self:%s recv:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), this.Head.RecvId.B58String())
//			if blockAddr[utils.Bytes2string(*this.Head.RecvId)] <= engine.MaxRetryCnt {
//				return true, this.Head.RecvId, nil, true
//			}
//		}
//	}
//
//	var targetId *nodeStore.AddressNet
//	if from == nil {
//		if this.Head.RecvProxyId == nil || len(*this.Head.RecvProxyId) == 0 {
//			// 如果没有指定接收者代理节点，则直接根据接收方id获取当前最近的节点id
//			targetId = nodeManager.FindNearInSuper(this.Head.RecvId, from, false, blockAddr)
//		} else {
//			// 指定了接收者代理节点，则根据接收方代理节点id获取最近的节点id
//			targetId = nodeManager.FindNearInSuper(this.Head.RecvProxyId, from, false, blockAddr)
//		}
//		// utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
//		return true, targetId, nil, false
//	} else {
//		if this.Head.RecvProxyId == nil || len(*this.Head.RecvProxyId) == 0 {
//			// 如果没有指定接收者代理节点，则直接根据接收方id获取当前最近的节点id
//			targetId = nodeManager.FindNearInSuper(this.Head.RecvId, from, true, blockAddr)
//		} else {
//			// 指定了接收者代理节点，则根据接收方代理节点id获取最近的节点id
//			targetId = nodeManager.FindNearInSuper(this.Head.RecvProxyId, from, true, blockAddr)
//		}
//		if targetId != nil && bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *targetId) {
//			// utils.Log.Info().Msgf("这个消息丢失了")
//			// utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
//			return true, nil, nil, true
//		}
//		if targetId == nil {
//			return true, nil, nil, true
//		}
//		// utils.Log.Info().Msgf("转发给self:%s targetID:%s", nodeManager.NodeSelf.IdInfo.Id.B58String(), targetId.B58String())
//		return true, targetId, nil, false
//	}
//}
//
///*
//路由搜索超级节点消息，给onebyone规则调用
//@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
//@return    *nodeStore.AddressNet  选出的发送/中转地址
//@return    error                  发送错误信息
//@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
//*/
//func (this *Message) routerP2POnebyone(version uint64, nodeManager *nodeStore.NodeManager, sessionEngine *engine.Engine,
//	vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
//	//p2p收消息人就是自己
//	if bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
//		//utils.Log.Info().Msgf("p2p收消息人就是自己self:%s", nodeManager.NodeSelf.IdInfo.Id.B58String())
//		return false, nil, nil, false
//	}
//	for _, one := range sessionEngine.GetAllDownUp(nodeManager.AreaNameSelf) {
//		if bytes.Equal([]byte(one), *this.Head.RecvId) {
//			//接收者地址时我的onebyone规则节点
//			if blockAddr[utils.Bytes2string(*this.Head.RecvId)] <= engine.MaxRetryCnt {
//				return true, this.Head.RecvId, nil, true
//			}
//		}
//	}
//
//	targetId := nodeManager.FindNearInSuperOnebyone(this.Head.RecvId, from, blockAddr, sessionEngine)
//	return true, targetId, nil, false
//
//}
//
//func (this *Message) GetSortSession(allAddr []nodeStore.AddressNet, sessionEngine *engine.Engine) ([]nodeStore.AddressNet, int) {
//	var addrKey []byte
//	var sortedr []nodeStore.AddressNet
//	var rec *big.Int
//	sortLargeNum := 0
//
//	//1. 给虚拟节点onebyone规则找 2. 给真实节点onebyone规则找
//	if this.Head.RecvVnode != nil && len(*this.Head.RecvVnode) != 0 {
//		rec = new(big.Int).SetBytes(*this.Head.RecvVnode)
//	} else {
//		rec = new(big.Int).SetBytes(*this.Head.RecvId)
//	}
//
//	sort.Sort(nodeStore.AddressBytes(allAddr))
//	for i := range allAddr {
//		addrKey = append(addrKey, allAddr[i]...)
//	}
//	addrkeyS := utils.Bytes2string(addrKey)
//
//	if sorted, ok := sessionEngine.GetSortSessionByKey(addrkeyS); ok {
//		for i := range sorted {
//			if new(big.Int).SetBytes(sorted[i]).Cmp(rec) == 1 {
//				sortLargeNum += 1
//			}
//			sortedr = append(sortedr, nodeStore.AddressNet(sorted[i]))
//		}
//		return sortedr, sortLargeNum
//	}
//
//	sortM := make(map[string]nodeStore.AddressNet)
//	onebyone := new(nodeStore.IdDESC)
//	var sortE []engine.AddressNet
//	for _, one := range allAddr {
//		sortM[one.B58String()] = one
//		*onebyone = append(*onebyone, new(big.Int).SetBytes(one))
//	}
//
//	sort.Sort(onebyone)
//
//	for i := 0; i < len(*onebyone); i++ {
//		one := (*onebyone)[i]
//		if one.Cmp(rec) == 1 {
//			sortLargeNum += 1
//		}
//		IdBs := one.Bytes()
//		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
//		sortedr = append(sortedr, sortM[nodeStore.AddressNet(*IdBsP).B58String()])
//		sortE = append(sortE, engine.AddressNet(sortM[nodeStore.AddressNet(*IdBsP).B58String()]))
//	}
//	sessionEngine.SetSortSessionByKey(addrkeyS, sortE)
//	return sortedr, sortLargeNum
//}
//
///*
//搜索虚拟节点消息,用在虚拟节点中查询中转的方法
//@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
//@return    *nodeStore.AddressNet  选出的发送/转发节点地址
//@return    error                  处理中错误
//@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
//*/
//func (this *Message) routerSearchVnodeOnebyone(version uint64, nodeManager *nodeStore.NodeManager,
//	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, timeout time.Duration,
//	blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
//	var an []virtual_node.VnodeinfoS
//	if vnodeinfo := vnc.FindInVnodeinfoSelf(*this.Head.RecvVnode); vnodeinfo != nil {
//		return false, nil, nil, true
//	}
//
//	if this.Head.RecvId == nil || len(*this.Head.RecvId) == 0 {
//		//TODO 优化消息第一次上onebyone的起始位置
//		//onebyone正常模式
//		//第一次上onebyone时，写死的第一个虚拟节点
//		var nVnodeIndex uint64
//		// 查找index最大的虚拟地址
//		vnc.RLock()
//		defer vnc.RUnlock()
//		for k := range vnc.VnodeMap {
//			if k > nVnodeIndex {
//				nVnodeIndex = k
//			} else if nVnodeIndex == 0 {
//				nVnodeIndex = k
//			}
//		}
//
//		if nVnodeIndex == 0 {
//			an = vnc.VnodeMap[nVnodeIndex].GetDownVnodeInfo()
//			an = append(an, vnc.VnodeMap[nVnodeIndex].GetUpVnodeInfo()...)
//		} else {
//			an = vnc.VnodeMap[nVnodeIndex].GetOnebyoneVnodeInfo()
//		}
//		//本地没有带虚拟空间的节点，把消息通过连接的逻辑节点传递出去
//		if len(an) <= 0 {
//			if from != nil && !bytes.Equal(*from, *this.Head.Sender) {
//				return true, nil, nil, true
//			}
//			if bytes.Equal(*this.Head.Sender, nodeManager.NodeSelf.IdInfo.Id) {
//				targetid := nodeManager.FindNearInSuper(&nodeManager.NodeSelf.IdInfo.Id, from, false, blockAddr)
//				return true, targetid, nil, false
//			} else {
//				return true, nil, nil, true
//			}
//		}
//		return this.searchMsgNextRecvVnode(vnc.VnodeMap[nVnodeIndex], &an)
//	} else if vnode := vnc.FindVnodeInAllSelf(*this.Head.RecvId); vnode != nil {
//		an := vnode.GetOnebyoneVnodeInfo()
//		return this.searchMsgNextRecvVnode(vnode, &an)
//	} else {
//		vnc.RLock()
//		defer vnc.RUnlock()
//		for i, _ := range vnc.VnodeMap {
//			if bytes.Equal(vnc.VnodeMap[i].Vnode.Vid, *this.Head.SelfVnodeId) {
//				var vnodesInfo []virtual_node.VnodeinfoS
//				//处理消息重试时，只用查询单方向连接的虚拟节点
//				if new(big.Int).SetBytes(vnc.VnodeMap[i].Vnode.Vid).Cmp(new(big.Int).SetBytes(*this.Head.RecvVnode)) == 1 {
//					an = vnc.VnodeMap[i].GetDownVnodeInfo()
//				} else {
//					an = vnc.VnodeMap[i].GetUpVnodeInfo()
//				}
//				for i, _ := range an {
//					if blockAddr[utils.Bytes2string(an[i].Vid)] > engine.MaxRetryCnt {
//						continue
//					}
//					vnodesInfo = append(vnodesInfo, an[i])
//				}
//				if len(vnodesInfo) == 0 {
//					return true, nil, nil, true
//				}
//				return this.searchMsgNextRecvVnode(vnc.VnodeMap[i], &vnodesInfo)
//			}
//		}
//	}
//	return true, nil, nil, false
//}
//
///*
//搜索真实节点消息,用onebyone在真实节点中寻找的方法
//修改版，选出 要求个数 的距离目标节点最近的点（自己 up down ...）
//
//**********************************************
//注意：此方法在消息路由到代理服务端节点以后只能再路由一次
//**********************************************
//
//@return    []nodeStore.AddressNet  选出的发送/转发节点地址数组
//@return    error                  处理中错误
//*/
//func (this *Message) routerSearchNodeOnebyone2(version uint64, nodeManager *nodeStore.NodeManager,
//	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager, blockAddr map[string]int, searchNum int) ([]nodeStore.AddressNet, error) {
//	// utils.Log.Info().Msgf("发送routerSearchVnode self:%s from:%s to:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(),
//	// 	this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())
//
//	var allAddr []nodeStore.AddressNet
//	rnode := make([]nodeStore.AddressNet, 0)
//
//	//1.1正常模式下的判断和转发
//	allAddr = append(allAddr, nodeManager.NodeSelf.IdInfo.Id)
//	ud := sessionEngine.GetAllDownUp(nodeManager.AreaNameSelf)
//	for _, v := range ud {
//		if this.Head.RecvId != nil && utils.Bytes2string(*this.Head.RecvId) == v {
//			nearId := nodeStore.AddressNet(v)
//			rnode = append(rnode, nearId)
//		}
//		if v := blockAddr[v]; v > engine.MaxRetryCnt {
//			continue
//		}
//		allAddr = append(allAddr, nodeStore.AddressNet(v))
//	}
//
//	//1.2自己不是服务端节点，给自己连接的非自己的节点发消息
//	if !nodeManager.NodeSelf.GetIsSuper() {
//		for i := range allAddr {
//			if !bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, allAddr[i]) {
//				target := allAddr[i]
//				rnode = append(rnode, target)
//				return rnode, nil
//			}
//		}
//		return rnode, errors.New("未发现服务端节点")
//	}
//
//	//2.计算接受节点和当前onebyone连接节点状态
//	sorted, times := this.GetSortSession(allAddr, sessionEngine)
//	// utils.Log.Info().Msgf("_______________________________________ len(sorted) %d, times %d len(allAddr) %d", len(sorted), times, len(allAddr))
//	//3.1当搜索地址落到比最大大 或者 比最小小 时候
//	if times == 0 || times == len(sorted) {
//		//当前6个节点，updown一共20连接情境下使用
//		if times == 0 {
//			for i := 0; i < searchNum; i++ {
//				if i < len(sorted) {
//					target := sorted[i]
//					rnode = append(rnode, target)
//				}
//			}
//		} else if times == len(sorted) {
//			for i := 0; i < searchNum; i++ {
//				if len(sorted)-1-i >= 0 {
//					target := sorted[len(sorted)-1-i]
//					rnode = append(rnode, target)
//				}
//			}
//		}
//		return rnode, nil
//	}
//
//	//3.2 当搜索地址在最大最小范围里的时候排序，根据要求个数返回
//	for i := 1; i < len(sorted)+1; i++ {
//		//找upsession，由近及远
//		if u := times - i; u >= 0 {
//			anode := sorted[u]
//			rnode = append(rnode, anode)
//		}
//		//找downsession，由近及远
//		if d := times + i - 1; d < len(sorted) {
//			anode := sorted[d]
//			rnode = append(rnode, anode)
//		}
//	}
//
//	if len(rnode) > searchNum {
//		rnode = rnode[:searchNum]
//	}
//	return rnode, nil
//}
//
///*
//搜索真实节点消息,用onebyone在真实节点中寻找的方法
//@return    bool                   是否发送给其他人，true=成功发送给其他人;false=发送给自己;
//@return    *nodeStore.AddressNet  选出的发送/转发节点地址
//@return    error                  处理中错误
//@return    bool                   是否强制推出此次消息发送/中转，一般当黑名单填满时候是true
//*/
//func (this *Message) routerSearchNodeOnebyone(version uint64, nodeManager *nodeStore.NodeManager, sessionEngine *engine.Engine,
//	vnc *virtual_node.VnodeManager, from *nodeStore.AddressNet, blockAddr map[string]int) (bool, *nodeStore.AddressNet, error, bool) {
//	// utils.Log.Info().Msgf("发送routerSearchVnode self:%s from:%s to:%s", vnc.GetVnodeDiscover().Vnode.Vid.B58String(),
//	// 	this.Head.SenderVnode.B58String(), this.Head.RecvVnode.B58String())
//
//	var allAddr []nodeStore.AddressNet
//	//判断nodeid是否是自己，是则自己处理
//	if nodeManager.NodeSelf.GetIsSuper() && bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
//		return false, nil, nil, true
//	}
//
//	//正常模式下的判断和转发
//	allAddr = append(allAddr, nodeManager.NodeSelf.IdInfo.Id)
//	ud := sessionEngine.GetAllDownUp(nodeManager.AreaNameSelf)
//	for _, v := range ud {
//		if blockAddr[v] > engine.MaxRetryCnt {
//			continue
//		}
//
//		allAddr = append(allAddr, nodeStore.AddressNet(v))
//	}
//
//	//计算接受节点和当前onebyone连接节点状态
//	sorted, times := this.GetSortSession(allAddr, sessionEngine)
//	// sorted, times := this.GetSortSession(allAddr)
//	if !nodeManager.NodeSelf.GetIsSuper() {
//		for i := range sorted {
//			if !bytes.Equal(nodeManager.NodeSelf.IdInfo.Id, sorted[i]) {
//				target := sorted[i]
//				return true, &target, nil, false
//			}
//		}
//		return true, nil, nil, true
//	}
//	if nodeManager.IsHeadTailModl {
//		if times == len(sorted) {
//			target := sorted[len(sorted)-1]
//			//如果自己是最后一个点，则自己处理
//			if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
//				return false, nil, nil, true
//			}
//			return true, &target, nil, false
//		}
//
//		if times == 0 {
//			target := sorted[0]
//			//如果自己是第一个点，则自己处理
//			if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
//				return false, nil, nil, true
//			}
//			return true, &target, nil, false
//		}
//		target := sorted[times-1]
//		if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
//			//如果自己是紧挨着真实节点的点，则自己处理
//			return false, nil, nil, true
//		}
//
//		return true, &target, nil, false
//	} else {
//		if times == 0 || times == len(sorted) {
//			//说明不在自己onebyone范围里，需要路由走
//			if times == 0 {
//				target := sorted[0]
//				if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
//					//如果自己是紧挨着真实节点的点，则自己处理
//					return false, nil, nil, false
//				}
//				return true, &target, nil, false
//			} else if times == len(sorted) {
//				target := sorted[len(sorted)-1]
//				if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
//					//如果自己是紧挨着真实节点的点，则自己处理
//					return false, nil, nil, false
//				}
//				return true, &target, nil, false
//			}
//		}
//
//		//utils.Log.Info().Msgf("寻找真实节点在自己onebyone范围里")
//		target := sorted[times-1]
//		//utils.Log.Info().Msgf("****%s", target.B58String())
//		if bytes.Equal(target, nodeManager.NodeSelf.IdInfo.Id) {
//			//如果自己是紧挨着真实节点的点，则自己处理
//			return false, nil, nil, false
//		}
//
//		return true, &target, nil, true
//	}
//
//}
//
///*
//给连接的节点排序(包括本节点)，结果相对target 由正到负 大到小，确认消息发送顺序
//+@return    []nodeStore.AddressNet   由远及近排列的节点地址;
//+@return    数组中有多少地址大于传入self地址；
//+
//*/
//func GetSortSessionForTarget(ss []engine.Session, self nodeStore.AddressNet, target nodeStore.AddressNet) ([]nodeStore.AddressNet, int) {
//	var targetAddrBI *big.Int
//	var sortedS []nodeStore.AddressNet
//	sortLargeNum := 0
//	onebyone := new(nodeStore.IdDESC)
//
//	if self != nil {
//		targetAddrBI = new(big.Int).SetBytes(target)
//	}
//
//	//1.把自己和上下session放在排序器中等待排序
//	*onebyone = append(*onebyone, new(big.Int).SetBytes(self))
//	for _, one := range ss {
//		*onebyone = append(*onebyone, new(big.Int).SetBytes(nodeStore.AddressNet([]byte(one.GetName()))))
//	}
//
//	//2.排序器排序
//	sort.Sort(onebyone)
//
//	for i := 0; i < len(*onebyone); i++ {
//		one := (*onebyone)[i]
//		if one.Cmp(targetAddrBI) == 1 {
//			sortLargeNum += 1
//		}
//
//		IdBs := one.Bytes()
//		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
//		sortedS = append(sortedS, nodeStore.AddressNet(*IdBsP))
//	}
//
//	return sortedS, sortLargeNum
//}
//
///*
//+给连接的节点排序，结果相对self 由正到负 大到小，确认消息发送顺序
//+@return    []engine.Session   由远及近排列的节点session数组;
//+@return    数组中有多少地址大于传入self地址；
//+
//*/
//func GetSortSession(ss []engine.Session, self nodeStore.AddressNet) ([]engine.Session, int) {
//	var selfAddrBI *big.Int
//	var sortedS []engine.Session
//	sortLargeNum := 0
//	sortM := make(map[string]engine.Session)
//	onebyone := new(nodeStore.IdDESC)
//
//	if self != nil {
//		selfAddrBI = new(big.Int).SetBytes(self)
//	}
//
//	for _, one := range ss {
//		sortM[nodeStore.AddressNet([]byte(one.GetName())).B58String()] = one
//		*onebyone = append(*onebyone, new(big.Int).SetBytes(nodeStore.AddressNet([]byte(one.GetName()))))
//	}
//
//	sort.Sort(onebyone)
//
//	for i := 0; i < len(*onebyone); i++ {
//		// utils.Log.Info().Msgf("selfAddrBI: ", selfAddrBI, " one", one).
//		one := (*onebyone)[i]
//		if one.Cmp(selfAddrBI) == 1 {
//			sortLargeNum += 1
//		}
//
//		IdBs := one.Bytes()
//		IdBsP := utils.FullHighPositionZero(&IdBs, 32)
//		sortedS = append(sortedS, sortM[nodeStore.AddressNet(*IdBsP).B58String()])
//	}
//
//	return sortedS, sortLargeNum
//}
