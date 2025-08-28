package engine

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"web3_gui/keystore/v1/base58"
	"web3_gui/utils"
)

/*
 * GodAddrInfo 上帝地址信息
 */
type GodAddrInfo struct {
	areaName string // areaname
	godIp    string // IP地址
	godPort  uint16 // 端口
}

/*
不能设置handler运行总量，不能容忍业务阻塞，所以用户自己去处理
*/
type Engine struct {
	name               string          //本机名称
	status             int             //服务器状态
	authes             map[string]Auth // 根据区域名称对应的区域信息管理器
	onceRead           *sync.Once
	interceptor        *InterceptorProvider
	sessionStore       *sessionStore
	closecallback      sync.Map // 根据区域名称对应的连接关闭回调方法
	lis                *net.TCPListener
	router             *Router
	isSuspend          bool                             //暂停服务器
	tcpHost            string                           //监听IP，格式0.0.0.0
	tcpPort            uint16                           //监听端口
	serverConnCallback map[string]ServerNewConnCallback // 根据区域名称对应的服务端连接回调方法
	clientConnCallback map[string]ClientNewConnCallback // 根据区域名称对应的客户端端连接回调方法
	godInfoes          map[string]GodAddrInfo           // 上帝地址信息
	areaLinkLock       *sync.RWMutex                    // 有关areaName操作锁
	onlyConnectList    sync.Map
	sessionSort        sync.Map
	lisQuic            *quic.Listener // Quic的监听listener
	quicPort           uint16         // 监听端口
	connecting         sync.Map       //正在连接中地址添加进map，成功连接后1s后删除
}

// @name   本服务器名称
func NewEngine(name string) *Engine {
	engine := new(Engine)
	engine.name = name
	engine.interceptor = NewInterceptor()
	engine.onceRead = new(sync.Once)
	engine.sessionStore = NewSessionStore()
	engine.router = NewRouter()
	engine.authes = make(map[string]Auth)
	engine.serverConnCallback = make(map[string]ServerNewConnCallback)
	engine.clientConnCallback = make(map[string]ClientNewConnCallback)
	engine.godInfoes = make(map[string]GodAddrInfo)
	engine.areaLinkLock = new(sync.RWMutex)
	engine.router.AddRouter([]byte(""), MSGID_heartbeat, engine.heartbeatHander)
	return engine
}

/*
注册一个普通消息
*/
func (this *Engine) RegisterMsg(areaName []byte, msgId uint64, handler MsgHandler) {
	//打印注册消息id
	utils.Log.Debug().Msgf("register message id %d", msgId)
	this.router.AddRouter(areaName, msgId, handler)
}

func (this *Engine) Listen(ip string, port uint32, async bool) error {
	listenHost := ip + ":" + strconv.Itoa(int(port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listenHost)
	if err != nil {
		utils.Log.Error().Msgf("%v", err)
		return err
	}

	this.lis, err = net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		utils.Log.Error().Msgf("%v", err)
		return err
	}

	// 设置engine监听的地址及端口
	this.tcpHost = ip
	this.tcpPort = uint16(port)

	//监听一个地址和端口
	utils.Log.Debug().Msgf("Listen to an IP：%s", ip+":"+strconv.Itoa(int(port)))
	if async {
		go this.listener(this.lis)
	} else {
		this.listener(this.lis)
	}
	return nil
	//	return this.net.Listen(ip, port)
}

func (this *Engine) ListenByListener(listener *net.TCPListener, async bool) error {
	if async {
		go this.listener(this.lis)
	} else {
		this.listener(this.lis)
	}
	return nil
}

func (this *Engine) listener(listener *net.TCPListener) {
	this.lis = listener
	//	this.ipPort = listener.Addr().String()
	var conn net.Conn
	var err error
	// var godHost string
	for !this.isSuspend {
		conn, err = this.lis.Accept()
		if err != nil {
			continue
		}
		if this.isSuspend {
			conn.Close()
			continue
		}

		// // 如果设置了上帝地址，则忽略其他非上帝地址的连接
		// if this.godIp != "" {
		// 	godHost = this.godIp + ":" + strconv.Itoa(int(this.godPort))
		// } else if godHost != "" {
		// 	godHost = ""
		// }
		// if godHost != "" {
		// 	if godHost != conn.RemoteAddr().String() {
		// 		utils.Log.Warn().Msgf("qlw-----不是上帝节点的地址, 不允许连接!!!, host:%s", conn.RemoteAddr().String())
		// 		conn.Close()
		// 		continue
		// 	}
		// }

		go this.newConnect(conn)
	}
}

// 自己比传入地址大 true
// 传入地址比自己大 false
func (this *Engine) CompareAddressNet(node AddressNet) bool {
	// utils.Log.Info().Msgf("需检查自己地址 this.name %s", this.name)
	// utils.Log.Info().Msgf("需检查自己地址是否正确", new(big.Int).SetBytes(base58.Decode(this.name)))
	if new(big.Int).SetBytes(base58.Decode(this.name)).Cmp(new(big.Int).SetBytes(node)) == 1 {
		return true
	} else {
		return false
	}
}

