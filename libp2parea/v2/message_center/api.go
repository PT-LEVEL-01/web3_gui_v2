package message_center

import (
	"bytes"
	"github.com/oklog/ulid/v2"
	"sync"
	"time"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
注册邻居节点消息，不转发
*/
func (this *MessageCenter) Register_neighbor(msgid uint64, handler MsgHandler) {
	this.router.Register(msgid, handler)
}

/*
发送一个邻居节点消息
发出去不管
*/
func (this *MessageCenter) SendNeighborMsgBySession(session engine.Session, msgid uint64, content *[]byte,
	timeout time.Duration) (*MessageBase, utils.ERROR) {
	mn := NewMessageNeighbor(msgid, this.nodeManager.NodeSelf.IdInfo.Id, content)
	bs, err := mn.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	_, ERR := session.Send(config.Version_neighbor, bs, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return mn, utils.NewErrorSuccess()
}

/*
发送一个邻居节点消息
发送出去，等待返回
*/
func (this *MessageCenter) SendNeighborMsgWaitRequest(session engine.Session, msgid uint64, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	mn := NewMessageNeighborWait(msgid, this.nodeManager.NodeSelf.IdInfo.Id, content)
	bs, err := mn.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Str("发送消息后等待返回", "开始").Send()
	bs, ERR := session.SendWait(config.Version_neighbor, bs, timeout)
	//utils.Log.Info().Str("发送消息后等待返回", "返回了").Send()
	if ERR.CheckFail() {
		//if ERR.Code== engine.ERROR_code_send_cache_close{}
		//this.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	return bs, ERR
}

/*
对邻居消息回复
*/
func (this *MessageCenter) SendNeighborReplyMsg(message *MessageBase, content *[]byte, timeout time.Duration) utils.ERROR {
	//mn := message_bean.NewMessageNeighborReply(message.GetBase(), this.nodeManager.NodeSelf.IdInfo.Id, content)
	//bs, err := mn.Proto()
	//if err != nil {
	//	return utils.NewErrorSysSelf(err)
	//}
	packet := message.GetPacket()
	return packet.Session.Reply(packet, content, timeout)
}

/*
注册广播消息
*/
func (this *MessageCenter) Register_multicast(msgid uint64, handler MsgHandler) {
	this.router.Register(msgid, handler)
}

/*
发送一个广播消息
当其中一个节点发送成功，则返回成功。
*/
func (this *MessageCenter) SendMulticastMsg(msgId uint64, content *[]byte, timeout time.Duration) utils.ERROR {
	if timeout == 0 {
		timeout = time.Minute
	}
	msgHash := ulid.Make().Bytes()
	//this.Log.Info().Msgf("发送广播 内容长度:%d hash:%s", len(*content), hex.EncodeToString(msgHash))
	ERR := AddMessageCacheByBytes(&msgHash, content, this.levelDB)
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}
	forwardMsg := NewMessageForward(config.Version_multicast_head, msgId, this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, nil, &msgHash)
	//this.Log.Info().Msgf("发送广播消息：%+v", forwardMsg)
	_, ERR = this.broadcastsAll(forwardMsg, timeout)
	return ERR
}

/*
发送一个新的广播消息
当其中一个节点发送成功，则返回成功。
返回第一个回复
*/
func (this *MessageCenter) SendMulticastMsgWaitRequest(msgId uint64, content *[]byte, timeout time.Duration) (*[]byte, utils.ERROR) {
	if timeout == 0 {
		timeout = time.Minute
	}
	msgHash := ulid.Make().Bytes()
	ERR := AddMessageCacheByBytes(&msgHash, content, this.levelDB)
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return nil, ERR
	}
	forwardMsg := NewMessageForward(config.Version_multicast_head, msgId, this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, nil, &msgHash)
	engine.RegisterRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
	defer engine.RemoveRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
	_, ERR = this.broadcastsAll(forwardMsg, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	bs, ERR := engine.WaitResponseByteKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return bs, ERR
}

/*
发送广播消息hash
@wait    bool    是否等待消息回复
*/
func (this *MessageCenter) broadcastsAll(msgHash *MessageBase, timeout time.Duration) (*[]*[]byte, utils.ERROR) {
	if timeout == 0 {
		timeout = time.Minute
	}
	bs, err := msgHash.Proto()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}

	//给已发送的节点放map里，避免重复发送
	allNodes := make(map[string]bool)
	var timeouterrorlock = new(sync.Mutex)
	var timeouterror utils.ERROR
	resultBss := make([]*[]byte, 0)
	cs := make(chan bool, config.CPUNUM)
	group := new(sync.WaitGroup)

	//先发送给白名单节点
	for _, nodeOne := range this.nodeManager.GetWhiltListNodes() {
		for _, one := range nodeOne.GetSessions() {
			//去重
			_, ok := allNodes[utils.Bytes2string(one.GetId())]
			if ok {
				continue
			}
			allNodes[utils.Bytes2string(one.GetId())] = false
			cs <- false
			group.Add(1)
			//广播给节点
			utils.Go(func() {
				//this.Log.Info().Msgf("广播给白名单节点:%s", one.GetRemoteMultiaddr().String())
				bss, ERR := this.broadcastsOne(one, bs, cs, group, timeout)
				if ERR.CheckSuccess() {
					return
				}
				timeouterrorlock.Lock()
				timeouterror = ERR
				resultBss = append(resultBss, bss)
				timeouterrorlock.Lock()
			}, this.Log)
		}
	}
	group.Wait()

	//再发送给超级节点
	for _, nodeOne := range this.nodeManager.GetLogicNodesWAN() {
		for _, one := range nodeOne.GetSessions() {
			//去重
			_, ok := allNodes[utils.Bytes2string(one.GetId())]
			if ok {
				continue
			}
			allNodes[utils.Bytes2string(one.GetId())] = false
			cs <- false
			group.Add(1)
			//区块广播给节点
			//this.Log.Info().Msgf("广播给超级节点:%s", one.GetRemoteMultiaddr().String())
			go this.broadcastsOne(one, bs, cs, group, timeout)
		}
	}
	group.Wait()

	//再发送给内网节点
	for _, nodeOne := range this.nodeManager.GetLogicNodesLAN() {
		for _, one := range nodeOne.GetSessions() {
			//utils.Log.Info().Msgf("广播给内网节点:%+v", one.GetRemoteMultiaddr())
			//去重
			_, ok := allNodes[utils.Bytes2string(one.GetId())]
			if ok {
				//this.Log.Info().Msgf("广播给内网节点:%+v", one.GetRemoteMultiaddr())
				continue
			}
			allNodes[utils.Bytes2string(one.GetId())] = false
			cs <- false
			group.Add(1)
			//区块广播给节点
			//this.Log.Info().Msgf("广播给内网节点:%+v", one.GetRemoteMultiaddr())
			go this.broadcastsOne(one, bs, cs, group, timeout)
		}
	}

	//再发送给游离的节点
	for _, nodeOne := range this.nodeManager.GetFreeNodes() {
		for _, one := range nodeOne.GetSessions() {
			//去重
			_, ok := allNodes[utils.Bytes2string(one.GetId())]
			if ok {
				continue
			}
			allNodes[utils.Bytes2string(one.GetId())] = false
			cs <- false
			group.Add(1)
			//区块广播给节点
			//this.Log.Info().Msgf("广播给游离节点:%s", one.GetRemoteMultiaddr().String())
			go this.broadcastsOne(one, bs, cs, group, timeout)
		}
	}

	//再发送给代理的节点
	for _, nodeOne := range this.nodeManager.GetProxyNodes() {
		for _, one := range nodeOne.GetSessions() {
			//去重
			_, ok := allNodes[utils.Bytes2string(one.GetId())]
			if ok {
				continue
			}
			allNodes[utils.Bytes2string(one.GetId())] = false
			cs <- false
			group.Add(1)
			//区块广播给节点
			//nodeInfo := GetNodeInfo(one)
			//this.Log.Info().Msgf("广播给代理节点:%s size:%d", nodeInfo.IdInfo.Id.B58String(), len(msgHash.Content))
			go this.broadcastsOne(one, bs, cs, group, timeout)
		}
	}
	group.Wait()
	// utils.Log.Info().Msgf("multicast proxy node time %s", time.Now().Sub(start))
	return &resultBss, timeouterror
}

/*
给一个节点发送广播消息
当其中一个节点发送成功，则返回成功。
返回第一个回复
@wait    bool    是否等待消息回复
*/
func (this *MessageCenter) broadcastsOne(session engine.Session, bs *[]byte, cs chan bool, group *sync.WaitGroup,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	//var bs *[]byte
	var ERR utils.ERROR
	//this.Log.Info().Msgf("不用回复的广播:%d %d", msgId, len(*bs2))
	_, ERR = session.Send(config.Version_multicast_head, bs, timeout)
	if ERR.CheckFail() {
		//this.Log.Info().Msgf("不用回复的广播:%d 错误:%s", msgId, ERR.String())
	}
	<-cs
	group.Done()
	return nil, ERR
}

/*
对广播消息回复
*/
func (this *MessageCenter) SendMulticastMsgReplyMsg(message *MessageBase, content *[]byte) utils.ERROR {
	//this.Log.Info().Str("回复消息", "").Send()
	if !this.CheckOnline() {
		return utils.NewErrorBus(config.ERROR_code_offline, "")
	}
	//this.Log.Info().Str("回复消息", "").Send()
	forwardMsg := NewMessageForwardReply(config.Version_multicast_recv, message.GetBase(), this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, content)
	//this.Log.Info().Interface("回复消息", forwardMsg.ConvertVO()).Send()
	isSelf, ERR := this.forward(forwardMsg, true, 0)
	//this.Log.Info().Str("回复消息", "").Send()
	if ERR.CheckFail() {
		return ERR
	}
	if isSelf {
		//this.Log.Info().Str("回复消息", "").Send()
		this.handlerProcess(forwardMsg)
	}
	//this.Log.Info().Str("回复消息", "").Send()
	return utils.NewErrorSuccess()
}

/*
注册点对点通信消息
*/
func (this *MessageCenter) Register_p2p(msgid uint64, handler MsgHandler) {
	this.router.Register(msgid, handler)
}

/*
给指定节点发送一个消息
@return    *Message     返回的消息
@return    error        发送失败
*/
func (this *MessageCenter) SendP2pMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte, timeout time.Duration) (*MessageBase, utils.ERROR) {
	if !this.CheckOnline() {
		return nil, utils.NewErrorBus(config.ERROR_code_offline, "")
	}
	if recvid == nil || len(recvid.Data()) == 0 {
		return nil, utils.NewErrorBus(config.ERROR_code_params_fail, "recvid")
	}
	forwardMsg := NewMessageForward(config.Version_p2p, msgid, this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, recvid, content)
	isSelf, ERR := this.forward(forwardMsg, true, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	if isSelf {
		this.handlerProcess(forwardMsg)
	}
	return forwardMsg, utils.NewErrorSuccess()
}

/*
给指定节点发送一个消息，并等待返回
@return    *[]byte      返回的内容
*/
func (this *MessageCenter) SendP2pMsgWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	if !this.CheckOnline() {
		return nil, utils.NewErrorBus(config.ERROR_code_offline, "")
	}
	if recvid == nil || len(recvid.Data()) == 0 {
		return nil, utils.NewErrorBus(config.ERROR_code_params_fail, "recvid")
	}
	resultBs, ERR := this.sendP2pMsgWaitRequest(config.Version_p2p, msgid, recvid, content, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return resultBs, utils.NewErrorSuccess()
}

/*
对某个消息回复
*/
func (this *MessageCenter) SendP2pReplyMsg(message *MessageBase, content *[]byte) utils.ERROR {
	//this.Log.Info().Str("回复消息", "").Send()
	if !this.CheckOnline() {
		return utils.NewErrorBus(config.ERROR_code_offline, "")
	}
	//this.Log.Info().Str("回复消息", "").Send()
	forwardMsg := NewMessageForwardReply(config.Version_p2p, message.GetBase(), this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, content)
	//this.Log.Info().Interface("回复消息", forwardMsg.ConvertVO()).Send()
	isSelf, ERR := this.forward(forwardMsg, true, 0)
	//this.Log.Info().Str("回复消息", "").Send()
	if ERR.CheckFail() {
		return ERR
	}
	if isSelf {
		//this.Log.Info().Str("回复消息", "").Send()
		this.handlerProcess(forwardMsg)
	}
	//this.Log.Info().Str("回复消息", "").Send()
	return utils.NewErrorSuccess()
}

/*
注册点对点通信消息
*/
func (this *MessageCenter) Register_p2pHE(msgid uint64, handler MsgHandler) {
	this.router.Register(msgid, handler)
}

// var securityStore = new(sync.Map)

/*
发送一个加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
//func (this *MessageCenter) SendP2pMsgHE(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) utils.ERROR {
//	_, ERR := this.SendP2pMsgHEWaitRequest(msgid, recvid, content, 0)
//	return ERR
//}

/*
发送一个加密消息，包括消息头也加密
@return    *[]byte        返回的消息
@return    utils.ERROR    是否发送成功
*/
func (this *MessageCenter) SendP2pMsgHEWaitRequest(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	if !this.CheckOnline() {
		return nil, utils.NewErrorBus(config.ERROR_code_offline, "")
	}
	if recvid == nil || len(recvid.Data()) == 0 {
		return nil, utils.NewErrorBus(config.ERROR_code_params_fail, "recvid")
	}
	//this.Log.Info().Msgf("SendP2pMsgHEWaitRequest发送加密消息")
	//是自己发送给自己，则不用协商密钥
	if bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), recvid.Data()) {
		forwardMsg := NewMessageForward(config.Version_p2pHE_wait, msgid, this.nodeManager.NodeSelf.IdInfo.Id,
			this.nodeManager.NodeSelf.MachineID, recvid, content)
		if timeout > 0 {
			engine.RegisterRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
			defer engine.RemoveRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
			this.handlerProcess(forwardMsg)
			resultBs, ERR := engine.WaitResponseByteKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID, timeout)
			if ERR.CheckFail() {
				return nil, ERR
			}
			//resultBs := resultItr.(*[]byte)
			return resultBs, utils.NewErrorSuccess()
		} else {
			this.handlerProcess(forwardMsg)
			return nil, utils.NewErrorSuccess()
		}
	}
	this.Log.Info().Str("发送加密消息", "").Send()
	resultBs, ERR := this.sendP2pMsgWaitHEOne(msgid, recvid, content, timeout)
	if ERR.CheckSuccess() {
		return resultBs, ERR
	}
	if ERR.Code == config.ERROR_code_recv_security_store_not_exist ||
		ERR.Code == config.ERROR_code_recv_ratchet_not_exist {
		utils.Log.Info().Str("对方重启过需重新发送加密消息", "").Send()
		//清理本地的双棘轮密钥后重试
		this.CleanHEInfo(recvid, nil)
		//再次发送
		resultBs, ERR = this.sendP2pMsgWaitHEOne(msgid, recvid, content, timeout)
		if ERR.CheckSuccess() {
			return resultBs, ERR
		}
	}
	//this.Log.Info().Str("发送加密消息", ERR.String()).Send()
	//返回其他错误
	return nil, ERR
}

