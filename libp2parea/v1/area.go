package libp2parea

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1/addr_manager"
	"web3_gui/libp2parea/v1/config"
	gconfig "web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/global"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/ntp"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	proxyrec "web3_gui/libp2parea/v1/proxy_rec"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/libp2parea/v1/virtual_node/manager"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

type Area struct {
	AreaName                  [32]byte                      //域网络名称
	Keystore                  keystore.KeystoreInterface    //
	Pwd                       string                        //密码
	NodeManager               *nodeStore.NodeManager        //
	MessageCenter             *message_center.MessageCenter //
	SessionEngine             *engine.Engine                //消息引擎
	Vm                        *virtual_node.VnodeManager    //
	Vc                        *manager.VnodeCenter          //
	contextRoot               context.Context               //
	canceRoot                 context.CancelFunc            //
	GetNearSuperAddr_recvLock *sync.Mutex                   //
	isStartCore               bool                          //
	isOnline                  chan bool                     //当连入网络，给他个信号
	reconnLock                *sync.Mutex                   //断线重连锁
	isReconn                  bool                          //
	reConnTimer               *utils.BackoffTimerChan       //重连计时器
	destroy                   bool                          //销毁此对象,销毁了不能重连网络
	findNearNodeTimer         *utils.BackoffTimerChan       //
	autonomyFinishChanLock    *sync.RWMutex                 //关闭管道锁
	autonomyFinishChan        chan bool                     //自治连接完成等待信号
	sqlite3dbPath             string                        //sqlite3数据库
	leveldbPath               string                        //leveldb数据库存放地方
	connecting                *sync.Map                     //保存正在连接的节点地址。key:string=[ip:port];
	lockNewConnCallback       *sync.RWMutex                 //
	levelDB                   *utilsleveldb.LevelDB         //
	addrManager               *addr_manager.AddrManager     //
	GodHost                   string                        // 上帝节点地址
	GodID                     *nodeStore.AddressNet         // 上帝节点Id
	setGodHostLock            *sync.Mutex                   // 设置上帝节点地址锁
	closedCallbackFunc        []NodeEventCallbackHandler    // 节点关闭连接回调
	beenGodAddrCallbackFunc   []NodeEventCallbackHandler    // 节点被设置为超级代理节点地址回调
	newConnCallbackFunc       []NodeEventCallbackHandler    // 节点新建连接回调
	ProxyData                 *proxyrec.ProxyData           // 代理节点记录信息
	ProxyDetail               *proxyrec.ProxyDetailInfo     // 代理详情信息
	ProxyCache                *proxyrec.ProxyCache          // 代理缓存信息
	selfAddrs                 sync.Map                      // 自己地址列表
	connLimitType             int16                         // 连接方式限制, 默认支持所有的连接
}

func NewArea(areaName [32]byte, key keystore.KeystoreInterface, pwd string) (*Area, error) {
	contextRoot, canceRoot := context.WithCancel(context.Background())
	_, puk, err := key.GetNetAddr(pwd)
	if err != nil && err.Error() == keystore.ERROR_address_empty.Error() {
		_, puk, err = key.CreateNetAddr(pwd, pwd)
	}
	if err != nil {
		// utils.Log.Info().Msgf("GetNetAddr error:%s", err.Error())
		return nil, err
	}
	addrNet := nodeStore.BuildAddr(puk)

	sessionEngine := engine.NewEngine(addrNet.B58String())
	nodeManager, err := nodeStore.NewNodeManager(key, pwd)
	if err != nil {
		// utils.Log.Info().Msgf("NewNodeManager error:%s", err.Error())
		return nil, err
	}
	nodeManager.AreaNameSelf = areaName[:]
	area := &Area{
		AreaName:                  areaName,
		Keystore:                  key,
		NodeManager:               nodeManager,
		SessionEngine:             sessionEngine,
		contextRoot:               contextRoot, //
		canceRoot:                 canceRoot,   //
		GetNearSuperAddr_recvLock: new(sync.Mutex),
		isOnline:                  make(chan bool, 1),
		reconnLock:                new(sync.Mutex),
		findNearNodeTimer:         utils.NewBackoffTimerChan(time.Second*30, time.Second*30, time.Minute*1, time.Minute*2, time.Minute*4, time.Minute*6, time.Minute*8, time.Minute*10),
		autonomyFinishChanLock:    new(sync.RWMutex),
		autonomyFinishChan:        make(chan bool, 1),
		connecting:                new(sync.Map),
		lockNewConnCallback:       new(sync.RWMutex),
		setGodHostLock:            new(sync.Mutex),
		connLimitType:             config.CONN_TYPE_ALL,
	}
	area.addrManager = addr_manager.NewAddrManager()
	sessionEngine.AddOnlyConnect(config.OnlyConnectList)
	vm := virtual_node.NewVnodeManager(nodeManager, contextRoot, area)
	mcm := message_center.NewMessageCenter(nodeManager, sessionEngine, vm, key, contextRoot, areaName[:])
	vc := manager.NewVnodeCenter(mcm, nodeManager, sessionEngine, vm, contextRoot)
	area.MessageCenter = mcm
	area.Vm = vm
	area.Vc = vc
	area.CloseVnode()
	area.ProxyData = proxyrec.NewProxyData(contextRoot)
	area.ProxyDetail = proxyrec.NewProxyDetail(contextRoot, area.MessageCenter)
	area.ProxyCache = proxyrec.NewProxyCache(contextRoot)
	return area, nil
}

/*
 * NewAreaByGlobal 根据大区创建区域信息
 *
 * @param	areaName	[]byte		大区名称
 * @param	key			keystore	keystore
 * @param	pwd			string		密码
 * @param	global		*Global		大区信息
 * @return	area		*Area		生成的大区
 * @return	err			error		错误信息
 */
func NewAreaByGlobal(areaName [32]byte, key keystore.KeystoreInterface, pwd string, global *global.Global) (*Area, error) {
	// 检查大区是否包含相同域名的区域信息
	if global.CheckOrAddAreaInfo(utils.Bytes2string(areaName[:])) {
		return nil, errors.New("重复添加相同areaName的节点")
	}

	contextRoot, canceRoot := context.WithCancel(context.Background())

	// 获取大区对应的消息引擎
	sessionEngine := global.GetEngine()

	nodeManager, err := nodeStore.NewNodeManager(key, pwd)
	if err != nil {
		// utils.Log.Info().Msgf("NewNodeManager error:%s", err.Error())
		canceRoot()
		global.RemoveAreaInfo(areaName[:])
		return nil, err
	}
	nodeManager.AreaNameSelf = areaName[:]
	area := &Area{
		AreaName:                  areaName,
		Keystore:                  key,
		NodeManager:               nodeManager,
		SessionEngine:             sessionEngine,
		contextRoot:               contextRoot, //
		canceRoot:                 canceRoot,   //
		GetNearSuperAddr_recvLock: new(sync.Mutex),
		isOnline:                  make(chan bool, 1),
		reconnLock:                new(sync.Mutex),
		findNearNodeTimer:         utils.NewBackoffTimerChan(time.Second*30, time.Second*30, time.Minute*1, time.Minute*2, time.Minute*4, time.Minute*6, time.Minute*8, time.Minute*10),
		autonomyFinishChanLock:    new(sync.RWMutex),
		autonomyFinishChan:        make(chan bool, 1),
		connecting:                new(sync.Map),
		lockNewConnCallback:       new(sync.RWMutex),
		setGodHostLock:            new(sync.Mutex),
		connLimitType:             config.CONN_TYPE_ALL,
	}
	area.addrManager = addr_manager.NewAddrManager()
	vm := virtual_node.NewVnodeManager(nodeManager, contextRoot, area)
	mcm := message_center.NewMessageCenter(nodeManager, sessionEngine, vm, key, contextRoot, areaName[:])
	vc := manager.NewVnodeCenter(mcm, nodeManager, sessionEngine, vm, contextRoot)
	area.MessageCenter = mcm
	area.Vm = vm
	area.Vc = vc
	area.CloseVnode()
	area.ProxyData = proxyrec.NewProxyData(contextRoot)
	area.ProxyDetail = proxyrec.NewProxyDetail(contextRoot, area.MessageCenter)
	area.ProxyCache = proxyrec.NewProxyCache(contextRoot)
	return area, nil
}

