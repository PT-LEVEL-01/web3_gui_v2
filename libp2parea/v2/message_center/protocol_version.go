package message_center

import (
	"encoding/hex"
	"time"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/message_center/doubleratchet"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func (this *MessageCenter) RegisterMsgVersion() {
	this.sessionEngine.RegisterMsg(config.Version_neighbor, this.NeighborHandler)                 //邻居节点消息，本消息不转发
	this.sessionEngine.RegisterMsg(config.Version_multicast_head, this.multicastHandler)          //广播消息hash
	this.sessionEngine.RegisterMsg(config.Version_multicast_body, this.multicastSyncHandler)      //通过广播消息hash值查询邻居节点的广播消息内容
	this.sessionEngine.RegisterMsg(config.Version_multicast_recv, this.multicastSyncHandler_recv) //通过广播消息hash值查询邻居节点的广播消息内容_返回
	this.sessionEngine.RegisterMsg(config.Version_p2p, this.p2pPlaintextHandler)                  //点对点明文消息
	//this.sessionEngine.RegisterMsg(config.Version_p2pHE, this.p2pCiphertextHandler)                    //点对点加密消息
	this.sessionEngine.RegisterMsg(config.Version_p2pHE_wait, this.p2pCiphertextWaitHandler) //点对点加密消息，需要等待返回消息
	//this.sessionEngine.RegisterMsg(config.Version_p2pHE_wait_recv, this.p2pCiphertextWaitHandler_recv) //点对点加密消息，需要等待返回消息_返回
	this.sessionEngine.RegisterMsg(config.MSGID_p2p_get_node_info, this.getNodeInfo) //获取对方的公钥
}

/*
统一解析消息和转发
@return    message_bean.MessageItr    为空则转发出去了，有值则是发送给自己，需要自己处理
*/
func (this *MessageCenter) handlerBase(packet *engine.Packet) *MessageBase {
	//this.Log.Info().Str("接收到消息", "").Send()
	message, err := ParseMessageBase(&packet.Data) //message_bean.ParseMessage(packet.MsgID, packet.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil
	}
	//this.Log.Info().Str("接收到消息", "").Send()
	message.SetPacket(packet)
	message.MsgEngineID = packet.MsgID
	//当接收者为空，则是邻居节点消息，不做转发
	if message.RecvAddr == nil || len(message.RecvAddr.Data()) == 0 {
		return message
	}
	//this.Log.Info().Str("接收到消息", "").Send()
	if !this.checkSelf(message) {
		//this.Log.Info().Str("接收到消息", "").Send()
		//不是自己接收，则转发出去
		this.forward(message, false, 0)
		return nil
	}
	//this.Log.Info().Str("接收到消息", "").Send()
	//是发送给自己的回复消息
	//if len(message.ReplyID) > 0 {
	//	//回复消息
	//	engine.ResponseByteKey(config.Wait_major_p2p_sys_msg, message.SendID, &message.Content)
	//	return nil
	//}
	//是发送给自己
	return message
}

/*
邻居节点消息控制器，不转发消息
*/
func (this *MessageCenter) NeighborHandler(packet *engine.Packet) {
	messageItr := this.handlerBase(packet)
	this.handlerProcess(messageItr)
}

/*
广播消息控制器
*/
func (this *MessageCenter) multicastHandler(msg *engine.Packet) {
	//this.Log.Info().Str("有广播消息", "").Send()
	messageItr, err := ParseMessageBase(&msg.Data) //message_bean.ParseMessage(config.Version_multicast, msg.Data)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	//this.Log.Info().Hex("有广播消息 hash:", messageItr.Content).Send()
	messageItr.SetPacket(msg)
	messageItr.MsgEngineID = msg.MsgID
	//先同步，同步了再广播给其他节点
	mhOne := MsgHolderOne{
		MsgHash:        messageItr.Content,
		Addr:           *messageItr.SenderAddr, //  nodeStore.AddressNet([]byte(msg.Session.GetName())), // *message.Head.SenderSuperId
		ForwordMessage: *messageItr,
		Message:        messageItr,
		Session:        msg.Session,
	}
	this.msgchannl.Add(&mhOne)
	return
}