/*
 * 获取配置的只连节点数量
 */
func (this *Engine) GetOnlyConnectNum() int {
	index := 0
	this.onlyConnectList.Range(func(key, value interface{}) bool {
		index++
		return true
	})
	return index
}

// 创建一个新的连接
func (this *Engine) newConnect(conn net.Conn) {
	defer utils.PrintPanicStack(nil)
	remoteName, machineId, remoteAreaName, setGodTime, params, _, err := this.RecvKey(conn, this.name)
	if err != nil {
		//接受连接错误
		if err.Error() == io.EOF.Error() {
		} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
		} else {
			utils.Log.Warn().Msgf("Accept connection error:%s engineName:%s", err.Error(), this.name)
		}
		conn.Close()
		return
	}

	serverConn := this.sessionStore.getServerConn(this, remoteAreaName)

	addrBs := AddressNet([]byte(remoteName))
	serverConn.sendQueue.targetName = addrBs.B58String()
	serverConn.name = remoteName
	serverConn.conn = conn
	serverConn.Ip = conn.RemoteAddr().String()
	serverConn.Connected_time = time.Now().String()
	serverConn.outChan = make(chan *[]byte, 10000)
	serverConn.outChanCloseLock = new(sync.Mutex)
	serverConn.outChanIsClose = false
	serverConn.machineID = machineId
	serverConn.setGodTime = setGodTime
	this.sessionStore.addSession(remoteName, serverConn, "", []byte(remoteAreaName))

	serverConnCalllback, exist := this.GetServerConnCallback(remoteAreaName)
	if exist && serverConnCalllback != nil {
		go func() {
			err := serverConnCalllback(serverConn, params)
			serverConn.allowClose <- false
			if err != nil {
				serverConn.Close()
			}
		}()
	}
	serverConn.run()
	// fmt.Println(time.Now().String(), "建立连接", conn.RemoteAddr().String())
	// utils.Log.Debug().Msgf("Accept remote addr:%s", conn.RemoteAddr().String())
}

/*
添加一个连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
@powerful      是否是强连接
@return  name  对方的名称
*/
func (this *Engine) AddClientConn(areaName []byte, ip string, port uint32, powerful bool, onlyAddOneSess string) (ss Session, err error) {
	//utils.Log.Info().Msgf("AddClientConn 1111 %s:%d", ip, port, len(this.onlyConnectList))
	if this.GetOnlyConnectNum() != 0 {
		netCon := fmt.Sprintf("%s:%d", ip, port)
		if _, has := this.onlyConnectList.Load(netCon); !has {
			return nil, errors.New("not only connect node")
		}
	}

	clientConn := this.sessionStore.getClientConn(areaName, this)
	clientConn.name = this.name
	remoteName, params, connectKey, err := clientConn.Connect(ip, port)
	// utils.Log.Info().Msgf("AddClientConn 2222 %s:%d", ip, port)
	if err != nil {
		// utils.Log.Info().Msgf("AddClientConn 3333 %s:%d", ip, port)
		return nil, err
	}
	addrBs := AddressNet([]byte(remoteName))
	clientConn.sendQueue.targetName = addrBs.B58String()
	clientConn.name = remoteName
	// 防止自己连自己
	if addrBs.B58String() == this.name {
		this.connecting.Delete(connectKey)
		return
	}
	//排除重复连接session
	if as, ok := this.sessionStore.getSessionAll(areaName, remoteName); ok && len(as) != 0 {
		for _, one := range as {
			if one.GetMachineID() == clientConn.machineID {
				this.connecting.Delete(connectKey)
				//utils.Log.Warn().Msgf("Node has been connect : %s  machineID: %s self:%s areaName:%s", addrBs.B58String(), clientConn.machineID, this.name, hex.EncodeToString([]byte(one.GetAreaName())))
				return nil, errors.New(fmt.Sprintf("Node has been connect : %s  machineID: %s", addrBs.B58String(), clientConn.machineID))
			}
		}
	}

	switch onlyAddOneSess {
	case BothMod:
		//正常存session,存入sessionByIndex 和 up或者down中
		if this.CompareAddressNet(addrBs) {
			this.sessionStore.addSession(remoteName, clientConn, AddressDown, areaName)
		} else {
			this.sessionStore.addSession(remoteName, clientConn, AddressUp, areaName)
		}
	case KadMod:
		//只存sessionByIndex中
		this.sessionStore.addSession(remoteName, clientConn, "", areaName)
	case OnebyoneMod:
		//只存入up或者down中
		if this.CompareAddressNet(addrBs) {
			this.sessionStore.addOnlyDownSession(remoteName, clientConn, areaName)
		} else {
			this.sessionStore.addOnlyUpSession(remoteName, clientConn, areaName)
		}
	case VnodeMod:
		//只存sessionByIndex中
		this.sessionStore.addSession(remoteName, clientConn, "", areaName)
	default:
		this.connecting.Delete(connectKey)
		return nil, errors.New("未找到 AddClientConn 模式")
	}

	// utils.Log.Info().Msgf("AddClientConn 4444 %s:%d", ip, port)

	clientConnCallback, exist := this.GetClientConnCallback(utils.Bytes2string(areaName))
	if exist && clientConnCallback != nil {
		go func() {
			err := clientConnCallback(clientConn, params)
			clientConn.allowClose <- false
			if err != nil {
				clientConn.Close()
			}
		}()
	}
	// utils.Log.Info().Msgf("AddClientConn 5555 %s:%d", ip, port)
	clientConn.run()
	// utils.Log.Info().Msgf("AddClientConn 6666 %s:%d", ip, port)
	return clientConn, nil
}

