package engine

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

// 本机向其他服务器的连接
type TcpClientConn struct {
	sessionBase
	id               []byte
	serverName       string
	conn             net.Conn
	inPack           chan *Packet //接收队列
	outChan          chan *[]byte //发送队列
	outChanCloseLock *sync.Mutex  //是否关闭发送队列
	outChanIsClose   bool         //是否关闭发送队列。1=关闭
	isPowerful       bool         //是否是强连接，强连接有短线重连功能
	engine           *Engine      //
	sendQueue        *SendQueue   //
	isClose          *atomic.Bool //
	allowClose       chan bool    //允许调用关闭回调函数，保证先后顺序，close函数最后调用
	heartbeat        chan bool    //
	getDataSign      chan bool    // 接收到数据信号，用来优化心跳超时关闭的问题
}

/*
获得一个未使用的服务器连接
*/
func NewClientConn(engine *Engine) *TcpClientConn {
	contextRoot, canceRoot := context.WithCancel(engine.contextRoot)
	sessionBase := NewSessionBase(CONN_TYPE_TCP_client)
	isClose := new(atomic.Bool)
	isClose.Store(false)
	clientConn := &TcpClientConn{
		sessionBase: *sessionBase,
		id:          engine.GetSessionId(),
		engine:      engine,
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, &engine.Log),
		allowClose:  make(chan bool, 1),
		heartbeat:   make(chan bool, 1),
		getDataSign: make(chan bool, 1),
		isClose:     isClose,
	}
	return clientConn
}

func (this *TcpClientConn) GetId() []byte {
	return this.id
}

func (this *TcpClientConn) run() {
	go this.loopSend()
	go this.recv()
}

func (this *TcpClientConn) Connect(ip string, port uint16) error {
	// 如果设置了上帝地址，则忽略其他非上帝地址的连接
	//this.engine.Log.Info().Str("start Connecting to", ip+":"+strconv.Itoa(int(port))).Send()
	//Log.Debug("start Connecting to:%s:%s selfAreaName:%s", ip, strconv.Itoa(int(port)), hex.EncodeToString([]byte(this.areaName)))
	this.outChan = make(chan *[]byte, SendQueueCacheNum)
	this.outChanCloseLock = new(sync.Mutex)
	this.outChanIsClose = false
	var err error
	this.conn, err = net.DialTimeout("tcp", ip+":"+strconv.Itoa(int(port)), time.Second*3)
	if err != nil {
		this.engine.Log.Error().Str("connect error", err.Error()).Send()
		//Log.Error("connect error:%s", err.Error())
		return err
	}
	// 通过TCP连接发送HTTP CONNECT请求
	_, err = this.conn.Write([]byte(fmt.Sprintf("CONNECT / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\nContent-Length: 5\r\n\r\nhello",
		ip+":"+strconv.Itoa(int(port)))))
	this.engine.Log.Info().Str("Connecting to", this.conn.LocalAddr().String()+"->"+this.conn.RemoteAddr().String()).Send()
	return nil
}

func (this *TcpClientConn) recv() {
	defer utils.PrintPanicStack(this.engine.Log)
	loopRecv(this.conn, this, this.engine.router, this.engine.authFun, this.engine.Log)
	// close(this.outChan)
	this.outChanCloseLock.Lock()
	this.outChanIsClose = true
	close(this.outChan)
	this.outChanCloseLock.Unlock()
	this.Close()
}

func (this *TcpClientConn) loopSend() {
	defer this.conn.Close()
	loopSend(this.sendQueue, this.conn, this.engine.Log)
}

//func (this *TcpClientConn) handlerProcess(handler MsgHandler, msg *Packet) {
//	//消息处理模块报错将不会引起宕机
//	defer utils.PrintPanicStack()
//	//	Log.Debug("engine client 888888888888")
//	handler(*msg)
//}

/*
发送序列化后的数据
*/
func (this *TcpClientConn) Send(msgID uint64, data *[]byte, timeout time.Duration) (*Packet, utils.ERROR) {
	return send(msgID, data, timeout, this.sendQueue)
}

/*
发送后等待返回
*/
func (this *TcpClientConn) SendWait(msgID uint64, data *[]byte, timeout time.Duration) (*[]byte, utils.ERROR) {
	return sendWait(msgID, data, timeout, this.sendQueue, this.engine.Log)
}

/*
回复消息
*/
func (this *TcpClientConn) Reply(packet *Packet, data *[]byte, timeout time.Duration) utils.ERROR {
	return reply(packet, data, timeout, this.sendQueue)
}

// 客户端关闭时,退出recv,send
func (this *TcpClientConn) Close() {
	//this.engine.Log.Info().Str("关闭tcp客户端连接", "11111111").Send()
	//保证只执行一次
	if !this.isClose.CompareAndSwap(false, true) {
		return
	}
	<-this.allowClose

	closeCallback := this.engine.GetCloseConnBeforeEvent()
	if closeCallback != nil {
		closeCallback(this)
	}
	// Log.Info("Close session ClientConn")
	this.engine.sessionStore.removeSession(this)
	// Log.Info("Close session ClientConn")
	err := this.conn.Close()
	if err != nil {
	}
	// Log.Info("Close session ClientConn")
	this.sendQueue.Destroy()
	closeCallback = this.engine.GetCloseConnAfterEvent()
	if closeCallback != nil {
		//this.engine.Log.Info().Interface("回调方法", closeCallback).Interface("回调方法", this).Send()
		closeCallback(this)
	}
}

/*
获取远程ip地址和端口
*/
func (this *TcpClientConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

/*
获取此会话本地ip地址和端口
*/
func (this *TcpClientConn) GetLocalHost() string {
	return this.conn.LocalAddr().String()
}

func (this *TcpClientConn) SetId(id []byte) {
	this.engine.sessionStore.renameSession(this.GetId(), id)
	this.id = id
}