/*
发送一个加密消息，包括消息头也加密
@return    []byte         返回消息节点机器号
@return    utils.ERROR    是否发送成功
*/
func (this *MessageCenter) sendP2pMsgWaitHEOne(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	//tempLock.Lock()
	//defer tempLock.Unlock()
	if timeout == 0 {
		timeout = time.Minute
	}
	//this.Log.Info().Msgf("sendP2pMsgWaitHEOne 11111111")
	ciphertext, ERR := this.encryptMessagesRatchet(recvid, nil, content, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Msgf("sendP2pMsgWaitHEOne 11111111")
	forwardMsg := NewMessageForward(config.Version_p2pHE_wait, msgid, this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, recvid, ciphertext)
	engine.RegisterRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
	defer engine.RemoveRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
	_, ERR = this.forward(forwardMsg, true, timeout)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//utils.Log.Info().Msgf("sendP2pMsgWaitHEOne 11111111")
	//等待返回。可能返回棘轮密钥未找到，则清除本地棘轮密钥重新发送；或者返回成功
	messageItr, ERR := engine.WaitResponseItrKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID, timeout)
	if ERR.CheckFail() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return nil, ERR
	}
	//utils.Log.Info().Msgf("sendP2pMsgWaitHEOne 11111111")
	message := messageItr.(*MessageBase)
	nr, err := utils.ParseNetResult(message.Content)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//this.Log.Info().Msgf("sendP2pMsgWaitHEOne 11111111")
	ERR = nr.ConvertERROR()
	if ERR.CheckFail() {
		//有可能对方解密不了消息
		return nil, ERR
	}
	//this.Log.Info().Msgf("sendP2pMsgWaitHEOne 11111111")
	bs, ERR := this.decryptMessage(message.SenderAddr, message.SenderMachineID, &nr.Data)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//this.Log.Info().Msgf("sendP2pMsgWaitHEOne 11111111")
	//解密返回的消息
	return bs, utils.NewErrorSuccess()
}

