package engine

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
	"unsafe"
	"web3_gui/utils"
)

type RpcWebsocketListen struct {
	engine      *Engine             //
	tcpAddr     *net.TCPAddr        //
	httpListen  *HttpListen         //
	closed      *atomic.Bool        //是否已经关闭
	upgrader    *websocket.Upgrader //
	contextRoot context.Context     //
	canceRoot   context.CancelFunc  //
}

func NewRpcWebsocketListen(engine *Engine) *RpcWebsocketListen {
	closed := new(atomic.Bool)
	closed.Store(true)
	wsListen := &RpcWebsocketListen{engine: engine, closed: closed}
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

func (this *RpcWebsocketListen) Listen(tcpAddr *net.TCPAddr, async bool) utils.ERROR {
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
	} else if this.engine.wsListen.httpListen != nil && reflect.DeepEqual(tcpAddr, this.engine.wsListen.tcpAddr) {
		this.httpListen = this.engine.wsListen.httpListen
	} else if this.engine.rpcHttpListen.httpListen != nil && reflect.DeepEqual(tcpAddr, this.engine.rpcHttpListen.tcpAddr) {
		this.httpListen = this.engine.rpcHttpListen.httpListen
	} else {
		httpListen := NewHttpListen(&this.engine.Log)
		ERR := httpListen.Listen(this.tcpAddr, true)
		if ERR.CheckFail() {
			return ERR
		}
		this.httpListen = httpListen
	}
	this.httpListen.SetRPCWebsocketHandler(this.rpcWebsocketUpgradeHander)
	if !async {
		<-this.contextRoot.Done()
	}
	ERR := utils.NewErrorSuccess()
	return ERR
}

func (this *RpcWebsocketListen) rpcWebsocketUpgradeHander(resp http.ResponseWriter, req *http.Request) {
	ws, err := this.upgrader.Upgrade(resp, req, nil)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return
	}
	go this.newRpcConnect(ws)
}

/*
接受一个新的连接
*/
func (this *RpcWebsocketListen) newRpcConnect(conn *websocket.Conn) {
	defer utils.PrintPanicStack(this.engine.Log)
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		this.engine.Log.Info().Int("rpc websocket 参数 消息类型", messageType).Str("rpc websocket 参数", string(p)).Send()
		/*
			// 字节转字符串
			fmt.Println(string(p))
			fstr := *(*string)(unsafe.Pointer(&p))
			fmt.Println(fstr)
		*/

		/*
			// 字符串转字节
			fmt.Println(*(*[]byte)(unsafe.Pointer(&fstr)))
		*/

		log.Printf("received: %s", p)

		//lower := strings.ToLower(fstr)
		//字符串转大写
		fstr := *(*string)(unsafe.Pointer(&p))
		upperStr := strings.ToUpper(fstr)

		err = conn.WriteMessage(messageType, *(*[]byte)(unsafe.Pointer(&upperStr)))
		if err != nil {
			log.Println(err)
			return
		}
	}

}

/*
销毁，断开连接，关闭监听
@param	areaName	[]byte	区域名
*/
func (this *RpcWebsocketListen) Destroy() {
	this.closed.Store(true)
	this.httpListen.SetRPCWebsocketHandler(nil)
}
