package engine

import (
	"context"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog"
	"math/big"
	"net"
	"strconv"
	"sync"
	"web3_gui/utils"
)

/*
不能设置handler运行总量，不能容忍业务阻塞，所以用户自己去处理
*/
type Engine struct {
	name               string              //本机名称
	sessionIdGenLock   *sync.Mutex         //
	sessionIdGen       *big.Int            //连接id自增长生成器
	router             *Router             //
	authFun            AuthFunc            //权限验证，每个接收的包都要验权
	rpcServer          *RPCServer          //
	sessionStore       *sessionStore       //
	tcpListen          *TcpListen          //
	wsListen           *WebsocketListen    //
	rpcHttpListen      *RpcHttpListen      //
	rpcWebsocketListen *RpcWebsocketListen //
	quicListen         *QuicListen         //
	dialBeforeEvent    NewConnBeforeEvent  //主动连接之前事件
	dialAfterEvent     NewConnAfterEvent   //主动连接之后事件
	acceptBeforeEvent  NewConnBeforeEvent  //被动连接之前事件
	acceptAfterEvent   NewConnAfterEvent   //被动连接之后事件
	closeBeforeEvent   CloseConnEvent      //连接被关闭之前事件
	closeAfterEvent    CloseConnEvent      //连接被关闭之后事件
	Log                *zerolog.Logger     //日志
	contextRoot        context.Context     //
	canceRoot          context.CancelFunc  //
}

// @name   本服务器名称
func NewEngine(name string) *Engine {
	engine := new(Engine)
	engine.contextRoot, engine.canceRoot = context.WithCancel(context.Background())
	engine.name = name
	engine.rpcServer = NewRPCServer(engine)
	engine.tcpListen = NewTcpListen(engine)
	engine.wsListen = NewWebsocketListen(engine)
	engine.rpcHttpListen = NewRpcHttpListen(engine)
	engine.rpcWebsocketListen = NewRpcWebsocketListen(engine)
	engine.quicListen = NewQuicListen(engine)
	engine.sessionStore = NewSessionStore()
	engine.router = NewRouter()
	engine.sessionIdGenLock = new(sync.Mutex)
	engine.sessionIdGen = big.NewInt(0)
	engine.Log = utils.Log
	return engine
}

/*
注册一个普通消息
*/
func (this *Engine) RegisterMsg(msgId uint64, handler MsgHandler) {
	//打印注册消息id
	//this.Log.Info().Uint64("register message id", msgId).Send()
	this.router.AddRouter(msgId, handler)
}

/*
合并监听到同一个端口
@port     uint16    监听端口
@async    bool      是否异步执行
*/
func (this *Engine) ListenOnePort(port uint16, async bool) utils.ERROR {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(int(port)))
	if err != nil {
		this.Log.Error().Str("err", err.Error()).Send()
		return utils.NewErrorSysSelf(err)
	}
	udpAddr, err := net.ResolveUDPAddr("udp4", tcpAddr.String())
	if err != nil {
		this.Log.Error().Str("err", err.Error()).Send()
		return utils.NewErrorSysSelf(err)
	}
	this.Log.Info().Str("Listen TCP、HTTP、WS、QUIC to an addr", tcpAddr.String()).Send()
	conf := ListenConfig{
		TcpAddr:     tcpAddr, //tcp服务器监听地址
		WsAddr:      tcpAddr, //websocket监听地址
		RpcHttpAddr: tcpAddr, //web服务器监听地址
		RpcWsAddr:   tcpAddr, //web服务器监听地址
		QuicAddr:    udpAddr, //quic协议监听地址
	}
	return this.Listen(&conf, async)
}

/*
监听
*/
func (this *Engine) Listen(config *ListenConfig, async bool) utils.ERROR {
	//当tcp、http、websocket 协议都使用同一个端口的情况
	ERR := this.tcpListen.Listen(config.TcpAddr, true)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = this.wsListen.Listen(config.WsAddr, true)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = this.rpcHttpListen.Listen(config.RpcHttpAddr, true)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = this.rpcWebsocketListen.Listen(config.RpcWsAddr, true)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = this.quicListen.Listen(config.QuicAddr, true)
	if ERR.CheckFail() {
		return ERR
	}
	if !async {
		<-this.contextRoot.Done()
	}
	return utils.NewErrorSuccess()
}

