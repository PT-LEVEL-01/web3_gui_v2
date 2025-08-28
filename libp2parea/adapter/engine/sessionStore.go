package engine

import (
	"time"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

type Session interface {
	GetIndex() []byte
	Send(msgID uint64, data, datapuls *[]byte, timeout time.Duration) error
	Close()
	Set(name string, value interface{})
	Get(name string) interface{}
	GetName() string
	SetName(name string)
	GetRemoteHost() string
	GetMachineID() string
	GetAreaName() string
	GetSetGodTime() int64
	GetConnType() string
}

type SessionImpl struct {
	engine.Session
}

func (this *SessionImpl) GetIndex() []byte {
	return this.GetId()
}

func (s SessionImpl) Send(msgID uint64, data, datapuls *[]byte, timeout time.Duration) error {
	//TODO implement me
	panic("implement me")
}

func (this *SessionImpl) GetName() string {
	return utils.Bytes2string(this.GetId())
}

func (s SessionImpl) SetName(name string) {
	//TODO implement me
	panic("implement me")
}

func (s SessionImpl) GetMachineID() string {
	//TODO implement me
	panic("implement me")
}

func (s SessionImpl) GetAreaName() string {
	//TODO implement me
	panic("implement me")
}

func (s SessionImpl) GetSetGodTime() int64 {
	//TODO implement me
	panic("implement me")
}

func (this *SessionImpl) GetConnType() string {
	return this.GetConnType()
}

//func (this *SessionImpl) GetName() string {
//
//}
