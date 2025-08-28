package engine

import (
	"net"
	"strconv"

	"github.com/quic-go/quic-go"
)

// var defaultAuth Auth = new(NoneAuth)

type Auth interface {
	SendKey(conn net.Conn, session Session, name string, setGodAddr bool) (remoteName, machineID string, setGodTime int64, params interface{}, connectKey string, err error)
	RecvKey(conn net.Conn, name string) (remoteName, machineID string, setGodTime int64, params interface{}, connectKey string, err error)
	GetAreaName() (areaName string)
	SendQuicKey(conn quic.Connection, stream quic.Stream, session Session, name string, setGodAddr bool) (remoteName, machineID string, setGodTime int64, params interface{}, connectKey string, err error)
	RecvQuicKey(conn quic.Connection, stream quic.Stream, name string) (remoteName, machineID string, setGodTime int64, params interface{}, connectKey string, err error)
}

type NoneAuth struct {
	session int64
}

// 发送
// @name                 本机服务器的名称
// @setGodAddr			 是否设置对方为自己的上帝节点标识
// @return  remoteName   对方服务器的名称
func (this *NoneAuth) SendKey(conn net.Conn, session Session, name string, setGodAddr bool) (remoteName, machineID string, setGodTime int64, params interface{}, err error) {
	remoteName = name
	return
}

// 接收
// @name                 本机服务器的名称
// @return  remoteName   对方服务器的名称
func (this *NoneAuth) RecvKey(conn net.Conn, name string) (remoteName, machineID string, setGodTime int64, params interface{}, err error) {
	this.session++
	// name = strconv.ParseInt(this.session, 10, )
	// name = strconv.Itoa(this.session)
	remoteName = strconv.FormatInt(this.session, 10)
	return
}