/*
 * 根据ip地址及端口，获取连接session及node信息
 * @param	ip			ip地址
 * @param	port		端口
 * @param	setGodAddr	是否设置对方为自己的上帝节点
 * @return  ss  		连接session
 * @return	node		连接节点node
 * @return	err			是否发生错误
 */
func (this *Engine) GetConnectInfoByIpAddr(areaName []byte, ip string, port uint32) (ss *Client, node interface{}, err error) {
	if this.GetOnlyConnectNum() != 0 {
		netCon := fmt.Sprintf("%s:%d", ip, port)
		if _, has := this.onlyConnectList.Load(netCon); !has {
			return nil, nil, errors.New("not only connect node")
		}
	}

	// 1. 获取客户端连接
	clientConn := this.sessionStore.getClientConn(areaName, this)
	clientConn.name = this.name

	// 2. 根据ip及端口建立连接
	remoteName, params, connectKey, err := clientConn.Connect(ip, port)
	defer this.connecting.Delete(connectKey)
	if err != nil {
		return nil, nil, err
	}

	// 3. 获取对方的节点地址
	addrBs := AddressNet([]byte(remoteName))
	clientConn.sendQueue.targetName = addrBs.B58String()
	clientConn.name = remoteName
	//排除重复连接session
	if as, ok := this.sessionStore.getSessionAll(areaName, remoteName); ok && len(as) != 0 {
		for _, one := range as {
			if one.GetMachineID() == clientConn.machineID {
				utils.Log.Warn().Msgf("Node has been connect : %s  machineID: %s", addrBs.B58String(), clientConn.machineID)
				return nil, nil, errors.New(fmt.Sprintf("Node has been connect : %s  machineID: %s", addrBs.B58String(), clientConn.machineID))
			}
		}
	}
	if this.CompareAddressNet(addrBs) {
		this.sessionStore.addSession(remoteName, clientConn, AddressDown, areaName)
	} else {
		this.sessionStore.addSession(remoteName, clientConn, AddressUp, areaName)
	}

	return clientConn, params, nil
}

/*
 * 根据客户端连接session，和节点信息，建立连接处理
 * @param	clientConn		客户端连接session
 * @param	params			连接的节点信息
 */
func (this *Engine) AddConnectBySessionAndNode(areaName []byte, clientConn *Client, params interface{}) {
	// 1. 根据回调，处理连接信息
	// clientConnCallback, exist := this.clientConnCallback[utils.Bytes2string(areaName)]
	clientConnCallback, exist := this.GetClientConnCallback(utils.Bytes2string(areaName))
	if exist && clientConnCallback != nil {
		err := clientConnCallback(clientConn, params)
		clientConn.allowClose <- false
		if err != nil {
			clientConn.Close()
		}
	}

	// 2. 启动客户端连接处理
	clientConn.run()
}

// 添加一个拦截器，所有消息到达业务方法之前都要经过拦截器处理
func (this *Engine) AddInterceptor(itpr Interceptor) {
	this.interceptor.addInterceptor(itpr)
}

// 获得session
func (this *Engine) GetSession(areaName []byte, name string) (Session, bool) {
	return this.sessionStore.getSession(areaName, name)
}

// 真实节点onebyone规则连接，up和down地址
func (this *Engine) IsNodeUpdownSess(areaName []byte, addr AddressNet) bool {
	ud := this.GetAllDownUp(areaName)
	for _, one := range ud {
		if bytes.Equal([]byte(one), addr) {
			return true
		}
	}
	return false
}

// 获得某个地址所有session
func (this *Engine) GetSessionAll(areaName []byte, name string) ([]Session, bool) {
	return this.sessionStore.getSessionAll(areaName, name)
}

//通过ip地址和端口获得session,可以用于是否有重复连接
// func (this *Engine) GetSessionByHost(host string) Session {
// 	return this.sessionStore.getSessionByHost(host)
// }

/*
 * 删除sessionStore中特定session
 */
func (this *Engine) RemoveSession(areaName []byte, ss Session) {
	this.sessionStore.removeSession(utils.Bytes2string(areaName), ss)
}

/*
 * 删除customNameStore中特定session
 */
func (this *Engine) RemoveCustomSession(ss Session) {
	this.sessionStore.removeCustomSession(ss)
}

/*
 * 获取区域名的所有session
 */
func (this *Engine) GetAllSession(areaName []byte) []Session {
	return this.sessionStore.getAllSession(areaName)
}

/*
 * upSession中添加session
 */
