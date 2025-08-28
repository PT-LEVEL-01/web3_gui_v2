package message_center

import (
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"github.com/oklog/ulid/v2"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/libp2parea/v2/protobuf/go_protobuf"
)

type MessageBase struct {
	MsgEngineID     uint64                //底层协议号
	SenderAddr      *nodeStore.AddressNet //发送者地址
	SenderSuperAddr *nodeStore.AddressNet //发送者超级节点地址
	SenderProxyAddr *nodeStore.AddressNet //发送者代理节点地址
	SenderMachineID []byte                //发送者设备机器id
	RecvAddr        *nodeStore.AddressNet //接收者地址
	RecvSuperAddr   *nodeStore.AddressNet //接收者的超级节点地址
	RecvProxyAddr   *nodeStore.AddressNet //接收者代理节点地址
	RecvMachineID   []byte                //接收者设备机器id
	MsgID           uint64                //消息id
	SendID          []byte                //发送消息id
	ReplyID         []byte                //返回消息id
	Content         []byte                //消息内容
	packet          *engine.Packet        //
	//messageCenter   *MessageCenter       //
}

func NewMessageBase(msgEngineID uint64, senderAddr, senderSuperAddr, senderProxyAddr *nodeStore.AddressNet, senderMID []byte, recvAddr,
	recvSuperAddr, recvProxyAddr *nodeStore.AddressNet, recvMID []byte, msgID uint64, sendID, replyID []byte, content *[]byte) *MessageBase {
	mb := &MessageBase{
		MsgEngineID:     msgEngineID,
		SenderAddr:      senderAddr,
		SenderSuperAddr: senderSuperAddr,
		SenderProxyAddr: senderProxyAddr,
		SenderMachineID: senderMID,
		RecvAddr:        recvAddr,
		RecvSuperAddr:   recvSuperAddr,
		RecvProxyAddr:   recvProxyAddr,
		RecvMachineID:   recvMID,
		MsgID:           msgID,
		SendID:          sendID,
		ReplyID:         replyID,
		//Content:         *content,
	}
	if content != nil {
		mb.Content = *content
	}
	return mb
}

func (this *MessageBase) GetBase() *MessageBase {
	return this
}

func (this *MessageBase) SetPacket(packet *engine.Packet) {
	this.packet = packet
}
func (this *MessageBase) GetPacket() *engine.Packet {
	return this.packet
}

