package engine

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/quic-go/quic-go"
	"math/big"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/utils"
)

type QuicListen struct {
	engine      *Engine            //
	udpAddr     *net.UDPAddr       //
	lisQuic     *quic.Listener     //Quic的监听listener
	closed      *atomic.Bool       //是否已经关闭
	contextRoot context.Context    //
	canceRoot   context.CancelFunc //
}

func NewQuicListen(engine *Engine) *QuicListen {
	closed := new(atomic.Bool)
	closed.Store(true)
	quicListen := &QuicListen{engine: engine, closed: closed}
	quicListen.contextRoot, quicListen.canceRoot = context.WithCancel(engine.contextRoot)
	return quicListen
}

func (this *QuicListen) Listen(udpAddr *net.UDPAddr, async bool) utils.ERROR {
	if udpAddr == nil || udpAddr.Port == 0 {
		return utils.NewErrorSuccess()
	}
	if !this.closed.CompareAndSwap(true, false) {
		return utils.NewErrorBus(ERROR_code_quic_listen_runing, "")
	}
	this.udpAddr = udpAddr
	conf := quic.Config{
		MaxIdleTimeout:  time.Second,
		KeepAlivePeriod: 500 * time.Millisecond,
	}
	lisQuic, err := quic.ListenAddr(udpAddr.String(), generateTLSConfig(), &conf)
	if err != nil {
		utils.Log.Error().Str("error", err.Error()).Send()
		//Log.Error("err:%s, listen type:%s", err.Error(), udpAddr.Network())
		return utils.NewErrorSysSelf(err)
	}
	//监听一个地址和端口
	//utils.Log.Info().Str("Quic Listen to an IP", udpAddr.String()).Send()
	//Log.Debug("Quic Listen to an IP: %s", ip+":"+strconv.Itoa(int(port)))
	go this.listenerQuic(lisQuic)
	if !async {
		<-this.contextRoot.Done()
	}
	return utils.NewErrorSuccess()
}

func (this *QuicListen) listenerQuic(lisQuic *quic.Listener) {
	this.lisQuic = lisQuic
	var conn quic.Connection
	var err error
	for !this.closed.Load() {
		conn, err = lisQuic.Accept(context.Background())
		if err != nil {
			utils.Log.Error().Str("error", err.Error()).Send()
			continue
		}
		go this.newQuicConnect(conn)
	}
}

// 创建一个新的连接
func (this *QuicListen) newQuicConnect(conn quic.Connection) {
	defer utils.PrintPanicStack(this.engine.Log)
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		utils.Log.Error().Str("error", err.Error()).Send()
		//Log.Error("conn.AcceptStream err:%s", err.Error())
		return
	}

	var ERR utils.ERROR
	//新连接回调函数
	h := this.engine.GetDialBeforeEvent()
	if h != nil {
		ERR = h(stream)
		if ERR.CheckFail() {
			utils.Log.Error().Str("ERR", ERR.String()).Send()
			return
		}
	}

	serverConn := getServerQuicConn(this.engine)
	serverConn.conn = conn
	serverConn.stream = stream
	serverConn.Ip = conn.RemoteAddr().String()
	serverConn.Connected_time = time.Now().String()
	serverConn.outChan = make(chan *[]byte, 10000)
	serverConn.outChanCloseLock = new(sync.Mutex)
	serverConn.outChanIsClose = false
	this.engine.sessionStore.addSession(serverConn)

	serverConn.run()
	// Log.Debug("Accept remote addr:%s", conn.RemoteAddr().String())
	//新连接回调函数
	ah := this.engine.GetDialAfterEvent()
	if ah != nil {
		ERR = ah(serverConn)
		if ERR.CheckFail() {
			utils.Log.Error().Str("ERR", ERR.String()).Send()
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
func (this *QuicListen) DialAddrInfo(info *AddrInfo) (Session, utils.ERROR) {
	clientConn := getClientQuicConn(this.engine)
	clientConn.remoteMultiaddr = info.Multiaddr
	port, err := strconv.Atoi(info.Port)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return this.connect(clientConn, info.Addr, uint16(port))
}

/*
连接
*/
func (this *QuicListen) DialAddr(ip string, port uint16) (Session, utils.ERROR) {
	clientConn := getClientQuicConn(this.engine)
	return this.connect(clientConn, ip, port)
}

/*
连接
*/
func (this *QuicListen) connect(clientConn *ClientQuic, ip string, port uint16) (Session, utils.ERROR) {

	//clientConn.name = this.name
	err := clientConn.Connect(ip, port)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	var ERR utils.ERROR
	//新连接回调函数
	h := this.engine.GetAcceptBeforeEvent()
	if h != nil {
		ERR = h(clientConn.stream)
		if ERR.CheckFail() {
			utils.Log.Error().Str("ERR", ERR.String()).Send()
			clientConn.Close()
			return nil, ERR
		}
	}
	// 防止自己连自己
	//排除重复连接session
	this.engine.sessionStore.addSession(clientConn)
	clientConn.run()
	// Log.Info("AddClientConn 6666 %s:%d", ip, port)
	//新连接回调函数
	ah := this.engine.GetAcceptAfterEvent()
	if ah != nil {
		ERR = ah(clientConn)
		if ERR.CheckFail() {
			utils.Log.Error().Str("ERR", ERR.String()).Send()
			clientConn.Close()
			return clientConn, ERR
		}
	}
	clientConn.allowClose <- true
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		clientConn.Close()
		return nil, ERR
	}
	return clientConn, utils.NewErrorSuccess()
}

/*
销毁，断开连接，关闭监听
*/
func (this *QuicListen) Destroy() {
	this.closed.Store(true)
	this.lisQuic.Close()
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"P2p Go project."},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 100),

		KeyUsage:              x509.KeyUsageContentCommitment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-p2p-project"},
	}
}