func (this *Area) StartUP(isFirst bool, addr string, port uint16) utils.ERROR {
	this.destroy = false
	utils.Log.Info().Msgf("Local netid is: %s", this.NodeManager.NodeSelf.IdInfo.Id.B58String())
	if CommitInfo != "" {
		utils.Log.Warn().Msgf("libp2p commit is %s", CommitInfo)
	}
	// config.Init_LocalIP = addr
	// config.Init_LocalPort = port
	this.NodeManager.NodeSelf.Addr = addr
	this.NodeManager.NodeSelf.TcpPort = port
	this.NodeManager.NodeSelf.QuicPort = port

	// 如果没有设置机器id, 会用当前时间的纳秒生成一个默认机器id
	if this.GetMachineID() == "" {
		this.SetMachineID(strconv.FormatInt(time.Now().UnixNano(), 10))
	}
	utils.Log.Info().Msgf("StartUP 1111111111")
	// go this.startUp()
	go this.read(this.contextRoot)
	go this.getNearSuperIP(this.contextRoot)
	go this.SendNearLogicSuperIP(this.contextRoot)
	go this.readOutCloseConnName(this.contextRoot)
	go this.loopCleanMessageCache(this.contextRoot)
	go this.loopSendVnodeInfo(this.contextRoot)
	go this.Vc.LoopSendVnodeInfo()
	go this.cleanData(this.contextRoot)
	//初始化数据库
	// if this.sqlite3dbPath == "" {
	// 	this.sqlite3dbPath = config.SQLITE3DB_name
	// }
	// sqlite3_db.Init(this.sqlite3dbPath)
	if this.leveldbPath == "" {
		this.leveldbPath = config.Path_leveldb
	}
	utils.Log.Info().Msgf("StartUP 1111111111")
	ERR := this.initDB(this.leveldbPath)
	if !ERR.CheckSuccess() {
		return ERR
	}
	utils.Log.Info().Msgf("StartUP 1111111111")
	this.MessageCenter.SetLevelDB(this.levelDB)
	this.addrManager.SetLevelDB(this.levelDB)

	// this.Start(isFirst)
	this.StartEngine()
	utils.Log.Info().Msgf("StartUP 1111111111")
	this.registerHandler()
	utils.Log.Info().Msgf("StartUP 1111111111")
	if isFirst {
		return utils.NewErrorSuccess()
	}
	this.startUp()
	utils.Log.Info().Msgf("StartUP end")
	return utils.NewErrorSuccess()
}

/*
通过域管理器启动网络
*/
func (this *Area) StartUPGlobal(isFirst bool, global *global.Global, addr string, port uint16) utils.ERROR {
	utils.Log.Info().Msgf("Local netid is: %s", this.NodeManager.NodeSelf.IdInfo.Id.B58String())
	if CommitInfo != "" {
		utils.Log.Warn().Msgf("libp2p commit is %s", CommitInfo)
	}
	this.destroy = false
	eIP, ePort := global.GetTcpHost()
	if eIP == "" || ePort == 0 {
		this.NodeManager.NodeSelf.Addr, this.NodeManager.NodeSelf.TcpPort = addr, port
		global.Addr = addr
		this.NodeManager.NodeSelf.QuicPort = port
	} else {
		this.NodeManager.NodeSelf.Addr, this.NodeManager.NodeSelf.TcpPort = eIP, ePort
		this.NodeManager.NodeSelf.QuicPort = ePort
	}
	utils.Log.Info().Msgf("area listen ip:%s, port:%d", this.NodeManager.NodeSelf.Addr, this.NodeManager.NodeSelf.TcpPort)

	// 如果没有设置机器id, 会用当前时间的纳秒生成一个默认机器id
	if this.GetMachineID() == "" {
		this.SetMachineID(strconv.FormatInt(time.Now().UnixNano(), 10))
	}

	go this.read(this.contextRoot)
	go this.getNearSuperIP(this.contextRoot)
	go this.SendNearLogicSuperIP(this.contextRoot)
	go this.readOutCloseConnName(this.contextRoot)
	go this.loopCleanMessageCache(this.contextRoot)
	go this.loopSendVnodeInfo(this.contextRoot)
	go this.cleanData(this.contextRoot)
	//初始化数据库
	if this.leveldbPath == "" {
		this.leveldbPath = config.Path_leveldb
	}
	// 判断大区是否初始化过leveldb, 如果初始化过，则大区内的area公用同一个leveldb
	if global.GetLevelDB() == nil {
		// 没有初始化过，则初始化leveldb
		ERR := this.initDB(this.leveldbPath)
		if !ERR.CheckSuccess() {
			return ERR
		}
		global.SetLevelDB(this.levelDB)
	} else {
		this.levelDB = global.GetLevelDB()
	}
	this.MessageCenter.SetLevelDB(this.levelDB)
	this.addrManager.SetLevelDB(this.levelDB)
	this.StartGlobalEngine(global)
	if isFirst {
		return utils.NewErrorSuccess()
	}
	this.startUp()
	this.registerHandler()
	// utils.Log.Info().Msgf("StartUP Global end")
	return utils.NewErrorSuccess()
}

/*
启动消息服务器
@isFirst    bool      是否是首节点
@addr       string    监听地址
*/
func (this *Area) StartEngine() bool {
	auth := NewAuth(this.AreaName, this.NodeManager, this.SessionEngine, this.Vc)
	this.SessionEngine.AddAuth(auth)
	this.SessionEngine.SetCloseCallback(this.AreaName[:], this.closeConnCallback)
	this.SessionEngine.SetClientConnCallback(this.AreaName[:], this.clientNewConnCallback)
	this.SessionEngine.SetServerConnCallback(this.AreaName[:], this.serverNewConnCallback)
	this.RegisterCoreMsg()
	//占用本机一个端口
	var err error
	if this.CheckSupportTcpConn() {
		for i := 0; i < 100; i++ {
			err = this.SessionEngine.Listen("0.0.0.0", uint32(this.NodeManager.NodeSelf.TcpPort+uint16(i)), true)
			if err != nil {
				continue
			} else {
				//得到本机可用端口
				port := this.NodeManager.NodeSelf.TcpPort + uint16(i)
				if !config.Init_IsMapping {
					this.NodeManager.NodeSelf.TcpPort = port
				}
				//加载超级节点ip地址
				// return true
				break
			}
		}
		if err != nil {
			utils.Log.Error().Msgf("listen tcp err:%s", err.Error())
			return false
		}
	}

	//启动quic
	if this.CheckSupportQuicConn() {
		if this.NodeManager.NodeSelf.QuicPort != 0 {
			for i := 0; i < 100; i++ {
				err = this.SessionEngine.ListenQuic("0.0.0.0", uint32(this.NodeManager.NodeSelf.QuicPort+uint16(i)), true)
				if err != nil {
					continue
				} else {
					//得到本机可用端口
					port := this.NodeManager.NodeSelf.QuicPort + uint16(i)
					if !config.Init_IsMapping {
						this.NodeManager.NodeSelf.QuicPort = port
					}
					break
				}
			}
			if err != nil {
				return false
			}
		}
	}

	return true
}

/*
 * 启动消息服务器
 *
 * @param	global		*Global	大区信息
 * @return  success     bool    是否设置成功
 */
func (this *Area) StartGlobalEngine(global *global.Global) bool {
	// 判断大区信息是否有效
	if global == nil {
		return false
	}

	auth := NewAuth(this.AreaName, this.NodeManager, this.SessionEngine, this.Vc)
	this.SessionEngine.AddAuth(auth)

	this.SessionEngine.SetCloseCallback(this.AreaName[:], this.closeConnCallback)
	this.SessionEngine.SetClientConnCallback(this.AreaName[:], this.clientNewConnCallback)
	this.SessionEngine.SetServerConnCallback(this.AreaName[:], this.serverNewConnCallback)
	this.RegisterCoreMsg()

	// Global版只会启动一次engine
	if global.StartEngine {
		if !config.Init_IsMapping {
			_, this.NodeManager.NodeSelf.TcpPort = global.GetTcpHost()
		}

		return true
	}

	// 占用本机一个端口
	var err error
	for i := 0; i < 100; i++ {
		err = this.SessionEngine.Listen("0.0.0.0", uint32(this.NodeManager.NodeSelf.TcpPort+uint16(i)), true)
		if err != nil {
			continue
		} else {
			// 得到本机可用端口
			port := this.NodeManager.NodeSelf.TcpPort + uint16(i)
			if !config.Init_IsMapping {
				this.NodeManager.NodeSelf.TcpPort = port
			}
			// 更新启动引擎状态
			global.StartEngine = true
			global.Port = port
			return true
		}
	}
	return false
}