func (this *MessageBase) GetProto() *go_protobuf.MessageHeadV2 {
	mh := &go_protobuf.MessageHeadV2{
		MsgEngineID: this.MsgEngineID,
		//SenderAddr:      this.SenderAddr.GetAddr(),
		//SenderSuperAddr: this.SenderSuperAddr.GetAddr(),
		//SenderProxyAddr: this.SenderProxyAddr.GetAddr(),
		SenderMachineID: this.SenderMachineID,
		//RecvAddr:        this.RecvAddr.GetAddr(),
		//RecvSuperAddr:   this.RecvSuperAddr.GetAddr(),
		//RecvProxyAddr:   this.RecvProxyAddr.GetAddr(),
		MsgID:         this.MsgID,
		RecvMachineID: this.RecvMachineID,
		SendID:        this.SendID,
		ReplyID:       this.ReplyID,
		Content:       this.Content,
	}
	if this.SenderAddr != nil {
		mh.SenderAddr = this.SenderAddr.GetAddr()
	}
	if this.SenderSuperAddr != nil {
		mh.SenderSuperAddr = this.SenderSuperAddr.GetAddr()
	}
	if this.SenderProxyAddr != nil {
		mh.SenderProxyAddr = this.SenderProxyAddr.GetAddr()
	}
	if this.RecvAddr != nil {
		mh.RecvAddr = this.RecvAddr.GetAddr()
	}
	if this.RecvSuperAddr != nil {
		mh.RecvSuperAddr = this.RecvSuperAddr.GetAddr()
	}
	if this.RecvProxyAddr != nil {
		mh.RecvProxyAddr = this.RecvProxyAddr.GetAddr()
	}
	return mh
}
func (this *MessageBase) Proto() (*[]byte, error) {
	proxyBase := this.GetProto()
	base := go_protobuf.MessageBaseV2{
		Base: proxyBase,
	}
	bs, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func ParseMessageBase(bs *[]byte) (*MessageBase, error) {
	basePro := new(go_protobuf.MessageBaseV2)
	err := proto.Unmarshal(*bs, basePro)
	if err != nil {
		return nil, err
	}
	mb := ConvertMessageBase(basePro)
	return mb, nil
}

func ConvertMessageBase(basePro *go_protobuf.MessageBaseV2) *MessageBase {
	base := MessageBase{
		MsgEngineID: basePro.Base.MsgEngineID, //底层协议号
		//SenderAddr:      nodeStore.NewAddressNet(basePro.Base.SenderAddr),      //发送者地址
		//SenderSuperAddr: nodeStore.NewAddressNet(basePro.Base.SenderSuperAddr), //发送者超级节点地址
		//SenderProxyAddr: nodeStore.NewAddressNet(basePro.Base.SenderProxyAddr), //发送者代理节点地址
		SenderMachineID: basePro.Base.SenderMachineID, //发送者设备机器id
		//RecvAddr:        nodeStore.NewAddressNet(basePro.Base.RecvAddr),        //接收者地址
		//RecvSuperAddr:   nodeStore.NewAddressNet(basePro.Base.RecvSuperAddr),   //接收者的超级节点地址
		//RecvProxyAddr:   nodeStore.NewAddressNet(basePro.Base.RecvProxyAddr),   //接收者代理节点地址
		RecvMachineID: basePro.Base.RecvMachineID, //接收者设备机器id
		MsgID:         basePro.Base.MsgID,         //消息id
		SendID:        basePro.Base.SendID,        //
		ReplyID:       basePro.Base.ReplyID,       //
		Content:       basePro.Base.Content,       //
	}
	if len(basePro.Base.SenderAddr) > 0 {
		base.SenderAddr = nodeStore.NewAddressNet(basePro.Base.SenderAddr)
	}
	if len(basePro.Base.SenderSuperAddr) > 0 {
		base.SenderSuperAddr = nodeStore.NewAddressNet(basePro.Base.SenderSuperAddr)
	}
	if len(basePro.Base.SenderProxyAddr) > 0 {
		base.SenderProxyAddr = nodeStore.NewAddressNet(basePro.Base.SenderProxyAddr)
	}
	if len(basePro.Base.RecvAddr) > 0 {
		base.RecvAddr = nodeStore.NewAddressNet(basePro.Base.RecvAddr)
	}
	if len(basePro.Base.RecvSuperAddr) > 0 {
		base.RecvSuperAddr = nodeStore.NewAddressNet(basePro.Base.RecvSuperAddr)
	}
	if len(basePro.Base.RecvProxyAddr) > 0 {
		base.RecvProxyAddr = nodeStore.NewAddressNet(basePro.Base.RecvProxyAddr)
	}
	return &base
}

type MessageVO struct {
	MsgEngineID     uint64 //底层协议号
	SenderAddr      string //发送者地址
	SenderSuperAddr string //发送者超级节点地址
	SenderProxyAddr string //发送者代理节点地址
	SenderMachineID string //发送者设备机器id
	RecvAddr        string //接收者地址
	RecvSuperAddr   string //接收者的超级节点地址
	RecvProxyAddr   string //接收者代理节点地址
	RecvMachineID   string //接收者设备机器id
	MsgID           uint64 //消息id
	SendID          string //发送消息id
	ReplyID         string //返回消息id
	Content         string //消息内容
}

func (this *MessageBase) ConvertVO() *MessageVO {
	mvo := MessageVO{
		MsgEngineID: this.MsgEngineID, //底层协议号
		//SenderAddr:      this.SenderAddr.B58String(),              //发送者地址
		//SenderSuperAddr: this.SenderSuperAddr.B58String(),         //发送者超级节点地址
		//SenderProxyAddr: this.SenderProxyAddr.B58String(),         //发送者代理节点地址
		SenderMachineID: hex.EncodeToString(this.SenderMachineID), //发送者设备机器id
		//RecvAddr:        this.RecvAddr.B58String(),                //接收者地址
		//RecvSuperAddr:   this.RecvSuperAddr.B58String(),           //接收者的超级节点地址
		//RecvProxyAddr:   this.RecvProxyAddr.B58String(),           //接收者代理节点地址
		RecvMachineID: hex.EncodeToString(this.RecvMachineID), //接收者设备机器id
		MsgID:         this.MsgID,                             //消息id
		SendID:        hex.EncodeToString(this.SendID),        //发送消息id
		ReplyID:       hex.EncodeToString(this.ReplyID),       //返回消息id
		Content:       hex.EncodeToString(this.Content),       //消息内容
	}
	if this.SenderAddr != nil {
		mvo.SenderAddr = this.SenderAddr.B58String()
	}
	if this.SenderSuperAddr != nil {
		mvo.SenderSuperAddr = this.SenderSuperAddr.B58String()
	}
	if this.SenderProxyAddr != nil {
		mvo.SenderProxyAddr = this.SenderProxyAddr.B58String()
	}
	if this.RecvAddr != nil {
		mvo.RecvAddr = this.RecvAddr.B58String()
	}
	if this.RecvSuperAddr != nil {
		mvo.RecvSuperAddr = this.RecvSuperAddr.B58String()
	}
	if this.RecvProxyAddr != nil {
		mvo.RecvProxyAddr = this.RecvProxyAddr.B58String()
	}
	return &mvo
}

func NewMessageNeighbor(msgID uint64, addrSelf *nodeStore.AddressNet, content *[]byte) *MessageBase {
	message := NewMessageBase(0, addrSelf, nil,
		nil, nil, nil, nil, nil, nil, msgID,
		nil, nil, content)
	return message
}

func NewMessageNeighborWait(msgID uint64, addrSelf *nodeStore.AddressNet, content *[]byte) *MessageBase {
	message := NewMessageBase(0, addrSelf, nil,
		nil, nil, nil, nil, nil, nil, msgID,
		ulid.Make().Bytes(), nil, content)
	return message
}

func NewMessageNeighborReply(base *MessageBase, addrSelf *nodeStore.AddressNet, content *[]byte) *MessageBase {
	message := NewMessageBase(0, addrSelf, nil,
		nil, nil, nil, nil, nil, nil, 0,
		base.SendID, ulid.Make().Bytes(), content)
	return message
}

func NewMessageForward(msgID_engine, msgID_p2p uint64, addrSelf *nodeStore.AddressNet, machineIdSelf []byte,
	recvId *nodeStore.AddressNet, content *[]byte) *MessageBase {
	message := NewMessageBase(msgID_engine, addrSelf, nil, nil, machineIdSelf, recvId,
		nil, nil, nil, msgID_p2p, ulid.Make().Bytes(), nil, content)
	return message
}

func NewMessageForwardReply(msgEngineId uint64, base *MessageBase, addrSelf *nodeStore.AddressNet, machineIdSelf []byte,
	content *[]byte) *MessageBase {
	message := NewMessageBase(msgEngineId, addrSelf, nil, nil, machineIdSelf, base.SenderAddr,
		base.SenderSuperAddr, base.SenderProxyAddr, base.SenderMachineID, 0, base.SendID, ulid.Make().Bytes(), content)
	return message
}
