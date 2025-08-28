package libp2parea

import (
	"context"
	"github.com/rs/zerolog"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/addr_manager"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/libp2parea/v2/session_manager"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

type Node struct {
	global                    *Global                         //域管理器
	AreaName                  [32]byte                        //域网络名称
	Keystore                  keystore.KeystoreInterface      //
	Pwd                       string                          //密码
	addrManager               *addr_manager.AddrManager       //外网节点端口管理
	NodeManager               *nodeStore.NodeManager          //
	SessionEngine             *engine.Engine                  //消息引擎
	SessionManager            *session_manager.SessionManager //
	MessageCenter             *message_center.MessageCenter   //
	contextRoot               context.Context                 //
	canceRoot                 context.CancelFunc              //
	GetNearSuperAddr_recvLock *sync.Mutex                     //
	isStartCore               bool                            //
	reconnLock                *sync.Mutex                     //断线重连锁
	isReconn                  bool                            //
	reConnTimer               *utils.BackoffTimerChan         //重连计时器
	findNearNodeTimer         *utils.BackoffTimerChan         //
	sqlite3dbPath             string                          //sqlite3数据库
	leveldbPath               string                          //leveldb数据库存放地方
	connecting                *sync.Map                       //保存正在连接的节点地址。key:string=[ip:port];
	lockNewConnCallback       *sync.RWMutex                   //
	levelDB                   *utilsleveldb.LevelDB           //
	GodHost                   string                          // 上帝节点地址
	GodID                     *nodeStore.AddressNet           // 上帝节点Id
	setGodHostLock            *sync.Mutex                     // 设置上帝节点地址锁
	closedCallbackFunc        []NodeEventCallbackHandler      // 节点关闭连接回调
	beenGodAddrCallbackFunc   []NodeEventCallbackHandler      // 节点被设置为超级代理节点地址回调
	newConnCallbackFunc       []NodeEventCallbackHandler      // 节点新建连接回调
	selfAddrs                 sync.Map                        // 自己地址列表
	connLimitType             int16                           // 连接方式限制, 默认支持所有的连接
	Log                       *zerolog.Logger                 //日志
}

/*
创建一个域网络
@areaName    [32]byte                      域名称
@key         keystore.KeystoreInterface    密钥接口
@pwd string                                密码
*/
func NewNode(areaName [32]byte, pre string, key keystore.KeystoreInterface, pwd string) (*Node, utils.ERROR) {
	log := utils.Log
	//(*log).Info().Msgf("日志指针:%p", log)
	//this.Log.Info().Str("NewArea", "11111111").Send()
	contextRoot, canceRoot := context.WithCancel(context.Background())
	g := NewGlobal()
	puk, _, ERR := key.GetNetAddrKeyPair(pwd)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//this.Log.Info().Str("NewArea", "11111111").Send()
	addrNet, ERR := nodeStore.BuildAddr(pre, puk)
	if ERR.CheckFail() {
		return nil, ERR
	}
	sessionEngine := engine.NewEngine(addrNet.B58String())
	nodeManager, ERR := nodeStore.NewNodeManager(areaName, pre, key, pwd, log)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//this.Log.Info().Str("NewArea", "11111111").Send()
	addrM := addr_manager.NewAddrManager()
	//this.Log.Info().Str("NewArea", "11111111").Send()
	mcm := message_center.NewMessageCenter(nodeManager, sessionEngine, key, pwd, contextRoot, areaName[:], log)
	//this.Log.Info().Str("NewArea", "11111111").Send()
	sm := session_manager.NewSessionManager(addrM, nodeManager, sessionEngine, log, contextRoot)
	//this.Log.Info().Str("NewArea", "11111111").Send()

	area := &Node{
		global:                    g,
		AreaName:                  areaName,
		Keystore:                  key,
		addrManager:               addrM,
		NodeManager:               nodeManager,
		SessionEngine:             sessionEngine,
		SessionManager:            sm,
		MessageCenter:             mcm,
		contextRoot:               contextRoot, //
		canceRoot:                 canceRoot,   //
		GetNearSuperAddr_recvLock: new(sync.Mutex),
		//isOnline:                  make(chan bool, 1),
		reconnLock: new(sync.Mutex),
		//destroy:                new(atomic.Bool),
		findNearNodeTimer: utils.NewBackoffTimerChan(config.FindNearNodeTimer...),
		//autonomyFinishChanLock: new(sync.RWMutex),
		//autonomyFinishChan:     make(chan bool, 1),
		connecting:          new(sync.Map),
		lockNewConnCallback: new(sync.RWMutex),
		setGodHostLock:      new(sync.Mutex),
		connLimitType:       config.CONN_TYPE_ALL,
		Log:                 log,
	}
	//nodeManager.Log = &area.Log
	sessionEngine.SetDialBeforeEvent(area.SessionManager.DialBeforeEvent)
	sessionEngine.SetDialAfterEvent(area.SessionManager.DialAfterEvent)
	sessionEngine.SetAcceptBeforeEvent(area.SessionManager.AcceptBeforeEvent)
	sessionEngine.SetAcceptAfterEvent(area.SessionManager.AcceptAfterEvent)
	sessionEngine.SetCloseConnBeforeEvent(area.SessionManager.CloseBeforeEvent)
	sessionEngine.SetCloseConnAfterEvent(area.SessionManager.CloseAfterEvent)
	//this.Log.Info().Str("NewArea", "11111111").Send()
	ERR = area.registerRPC()
	return area, ERR
}

/*
启动节点
@port    uint16    节点监听的端口
*/
func (this *Node) StartUP(port uint16) utils.ERROR {
	//this.destroy.Store(false)
	this.Log.Info().Msgf("Local netid is: %s", this.NodeManager.NodeSelf.IdInfo.Id.B58String())

	this.SessionManager.RegisterHanders()

	if this.leveldbPath == "" {
		this.leveldbPath = config.Path_leveldb
	}
	//this.Log.Info().Msgf("StartUP 1111111111")
	ERR := this.initDB(this.leveldbPath)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//this.Log.Info().Msgf("StartUP 1111111111")
	go this.cleanData(this.contextRoot)
	//初始化数据库
	//this.Log.Info().Msgf("StartUP 1111111111")
	this.MessageCenter.SetLevelDB(this.levelDB)
	this.addrManager.SetLevelDB(this.levelDB)
	//utils.Log.Info().Uint16("打印端口", port).Send()
	this.NodeManager.NodeSelf.Port = port
	this.startEngine(port)
	//this.Log.Info().Msgf("StartUP 1111111111")
	this.SessionManager.StartConnectNet()
	//this.Log.Info().Msgf("StartUP 1111111111")
	return utils.NewErrorSuccess()
}

/*
创建一个相互隔离的域网络
@areaName    [32]byte    新的域名称
*/
func (this *Node) With(areaName [32]byte, addr string, port uint16) (*Node, utils.ERROR) {
	area := &Node{}
	return area, utils.NewErrorSuccess()
}

/*
启动消息服务器
@isFirst    bool      是否是首节点
@addr       string    监听地址
*/
func (this *Node) startEngine(port uint16) utils.ERROR {
	//auth := NewAuth(this.AreaName, this.NodeManager, this.SessionEngine, this.Vc)
	//this.SessionEngine.AddAuth(auth)
	//this.SessionEngine.SetCloseCallback(this.AreaName[:], this.closeConnCallback)
	//this.SessionEngine.SetClientConnCallback(this.AreaName[:], this.clientNewConnCallback)
	//this.SessionEngine.SetServerConnCallback(this.AreaName[:], this.serverNewConnCallback)
	//this.RegisterCoreMsg()
	this.MessageCenter.RegisterMsgVersion()
	ERR := this.SessionEngine.ListenOnePort(port, true)
	if ERR.CheckFail() {
		return ERR
	}
	return utils.NewErrorSuccess()
}

/*
 * 主动移除服务端节点旧session，保持每个machineid只有一个session连接
 */
//func (this Area) removeOldConn(ss engine.Session, params interface{}) {
//	node, ok := params.(*nodeStore.Node)
//	if !ok {
//		return
//	}
//	existNode := this.NodeManager.FindNode(&node.IdInfo.Id)
//	if existNode != nil {
//		sessions := existNode.GetSessions()
//		for i := range sessions {
//			//this.Log.Info().Msgf("serverNewConnCallback 11111 self:%s %s", this.GetNetId().B58String(), node.IdInfo.Id.B58String())
//			if sessions[i].GetMachineID() != "" && sessions[i].GetMachineID() == ss.GetMachineID() {
//				sessions[i].Close()
//				existNode.RemoveSession(sessions[i])
//				this.SessionEngine.RemoveCustomSession(sessions[i])
//				this.SessionEngine.RemoveSession(this.AreaName[:], sessions[i])
//			}
//		}
//	}
//}

/*
关闭服务器回调函数
*/
func (this *Node) ShutdownCallback() {
	//回收映射的端口
	//Reclaim()
	// addrm.CloseBroadcastServer()
	// fmt.Println("Close over")
}

/*
等待网络自治完成
*/
func (this *Node) WaitAutonomyFinish() {
	this.SessionManager.WaitAutonomyFinish()
}

/*
重置网络自治接口
*/
func (this *Node) ResetAutonomyFinish() {
	this.SessionManager.ResetAutonomyFinish()
}

/*
 * 检查网络自治是否已经完成
 */
func (this *Node) CheckAutonomyFinish() (autoFinish bool) {
	return this.SessionManager.CheckAutonomyFinish()
}

/*
接收磁力节点查询地址
*/
func (this *Node) initDB(dbpath string) utils.ERROR {
	var err error
	this.levelDB, err = utilsleveldb.CreateLevelDB(dbpath)
	// 检查广播消息自增key信息是否存在
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//this.Log.Info().Msgf("1111111111")
	// 检查信息是否存在
	//newVersionKey, ERR := utilsleveldb.NewLeveldbKey(config.DBKEY_version)
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	bsVer, err := this.levelDB.Find(*config.DBKEY_version)
	if err != nil {
		// 关闭db连接信息
		this.levelDB.Close()
		this.levelDB = nil
		// 删除文件夹
		err = os.RemoveAll(dbpath)
		if err != nil {
			this.Log.Warn().Msgf("删除p2pmessage出错:%s", err.Error())
			return utils.NewErrorSysSelf(err)
		}
		this.Log.Warn().Msgf("删除数据库成功!!!!!")
		// 重新建立连接
		this.levelDB, err = utilsleveldb.CreateLevelDB(dbpath)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		this.Log.Error().Msgf("新建数据库成功!!!!!!")
	}
	//this.Log.Info().Msgf("1111111111")
	// 保存当前版本号
	verValue := utils.Uint64ToBytesByBigEndian(config.CUR_VERSION)
	ERR := this.levelDB.Save(*config.DBKEY_version, &verValue)
	if !ERR.CheckSuccess() {
		return ERR
	}
	if bsVer != nil && bsVer.Value != nil && len(bsVer.Value) > 0 {
		oldVersion := utils.BytesToUint64ByBigEndian(bsVer.Value)
		if oldVersion != config.CUR_VERSION {
			this.Log.Warn().Msgf("当前数据版本号:%d, 更新到:%d", oldVersion, config.CUR_VERSION)
		} else {
			this.Log.Warn().Msgf("当前数据版本号:%d", config.CUR_VERSION)
		}
	} else {
		this.Log.Warn().Msgf("当前数据版本号:%d", config.CUR_VERSION)
	}
	//this.Log.Info().Msgf("1111111111")
	// 确认广播自增数值
	//newBroadcastAddKey, ERR := utilsleveldb.NewLeveldbKey(config.DBKEY_broadcast_msg_add)
	//if !ERR.CheckSuccess() {
	//	return ERR
	//}
	addValueKey, err := this.levelDB.Find(*config.DBKEY_broadcast_msg_add)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if addValueKey != nil && len(addValueKey.Value) > 0 {
		this.Log.Info().Msgf("value:%v", addValueKey.Value)
		config.CurBroadcastAddValue = utils.BytesToUint64ByBigEndian(addValueKey.Value)
	} else {
		value := utils.Uint64ToBytesByBigEndian(config.CurBroadcastAddValue)
		ERR = this.levelDB.Save(*config.DBKEY_broadcast_msg_add, &value)
		if !ERR.CheckSuccess() {
			return ERR
		}
	}
	//this.Log.Info().Msgf("1111111111")
	// 确认广播最后清理数值
	{
		//newBroadcastClearKey, ERR := utilsleveldb.NewLeveldbKey(config.DBKEY_broadcast_msg_clear)
		//if !ERR.CheckSuccess() {
		//	return ERR
		//}
		//this.Log.Info().Msgf("1111111111")
		clearValueKey, err2 := this.levelDB.Find(*config.DBKEY_broadcast_msg_clear)
		if err2 != nil {
			return utils.NewErrorSysSelf(err2)
		}
		//this.Log.Info().Msgf("1111111111")
		if clearValueKey != nil && clearValueKey.Value != nil && len(clearValueKey.Value) > 0 {
			//this.Log.Info().Msgf("value:%v", addValueKey.Value)
			config.CurBroadcastClearValue = utils.BytesToUint64ByBigEndian(clearValueKey.Value)
		}
	}
	//this.Log.Info().Msgf("1111111111")
	return utils.NewErrorSuccess()
}

/*
 * 清理数据
 */
func (this *Node) cleanData(c context.Context) {
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
			//newBroadcastAddKey, ERR := utilsleveldb.NewLeveldbKey(config.DBKEY_broadcast_msg_add)
			//if !ERR.CheckSuccess() {
			//	return
			//}
			this.levelDB.Save(*config.DBKEY_broadcast_msg_add, &value)
		}
		// this.Log.Error().Msgf("add value:%d", multiMsgAddValue)
		if multiMsgAddValue < 3 {
			continue
		}

		// 3. 获取最后清理值信息
		var multiMsgClearValue = atomic.LoadUint64(&config.CurBroadcastClearValue)
		// this.Log.Error().Msgf("clear value:%d", multiMsgClearValue)
		if multiMsgClearValue >= multiMsgAddValue-1 {
			multiMsgClearValue = multiMsgAddValue - 2
		}

		// 4. 清理leveldb中的广播信息
		for multiMsgClearValue < multiMsgAddValue-1 {
			keyOut := utils.Uint64ToBytesByBigEndian(multiMsgClearValue)
			//newBroadcastKey, ERR := utilsleveldb.NewLeveldbKey(config.DBKEY_broadcast_msg)
			//if !ERR.CheckSuccess() {
			//	return
			//}
			newKeyOut, ERR := utilsleveldb.BuildLeveldbKey(keyOut)
			if !ERR.CheckSuccess() {
				return
			}
			this.levelDB.RemoveMapInMapByKeyOutInterval(*config.DBKEY_broadcast_msg, *newKeyOut, config.MAX_CLEAN_DATA_LENGTH,
				config.CLEAN_DATA_INTERVAL_TIME)

			time.Sleep(time.Second * 1)

			multiMsgClearValue = atomic.AddUint64(&config.CurBroadcastClearValue, 1)
			keyOut2 := utils.Uint64ToBytesByBigEndian(multiMsgClearValue)
			//newBroadcastClearKey, ERR := utilsleveldb.NewLeveldbKey(config.DBKEY_broadcast_msg_clear)
			//if !ERR.CheckSuccess() {
			//	return
			//}
			this.levelDB.Save(*config.DBKEY_broadcast_msg_clear, &keyOut2)
		}
	}
}

/*
设置日志
*/
func (this *Node) SetLog(log *zerolog.Logger) {
	//log.Info().Msgf("修改了日志 指针:%p", log)
	this.Log = log
	this.NodeManager.SetLog(log)
	this.SessionEngine.SetLog(this.Log)
	this.SessionManager.Log = this.Log
	this.MessageCenter.Log = this.Log
}

func (this *Node) registerRPC() utils.ERROR {
	return this.SessionEngine.RegisterRPC(1, config.RPC_method_netinfo, this.getinfo_h, "Network information")
}

func (this *Node) getinfo_h(params *map[string]interface{}) engine.PostResult {
	pr := engine.NewPostResult()
	pr.Data["address"] = this.NodeManager.NodeSelf.IdInfo.Id.B58String()

	addrs := make([]string, 0)
	for _, one := range this.GetNetworkInfo() {
		addrs = append(addrs, one.IdInfo.Id.B58String())
	}
	pr.Data["connects"] = addrs
	//pr.Data["AreaName"] =
	return *pr
}
