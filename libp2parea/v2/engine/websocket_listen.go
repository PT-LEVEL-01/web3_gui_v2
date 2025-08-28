package engine

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
	"net"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

type WebsocketListen struct {
	engine      *Engine             //
	tcpAddr     *net.TCPAddr        //
	httpListen  *HttpListen         //
	closed      *atomic.Bool        //是否已经关闭
	upgrader    *websocket.Upgrader //
	contextRoot context.Context     //
	canceRoot   context.CancelFunc  //
}

func NewWebsocketListen(engine *Engine) *WebsocketListen {
	closed := new(atomic.Bool)
	closed.Store(true)
	wsListen := &WebsocketListen{engine: engine, closed: closed}
	// 我们去定义一个 Upgrader
	// 这需要一个 Read 和 Write 的缓冲大小
	wsListen.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// 我们需要检查连接的来源
		// 这将允许我们根据我们的 React 发出请求
		// 现在,我们将不检查就允许任何连接，允许跨域请求
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	wsListen.contextRoot, wsListen.canceRoot = context.WithCancel(engine.contextRoot)
	return wsListen
}

func (this *WebsocketListen) Listen(tcpAddr *net.TCPAddr, async bool) utils.ERROR {
	if tcpAddr == nil || tcpAddr.Port == 0 {
		return utils.NewErrorSuccess()
	}
	if !this.closed.CompareAndSwap(true, false) {
		return utils.NewErrorBus(ERROR_code_http_listen_runing, "")
	}
	this.tcpAddr = tcpAddr
	//
	if this.engine.tcpListen.httpListen != nil && reflect.DeepEqual(tcpAddr, this.engine.tcpListen.tcpAddr) {
		this.httpListen = this.engine.tcpListen.httpListen
	} else if this.engine.rpcHttpListen.httpListen != nil && reflect.DeepEqual(tcpAddr, this.engine.rpcHttpListen.tcpAddr) {
		this.httpListen = this.engine.rpcHttpListen.httpListen
	} else if this.engine.rpcWebsocketListen.httpListen != nil && reflect.DeepEqual(tcpAddr, this.engine.rpcWebsocketListen.tcpAddr) {
		this.httpListen = this.engine.rpcWebsocketListen.httpListen
	} else {
		httpListen := NewHttpListen(&this.engine.Log)
		ERR := httpListen.Listen(this.tcpAddr, true)
		if ERR.CheckFail() {
			return ERR
		}
		this.httpListen = httpListen
	}
	this.httpListen.SetWebSocketHandler(this.UpgradeHander)
	if !async {
		<-this.contextRoot.Done()
	}
	ERR := utils.NewErrorSuccess()
	return ERR
}

func (this *WebsocketListen) UpgradeHander(resp http.ResponseWriter, req *http.Request) {
	ws, err := this.upgrader.Upgrade(resp, req, nil)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	go this.newConnect(ws.NetConn())
}

/*
接受一个新的连接
*/
func (this *WebsocketListen) newConnect(conn net.Conn) {
	defer utils.PrintPanicStack(this.engine.Log)

	//this.engine.Log.Info().Str("Accept TCP conn", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()

	serverConn := NewServerConn(this.engine)
	var ERR utils.ERROR
	//新连接回调函数
	h := this.engine.GetAcceptBeforeEvent()
	if h != nil {
		ERR = h(conn)
		if ERR.CheckFail() {
			this.engine.Log.Error().Str("ERR", ERR.String()).Send()
			serverConn.Close()
			return
		}
	}

	serverConn.connType = CONN_TYPE_WS_server
	serverConn.conn = conn
	serverConn.Ip = conn.RemoteAddr().String()
	serverConn.Connected_time = time.Now().String()
	serverConn.outChan = make(chan *[]byte, 10000)
	serverConn.outChanCloseLock = new(sync.Mutex)
	serverConn.outChanIsClose = false
	serverConn.machineID = ulid.Make().Bytes()

	this.engine.sessionStore.addSession(serverConn)
	serverConn.run()

	//新连接回调函数
	ah := this.engine.GetAcceptAfterEvent()
	if ah != nil {
		ERR = ah(serverConn)
		if ERR.CheckFail() {
			this.engine.Log.Error().Str("ERR", ERR.String()).Send()
			serverConn.Close()
			return
		}
	}
	serverConn.allowClose <- true
	if ERR.CheckFail() {
		serverConn.Close()
	}
}

/*
连接
*/
func (this *WebsocketListen) DialAddrInfo(info *AddrInfo) (Session, utils.ERROR) {
	clientConn, ERR := NewWsClientConn(this.engine, info.Addr, info.Port)
	if ERR.CheckFail() {
		//this.engine.Log.Info().Str("传进来的地址", info.Multiaddr.String()).Send()
		return nil, ERR
	}
	clientConn.remoteMultiaddr = info.Multiaddr
	return this.connect(clientConn)
}

/*
连接
*/
func (this *WebsocketListen) DialAddr(ip string, port string) (Session, utils.ERROR) {
	clientConn, ERR := NewWsClientConn(this.engine, ip, port)
	if ERR.CheckFail() {
		return nil, ERR
	}
	return this.connect(clientConn)
}

func (this *WebsocketListen) connect(clientConn *WsClientConn) (Session, utils.ERROR) {
	var ERR utils.ERROR
	//新连接回调函数
	h := this.engine.GetDialBeforeEvent()
	if h != nil {
		ERR = h(clientConn.conn.NetConn())
		if ERR.CheckFail() {
			this.engine.Log.Error().Str("ERR", ERR.String()).Send()
			clientConn.Close()
			return nil, ERR
		}
	}
	//排除重复连接session
	this.engine.sessionStore.addSession(clientConn)
	clientConn.run()
	//新连接回调函数
	ah := this.engine.GetDialAfterEvent()
	if ah != nil {
		ERR = ah(clientConn)
		if ERR.CheckFail() {
			this.engine.Log.Error().Str("ERR", ERR.String()).Send()
			clientConn.Close()
			return clientConn, ERR
		}
	}
	clientConn.allowClose <- true
	if ERR.CheckFail() {
		this.engine.Log.Error().Str("ERR", ERR.String()).Send()
		clientConn.Close()
		return nil, ERR
	}
	return clientConn, utils.NewErrorSuccess()
}

/*
销毁，断开连接，关闭监听
@param	areaName	[]byte	区域名
*/
func (this *WebsocketListen) Destroy() {
	this.closed.Store(true)
	this.httpListen.SetWebSocketHandler(nil)
}
