package message_center

import "encoding/json"

// jsoniter "github.com/json-iterator/go"

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

type MessageHeadVO struct {
	RecvId        string //接收者id
	RecvSuperId   string //接收者的超级节点id
	RecvVnode     string //接收者虚拟节点id
	Sender        string //发送者id
	SenderSuperId string //发送者超级节点id
	SenderVnode   string //发送者虚拟节点id
	Accurate      bool   //是否准确发送给一个节点，如果
	SelfVnodeId   string //查询磁力节点的时候，落到自己的哪个虚拟节点上。此字段不作序列化。
}

func (this *MessageHead) JSON() ([]byte, error) {
	vo := MessageHeadVO{
		Accurate: this.Accurate, //是否准确发送给一个节点，如果
	}
	if this.RecvId != nil {
		vo.RecvId = this.RecvId.B58String()
	}
	if this.RecvSuperId != nil {
		vo.RecvSuperId = this.RecvSuperId.B58String()
	}
	if this.RecvVnode != nil {
		vo.RecvVnode = this.RecvVnode.B58String()
	}
	if this.Sender != nil {
		vo.Sender = this.Sender.B58String()
	}
	if this.SenderSuperId != nil {
		vo.SenderSuperId = this.SenderSuperId.B58String()
	}
	if this.SenderVnode != nil {
		vo.SenderVnode = this.SenderVnode.B58String()
	}
	if this.SelfVnodeId != nil {
		vo.SelfVnodeId = this.SelfVnodeId.B58String()
	}
	return json.Marshal(vo)
}
