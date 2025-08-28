package session_manager

import (
	"bytes"
	"context"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/libp2parea/v2/addr_manager"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/message_center"
	"web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

type SessionManager struct {
	addrManager             *addr_manager.AddrManager     //外网节点端口管理
	nodeManager             *nodeStore.NodeManager        //
	sessionEngine           *engine.Engine                //
	messageCenter           *message_center.MessageCenter //
	ConnStatus              *ConnStatus                   //连接状态管理，连接去重
	guestSession            *sync.Map                     //连接上，未注册用户。key:string=sessionId;value:engine.Session=;
	contextRoot             context.Context               //
	canceRoot               context.CancelFunc            //
	isReconn                atomic.Bool                   //是否正在运行重连
	reConnTimer             *utils.BackoffTimerChan       //重连计时器
	destroy                 *atomic.Bool                  //销毁此对象,销毁了不能重连网络
	autonomyFinishState     *NodeAutonomyState            //节点自治状态
	HaveChengeNodeAreaSelf  chan *nodeStore.NodeInfo      //自己域网络有节点变动(节点上线或下线)
	HaveChengeNodeAreaOther chan *nodeStore.NodeInfo      //其他域网络有节点变动(节点上线或下线)
	OutCloseConnName        chan engine.Session           //询问节点关闭
	Log                     *zerolog.Logger               //日志
	CloseConnLock           *sync.Mutex                   //关闭连接锁
	mam                     *MulticastAddressManager      //局域网节点地址广播与接收
}

func NewSessionManager(addrM *addr_manager.AddrManager, nodeManager *nodeStore.NodeManager, sessionEngine *engine.Engine,
	log *zerolog.Logger, ctx context.Context) *SessionManager {
	ctxRoot, cancelRoot := context.WithCancel(ctx)
	mam := NewMulticastAddressManager(nodeManager, ctxRoot)
	sm := &SessionManager{
		addrManager:   addrM,
		nodeManager:   nodeManager,
		sessionEngine: sessionEngine,
		//messageCenter: messageCenter,
		guestSession: new(sync.Map),
		reConnTimer: utils.NewBackoffTimerChan(time.Second*1, time.Second*2, time.Second*4,
			time.Second*8, time.Second*16, time.Second*32, time.Second*64),
		destroy:                 new(atomic.Bool),
		HaveChengeNodeAreaSelf:  make(chan *nodeStore.NodeInfo, 1),
		HaveChengeNodeAreaOther: make(chan *nodeStore.NodeInfo, 1),
		OutCloseConnName:        make(chan engine.Session, 100),
		Log:                     log,
		CloseConnLock:           new(sync.Mutex),
		mam:                     mam,
	}
	mam.sessionManager = sm
	sm.ConnStatus = NewConnStatus(&sm.Log)
	sm.autonomyFinishState = NewNodeAutonomyState(sm)
	sm.contextRoot, sm.canceRoot = ctxRoot, cancelRoot
	sm.loopCleanSession() //循环清理连上超时未注册的节点
	//sm.loopClassificationSession() //对注册的节点分类存放
	sm.loop_getNearSuperIP()   //定时获得相邻节点的超级节点ip地址
	sm.loop_getAreaNet()       //
	sm.readOutCloseConnName()  //读取需要询问关闭的连接名称
	sm.loopCleanMessageCache() //定时删除数据库中过期的消息缓存
	sessionEngine.SetAuthFun(sm.authFun)
	//sm.RegisterHanders()
	return sm
}

func (this *SessionManager) AddNode(node *nodeStore.NodeInfo) {

}

/*
查询一个网络地址下有多少个协议端口
*/
//func (this *SessionManager) FindMultiaddr(addr []nodeStore.AddressNet) []string {
//	this.lock.RLock()
//	defer this.lock.RUnlock()
//	as := make([]string, 0, len(addr))
//	for _, addr := range addr {
//		ss, ok := this.ss[utils.Bytes2string(addr)]
//		if ok {
//			as = append(as, ss.Multiaddr)
//		}
//	}
//	return as
//}

/*
通过节点地址，查询多个会话，包括游离的会话
*/
func (this *SessionManager) FindSessionAreaSelfByAddr(addr *nodeStore.AddressNet) []engine.Session {
	nodeInfo := this.nodeManager.FindNodeInfoLANByAddr(addr)
	if nodeInfo != nil {
		return nodeInfo.GetSessions()
	}
	nodeInfo = this.nodeManager.FindNodeInfoWANByAddr(addr)
	if nodeInfo != nil {
		return nodeInfo.GetSessions()
	}
	nodeInfo = this.nodeManager.FindNodeInfoWhiteByAddr(addr)
	if nodeInfo != nil {
		return nodeInfo.GetSessions()
	}
	nodeInfo = this.nodeManager.FindNodeInfoFreeByAddr(addr)
	if nodeInfo != nil {
		return nodeInfo.GetSessions()
	}
	nodeInfo = this.nodeManager.FindNodeInfoProxyByAddr(addr)
	if nodeInfo != nil {
		return nodeInfo.GetSessions()
	}
	return nil
}

/*
获取自己域的所有会话
*/
//func (this *SessionManager) GetSessionAllAreaSelf() []engine.Session {
//	this.lock.RLock()
//	defer this.lock.RUnlock()
//	ss := make([]engine.Session, 0, len(this.sessionAreaSelfWAN))
//	for _, ssMap := range this.sessionAreaSelfWAN {
//		for _, one := range *ssMap {
//			ss = append(ss, one)
//		}
//	}
//	return ss
//}

/*
获取白名单中的所有会话
*/
func (this *SessionManager) GetWhiteListSession() []engine.Session {
	nodeInfos := this.nodeManager.GetWhiltListNodes()
	sss := make([]engine.Session, 0, len(nodeInfos))
	for _, one := range nodeInfos {
		for _, one := range one.Sessions {
			sss = append(sss, one)
		}
	}
	return sss
}

/*
获取超级节点中的所有会话
*/
func (this *SessionManager) GetAllNodeSessionWAN() []engine.Session {
	nodeInfos := this.nodeManager.GetLogicNodesWAN()
	sss := make([]engine.Session, 0, len(nodeInfos))
	for _, one := range nodeInfos {
		for _, one := range one.Sessions {
			sss = append(sss, one)
		}
	}
	return sss
}

/*
获取内网节点中的所有会话
*/
func (this *SessionManager) GetAllNodeSessionLAN() []engine.Session {
	nodeInfos := this.nodeManager.GetLogicNodesLAN()
	sss := make([]engine.Session, 0, len(nodeInfos))
	for _, one := range nodeInfos {
		for _, one := range one.Sessions {
			sss = append(sss, one)
		}
	}
	return sss
}

/*
获取被代理节点中的所有会话
*/
func (this *SessionManager) GetFreeNodeSession() []engine.Session {
	nodeInfos := this.nodeManager.GetFreeNodes()
	sss := make([]engine.Session, 0, len(nodeInfos))
	for _, one := range nodeInfos {
		for _, one := range one.Sessions {
			sss = append(sss, one)
		}
	}
	return sss
}

/*
获取被代理节点中的所有会话
*/
func (this *SessionManager) GetProxyNodeSession() []engine.Session {
	nodeInfos := this.nodeManager.GetProxyNodes()
	sss := make([]engine.Session, 0, len(nodeInfos))
	for _, one := range nodeInfos {
		for _, one := range one.Sessions {
			sss = append(sss, one)
		}
	}
	return sss
}

/*
通过sessionID查询节点地址
*/
//func (this *SessionManager) FindAddrBySessionID(sid []byte) *nodeStore.AddressNet {
//	session := this.sessionEngine.GetSession(sid)
//	if session == nil {
//		return nil
//	}
//	nodeInfo := GetNodeInfo(session)
//	if nodeInfo == nil {
//		return nil
//	}
//	return &nodeInfo.IdInfo.Id
//}

//func (this *SessionManager) RemoveNode(node *nodeStore.Node) {}

/*
循环清理连上超时未注册的节点
*/
func (this *SessionManager) loopCleanSession() {
	go func() {
		tiker := time.NewTicker(config.Not_reg_timeout)
		for range tiker.C {
			//this.CloseConnLock.Lock()
			//defer this.CloseConnLock.Unlock()
			now := time.Now().Unix()
			//for _, ss := range this.guestSession {
			//	//ss := v.(engine.Session)
			//	createTime := GetSessionCreateTime(ss)
			//	if now-createTime > int64(config.Not_reg_timeout/time.Second) {
			//		this.Log.Info().Str("关闭连接，因超时未注册", "").Str(ss.GetLocalHost(), ss.GetRemoteHost()).Send()
			//		//超时关闭
			//		ss.Close()
			//	}
			//}
			this.guestSession.Range(func(k, v interface{}) bool {
				ss := v.(engine.Session)
				createTime := GetSessionCreateTime(ss)
				if now-createTime > int64(config.Not_reg_timeout/time.Second) {
					this.Log.Info().Interface("会话", ss).Hex("sid", ss.GetId()).Uint8("连接类型",
						ss.GetConnType()).Str("关闭连接，因超时未注册", "").Str(ss.GetLocalHost(), ss.GetRemoteHost()).Send()
					//超时关闭
					ss.Close()
					this.guestSession.Delete(utils.Bytes2string(ss.GetId()))
				}
				return true
			})
			//this.registSession.Range(func(k, v interface{}) bool {
			//	ss := v.(engine.Session)
			//	createTime := GetSessionCreateTime(ss)
			//	if now-createTime > int64(config.Not_reg_timeout/time.Second) {
			//		this.Log.Info().Str("关闭连接，因超时未分类", "").Str(ss.GetLocalHost(), ss.GetRemoteHost()).Send()
			//		//超时关闭
			//		ss.Close()
			//	}
			//	return true
			//})
		}
	}()
}

/*
权限验证
*/
func (this *SessionManager) authFun(msg *engine.Packet) utils.ERROR {
	v := msg.Session.Get(config.Session_attribute_key_auth)
	if v != nil {

	}
	if msg.MsgID == config.MSGID_register_node {

	}
	return utils.NewErrorSuccess()
}

/*
定时获得相邻节点的超级节点ip地址
*/
func (this *SessionManager) loop_getNearSuperIP() {
	go func() {
		var newNode *nodeStore.NodeInfo
		for {
			select {
			case <-this.contextRoot.Done():
				return
			case newNode = <-this.HaveChengeNodeAreaSelf:
				//this.Log.Info().Str("有节点变动", "11111").Send()
			}
			if newNode != nil {
				//有节点上线
				//通过事件驱动，给邻居节点推送对方需要的逻辑节点
				this.SendNearLogicSuperIP(newNode)
			} else {
				//有节点下线
			}
			this.searchArea()
			this.getNearSuperIP()
		}
	}()
}

/*
定时获得相邻节点的超级节点ip地址
*/
func (this *SessionManager) getNearSuperIP() {
	this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
	total := 0
	for {
		//重新整理连接到本机的节点是否可以作为逻辑节点
		this.reorganizeSuperNodeIsLogicNode()
		//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
		//保存旧有逻辑节点，待对比是否有改变
		logicNodes := this.nodeManager.GetLogicNodesWAN()
		lanNodes := this.nodeManager.GetLogicNodesLAN()
		//询问所有的连接，自己的逻辑地址有哪些
		ss := this.sessionEngine.GetSessionAll()
		var ERRout utils.ERROR
		for _, sOne := range ss {
			//连接上，但还未注册的连接，不询问
			nodeInfo := GetNodeInfo(sOne)
			if nodeInfo == nil {
				//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
				continue
			}
			//不同域的连接，不询问
			if !bytes.Equal(nodeInfo.AreaName, this.nodeManager.NodeSelf.AreaName) {
				//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
				continue
			}
			//从邻居节点获取其他节点
			nodesWAN, nodesLAN, ERR := this.getNearSuperIP_net(sOne)
			if ERR.CheckFail() {
				//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
				ERRout = ERR
				continue
			}
			for _, one := range nodesWAN {
				this.Log.Info().Str("接收到的逻辑节点WAN", one.IdInfo.Id.B58String()).Send()
			}
			for _, one := range nodesLAN {
				this.Log.Info().Str("接收到的逻辑节点LAN", one.IdInfo.Id.B58String()).Send()
			}
			//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
			for _, one := range nodesWAN {
				//检查这个节点是否需要
				need, _, exist := this.nodeManager.CheckNeedNodeWan(&one, false)
				//不重复并且自己需要的节点，才去连接
				if !exist && need {
					for _, one := range one.GetMultiaddrWAN() {
						//this.Log.Info().Str("开始搜索逻辑节点 开始连接节点", "").Send()
						this.syncConnectNetOne(one.Multiaddr, nil, nil)
					}
				}
			}
			//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
			for _, one := range nodesLAN {
				//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
				//检查这个节点是否需要
				need, _, exist := this.nodeManager.CheckNeedNodeLan(&one, false)
				//不重复并且自己需要的节点，才去连接
				if !exist && need {
					for _, one := range one.GetMultiaddrLAN() {
						//this.Log.Info().Str("开始搜索逻辑节点 开始连接节点", "").Send()
						this.syncConnectNetOne(one.Multiaddr, nil, nil)
					}
				}
				//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
			}
		}
		//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
		//如果存在有的节点发送超时，则重新发送
		if ERRout.CheckFail() {
			//this.Log.Info().Str("搜索节点错误", ERRout.String()).Send()
			//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
			continue
		}
		//this.Log.Info().Str("开始搜索逻辑节点", "开始").Send()
		//检查逻辑节点是否有变化，如果两次无变化，则停止寻找逻辑节点
		if this.nodeManager.EqualLogicNodesWAN(logicNodes) || this.nodeManager.EqualLogicNodesLAN(lanNodes) {
			//this.Log.Info().Str("开始搜索逻辑节点", "有变化").Send()
			total = 0
		} else {
			//this.Log.Info().Str("开始搜索逻辑节点", "没变化").Send()
			total++
		}
		if total > 2 {
			break
		}
	}
	//this.Log.Info().Str("开始搜索逻辑节点", "结束").Send()
	this.autonomyFinishState.SetAutonomyFinish()
}

/*
定时获得其他域网络的节点
*/
func (this *SessionManager) loop_getAreaNet() {
	go func() {
		var newNode *nodeStore.NodeInfo
		for {
			select {
			case <-this.contextRoot.Done():
				return
			case newNode = <-this.HaveChengeNodeAreaOther:
				//this.Log.Info().Str("域网络有变动", "11111").Send()
			}
			if newNode != nil {
				//有节点上线
				//通过事件驱动，给邻居节点推送对方需要的逻辑节点
				this.SendOtherAreaNodeSuperIP(newNode)
			} else {
				//有节点下线
			}
			this.searchArea()
			this.getAreaNames()
		}
	}()
}

/*
查找逻辑域节点
*/
func (this *SessionManager) getAreaNames() {
	//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
	total := 0
	for {
		//重新整理连接到本机的节点是否可以作为逻辑节点
		//this.reorganizeSuperNodeIsLogicNode()
		//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
		//保存旧有逻辑节点，待对比是否有改变
		oldAddrsWAN := this.nodeManager.AreaLogicAddrWAN.GetNodeInfos()
		oldAddrsLAN := this.nodeManager.AreaLogicAddrLAN.GetNodeInfos()
		//logicNodes := this.nodeManager.GetLogicNodesWAN()
		//lanNodes := this.nodeManager.GetLogicNodesLAN()
		//询问所有的连接，自己的逻辑地址有哪些
		ss := this.sessionEngine.GetSessionAll()
		var ERRout utils.ERROR
		for _, sOne := range ss {
			//连接上，但还未注册的连接，不询问
			nodeInfo := GetNodeInfo(sOne)
			if nodeInfo == nil {
				//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
				continue
			}
			//不同域的连接，不询问
			//if bytes.Equal(nodeInfo.AreaName, this.nodeManager.NodeSelf.AreaName) {
			//	//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
			//	continue
			//}
			//从邻居节点获取其他域节点
			nodesWAN, nodesLAN, ERR := this.getLogicAreaName_net(sOne)
			if ERR.CheckFail() {
				//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
				ERRout = ERR
				continue
			}
			//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
			for _, one := range nodesWAN {
				//检查这个节点是否需要
				need, _, exist := this.nodeManager.CheckNeedNodeWan(&one, false)
				//不重复并且自己需要的节点，才去连接
				if !exist && need {
					for _, one := range one.GetMultiaddrWAN() {
						//this.Log.Info().Str("开始搜索逻辑节点 开始连接节点", "").Send()
						this.syncConnectNetOne(one.Multiaddr, nil, nil)
					}
				}
			}
			//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
			for _, one := range nodesLAN {
				//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
				//检查这个节点是否需要
				need, _, exist := this.nodeManager.CheckNeedNodeLan(&one, false)
				//不重复并且自己需要的节点，才去连接
				if !exist && need {
					for _, one := range one.GetMultiaddrLAN() {
						//this.Log.Info().Str("开始搜索逻辑节点 开始连接节点", "").Send()
						this.syncConnectNetOne(one.Multiaddr, nil, nil)
					}
				}
				//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
			}
		}
		//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
		//如果存在有的节点发送超时，则重新发送
		if ERRout.CheckFail() {
			//this.Log.Info().Str("搜索节点错误", ERRout.String()).Send()
			//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
			continue
		}
		//this.Log.Info().Str("开始搜索逻辑域", "开始").Send()
		//检查逻辑节点是否有变化，如果两次无变化，则停止寻找逻辑节点
		if this.nodeManager.AreaLogicAddrWAN.EqualLogicNodes(oldAddrsWAN) || this.nodeManager.AreaLogicAddrLAN.EqualLogicNodes(oldAddrsLAN) {
			//}
			//if this.nodeManager.EqualLogicNodesWAN(logicNodes) || this.nodeManager.EqualLogicNodesLAN(lanNodes) {
			//this.Log.Info().Str("开始搜索逻辑域", "有变化").Send()
			total = 0
		} else {
			//this.Log.Info().Str("开始搜索逻辑域", "没变化").Send()
			total++
		}
		if total > 2 {
			break
		}
	}
	//this.Log.Info().Str("开始搜索逻辑域", "结束").Send()
	//if this.destroy.Load() {
	//	return
	//}
	//select {
	//case this.autonomyFinishChan <- false:
	//	this.Log.Info().Str("开始搜索逻辑节点", "结束").Send()
	//default:
	//	//this.Log.Info().Str("开始搜索逻辑节点", "结束").Send()
	//}
	return
}

/*
检查连接到本机的节点是否可以作为逻辑节点
*/
func (this *SessionManager) reorganizeSuperNodeIsLogicNode() {
	//this.superSession.Range(func(key, value any) bool {
	//	session := value.(engine.Session)
	//	addr := this.FindAddrBySessionID(session.GetId())
	//	if addr == nil {
	//		this.Log.Error().Str("超级节点连接session中没有节点信息", "").Send()
	//		return true
	//	}
	//	this.nodeManager.NodeLogicAddrWAN.CheckNeedAddr(*addr, true)
	//	return true
	//})
}

/*
通过事件驱动，给邻居节点推送对方需要的逻辑节点
*/
func (this *SessionManager) SendNearLogicSuperIP(newNode *nodeStore.NodeInfo) {
	sss := this.sessionEngine.GetSessionAll()
	//this.Log.Info().Int("激活推送新节点程序", len(sss)).Send()
	for _, ssOne := range sss {
		nodeOne := GetNodeInfo(ssOne)
		if nodeOne == nil {
			//this.Log.Info().Str("激活推送新节点程序", "").Send()
			continue
		}
		//不再往回推送，否则死循环
		if bytes.Equal(nodeOne.AreaName, newNode.AreaName) &&
			bytes.Equal(nodeOne.IdInfo.Id.Data(), newNode.IdInfo.Id.Data()) &&
			bytes.Equal(nodeOne.MachineID, newNode.MachineID) {
			//this.Log.Info().Str("激活推送新节点程序", "").Send()
			continue
		}
		this.sendNewNode_net(ssOne, newNode)
	}
}

/*
通过事件驱动，给邻居节点推送其他域节点信息
*/
func (this *SessionManager) SendOtherAreaNodeSuperIP(newNode *nodeStore.NodeInfo) {
	sss := this.sessionEngine.GetSessionAll()
	//this.Log.Info().Int("激活推送新节点程序", len(sss)).Send()
	for _, ssOne := range sss {
		nodeOne := GetNodeInfo(ssOne)
		if nodeOne == nil {
			//this.Log.Info().Str("激活推送新节点程序", "").Send()
			continue
		}
		//不再往回推送，否则死循环
		if bytes.Equal(nodeOne.AreaName, newNode.AreaName) &&
			bytes.Equal(nodeOne.IdInfo.Id.Data(), newNode.IdInfo.Id.Data()) &&
			bytes.Equal(nodeOne.MachineID, newNode.MachineID) {
			//this.Log.Info().Str("激活推送新节点程序", "").Send()
			continue
		}
		this.sendNewNode_net(ssOne, newNode)
	}
}

/*
读取需要询问关闭的连接名称
*/
func (this *SessionManager) readOutCloseConnName() {
	go func() {
		for {
			select {
			case ss := <-this.OutCloseConnName:
				//this.Log.Info().Str("询问关闭", "").Send()
				this.askCloseConn_net(ss)
			case <-this.contextRoot.Done():
				// this.Log.Info().Msgf("readOutCloseConnName done!")
				return
			}
		}
	}()
}

/*
定时删除数据库中过期的消息缓存
*/
func (this *SessionManager) loopCleanMessageCache() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-this.contextRoot.Done():
				// this.Log.Info().Msgf("loopCleanMessageCache done!")
				return
			}
			//计算24小时以前的时间UNIX
			// overtime := time.Now().Unix() - config.MsgCacheTimeOver
			// new(sqlite3_db.MessageCache).Remove(overtime)
		}
	}()
}

/*
保存超级节点ip地址到数据库做缓存
*/
func (this *SessionManager) saveLogicNodeIP() {
	ips := make([]ma.Multiaddr, 0)
	for _, one := range this.sessionEngine.GetSessionAll() {
		nodeInfo := GetNodeInfo(one)
		if nodeInfo == nil {
			continue
		}
		for _, addrInfo := range nodeInfo.GetMultiaddrWAN() {
			ips = append(ips, addrInfo.Multiaddr)
		}
	}
	this.addrManager.SavePeerEntryToDB(ips)
}

/*
设置状态为销毁
*/
func (this *SessionManager) SetDestroy() {
	this.destroy.Store(true)
	this.autonomyFinishState.Destroy()
	this.mam.SetDestroy()
}
