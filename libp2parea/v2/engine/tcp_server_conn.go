package engine

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

// 其他计算机对本机的连接
type TcpServerConn struct {
	sessionBase                   //
	conn             net.Conn     //
	Ip               string       //
	Connected_time   string       //
	CloseTime        string       //
	outChan          chan *[]byte //发送队列
	outChanCloseLock *sync.Mutex  //是否关闭发送队列
	outChanIsClose   bool         //是否关闭发送队列。1=关闭
	engine           *Engine      //
	sendQueue        *SendQueue   //
	isClose          *atomic.Bool //
	allowClose       chan bool    //允许调用关闭回调函数，保证先后顺序，close函数最后调用
}

/*
获得一个未使用的服务器连接
*/
func NewServerConn(engine *Engine) *TcpServerConn {
	contextRoot, canceRoot := context.WithCancel(context.Background())
	//创建一个新的session
	sessionBase := NewSessionBase(CONN_TYPE_TCP_server)
	sessionBase.id = engine.GetSessionId()
	isClose := atomic.Bool{}
	isClose.Store(false)
	serverConn := &TcpServerConn{
		sessionBase: *sessionBase,
		engine:      engine,
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, &engine.Log),
		allowClose:  make(chan bool, 1),
		isClose:     &isClose,
	}
	return serverConn
}

func (this *TcpServerConn) GetId() []byte {
	return this.id
}

func (this *TcpServerConn) run() {
	//this.engine.Log.Info().Str("run方法", "1111111").Send()
	go this.loopSend()
	//this.engine.Log.Info().Str("run方法", "1111111").Send()
	go this.recv()
}

// 接收客户端消息协程
func (this *TcpServerConn) recv() {
	//this.engine.Log.Info().Str("recv方法", "1111111").Send()
	defer utils.PrintPanicStack(this.engine.Log)
	loopRecv(this.conn, this, this.engine.router, this.engine.authFun, this.engine.Log)
	this.Close()
	//最后一个包接收了之后关闭chan
	//如果有超时包需要等超时了才关闭，目前未做处理
	this.outChanCloseLock.Lock()
	this.outChanIsClose = true
	close(this.outChan)
	this.outChanCloseLock.Unlock()
	// fmt.Println("关闭连接")
}

//func (this *TcpServerConn) handlerProcess(handler MsgHandler, msg *Packet) {
//	//消息处理模块报错将不会引起宕机
//	defer utils.PrintPanicStack()
//	handler(*msg)
//}

func (this *TcpServerConn) loopSend() {
	defer this.conn.Close()
	loopSend(this.sendQueue, this.conn, this.engine.Log)
}

/*
发送序列化后的数据
*/
func (this *TcpServerConn) Send(msgID uint64, data *[]byte, timeout time.Duration) (*Packet, utils.ERROR) {
	return send(msgID, data, timeout, this.sendQueue)
}

/*
回复消息
*/
func (this *TcpServerConn) Reply(packet *Packet, data *[]byte, timeout time.Duration) utils.ERROR {
	return reply(packet, data, timeout, this.sendQueue)
}

/*
发送后等待返回
*/
func (this *TcpServerConn) SendWait(msgID uint64, data *[]byte, timeout time.Duration) (*[]byte, utils.ERROR) {
	return sendWait(msgID, data, timeout, this.sendQueue, this.engine.Log)
}

// 关闭这个连接
func (this *TcpServerConn) Close() {
	//this.engine.Log.Info().Str("关闭tcp服务器连接", "11111111").Send()
	//保证只执行一次
	if !this.isClose.CompareAndSwap(false, true) {
		return
	}
	<-this.allowClose
	closeCallback := this.engine.GetCloseConnBeforeEvent()
	if closeCallback != nil {
		closeCallback(this)
	}
	this.engine.sessionStore.removeSession(this)
	err := this.conn.Close()
	if err != nil {
	}
	this.sendQueue.Destroy()
	closeCallback = this.engine.GetCloseConnAfterEvent()
	if closeCallback != nil {
		closeCallback(this)
	}
}

func (this *TcpServerConn) SetId(id []byte) {
	this.engine.sessionStore.renameSession(this.GetId(), id)
	this.id = id
}

// 获取远程ip地址和端口
func (this *TcpServerConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

/*
获取此会话本地ip地址和端口
*/
func (this *TcpServerConn) GetLocalHost() string {
	return this.conn.LocalAddr().String()
}