/*
有新地址就连接到网络中去
已经连入网络，有重复链接也返回正确
*/
func (this *Area) startUp() (success bool) {
	//1.实现同时连接多个节点。
	//2.有一个节点连接成功了就返回，不要一直等待。
	//3.返回了，也继续连接其他节点。

	// utils.Log.Info().Msgf("startUp start")
	success = false
	// utils.Log.Debug().Msgf("获取所有发现节点地址 start")
	addrs, dns := this.addrManager.LoadAddrForAll(this.AreaName[:])

	//电信网络移动端连接域名很慢
	//先连接缓存中的ip地址
	success = this.syncConnectNet(addrs)
	if success {
		utils.Log.Info().Msgf("startUp end:%t", success)
		return
	}
	//再连接域名。连接域名需要一次预连接，速度慢
	success = this.syncConnectNet(dns)

	utils.Log.Info().Msgf("startUp end:%t", success)
	return success
	// return this.CheckOnline()
	// utils.Log.Debug().Msgf("获取所有发现节点地址 end")
	// return
}

/*
异步链接到网络中去
@return    bool    是否连接成功
*/
func (this *Area) syncConnectNet(addrs []string) (success bool) {
	success = false
	if len(addrs) == 0 {
		return
	}
	//去除重复
	addrs = utils.DistinctString(addrs)
	// for _, one := range addrs {
	// 	utils.Log.Info().Msgf("去除重复后的ip:%s", one)
	// }

	resultSuccess := make(chan bool, len(addrs))
	token := make(chan bool, config.GetCPUSyncNum())

	countResult := 0 //返回计数
	for _, addr := range addrs {
		if success {
			//连接成功，也继续连接其他节点
			go this.syncConnectNetOne(addr, token, resultSuccess)
			continue
		}
		select {
		case token <- false:
			go this.syncConnectNetOne(addr, token, resultSuccess)
		case <-resultSuccess:
			// utils.Log.Info().Msgf("等待返回:%d", countResult)
			countResult++
			success = true
		}
	}
	// utils.Log.Info().Msgf("还需等待返回:%d", len(addrs)-countResult)
	for i := 0; i < len(addrs)-countResult; i++ {
		// utils.Log.Info().Msgf("等待返回:%d", i)
		ok := <-resultSuccess
		if ok {
			success = true
			break
		}
	}
	return
}

/*
异步链接到网络中去
@return    bool    是否连接成功
*/
func (this *Area) syncConnectNetOne(addr string, token chan bool, resultSuccess chan bool) {
	connSuccess := false
	// utils.Log.Info().Msgf("syncConnectNet start")
	defer func() {
		// utils.Log.Info().Msgf("syncConnectNet end 111111")
		resultSuccess <- connSuccess
		// utils.Log.Info().Msgf("syncConnectNet end 222222", ok)
		<-token
		// utils.Log.Info().Msgf("syncConnectNet end 3333333")
	}()

	//如果是域名
	if addr_manager.IsDNS(addr) {
		addr = addr_manager.AnalysisDNS(addr, time.Second*10)
		if addr == "" {
			return
		}
	}

	//接收到超级节点地址消息
	// utils.Log.Debug().Msgf("有新的地址 %s", addr)
	host, portStr, _ := net.SplitHostPort(addr)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return
	}
	//判断节点地址是否是自己
	// utils.Log.Debug().Msgf("self node %s %d %s %d", ip, port, nodeStore.NodeSelf.Addr, nodeStore.NodeSelf.TcpPort)
	var usePort uint16
	if this.CheckSupportQuicConn() {
		usePort = this.NodeManager.NodeSelf.QuicPort
	} else {
		usePort = this.NodeManager.NodeSelf.TcpPort
	}
	if this.NodeManager.NodeSelf.Addr == host && usePort == uint16(port) {
		return
	}
	//查询是否已经有这个连接了，有了就不连接
	// session := this.SessionEngine.GetSessionByHost(host + ":" + strconv.Itoa(int(port)))
	// if session != nil {
	// 	return
	// }

	// 判断连接的地址是不是自己地址
	strConnAddr := host + ":" + portStr
	if _, exist := this.selfAddrs.Load(strConnAddr); exist {
		// utils.Log.Error().Msgf("自己连接自己, 直接退出!!!!!")
		return
	}

	// 如果设置了上帝地址，则忽略所有非上帝地址的连接
	if this.GodHost != "" && this.GodHost != addr {
		// utils.Log.Warn().Msgf("qlw-----不是上帝节点的地址, 不允许连接!!! addr:%s", addr)
		return
	}

	connSuccess, err = this.connectNet(host, uint16(port))
	if err != nil && err.Error() == config.ERROR_conn_self.Error() {
		this.selfAddrs.Store(strConnAddr, struct{}{})
	}

}

/*
链接到网络中去
@return    bool    是否连接成功
*/
func (this *Area) connectNet(ip string, port uint16) (bool, error) {
	//判断节点地址是否是自己
	// utils.Log.Debug().Msgf("self node %s %d %s %d", ip, port, this.NodeManager.NodeSelf.Addr, this.NodeManager.NodeSelf.TcpPort)
	var usePort uint16
	if this.CheckSupportQuicConn() {
		usePort = this.NodeManager.NodeSelf.QuicPort
	} else {
		usePort = this.NodeManager.NodeSelf.TcpPort
	}
	if this.NodeManager.NodeSelf.Addr == ip && usePort == port {
		return false, nil
	}
	//查询是否已经有这个连接了，有了就不连接
	// session := this.SessionEngine.GetSessionByHost(ip + ":" + strconv.Itoa(int(port)))
	// if session != nil {
	// 	//连接已经存在
	// 	return true
	// }

	//开始连接节点
	// utils.Log.Debug().Msgf("Start connecting nodes:%s:%d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), ip, port)
	// utils.Log.Info().Msgf("开始连接 111111")
	//判断正在连接的队列中有这个连接
	connectingKey := ip + ":" + strconv.Itoa(int(port))
	// utils.Log.Info().Msgf("%s 添加 正在连接的地址:%s", this.GetNetId().B58String(), connectingKey)
	_, ok := this.connecting.LoadOrStore(connectingKey, 0)
	if ok {
		// utils.Log.Info().Msgf("有重复的")
		return false, nil
	}
	var bConnSuccess bool
	var session engine.Session
	var err error
	// 优先连接quic
	if this.CheckSupportQuicConn() {
		session, err = this.SessionEngine.AddClientQuicConn(this.AreaName[:], ip, uint32(port), false, engine.BothMod)
		// 如果连接失败, 则再使用tcp进行连接
		if err != nil {
			utils.Log.Error().Msgf("连接quic失败,err: %s", err.Error())
		} else {
			bConnSuccess = true
		}
	}
	if !bConnSuccess && this.CheckSupportTcpConn() {
		session, err = this.SessionEngine.AddClientConn(this.AreaName[:], ip, uint32(port), false, engine.BothMod)
		this.connecting.Delete(connectingKey)
		// utils.Log.Info().Msgf("%s 删除 正在连接的地址:%s", this.GetNetId().B58String(), connectingKey)
		if err != nil {
			//连接失败
			// utils.Log.Error().Msgf("[%s] connection failed %s %d %v", hex.EncodeToString(this.AreaName[:]), ip, port, err)
			if err.Error() == engine.Error_different_netid.Error() {
				// addr_manager.RemoveIP(ip, port)
			}
			return false, err
		}
		bConnSuccess = true
	}
	this.connecting.Delete(connectingKey)
	if !bConnSuccess {
		if err == nil {
			err = errors.New("没有可用的连接")
		}
		return false, err
	}

	mh := nodeStore.AddressNet([]byte(session.GetName()))
	// utils.Log.Info().Msgf("更换超级节点id:%s", mh.B58String())
	// this.NodeManager.SuperPeerId = &mh
	this.NodeManager.SetSuperPeerId(&mh)

	// 记录节点的超级代理地址
	if this.GodHost != "" && this.GodHost == connectingKey {
		this.GodID = &mh
	}

	// utils.Log.Debug().Msgf("超级节点为: %s", nodeStore.SuperPeerId.B58String())
	config.IsOnline = true

	select {
	case this.isOnline <- true:
	default:
	}
	return true, nil
}

/*
 * 主动移除服务端节点旧session，保持每个machineid只有一个session连接
 */
func (this Area) removeOldConn(ss engine.Session, params interface{}) {
	node, ok := params.(*nodeStore.Node)
	if !ok {
		return
	}
	existNode := this.NodeManager.FindNode(&node.IdInfo.Id)
	if existNode != nil {
		sessions := existNode.GetSessions()
		for i := range sessions {
			//utils.Log.Info().Msgf("serverNewConnCallback 11111 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
			if sessions[i].GetMachineID() != "" && sessions[i].GetMachineID() == ss.GetMachineID() {
				sessions[i].Close()
				existNode.RemoveSession(sessions[i])
				this.SessionEngine.RemoveCustomSession(sessions[i])
				this.SessionEngine.RemoveSession(this.AreaName[:], sessions[i])
			}
		}
	}
}