func (this *Engine) AddUpSession(areaName []byte, ss Session) {
	this.sessionStore.addUpSession(areaName, ss)
}

/*
 * downSession中添加session
 */
func (this *Engine) AddDownSession(areaName []byte, ss Session) {
	this.sessionStore.addDownSession(areaName, ss)
}

/*
 * 获取本节点所有upSession
 */
func (this *Engine) GetAllUpSession(areaName []byte) []Session {
	return this.sessionStore.getAllUpSession(areaName)
}

/*
 * 获取本节点所有downSession
 */
func (this *Engine) GetAllDownSession(areaName []byte) []Session {
	return this.sessionStore.getAllDownSession(areaName)
}

/*
 * 获取本节点所有downSession 和 upSession
 */
func (this *Engine) GetAllDownUp(areaName []byte) []string {
	var addr []string
	for _, v := range this.sessionStore.getAllDownSession(areaName) {
		addr = append(addr, v.GetName())
	}
	for _, v := range this.sessionStore.getAllUpSession(areaName) {
		addr = append(addr, v.GetName())
	}
	return addr
}

/*
 * 加载并保存到正处于连接状态的地址
 * 返回 true map中存在
 * 返回 false map中不存在并保存
 */
func (this *Engine) LoadOrStoreConnecting(name string) bool {
	_, ok := this.connecting.LoadOrStore(name, struct{}{})
	// utils.Log.Warn().Msgf("加载或存 正在连接锁 !! %t", ok)
	return ok
}

/*
 * 保存正在连接状态的地址
 */
func (this *Engine) StoreConnecting(name string) {
	// utils.Log.Warn().Msgf("存储 正在连接锁 ！！ %s", name)
	this.connecting.Store(name, struct{}{})
}

/*
 * 直接删除正在处于连接状态中的地址
 */
func (this *Engine) DeleteConnecting(name string) {
	// utils.Log.Warn().Msgf("删除 正在连接锁 ！！ %s ", name)
	this.connecting.Delete(name)
}

/*
 * 延迟一秒删除正在处于连接状态中的地址
 */
func (this *Engine) DelyDeleteConnecting(name string) {
	// utils.Log.Warn().Msgf("延迟删除 正在连接锁 ！！ %s ", name)
	time.Sleep(2 * time.Second)
	this.DeleteConnecting(name)
}

/*
 * 检查是否存在特定session
 */
func (this *Engine) CheckInSessionByIndex(ss Session) bool {
	return this.sessionStore.checkInSessionByIndex(ss)
}

/*
 * 删除downSession中某个特定session
 */
func (this *Engine) DelNodeDownSession(areaName []byte, ss Session) {
	this.sessionStore.delNodeDownSession(areaName, ss)
}

/*
 * 删除upSession中某个特定session
 */
func (this *Engine) DelNodeUpSession(areaName []byte, ss Session) {
	this.sessionStore.delNodeUpSession(areaName, ss)
}

/*
 * 通过字符串获取排序好的网络地址数组
 */
func (this *Engine) GetSortSessionByKey(key string) ([]AddressNet, bool) {
	if addrs, ok := this.sessionSort.Load(key); !ok {
		return nil, false
	} else {
		if st, ok := addrs.([]AddressNet); ok && len(st) != 0 {
			return st, true
		} else {
			return nil, false
		}
	}
}

/*
 * 使用字符串作为key存储排序好的地址数组
 */
func (this *Engine) SetSortSessionByKey(key string, addrs []AddressNet) {
	if len(addrs) != 0 && key != "" {
		this.sessionSort.Store(key, addrs)
	}
}

// 添加自定义权限验证
func (this *Engine) AddAuth(auth Auth) {
	if auth == nil {
		return
	}

	areaName := auth.GetAreaName()
	if len(areaName) == 0 {
		return
	}

	this.areaLinkLock.Lock()
	defer this.areaLinkLock.Unlock()

	this.authes[auth.GetAreaName()] = auth
}

/*
设置关闭连接回调方法
*/
func (this *Engine) SetCloseCallback(areaName []byte, call CloseCallback) {
	if len(areaName) == 0 || call == nil {
		return
	}

	this.closecallback.Store(utils.Bytes2string(areaName), call)
}

/*
 * 设置服务器有新连接的回调方法
 */
func (this *Engine) SetServerConnCallback(areaName []byte, call ServerNewConnCallback) {
	if len(areaName) == 0 || call == nil {
		return
	}

	this.areaLinkLock.Lock()
	defer this.areaLinkLock.Unlock()

	this.serverConnCallback[utils.Bytes2string(areaName)] = call
}

/*
设置客户端新连接的回调方法
*/
func (this *Engine) SetClientConnCallback(areaName []byte, call ClientNewConnCallback) {
	if len(areaName) == 0 || call == nil {
		return
	}

	this.areaLinkLock.Lock()
	defer this.areaLinkLock.Unlock()

	this.clientConnCallback[utils.Bytes2string(areaName)] = call
}

/*
暂停服务器
*/
func (this *Engine) Suspend() {
	// utils.Log.Debug().Msgf("暂停服务器")
	this.isSuspend = true
}