/*
查询广播消息内容
*/
func (this *MessageCenter) multicastSyncHandler(msg *engine.Packet) {
	//this.Log.Info().Msgf("来获取广播消息内容了")
	messageItr, err := ParseMessageBase(&msg.Data) //
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	messageItr.SetPacket(msg)
	messageItr.MsgEngineID = msg.MsgID
	//this.Log.Info().Hex("来获取广播消息内容了 hash:", messageItr.Content).Send()
	content, err := FindMessageCacheByHash(messageItr.GetBase().Content, this.levelDB)
	if err != nil {
		this.Log.Error().Msgf("get multicast message :%s error:%s", hex.EncodeToString(messageItr.GetBase().Content), err.Error())
		return
	}
	//this.Log.Info().Hex("来获取广播消息内容了", messageItr.Content).Int("长度", len(*content)).Send()
	if content == nil {
		return
	}

	//this.Log.Info().Hex("来获取广播消息内容了", messageItr.Content).Send()
	message := NewMessageBase(0, this.nodeManager.NodeSelf.IdInfo.Id, nil,
		nil, nil, nil, nil, nil, nil, 0,
		nil, nil, content)
	bs, err := message.Proto()
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}

	//this.Log.Info().Hex("来获取广播消息内容了 hash", messageItr.Content).Int("内容长度", len(*bs)).Send()
	ERR := msg.Session.Reply(msg, bs, time.Minute)
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
	}
	return
}

/*
查询广播消息内容 返回
*/
func (this *MessageCenter) multicastSyncHandler_recv(packet *engine.Packet) {
	messageItr := this.handlerBase(packet)
	if messageItr == nil {
		return
	}
	//是自己的消息
	this.Log.Info().Msgf("有广播消息")
	messageItr, err := ParseMessageBase(&packet.Data) //
	if err != nil {
		this.Log.Error().Err(err).Send()
		return
	}
	engine.ResponseByteKey(config.Wait_major_p2p_sys_msg, messageItr.SendID, &messageItr.Content)
	return
}

/*
点对点消息控制器
*/
func (this *MessageCenter) p2pPlaintextHandler(packet *engine.Packet) {
	//this.Log.Info().Str("收到一条p2p明文消息", "").Send()
	messageItr := this.handlerBase(packet)
	this.handlerProcess(messageItr)
}

/*
点对点加密消息控制器
*/
//func (this *MessageCenter) p2pCiphertextHandler(packet *engine.Packet) {
//	this.Log.Info().Str("收到加密消息", "").Send()
//	//解析消息
//	message := this.handlerBase(packet)
//	//返回空，则代表转发出去了
//	if message == nil {
//		//this.Log.Info().Str("收到加密消息", "").Send()
//		return
//	}
//	//自己的消息
//	lists, err := config.ParseByteList(&message.Content)
//	if err != nil {
//		this.Log.Error().Err(err).Send()
//		return
//	}
//	this.Log.Info().Str("收到加密消息", "").Send()
//
//	//解密消息
//	//headSize := binary.LittleEndian.Uint64(packet.Data[:8])
//	//headData := packet.Data[8 : headSize+8]
//	//bodyData := packet.Data[headSize+8+8:]
//	msgHE := doubleratchet.MessageHE{
//		Header:     lists[0][0], //headData,
//		Ciphertext: lists[0][1], //bodyData,
//	}
//	sessionHE := this.RatchetSession.GetRecvRatchet(message.SenderAddr, utils.Bytes2string(message.SenderMachineID))
//	if sessionHE == nil {
//		//this.Log.Info().Str("收到加密消息", "").Send()
//		//双棘轮key未找到
//		nr := utils.NewNetResult(config.Version_1, config.ERROR_code_recv_security_store_not_exist, "", this.nodeManager.NodeSelf.MachineID)
//		bs, err := nr.Proto()
//		if err != nil {
//			utils.Log.Error().Err(err).Send()
//			return
//		}
//		this.SendP2pReplyMsg(message, bs)
//		return
//	}
//	this.Log.Info().Str("收到加密消息", "").Send()
//	//解密消息
//	bs, err := sessionHE.RatchetDecrypt(msgHE, nil)
//	if err != nil {
//		this.Log.Info().Str("收到加密消息", "").Send()
//		//解密消息出错
//		nr := utils.NewNetResult(config.Version_1, config.ERROR_code_recv_ratchet_not_exist, "", this.nodeManager.NodeSelf.MachineID)
//		bs, err := nr.Proto()
//		if err != nil {
//			utils.Log.Error().Err(err).Send()
//			return
//		}
//		this.SendP2pReplyMsg(message, bs)
//		return
//	}
//	message.Content = bs
//	this.Log.Info().Interface("收到加密消息", message.ConvertVO()).Send()
//
//	//执行消息号对应的handler
//	this.handlerProcess(message)
//}

