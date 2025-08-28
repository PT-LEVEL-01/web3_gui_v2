package message_center

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/message_center/security_signal/doubleratchet"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

// 消息协议版本号
const (
	version_neighbor            = 1  //邻居节点消息，本消息不转发
	version_multicast           = 2  //广播消息
	version_search_super        = 3  //搜索节点消息
	version_search_all          = 4  //搜索节点消息
	version_p2p                 = 5  //点对点消息
	version_p2pHE               = 6  //点对点可靠传输加密消息
	version_vnode_search        = 7  //搜索虚拟节点消息
	version_vnode_p2pHE         = 8  //虚拟节点间点对点可靠传输加密消息
	version_multicast_sync      = 9  //查询邻居节点的广播消息
	version_multicast_sync_recv = 10 //查询广播消息的回应
	version_vnode_search_end    = 11 //搜索最终接收者虚拟节点
	//#新增onebyone消息协议
	version_p2p_onebyone              = 12 //点对点消息，onebyone规则路由
	version_vnode_search_end_onebyone = 13 //搜索最终接收者虚拟节点，onebyone规则路由
	version_vnode_p2pHE_onebyone      = 14 //虚拟节点间点对点可靠传输加密消息，onebyone规则路由
	version_vnode_search_onebyone     = 15 //搜索虚拟节点消息，onebyone规则路由
	version_search_super_onebyone     = 16 //搜索节点消息，onebyone规则路由
	version_message_recv              = 17 //下层节点收到消息并向上层回复收到的处理id
	version_node_search_onebyone      = 18 // 真实节点onebyone规则路由消息
)

// var do = new(sync.Once)

func (this *MessageCenter) RegisterMsgVersion() {

	this.sessionEngine.RegisterMsg(this.areaName, version_neighbor, this.NeighborHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_multicast, this.multicastHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_search_super, this.searchSuperHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_search_all, this.searchAllHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_p2p, this.p2pHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_p2pHE, this.p2pHEHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_vnode_search, this.vnodeSearchHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_vnode_p2pHE, this.vnodeP2pHEHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_multicast_sync, this.multicastSyncHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_multicast_sync_recv, this.multicastSyncRecvHandler)
	this.sessionEngine.RegisterMsg(this.areaName, version_message_recv, this.messageRecvHandler)
}

type MsgHandler func(c engine.Controller, msg engine.Packet, message *Message)

