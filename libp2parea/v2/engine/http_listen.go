package engine

import (
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"web3_gui/libp2parea/v2/engine/html"
	"web3_gui/utils"
)

type HttpListen struct {
	tcpAddr             *net.TCPAddr     //
	lis                 *net.TCPListener //
	tcpHandler          http.HandlerFunc //
	websocketHandler    http.HandlerFunc //
	rpcHttpHandler      http.HandlerFunc //
	rpcWebsocketHandler http.HandlerFunc //
	Log                 **zerolog.Logger //日志
	once                *sync.Once       //
}

func NewHttpListen(log **zerolog.Logger) *HttpListen {
	closed := new(atomic.Bool)
	closed.Store(true)
	httpListen := &HttpListen{Log: log, once: new(sync.Once)}
	return httpListen
}

func (this *HttpListen) Listen(tcpAddr *net.TCPAddr, async bool) utils.ERROR {
	if tcpAddr == nil || tcpAddr.Port == 0 {
		return utils.NewErrorSuccess()
	}
	this.tcpAddr = tcpAddr
	var err error
	//utils.Log.Info().Str("Listen HTTP to an IP", tcpAddr.String()).Send()
	this.lis, err = net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}
	go this.runHttpServer()
	ERR := utils.NewErrorSuccess()
	return ERR
}

func (this *HttpListen) runHttpServer() {
	err := http.Serve(this.lis, this)
	if err != nil {
		utils.Log.Error().Err(err).Send()
	}
}

func (this *HttpListen) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	//tcp连接
	if req.Method == HTTP_Method_connect {
		//utils.Log.Info().Interface("使用tcpHandler", this.tcpHandler).Send()
		this.tcpHandler(resp, req)
		return
	}
	switch req.URL.String() {
	case HTTP_URL_websocket: //websocket连接
		if this.websocketHandler == nil {
			return
		}
		this.websocketHandler(resp, req)
	case HTTP_URL_rpc_websocket: //websocket rpc连接
		if this.rpcWebsocketHandler == nil {
			return
		}
		this.rpcWebsocketHandler(resp, req)
	case HTTP_URL_rpc_json: //json rpc请求
		if this.rpcHttpHandler == nil {
			return
		}
		this.rpcHttpHandler(resp, req)
	default:
		//不是rpc路径的请求
		this.defaultHanderFUN(resp, req)
	}
}

/*
这是一个当http请求路径输入错误的情况下，默认的输出。
*/
func (this *HttpListen) defaultHanderFUN(resp http.ResponseWriter, req *http.Request) {
	//this.useVue()
	var resultTpl string
	switch req.URL.String() {
	case HTTP_URL_Page_rpclistjson:
		resultTpl = html.HTML_rpc_list_json_page
	case HTTP_URL_Page_errors_desc:
		resultTpl = html.HTML_rpc_errors_desc_page
	case HTTP_URL_Page_websocket_client:
		resultTpl = html.HTML_rpc_websocket_page
	//case HTTP_URL_Static_JS_vue_v3_5_13:
	//	resultTpl = html.JS_vue_v3_5_13
	//case HTTP_URL_Static_JS_element_plus_v2_9_8:
	//	resultTpl = html.JS_element_plus_v2_9_8
	//case HTTP_URL_Static_CSS_element_plus_index:
	//	resultTpl = html.CSS_element_plus_index
	default:
		resultTpl = html.HTML_index
	}
	//未找到的地址，都显示首页
	//utils.Log.Info().Str("url", req.URL.String()).Send()
	_, err := resp.Write([]byte(resultTpl))
	//tpl, err := template.New(req.URL.String()).Parse(resultTpl)
	//if err != nil {
	//	(*this.Log).Error().Err(err).Send()
	//	return
	//}
	//err = tpl.Execute(resp, RpcUserParams)
	if err != nil {
		(*this.Log).Error().Err(err).Send()
		return
	}
}

//func (this *HttpListen) useVue() {
//	this.once.Do(func() {
//		RpcUserParams[TPL_KEY_vuejs] = html.JS_vue_global_min
//	})
//}

func (this *HttpListen) SetWebSocketHandler(websocketHandler http.HandlerFunc) {
	this.websocketHandler = websocketHandler
	this.CheckClose()
}

func (this *HttpListen) SetRPCHandler(rpcHandler http.HandlerFunc) {
	this.rpcHttpHandler = rpcHandler
	this.CheckClose()
}

func (this *HttpListen) SetRPCWebsocketHandler(rpcWebsocketHandler http.HandlerFunc) {
	this.rpcWebsocketHandler = rpcWebsocketHandler
	this.CheckClose()
}

func (this *HttpListen) SetTcpHandler(tcpHandler http.HandlerFunc) {
	//utils.Log.Info().Interface("设置tcpHandler", tcpHandler).Send()
	this.tcpHandler = tcpHandler
	this.CheckClose()
}

func (this *HttpListen) CheckClose() {
	if this.websocketHandler == nil && this.rpcHttpHandler == nil && this.tcpHandler == nil && this.rpcWebsocketHandler == nil {
		this.lis.Close()
	}
}
