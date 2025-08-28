package engine

import (
	"context"
	"crypto/tls"
	"github.com/quic-go/quic-go"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

// 本机向其他服务器的连接
type ClientQuic struct {
	sessionBase
	id               []byte
	serverName       string
	ip               string
	port             uint16
	conn             quic.Connection
	stream           quic.Stream  // quic写和读取数据的流
	inPack           chan *Packet //接收队列
	outChan          chan *[]byte //发送队列
	outChanCloseLock *sync.Mutex  //是否关闭发送队列
	outChanIsClose   bool         //是否关闭发送队列。1=关闭
	engine           *Engine      //
	sendQueue        *SendQueue   //
	isClose          *atomic.Bool //
	allowClose       chan bool    //允许调用关闭回调函数，保证先后顺序，close函数最后调用
	heartbeat        chan bool    //
	getDataSign      chan bool    // 接收到数据信号，用来优化心跳超时关闭的问题
	quicConf         *quic.Config // quic客户端连接配置信息
	tlsConf          *tls.Config  // quic连接需要的tls配置信息
}

func (this *ClientQuic) GetId() []byte {
	return this.id
}

func (this *ClientQuic) run() {
	go this.loopSend()
	go this.recv()
}

func (this *ClientQuic) Connect(ip string, port uint16) error {
	// 如果设置了上帝地址，则忽略其他非上帝地址的连接

	utils.Log.Info().Str("start Connecting to", ip+":"+strconv.Itoa(int(port))).Send()
	//Log.Debug("start Connecting to:%s:%s selfAreaName:%s", ip, strconv.Itoa(int(port)), hex.EncodeToString([]byte(this.areaName)))
	this.ip = ip
	this.port = port
	this.outChan = make(chan *[]byte, SendQueueCacheNum)
	this.outChanCloseLock = new(sync.Mutex)
	this.outChanIsClose = false

	ctx, cannel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cannel()
	var err error
	this.conn, err = quic.DialAddr(ctx, ip+":"+strconv.Itoa(int(port)), this.tlsConf, this.quicConf)
	if err != nil {
		utils.Log.Error().Str("connect error", err.Error()).Send()
		//Log.Error("connect error:%s ip:%s, port:%d", err.Error(), ip, port)
		return err
	}
	stream, err := this.conn.OpenStreamSync(context.Background())
	if err != nil {
		// Log.Error("conn.OpenStreamSync:%s", err)
		return err
	}
	this.stream = stream
	// Log.Debug("Connecting to 11111:%s:%s", ip, strconv.Itoa(int(port)))
	host, portStr, err := net.SplitHostPort(this.conn.RemoteAddr().String())
	if err != nil {
		utils.Log.Error().Str("connect error", err.Error()).Send()
		//Log.Error("connect error:%s", err.Error())
		return err
	}
	portTemp, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}
	this.ip = host
	this.port = uint16(portTemp)
	// Log.Debug("Connecting to 22222:%s:%s", ip, strconv.Itoa(int(port)))
	//权限验证
	//auth, exist := this.engine.GetAuth(this.areaName)
	//if !exist || auth == nil {
	//	Log.Error("areaName auth don't exist!!!")
	//	err = errors.New("areaName auth don't exist")
	//	return
	//}
	//var setGodTime int64
	//remoteName, this.machineID, setGodTime, params, connectKey, err = auth.SendQuicKey(this.conn, this.stream, this, this.serverName, setGodAddr)
	//this.setGodTime = setGodTime
	//// Log.Debug("Connecting to 33333:%s:%s", ip, strconv.Itoa(int(port)))
	//if err != nil {
	//	Log.Error("connect error:%s", err.Error())
	//	this.conn.CloseWithError(0, "")
	//	// this.Close()
	//	if err.Error() == Error_different_netid.Error() {
	//		return
	//	}
	//	if err.Error() == Error_node_unwanted.Error() {
	//		return
	//	}
	//	return
	//}

	utils.Log.Info().Str("end Connecting to", this.conn.LocalAddr().String()).Send()
	//Log.Debug("end Connecting to:%s:%s local:%s machineId:%s", ip, strconv.Itoa(int(port)), this.conn.LocalAddr().String(), this.machineID)

	return nil
}