/*
恢复服务器
*/
func (this *Engine) Recovery() {
	// utils.Log.Debug().Msgf("恢复服务器")
	this.isSuspend = false
	go this.listener(this.lis)
}

/*
 * 销毁，断开连接，关闭监听
 *
 * @param	areaName	[]byte	区域名
 */
func (this *Engine) Destroy(areaName []byte) {
	this.areaLinkLock.Lock()
	defer this.areaLinkLock.Unlock()

	// 获取区域名
	strAreaName := utils.Bytes2string(areaName)
	// 从区域信息管理器中找到区域信息
	_, exist := this.authes[strAreaName]
	if !exist {
		// 不存在，则直接退出
		return
	}

	// 根据区域名，获取所有的session，依次进行关闭操作
	for _, session := range this.sessionStore.getAllSession(areaName) {
		session.Close()
	}

	// 删除区域名对应的相关信息及回调
	delete(this.authes, strAreaName)
	this.closecallback.Delete(strAreaName)
	delete(this.serverConnCallback, strAreaName)
	delete(this.clientConnCallback, strAreaName)

	// 删除区域名对应的所有消息路由
	this.router.RemoveHandlers(strAreaName)

	// 删除区域名称对应的上帝地址信息
	delete(this.godInfoes, strAreaName)

	// 没有任何区域信息时，关闭engine
	if len(this.authes) == 0 {
		this.isSuspend = true
		this.lis.Close()
	}
}

/*
获取监听地址和端口
格式：0.0.0.0:8888
*/
func (this *Engine) GetTcpHost() (string, uint16) {
	return this.tcpHost, this.tcpPort
}

/*
心跳
*/
func (this *Engine) heartbeatHander(c Controller, msg Packet) {
}

/*
 * 设置上帝节点地址
 *
 * @auth qlw
 * @param ip	string		ip地址
 * @param port	uint16		端口
 */
func (this *Engine) SetGodAddr(areaName []byte, ip string, port uint16) {
	this.areaLinkLock.Lock()
	defer this.areaLinkLock.Unlock()

	// utils.Log.Warn().Msgf("qlw-----设置上帝节点地址, ip:%s, port:%d", ip, port)
	if len(areaName) == 0 {
		return
	}

	// 构建god地址信息
	strAreaName := utils.Bytes2string(areaName)
	var godAddrInfo GodAddrInfo = GodAddrInfo{
		areaName: strAreaName,
		godIp:    ip,
		godPort:  port,
	}

	// 保存god地址信息
	this.godInfoes[strAreaName] = godAddrInfo
}

/*
 * 获取当前上帝节点地址
 *
 * @auth qlw
 * @param areaName	[]byte		区域名称
 * @return ip		string		ip地址
 * @return port		uint16		端口
 */
func (this *Engine) GetGodAddr(areaName []byte) (ip string, port uint16) {
	this.areaLinkLock.RLock()
	defer this.areaLinkLock.RUnlock()

	// 转化areaName
	strAreaName := utils.Bytes2string(areaName)
	// 根据areaName查询记录是否存在
	if godAddrInfo, exist := this.godInfoes[strAreaName]; exist {
		// 返回记录信息
		return godAddrInfo.godIp, godAddrInfo.godPort
	}

	// 返回空值
	return "", 0
}

/*
 * GetAuth 获取区域对应的区域信息管理器
 *
 * @param	areaName	string			区域名称
 * @return	auth		Auth			区域信息管理器
 * @return	exist		bool			是否存在回调标识
 */
func (this *Engine) GetAuth(areaName string) (auth Auth, exist bool) {
	this.areaLinkLock.RLock()
	defer this.areaLinkLock.RUnlock()

	auth, exist = this.authes[areaName]

	return
}

/*
 * GetCloseCallback 获取区域对应的关闭回调方法
 *
 * @param	areaName	string			区域名称
 * @return	call		CloseCallback	关闭回调方法
 * @return	exist		bool			是否存在回调标识
 */
func (this *Engine) GetCloseCallback(areaName string) (callback CloseCallback, exist bool) {
	v, exist := this.closecallback.Load(areaName)
	if exist {
		callback, ok := v.(CloseCallback)
		if ok {
			return callback, true
		}
	}

	return nil, false
}

/*
 * GetServerConnCallback 获取区域对应的服务端连接回调方法
 *
 * @param	areaName		string					区域名称
 * @return	closeCallback	ServerNewConnCallback	回调方法
 * @return	exist			bool					是否存在回调标识
 */
func (this *Engine) GetServerConnCallback(areaName string) (callback ServerNewConnCallback, exist bool) {
	this.areaLinkLock.RLock()
	defer this.areaLinkLock.RUnlock()

	callback, exist = this.serverConnCallback[areaName]

	return
}

/*
 * GetClientConnCallback 获取区域对应的客户端端连接回调方法
 *
 * @param	areaName		string					区域名称
 * @return	closeCallback	ClientNewConnCallback	回调方法
 * @return	exist			bool					是否存在回调标识
 */