/*
监听一个tcp端口
*/
func (this *Engine) ListenTCP(ip string, port uint16, async bool) utils.ERROR {
	listenHost := ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listenHost)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}
	return this.tcpListen.Listen(tcpAddr, async)
}

/*
 * 监听quic信息
 */
func (this *Engine) ListenQuic(ip string, port uint16, async bool) utils.ERROR {
	listenHost := ip + ":" + strconv.Itoa(int(port))
	udpAddr, err := net.ResolveUDPAddr("udp4", listenHost)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		this.Log.Error().Str("ERR", ERR.String()).Send()
		return ERR
	}
	return this.quicListen.Listen(udpAddr, async)
}

/*
添加一个连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
@powerful      是否是强连接
@return  name  对方的名称
*/
func (this *Engine) Dial(addr string) (Session, utils.ERROR) {
	a, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	addrInfo, ERR := CheckAddr(a)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//this.Log.Info().Interface("addrInfo", addrInfo).Send()
	var session Session
	switch addrInfo.NetType {
	case "tcp4":
		switch addrInfo.Proto {
		case "ws":
			session, ERR = this.wsListen.DialAddrInfo(addrInfo)
		case "":
			session, ERR = this.tcpListen.DialAddrInfo(addrInfo)
		default:
			this.Log.Error().Str("不支持的协议", addrInfo.NetType).Send()
			return nil, utils.NewErrorBus(ERROR_code_unsupported_protocols, addrInfo.NetType)
		}
	case "tcp6":
		switch addrInfo.Proto {
		case "ws":
			session, ERR = this.wsListen.DialAddrInfo(addrInfo)
		case "":
			session, ERR = this.tcpListen.DialAddrInfo(addrInfo)
		default:
			this.Log.Error().Str("不支持的协议", addrInfo.NetType).Send()
			return nil, utils.NewErrorBus(ERROR_code_unsupported_protocols, addrInfo.NetType)
		}
	case "quic":
		session, ERR = this.quicListen.DialAddrInfo(addrInfo)
	default:
		this.Log.Error().Str("不支持的协议", addrInfo.NetType).Send()
		return nil, utils.NewErrorBus(ERROR_code_unsupported_protocols, addrInfo.NetType)
	}
	return session, ERR
}

/*
添加一个连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
@powerful      是否是强连接
@return  name  对方的名称
*/
func (this *Engine) DialTcpAddr(ip string, port uint16) (Session, utils.ERROR) {
	return this.tcpListen.DialAddr(ip, port)
}

/*
添加一个连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
@powerful      是否是强连接
@return  name  对方的名称
*/
func (this *Engine) DialWebsocketAddr(ip string, port uint16) (Session, utils.ERROR) {
	return this.wsListen.DialAddr(ip, strconv.Itoa(int(port)))
}

/*
添加一个Quic连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
@powerful      是否是强连接
@return  name  对方的名称
*/
func (this *Engine) DialQuicAddr(ip string, port uint16) (Session, utils.ERROR) {
	return this.quicListen.DialAddr(ip, port)
}

func (this *Engine) GetSessionId() []byte {
	this.sessionIdGenLock.Lock()
	this.sessionIdGen = new(big.Int).Add(this.sessionIdGen, big.NewInt(1))
	sessionId := this.sessionIdGen.Bytes()
	this.sessionIdGenLock.Unlock()
	return sessionId
}

// 获得session
func (this *Engine) GetSession(id []byte) Session {
	return this.sessionStore.getSessionById(id)
}

// 获得某个地址所有session
func (this *Engine) GetSessionAll() []Session {
	return this.sessionStore.getSessionAll()
}

/*
 * 删除sessionStore中特定session
 */
func (this *Engine) RemoveSession(ss Session) {
	this.sessionStore.removeSession(ss)
}

