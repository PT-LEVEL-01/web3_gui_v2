package session_manager

import (
	"bytes"
	"context"
	ma "github.com/multiformats/go-multiaddr"
	"time"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

/*
开始连接到网络
*/
func (this *SessionManager) StartConnectNet() {
	//this.Log.Debug().Msgf("获取所有发现节点地址 start")
	addrs, dns := this.addrManager.LoadAddrForAll()
	//电信网络移动端连接域名很慢

	//this.Log.Info().Interface("开始连接节点", addrs).Send()
	//先连接缓存中的ip地址
	success := this.syncConnectNet(addrs)
	if !success {
		//this.Log.Info().Str("开始连接节点", "").Send()
		//再连接域名。连接域名需要一次预连接，速度慢
		success = this.syncConnectNet(dns)
	}

	this.searchArea()

	//this.Log.Info().Str("开始连接节点", "").Send()
	if !success {
		//this.Log.Info().Str("开始连接节点", "").Send()
		//当首节点无其他节点可以连接时，这时候要给组网完成一个信号
		this.autonomyFinishState.SetAutonomyFinish()
		//this.Log.Info().Bool("startUp end", success).Send()
	} else {
		//判断是否找到自己的域
		found := this.checkAreaNetOnlineWAN()
		if !found {
			found = this.checkAreaNetOnlineLAN()
		}
		if !found {
			this.Log.Info().Str("给组网一个完成信号", "").Send()
			//自己域下的首节点，给组网一个完成信号
			this.autonomyFinishState.SetAutonomyFinish()
		}
	}
	this.Log.Info().Bool("startUp end", success).Send()
}

/*
异步链接到网络中去
所有地址都会去尝试连接
其中一个连接成功，则立即返回true
全部连接失败，则返回false
@return    bool    是否连接成功
*/
func (this *SessionManager) syncConnectNet(addrs []ma.Multiaddr) (success bool) {
	success = false
	if len(addrs) == 0 {
		return
	}
	//去除重复
	//addrs = utils.DistinctString(addrs)
	// for _, one := range addrs {
	// 	this.Log.Info().Msgf("去除重复后的ip:%s", one)
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
			// this.Log.Info().Msgf("等待返回:%d", countResult)
			countResult++
			success = true
		}
	}
	// this.Log.Info().Msgf("还需等待返回:%d", len(addrs)-countResult)
	for i := 0; i < len(addrs)-countResult; i++ {
		// this.Log.Info().Msgf("等待返回:%d", i)
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
func (this *SessionManager) syncConnectNetOne(addr ma.Multiaddr, token chan bool, resultSuccess chan bool) (*engine.AddrInfo, utils.ERROR) {
	connSuccess := false
	//this.Log.Info().Str("syncConnectNetOne 开始连接节点", addr.String()).Send()
	defer func() {
		// this.Log.Info().Msgf("syncConnectNet end 111111")
		if resultSuccess != nil {
			resultSuccess <- connSuccess
		}
		// this.Log.Info().Msgf("syncConnectNet end 222222", ok)
		if token != nil {
			<-token
		}
		// this.Log.Info().Msgf("syncConnectNet end 3333333")
	}()
	addrInfo, ERR := engine.CheckAddr(addr)
	if ERR.CheckFail() {
		connSuccess = false
		return nil, ERR
	}
	//this.Log.Info().Str("syncConnectNetOne 开始连接节点", addr.String()).Send()
	ERR = this.connectNet(*addrInfo)
	if ERR.CheckFail() {
		connSuccess = false
		return addrInfo, ERR
	}
	//this.Log.Info().Str("syncConnectNetOne 开始连接节点", addr.String()).Send()
	connSuccess = true
	return addrInfo, utils.NewErrorSuccess()
}

/*
链接到网络中去
@return    bool    是否连接成功
*/
func (this *SessionManager) connectNet(addrInfo engine.AddrInfo) utils.ERROR {
	if addrInfo.Port == "0" {
		return utils.NewErrorSuccess()
	}
	return this.ConnStatus.Conn(addrInfo.Multiaddr, func() (engine.Session, utils.ERROR) {
		//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
		//_, exist := this.sessionMultiaddr[addrInfo.Multiaddr.String()]
		//if exist {
		//	this.Log.Info().Str("这个连接已经存在", addrInfo.Multiaddr.String()).Send()
		//	return utils.NewErrorSuccess()
		//}
		//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
		session, ERR := this.sessionEngine.Dial(addrInfo.Multiaddr.String())
		if ERR.CheckFail() {
			//this.Log.Error().Str("创建会话 错误", ERR.String()).Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
			return session, ERR
		}
		//this.Log.Info().Interface("创建会话 主动连接", session).Str("开始连接节点", addrInfo.Multiaddr.String()).Str("返回的错误", ERR.String()).Send()
		//this.Log.Info().Str("DialAfterEvent", "11111111").Send()
		//向对方注册本节点
		nodeRemote, ERR := this.exchangeNodeInfo_net(session)
		if ERR.CheckFail() {
			this.Log.Error().Str("创建会话 错误", ERR.String()).Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
			return session, ERR
		}

		//设置可连接节点的公开端口
		SetSessionCanConnPort(session, &addrInfo)
		//this.Log.Info().Str("开始连接节点", "11111111").Send()
		//节点信息放入session
		SetNodeInfo(session, nodeRemote)
		//把会话创建时间放入session
		SetSessionCreateTime(session, time.Now().Unix())
		//放入已经注册节点管理器
		//this.guestSession.Delete(utils.Bytes2string(ss.GetId()))
		//this.registSession.Store(utils.Bytes2string(ss.GetId()), ss)
		//this.Log.Info().Str("DialAfterEvent", "11111111").Send()

		//this.lock.Lock()
		//this.sessionMultiaddr[ss.GetRemoteMultiaddr().String()] = false
		//this.lock.Unlock()

		//this.Log.Info().Hex("新连接id", session.GetId()).Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
		//nodeRemote := GetNodeInfo(session)
		equlAreaName := bytes.Equal(nodeRemote.AreaName, this.nodeManager.NodeSelf.AreaName)
		//是否自己连接自己
		if equlAreaName && bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), nodeRemote.IdInfo.Id.Data()) &&
			bytes.Equal(this.nodeManager.NodeSelf.MachineID, nodeRemote.MachineID) {
			//this.Log.Info().Str("关闭连接，自己连接自己，关闭这个连接", addrInfo.Multiaddr.String()).
			//	Str("本地端口", session.GetLocalHost()).Str("远端端口", session.GetRemoteHost()).Send()

			//this.Log.Info().Str("关闭节点", "--------").Send()
			//是自己连接自己
			session.Close()
			ERR = utils.NewErrorBus(config.ERROR_code_dial_self, "")
			return session, ERR
		}
		//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
		//向对方注册本节点
		ERR = this.registerNodeInfo_net(session)
		if ERR.CheckFail() {
			if ERR.Code != config.ERROR_code_Repeated_connections {
				this.Log.Error().Str("创建会话 错误", ERR.String()).Send()
			}
			//this.Log.Info().Str("关闭节点", "--------").Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
			session.Close()
			return session, ERR
		}

		//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
		this.ClassificationSession(session, nodeRemote, true, addrInfo.IsOnlyIp(), &addrInfo)
		if !equlAreaName {
			//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
			//不同域，需要寻找自己的域网络
			select {
			case this.HaveChengeNodeAreaOther <- nodeRemote:
			default:
			}
		} else {
			//this.ClassificationSession(session, nodeRemote)
			select {
			case this.HaveChengeNodeAreaSelf <- nodeRemote:
			default:
			}
		}
		config.IsOnline.Store(true)
		//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
		// this.Log.Debug().Msgf("超级节点为: %s", nodeStore.SuperPeerId.B58String())
		//select {
		//case this.registNodeTag <- false:
		//default:
		//}
		//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
		return session, utils.NewErrorSuccess()
	})

}

