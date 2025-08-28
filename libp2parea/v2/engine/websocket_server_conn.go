package engine

//
//import (
//	"context"
//	"github.com/gorilla/websocket"
//	"sync"
//	"sync/atomic"
//	"time"
//	"web3_gui/utils"
//)
//
//// 其他计算机对本机的连接
//type WsServerConn struct {
//	sessionBase                      //
//	conn             *websocket.Conn //
//	Ip               string          //
//	Connected_time   string          //
//	CloseTime        string          //
//	outChan          chan *[]byte    //发送队列
//	outChanCloseLock *sync.Mutex     //是否关闭发送队列
//	outChanIsClose   bool            //是否关闭发送队列。1=关闭
//	engine           *Engine         //
//	sendQueue        *SendQueue      //
//	isClose          *atomic.Bool    //
//	allowClose       chan bool       //允许调用关闭回调函数，保证先后顺序，close函数最后调用
//}
//
///*
//获得一个未使用的服务器连接
//*/
//func NewWsServerConn(engine *Engine) *WsServerConn {
//	contextRoot, canceRoot := context.WithCancel(context.Background())
//	//创建一个新的session
//	sessionBase := NewSessionBase(CONN_TYPE_TCP)
//	isClose := atomic.Bool{}
//	isClose.Store(false)
//	serverConn := &WsServerConn{
//		sessionBase: *sessionBase,
//		engine:      engine,
//		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot),
//		allowClose:  make(chan bool, 1),
//		isClose:     &isClose,
//	}
//	return serverConn
//}
//
//func (this *WsServerConn) GetId() []byte {
//	return this.id
//}
//
//func (this *WsServerConn) run() {
//	// this.packet.Session = this
//	go this.loopSend()
//	go this.recv()
//
//}
//
//// 接收客户端消息协程
//func (this *WsServerConn) recv() {
//	defer utils.PrintPanicStack()
//	//处理客户端主动断开连接的情况
//	var ERR utils.ERROR
//	var handler MsgHandler
//	for {
//		var packet *Packet
//		packet, ERR = ReadStream(this.conn)
//		if ERR.CheckFail() {
//			break
//		}
//		packet.Session = this
//		//Log.Debug("server conn recv: %d %s <- %s %d", packet.MsgID, this.Ip, this.conn.RemoteAddr(), len(packet.Data)+len(packet.Dataplus)+16)
//		handler = this.engine.router.GetHandler(packet.MsgID)
//		if handler == nil {
//			utils.Log.Warn().Uint64("server The message is not registered", packet.MsgID).Send()
//		} else {
//			//这里决定了消息是否异步处理
//			go this.handlerProcess(handler, packet)
//		}
//	}
//	this.Close()
//	//最后一个包接收了之后关闭chan
//	//如果有超时包需要等超时了才关闭，目前未做处理
//	this.outChanCloseLock.Lock()
//	this.outChanIsClose = true
//	close(this.outChan)
//	this.outChanCloseLock.Unlock()
//	// fmt.Println("关闭连接")
//}
//
//func (this *WsServerConn) handlerProcess(handler MsgHandler, msg *Packet) {
//	//消息处理模块报错将不会引起宕机
//	defer utils.PrintPanicStack()
//	handler(*msg)
//}
//
//func (this *WsServerConn) loopSend() {
//	defer this.conn.Close()
//	loopSend(this.sendQueue, this.conn)
//}
//
///*
//发送序列化后的数据
//*/
//func (this *WsServerConn) Send(msgID uint64, data *[]byte, timeout time.Duration) utils.ERROR {
//	bs, err := NewPacket(msgID, data).Proto()
//	if err != nil {
//		return utils.NewErrorSysSelf(err)
//	}
//	return this.sendQueue.AddAndWaitTimeout(bs, timeout)
//}
//
//// 关闭这个连接
//func (this *WsServerConn) Close() {
//	//保证只执行一次
//	if !this.isClose.CompareAndSwap(false, true) {
//		return
//	}
//	<-this.allowClose
//	closeCallback := this.engine.GetCloseCallback()
//	if closeCallback != nil {
//		closeCallback(this)
//	}
//	this.engine.sessionStore.removeSession(this)
//	err := this.conn.Close()
//	if err != nil {
//	}
//	this.sendQueue.Destroy()
//}
//
//func (this *WsServerConn) SetId(id []byte) {
//	this.engine.sessionStore.renameSession(this.GetId(), id)
//	this.id = id
//}
//
//// 获取远程ip地址和端口
//func (this *WsServerConn) GetRemoteHost() string {
//	return this.conn.RemoteAddr().String()
//}
