package message_center

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/utils"
)

/*
发送消息序列化对象
*/
type Message struct {
	msgid    uint64                      //
	Head     *MessageHead                `json:"head"` //
	Body     *MessageBody                `json:"body"` //
	DataPlus *[]byte                     `json:"dp"`   //body部分加密数据，消息路由时候不需要解密，临时保存
	Message  *message_center.MessageBase //
}

type MessageHead struct {
	RecvId      *nodeStore.AddressNet `json:"r_id"`   //接收者id
	RecvSuperId *nodeStore.AddressNet `json:"r_s_id"` //接收者的超级节点id
	//RecvVnode        *virtual_node.AddressNetExtend `json:"r_v_id"`   //接收者虚拟节点id
	Sender        *nodeStore.AddressNet `json:"s_id"`   //发送者id
	SenderSuperId *nodeStore.AddressNet `json:"s_s_id"` //发送者超级节点id
	//SenderVnode      *virtual_node.AddressNetExtend `json:"s_v_id"`   //发送者虚拟节点id
	Accurate      bool                  `json:"a"`        //是否准确发送给一个节点，如果
	OneByOne      bool                  `json:"onebyone"` //是否使用onebyone规则去路由
	RecvProxyId   *nodeStore.AddressNet `json:"r_p_id"`   //接收者代理节点id
	SenderProxyId *nodeStore.AddressNet `json:"s_p_id"`   //发送者代理节点id
	//SearchVnodeEndId *virtual_node.AddressNetExtend `json:"s_v_e_id"` //最终接收者虚拟节点id
	//SelfVnodeId      *virtual_node.AddressNetExtend //查询磁力节点的时候，落到自己的哪个虚拟节点上。此字段不作序列化。
	SenderMachineID string `json:"s_m_id"` // 发送者设备机器id
	RecvMachineID   string `json:"r_m_id"` // 接收者者设备机器id
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

/*
构建消息内容
*/
func BuildMessage(msg *message_center.MessageBase) *Message {
	//recvId := nodeStore.AddressNet(msg.RecvAddr.GetAddr())
	//recvSuper := nodeStore.AddressNet(msg.RecvSuperAddr.GetAddr())
	//recvProxy := nodeStore.AddressNet(msg.RecvProxyAddr.GetAddr())
	//sendAddr := nodeStore.AddressNet(msg.SenderAddr.GetAddr())
	//sendSuper := nodeStore.AddressNet(msg.SenderSuperAddr.GetAddr())
	//sendProxy := nodeStore.AddressNet(msg.SenderProxyAddr.GetAddr())
	head := MessageHead{
		//RecvId:          &recvId,
		//RecvSuperId:     &recvSuper,
		//Sender:          &sendAddr,
		//SenderSuperId:   &sendSuper,
		Accurate: false,
		OneByOne: false,
		//RecvProxyId:     &recvProxy,
		//SenderProxyId:   &sendProxy,
		SenderMachineID: utils.Bytes2string(msg.SenderMachineID),
		RecvMachineID:   utils.Bytes2string(msg.RecvMachineID),
	}
	if msg.RecvAddr != nil {
		recvId := nodeStore.AddressNet(msg.RecvAddr.GetAddr())
		head.RecvId = &recvId
	}
	if msg.RecvSuperAddr != nil {
		recvId := nodeStore.AddressNet(msg.RecvSuperAddr.GetAddr())
		head.RecvId = &recvId
	}
	if msg.RecvProxyAddr != nil {
		recvId := nodeStore.AddressNet(msg.RecvProxyAddr.GetAddr())
		head.RecvSuperId = &recvId
	}

	if msg.SenderAddr != nil {
		recvId := nodeStore.AddressNet(msg.SenderAddr.GetAddr())
		head.Sender = &recvId
	}
	if msg.SenderSuperAddr != nil {
		recvId := nodeStore.AddressNet(msg.SenderSuperAddr.GetAddr())
		head.SenderSuperId = &recvId
	}
	if msg.SenderProxyAddr != nil {
		recvId := nodeStore.AddressNet(msg.SenderProxyAddr.GetAddr())
		head.SenderProxyId = &recvId
	}
	body := MessageBody{
		MessageId:  msg.MsgID,
		CreateTime: 0,
		ReplyTime:  0,
		Hash:       msg.SendID,
		ReplyHash:  msg.ReplyID,
		SendRand:   0,
		RecvRand:   0,
		Content:    &msg.Content,
	}
	message := Message{
		//msgid    :msg. //
		Head: &head, //
		Body: &body, //
		//DataPlus *[]byte     //body部分加密数据，消息路由时候不需要解密，临时保存
		Message: msg, //
	}
	return &message
}

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
	//if mhp.RecvVnode != nil && len(mhp.RecvVnode) > 0 {
	//	recvVnode := virtual_node.AddressNetExtend(mhp.RecvVnode)
	//	head.RecvVnode = &recvVnode
	//}
	if mhp.Sender != nil && len(mhp.Sender) > 0 {
		sender := nodeStore.AddressNet(mhp.Sender)
		head.Sender = &sender
	}
	if mhp.SenderSuperId != nil && len(mhp.SenderSuperId) > 0 {
		senderSuperId := nodeStore.AddressNet(mhp.SenderSuperId)
		head.SenderSuperId = &senderSuperId
	}
	//if mhp.SenderVnode != nil && len(mhp.SenderVnode) > 0 {
	//	senderVnode := virtual_node.AddressNetExtend(mhp.SenderVnode)
	//	head.SenderVnode = &senderVnode
	//}
	if mhp.RecvProxyId != nil && len(mhp.RecvProxyId) > 0 {
		recvProxyId := nodeStore.AddressNet(mhp.RecvProxyId)
		head.RecvProxyId = &recvProxyId
	}
	if mhp.SenderProxyId != nil && len(mhp.SenderProxyId) > 0 {
		senderProxyId := nodeStore.AddressNet(mhp.SenderProxyId)
		head.SenderProxyId = &senderProxyId
	}
	//if mhp.SearchVnodeEndId != nil && len(mhp.SearchVnodeEndId) > 0 {
	//	searchVnodeEndId := virtual_node.AddressNetExtend(mhp.SearchVnodeEndId)
	//	head.SearchVnodeEndId = &searchVnodeEndId
	//}
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