func (this *Engine) GetClientConnCallback(areaName string) (callback ClientNewConnCallback, exist bool) {
	this.areaLinkLock.RLock()
	defer this.areaLinkLock.RUnlock()

	callback, exist = this.clientConnCallback[areaName]

	return
}

/*
 * RecvKey 接收信息
 *
 * @param	conn			Conn		tcp连接
 * @param	name			string		自己节点id
 * @return	remoteName		string		对方节点id
 * @return	params			interface	对方节点信息
 * @return	err				error		错误信息
 * @return	remoteAreaName	string		对方的域名，主要用于当本地启动了多个area时，根据域名来查找对应的回调
 */
func (this *Engine) RecvKey(conn net.Conn, name string) (remoteName, machineID, remoteAreaName string, setGodTime int64, params interface{}, connectKey string, err error) {
	// utils.utils.Log.Info().Msgf().Msgf("%s RecvKey 11111 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
	//设置此方法总共11秒钟内完成验证，否则超时。
	conn.SetDeadline(time.Now().Add(time.Second * 11))
	defer conn.SetDeadline(time.Time{})

	//旧版接收对方网络id
	oldNetIdBs := make([]byte, 4)
	_, err = io.ReadFull(conn, oldNetIdBs)
	if err != nil {
		return "", "", "", 0, nil, "", err
	}
	areaNameResult := []byte{AreaNameResult_same}

	//接收对方网络id
	netIdBs := make([]byte, 32-4)
	_, err = io.ReadFull(conn, netIdBs)
	if err != nil {
		return "", "", "", 0, nil, "", err
	}
	targetAreaName := append(oldNetIdBs, netIdBs...)

	this.areaLinkLock.Lock()
	auth, exist := this.authes[utils.Bytes2string(targetAreaName)]
	this.areaLinkLock.Unlock()
	if !exist || auth == nil {
		areaNameResult = []byte{AreaNameResult_different}
	}
	_, err = conn.Write(areaNameResult)
	if err != nil {
		return "", "", "", 0, nil, "", err
	}

	// 节点areaName不相同，直接返回
	if !exist || auth == nil {
		return "", "", "", 0, nil, "", Error_different_netid
	}

	// utils.Log.Info().Msgf("[%s] recv connect", hex.EncodeToString(targetAreaName))
	remoteName, machineID, setGodTime, param, connectKey, err := auth.RecvKey(conn, name)
	return remoteName, machineID, utils.Bytes2string(targetAreaName), setGodTime, param, connectKey, err
}

/*
 * 白名单传递给Engine
 */
func (this *Engine) AddOnlyConnect(ss []string) {
	for i := range ss {
		this.onlyConnectList.Store(ss[i], struct{}{})
	}
}

/*
 * 获取本节点session总数
 */
func (this *Engine) GetSessionCnt(areaName []byte) int64 {
	return this.sessionStore.getSessionCnt(areaName)
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

/*
 * 监听quic信息
 */
func (this *Engine) ListenQuic(ip string, port uint32, async bool) error {
	listenHost := ip + ":" + strconv.Itoa(int(port))
	udpAddr, err := net.ResolveUDPAddr("udp4", listenHost)
	if err != nil {
		utils.Log.Error().Msgf("%v", err)
		return err
	}

	conf := quic.Config{
		MaxIdleTimeout:  time.Second,
		KeepAlivePeriod: 500 * time.Millisecond,
	}

	this.lisQuic, err = quic.ListenAddr(udpAddr.String(), generateTLSConfig(), &conf)
	if err != nil {
		utils.Log.Error().Msgf("err:%s, listen type:%s", err.Error(), udpAddr.Network())
		return err
	}

	// 设置engine监听的端口
	this.quicPort = uint16(port)

	//监听一个地址和端口
	utils.Log.Debug().Msgf("Quic Listen to an IP: %s", ip+":"+strconv.Itoa(int(port)))
	if async {
		go this.listenerQuic(this.lisQuic)
	} else {
		this.listenerQuic(this.lisQuic)
	}
	return nil
}

func (this *Engine) listenerQuic(listener *quic.Listener) {
	this.lisQuic = listener
	var conn quic.Connection
	var err error
	for !this.isSuspend {
		conn, err = this.lisQuic.Accept(context.Background())
		if err != nil {
			continue
		}
		if this.isSuspend {
			conn.CloseWithError(0, "")
			continue
		}

		go this.newQuicConnect(conn)
	}
}

// 创建一个新的连接
func (this *Engine) newQuicConnect(conn quic.Connection) {
	defer utils.PrintPanicStack(nil)
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		utils.Log.Error().Msgf("conn.AcceptStream err:%s", err.Error())
		return
	}

	remoteName, machineId, remoteAreaName, setGodTime, params, _, err := this.RecvQuicKey(conn, stream, this.name)
	if err != nil {
		//接受连接错误
		if err.Error() == io.EOF.Error() {
		} else if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
		} else {
			utils.Log.Warn().Msgf("Accept connection error:%s engineName:%s", err.Error(), this.name)
		}
		conn.CloseWithError(0, "")
		return
	}

	serverConn := this.sessionStore.getServerQuicConn(this, remoteAreaName)

	addrBs := AddressNet([]byte(remoteName))
	serverConn.sendQueue.targetName = addrBs.B58String()
	serverConn.name = remoteName
	serverConn.conn = conn
	serverConn.stream = stream
	serverConn.Ip = conn.RemoteAddr().String()
	serverConn.Connected_time = time.Now().String()
	serverConn.outChan = make(chan *[]byte, 10000)
	serverConn.outChanCloseLock = new(sync.Mutex)
	serverConn.outChanIsClose = false
	serverConn.machineID = machineId
	serverConn.setGodTime = setGodTime
	this.sessionStore.addSession(remoteName, serverConn, "", []byte(remoteAreaName))

	serverConnCalllback, exist := this.GetServerConnCallback(remoteAreaName)
	if exist && serverConnCalllback != nil {
		go func() {
			err := serverConnCalllback(serverConn, params)
			serverConn.allowClose <- false
			if err != nil {
				serverConn.Close()
			}
		}()
	}
	serverConn.run()
	// fmt.Println(time.Now().String(), "建立连接", conn.RemoteAddr().String())
	// utils.Log.Debug().Msgf("Accept remote addr:%s", conn.RemoteAddr().String())
}