/*
服务器有新连接，触发的回调函数
*/
func (this *Area) serverNewConnCallback(ss engine.Session, params interface{}) error {
	this.removeOldConn(ss, params)
	node := params.(*nodeStore.Node)
	connectKey := fmt.Sprintf("%s_%s", utils.Bytes2string(node.IdInfo.Id), node.MachineIDStr)
	defer func() {
		go this.SessionEngine.DelyDeleteConnecting(connectKey)
		// for i, ss := range this.SessionEngine.GetAllSession(this.AreaName[:]) {
		// 	utils.Log.Info().Msgf(" server index %d addr %s", i, nodeStore.AddressNet(ss.GetName()).B58String())
		// }
	}()
	node.AddSession(ss)
	// utils.Log.Info().Msgf("serverNewConnCallback self:%s target:%s index:%d machineID:%s", this.GetNetId().B58String(), node.IdInfo.Id.B58String(), ss.GetIndex(), ss.GetMachineID())
	this.lockNewConnCallback.Lock()
	defer this.lockNewConnCallback.Unlock()

	// utils.Log.Info().Msgf("%s serverNewConnCallback 11111 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
	//查询这个节点是否存在
	existNode := this.NodeManager.FindNode(&node.IdInfo.Id)
	// utils.Log.Info().Msgf("serverNewConnCallback 11111 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
	if existNode != nil {
		// utils.Log.Info().Msgf("serverNewConnCallback 22222 self:%s target:%s index:%d", this.GetNetId().B58String(), node.IdInfo.Id.B58String(), ss.GetIndex())
		// utils.Log.Info().Msgf("%s serverNewConnCallback 22222 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
		// utils.Log.Info().Msgf("serverNewConnCallback 22222 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
		//已经存在
		// existNode.AddSession(ss)

		// 如果当前节点被连接的客户端设置成为上帝节点，则调用相应回调方法
		if node.SetGod {
			existNode.AddSession(ss)
			// 节点被设置为上帝节点，调用节点回调方法
			for _, h := range this.beenGodAddrCallbackFunc {
				// 开启协程进行处理
				go h(node.IdInfo.Id, node.MachineIDStr)
			}

			// 设置代理信息
			// 如果客户端没有设置版本号, 则服务器进行设置
			if node.SetGodTime == 0 {
				node.SetGodTime = ntp.GetNtpTime().UnixMilli()
				if node.SetGodTime == 0 {
					node.SetGodTime = time.Now().UnixMilli()
				}
			}
			this.ProxyDetail.AddProxyDetail(node.IdInfo.Id, this.GetNetId(), node.MachineIDStr, node.SetGodTime)
		}

		return nil
	}
	// utils.Log.Info().Msgf("%s serverNewConnCallback 33333 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("serverNewConnCallback 33333 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
	if node.Type == nodeStore.Node_type_proxy {
		//不是公网ip地址
		this.NodeManager.AddProxyNode(*node)
	} else {
		ok := this.NodeManager.AddNode(*node)
		// utils.Log.Info().Msgf("add super nodeid self:%s targetID:%s %t", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), node.IdInfo.Id.B58String(), ok)
		if !ok {
			// utils.Log.Info().Msgf("%s serverNewConnCallback 44444 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
			//不是自己的逻辑节点，则是对方的逻辑节点
			this.NodeManager.AddNodesClient(*node)
		}
	}

	// 如果当前节点被连接的客户端设置成为上帝节点，则调用相应回调方法
	if node.SetGod {
		// 节点被设置为上帝节点，调用节点回调方法
		for _, h := range this.beenGodAddrCallbackFunc {
			// 开启协程进行处理
			go h(node.IdInfo.Id, node.MachineIDStr)
		}

		// 设置代理信息
		// 如果客户端没有设置版本号, 则服务器进行设置
		if node.SetGodTime == 0 {
			node.SetGodTime = ntp.GetNtpTime().UnixMilli()
			if node.SetGodTime == 0 {
				node.SetGodTime = time.Now().UnixMilli()
			}
		}
		this.ProxyDetail.AddProxyDetail(node.IdInfo.Id, this.GetNetId(), node.MachineIDStr, node.SetGodTime)
	}

	// 新建连接后，调用节点新建连接的回调方法
	for _, h := range this.newConnCallbackFunc {
		// 开启协程进行处理
		go h(node.IdInfo.Id, ss.GetMachineID())
	}

	//发送节点上线信号
	select {
	case this.isOnline <- true:
	default:
	}
	// utils.Log.Info().Msgf("serverNewConnCallback 55555 self:%s target:%s index:%d", this.GetNetId().B58String(), node.IdInfo.Id.B58String(), ss.GetIndex())
	// utils.Log.Info().Msgf("%s serverNewConnCallback 55555 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
	return nil
}

/*
客户端有新连接，触发的回调函数
*/
func (this *Area) clientNewConnCallback(ss engine.Session, params interface{}) error {
	node := params.(*nodeStore.Node)
	connectKey := fmt.Sprintf("%s_%s", utils.Bytes2string(node.IdInfo.Id), node.MachineIDStr)
	defer func() {
		go this.SessionEngine.DelyDeleteConnecting(connectKey)
		// for i, ss := range this.SessionEngine.GetAllSession(this.AreaName[:]) {
		// 	utils.Log.Info().Msgf(" client index %d addr %s", i, nodeStore.AddressNet(ss.GetName()).B58String())
		// }
	}()
	// utils.Log.Info().Msgf("%s clientNewConnCallback 11111 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("clientNewConnCallback self:%s target:%s index:%d machineID:%s", this.GetNetId().B58String(), node.IdInfo.Id.B58String(), ss.GetIndex(), ss.GetMachineID())
	this.lockNewConnCallback.Lock()
	defer this.lockNewConnCallback.Unlock()

	node.AddSession(ss)
	//能直接通过ip地址访问的节点，一定是超级节点。
	node.SetIsSuper(true)

	//查询这个节点是否存在
	existNode := this.NodeManager.FindNode(&node.IdInfo.Id)
	if existNode != nil && (existNode.Type == nodeStore.Node_type_oneByone || existNode.Type == nodeStore.Node_type_logic || existNode.Type == nodeStore.Node_type_client) {
		return nil
	} else if existNode != nil {
		// utils.Log.Info().Msgf("clientNewConnCallback 22222 self:%s target:%s index:%d", this.GetNetId().B58String(), node.IdInfo.Id.B58String(), ss.GetIndex())
		// utils.Log.Info().Msgf("%s clientNewConnCallback 22222 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
		// utils.Log.Info().Msgf("clientNewConnCallback 22222 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
		// ss.Close()
		// utils.Log.Info().Msgf("%s clientNewConnCallback 22222 33333 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
		return engine.Error_node_unwanted
	}

	ok := this.NodeManager.AddNode(*node)
	// utils.Log.Info().Msgf("add super nodeid self:%s targetID:%s %t", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), node.IdInfo.Id.B58String(), ok)
	if !ok {
		// utils.Log.Info().Msgf("%s clientNewConnCallback 33333 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
		// utils.Log.Info().Msgf("clientNewConnCallback 33333 self:%s target:%s index:%d", this.GetNetId().B58String(), node.IdInfo.Id.B58String(), ss.GetIndex())
		// utils.Log.Info().Msgf("不需要的节点:%s", hex.EncodeToString(node.IdInfo.Id))
		// utils.Log.Info().Msgf("clientNewConnCallback 33333 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
		return engine.Error_node_unwanted
	}
	this.Vc.NoticeAddNode(node.IdInfo.Id)
	// utils.Log.Info().Msgf("%s clientNewConnCallback 44444 %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
	// utils.Log.Info().Msgf("clientNewConnCallback 55555 self:%s target:%s index:%d", this.GetNetId().B58String(), node.IdInfo.Id.B58String(), ss.GetIndex())
	// utils.Log.Info().Msgf("clientNewConnCallback 44444 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())

	// 新建连接后，调用节点新建连接的回调方法
	for _, h := range this.newConnCallbackFunc {
		// 开启协程进行处理
		go h(node.IdInfo.Id, ss.GetMachineID())
	}

	//发送节点上线信号
	select {
	case this.isOnline <- true:
	default:
	}

	return nil
}

/*
关闭服务器回调函数
*/
func (this *Area) ShutdownCallback() {
	//回收映射的端口
	Reclaim()
	// addrm.CloseBroadcastServer()
	// fmt.Println("Close over")
}

