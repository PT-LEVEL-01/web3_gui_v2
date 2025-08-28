package engine

import (
	"context"
	"github.com/oklog/ulid/v2"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

type TcpListen struct {
	engine      *Engine            //
	tcpAddr     *net.TCPAddr       //
	httpListen  *HttpListen        //
	closed      *atomic.Bool       //是否已经关闭
	contextRoot context.Context    //
	canceRoot   context.CancelFunc //
}

func NewTcpListen(engine *Engine) *TcpListen {
	closed := new(atomic.Bool)
	closed.Store(true)
	tcpListen := &TcpListen{engine: engine, closed: closed}
	tcpListen.contextRoot, tcpListen.canceRoot = context.WithCancel(engine.contextRoot)
	return tcpListen
}

func (this *TcpListen) Listen(tcpAddr *net.TCPAddr, async bool) utils.ERROR {
	if tcpAddr == nil || tcpAddr.Port == 0 {
		return utils.NewErrorSuccess()
	}
	if !this.closed.CompareAndSwap(true, false) {
		return utils.NewErrorBus(ERROR_code_tcp_listen_runing, "")
	}
	this.tcpAddr = tcpAddr

	//
	if this.engine.rpcServer.httpListen != nil && reflect.DeepEqual(tcpAddr, this.engine.rpcServer.tcpAddr) {
		this.httpListen = this.engine.rpcServer.httpListen
	} else if this.engine.wsListen.httpListen != nil && reflect.DeepEqual(tcpAddr, this.engine.wsListen.tcpAddr) {
		this.httpListen = this.engine.wsListen.httpListen
	} else {
		httpListen := NewHttpListen(&this.engine.Log)
		ERR := httpListen.Listen(this.tcpAddr, true)
		if ERR.CheckFail() {
			return ERR
		}
		this.httpListen = httpListen
	}
	this.httpListen.SetTcpHandler(this.newConnHandler)

	//监听一个地址和端口
	this.engine.Log.Info().Str("Listen TCP to an IP", tcpAddr.String()).Send()
	//go this.listener()
	if !async {
		<-this.contextRoot.Done()
	}
	return utils.NewErrorSuccess()
}

//
//func (this *TcpListen) listener() {
//	for !this.closed.Load() {
//		conn, err := this.lis.Accept()
//		if err != nil {
//			this.engine.Log.Error().Str("error", err.Error()).Send()
//			continue
//		}
//		go this.newConnect(conn)
//	}
//	this.engine.Log.Info().Str("tcp服务器关闭", "").Send()
//}

func (this *TcpListen) newConnHandler(resp http.ResponseWriter, req *http.Request) {
	//this.engine.Log.Info().Str("Server", "22222222222222").Send()
	hj, ok := resp.(http.Hijacker)
	if !ok {
		http.Error(resp, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	if tcp, ok := conn.(*net.TCPConn); ok {
		tcp.SetKeepAlive(true)
		tcp.SetKeepAlivePeriod(30 * 12 * 30 * 24 * time.Hour)
	}
	this.newConnect(conn)
}

//func (this *TcpListen) ListenByListener(listener *net.TCPListener, async bool) error {
//	if async {
//		go this.listener(listener)
//	} else {
//		this.listener(listener)
//	}
//	return nil
//}

// 创建一个新的连接
func (this *TcpListen) newConnect(conn net.Conn) {
	defer utils.PrintPanicStack(this.engine.Log)
	this.engine.Log.Info().Str("Accept TCP conn", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
	var ERR utils.ERROR
	serverConn := NewServerConn(this.engine)
	serverConn.conn = conn
	serverConn.Ip = conn.RemoteAddr().String()
	serverConn.Connected_time = time.Now().String()
	serverConn.outChan = make(chan *[]byte, 10000)
	serverConn.outChanCloseLock = new(sync.Mutex)
	serverConn.outChanIsClose = false
	serverConn.machineID = ulid.Make().Bytes()

	//能发送，不能接收
	//serverConn.loopSend()
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

	//this.engine.Log.Info().Str("Accept TCP conn", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
	//serverConn.setGodTime = setGodTime
	this.engine.sessionStore.addSession(serverConn)
	serverConn.run()
	// Log.Debug("Accept remote addr:%s", conn.RemoteAddr().String())

	//this.engine.Log.Info().Str("Accept TCP conn", conn.LocalAddr().String()+"<-"+conn.RemoteAddr().String()).Send()
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
func (this *TcpListen) DialAddrInfo(info *AddrInfo) (Session, utils.ERROR) {
	clientConn := NewClientConn(this.engine)
	clientConn.remoteMultiaddr = info.Multiaddr
	//this.engine.Log.Info().Str("远端地址", clientConn.remoteMultiaddr.String()).Send()
	port, err := strconv.Atoi(info.Port)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return this.connect(clientConn, info.Addr, uint16(port))
}

/*
连接
*/
func (this *TcpListen) DialAddr(ip string, port uint16) (Session, utils.ERROR) {
	clientConn := NewClientConn(this.engine)
	return this.connect(clientConn, ip, port)
}

/*
连接
*/
func (this *TcpListen) connect(clientConn *TcpClientConn, ip string, port uint16) (Session, utils.ERROR) {
	//clientConn := NewClientConn(this.engine)
	//clientConn.name = this.name
	err := clientConn.Connect(ip, port)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	var ERR utils.ERROR
	//新连接回调函数
	h := this.engine.GetDialBeforeEvent()
	if h != nil {
		ERR = h(clientConn.conn)
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
func (this *TcpListen) Destroy() {
	this.closed.Store(true)
	this.httpListen.SetTcpHandler(nil)
}