/*
查询邻居节点的广播消息
*/
func (this *MessageCenter) multicastSyncHandler(c engine.Controller, msg engine.Packet) {
	//	fmt.Println("接收到查询最近超级节点请求")
	fromAddr := nodeStore.AddressNet(msg.Session.GetName())
	is := false
	if this.nodeManager.FindWhiteList(&fromAddr) {
		is = true
	}
	if this.nodeManager.FindOneByOneList(&fromAddr) {
		is = true
	}
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		if is {
			utils.Log.Info().Msgf("get multicast message error from:%s %s", fromAddr.B58String(), err.Error())
		}
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		if is {
			utils.Log.Info().Msgf("get multicast message error from:%s %s", fromAddr.B58String(), err.Error())
		}
		return
	}

	// messageCache, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	// if err != nil {
	// 	if is {
	// 		utils.Log.Info().Msgf("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
	// 	}
	// 	utils.Log.Error().Msgf("get multicast message :%s error:%s", hex.EncodeToString(*message.Body.Content), err.Error())
	// 	return
	// }
	// mmp := go_protobuf.MessageMulticast{
	// 	Head: messageCache.Head,
	// 	Body: messageCache.Body,
	// }
	// content, err := mmp.Marshal()
	// if err != nil {
	// 	if is {
	// 		utils.Log.Info().Msgf("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
	// 	}
	// 	utils.Log.Error().Msgf(err.Error())
	// 	return
	// }

	content, err := FindMessageCacheByHash(*message.Body.Content, this.levelDB)
	if err != nil {
		if is {
			utils.Log.Info().Msgf("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		}
		utils.Log.Error().Msgf("get multicast message :%s error:%s", hex.EncodeToString(*message.Body.Content), err.Error())
		return
	}
	if content == nil {
		return
	}

	head := NewMessageHead(this.nodeManager.NodeSelf, this.nodeManager.GetSuperPeerId(), message.Head.Sender, message.Head.SenderSuperId, true, this.nodeManager.GetMachineID(), message.Head.SenderMachineID)
	body := NewMessageBody(0, content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	newmessage := NewMessage(head, body)
	newmessage.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)

	mheadBs := head.Proto()
	mbodyBs, err := body.Proto()
	if err != nil {
		utils.Log.Error().Msgf(err.Error())
		return
	}
	err = msg.Session.Send(version_multicast_sync_recv, &mheadBs, &mbodyBs, 0)
	if err != nil {
		if is {
			utils.Log.Info().Msgf("get multicast message success error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		}
	} else {
		if is {
			// utils.Log.Info().Msgf("get multicast message success from:%s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content))
		}
	}
	// SendNeighborReplyMsg(message, config.MSGID_multicast_return, nil, msg.Session)
	// //
	// msg.Session.Send(version_multicast_sync_recv, messageCache.Head, messageCache.Body, false)
}

func (this *MessageCenter) multicastSyncRecvHandler(c engine.Controller, msg engine.Packet) {
	//	fmt.Println("接收到查询最近超级节点请求")
	fromAddr := nodeStore.AddressNet(msg.Session.GetName())
	is := false
	if this.nodeManager.FindWhiteList(&fromAddr) {
		is = true
		// utils.Log.Info().Msgf("recv multicast message content from:%s", fromAddr.B58String())
	}

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		if is {
			utils.Log.Info().Msgf("recv multicast message content error from:%s", fromAddr.B58String(), err.Error())
		}
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		if is {
			utils.Log.Info().Msgf("recv multicast message content error from:%s", fromAddr.B58String(), err.Error())
		}
		return
	}

	//自己处理
	// flood.ResponseWait(config.CLASS_engine_multicast_sync, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
邻居节点消息控制器，不转发消息
*/
func (this *MessageCenter) NeighborHandler(c engine.Controller, msg engine.Packet) {
	//	fmt.Println("接收到查询最近超级节点请求")
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }

	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	return
	// }
	// utils.Log.Info().Msgf("------自己处理-----------")
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		return
	}
	//自己处理
	h := this.router.GetHandler(message.Body.MessageId)
	if h == nil {
		// fmt.Println("This neighbor message is not registered:", message.Body.MessageId)
		// utils.Log.Info().Msgf("This neighbor message is not registered:%d", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
广播消息控制器
*/
func (this *MessageCenter) multicastHandler(c engine.Controller, msg engine.Packet) {
	// utils.Log.Debug().Msgf("收到广播消息")

	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	//广播消息头解析失败
	// 	utils.Log.Warn().Msgf("Parsing of this broadcast header failed")
	// 	return
	// }
	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	utils.Log.Info().Msgf("Content parsing of this broadcast message failed %s", err.Error())
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		//广播消息头解析失败
		utils.Log.Warn().Msgf("Parsing of this broadcast header failed")
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		utils.Log.Info().Msgf("Content parsing of this broadcast message failed %s", err.Error())
		return
	}
	//第一时间回复广播消息
	// go SendNeighborReplyMsg(message, config.MSGID_multicast_return, nil, msg.Session)

	// utils.Log.Info().Msgf("广播消息hash %s", hex.EncodeToString(message.Body.Hash))

	//判断重复的广播
	// if !message.CheckSendhash() {
	// 	utils.Log.Info().Msgf("This multicast message is repeated")
	// 	// utils.Log.Info().Msgf("有重复的广播")
	// 	return
	// }
	// utils.Log.Info().Msgf("这个广播不重复")

	//先同步，同步了再广播给其他节点
	mhOne := MsgHolderOne{
		MsgHash: *message.Body.Content,
		Addr:    nodeStore.AddressNet([]byte(msg.Session.GetName())), // *message.Head.SenderSuperId, // nodeStore.AddressNet
		Message: message,
		Session: msg.Session,
	}

	this.msgchannl.Add(&mhOne)

	// select {
	// case this.msgchannl <- &mhOne:
	// default:
	// 	utils.Log.Error().Msgf("msgchannl is full")
	// }

	return

	// //继续广播给其他节点
	// // utils.Log.Info().Msgf("自己是否是超级节点 %v", nodeStore.NodeSelf.IsSuper)
	// if nodeStore.NodeSelf.IsSuper {
	// 	//广播给其他超级节点
	// 	go func() {
	// 		//先发送给超级节点
	// 		superNodes := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
	// 		//广播给代理对象
	// 		proxyNodes := nodeStore.GetProxyAll()
	// 		broadcastsAll(superNodes, proxyNodes, message)

	// 		return
	// 	}()
	// }

	// //自己处理
	// h := router.GetHandler(message.Body.MessageId)
	// if h == nil {
	// 	utils.Log.Info().Msgf("This broadcast message is not registered:", message.Body.MessageId)
	// 	return
	// }
	// // utils.Log.Info().Msgf("有广播消息，消息编号 %d", message.Body.MessageId)
	// h(c, msg, message)

}

/*
逐层回复时，遇到某一层层没回复返回给消息发送者的报告
*/
func (this *MessageCenter) sendBackToMsgSender(message *Message) {
	//	utils.Log.Info().Msgf("sendBackToMsgSender messge sender: %s", message.Head.Sender.B58String())
	if err := message.ParserContentProto(); err != nil {
		return
	}
	content := []byte(config.CLASS_router_err)
	if _, err := this.SendP2pMsgEX(config.MsGID_recv_router_err, message.Head.Sender, message.Head.SenderSuperId, &content, &message.Body.Hash); err != nil {
		utils.Log.Warn().Msgf("send back err :%s", err.Error())
	}
	return
}

/*
从超级节点中搜索目标节点消息控制器
*/
func (this *MessageCenter) searchSuperHandler(c engine.Controller, msg engine.Packet) {

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	from := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.DataPlus != nil && !bytes.Equal(from, *message.Head.Sender) {
	// 	//utils.Log.Info().Msgf("vnodeSearchHandler 逐层回复 self : %s  DataPlus : %v from :  %s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), *message.DataPlus, from.B58String())
	// 	//逐层回复
	// 	err = this.layerReply(message, this.sessionEngine, &msg)
	// 	if err != nil {
	// 		utils.Log.Warn().Msgf("layer reply error : %s", err.Error())
	// 		return
	// 	}

	// }

	// isOther, err := message.IsSendOther(&from, this, 0)
	isOther, err := message.Forward(message.msgid, this.nodeManager, this.sessionEngine, this.vm, &from, 0)
	if err != nil {
		if err == config.ERROR_send_to_sender {
			this.sendBackToMsgSender(message)
		}
		// utils.Log.Info().Msgf("Forward error:%s", err.Error())
		return
	}
	if isOther {
		return
	}
	// utils.Log.Info().Msgf("------自己处理-----------")
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}

	//自己处理
	h := this.router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This searchsuper message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
从所有节点中搜索目标节点消息控制器
*/
func (this *MessageCenter) searchAllHandler(c engine.Controller, msg engine.Packet) {
	this.searchSuperHandler(c, msg)
	// message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// form := nodeStore.AddressNet(msg.Session.GetName())
	// ok, err := message.IsSendOther(&form, this, 0)
	// if err != nil {
	// 	// utils.Log.Info().Msgf("IsSendOther error:%s", err.Error())
	// 	return
	// }
	// if ok {
	// 	return
	// }
	// // utils.Log.Info().Msgf("------自己处理-----------")
	// //解析包体内容
	// if err = message.ParserContentProto(); err != nil {
	// 	return
	// }

	// //自己处理
	// h := this.router.GetHandler(message.Body.MessageId)
	// if h == nil {
	// 	fmt.Println("This searchAll message is not registered:", message.Body.MessageId)
	// 	return
	// }
	// h(c, msg, message)
}

/*
点对点消息控制器
*/
func (this *MessageCenter) p2pHandler(c engine.Controller, msg engine.Packet) {
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	from := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.DataPlus != nil && !bytes.Equal(from, *message.Head.Sender) {
	// 	//utils.Log.Info().Msgf("vnodeSearchHandler 逐层回复 self : %s  DataPlus : %v from :  %s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), *message.DataPlus, from.B58String())
	// 	//逐层回复
	// 	err = this.layerReply(message, this.sessionEngine, &msg)
	// 	if err != nil {
	// 		utils.Log.Warn().Msgf("layer reply error : %s", err.Error())
	// 		return
	// 	}
	// }

	isOther, err := message.Forward(message.msgid, this.nodeManager, this.sessionEngine, this.vm, &from, 0)
	if err != nil {
		// utils.Log.Info().Msgf("Forward error:%s", err.Error())
		if err == config.ERROR_send_to_sender {
			this.sendBackToMsgSender(message)
		}
		return
	}
	if isOther {
		return
	}

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		return
	}
	// utils.Log.Info().Msgf("------自己处理-----------")
	//自己处理
	h := this.router.GetHandler(message.Body.MessageId)
	if h == nil {
		utils.Log.Info().Msgf("This P2P message is not registered,msgId:%d", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
点对点消息控制器
*/
func (this *MessageCenter) p2pHEHandler(c engine.Controller, msg engine.Packet) {
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	from := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.DataPlus != nil && !bytes.Equal(from, *message.Head.Sender) {
	// 	//		utils.Log.Info().Msgf("vnodeSearchHandler 逐层回复 self : %s  %v", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), msg.Dataplus)
	// 	//逐层回复
	// 	err = this.layerReply(message, this.sessionEngine, &msg)
	// 	if err != nil {
	// 		utils.Log.Warn().Msgf("layer reply error : %s", err.Error())
	// 		return
	// 	}
	// }

	isOther, err := message.Forward(message.msgid, this.nodeManager, this.sessionEngine, this.vm, &from, 0)
	if err != nil {
		// utils.Log.Info().Msgf("Forward error:%s", err.Error())
		if err == config.ERROR_send_to_sender {
			this.sendBackToMsgSender(message)
		}
		return
	}
	if isOther {
		return
	}
	// utils.Log.Info().Msgf("------自己处理-----------")
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}
	headSize := binary.LittleEndian.Uint64((*message.Body.Content)[:8])
	headData := (*message.Body.Content)[8 : headSize+8]
	// bodySize := binary.LittleEndian.Uint64((*message.Body.Content)[headSize+8 : headSize+8+8])
	bodyData := (*message.Body.Content)[headSize+8+8:]

	msgHE := doubleratchet.MessageHE{
		Header:     headData,
		Ciphertext: bodyData,
	}
	// fmt.Println("head", hex.EncodeToString(msgHE.Header))
	// fmt.Println("body", hex.EncodeToString(msgHE.Ciphertext))

	sessionHE := this.RatchetSession.GetRecvRatchet(*message.Head.Sender, message.Head.SenderMachineID)
	if sessionHE == nil {
		// fmt.Println("双棘轮接收key未找到")
		//双棘轮key未找到
		if message.Head.SenderProxyId == nil || len(*message.Head.SenderProxyId) == 0 {
			go this.SendP2pMsgHE(config.MSGID_security_pipe_error, message.Head.Sender, nil)
		} else {
			go this.SendP2pMsgHEProxy(config.MSGID_security_pipe_error, message.Head.Sender, message.Head.SenderProxyId, message.Head.RecvProxyId, message.Head.SenderMachineID, nil)
		}
		return
	}
	bs, err := sessionHE.RatchetDecrypt(msgHE, nil)
	if err != nil {
		// fmt.Println("解密消息出错", err)
		return
	}
	*message.Body.Content = bs
	// fmt.Println("开始解密消息 33333333333333")
	//自己处理
	h := this.router.GetHandler(message.Body.MessageId)
	if h == nil {
		utils.Log.Info().Msgf("This P2PHE message is not registered,msgId:%d", message.Body.MessageId)
		return
	}
	//计算出落到自己的哪个虚拟节点上
	if message.Head.RecvVnode != nil && len(*message.Head.RecvVnode) != 0 {
		if message.checkRouterType() == version_vnode_p2pHE {
			targetVnodeID := this.vm.FindNearVnodeP2P(message.Head.RecvVnode, nil, true, nil)
			message.Head.SelfVnodeId = &targetVnodeID
		}
		if message.checkRouterType() == version_vnode_search {
			targetVnodeID := this.vm.FindNearVnodeSearchVnode(message.Head.RecvVnode, nil, true, false, nil)
			message.Head.SelfVnodeId = &targetVnodeID
		}
		if message.checkRouterType() == version_vnode_search_end {
			targetVnodeID := this.vm.FindNearVnodeSearchVnode(message.Head.RecvVnode, nil, true, true, nil)
			message.Head.SelfVnodeId = &targetVnodeID
		}
	}
	h(c, msg, message)
	// fmt.Println("开始解密消息 44444444444444444")
}

/*
从所有虚拟节点中搜索目标节点消息控制器
*/
func (this *MessageCenter) vnodeSearchHandler(c engine.Controller, msg engine.Packet) {
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	from := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.DataPlus != nil && !bytes.Equal(from, *message.Head.Sender) {
	// 	//utils.Log.Info().Msgf("vnodeSearchHandler 逐层回复 self : %s  %v", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), msg.Dataplus)
	// 	//逐层回复
	// 	err = this.layerReply(message, this.sessionEngine, &msg)
	// 	if err != nil {
	// 		utils.Log.Warn().Msgf("layer reply error : %s", err.Error())
	// 		return
	// 	}
	// }

	isOther, err := message.Forward(message.msgid, this.nodeManager, this.sessionEngine, this.vm, &from, 0)
	if err != nil {
		// utils.Log.Info().Msgf("Forward error:%s", err.Error())
		if err == config.ERROR_send_to_sender {
			//			utils.Log.Info().Msgf("Forward back")
			this.sendBackToMsgSender(message)
		}
		return
	}
	if isOther {
		return
	}

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		return
	}
	// utils.Log.Info().Msgf("------自己处理-----------")
	//自己处理
	h := this.router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This searchAll message is not registered:", message.Body.MessageId)
		return
	}
	//计算出落到自己的哪个虚拟节点上
	// 如果指定最终处理的虚拟节点，并且判定是自己处理，则直接落到处理的虚拟节点
	if message.Head.SearchVnodeEndId != nil && len(*message.Head.SearchVnodeEndId) != 0 {
		message.Head.SelfVnodeId = message.Head.SearchVnodeEndId
	} else if message.Head.RecvVnode != nil && len(*message.Head.RecvVnode) != 0 {
		// 既然指定了自己处理，所以应该只在自己的虚拟节点上查询最近的节点
		// 派出index为0的节点进行查找
		targetVnodeID := this.vm.FindNearVnodeInSelfAppIndex0(message.Head.RecvVnode, false)
		if len(targetVnodeID) == 0 {
			// 如果不存在虚拟节点，这个时候只能包含index为0的进行查找
			targetVnodeID = this.vm.FindNearVnodeInSelfAppIndex0(message.Head.RecvVnode, true)
		}
		message.Head.SelfVnodeId = &targetVnodeID
	}
	h(c, msg, message)
}

/*
虚拟节点间点对点可靠传输加密消息
*/
func (this *MessageCenter) vnodeP2pHEHandler(c engine.Controller, msg engine.Packet) {
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	from := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.DataPlus != nil && !bytes.Equal(from, *message.Head.Sender) {
	// 	//utils.Log.Info().Msgf("vnodeSearchHandler 逐层回复 self : %s  %v", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), msg.Dataplus)
	// 	err = this.layerReply(message, this.sessionEngine, &msg)
	// 	if err != nil {
	// 		utils.Log.Warn().Msgf("layer reply error : %s", err.Error())
	// 		return
	// 	}
	// }

	isOther, err := message.Forward(message.msgid, this.nodeManager, this.sessionEngine, this.vm, &from, 0)
	if err != nil {
		// utils.Log.Info().Msgf("Forward error:%s", err.Error())
		if err == config.ERROR_send_to_sender {
			this.sendBackToMsgSender(message)
		}
		return
	}
	if isOther {
		return
	}

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		return
	}
	// utils.Log.Info().Msgf("------自己处理-----------self:%s", this.vm.DiscoverVnodes.Vnode.Vid.B58String())
	//自己处理
	h := this.router.GetHandler(message.Body.MessageId)
	if h == nil {
		utils.Log.Info().Msgf("This searchAll message is not registered:%d", message.Body.MessageId)
		return
	}
	//计算出落到自己的哪个虚拟节点上
	if message.Head.RecvVnode != nil && len(*message.Head.RecvVnode) != 0 {
		targetVnodeID := this.vm.FindNearVnodeP2P(message.Head.RecvVnode, nil, true, nil)
		message.Head.SelfVnodeId = &targetVnodeID
	}
	h(c, msg, message)
}
func (this *MessageCenter) messageRecvHandler(c engine.Controller, msg engine.Packet) {
	//	utils.Log.Info().Msgf("messageRecvHandler self : %s", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	fromAddr := nodeStore.AddressNet(msg.Session.GetName())
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		utils.Log.Warn().Msgf("messageRecvHandler message content error from:%s", fromAddr.B58String(), err.Error())
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		utils.Log.Warn().Msgf("messageRecvHandler message content error from:%s", fromAddr.B58String(), err.Error())
		return
	}

	//自己处理
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), nil)
}

// 这种消息不走路由好一点重写方法
// func (this *MessageCenter) layerReply(message *Message, recvid *nodeStore.AddressNet, sessionEngine *engine.Engine) error {
// 	mhead := NewMessageHeadProxy(this.nodeManager.NodeSelf, nil, recvid,
// 		recvid, nil, nil, true)
// 	//msg_id_p2p_recv = 1002,1002是测试方法的p2p_recv
// 	mbody := NewMessageBody(1002, nil, 0, *message.DataPlus, 0)
// 	newmessage := NewMessage(mhead, mbody)
// 	newmessage.BuildReplyHash(0, *message.DataPlus, 0)
// 	isOther, err := newmessage.Reply(version_p2p, this.nodeManager, this.sessionEngine, this.vm)
// 	if err != nil || !isOther {
// 		if err != nil {
// 			utils.Log.Info().Msgf("layerReply error ,error : %s isOther : %t to : %s self : %s", err.Error(), isOther, newmessage.Head.RecvId.B58String(), newmessage.Head.Sender.B58String())
// 		}
// 		utils.Log.Info().Msgf("layerReply error , isOther : %t to : %s self : %s", isOther, newmessage.Head.RecvId.B58String(), newmessage.Head.Sender.B58String())

// 		return errors.New("layerReply ERRR")
// 	}
// 	utils.Log.Info().Msgf("回复成功 id %d DataPlus : %v", message.msgid, newmessage.Body.Hash)
// 	return nil
// }

func (this *MessageCenter) layerReply(message *Message, sessionEngine *engine.Engine, msg *engine.Packet) error {
	mhead := NewMessageHead(this.nodeManager.NodeSelf, this.nodeManager.GetSuperPeerId(), message.Head.Sender, message.Head.SenderSuperId, true, this.nodeManager.GetMachineID(), message.Head.SenderMachineID)
	from := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	waitr := []byte(from)
	waitr = append(waitr, *message.DataPlus...)
	mbody := NewMessageBody(0, nil, 0, waitr, 0)
	newmessage := NewMessage(mhead, mbody)
	newmessage.BuildReplyHash(0, waitr, 0)
	mheadBs := mhead.Proto()
	mbodyBs, err := mbody.Proto()
	if err != nil {
		utils.Log.Error().Msgf(err.Error())
		return err
	}
	err = msg.Session.Send(version_message_recv, &mheadBs, &mbodyBs, 0)
	if err != nil {
		utils.Log.Error().Msgf(err.Error())
		return err
	}
	//utils.Log.Info().Msgf("layerReply msgid:%d self : %s  ", message.msgid, this.nodeManager.NodeSelf.IdInfo.Id.B58String(), *message.DataPlus)
	//	utils.Log.Info().Msgf("回复成功 id %d DataPlus : %v", message.msgid, newmessage.Body.Hash)
	return nil
}