/*
一个连接断开后的回调方法
*/
func (this *Area) closeConnCallback(ss engine.Session) {
	// utils.Log.Info().Msgf("closeConnCallback start self:%s", this.GetNetId().B58String())
	this.lockNewConnCallback.Lock()
	defer this.lockNewConnCallback.Unlock()

	name := ss.GetName()
	addrNet := nodeStore.AddressNet([]byte(name))
	utils.Log.Debug().Msgf("Node offline self:%s target:%s index:%d machineID:%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), ss.GetIndex(), ss.GetMachineID())

	this.SessionEngine.RemoveCustomSession(ss)
	this.SessionEngine.RemoveSession(this.AreaName[:], ss)
	//删除虚拟节点之中的真实节点对应本次断开的地址
	this.Vm.RLock()
	for _, one := range this.Vm.VnodeMap {
		one.Lock.Lock()
		for _, up := range one.GetUpVnodeInfo() {
			if bytes.Equal(up.Nid, addrNet) {
				one.UpVnodeInfo.Delete(utils.Bytes2string(up.Vid))
			}
		}
		for _, down := range one.GetDownVnodeInfo() {
			if bytes.Equal(down.Nid, addrNet) {
				one.DownVnodeInfo.Delete(utils.Bytes2string(down.Vid))
			}
		}
		one.Lock.Unlock()
	}
	this.Vm.RUnlock()
	flood.GroupWaitRecv.ResponseItrGroup(strconv.Itoa(int(ss.GetIndex())), engine.ERROR_send_timeout)

	//不在管理节点内
	node := this.NodeManager.FindNode(&addrNet)
	if node == nil {
		// utils.Log.Debug().Msgf("Node offline 22222 self:%s target:%s index:%d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), ss.GetIndex())
		return
	}
	node.RemoveSession(ss)

	// 删除节点对应的代理信息, 或者更新代理节点保存的节点信息
	if node.IsApp {
		if node.SetGod {
			this.ProxyDetail.RemoveProxyDetail(&addrNet, ss.GetMachineID(), ss.GetSetGodTime())
		}
	}

	//当允许一个节点有多个连接，那么多个连接都离线之后才能删除管理
	//检查是否还有其他连接存在
	if node.CheckHaveOtherSessions(ss) {
		// utils.Log.Debug().Msgf("Node offline 33333 self:%s target:%s index:%d machineID:%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), ss.GetIndex(), ss.GetMachineID())
		// 连接断开后，调用节点断开连接的回调方法
		for _, h := range this.closedCallbackFunc {
			// 开启协程进行处理
			go h(addrNet, ss.GetMachineID())
		}

		return
	}

	this.NodeManager.DelNode(&addrNet)
	DelNodeAddrSpeed(addrNet)
	this.Vc.NoticeRemoveNode(addrNet)
	// 删除客户端节点信息
	this.Vm.DelClientVnodeinfo(addrNet)

	//自己是客户端节点，不需要广播其他节点下线
	//只广播超级节点下线信息
	// if len(this.NodeManager.GetNodesClient()) > 0 {
	if !this.NodeManager.NodeSelf.IsApp && node.GetIsSuper() {
		// 检查对象是否在线
		_, _, _, err := this.SendP2pMsgWaitRequest(gconfig.MSGID_checkAddrOnline, &addrNet, nil, 5*time.Second)
		// 不在线，则通知其他节点，并删除对应代理记录信息
		if err != nil {
			// 广播通知该节点已下线
			content, err := node.Proto()
			if err == nil {
				// utils.Log.Info().Msgf("发送节点下线广播消息: nid:%s, self:%s", addrNet.B58String(), this.Vm.DiscoverVnodes.Vnode.Nid.B58String())
				if err := this.SendMulticastMsg(gconfig.MSGID_multicast_offline_recv, &content); err != nil {
					// utils.Log.Info().Msgf("发送节点下线广播消息 err:%s", err)
				}
			}

			if node.GetIsSuper() {
				this.ProxyDetail.NodeOfflineDeal(&addrNet)
			}
		}
	}

	// 连接断开后，调用节点断开连接的回调方法
	for _, h := range this.closedCallbackFunc {
		// 开启协程进行处理
		go h(addrNet, ss.GetMachineID())
	}

	// 连接断开后，删除连接对应的加密管道
	if this.MessageCenter != nil && this.MessageCenter.RatchetSession != nil {
		this.MessageCenter.CleanHEInfo(addrNet, ss.GetMachineID())
	}

	//对比删除此节点，前后，是否有变化
	// if this.NodeManager.EqualLogicNodes(logicNodes) {
	// 	utils.Log.Info().Msgf("有变化就重新查询自己的逻辑节点")
	// 	//有变化就重新查询自己的逻辑节点
	// }
	// if node.Type == nodeStore.Node_type_logic ||
	// 	node.Type == nodeStore.Node_type_client ||
	// 	node.Type == nodeStore.Node_type_white_list {
	// }
	this.findNearNodeTimer.Release()

	//更换超级节点
	superID := this.NodeManager.GetSuperPeerId()
	if superID != nil && bytes.Equal([]byte(name), *superID) {
		newSuperID := this.NodeManager.FindNearInSuper(&this.NodeManager.NodeSelf.IdInfo.Id, nil, false, nil)
		this.NodeManager.SetSuperPeerId(newSuperID)
	}

	//判断是否仍然在线
	ses := this.SessionEngine.GetAllSession(this.AreaName[:])
	if len(ses) > 1 {
		// utils.Log.Debug().Msgf("Node offline 44444 self:%s target:%s index:%d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), ss.GetIndex())
		//在线
		return
	} else if this.GodHost != "" && len(ses) == 1 {
		return
	}

	//该节点没有邻居节点，已经离开了网络，没有连入网站中。
	utils.Log.Debug().Msgf("----------- Left the network ---------------")
	this.ResetAutonomyFinish()
	select {
	case <-this.isOnline:
	default:
	}
	//启动定时重连机制
	go this.reConnect(this.contextRoot)
	// utils.Log.Debug().Msgf("Node offline 55555 self:%s target:%s index:%d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), ss.GetIndex())
	return
}

/*
处理查找节点的请求
定期查询已知节点是否在线，更新节点信息
*/
func (this *Area) read(c context.Context) {
	var nodeIdStr *nodeStore.AddressNet
	for {
		select {
		case nodeIdStr = <-this.NodeManager.OutFindNode:
		case <-c.Done():
			// utils.Log.Info().Msgf("read done!")
			return
		}
		this.MessageCenter.SendSearchSuperMsg(gconfig.MSGID_checkNodeOnline, nodeIdStr, nil)
	}
}

/*
定时获得相邻节点的超级节点ip地址
*/
func (this *Area) getNearSuperIP(c context.Context) {
	timeOutM := make(map[string]int64)
	total := 0
	//用于判断首尾模式
	//当连续自治成功3次后，判断是否可以启动首尾模式
	headTailModlTimes := 0
	for {
		select {
		case <-c.Done():
			return
		case <-this.isOnline:
			select {
			case this.isOnline <- false:
			default:
			}
		}
		logicNodes := this.NodeManager.GetLogicNodes()
		clientNodes := this.NodeManager.GetNodesClient()
		nodesAll := append(logicNodes, clientNodes...)
		onebyoneNodes := this.NodeManager.GetOneByOneNodes()
		nodesAll = append(nodesAll, onebyoneNodes...)
		allUpDown := this.SessionEngine.GetAllDownUp(this.AreaName[:])
		for _, v := range allUpDown {
			nodesAll = append(nodesAll, nodeStore.AddressNet([]byte(v)))
		}

		if len(nodesAll) <= 0 {
			// 修复: 在只有一个服务器节点, 并且客户端连接了服务器后, 服务器尝试自治, 没有任何逻辑节点, 导致死循环占用CPU过高的问题
			time.Sleep(time.Second)
			continue
		}

		// 如果设置了超级代理节点，只要和超级代理节点建立了连接后，就认定网络自治完成
		var godAutonomyFinish bool
		if total < 2 && this.GodHost != "" {
			allSessions := this.SessionEngine.GetAllSession(this.AreaName[:])
			for i := range allSessions {
				if allSessions[i].GetRemoteHost() != this.GodHost {
					continue
				}

				godAutonomyFinish = true
				break
			}
		}

		if !godAutonomyFinish {
			//检查连接到本机的节点是否可以作为逻辑节点
			this.NodeManager.CheckClientNodeIsLogicNode()
			haveFail := false
			//防止所有节点都在小黑屋
			allWait := true
			for i := range nodesAll {
				select {
				case <-c.Done():
					return
				default:
				}

				//发送超时的节点，小黑屋冷静1分钟后再使用
				if t, ok := timeOutM[nodesAll[i].B58String()]; ok {
					if t+60 < time.Now().Unix() {
						delete(timeOutM, nodesAll[i].B58String())
					} else {
						continue
					}
				}

				bs, err := this.MessageCenter.SendNeighborMsgWaitRequest(gconfig.MSGID_getNearSuperIP, &nodesAll[i], nil, time.Second*8)
				if err != nil {
					//utils.Log.Warn().Msgf("send MSGID_getNearSuperIP error:%s self:%s target:%s ", err.Error(), this.GetNetId().B58String(), nodesAll[i].B58String())
					timeOutM[nodesAll[i].B58String()] = time.Now().Unix()
					haveFail = true
					continue
				}

				this.recvNearLogicNodes(bs, &nodesAll[i])
				allWait = false
			}
			//全部被关在小黑屋里时重新循环
			if allWait {
				for k, _ := range timeOutM {
					delete(timeOutM, k)
				}
				time.Sleep(time.Second)
				continue
			}

			if haveFail {
				total = 0
				headTailModlTimes = 0
				continue
			}

			//检查逻辑节点是否有变化，如果两次无变化，则停止寻找逻辑节点
			if this.NodeManager.EqualLogicNodes(logicNodes) {
				total = 0
				headTailModlTimes = 0
			} else {
				total++
			}
		}

		if total >= 2 || godAutonomyFinish {
			this.autonomyFinishChanLock.RLock()
			if this.destroy {
				this.autonomyFinishChanLock.RUnlock()
				return
			}
			select {
			case this.autonomyFinishChan <- false:
				this.sendOnlineMulticast()
			default:
			}
			this.autonomyFinishChanLock.RUnlock()

			this.saveLogicNodeIP()
			this.Vc.TriggerLoopGetVnodeinfo()

			//////查看本节点 Vnode 获取所有seesion中的位次排序
			/////
			// for i, _ := range this.Vm.VnodeMap {
			// 	downSession := this.Vm.VnodeMap[i].GetDownVnodeInfo()
			// 	upSession := this.Vm.VnodeMap[i].GetUpVnodeInfo()
			// 	aa := ""
			// 	for v := 0; v < len(upSession); v++ {
			// 		nn := fmt.Sprintf("\n upVNode : %s  \n", nodeStore.AddressNet([]byte(upSession[v].Vid)).B58String())
			// 		aa += nn
			// 	}
			// 	for v := 0; v < len(downSession); v++ {
			// 		nn := fmt.Sprintf("\n downVNode : %s  \n", nodeStore.AddressNet([]byte(downSession[v].Vid)).B58String())
			// 		aa += nn
			// 	}
			// 	utils.Log.Info().Msgf("22222排序位次 %d:\nziji :%s node:%s \n----> %s, ", i, this.Vm.VnodeMap[i].Vnode.Vid.B58String(), this.Vm.VnodeMap[i].Vnode.Nid.B58String(), aa)
			// }

			// {
			// 	////查看本节点获取所有seesion中的位次排序
			// 	downSession := this.SessionEngine.GetAllDownSession(this.AreaName[:])
			// 	upSession := this.SessionEngine.GetAllUpSession(this.AreaName[:])
			// 	///
			// 	aa := ""
			// 	for v := 0; v < len(upSession); v++ {
			// 		nn := fmt.Sprintf("\n upSession : %s  %d\n", nodeStore.AddressNet([]byte(upSession[v].GetName())).B58String(), upSession[v].GetIndex())
			// 		aa += nn
			// 	}
			// 	for v := 0; v < len(downSession); v++ {
			// 		nn := fmt.Sprintf("\n downSession : %s  %d\n", nodeStore.AddressNet([]byte(downSession[v].GetName())).B58String(), downSession[v].GetIndex())
			// 		aa += nn
			// 	}
			// 	utils.Log.Info().Msgf("44444排序位次 :\nziji :%s  \n----> %s, ", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), aa)
			// }
			// ////////////

			// // ////////////
			// {
			// 	LogicNode := this.NodeManager.GetLogicNodeInfo()
			// 	bb := ""
			// 	for v := 0; v < len(LogicNode); v++ {
			// 		nn := fmt.Sprintf("\n LogicNode : %s \n", nodeStore.AddressNet(LogicNode[v].IdInfo.Id.B58String()))
			// 		bb += nn
			// 	}
			// 	utils.Log.Info().Msgf("66666排序位次 :\nziji :%s  \n----> %s, ", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), bb)
			// }

			if this.findNearNodeTimer.Wait(c) == 0 {
				//对象销毁了，退出
				return
			}
			total = 1
			headTailModlTimes += 1

			// 连续两次节点没变化，判断去开首尾模式
			if headTailModlTimes >= 2 {
				downSession := this.SessionEngine.GetAllDownSession(this.AreaName[:])
				upSession := this.SessionEngine.GetAllUpSession(this.AreaName[:])
				if !this.NodeManager.IsHeadTailModl {
					if (len(upSession) < config.GreaterThanSelfMaxConn/2 || len(downSession) < config.GreaterThanSelfMaxConn/2) &&
						!(len(upSession) <= config.GreaterThanSelfMaxConn/2 && len(downSession) <= config.GreaterThanSelfMaxConn/2) {
						this.NodeManager.IsHeadTailModl = true
					}
				}
			}
			//utils.Log.Info().Msgf("确定节点是否开启了首尾模式 %t, %s", this.NodeManager.IsHeadTailModl, this.NodeManager.NodeSelf.IdInfo.Id.B58String())
		} else {
			this.findNearNodeTimer.Reset()
		}
	}
}

/*
保存超级节点ip地址到数据库做缓存
*/
func (this *Area) saveLogicNodeIP() {

	nodeInfos := this.NodeManager.GetLogicNodeInfo()
	ips := make([]string, 0, len(nodeInfos))
	for _, one := range nodeInfos {
		ips = append(ips, one.Addr+":"+strconv.Itoa(int(one.TcpPort)))
	}
	this.addrManager.SavePeerEntryToDB(ips, this.AreaName[:])
}

/*
	定时广播自己在线
*/
// func broadcastSelfOnline() {
// 	//TODO 应该只初始化一次
// 	bs, _ := json.Marshal(nodeStore.NodeSelf)
// 	message_center.SendMulticastMsg(gconfig.MSGID_multicast_online_recv, &bs)

// }

/*
通过事件驱动，给邻居节点推送对方需要的逻辑节点
*/
func (this *Area) SendNearLogicSuperIP(c context.Context) {
	var nodeOne *nodeStore.Node
	for {
		select {
		case nodeOne = <-this.NodeManager.HaveNewNode:
			// utils.Log.Info().Msgf("nodeinfo:%+v", nodeOne)
		case <-c.Done():
			// utils.Log.Info().Msgf("SendNearLogicSuperIP done!")
			return
		}
		// utils.Log.Info().Msgf("nodeinfo:%+v", nodeOne)
		// nodes := make([]nodeStore.Node, 0)
		// nodes = append(nodes, *nodeOne)
		// data, _ := json.Marshal(nodes)

		// utils.Log.Info().Msgf("SendNearLogicSuperIP :%s", nodeOne.IdInfo.Id.B58String())

		for _, session := range this.SessionEngine.GetAllSession(this.AreaName[:]) {
			sessionAddr := nodeStore.AddressNet([]byte(session.GetName()))
			ns := this.NodeManager.GetLogicNodes()
			ns = append(ns, this.NodeManager.GetNodesClient()...)
			// utils.Log.Info().Msgf("nodeinfo:%+v", nodeOne)
			ns = append(ns, nodeOne.IdInfo.Id)

			idsm := nodeStore.NewIds(sessionAddr, nodeStore.NodeIdLevel)
			for _, one := range ns {
				if bytes.Equal(sessionAddr, one) {
					continue
				}
				idsm.AddId(one)
			}
			ids := idsm.GetIds()

			nodes := make([]nodeStore.Node, 0)
			have := false //标记是否有这个新节点
			for _, one := range ids {
				if bytes.Equal(one, nodeOne.IdInfo.Id) {
					have = true
					nodes = append(nodes, *nodeOne)
					continue
				}
				addrNet := nodeStore.AddressNet(one)
				node := this.NodeManager.FindNode(&addrNet)
				if node != nil {
					nodes = append(nodes, *node)
				} else {
					// fmt.Println("这个节点为空")
				}
			}
			if !have {
				//没有新节点,则不发送推送消息
				continue
			}
			// data, _ := json.Marshal(nodes)
			nodeRepeated := go_protobuf.NodeRepeated{
				Nodes: make([]*go_protobuf.Node, 0),
			}
			for _, one := range nodes {
				if one.SetGod {
					continue
				}
				idinfo := go_protobuf.IdInfo{
					Id:   one.IdInfo.Id,
					EPuk: one.IdInfo.EPuk,
					CPuk: one.IdInfo.CPuk[:],
					V:    one.IdInfo.V,
					Sign: one.IdInfo.Sign,
				}

				nodeOne := go_protobuf.Node{
					IdInfo:       &idinfo,
					IsSuper:      one.GetIsSuper(),
					Addr:         one.Addr,
					TcpPort:      uint32(one.TcpPort),
					IsApp:        one.IsApp,
					MachineID:    one.MachineID,
					Version:      one.Version,
					MachineIDStr: one.MachineIDStr,
					QuicPort:     uint32(one.QuicPort),
				}

				nodeRepeated.Nodes = append(nodeRepeated.Nodes, &nodeOne)
			}

			//增加自己节点信息
			if this.GodHost == "" && this.NodeManager.NodeSelf.GetIsSuper() {
				idinfoSelf := go_protobuf.IdInfo{
					Id:   this.NodeManager.NodeSelf.IdInfo.Id,
					EPuk: this.NodeManager.NodeSelf.IdInfo.EPuk,
					CPuk: this.NodeManager.NodeSelf.IdInfo.CPuk[:],
					V:    this.NodeManager.NodeSelf.IdInfo.V,
					Sign: this.NodeManager.NodeSelf.IdInfo.Sign,
				}
				nodeRepeated.Nodes = append(nodeRepeated.Nodes, &go_protobuf.Node{
					IdInfo:       &idinfoSelf,
					IsSuper:      this.NodeManager.NodeSelf.GetIsSuper(),
					Addr:         this.NodeManager.NodeSelf.Addr,
					TcpPort:      uint32(this.NodeManager.NodeSelf.TcpPort),
					IsApp:        this.NodeManager.NodeSelf.IsApp,
					MachineID:    this.NodeManager.NodeSelf.MachineID,
					Version:      this.NodeManager.NodeSelf.Version,
					MachineIDStr: this.NodeManager.NodeSelf.MachineIDStr,
					QuicPort:     uint32(this.NodeManager.NodeSelf.QuicPort),
				})
			}
			data, _ := nodeRepeated.Marshal()

			// message_center.SendNeighborReplyMsg(message, config.MSGID_getNearSuperIP_recv, &data, msg.Session)
			this.MessageCenter.SendNeighborMsg(gconfig.MSGID_getNearSuperIP_recv, &sessionAddr, &data)

		}
	}

}

/*
读取需要询问关闭的连接名称
*/
func (this *Area) readOutCloseConnName(c context.Context) {
	var name *nodeStore.AddressNet
	for {
		select {
		case name = <-this.NodeManager.OutCloseConnName:
		case <-c.Done():
			// utils.Log.Info().Msgf("readOutCloseConnName done!")
			return
		}
		// message_center.AskCloseConn(name.B58String())
		// utils.Log.Info().Msgf("ReadOutCloseConnName:%s", name.B58String())
		this.MessageCenter.SendNeighborMsg(gconfig.MSGID_ask_close_conn_recv, name, nil)
	}

}

/*
定时删除数据库中过期的消息缓存
*/
func (this *Area) loopCleanMessageCache(c context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-c.Done():
			// utils.Log.Info().Msgf("loopCleanMessageCache done!")
			return
		}
		//计算24小时以前的时间UNIX
		// overtime := time.Now().Unix() - config.MsgCacheTimeOver
		// new(sqlite3_db.MessageCache).Remove(overtime)
	}

}

