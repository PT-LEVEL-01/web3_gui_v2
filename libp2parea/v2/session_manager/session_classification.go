package session_manager

import (
	"bytes"
	ma "github.com/multiformats/go-multiaddr"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
根据连接信息，检查自己节点的公开端口信息
*/
func (this *SessionManager) checkSelfOpenAddr(nodeRemote *nodeStore.NodeInfo) {
	addrInfo, ERR := engine.CheckAddr(nodeRemote.RemoteMultiaddr)
	if ERR.CheckFail() {
		this.Log.Error().Str("节点信息不合法", "").Send()
		return
	}
	//nodeRemote.AddSession(msg.Session)
	////节点信息放入session
	//SetNodeInfo(msg.Session, nodeRemote)

	//判断对方是通过哪个协议和端口连接的自己
	isOnlyIP := addrInfo.IsOnlyIp()
	multiaddrs := this.nodeManager.NodeSelf.GetMultiaddrLAN()
	if isOnlyIP {
		multiaddrs = this.nodeManager.NodeSelf.GetMultiaddrWAN()
	}
	//判断自己是否有收录这个端口
	have := false
	for _, mawOne := range multiaddrs {
		if mawOne.Multiaddr.Equal(nodeRemote.RemoteMultiaddr) {
			have = true
			break
		}
	}
	if !have {
		//没有收录这个公开端口，则自己尝试连接自己，看是否可行
		//exist, ERR := this.ConnStatus.QueryConn(addrInfo.Multiaddr)
		//if ERR.CheckFail() {
		//	return
		//}
		//if !exist {
		//	//this.Log.Info().Str("这个连接已经存在", addr.String()).Send()
		//	this.Log.Info().Str("开始连接节点", "尝试自己连接自己").Send()
		//	ERR = this.TryConnectNet(addrInfo.Multiaddr)
		//}
		ERR = this.TryConnectNet(addrInfo.Multiaddr)
		//ERR = this.connectNet(*addrInfo)
		//自己连接自己，会报错，这里不报错则不对
		if ERR.CheckSuccess() {
			//this.Log.Error().Str("是自己连接自己", "").Send()
			if isOnlyIP {
				this.nodeManager.NodeSelf.AddMultiaddrWAN(addrInfo)
			} else {
				this.nodeManager.NodeSelf.AddMultiaddrLAN(addrInfo)
			}
		}
	}
}

/*
对注册的节点分类存放
*/
func (this *SessionManager) ClassificationSession(session engine.Session, nodeRemote *nodeStore.NodeInfo, connectCan,
	isOnlyIp bool, addrInfo *engine.AddrInfo) {
	//this.Log.Info().Str("节点分类存放", nodeRemote.IdInfo.Id.B58String()).Hex("sid", nodeRemote.GetSessions()[0].GetId()).Send()
	//检查对方是否可以连接
	//connectCan, isOnlyIp, addrInfo, ERR := this.checkNodeCanConnect(session, nodeRemote)
	//if ERR.CheckFail() {
	//	return
	//}
	//SetSessionAddrMultiaddr(session, addrInfo)
	//this.Log.Info().Interface("是否能连接上", connectCan).Send()
	//域名称是否相同
	if bytes.Equal(nodeRemote.AreaName, this.nodeManager.NodeSelf.AreaName) {
		if connectCan {
			select {
			case this.HaveChengeNodeAreaSelf <- nodeRemote:
			default:
			}
		}
		//对方通过内网地址还是外网地址发起的连接
		if isOnlyIp {
			if connectCan {
				//可以连接，保存这个连接
				nodeRemote.AddMultiaddrWAN(addrInfo)
				this.saveSessionToWAN(session, nodeRemote)
			} else {
				//不能连接，则放入代理节点中
				this.nodeManager.AddNodeInfoProxy(nodeRemote)
				//this.proxySession.Store(utils.Bytes2string(session.GetId()), session)
			}
		} else {
			if connectCan {
				//this.Log.Info().Str("保存这个连接到LAN", nodeRemote.IdInfo.Id.B58String()).Send()
				//可以连接，保存这个连接
				nodeRemote.AddMultiaddrLAN(addrInfo)
				this.saveSessionToLAN(session, nodeRemote)
			} else {
				//不能连接
				this.nodeManager.AddNodeInfoProxy(nodeRemote)
				//this.FreeSession.Store(utils.Bytes2string(session.GetId()), session)
			}
		}
	} else {
		//域不相同
		if connectCan {
			select {
			case this.HaveChengeNodeAreaOther <- nodeRemote:
			default:
			}
		}
		//对方通过内网地址还是外网地址发起的连接
		if isOnlyIp {
			if connectCan {
				//可以连接
				nodeRemote.AddMultiaddrWAN(addrInfo)
				this.saveAreaSessionToWAN(session, nodeRemote)
				//this.nodeManager.CheckNeedAreaWan(nodeRemote, true)
			} else {
				//不能连接
				this.nodeManager.AddNodeInfoProxy(nodeRemote)
				//this.FreeSession.Store(utils.Bytes2string(session.GetId()), session)
			}
		} else {
			if connectCan {
				//可以连接
				nodeRemote.AddMultiaddrLAN(addrInfo)
				this.saveAreaSessionToLAN(session, nodeRemote)
			} else {
				//不能连接
				this.nodeManager.AddNodeInfoProxy(nodeRemote)
				//this.FreeSession.Store(utils.Bytes2string(session.GetId()), session)
			}
		}
	}
}

/*
检查此节点是否可以连上
@return    bool    是否可以连接
@return    bool    是否公网ip
@return    *engine.AddrInfo    能连接的对方节点信息
@return    utils.ERROR         错误信息
*/
func (this *SessionManager) checkNodeCanConnect(session engine.Session, nodeRemote *nodeStore.NodeInfo) (bool, bool, *engine.AddrInfo, utils.ERROR) {
	//是我向对方发起的会话
	if session.GetConnType() == engine.CONN_TYPE_TCP_client ||
		session.GetConnType() == engine.CONN_TYPE_WS_client ||
		session.GetConnType() == engine.CONN_TYPE_QUIC_client {
		mAddr := session.GetRemoteMultiaddr()
		addrInfo, ERR := engine.CheckAddr(mAddr)
		if ERR.CheckFail() {
			return true, false, addrInfo, ERR
		}
		return true, addrInfo.IsOnlyIp(), addrInfo, utils.NewErrorSuccess()
	}
	//是对方向我发起的会话
	addrInfo, ERR := engine.CheckAddr(nodeRemote.RemoteMultiaddr)
	if ERR.CheckFail() {
		return false, false, nil, ERR
	}
	isOnlyIp := addrInfo.IsOnlyIp()

	//ip := strings.SplitN(session.GetRemoteHost(), ":", 2)[0]
	//isOnlyIp := utils.IsOnlyIp(ip)

	addrInfos := nodeRemote.GetMultiaddrWAN()
	if !isOnlyIp {
		addrInfos = nodeRemote.GetMultiaddrLAN()
	}
	if len(addrInfos) > 0 {
		mas := make([]ma.Multiaddr, 0, len(addrInfos))
		for _, one := range addrInfos {
			mas = append(mas, one.Multiaddr)
		}
		for _, one := range mas {
			ERR := this.TryConnectNet(one)
			//newSession, addrInfo, ERR := this.syncConnectNetOne(one, nil, nil)
			if ERR.CheckFail() {
				//未连接上
			} else {
				//连接上了
				addrInfo, ERR = engine.CheckAddr(one)
				if ERR.CheckFail() {
					return true, isOnlyIp, addrInfo, ERR
				}
				return true, isOnlyIp, addrInfo, utils.NewErrorSuccess()
			}
		}
	}
	//this.Log.Info().Str("开始拼接节点地址", "").Send()
	//拼接一个地址试试
	mAddr, ERR := BuildMultiaddrBySession(session)
	if ERR.CheckFail() {
		//this.Log.Info().Str("开始拼接节点地址", ERR.String()).Send()
		return false, isOnlyIp, nil, ERR
	}
	//this.Log.Info().Str("开始拼接节点地址", mAddr.String()).Send()
	ERR = this.TryConnectNet(mAddr)
	//newSession, addrInfo, ERR := this.syncConnectNetOne(mAddr, nil, nil)
	if ERR.CheckFail() {
		//未连接上
		return false, isOnlyIp, nil, utils.NewErrorSuccess()
	} else {
		//连接上了
		addrInfo, ERR = engine.CheckAddr(mAddr)
		if ERR.CheckFail() {
			return true, isOnlyIp, addrInfo, ERR
		}
		return true, isOnlyIp, addrInfo, utils.NewErrorSuccess()
	}
}

/*
保存会话到外网
*/
func (this *SessionManager) saveAreaSessionToWAN(session engine.Session, nodeRemote *nodeStore.NodeInfo) {
	//key := utils.Bytes2string(append(nodeRemote.AreaName, nodeRemote.IdInfo.Id...))
	//判断是否自己需要的逻辑域
	need, removeNodes, exist := this.nodeManager.CheckNeedAreaWan(nodeRemote, true)
	if exist {
		//重复连接
		//判断是多端登录，还是相同节点多个连接
	} else if need {
		//保存这个连接
	} else {
		//不需要的节点
		this.nodeManager.AddNodeInfoWAN(nodeRemote)
		select {
		case this.OutCloseConnName <- session:
		default:
		}
	}
	//需要删除的节点
	//this.lock.Lock()
	for _, one := range removeNodes {
		//this.Log.Info().Int("添加到游离的节点", len(one.GetSessions())).Send()
		this.nodeManager.AddNodeInfoFree(&one)
		for _, one := range one.GetSessions() {
			select {
			case this.OutCloseConnName <- one:
			default:
			}
		}
	}
	//this.lock.Unlock()
	return
}

/*
保存会话到局域网
*/
func (this *SessionManager) saveAreaSessionToLAN(session engine.Session, nodeRemote *nodeStore.NodeInfo) {
	//key := utils.Bytes2string(append(nodeRemote.AreaName, nodeRemote.IdInfo.Id...))
	//this.Log.Info().Hex("保存逻辑域节点到 LAN", nodeRemote.AreaName).Send()
	//判断是否自己需要的逻辑域
	need, removeNodes, exist := this.nodeManager.CheckNeedAreaLan(nodeRemote, true)
	//this.Log.Info().Bool("保存节点到LAN", need).Send()
	//this.Log.Info().Interface("要删除的地址", removeIDs).Send()
	if exist {
		//重复连接
		//判断是多端登录，还是相同节点多个连接
	} else if need {
		//保存这个连接
	} else {
		//不需要的节点
		//this.Log.Info().Str("添加到游离的节点", nodeRemote.IdInfo.Id.B58String()).Int("添加到游离的节点", len(nodeRemote.GetSessions())).Send()
		this.nodeManager.AddNodeInfoFree(nodeRemote)
		select {
		case this.OutCloseConnName <- session:
		default:
		}
	}
	//需要删除的节点
	for _, one := range removeNodes {
		this.Log.Info().Str("添加到游离的节点", one.IdInfo.Id.B58String()).Int("添加到游离的节点", len(one.GetSessions())).Send()
		this.nodeManager.AddNodeInfoFree(&one)
		for _, one := range one.GetSessions() {
			select {
			case this.OutCloseConnName <- one:
			default:
			}
		}
	}
	return
}

/*
保存会话到外网
*/
func (this *SessionManager) saveSessionToWAN(session engine.Session, nodeRemote *nodeStore.NodeInfo) {
	//可以连接，判断是否自己需要的逻辑节点
	need, removeNodes, exist := this.nodeManager.CheckNeedNodeWan(nodeRemote, true)
	//this.Log.Info().Bool("保存节点到WAN", need).Send()
	//this.Log.Info().Interface("要删除的地址", removeIDs).Send()
	if exist {
		//重复连接
		//判断是多端登录，还是相同节点多个连接
	} else if need {
		//保存这个连接
	} else {
		//不需要的节点
		//this.Log.Info().Str("添加到游离的节点", nodeRemote.IdInfo.Id.B58String()).Int("添加到游离的节点", len(nodeRemote.GetSessions())).Send()
		this.nodeManager.AddNodeInfoFree(nodeRemote)
		select {
		case this.OutCloseConnName <- session:
		default:
		}
	}
	//需要删除的节点
	for _, one := range removeNodes {
		//this.Log.Info().Str("添加到游离的节点", one.IdInfo.Id.B58String()).Int("添加到游离的节点", len(one.GetSessions())).Send()
		this.nodeManager.AddNodeInfoFree(one)
		for _, one := range one.GetSessions() {
			select {
			case this.OutCloseConnName <- one:
			default:
			}
		}
	}
	return
}

/*
保存会话到内网
*/
func (this *SessionManager) saveSessionToLAN(session engine.Session, nodeRemote *nodeStore.NodeInfo) {
	//可以连接，判断是否自己需要的逻辑节点
	need, removeNodes, exist := this.nodeManager.CheckNeedNodeLan(nodeRemote, true)
	//this.Log.Info().Bool("保存节点到LAN", need).Send()
	if exist {
		//重复连接
		//this.Log.Info().Str("保存连接 重复连接", nodeRemote.IdInfo.Id.B58String()).Send()
	} else if need {
		//保存这个连接
		//this.Log.Info().Str("保存连接 需要连接", nodeRemote.IdInfo.Id.B58String()).Send()
	} else {
		//this.Log.Info().Str("保存连接 不需要连接", nodeRemote.IdInfo.Id.B58String()).Send()
		//不需要的节点
		//this.Log.Info().Str("添加到游离的节点", nodeRemote.IdInfo.Id.B58String()).Int("添加到游离的节点", len(nodeRemote.GetSessions())).Send()
		this.nodeManager.AddNodeInfoFree(nodeRemote)
		//this.Log.Info().Str("保存连接 不需要连接", nodeRemote.IdInfo.Id.B58String()).Send()
		//this.FreeSession.Store(utils.Bytes2string(session.GetId()), session)
		select {
		case this.OutCloseConnName <- session:
		default:
		}
		//this.Log.Info().Str("保存连接 不需要连接", nodeRemote.IdInfo.Id.B58String()).Send()
	}
	//需要删除的节点
	for _, one := range removeNodes {
		//this.Log.Info().Str("添加到游离的节点", one.IdInfo.Id.B58String()).Int("添加到游离的节点", len(one.GetSessions())).Send()
		//this.Log.Info().Str("保存连接 不需要连接", one.IdInfo.Id.B58String()).Hex("sid", one.GetSessions()[0].GetId()).Send()
		this.nodeManager.AddNodeInfoFree(one)
		for _, one := range one.GetSessions() {
			select {
			case this.OutCloseConnName <- one:
			default:
			}
		}
	}
	return
}

/*
在外网会话中查询指定地址和机器码
*/
func (this *SessionManager) FindSessionAreaSelfWANByAddrAndMid(addr nodeStore.AddressNet, mid []byte) engine.Session {
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	//ssMap, ok := this.sessionAreaSelfWAN[utils.Bytes2string(addr)]
	//if !ok {
	//	return nil
	//}
	//for _, one := range *ssMap {
	//	nodeRemote := GetNodeInfo(one)
	//	if nodeRemote == nil {
	//		continue
	//	}
	//	if bytes.Equal(nodeRemote.MachineID, mid) {
	//		return one
	//	}
	//}
	return nil
}

/*
在局域网会话中查询指定地址和机器码
*/
func (this *SessionManager) FindSessionAreaSelfLANByAddrAndMid(addr nodeStore.AddressNet, mid []byte) engine.Session {
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	//ssMap, ok := this.sessionAreaSelfLAN[utils.Bytes2string(addr)]
	//if !ok {
	//	return nil
	//}
	//for _, one := range *ssMap {
	//	nodeRemote := GetNodeInfo(one)
	//	if nodeRemote == nil {
	//		continue
	//	}
	//	if bytes.Equal(nodeRemote.MachineID, mid) {
	//		return one
	//	}
	//}
	return nil
}

/*
在外网会话中查询指定地址和机器码
*/
func (this *SessionManager) FindSessionAreaOtherWANByAddrAndMid(areaName []byte, addr nodeStore.AddressNet, mid []byte) engine.Session {
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	//ssMap, ok := this.sessionAreaSelfWAN[utils.Bytes2string(append(areaName, addr...))]
	//if !ok {
	//	return nil
	//}
	//for _, one := range *ssMap {
	//	nodeRemote := GetNodeInfo(one)
	//	if nodeRemote == nil {
	//		continue
	//	}
	//	if bytes.Equal(nodeRemote.MachineID, mid) {
	//		return one
	//	}
	//}
	return nil
}

/*
在局域网会话中查询指定地址和机器码
*/
func (this *SessionManager) FindSessionAreaOtherLANByAddrAndMid(areaName []byte, addr nodeStore.AddressNet, mid []byte) engine.Session {
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	//ssMap, ok := this.sessionAreaSelfLAN[utils.Bytes2string(append(areaName, addr...))]
	//if !ok {
	//	return nil
	//}
	//for _, one := range *ssMap {
	//	nodeRemote := GetNodeInfo(one)
	//	if nodeRemote == nil {
	//		continue
	//	}
	//	if bytes.Equal(nodeRemote.MachineID, mid) {
	//		return one
	//	}
	//}
	return nil
}