/*
设置权限方法
*/
func (this *Engine) SetAuthFun(auth AuthFunc) {
	this.authFun = auth
}

/*
设置主动连接之前事件
*/
func (this *Engine) SetDialBeforeEvent(event NewConnBeforeEvent) {
	this.dialBeforeEvent = event
}

/*
获取主动连接之前事件
*/
func (this *Engine) GetDialBeforeEvent() NewConnBeforeEvent {
	return this.dialBeforeEvent
}

/*
设置主动连接之后事件
*/
func (this *Engine) SetDialAfterEvent(event NewConnAfterEvent) {
	this.dialAfterEvent = event
}

/*
获取主动连接之后事件
*/
func (this *Engine) GetDialAfterEvent() NewConnAfterEvent {
	return this.dialAfterEvent
}

/*
设置被动连接之前事件
*/
func (this *Engine) SetAcceptBeforeEvent(event NewConnBeforeEvent) {
	this.acceptBeforeEvent = event
}

/*
获取被动连接之前事件
*/
func (this *Engine) GetAcceptBeforeEvent() NewConnBeforeEvent {
	return this.acceptBeforeEvent
}

/*
设置被动连接之后事件
*/
func (this *Engine) SetAcceptAfterEvent(event NewConnAfterEvent) {
	this.acceptAfterEvent = event
}

/*
获取新连接之后事件
*/
func (this *Engine) GetAcceptAfterEvent() NewConnAfterEvent {
	return this.acceptAfterEvent
}

/*
设置关闭连接之前事件
*/
func (this *Engine) SetCloseConnBeforeEvent(event CloseConnEvent) {
	this.closeBeforeEvent = event
}

/*
获取关闭连接之前事件
*/
func (this *Engine) GetCloseConnBeforeEvent() CloseConnEvent {
	return this.closeBeforeEvent
}

/*
设置关闭连接之后事件
*/
func (this *Engine) SetCloseConnAfterEvent(event CloseConnEvent) {
	this.closeAfterEvent = event
}

/*
获取关闭连接之后事件
*/
func (this *Engine) GetCloseConnAfterEvent() CloseConnEvent {
	return this.closeAfterEvent
}

/*
获取关闭连接之后事件
*/
func (this *Engine) RegisterRPC(sortNumber int, rpcName string, handler any, desc string, pvs ...ParamValid) utils.ERROR {
	return this.rpcServer.RegisterRpcHandler(sortNumber, rpcName, handler, desc, pvs...)
}

/*
添加用户名和密码
*/
func (this *Engine) AddRpcUser(user, password string) utils.ERROR {
	return this.rpcServer.AddRpcUser(user, password)
}

/*
设置用户名和密码
*/
func (this *Engine) UpdateRpcUser(user, password string) utils.ERROR {
	return this.rpcServer.UpdateRpcUser(user, password)
}

/*
删除用户名和密码
*/
func (this *Engine) DelRpcUser(user string) utils.ERROR {
	return this.rpcServer.DelRpcUser(user)
}

/*
执行方法
*/
func (this *Engine) RunRpcMethod(method string, params map[string]interface{}) (map[string]interface{}, utils.ERROR) {
	return this.rpcServer.proessMethodHandler(method, params)
}

/*
设置日志
*/
func (this *Engine) SetLog(log *zerolog.Logger) {
	this.Log = log
}

/*
 * 销毁，断开连接，关闭监听
 * @param	areaName	[]byte	区域名
 */
func (this *Engine) Destroy() {
	//this.Log.Info().Str("销毁1111", "").Send()
	this.tcpListen.Destroy()
	this.rpcServer.Destroy()
	this.wsListen.Destroy()
	this.quicListen.Destroy()
	//this.Log.Info().Str("销毁1111", "").Send()
	// 根据区域名，获取所有的session，依次进行关闭操作
	ss := this.sessionStore.getSessionAll()
	//this.Log.Info().Str("销毁1111", "").Send()
	for _, session := range ss {
		session.Close()
	}
	//this.Log.Info().Str("销毁1111", "").Send()
}