/*
 * RecvQuicKey 接收信息
 *
 * @param	conn			Conn		tcp连接
 * @param	name			string		自己节点id
 * @return	remoteName		string		对方节点id
 * @return	params			interface	对方节点信息
 * @return	err				error		错误信息
 * @return	remoteAreaName	string		对方的域名，主要用于当本地启动了多个area时，根据域名来查找对应的回调
 */
func (this *Engine) RecvQuicKey(conn quic.Connection, stream quic.Stream, name string) (remoteName, machineID, remoteAreaName string, setGodTime int64, params interface{}, connectKey string, err error) {
	// utils.utils.Log.Info().Msgf().Msgf("%s RecvQuicKey 11111 %s remote:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), conn.LocalAddr().String(), conn.RemoteAddr().String())
	//设置此方法总共11秒钟内完成验证，否则超时。
	stream.SetDeadline(time.Now().Add(time.Second * 11))
	defer stream.SetDeadline(time.Time{})

	//旧版接收对方网络id
	oldNetIdBs := make([]byte, 4)
	_, err = io.ReadFull(stream, oldNetIdBs)
	if err != nil {
		return "", "", "", 0, nil, "", err
	}
	areaNameResult := []byte{AreaNameResult_same}

	//接收对方网络id
	netIdBs := make([]byte, 32-4)
	_, err = io.ReadFull(stream, netIdBs)
	if err != nil {
		return "", "", "", 0, nil, "", err
	}
	targetAreaName := append(oldNetIdBs, netIdBs...)

	this.areaLinkLock.Lock()
	auth, exist := this.authes[utils.Bytes2string(targetAreaName)]
	this.areaLinkLock.Unlock()
	if !exist || auth == nil {
		areaNameResult = []byte{AreaNameResult_different}
	}
	_, err = stream.Write(areaNameResult)
	if err != nil {
		return "", "", "", 0, nil, "", err
	}

	// 节点areaName不相同，直接返回
	if !exist || auth == nil {
		return "", "", "", 0, nil, "", Error_different_netid
	}

	// utils.Log.Info().Msgf("[%s] recv connect", hex.EncodeToString(targetAreaName))
	remoteName, machineID, setGodTime, param, connectKey, err := auth.RecvQuicKey(conn, stream, name)
	return remoteName, machineID, utils.Bytes2string(targetAreaName), setGodTime, param, connectKey, err
}

