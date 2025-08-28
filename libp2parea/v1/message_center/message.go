package message_center

import (
	"github.com/gogo/protobuf/proto"
	"time"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

type MessageHead struct {
	RecvId           *nodeStore.AddressNet          `json:"r_id"`     //接收者id
	RecvSuperId      *nodeStore.AddressNet          `json:"r_s_id"`   //接收者的超级节点id
	RecvVnode        *virtual_node.AddressNetExtend `json:"r_v_id"`   //接收者虚拟节点id
	Sender           *nodeStore.AddressNet          `json:"s_id"`     //发送者id
	SenderSuperId    *nodeStore.AddressNet          `json:"s_s_id"`   //发送者超级节点id
	SenderVnode      *virtual_node.AddressNetExtend `json:"s_v_id"`   //发送者虚拟节点id
	Accurate         bool                           `json:"a"`        //是否准确发送给一个节点，如果
	OneByOne         bool                           `json:"onebyone"` //是否使用onebyone规则去路由
	RecvProxyId      *nodeStore.AddressNet          `json:"r_p_id"`   //接收者代理节点id
	SenderProxyId    *nodeStore.AddressNet          `json:"s_p_id"`   //发送者代理节点id
	SearchVnodeEndId *virtual_node.AddressNetExtend `json:"s_v_e_id"` //最终接收者虚拟节点id
	SelfVnodeId      *virtual_node.AddressNetExtend //查询磁力节点的时候，落到自己的哪个虚拟节点上。此字段不作序列化。
	SenderMachineID  string                         `json:"s_m_id"` // 发送者设备机器id
	RecvMachineID    string                         `json:"r_m_id"` // 接收者者设备机器id
}

func NewMessageHead(nodeSelf *nodeStore.Node, superId, recvid, recvSuperid *nodeStore.AddressNet, accurate bool, senderMachineId, recvMachineID string) *MessageHead {
	if nodeSelf.GetIsSuper() {
		//		head := NewMessageHead(nil, nil, nil, nodeStore.NodeSelf.IdInfo.Id, false)
		return &MessageHead{
			RecvId:          recvid,              //接收者id
			RecvSuperId:     recvSuperid,         //接收者的超级节点id
			Sender:          &nodeSelf.IdInfo.Id, //发送者id
			SenderSuperId:   &nodeSelf.IdInfo.Id, //发送者超级节点id
			Accurate:        accurate,            //是否准确发送给一个节点
			SenderMachineID: senderMachineId,     //发送者设备机器Id
			RecvMachineID:   recvMachineID,       //接收者设备机器Id
		}
	} else {
		return &MessageHead{
			RecvId:          recvid,              //接收者id
			RecvSuperId:     recvSuperid,         //接收者的超级节点id
			Sender:          &nodeSelf.IdInfo.Id, //发送者id
			SenderSuperId:   superId,             //发送者超级节点id
			Accurate:        accurate,            //是否准确发送给一个节点
			SenderMachineID: senderMachineId,     //发送者设备机器Id
			RecvMachineID:   recvMachineID,       //接收者设备机器Id
		}
	}
}

func NewMessageHeadOneByOne(nodeSelf *nodeStore.Node, superId, recvid, recvSuperid *nodeStore.AddressNet, accurate, oneByOne bool, senderMachineId, recvMachineID string) *MessageHead {
	if nodeSelf.GetIsSuper() {
		//		head := NewMessageHead(nil, nil, nil, nodeStore.NodeSelf.IdInfo.Id, false)
		return &MessageHead{
			RecvId:          recvid,              //接收者id
			RecvSuperId:     recvSuperid,         //接收者的超级节点id
			Sender:          &nodeSelf.IdInfo.Id, //发送者id
			SenderSuperId:   &nodeSelf.IdInfo.Id, //发送者超级节点id
			Accurate:        accurate,            //是否准确发送给一个节点
			OneByOne:        oneByOne,            // 是否采用oneByOne方式
			SenderMachineID: senderMachineId,     //发送者设备机器Id
			RecvMachineID:   recvMachineID,       //接收者设备机器Id
		}
	} else {
		return &MessageHead{
			RecvId:          recvid,              //接收者id
			RecvSuperId:     recvSuperid,         //接收者的超级节点id
			Sender:          &nodeSelf.IdInfo.Id, //发送者id
			SenderSuperId:   superId,             //发送者超级节点id
			Accurate:        accurate,            //是否准确发送给一个节点
			OneByOne:        oneByOne,            // 是否采用oneByOne方式
			SenderMachineID: senderMachineId,     //发送者设备机器Id
			RecvMachineID:   recvMachineID,       //接收者设备机器Id
		}
	}
}

// 创建含有接收者代理节点和发送者代理节点的消息头
func NewMessageHeadProxy(nodeSelf *nodeStore.Node, superId, recvid, recvSuperid, recvProxyid, senderProxyid *nodeStore.AddressNet, accurate bool, senderMachineId, recvMachineID string) *MessageHead {
	if nodeSelf.GetIsSuper() {
		//		head := NewMessageHead(nil, nil, nil, nodeStore.NodeSelf.IdInfo.Id, false)
		return &MessageHead{
			RecvId:          recvid,              //接收者id
			RecvSuperId:     recvSuperid,         //接收者的超级节点id
			Sender:          &nodeSelf.IdInfo.Id, //发送者id
			SenderSuperId:   &nodeSelf.IdInfo.Id, //发送者超级节点id
			Accurate:        accurate,            //是否准确发送给一个节点
			RecvProxyId:     recvProxyid,         //接收者代理节点id
			SenderProxyId:   senderProxyid,       //发送者代理节点id
			SenderMachineID: senderMachineId,     //发送者设备机器Id
			RecvMachineID:   recvMachineID,       //接收者设备机器Id
		}
	} else {
		return &MessageHead{
			RecvId:          recvid,              //接收者id
			RecvSuperId:     recvSuperid,         //接收者的超级节点id
			Sender:          &nodeSelf.IdInfo.Id, //发送者id
			SenderSuperId:   superId,             //发送者超级节点id
			Accurate:        accurate,            //是否准确发送给一个节点
			RecvProxyId:     recvProxyid,         //接收者代理节点id
			SenderProxyId:   senderProxyid,       //发送者代理节点id
			SenderMachineID: senderMachineId,     //发送者设备机器Id
			RecvMachineID:   recvMachineID,       //接收者设备机器Id
		}
	}
}

/*
创建一个虚拟节点消息
*/
func NewMessageHeadVnode(senderId *nodeStore.AddressNet, sendVid, recvVid *virtual_node.AddressNetExtend, recvProxyid, senderProxyid *nodeStore.AddressNet, accurate bool, senderMachineId, recvMachineID string) *MessageHead {
	return &MessageHead{
		// RecvId:        recvid,                        //接收者id
		// RecvSuperId:   recvSuperid,                   //接收者的超级节点id
		RecvVnode: recvVid,  //
		Sender:    senderId, //发送者id
		// SenderSuperId: senderId, //发送者超级节点id
		SenderVnode:     sendVid,         //
		Accurate:        accurate,        //是否准确发送给一个节点
		RecvProxyId:     recvProxyid,     //接收者代理节点id
		SenderProxyId:   senderProxyid,   //发送者代理节点id
		SenderMachineID: senderMachineId, //发送者设备机器Id
		RecvMachineID:   recvMachineID,   //接收者设备机器Id
	}
}

/*
检查参数是否合法
*/
func (this *MessageHead) Check() bool {
	if this.RecvId == nil {
		return false
	}
	if this.RecvSuperId == nil {
		return false
	}
	if this.Sender == nil {
		return false
	}
	if this.SenderSuperId == nil {
		return false
	}
	return true
}

func (this *MessageHead) Proto() []byte {
	mhp := new(go_protobuf.MessageHead)
	mhp.Accurate = this.Accurate
	mhp.OneByOne = this.OneByOne
	if this.RecvId != nil {
		mhp.RecvId = *this.RecvId
	}

	if this.RecvSuperId != nil {
		mhp.RecvSuperId = *this.RecvSuperId
	}

	if this.RecvVnode != nil {
		mhp.RecvVnode = *this.RecvVnode
	}

	if this.Sender != nil {
		mhp.Sender = *this.Sender
	}

	if this.SenderSuperId != nil {
		mhp.SenderSuperId = *this.SenderSuperId
	}

	if this.SenderVnode != nil {
		mhp.SenderVnode = *this.SenderVnode
	}

	if this.RecvProxyId != nil {
		mhp.RecvProxyId = *this.RecvProxyId
	}

	if this.SenderProxyId != nil {
		mhp.SenderProxyId = *this.SenderProxyId
	}

	if this.SearchVnodeEndId != nil {
		mhp.SearchVnodeEndId = *this.SearchVnodeEndId
	}

	if this.SenderMachineID != "" {
		mhp.SenderMachineID = this.SenderMachineID
	}

	if this.RecvMachineID != "" {
		mhp.RecvMachineID = this.RecvMachineID
	}

	bs, _ := mhp.Marshal()
	return bs
}

type MessageBody struct {
	MessageId  uint64  `json:"m_id"`    //消息协议编号
	CreateTime uint64  `json:"c_time"`  //消息创建时间unix
	ReplyTime  uint64  `json:"r_time"`  //消息回复时间unix
	Hash       []byte  `json:"hash"`    //消息的hash值
	ReplyHash  []byte  `json:"r_hash"`  //回复消息的hash
	SendRand   uint64  `json:"s_rand"`  //发送随机数
	RecvRand   uint64  `json:"r_rand"`  //接收随机数
	Content    *[]byte `json:"content"` //发送的内容
}

func NewMessageBody(msgid uint64, content *[]byte, creatTime uint64, hash []byte, sendRand uint64) *MessageBody {
	return &MessageBody{
		MessageId:  msgid,
		CreateTime: creatTime,
		Hash:       hash,
		SendRand:   sendRand,
		Content:    content, //发送的内容
	}
}

func (this *MessageBody) Proto() ([]byte, error) {
	mbp := go_protobuf.MessageBody{
		MessageId:  this.MessageId,
		CreateTime: this.CreateTime,
		ReplyTime:  this.ReplyTime,
		Hash:       this.Hash,
		ReplyHash:  this.ReplyHash,
		SendRand:   this.SendRand,
		RecvRand:   this.RecvRand,
	}

	if this.Content != nil {
		mbp.Content = *this.Content
	}
	bs, err := mbp.Marshal()
	return bs, err
}

/*
发送消息序列化对象
*/
type Message struct {
	msgid    uint64       //
	Head     *MessageHead `json:"head"` //
	Body     *MessageBody `json:"body"` //
	DataPlus *[]byte      `json:"dp"`   //body部分加密数据，消息路由时候不需要解密，临时保存
}

func (this *Message) Proto() ([]byte, error) {
	head := go_protobuf.MessageHead{
		Accurate: this.Head.Accurate,
		OneByOne: this.Head.OneByOne,
	}
	if this.Head.RecvId != nil {
		head.RecvId = *this.Head.RecvId
	}
	if this.Head.RecvSuperId != nil {
		head.RecvSuperId = *this.Head.RecvSuperId
	}
	if this.Head.RecvVnode != nil {
		head.RecvVnode = *this.Head.RecvVnode
	}
	if this.Head.Sender != nil {
		head.Sender = *this.Head.Sender
	}
	if this.Head.SenderSuperId != nil {
		head.SenderSuperId = *this.Head.SenderSuperId
	}
	if this.Head.SenderVnode != nil {
		head.SenderVnode = *this.Head.SenderVnode
	}
	if this.Head.RecvProxyId != nil {
		head.RecvProxyId = *this.Head.RecvProxyId
	}
	if this.Head.SenderProxyId != nil {
		head.SenderProxyId = *this.Head.SenderProxyId
	}
	if this.Head.SearchVnodeEndId != nil {
		head.SearchVnodeEndId = *this.Head.SearchVnodeEndId
	}
	if this.Head.SenderMachineID != "" {
		head.SenderMachineID = this.Head.SenderMachineID
	}
	if this.Head.RecvMachineID != "" {
		head.RecvMachineID = this.Head.RecvMachineID
	}
	body := go_protobuf.MessageBody{
		MessageId:  this.Body.MessageId,
		CreateTime: this.Body.CreateTime,
		ReplyTime:  this.Body.ReplyTime,
		Hash:       this.Body.Hash,
		ReplyHash:  this.Body.ReplyHash,
		SendRand:   this.Body.SendRand,
		RecvRand:   this.Body.RecvRand,
	}
	if this.Body.Content != nil {
		body.Content = *this.Body.Content
	}
	message := go_protobuf.Message{
		Head: &head,
		Body: &body,
	}
	if this.DataPlus != nil {
		message.DataPlus = *this.DataPlus
	}
	return message.Marshal()
}
func (this *Message) BuildHash() {
	if this.Body.Hash != nil && len(this.Body.Hash) > 0 {
		return
	}
	this.Body.ReplyHash = nil
	this.Body.Hash = nil
	if this.Body.SendRand == 0 {
		this.Body.SendRand = utils.GetAccNumber()
	}
	this.Body.RecvRand = 0
	this.Body.CreateTime = uint64(time.Now().Unix()) //
	bs, _ := this.Proto()                            //
	this.Body.Hash = utils.Hash_SHA3_256(bs)
}
func (this *Message) BuildReplyHash(createtime uint64, sendhash []byte, sendrand uint64) error {
	if this.Body.ReplyHash != nil && len(this.Body.ReplyHash) > 0 {
		return nil
	}
	this.Body.CreateTime = createtime
	this.Body.Hash = sendhash
	this.Body.SendRand = sendrand
	this.Body.ReplyHash = nil
	if this.Body.RecvRand == 0 {
		this.Body.RecvRand = utils.GetAccNumber()
	}
	this.Body.ReplyTime = uint64(time.Now().Unix()) //utils.TimeFormatToNanosecond()
	bs, err := this.Proto()                         //
	if err != nil {
		return nil
	}
	this.Body.ReplyHash = utils.Hash_SHA3_256(bs)
	return nil
}

/*
	解析内容
*/
// func (this *Message) ParserContent() error {
// 	//TODO 解密内容

// 	this.Body = new(MessageBody)
// 	// err := json.Unmarshal(*this.DataPlus, this.Body)
// 	decoder := json.NewDecoder(bytes.NewBuffer(*this.DataPlus))
// 	decoder.UseNumber()
// 	err := decoder.Decode(this.Body)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

/*
	验证hash
*/
// func (this *Message) CheckSendhash() bool {
// 	//TODO 验证sendhash是否正确
// 	//TODO 验证时间不能相差太远

// 	//验证sendhash是否已经接受过此消息
// 	return CheckHash(this.Body.Hash)
// }

/*
	验证hash
*/
// func (this *Message) CheckReplyHash() bool {
// 	//TODO 验证sendhash是否正确
// 	//TODO 验证时间不能相差太远

// 	//验证sendhash是否已经接受过此消息
// 	return CheckHash(this.Body.ReplyHash)
// }

/*
解析内容
*/
func (this *Message) ParserContentProto() error {
	if this.DataPlus == nil {
		return nil
	}
	mbp := new(go_protobuf.MessageBody)
	err := proto.Unmarshal(*this.DataPlus, mbp)
	if err != nil {
		return err
	}
	this.Body = &MessageBody{
		MessageId:  mbp.MessageId,
		CreateTime: mbp.CreateTime,
		ReplyTime:  mbp.ReplyTime,
		Hash:       mbp.Hash,
		ReplyHash:  mbp.ReplyHash,
		SendRand:   mbp.SendRand,
		RecvRand:   mbp.RecvRand,
		// Content:    mbp.Content,
	}
	if mbp.Content != nil && len(mbp.Content) > 0 {
		this.Body.Content = &mbp.Content
	}

	// err := json.Unmarshal(*this.DataPlus, this.Body)
	// decoder := json.NewDecoder(bytes.NewBuffer(*this.DataPlus))
	// decoder.UseNumber()
	// err := decoder.Decode(this.Body)
	// if err != nil {
	// 	return err
	// }
	return nil
}

/*
	验证hash
*/
// func (this *Message) CheckReplyhash() bool {
// 	//TODO 验证replyhash是否正确
// 	//TODO 验证时间不能相差太远

// 	//验证replyhash是否已经接受过此消息
// 	return CheckHash(this.Body.ReplyHash)
// }

/*
检查该消息是否是自己的
不是自己的则自动转发出去
@return    sendOk    bool    是否发送给其他人。true=发送给其他人了;false=自己的消息;
*/
func (this *Message) Reply(version uint64, nodeManager *nodeStore.NodeManager,
	sessionEngine *engine.Engine, vnc *virtual_node.VnodeManager) (bool, error) {

	err := this.BuildReplyHash(this.Body.CreateTime, this.Body.Hash, this.Body.SendRand)
	if err != nil {
		utils.Log.Info().Msgf("reply build hash error:%s", err.Error())
		return true, err
	}
	// bs, _ := this.Head.JSON()
	// utils.Log.Info().Msgf("Reply路由消息类型:%s", string(bs))

	return this.Send(version, nodeManager, sessionEngine, vnc, 0)

	// if nodeManager.NodeSelf.GetIsSuper() {
	// 	// return IsSendToOtherSuperToo(this.Head, this.Body.JSON(), version, nil)
	// 	mbodyBs, err := this.Body.Proto()
	// 	if err != nil {
	// 		utils.Log.Info().Msgf("proto error:%s", err.Error())
	// 		return true, err
	// 	}
	// 	ok, err := IsSendToOtherSuperToo(this.Head, &mbodyBs, version, nil, nodeManager, sessionEngine, 0)
	// 	// utils.Log.Info().Msgf("Reply IsSendToOtherSuperToo:%t", ok)
	// 	return ok, err
	// } else {
	// 	if nodeManager.SuperPeerId == nil {
	// 		// utils.Log.Info().Msgf("没有可用的超级节点")
	// 		return true, nil
	// 	}
	// 	// if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
	// 	if session, ok := sessionEngine.GetSession(utils.Bytes2string(*nodeManager.GetSuperPeerId())); ok {
	// 		// session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
	// 		mheadBs := this.Head.Proto()
	// 		mbodyBs, err := this.Body.Proto()
	// 		if err != nil {
	// 			return true, err
	// 		}
	// 		err = session.Send(version, &mheadBs, &mbodyBs, 0)
	// 	} else {
	// 		// fmt.Println("超级节点的session未找到")
	// 	}
	// 	return true, err
	// }
}

func NewMessage(head *MessageHead, body *MessageBody) *Message {
	return &Message{
		Head: head,
		Body: body,
	}
}

// func ParserMessage(data, dataplus *[]byte, msgId uint64) (*Message, error) {
// 	head := new(MessageHead)
// 	// err := json.Unmarshal(*data, head)
// 	decoder := json.NewDecoder(bytes.NewBuffer(*data))
// 	decoder.UseNumber()
// 	err := decoder.Decode(head)
// 	if err != nil {
// 		return nil, err
// 	}

// 	msg := Message{
// 		msgid:    msgId,
// 		Head:     head,
// 		DataPlus: dataplus,
// 	}
// 	return &msg, nil
// }

func ParserMessageProto(data, dataplus []byte, msgId uint64) (*Message, error) {
	mhp := new(go_protobuf.MessageHead)
	err := proto.Unmarshal(data, mhp)
	if err != nil {
		return nil, err
	}

	head := &MessageHead{
		// RecvId:        nodeStore.AddressNet(mhp.RecvId),
		// RecvSuperId:   nodeStore.AddressNet(mhp.RecvSuperId),
		// RecvVnode:     virtual_node.AddressNetExtend(mhp.RecvVnode),
		// Sender:        nodeStore.AddressNet(mhp.Sender),
		// SenderSuperId: nodeStore.AddressNet(mhp.SenderSuperId),
		// SenderVnode:   virtual_node.AddressNetExtend(mhp.SenderVnode),
		Accurate: mhp.Accurate,
		OneByOne: mhp.OneByOne,
	}

	if mhp.RecvId != nil && len(mhp.RecvId) > 0 {
		recvId := nodeStore.AddressNet(mhp.RecvId)
		head.RecvId = &recvId
	}
	if mhp.RecvSuperId != nil && len(mhp.RecvSuperId) > 0 {
		recvSuperId := nodeStore.AddressNet(mhp.RecvSuperId)
		head.RecvSuperId = &recvSuperId
	}
	if mhp.RecvVnode != nil && len(mhp.RecvVnode) > 0 {
		recvVnode := virtual_node.AddressNetExtend(mhp.RecvVnode)
		head.RecvVnode = &recvVnode
	}
	if mhp.Sender != nil && len(mhp.Sender) > 0 {
		sender := nodeStore.AddressNet(mhp.Sender)
		head.Sender = &sender
	}
	if mhp.SenderSuperId != nil && len(mhp.SenderSuperId) > 0 {
		senderSuperId := nodeStore.AddressNet(mhp.SenderSuperId)
		head.SenderSuperId = &senderSuperId
	}
	if mhp.SenderVnode != nil && len(mhp.SenderVnode) > 0 {
		senderVnode := virtual_node.AddressNetExtend(mhp.SenderVnode)
		head.SenderVnode = &senderVnode
	}
	if mhp.RecvProxyId != nil && len(mhp.RecvProxyId) > 0 {
		recvProxyId := nodeStore.AddressNet(mhp.RecvProxyId)
		head.RecvProxyId = &recvProxyId
	}
	if mhp.SenderProxyId != nil && len(mhp.SenderProxyId) > 0 {
		senderProxyId := nodeStore.AddressNet(mhp.SenderProxyId)
		head.SenderProxyId = &senderProxyId
	}
	if mhp.SearchVnodeEndId != nil && len(mhp.SearchVnodeEndId) > 0 {
		searchVnodeEndId := virtual_node.AddressNetExtend(mhp.SearchVnodeEndId)
		head.SearchVnodeEndId = &searchVnodeEndId
	}
	if mhp.SenderMachineID != "" {
		head.SenderMachineID = mhp.SenderMachineID
	}
	if mhp.RecvMachineID != "" {
		head.RecvMachineID = mhp.RecvMachineID
	}

	msg := Message{
		msgid:    msgId,
		Head:     head,
		DataPlus: &dataplus,
	}
	return &msg, nil
}

/*
	得到一条消息的hash值
*/
//func GetHash(msg *Message) string {
//	hash := sha256.New()
//	hash.Write([]byte(msg.RecvId))
//	//	binary.Write(hash, binary.LittleEndian, uint64(msg.ProtoId))
//	binary.Write(hash, binary.LittleEndian, msg.CreateTime)
//	// hash.Write([]byte(int64(msg.ProtoId)))
//	// hash.Write([]byte(msg.CreateTime))
//	hash.Write([]byte(msg.Sender))
//	// hash.Write([]byte(msg.RecvTime))
//	binary.Write(hash, binary.LittleEndian, msg.ReplyTime)
//	hash.Write(msg.Content)
//	hash.Write([]byte(msg.ReplyHash))
//	return hex.EncodeToString(hash.Sum(nil))
//}

/*
消息超时删除md5
*/
func msgTimeOutProsess(class, params string) {
	switch class {
	case config.TSK_msg_timeout_remove: //删除超时的消息md5
		//		fmt.Println("开始删除临时域名", tempName)
		//		tempNameLock.Lock()
		//		delete(tempName, params)
		//		tempNameLock.Unlock()
		//		fmt.Println("删除了这个临时域名", params, tempName)
	default:
		//		//剩下是需要更新的域名
		//		flashName := FlashName{
		//			Name:  params,
		//			Class: class,
		//		}
		//		OutFlashName <- &flashName
	}

}