/*
 * 定时广播本节点的所有虚拟节点信息
 */
func (this *Area) loopSendVnodeInfo(c context.Context) {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-c.Done():
			// utils.Log.Info().Msgf("loopCleanMessageCache done!")
			return
		}
		this.Vc.TriggerLoopSendVnodeinfo()
	}

}

/*
断线重连
*/
func (this *Area) reConnect(c context.Context) {
	isRun := false
	this.reconnLock.Lock()
	if this.reConnTimer == nil {
		this.reConnTimer = utils.NewBackoffTimerChan(time.Second*1, time.Second*2, time.Second*4,
			time.Second*8, time.Second*16, time.Second*32, time.Second*64)
	} else {
		this.reConnTimer.Reset()
		this.reConnTimer.Release()
		// utils.Log.Info().Msgf("释放并重新设置时间")
	}
	if this.isReconn {
		isRun = true
	} else {
		this.isReconn = true
	}
	this.isReconn = true
	this.reconnLock.Unlock()
	if isRun {
		return
	}
	if this.destroy {
		return
	}
	// utils.Log.Info().Msgf("开始重连:%s", this.GetNetId().B58String())
	// this.reConnTimer := utils.NewBackoffTimerChan(time.Second*1, time.Second*2, time.Second*4,
	// 	time.Second*8, time.Second*16, time.Second*32, time.Second*64)
	for {
		if this.CheckOnline() {
			this.reconnLock.Lock()
			this.isReconn = false
			this.reconnLock.Unlock()
			return
		}
		if this.reConnTimer.Wait(c) == 0 {
			this.reconnLock.Lock()
			this.isReconn = false
			this.reconnLock.Unlock()
			return
		}
		if this.CheckOnline() {
			this.reconnLock.Lock()
			this.isReconn = false
			this.reconnLock.Unlock()
			return
		}
		ok := this.startUp()
		if ok {
			this.reconnLock.Lock()
			this.isReconn = false
			this.reconnLock.Unlock()
			// utils.Log.Info().Msgf("连接成功:%s", this.GetNetId().B58String())
			return
		}
		// utils.Log.Info().Msgf("连接失败:%s", this.GetNetId().B58String())
	}
}

