package session_manager

import (
	"bytes"
	"io"
	"time"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

/*
主动连接之前事件
*/
func (this *SessionManager) DialBeforeEvent(conn io.ReadWriter) utils.ERROR {
	return utils.NewErrorSuccess()
}

/*
主动连接之后事件
*/
func (this *SessionManager) DialAfterEvent(ss engine.Session) utils.ERROR {
	//this.Log.Info().Hex("创建会话 主动连接", ss.GetId()).Send()
	return utils.NewErrorSuccess()
}

/*
被动连接之前事件
*/
func (this *SessionManager) AcceptBeforeEvent(conn io.ReadWriter) utils.ERROR {
	return utils.NewErrorSuccess()
}

/*
被动连接之后事件
*/
func (this *SessionManager) AcceptAfterEvent(ss engine.Session) utils.ERROR {
	//this.Log.Info().Hex("创建会话 被动连接", ss.GetId()).Send()
	SetSessionCreateTime(ss, time.Now().Unix())
	//this.CloseConnLock.Lock()
	//defer this.CloseConnLock.Unlock()
	//this.guestSession[utils.Bytes2string(ss.GetId())] = ss
	this.guestSession.Store(utils.Bytes2string(ss.GetId()), ss)
	return utils.NewErrorSuccess()
}

/*
关闭连接之前事件
*/
func (this *SessionManager) CloseBeforeEvent(ss engine.Session) utils.ERROR {
	return utils.NewErrorSuccess()
}

/*
关闭连接之后事件
*/
func (this *SessionManager) CloseAfterEvent(ss engine.Session) utils.ERROR {
	//this.Log.Info().Hex("关闭会话", ss.GetId()).Send()
	this.CloseConnLock.Lock()
	defer this.CloseConnLock.Unlock()

	this.guestSession.Delete(utils.Bytes2string(ss.GetId()))

	addrInfoRemote := GetSessionCanConnPort(ss)
	if addrInfoRemote != nil {
		//this.Log.Info().Str("会话关闭回调", addrInfoRemote.Multiaddr.String()).Hex("sid", ss.GetId()).
		//	Uint8("连接类型", ss.GetConnType()).Send()
		//this.deleteAddrMulticast(addrInfo.Multiaddr)
		//delete(this.sessionMultiaddr, addrInfo.Multiaddr.String())
		//this.ConnStatus.QueryConn(addrInfo.Multiaddr)
		this.ConnStatus.Close(addrInfoRemote.Multiaddr, ss.GetId())
	} else {
		//this.Log.Info().Hex("关闭会话 此会话无节点信息", ss.GetId()).Send()
	}
	//删除组播中的缓存
	this.mam.DelAddrInfo(addrInfoRemote)

	//自己连接自己的会话断开，不触发重连
	remoteNode := GetNodeInfo(ss)
	if remoteNode == nil {
		return utils.NewErrorSuccess()
	}
	//addr1 := nodeStore.AddressFromB58String("7451t9dxbhttK7H9NUJefE26GrXLxuJhs6wjzvhSi12N")
	//if bytes.Equal(addr1, remoteNode.IdInfo.Id) {
	//	this.Log.Info().Str("要删除这个session", remoteNode.IdInfo.Id.B58String()).Send()
	//}
	remoteNode.RemoveSession(ss)
	equlAreaName := bytes.Equal(remoteNode.AreaName, this.nodeManager.NodeSelf.AreaName)
	//自己连自己
	if equlAreaName && bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), remoteNode.IdInfo.Id.Data()) &&
		bytes.Equal(this.nodeManager.NodeSelf.MachineID, remoteNode.MachineID) {
		//this.Log.Info().Hex("检查sessionid", one.GetId()).Str("有其他节点连接在线", "").Send()
		//自己连接自己的会话断开，不触发重连
		return utils.NewErrorSuccess()
	}
	//域名称相同
	if equlAreaName {
		select {
		case this.HaveChengeNodeAreaSelf <- nil:
		default:
		}
	} else {
		select {
		case this.HaveChengeNodeAreaOther <- nil:
		default:
		}
	}
	//清理节点信息中会话数量为0的记录
	if len(remoteNode.GetSessions()) == 0 {
		this.nodeManager.CleanNodeInfo()
	}
	//获取所有连接会话，看是否离线
	sss := this.sessionEngine.GetSessionAll()
	for _, one := range sss {
		//this.Log.Info().Hex("检查sessionid", one.GetId()).Send()
		remoteNode := GetNodeInfo(one)
		if remoteNode == nil {
			//this.Log.Info().Hex("检查sessionid", one.GetId()).Str("未注册节点不管", "").Send()
			//未注册节点不管
			continue
		}
		if !bytes.Equal(this.nodeManager.NodeSelf.AreaName, remoteNode.AreaName) ||
			!bytes.Equal(this.nodeManager.NodeSelf.IdInfo.Id.Data(), remoteNode.IdInfo.Id.Data()) ||
			!bytes.Equal(this.nodeManager.NodeSelf.MachineID, remoteNode.MachineID) {
			//this.Log.Info().Hex("检查sessionid", one.GetId()).Str("有其他节点连接在线", "").Send()
			//所有连接中，有一个其他节点的连接，就算在线
			return utils.NewErrorSuccess()
		}
	}
	//已经离线，需要断线重连
	this.reConnect(this.contextRoot)
	return utils.NewErrorSuccess()
}
