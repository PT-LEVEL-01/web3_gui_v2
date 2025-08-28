package engine

import (
	ma "github.com/multiformats/go-multiaddr"
	"io"
	"net"
	"time"
	"web3_gui/utils"
)

type MsgHandler func(msg *Packet)

//type MsgWaitHandler func(msg *Packet) *[]byte

type Session interface {
	GetId() []byte
	SetId(id []byte)
	Send(msgID uint64, data *[]byte, timeout time.Duration) (*Packet, utils.ERROR)
	SendWait(msgID uint64, data *[]byte, timeout time.Duration) (*[]byte, utils.ERROR)
	Reply(packet *Packet, data *[]byte, timeout time.Duration) utils.ERROR
	Close()
	GetRemoteHost() string
	GetLocalHost() string
	GetConnType() uint8
	Set(key string, value interface{})
	Get(key string) interface{}
	GetRemoteMultiaddr() ma.Multiaddr
}

/*
权限验证接口
*/
type AuthFunc func(msg *Packet) utils.ERROR

/*
关闭连接回调接口
*/
type CloseConnEvent func(ss Session) utils.ERROR

/*
有新连接，触发的回调方法
*/
type NewConnBeforeEvent func(conn io.ReadWriter) utils.ERROR

/*
有新连接，触发的回调方法
*/
type NewConnAfterEvent func(ss Session) utils.ERROR

/*
能读写的网络接口，并且能获取对方地址和端口号
*/
type IOConn interface {
	io.ReadWriter
	//返回本地地址
	LocalAddr() net.Addr
	//返回远程地址
	RemoteAddr() net.Addr
}

/*
参数自定义验证接口
*/
type CustomValidFun func(p interface{}) (any, utils.ERROR)

/*
rpc回调函数接口
*/
//type RpcHandler func(params map[string]interface{}) *PostResult

/*
rpc请求返回消息
*/
type PostResult struct {
	HTTPCode int                    `json:"hcode"` //HTTP返回状态码
	Code     uint64                 `json:"code"`  //成功或错误编号
	Msg      string                 `json:"msg"`   //错误信息
	Data     map[string]interface{} `json:"data"`  //返回的数据
}

func NewPostResult() *PostResult {
	p := PostResult{
		Code: utils.ERROR_CODE_success,
		Data: make(map[string]interface{}),
	}
	return &p
}

/*
检查这个错误类型是否是失败
*/
func (this *PostResult) ConverERR() utils.ERROR {
	return utils.NewErrorBus(this.Code, this.Msg)
}