/*
尝试连接一个节点，连接成功后关闭会话
作用：判断此地址是否可以连接上。
*/
func (this *SessionManager) TryConnectNet(addr ma.Multiaddr) utils.ERROR {
	exist := this.ConnStatus.QueryConn(addr)
	//if ERR.CheckFail() {
	//	return ERR
	//}
	if exist {
		//this.Log.Info().Str("这个连接已经存在", addr.String()).Send()
		return utils.NewErrorSuccess()
	}
	//this.Log.Info().Str("开始连接节点", addr.String()).Send()
	//this.Log.Info().Str("开始连接节点", addrInfo.Multiaddr.String()).Send()
	session, ERR := this.sessionEngine.Dial(addr.String())
	if ERR.CheckFail() {
		return ERR
	}
	//this.Log.Info().Str("开始连接节点", addr.String()).Send()
	_, ERR = this.exchangeNodeInfo_net(session)
	if ERR.CheckFail() {
		return ERR
	}
	//this.Log.Info().Str("开始连接节点", addr.String()).Send()
	session.Close()
	//this.Log.Info().Str("开始连接节点", addr.String()).Send()
	return utils.NewErrorSuccess()
}

/*
断线重连
*/
func (this *SessionManager) reConnect(c context.Context) {
	//有调用此接口，则激活立即重连
	this.reConnTimer.Reset()
	this.reConnTimer.Release()
	if !this.isReconn.CompareAndSwap(false, true) {
		//有重连程序在跑，则退出
		return
	}
	//销毁则不重连
	if this.destroy.Load() {
		return
	}
	//this.Log.Info().Msgf("开始重连:%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String())
	for {
		//在线了
		if this.CheckOnline() {
			this.isReconn.Store(false)
			return
		}
		//销毁了
		if this.reConnTimer.Wait(c) == 0 {
			this.isReconn.Store(false)
			return
		}
		//在线状态
		if this.CheckOnline() {
			this.isReconn.Store(false)
			return
		}
		//开始重新连入网络
		this.StartConnectNet()
	}
}