/*
添加一个Quic连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
@powerful      是否是强连接
@return  name  对方的名称
*/
func (this *Engine) AddClientQuicConn(areaName []byte, ip string, port uint32, powerful bool, onlyAddOneSess string) (ss Session, err error) {
	//utils.Log.Info().Msgf("AddClientConn 1111 %s:%d", ip, port, len(this.onlyConnectList))
	if this.GetOnlyConnectNum() != 0 {
		netCon := fmt.Sprintf("%s:%d", ip, port)
		if _, has := this.onlyConnectList.Load(netCon); !has {
			return nil, errors.New("not only connect node")
		}
	}

	clientConn := this.sessionStore.getClientQuicConn(areaName, this)
	clientConn.name = this.name
	remoteName, params, connectKey, err := clientConn.Connect(ip, port)
	// utils.Log.Info().Msgf("AddClientConn 2222 %s:%d", ip, port)
	if err != nil {
		// utils.Log.Info().Msgf("AddClientConn 3333 %s:%d", ip, port)
		return nil, err
	}
	addrBs := AddressNet([]byte(remoteName))
	clientConn.sendQueue.targetName = addrBs.B58String()
	clientConn.name = remoteName
	// 防止自己连自己
	if addrBs.B58String() == this.name {
		this.connecting.Delete(connectKey)
		return
	}
	//排除重复连接session
	if as, ok := this.sessionStore.getSessionAll(areaName, remoteName); ok && len(as) != 0 {
		for _, one := range as {
			if one.GetMachineID() == clientConn.machineID {
				this.connecting.Delete(connectKey)
				//utils.Log.Warn().Msgf("Node has been connect : %s  machineID: %s self:%s areaName:%s", addrBs.B58String(), clientConn.machineID, this.name, hex.EncodeToString([]byte(one.GetAreaName())))
				return nil, errors.New(fmt.Sprintf("Node has been connect : %s  machineID: %s", addrBs.B58String(), clientConn.machineID))
			}
		}
	}

	switch onlyAddOneSess {
	case BothMod:
		//正常存session,存入sessionByIndex 和 up或者down中
		if this.CompareAddressNet(addrBs) {
			this.sessionStore.addSession(remoteName, clientConn, AddressDown, areaName)
		} else {
			this.sessionStore.addSession(remoteName, clientConn, AddressUp, areaName)
		}
	case KadMod:
		//只存sessionByIndex中
		this.sessionStore.addSession(remoteName, clientConn, "", areaName)
	case OnebyoneMod:
		//只存入up或者down中
		if this.CompareAddressNet(addrBs) {
			this.sessionStore.addOnlyDownSession(remoteName, clientConn, areaName)
		} else {
			this.sessionStore.addOnlyUpSession(remoteName, clientConn, areaName)
		}
	case VnodeMod:
		//只存sessionByIndex中
		this.sessionStore.addSession(remoteName, clientConn, "", areaName)
	default:
		this.connecting.Delete(connectKey)
		return nil, errors.New("未找到 AddClientConn 模式")
	}

	// utils.Log.Info().Msgf("AddClientConn 4444 %s:%d", ip, port)

	clientConnCallback, exist := this.GetClientConnCallback(utils.Bytes2string(areaName))
	if exist && clientConnCallback != nil {
		go func() {
			err := clientConnCallback(clientConn, params)
			clientConn.allowClose <- false
			if err != nil {
				clientConn.Close()
			}
		}()
	}
	// utils.Log.Info().Msgf("AddClientConn 5555 %s:%d", ip, port)
	clientConn.run()
	// utils.Log.Info().Msgf("AddClientConn 6666 %s:%d", ip, port)
	return clientConn, nil
}

/*
 * 根据ip地址及端口，获取Quic连接session及node信息
 * @param	ip			ip地址
 * @param	port		端口
 * @param	setGodAddr	是否设置对方为自己的上帝节点
 * @return  ss  		连接session
 * @return	node		连接节点node
 * @return	err			是否发生错误
 */
func (this *Engine) GetQuicConnectInfoByIpAddr(areaName []byte, ip string, port uint32) (ss *ClientQuic, node interface{}, err error) {
	if this.GetOnlyConnectNum() != 0 {
		netCon := fmt.Sprintf("%s:%d", ip, port)
		if _, has := this.onlyConnectList.Load(netCon); !has {
			return nil, nil, errors.New("not only connect node")
		}
	}

	// 1. 获取客户端连接
	clientConn := this.sessionStore.getClientQuicConn(areaName, this)
	clientConn.name = this.name

	// 2. 根据ip及端口建立连接
	remoteName, params, connectKey, err := clientConn.Connect(ip, port)
	defer this.connecting.Delete(connectKey)
	if err != nil {
		return nil, nil, err
	}

	// 3. 获取对方的节点地址
	addrBs := AddressNet([]byte(remoteName))
	clientConn.sendQueue.targetName = addrBs.B58String()
	clientConn.name = remoteName
	//排除重复连接session
	if as, ok := this.sessionStore.getSessionAll(areaName, remoteName); ok && len(as) != 0 {
		for _, one := range as {
			if one.GetMachineID() == clientConn.machineID {
				utils.Log.Warn().Msgf("Node has been connect : %s  machineID: %s", addrBs.B58String(), clientConn.machineID)
				return nil, nil, errors.New(fmt.Sprintf("Node has been connect : %s  machineID: %s", addrBs.B58String(), clientConn.machineID))
			}
		}
	}
	if this.CompareAddressNet(addrBs) {
		this.sessionStore.addSession(remoteName, clientConn, AddressDown, areaName)
	} else {
		this.sessionStore.addSession(remoteName, clientConn, AddressUp, areaName)
	}

	return clientConn, params, nil
}

/*
 * 根据客户端连接session，和节点信息，建立连接处理
 * @param	clientConn		客户端连接session
 * @param	params			连接的节点信息
 */
func (this *Engine) AddQuicConnectBySessionAndNode(areaName []byte, clientConn *ClientQuic, params interface{}) {
	// 1. 根据回调，处理连接信息
	// clientConnCallback, exist := this.clientConnCallback[utils.Bytes2string(areaName)]
	clientConnCallback, exist := this.GetClientConnCallback(utils.Bytes2string(areaName))
	if exist && clientConnCallback != nil {
		err := clientConnCallback(clientConn, params)
		clientConn.allowClose <- false
		if err != nil {
			clientConn.Close()
		}
	}

	// 2. 启动客户端连接处理
	clientConn.run()
}
