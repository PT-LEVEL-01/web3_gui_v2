package engine

type Packet struct {
	MsgID    uint64 //
	Size     uint64 //数据包长度，包含头部4字节
	Data     []byte
	Dataplus []byte //未加密部分数据分开
	Session  Session
}