/*
判断是否在线
*/
func (this *SessionManager) CheckOnline() bool {
	if len(this.sessionEngine.GetSessionAll()) > 0 {
		return true
	}
	return false
}

/*
等待网络自治完成，阻塞接口，需要等待
*/
func (this *SessionManager) WaitAutonomyFinish() {
	this.autonomyFinishState.WaitAutonomyFinish()
}

/*
检查网络自治是否已经完成，立即返回，不等待
*/
func (this *SessionManager) CheckAutonomyFinish() (autoFinish bool) {
	return this.autonomyFinishState.CheckAutonomyFinish()
}

/*
重置网络自治接口
*/
func (this *SessionManager) ResetAutonomyFinish() {
	this.autonomyFinishState.ResetAutonomyFinish()
}

/*
发送节点上线广播信息
*/
func (this *SessionManager) sendOnlineMulticast() {
	//bs, err := this.NodeManager.NodeSelf.Proto()
	//if err != nil {
	//	return
	//}
	//
	//// this.Log.Info().Msgf("发送节点在线广播消息: self:%s", addrNet.B58String())
	//if err := this.SendMulticastMsg(config.MSGID_multicast_online_recv, &bs); err != nil {
	//	// this.Log.Info().Msgf("发送节点下线广播消息 err:%s", err)
	//}
}

/*
检查并保存主动连接已有节点地址
@return    bool    是否存在
*/
//func (this *SessionManager) checkAndSaveAddrMulticast(addr ma.Multiaddr, ss engine.Session) (engine.Session, bool) {
//	this.Log.Info().Str("检查并保存节点地址", addr.String()).Send()
//	sessionItr, exist := this.sessionMultiaddr.LoadOrStore(addr.String(), ss)
//	if !exist {
//		return nil, false
//	}
//	session := sessionItr.(engine.Session)
//	return session, exist
//}

/*
删除主动连接已有节点地址
*/
//func (this *SessionManager) deleteAddrMulticast(addr ma.Multiaddr) {
//	if addr == nil || addr.String() == "" {
//		return
//	}
//	this.Log.Info().Str("删除节点地址", addr.String()).Send()
//	this.sessionMultiaddr.Delete(addr.String())
//}
