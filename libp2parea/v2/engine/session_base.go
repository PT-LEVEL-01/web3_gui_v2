package engine

import (
	ma "github.com/multiformats/go-multiaddr"
	"sync"
)

type sessionBase struct {
	id              []byte       //
	machineID       []byte       // 机器Id
	connType        uint8        // 连接类型
	attribute       *sync.Map    //
	remoteMultiaddr ma.Multiaddr //
}

func NewSessionBase(connType uint8) *sessionBase {
	return &sessionBase{
		connType:  connType,
		attribute: new(sync.Map),
	}
}

func (this *sessionBase) GetId() []byte {
	return this.id
}

func (this *sessionBase) SetId(id []byte) {
	this.id = id
}

func (this sessionBase) GetConnType() uint8 {
	return this.connType
}

func (this *sessionBase) Close() {}

func (this *sessionBase) Set(key string, value interface{}) {
	this.attribute.Store(key, value)
}

func (this *sessionBase) Get(key string) interface{} {
	if value, ok := this.attribute.Load(key); ok {
		return value
	}
	return nil
}

func (this *sessionBase) GetRemoteMultiaddr() ma.Multiaddr {
	return this.remoteMultiaddr
}
