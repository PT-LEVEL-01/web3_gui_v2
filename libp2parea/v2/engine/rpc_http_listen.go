package engine

import (
	"context"
	"net"
	"net/http"
	"reflect"
	"sync/atomic"
	"web3_gui/utils"
)

type RpcHttpListen struct {
	engine      *Engine            //
	tcpAddr     *net.TCPAddr       //
	httpListen  *HttpListen        //
	closed      *atomic.Bool       //是否已经关闭
	contextRoot context.Context    //
	canceRoot   context.CancelFunc //
}

func NewRpcHttpListen(engine *Engine) *RpcHttpListen {
	closed := new(atomic.Bool)
	closed.Store(true)
	httpListen := &RpcHttpListen{engine: engine, closed: closed}
	httpListen.contextRoot, httpListen.canceRoot = context.WithCancel(engine.contextRoot)
	return httpListen
}

func (this *RpcHttpListen) Listen(tcpAddr *net.TCPAddr, async bool) utils.ERROR {
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
	this.httpListen.SetRPCHandler(this.rpcHttpHandler)
	if !async {
		<-this.contextRoot.Done()
	}
	ERR := utils.NewErrorSuccess()
	return ERR
}

/*
rpc请求内容解析
*/
func (this *RpcHttpListen) rpcHttpHandler(resp http.ResponseWriter, req *http.Request) {
	//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
	result := make(map[string]interface{}) //:= &PostResult{}
	//保证无论如何都要返回给用户
	defer func() {
		ok := utils.PrintPanicStack(this.engine.Log)
		if ok {
			result["Code"] = utils.ERROR_CODE_system_error_self
		}
		//utils.Log.Info().Interface("RPC 调用返回", result).Send()
		bs, err := json.Marshal(result)
		if err != nil {
			utils.Log.Error().Err(err).Send()
			return
		}
		resp.Header().Add("Content-Type", "application/json")
		//resp.Header().Add("Content-Length", strconv.Itoa(len(bs)))
		_, err = resp.Write(bs)
		if err != nil {
			utils.Log.Error().Err(err).Send()
			return
		}
	}()
	params := make(map[string]interface{})
	decoder := json.NewDecoder(req.Body)
	decoder.UseNumber()
	err := decoder.Decode(&params)
	if err != nil {
		//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
		ERR := utils.NewErrorSysSelf(err)
		result["Code"] = ERR.Code
		result["Msg"] = ERR.Msg
		return
	}
	//utils.Log.Info().Interface("输入参数", params).Send()
	//utils.Log.Info().Str("RPC 调用开始", "11111111111111").Send()
	//解析rpc方法名称
	methodName, ERR := this.engine.rpcServer.rpcGetMethod(params)
	if ERR.CheckFail() {
		result["Code"] = ERR.Code
		result["Msg"] = ERR.Msg
		return
	}

	ERR = this.engine.rpcServer.checkUserPassword(params, methodName)
	if ERR.CheckFail() {
		result["Code"] = ERR.Code
		result["Msg"] = ERR.Msg
		return
	}

	r, ERR := this.engine.rpcServer.proessMethodHandler(methodName, params)
	if ERR.CheckFail() {
		result["Code"] = ERR.Code
		result["Msg"] = ERR.Msg
		return
	}
	result = r
	return
}