/*
加密消息，包括消息头也加密
@return    *Message     返回的消息
@return    bool         是否发送成功
@return    bool         消息是发给自己
*/
func (this *MessageCenter) encryptMessagesRatchet(recvid *nodeStore.AddressNet, recvMachineId []byte, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	if content == nil {
		return nil, utils.NewErrorSuccess()
	}
	strKey := utils.Bytes2string(recvid.Data())
	unlock := createHeMutex.Lock(strKey)
	defer unlock()
	//this.Log.Info().Str("获取棘轮密钥加密消息", "").Send()
	//这里查询缓存中的节点信息，拿到节点的超级节点信息
	ratchet := this.RatchetSession.GetSendRatchet(recvid, recvMachineId)

	//this.securityStoreLock.Lock()
	//this.securityStoreLock.Unlock()

	exist := true
	machineValue, ok := this.securityStore[strKey]
	if ok && machineValue != nil && len(recvMachineId) > 0 {
		_, exist = machineValue[utils.Bytes2string(recvMachineId)]
	}
	if ratchet == nil || !ok || !exist {
		//没有协商密钥
		//this.Log.Info().Str("获取棘轮密钥加密消息", "没有协商密钥").Send()
		//获取一个随机数
		sharedHkaRand, err := utils.Rand32Byte()
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		//使用随机数生成一个密钥对
		sharedHkaDH, ERR := keystore.GenerateKeyPair(sharedHkaRand[:])
		if ERR.CheckFail() {
			return nil, ERR
		}
		//获取一个随机数
		sharedNhkbRand, err := utils.Rand32Byte()
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		//使用随机数生成一个密钥对
		sharedNhkbDH, ERR := keystore.GenerateKeyPair(sharedNhkbRand[:])
		if ERR.CheckFail() {
			return nil, ERR
		}
		shareKey := ShareKey{
			Idinfo:   this.nodeManager.NodeSelf.IdInfo,
			A_DH_PUK: sharedHkaDH.GetPublicKey(),  //
			B_DH_PUK: sharedNhkbDH.GetPublicKey(), //B公钥
		}
		bs, err := shareKey.Proto() //
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}

		//this.Log.Info().Str("获取棘轮密钥加密消息", "").Send()
		//给目标节点发送自己的公钥，并获取对方身份公钥
		bs2, ERR := this.sendP2pMsgWaitRequest(config.MSGID_p2p_get_node_info, 0, recvid, &bs, timeout)
		//bs2, ERR := this.SendP2pMsgWaitRequest(config.MSGID_p2p_get_node_info, recvid, &bs, timeout)
		if ERR.CheckFail() {
			//this.Log.Info().Msgf("获取对方身份公钥失败:%s", ERR.String())
			return nil, ERR
		}

		//this.Log.Info().Str("获取棘轮密钥加密消息", "").Send()
		nodeInfo, err := nodeStore.ParseNodeProto(*bs2)
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		sni := SearchNodeInfo{
			Id: recvid,
			//SuperId: message.Head.SenderSuperId,
			CPuk: nodeInfo.IdInfo.CPuk,
		}
		recvMachineId = nodeInfo.MachineID

		// utils.Log.Info().Msgf("============ 发送加密消息 7777777777")
		//生成一个共享密钥Hka
		sharedHka, err := keystore.KeyExchange(keystore.NewDHPair(sharedHkaDH.GetPrivateKey(), sni.CPuk))
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}

		// utils.Log.Info().Msgf("============ 发送加密消息 99999")
		//使用自己的私钥和对方公钥生成共享密钥Nhkb
		sharedNhkb, err := keystore.KeyExchange(keystore.NewDHPair(sharedNhkbDH.GetPrivateKey(), sni.CPuk))
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}

		//生成共享密钥SK
		keyPairSelf, ERR := this.key.GetDhAddrKeyPair(this.pwd)
		if ERR.CheckFail() {
			return nil, ERR
		}
		sk, err := keystore.KeyExchange(keystore.NewDHPair(keyPairSelf.GetPrivateKey(), sni.CPuk))
		if err != nil {
			return nil, utils.NewErrorSysSelf(err)
		}
		//this.Log.Info().Msgf("获取棘轮密钥加密消息 sk:%s shka:%s snhkb:%s", hex.EncodeToString(sk[:]),
		//	hex.EncodeToString(sharedHka[:]), hex.EncodeToString(sharedNhkb[:]))
		err = this.RatchetSession.AddSendPipe(sni.Id, recvMachineId, sk, sharedHka, sharedNhkb, keyPairSelf)
		if err != nil {
			// utils.Log.Info().Msgf("============ 发送加密消息 1010101022222")
			return nil, utils.NewErrorSysSelf(err)
		}
		//
		machineValue = make(map[string]*SearchNodeInfo)
		machineValue[utils.Bytes2string(recvMachineId)] = &sni
		this.securityStore[strKey] = machineValue
	}
	//this.Log.Info().Str("获取棘轮密钥加密消息", "").Send()
	//utils.Log.Info().Msgf("============ 发送加密消息 22222")
	machineValue, ok = this.securityStore[strKey]
	if !ok {
		return nil, utils.NewErrorBus(config.ERROR_code_security_store_not_exist, "")
	}

	var sni *SearchNodeInfo
	if len(recvMachineId) > 0 {
		sni, ok = machineValue[utils.Bytes2string(recvMachineId)]
		if !ok {
			return nil, utils.NewErrorBus(config.ERROR_code_search_node_info_not_exist, "")
		}
	} else {
		for _, sni = range machineValue {
			if sni != nil {
				break
			}
		}
	}
	if sni == nil {
		return nil, utils.NewErrorBus(config.ERROR_code_security_store_not_exist, "")
	}
	//this.Log.Info().Msgf("通过棘轮加密:%s %s", sni.Id.B58String(), hex.EncodeToString(recvMachineId))
	//开始发送真正的消息
	//先将内容加密
	ratchet = this.RatchetSession.GetSendRatchet(sni.Id, recvMachineId)
	if ratchet == nil {
		// utils.Log.Info().Msgf("============ 发送加密消息 5555555")
		return nil, utils.NewErrorBus(config.ERROR_code_send_ratchet_not_exist, "") //errors.New("sendRatchet not exist")
	}
	//utils.Log.Info().Msgf("============ 发送加密消息 6666666")
	msgHE := ratchet.RatchetEncrypt(*content, nil)
	bs, err := config.ByteListProto([][][]byte{[][]byte{msgHE.Header, msgHE.Ciphertext}})
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return &bs, utils.NewErrorSuccess()
}

