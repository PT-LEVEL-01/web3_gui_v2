package engine

import (
	"context"
	"github.com/gorilla/websocket"
	"net/url"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

/*
websocket客户端连接
*/
type WsClientConn struct {
	sessionBase
	id          []byte
	serverName  string
	ip          string
	port        uint32
	conn        *websocket.Conn
	engine      *Engine      //
	sendQueue   *SendQueue   //
	isClose     *atomic.Bool //
	allowClose  chan bool    //允许调用关闭回调函数，保证先后顺序，close函数最后调用
	heartbeat   chan bool    //
	getDataSign chan bool    // 接收到数据信号，用来优化心跳超时关闭的问题
}

/*
获得一个未使用的服务器连接
*/
func NewWsClientConn(engine *Engine, ip string, port string) (*WsClientConn, utils.ERROR) {
	addr := ip + ":" + port
	wsDsn, err := url.Parse("ws://" + addr + HTTP_URL_websocket)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//engine.Log.Info().Str("wsdsn", wsDsn.String()).Send()
	//origin := "http://" + addr + "/"
	//url := "ws://localhost:12345/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsDsn.String(), nil)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//conn.NetConn()
	//conn, err := websocket.Dial(wsDsn.String(), "", origin)
	contextRoot, canceRoot := context.WithCancel(context.Background())
	sessionBase := NewSessionBase(CONN_TYPE_WS_client)
	isClose := new(atomic.Bool)
	isClose.Store(false)
	clientConn := &WsClientConn{
		sessionBase: *sessionBase,
		id:          engine.GetSessionId(),
		conn:        conn,
		engine:      engine,
		//outChanCloseLock: new(sync.Mutex),
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, &engine.Log),
		allowClose:  make(chan bool, 1),
		heartbeat:   make(chan bool, 1),
		getDataSign: make(chan bool, 1),
		isClose:     isClose,
	}
	return clientConn, utils.NewErrorSuccess()
}

func (this *WsClientConn) GetId() []byte {
	return this.id
}
func (this *WsClientConn) SetId(id []byte) {
	this.engine.sessionStore.renameSession(this.GetId(), id)
	this.id = id
}

func (this *WsClientConn) run() {
	go this.loopSend()
	go this.recv()
}

func (this *WsClientConn) recv() {
	defer utils.PrintPanicStack(this.engine.Log)
	loopRecv(this.conn.NetConn(), this, this.engine.router, this.engine.authFun, this.engine.Log)
	// close(this.outChan)
	//this.outChanCloseLock.Lock()
	//this.outChanIsClose = true
	//close(this.outChan)
	//this.outChanCloseLock.Unlock()
	this.Close()
}

func (this *WsClientConn) loopSend() {
	defer this.conn.Close()
	loopSend(this.sendQueue, this.conn.NetConn(), this.engine.Log)
}

//func (this *WsClientConn) handlerProcess(handler MsgHandler, msg *Packet) {
//	//消息处理模块报错将不会引起宕机
//	defer utils.PrintPanicStack()
//	//	Log.Debug("engine client 888888888888")
//	handler(*msg)
//}

/*
发送序列化后的数据
*/
func (this *WsClientConn) Send(msgID uint64, data *[]byte, timeout time.Duration) (*Packet, utils.ERROR) {
	return send(msgID, data, timeout, this.sendQueue)
}

/*
发送后等待返回
*/
func (this *WsClientConn) SendWait(msgID uint64, data *[]byte, timeout time.Duration) (*[]byte, utils.ERROR) {
	return sendWait(msgID, data, timeout, this.sendQueue, this.engine.Log)
}

/*
回复消息
*/
func (this *WsClientConn) Reply(packet *Packet, data *[]byte, timeout time.Duration) utils.ERROR {
	return reply(packet, data, timeout, this.sendQueue)
}

// 客户端关闭时,退出recv,send
func (this *WsClientConn) Close() {
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
		closeCallback(this)
	}
}

// 获取远程ip地址和端口
func (this *WsClientConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

/*
获取此会话本地ip地址和端口
*/
func (this *WsClientConn) GetLocalHost() string {
	return this.conn.LocalAddr().String()
}