//var tempLock = new(sync.RWMutex)

/*
点对点加密消息控制器，需要等待消息返回
*/
func (this *MessageCenter) p2pCiphertextWaitHandler(packet *engine.Packet) {
	//解析消息
	message := this.handlerBase(packet)
	//返回空，则代表转发出去了
	if message == nil {
		//this.Log.Info().Str("收到加密消息", "").Send()
		return
	}
	if len(message.ReplyID) > 0 {
		engine.ResponseItrKey(config.Wait_major_p2p_sys_msg, message.SendID, message)
		return
	}
	//重复消息
	if this.CheckRepeatHash(message.SendID) {
		return
	}
	//tempLock.Lock()
	//defer tempLock.Unlock()
	//this.Log.Info().Hex("p2pCiphertextWaitHandler收到加密等待消息", packet.GetSendId()).Str("远端地址", packet.Session.GetRemoteHost()).Send()
	//解密消息内容
	bs, ERR := this.decryptMessage(message.SenderAddr, message.SenderMachineID, &message.Content)
	if ERR.CheckFail() {
		if ERR.Code == config.ERROR_code_recv_security_store_not_exist {
			//节点重启，会造成缓存中未找到接收加密会话棘轮，返回一个错误
		}
		if len(message.ReplyID) > 0 {
			utils.Log.Info().Msgf("是返回的消息解密不了")
			return
		}

		//this.Log.Error().Str("收到加密等待消息", ERR.String()).Send()
		//回复消息：解密失败
		nr := utils.NewNetResult(config.Version_1, ERR.Code, ERR.Msg, this.nodeManager.NodeSelf.MachineID)
		bs, err := nr.Proto()
		if err != nil {
			utils.Log.Error().Err(err).Send()
			return
		}
		forwardMsg := NewMessageForwardReply(config.Version_p2pHE_wait, message, this.nodeManager.NodeSelf.IdInfo.Id,
			this.nodeManager.NodeSelf.MachineID, bs)
		_, ERR = this.forward(forwardMsg, true, 0)
		if ERR.CheckFail() {
			this.Log.Error().Str("转发失败", ERR.String()).Send()
			return
		}
		//this.SendP2pReplyMsg(message, bs)
		return
	}

	//this.Log.Info().Str("收到加密等待消息", "").Send()
	//自己的消息
	message.Content = *bs
	//this.Log.Info().Interface("收到加密消息", message.ConvertVO()).Send()

	//执行消息号对应的handler
	this.handlerProcess(message)
}

/*
点对点加密消息控制器，需要等待消息返回
*/
func (this *MessageCenter) p2pCiphertextWaitHandler_recv(packet *engine.Packet) {
	//this.Log.Info().Str("收到加密消息", "").Send()
	//解析消息
	message := this.handlerBase(packet)
	//返回空，则代表转发出去了
	if message == nil {
		//this.Log.Info().Str("收到加密消息", "").Send()
		return
	}
	//np, err := utils.ParseNetParams(message.Content)
	//if err != nil {
	//	utils.Log.Error().Msgf("错误:%s", ERR.String())
	//	return
	//}
	//自己的消息
	engine.ResponseItrKey(config.Wait_major_p2p_sys_msg, message.SendID, message)
	return
}