/*
对某个加密消息回复
*/
func (this *MessageCenter) SendP2pReplyMsgHE(message *MessageBase, content *[]byte) utils.ERROR {
	if !this.CheckOnline() {
		return utils.NewErrorBus(config.ERROR_code_offline, "")
	}
	//是自己发送给自己，则不用协商密钥
	if bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), message.GetBase().SenderAddr.Data()) {
		forwardMsg := NewMessageForwardReply(config.Version_p2p, message.GetBase(), this.nodeManager.NodeSelf.IdInfo.Id,
			this.nodeManager.NodeSelf.MachineID, content)
		this.handlerProcess(forwardMsg)
		return utils.NewErrorSuccess()
	}
	//加密
	ciphertext, ERR := this.encryptMessagesRatchet(message.SenderAddr, message.SenderMachineID, content, 0)
	if ERR.CheckFail() {
		return ERR
	}

	netResult := utils.NewNetResult(config.Version_1, ERR.Code, ERR.Msg, *ciphertext)
	bs, err := netResult.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	forwardMsg := NewMessageForwardReply(config.Version_p2pHE_wait, message, this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, bs)
	_, ERR = this.forward(forwardMsg, true, 0)
	if ERR.CheckFail() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
发送一个自定义消息
*/
func (this *MessageCenter) sendP2pMsgWaitRequest(msgID_engine, msgID_p2p uint64, recvid *nodeStore.AddressNet, content *[]byte,
	timeout time.Duration) (*[]byte, utils.ERROR) {
	if timeout == 0 {
		timeout = time.Minute
	}
	forwardMsg := NewMessageForward(msgID_engine, msgID_p2p, this.nodeManager.NodeSelf.IdInfo.Id,
		this.nodeManager.NodeSelf.MachineID, recvid, content)
	engine.RegisterRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
	defer engine.RemoveRequestKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID)
	isSelf, ERR := this.forward(forwardMsg, true, timeout)
	if ERR.CheckFail() {
		//this.Log.Info().Str("转发错误", ERR.String()).Send()
		return nil, ERR
	}
	//this.Log.Info().Str("发送消息", "111111").Send()
	if isSelf {
		//this.Log.Info().Str("是自己发送给自己", "").Send()
		this.handlerProcess(forwardMsg)
	}
	//this.Log.Info().Str("发送消息", "111111").Send()
	resultBs, ERR := engine.WaitResponseByteKey(config.Wait_major_p2p_sys_msg, forwardMsg.SendID, timeout)
	if ERR.CheckFail() {
		//this.Log.Info().Str("等待消息错误", ERR.String()).Send()
		return nil, ERR
	}
	return resultBs, utils.NewErrorSuccess()
}
