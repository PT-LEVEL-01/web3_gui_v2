package libp2parea

import (
	ma "github.com/multiformats/go-multiaddr"
	"web3_gui/libp2parea/v2/engine"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
设置发现节点
*/
func (this *Node) SetDiscoverPeer(addrs []string) utils.ERROR {
	for _, one := range addrs {
		addrOne, err := ma.NewMultiaddr(one)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		ERR := this.addrManager.AddSuperPeerAddr(addrOne)
		if ERR.CheckFail() {
			return ERR
		}
	}
	return utils.NewErrorSuccess()
}

/*
获取本节点地址
*/
func (this *Node) GetNetId() *nodeStore.AddressNet {
	return this.NodeManager.NodeSelf.IdInfo.Id
}

/*
链接到网络中去
*/
func (this *Node) CheckRepeatHash(sendHash []byte) bool {
	return this.MessageCenter.CheckRepeatHash(sendHash)
}

/*
获取idinfo
*/
func (this *Node) GetIdInfo() nodeStore.IdInfo {
	return this.NodeManager.NodeSelf.IdInfo
}

/*
获取NodeSelf
*/
func (this *Node) GetNodeSelf() nodeStore.NodeInfo {
	return *this.NodeManager.NodeSelf
}

/*
关闭所有网络连接
移动端关闭移动网络切换到wifi网络时调用
*/
func (this *Node) CloseNet() {
	ss := this.SessionEngine.GetSessionAll()
	for _, one := range ss {
		one.Close()
	}
	this.ResetAutonomyFinish()
}

/*
重新链接网络
*/
func (this *Node) ReconnectNet() bool {
	this.SessionManager.StartConnectNet()
	return false
}

/*
检查是否在线(链接有没有断开)
*/
func (this *Node) CheckOnline() bool {
	return this.MessageCenter.CheckOnline()
}

/*
销毁
关闭连接，停止服务器，清理所有协程
*/
func (this *Node) Destroy() {
	this.SessionManager.SetDestroy()
	//this.destroy.Store(true)
	this.SessionEngine.Destroy()
	this.MessageCenter.Destroy()
	this.canceRoot()
	//this.autonomyFinishChanLock.Lock()
	//close(this.autonomyFinishChan)
	//this.autonomyFinishChanLock.Unlock()
	this.levelDB.Close()
}

/*
添加一个地址到白名单
*/
func (this *Node) AddWhiteListByAddr(addr nodeStore.AddressNet) bool {
	return this.NodeManager.AddWhiteListByAddr(addr)
}

/*
删除一个地址到白名单
*/
func (this *Node) RemoveWhiteList(addr nodeStore.AddressNet) bool {
	return this.NodeManager.DelWhiteListByAddr(addr)
}

/*
根据目标ip地址及端口添加到白名单
*/
func (this *Node) AddWhiteListByIp(ip ma.Multiaddr) {
	this.NodeManager.AddWhiteListByIP(ip)
}

/*
添加一个连接
*/
func (this *Node) AddConnect(ip ma.Multiaddr) (engine.Session, utils.ERROR) {
	return this.SessionEngine.Dial(ip.String())
}

/*
获取节点管理器
*/
func (this *Node) GetNodeManager() *nodeStore.NodeManager {
	return this.NodeManager
}

/*
获取所有网络连接
*/
func (this *Node) GetNetworkInfo() []nodeStore.NodeInfo {
	logicAddr := this.NodeManager.GetLogicNodesWAN()
	logicAddr = append(logicAddr, this.NodeManager.GetLogicNodesLAN()...)
	return logicAddr
}

/*
设置leveldb数据库存放位置
*/
func (this *Node) SetLeveldbPath(dbpath string) {
	this.leveldbPath = dbpath
}

/*
设置本节点为手机端节点
*/
func (this *Node) GetLevelDB() *utilsleveldb.LevelDB {
	return this.levelDB
}

/*
 * 获取本节点的机器id
 *
 * @return	machineID		string		机器id
 */
func (this *Node) GetMachineID() []byte {
	if this.NodeManager == nil {
		//utils.Log.Error().Msgf("还没有初始化NodeManager, 无法进行[获取机器id]操作!!!!")
		return nil
	}
	return this.NodeManager.NodeSelf.MachineID
}

//// 检查是否支持tcp连接
//func (this *Area) CheckSupportTcpConn() bool {
//	return this.connLimitType&config.CONN_TYPE_TCP != 0
//}
//
//// 检查是否支持quic连接
//func (this *Area) CheckSupportQuicConn() bool {
//	return this.connLimitType&config.CONN_TYPE_QUIC != 0
//}