/*
获取节点地址和身份公钥
*/
func (this *MessageCenter) getNodeInfo(packet *engine.Packet) {
	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	//if this.CheckRepeatHash(message.Body.Hash) {
	//	// utils.Log.Info().Msgf("验证返回消息错误")
	//	return
	//}
	message := this.handlerBase(packet)
	//返回空，则代表转发出去了
	if message == nil {
		return
	}
	//this.Log.Info().Interface("收到获取身份公钥消息", message.ConvertVO()).Send()
	//不为空则代表发送给自己，当回复id不为空，则代表是回复消息
	if len(message.ReplyID) > 0 {
		//this.Log.Info().Msgf("收到对方公钥回复消息")
		//回复消息
		engine.ResponseByteKey(config.Wait_major_p2p_sys_msg, message.SendID, &message.Content)
		return
	}

	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	//不为空则代表是发送给自己的消息
	shareKey, err := ParseShareKey(message.Content)
	if err != nil {
		return
	}

	//控制一个用户10秒内只能协商一次密钥
	if this.CheckHaveMsgHash(shareKey) {
		return
	}

	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	keyPair, ERR := this.key.GetDhAddrKeyPair(this.pwd)
	if ERR.CheckFail() {
		return
	}
	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	sk, err := keystore.KeyExchange(keystore.NewDHPair(keyPair.GetPrivateKey(), shareKey.Idinfo.CPuk))
	if err != nil {
		return
	}
	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	sharedHka, err := keystore.KeyExchange(keystore.NewDHPair(keyPair.GetPrivateKey(), shareKey.A_DH_PUK))
	if err != nil {
		return
	}
	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	sharedNhkb, err := keystore.KeyExchange(keystore.NewDHPair(keyPair.GetPrivateKey(), shareKey.B_DH_PUK))
	if err != nil {
		return
	}
	//this.Log.Info().Msgf("收到对方公钥并创建加密通道sk:%s shka:%s snhkb:%s", hex.EncodeToString(sk[:]),
	//	hex.EncodeToString(sharedHka[:]), hex.EncodeToString(sharedNhkb[:]))
	err = this.RatchetSession.AddRecvPipe(message.GetBase().SenderAddr, message.GetBase().SenderMachineID,
		sk, sharedHka, sharedNhkb, shareKey.Idinfo.CPuk)
	if err != nil {
		return
	}
	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	//回复消息
	data, err := this.nodeManager.NodeSelf.Proto()
	//data, err := this.nodeManager.NodeSelf.IdInfo.Proto()
	if err != nil {
		this.Log.Error().Msg(err.Error())
		return
	}
	//this.Log.Info().Msgf("收到对方公钥并创建加密通道")
	ERR = this.SendP2pReplyMsg(message, data)
	//ERR = packet.Reply(&data, 0)
	if ERR.CheckFail() {
		this.Log.Error().Str("ERR", ERR.String()).Send()
	}
}

/*
解密对方发送过来的加密消息
*/
func (this *MessageCenter) decryptMessage(senderAddr *nodeStore.AddressNet, senderMachineID []byte, content *[]byte) (*[]byte, utils.ERROR) {
	lists, err := config.ParseByteList(content)
	if err != nil {
		this.Log.Error().Err(err).Send()
		return nil, utils.NewErrorSysSelf(err)
	}
	//解密消息
	msgHE := doubleratchet.MessageHE{
		Header:     lists[0][0], //headData,
		Ciphertext: lists[0][1], //bodyData,
	}
	//this.Log.Info().Msgf("开始解密消息:%s %s", senderAddr.B58String(), hex.EncodeToString(senderMachineID))
	sessionHE := this.RatchetSession.GetRecvRatchet(senderAddr, utils.Bytes2string(senderMachineID))
	if sessionHE == nil {
		//双棘轮key未找到
		//this.Log.Info().Str("收到加密消息", "").Send()
		return nil, utils.NewErrorBus(config.ERROR_code_recv_security_store_not_exist, "")
	}
	//this.Log.Info().Str("收到加密消息", "").Send()
	//解密消息
	bs, err := sessionHE.RatchetDecrypt(msgHE, nil)
	if err != nil {
		//解密消息出错
		this.Log.Error().Str("解密消息 错误", err.Error()).Send()
		return nil, utils.NewErrorSysSelf(err)
	}
	return &bs, utils.NewErrorSuccess()
}