/*
等待网络自治完成
*/
func (this *Area) WaitAutonomyFinish() {
	this.autonomyFinishChanLock.RLock()
	defer this.autonomyFinishChanLock.RUnlock()

	// 等待自治完成或area销毁信号
	select {
	case _, ok := <-this.autonomyFinishChan: // 自治完成
		if !ok {
			return
		}
	case <-this.contextRoot.Done(): // area销毁,在p2p网络启动未成功时，如果调用销毁,避免卡住Destory逻辑
		return
	}

	// 如果area已经被销毁了，则直接返回
	if this.destroy {
		return
	}

	// 触发自治完成信号
	select {
	case this.autonomyFinishChan <- false:
	default:
	}
	// utils.Log.Info().Msgf("WaitAutonomyFinish finish")
}

/*
重置网络自治接口
*/
func (this *Area) ResetAutonomyFinish() {
	this.autonomyFinishChanLock.RLock()
	select {
	case <-this.autonomyFinishChan:
	default:
	}
	this.autonomyFinishChanLock.RUnlock()
	// utils.Log.Info().Msgf("ResetAutonomyFinish reset")
}

/*
等待虚拟节点网络自治完成
*/
func (this *Area) WaitAutonomyFinishVnode() {
	this.Vm.WaitAutonomyFinish()
	// utils.Log.Info().Msgf("WaitAutonomyFinishVnode finish")
}

/*
注册系统自带消息
*/
func (this *Area) registerHandler() {
	this.MessageCenter.Register_search_super(config.MSGID_searchID, this.searchNetAddr, true)
	this.MessageCenter.Register_p2p(config.MSGID_searchID_recv, this.searchNetAddrRecv, true)
}

/*
接收磁力节点查询地址
*/
func (this *Area) searchNetAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// if msg.Session != nil {
	// 	utils.Log.Info().Msgf("搜到磁力消息self:%s  上级addr %s  Head.SearchNodeCount : %d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet(msg.Session.GetName()).B58String(), message.Head.SearchNodeCount)
	// } else {
	// 	utils.Log.Info().Msgf("搜到磁力消息self:%s  Head.SearchNodeCount : %d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), message.Head.SearchNodeCount)
	// }
	// utils.Log.Info().Msgf("搜到磁力消息self:%s  ", this.NodeManager.NodeSelf.IdInfo.Id.B58String())
	getNodeCount := 1
	if message.Body.Content != nil && len(*message.Body.Content) != 0 {
		getNodeCount = int(binary.LittleEndian.Uint16(*message.Body.Content))
	}
	ar := go_protobuf.NodeAddrRepeated{
		Nodes: make([]*go_protobuf.Addr, 0),
	}

	if message.Head.OneByOne && getNodeCount != 0 {
		ud := this.SessionEngine.GetAllDownSession(this.AreaName[:])
		ud = append(ud, this.SessionEngine.GetAllUpSession(this.AreaName[:])...)
		sorted, times := message_center.GetSortSessionForTarget(ud, this.NodeManager.NodeSelf.IdInfo.Id, *message.Head.RecvId)

		for i := 1; i < len(sorted)+1; i++ {
			//找upsession，由近及远
			if u := times - i; u >= 0 {
				ar.Nodes = append(ar.Nodes, &go_protobuf.Addr{
					Nid: sorted[u],
				})
			}
			// 找downsession，由近及远
			if d := times + i - 1; d < len(sorted) {
				ar.Nodes = append(ar.Nodes, &go_protobuf.Addr{
					Nid: sorted[d],
				})
			}
		}

		if !(getNodeCount <= 0 || int(getNodeCount) >= len(ar.Nodes)) {
			ar.Nodes = ar.Nodes[:getNodeCount]
		}
	}

	udb, err := ar.Marshal()
	if err != nil {
		utils.Log.Warn().Msgf("Marshal error : %s", err.Error())
		return
	}

	this.SendSearchSuperReplyMsg(message, config.MSGID_searchID_recv, &udb)
}