func (this *ClientQuic) recv() {
	defer utils.PrintPanicStack(this.engine.Log)
	quicConn := NewQuicIOConn(this.conn, this.stream)
	loopRecv(quicConn, this, this.engine.router, this.engine.authFun, this.engine.Log)

	// close(this.outChan)
	this.outChanCloseLock.Lock()
	this.outChanIsClose = true
	close(this.outChan)
	this.outChanCloseLock.Unlock()
	this.Close()
	//if this.isPowerful {
	//	go this.reConnect()
	//}
}

func (this *ClientQuic) loopSend() {
	defer this.conn.CloseWithError(0, "")
	quicConn := NewQuicIOConn(this.conn, this.stream)
	loopSend(this.sendQueue, quicConn, this.engine.Log)
}

//func (this *ClientQuic) handlerProcess(handler MsgHandler, msg *Packet) {
//	//消息处理模块报错将不会引起宕机
//	defer utils.PrintPanicStack()
//	handler(*msg)
//}

/*
发送序列化后的数据
*/
func (this *ClientQuic) Send(msgID uint64, data *[]byte, timeout time.Duration) (*Packet, utils.ERROR) {
	return send(msgID, data, timeout, this.sendQueue)
}

/*
发送后等待返回
*/
func (this *ClientQuic) SendWait(msgID uint64, data *[]byte, timeout time.Duration) (*[]byte, utils.ERROR) {
	return sendWait(msgID, data, timeout, this.sendQueue, this.engine.Log)
}

/*
回复消息
*/
func (this *ClientQuic) Reply(packet *Packet, data *[]byte, timeout time.Duration) utils.ERROR {
	return reply(packet, data, timeout, this.sendQueue)
}

// 客户端关闭时,退出recv,send
func (this *ClientQuic) Close() {
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
	err := this.conn.CloseWithError(0, "")
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
func (this *ClientQuic) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

/*
获取此会话本地ip地址和端口
*/
func (this *ClientQuic) GetLocalHost() string {
	return this.conn.LocalAddr().String()
}

// 获取节点机器Id
func (this *ClientQuic) GetMachineID() []byte {
	return this.machineID
}

func (this *ClientQuic) SetId(id []byte) {
	this.engine.sessionStore.renameSession(this.id, id)
	this.id = id
}

/*
获得一个未使用的服务器连接
*/
func getClientQuicConn(engine *Engine) *ClientQuic {
	contextRoot, canceRoot := context.WithCancel(context.Background())
	sessionBase := NewSessionBase(CONN_TYPE_QUIC_client)
	isClose := atomic.Bool{}
	isClose.Store(false)
	clientConn := &ClientQuic{
		sessionBase: *sessionBase,
		id:          engine.GetSessionId(),
		engine:      engine,
		sendQueue:   NewSendQueue(SendQueueCacheNum, contextRoot, canceRoot, &engine.Log),
		isClose:     &isClose,
		allowClose:  make(chan bool, 1),
		heartbeat:   make(chan bool, 1),
		getDataSign: make(chan bool, 1),
		quicConf: &quic.Config{
			MaxIdleTimeout:  time.Second,
			KeepAlivePeriod: 500 * time.Millisecond,
		},
		tlsConf: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-p2p-project"},
		},
	}

	return clientConn
}

type QuicIOConn struct {
	conn   quic.Connection
	stream quic.Stream // quic写和读取数据的流
}

func NewQuicIOConn(conn quic.Connection, stream quic.Stream) *QuicIOConn {
	return &QuicIOConn{conn: conn, stream: stream}
}

func (this *QuicIOConn) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *QuicIOConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *QuicIOConn) Read(p []byte) (n int, err error) {
	return this.stream.Read(p)
}

func (this *QuicIOConn) Write(p []byte) (n int, err error) {
	return this.stream.Write(p)
}