/*
接收磁力节点查询地址
*/
func (this *Area) searchNetAddrRecv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("搜到磁力消息返回self:%s", message.Head.Sender.B58String())
	var bs []byte
	if message.Body.Content != nil && len(*message.Body.Content) != 0 {
		bs = *message.Body.Content
	} else {
		if message.Head.Sender != nil && len(*message.Head.Sender) != 0 {
			bs = []byte(*message.Head.Sender)
		}
	}

	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), &bs)
}

/*
接收磁力节点查询地址
*/
func (this *Area) initDB(dbpath string) utils.ERROR {
	var err error
	this.levelDB, err = utilsleveldb.CreateLevelDB(dbpath)
	// 检查广播消息自增key信息是否存在
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("1111111111")
	// 检查信息是否存在
	newVersionKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_version)
	if !ERR.CheckSuccess() {
		return ERR
	}
	bsVer, err := this.levelDB.Find(*newVersionKey)
	if err != nil {
		// 关闭db连接信息
		this.levelDB.Close()
		this.levelDB = nil
		// 删除文件夹
		err = os.RemoveAll(dbpath)
		if err != nil {
			utils.Log.Warn().Msgf("删除p2pmessage出错:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		utils.Log.Warn().Msgf("删除数据库成功!!!!!")
		// 重新建立连接
		this.levelDB, err = utilsleveldb.CreateLevelDB(dbpath)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		utils.Log.Error().Msgf("新建数据库成功!!!!!!")
	}
	//utils.Log.Info().Msgf("1111111111")
	// 保存当前版本号
	verValue := utils.Uint64ToBytesByBigEndian(config.CUR_VERSION)
	ERR = this.levelDB.Save(*newVersionKey, &verValue)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if bsVer != nil && bsVer.Value != nil && len(bsVer.Value) > 0 {
		oldVersion := utils.BytesToUint64ByBigEndian(bsVer.Value)
		if oldVersion != config.CUR_VERSION {
			utils.Log.Warn().Msgf("当前数据版本号:%d, 更新到:%d", oldVersion, config.CUR_VERSION)
		} else {
			utils.Log.Warn().Msgf("当前数据版本号:%d", config.CUR_VERSION)
		}
	} else {
		utils.Log.Warn().Msgf("当前数据版本号:%d", config.CUR_VERSION)
	}
	//utils.Log.Info().Msgf("1111111111")
	// 确认广播自增数值
	newBroadcastAddKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_broadcast_msg_add)
	if !ERR.CheckSuccess() {
		return ERR
	}
	addValueKey, err := this.levelDB.Find(*newBroadcastAddKey)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if addValueKey != nil && len(addValueKey.Value) > 0 {
		utils.Log.Info().Msgf("value:%v", addValueKey.Value)
		config.CurBroadcastAddValue = utils.BytesToUint64ByBigEndian(addValueKey.Value)
	} else {
		value := utils.Uint64ToBytesByBigEndian(config.CurBroadcastAddValue)
		ERR = this.levelDB.Save(*newBroadcastAddKey, &value)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	//utils.Log.Info().Msgf("1111111111")
	// 确认广播最后清理数值
	{
		newBroadcastClearKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_broadcast_msg_clear)
		if !ERR.CheckSuccess() {
			return ERR
		}
		//utils.Log.Info().Msgf("1111111111")
		clearValueKey, err2 := this.levelDB.Find(*newBroadcastClearKey)
		if err2 != nil {
			return utils.NewErrorSysSelf(err2)
		}
		//utils.Log.Info().Msgf("1111111111")
		if clearValueKey != nil && clearValueKey.Value != nil && len(clearValueKey.Value) > 0 {
			//utils.Log.Info().Msgf("value:%v", addValueKey.Value)
			config.CurBroadcastClearValue = utils.BytesToUint64ByBigEndian(clearValueKey.Value)
		}
	}
	//utils.Log.Info().Msgf("1111111111")
	return utils.NewErrorSuccess()
}

/*
 * 发送节点上线广播信息
 */
func (this *Area) sendOnlineMulticast() {
	bs, err := this.NodeManager.NodeSelf.Proto()
	if err != nil {
		return
	}

	// utils.Log.Info().Msgf("发送节点在线广播消息: self:%s", addrNet.B58String())
	if err := this.SendMulticastMsg(gconfig.MSGID_multicast_online_recv, &bs); err != nil {
		// utils.Log.Info().Msgf("发送节点下线广播消息 err:%s", err)
	}
}

/*
 * 清理数据
 */
func (this *Area) cleanData(c context.Context) {
	// return

	ticker := time.NewTicker(config.CLEAN_DATA_TIME)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-c.Done():
			return
		}

		// 1. 判断是否初始化
		if this == nil || this.levelDB == nil {
			continue
		}

		// 2. 获取leveldb中广播自增值, 更新加1
		multiMsgAddValue := atomic.AddUint64(&config.CurBroadcastAddValue, 1)
		{
			value := utils.Uint64ToBytesByBigEndian(multiMsgAddValue)
			newBroadcastAddKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_broadcast_msg_add)
			if !ERR.CheckSuccess() {
				return
			}
			this.levelDB.Save(*newBroadcastAddKey, &value)
		}
		// utils.Log.Error().Msgf("add value:%d", multiMsgAddValue)
		if multiMsgAddValue < 3 {
			continue
		}

		// 3. 获取最后清理值信息
		var multiMsgClearValue = atomic.LoadUint64(&config.CurBroadcastClearValue)
		// utils.Log.Error().Msgf("clear value:%d", multiMsgClearValue)
		if multiMsgClearValue >= multiMsgAddValue-1 {
			multiMsgClearValue = multiMsgAddValue - 2
		}

		// 4. 清理leveldb中的广播信息
		for multiMsgClearValue < multiMsgAddValue-1 {
			keyOut := utils.Uint64ToBytesByBigEndian(multiMsgClearValue)
			newBroadcastKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_broadcast_msg)
			if !ERR.CheckSuccess() {
				return
			}
			newKeyOut, ERR := utilsleveldb.BuildLeveldbKey(keyOut)
			if !ERR.CheckSuccess() {
				return
			}
			this.levelDB.RemoveMapInMapByKeyOutInterval(*newBroadcastKey, *newKeyOut, config.MAX_CLEAN_DATA_LENGTH, config.CLEAN_DATA_INTERVAL_TIME)

			time.Sleep(time.Second * 1)

			multiMsgClearValue = atomic.AddUint64(&config.CurBroadcastClearValue, 1)
			keyOut2 := utils.Uint64ToBytesByBigEndian(multiMsgClearValue)
			newBroadcastClearKey, ERR := utilsleveldb.BuildLeveldbKey(config.DBKEY_broadcast_msg_clear)
			if !ERR.CheckSuccess() {
				return
			}
			this.levelDB.Save(*newBroadcastClearKey, &keyOut2)
		}
	}
}

/*
 * 检查网络自治是否已经完成
 */
func (this *Area) CheckAutonomyFinish() (autoFinish bool) {
	this.autonomyFinishChanLock.RLock()
	defer this.autonomyFinishChanLock.RUnlock()

	// 等待自治完成或area销毁信号
	select {
	case <-this.autonomyFinishChan: // 自治完成
		autoFinish = true
	case <-this.contextRoot.Done(): // area销毁,在p2p网络启动未成功时，如果调用销毁,避免卡住Destory逻辑
	default:
	}

	// 触发自治完成信号
	if autoFinish {
		select {
		case this.autonomyFinishChan <- false:
		default:
		}
	}

	return autoFinish
}
