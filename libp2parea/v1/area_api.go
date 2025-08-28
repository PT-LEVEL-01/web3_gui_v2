package libp2parea

import (
	"bytes"
	"net"
	"strconv"
	"strings"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/global"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

/*
设置发现节点
*/
func (this *Area) SetDiscoverPeer(hosts ...string) {
	for _, one := range hosts {
		this.addrManager.AddSuperPeerAddr(this.AreaName[:], one)
	}
}

/*
获取本节点地址
*/
func (this *Area) GetNetId() nodeStore.AddressNet {
	return this.NodeManager.NodeSelf.IdInfo.Id
}

/*
获取本节点虚拟节点地址
*/
func (this *Area) GetVnodeId() virtual_node.AddressNetExtend {
	return this.Vm.DiscoverVnodes.Vnode.Vid
}

/*
链接到网络中去
*/
func (this *Area) CheckRepeatHash(sendHash []byte) bool {
	return this.MessageCenter.CheckRepeatHash(sendHash)
}

/*
获取idinfo
*/
func (this *Area) GetIdInfo() nodeStore.IdInfo {
	return this.NodeManager.NodeSelf.IdInfo
}

/*
获取NodeSelf
*/
func (this *Area) GetNodeSelf() nodeStore.Node {
	return *this.NodeManager.NodeSelf
}

/*
设置为测试网络
*/
func (this *Area) SetNetTypeToTest() {
	this.NodeManager.SetNetType(config.NetType_test)
}

/*
关闭所有网络连接
移动端关闭移动网络切换到wifi网络时调用
*/
func (this *Area) CloseNet() {
	// this.nodeManager.SetNetType(config.NetType_test)
	// this.sessionEngine.Suspend()
	ss := this.SessionEngine.GetAllSession(this.AreaName[:])
	for _, one := range ss {
		one.Close()
	}
	this.ResetAutonomyFinish()
}

/*
重新链接网络
*/
func (this *Area) ReconnectNet() bool {
	this.reConnect(this.contextRoot)
	return false
}

/*
检查是否在线(链接有没有断开)
*/
func (this *Area) CheckOnline() bool {
	return this.MessageCenter.CheckOnline()
}

/*
销毁
关闭连接，停止服务器，清理所有协程
*/
func (this *Area) Destroy() {
	this.destroy = true
	this.SessionEngine.Destroy(this.AreaName[:])
	this.MessageCenter.Destroy()
	this.canceRoot()
	this.autonomyFinishChanLock.Lock()
	close(this.autonomyFinishChan)
	this.autonomyFinishChanLock.Unlock()
	this.levelDB.Close()
}

/*
销毁
关闭连接，停止服务器，清理所有协程
*/
func (this *Area) DestroyByGlobal(global *global.Global) {
	this.destroy = true
	this.SessionEngine.Destroy(this.AreaName[:])
	this.MessageCenter.Destroy()
	this.canceRoot()
	this.autonomyFinishChanLock.Lock()
	close(this.autonomyFinishChan)
	this.autonomyFinishChanLock.Unlock()
	if global != nil {
		global.RemoveAreaInfo(this.AreaName[:])
	} else {
		this.levelDB.Close()
	}
}

/*
添加一个地址到白名单
*/
func (this *Area) AddWhiteList(addr nodeStore.AddressNet) bool {
	return this.NodeManager.AddWhiteList(addr)
}

/*
 * 根据目标ip地址及端口添加到白名单
 */
func (this *Area) AddAddrWhiteList(ip string, port uint16) (nodeStore.AddressNet, bool) {
	// 根据ip地址及端口，获取session和node信息
	if this.CheckSupportQuicConn() {
		ss, nodePoint, err := this.SessionEngine.GetQuicConnectInfoByIpAddr(this.AreaName[:], ip, uint32(port))
		if err == nil && nodePoint != nil {
			// 添加白名单
			node := nodePoint.(*nodeStore.Node)
			this.NodeManager.AddWhiteList(node.IdInfo.Id)

			// 添加节点信息
			this.SessionEngine.AddQuicConnectBySessionAndNode(this.AreaName[:], ss, nodePoint)

			return node.IdInfo.Id, true
		}
	}

	if this.CheckSupportTcpConn() {
		ss, nodePoint, err := this.SessionEngine.GetConnectInfoByIpAddr(this.AreaName[:], ip, uint32(port))
		if err == nil && nodePoint != nil {
			// 添加白名单
			node := nodePoint.(*nodeStore.Node)
			this.NodeManager.AddWhiteList(node.IdInfo.Id)

			// 添加节点信息
			this.SessionEngine.AddConnectBySessionAndNode(this.AreaName[:], ss, nodePoint)

			return node.IdInfo.Id, true
		}
	}

	return nil, false
}

/*
 * 根据目标ip地址及端口添加到白名单
 * @auth	qlw
 * @param 	ip			string		IP地址
 * @param 	port		uint16		端口
 * @return	sess		uint64		session连接信息
 * @return	success 	bool		是否添加成功标识
 */
func (this *Area) AddAddrWhiteListWithSessIndexBack(ip string, port uint16) (engine.Session, bool) {
	// 根据ip地址及端口，获取session和node信息
	if this.CheckSupportQuicConn() {
		ss, nodePoint, err := this.SessionEngine.GetQuicConnectInfoByIpAddr(this.AreaName[:], ip, uint32(port))
		if err == nil && nodePoint != nil && ss != nil {
			// 添加白名单
			node := nodePoint.(*nodeStore.Node)
			this.NodeManager.AddWhiteList(node.IdInfo.Id)

			// 添加节点信息
			this.SessionEngine.AddQuicConnectBySessionAndNode(this.AreaName[:], ss, nodePoint)

			return ss, true
		}
	}

	// 如果支持tcp连接, 则尝试进行tcp连接
	if this.CheckSupportTcpConn() {
		ss, nodePoint, err := this.SessionEngine.GetConnectInfoByIpAddr(this.AreaName[:], ip, uint32(port))
		if err == nil && nodePoint != nil && ss != nil {
			// 添加白名单
			node := nodePoint.(*nodeStore.Node)
			this.NodeManager.AddWhiteList(node.IdInfo.Id)

			// 添加节点信息
			this.SessionEngine.AddConnectBySessionAndNode(this.AreaName[:], ss, nodePoint)

			return ss, true
		}
	}

	return nil, false
}

/*
删除一个地址到白名单
*/
func (this *Area) RemoveWhiteList(addr nodeStore.AddressNet) bool {
	return this.NodeManager.DelWhiteList(addr)
}

/*
添加一个连接
*/
func (this *Area) AddConnect(ip string, port uint16) (engine.Session, error) {
	return this.SessionEngine.AddClientConn(this.AreaName[:], ip, uint32(port), false, engine.BothMod)
}

/*
关闭虚拟节点功能
*/
func (this *Area) CloseVnode() {
	this.Vc.Close()
}

/*
打开虚拟节点功能
*/
func (this *Area) OpenVnode() {
	this.Vc.Open()
}

/*
获取节点管理器
*/
func (this *Area) GetNodeManager() *nodeStore.NodeManager {
	return this.NodeManager
}

/*
搜索磁力节点网络地址
*/
func (this *Area) SearchNetAddr(nodeId *nodeStore.AddressNet) (*nodeStore.AddressNet, error) {
	res, err := this.MessageCenter.GetSearchNetAddr(nodeId, nil, nil, true, 1)
	if err != nil || len(res) == 0 {
		return nil, err
	}
	return &res[0], err
}

/*
 * 搜索磁力节点网络地址, 指定返回的数量地址
 *
 * @param	nodeId		AddressNet		磁力地址
 * @param	num			uint16			返回的数量
 * @return 	addrs		[]*AddressNet	地址数组
 * @return	err			error			错误信息
 */
func (this *Area) SearchNetAddrWithNum(nodeId *nodeStore.AddressNet, num uint16) ([]nodeStore.AddressNet, error) {
	res, err := this.MessageCenter.GetSearchNetAddr(nodeId, nil, nil, true, num)
	return res, err
}

/*
搜索磁力节点网络地址
*/
func (this *Area) SearchNetAddrOneByOne(nodeId *nodeStore.AddressNet, num uint16) ([]nodeStore.AddressNet, error) {
	return this.MessageCenter.GetSearchNetAddr(nodeId, nil, nil, true, num)
}

/*
 * 搜索磁力节点网络地址，可以指定发送端的代理节点
 */
func (this *Area) SearchNetAddrProxy(nodeId, recvProxyId, senderProxyId *nodeStore.AddressNet) (*nodeStore.AddressNet, error) {
	return this.MessageCenter.GetSearchNetAddrProxy(nodeId, recvProxyId, senderProxyId)
}

/*
 * 搜索磁力节点网络地址, 指定返回的数量地址
 *
 * @param	nodeId		AddressNet		磁力地址
 * @param	num			uint16			返回的数量
 * @return 	addrs		[]*AddressNet	地址数组
 * @return	err			error			错误信息
 */
func (this *Area) SearchNetAddrWithNumProxy(nodeId, recvProxyId, senderProxyId *nodeStore.AddressNet, num uint16) ([]nodeStore.AddressNet, error) {
	res, err := this.MessageCenter.GetSearchNetAddr(nodeId, recvProxyId, senderProxyId, true, num)
	return res, err
}

/*
搜索磁力节点网络地址
*/
func (this *Area) SearchNetAddrOneByOneProxy(nodeId, recvProxyId, senderProxyId *nodeStore.AddressNet, num uint16) ([]nodeStore.AddressNet, error) {
	return this.MessageCenter.GetSearchNetAddr(nodeId, recvProxyId, senderProxyId, true, num)
}

/*
 * 搜索磁力虚拟节点网络地址
 * TODO zyz 优化使用细节
 */
func (this *Area) SearchNetAddrVnode(nodeId *virtual_node.AddressNetExtend) (*virtual_node.AddressNetExtend, error) {
	return this.SearchNetAddrVnodeProxy(nodeId, nil, nil)
}

/*
 * 搜索磁力虚拟节点网络地址，可以指定接收端和发送端的代理节点
 * TODO zyz 优化使用细节
 */
func (this *Area) SearchNetAddrVnodeProxy(nodeId *virtual_node.AddressNetExtend, recvProxyId, senderProxyId *nodeStore.AddressNet) (*virtual_node.AddressNetExtend, error) {
	netIdBytes, err := this.Vc.SearchVnodeId(nodeId, recvProxyId, senderProxyId, false, 1)
	if err != nil {
		return nil, err
	} else if netIdBytes == nil || len(*netIdBytes) == 0 {
		return nil, nil
	}

	vnodeId := virtual_node.AddressNetExtend(*netIdBytes)
	return &vnodeId, nil
}

/*
获取所有网络连接
*/
func (this *Area) GetNetworkInfo() *[]nodeStore.AddressNet {
	logicAddr := this.NodeManager.GetLogicNodes()
	return &logicAddr
}

/*
设置leveldb数据库存放位置
*/
func (this *Area) SetLeveldbPath(dbpath string) {
	this.leveldbPath = dbpath
}

/*
设置本节点为手机端节点
*/
func (this *Area) SetPhoneNode() {
	this.NodeManager.NodeSelf.IsApp = true
}

/*
设置本节点为手机端节点
*/
func (this *Area) GetLevelDB() *utilsleveldb.LevelDB {
	return this.levelDB
}

/*
 * 设置区域上帝地址信息
 *
 * @auth qlw
 * @param 	ip			string	服务器ip地址
 * @param 	port 		int		服务器端口, 端口范围 (1024, 65535)
 * @return 	success		bool	是否添加成功
 * @return 	proxyAddr 	string	上帝节点地址，注意：如果区域没有启动，则该值为空串
 */
func (this *Area) SetAreaGodAddr(ip string, port int) (bool, string) {
	// 防止空的area操作, 出现空指针异常的问题
	if this == nil {
		return false, ""
	}

	success, godAddr, _ := this.SetAreaGodAddrWithConnRes(ip, port)

	return success, godAddr
}

/*
 * 设置区域上帝地址信息, 返回是否连接成功标识
 *
 * @auth qlw
 * @param 	ip			string	服务器ip地址
 * @param 	port 		int		服务器端口, 端口范围 (1024, 65535)
 * @return 	success		bool	是否添加成功
 * @return 	proxyAddr 	string	上帝节点地址，注意：如果区域没有启动，则该值为空串
 * @return 	connRel 	bool	是否连接成功标识
 */
func (this *Area) SetAreaGodAddrWithConnRes(ip string, port int) (bool, string, bool) {
	// 防止空的area操作, 出现空指针异常的问题
	if this == nil {
		return false, "", false
	}

	// utils.Log.Warn().Msgf("qlw-----设置上帝节点地址, ip:%s, port:%d", ip, port)

	// 是否需要加锁处理，防止多线程异常问题
	this.setGodHostLock.Lock()
	defer this.setGodHostLock.Unlock()

	// 1. 检查地址及端口是否合法
	// 1.1 检查地址是否合法
	if net.ParseIP(ip) == nil {
		// utils.Log.Info().Msgf("SetAreaGodAddr 输入的ip地址非法")
		return false, "", false
	}
	// 1.2 检查端口是否合法
	if port <= 0 || port > 65535 {
		utils.Log.Info().Msgf("SetAreaGodAddr 输入的端口地址非法")
		return false, "", false
	}
	// 1.3 检查设置的代理地址是不是自己
	if this.NodeManager.NodeSelf.Addr == ip && this.NodeManager.NodeSelf.TcpPort == uint16(port) {
		utils.Log.Error().Msgf("不能设置自己为超级代理!!!!!")
		return false, "", false
	}

	// 上帝节点host
	godHost := ip + ":" + strconv.Itoa(port)
	oldGodHost := this.GodHost
	oldIp, oldPort := this.SessionEngine.GetGodAddr(this.AreaName[:])

	// 2. 设置相关上帝节点信息
	// 2.1 设置发现节点地址
	this.SetDiscoverPeer(godHost)
	// 2.2 设置区域engine中的上帝节点信息
	this.SessionEngine.SetGodAddr(this.AreaName[:], ip, uint16(port))

	// 3. 设置黑名单节点地址信息
	this.GodHost = godHost

	// 4. 判断节点是否已启动
	var godAddr string
	var connRes bool
	// 4.1 如果没有启动，则不做任何处理
	// 已经启动了
	if this.MessageCenter.CheckOnline() {
		// utils.Log.Warn().Msgf("qlw-----设置上帝节点地址, 节点已经启动了")
		// 4.2 如果启动了，判断本地是否持有该地址及端口对应的连接信息
		var godNodeAddrStr string
		var sessionIndex uint64
		// var godNodeAddr nodeStore.AddressNet
		// 4.2.1 如果存在则不用再建立连接信息
		allSession := this.SessionEngine.GetAllSession(this.AreaName[:])
		for i := range allSession {
			if allSession[i].GetRemoteHost() != godHost {
				continue
			}
			// utils.Log.Warn().Msgf("qlw-----设置上帝节点地址, 连接信息之前已经存在了!!!")
			godNodeAddrStr = allSession[i].GetName()
			sessionIndex = allSession[i].GetIndex()

			// 如果之前没有设置上帝地址,但已经连接上了上帝节点，需要断开重连
			// 因为重连时，会告知对方他是我的上帝节点，它才能正确的更新服务器信息
			if oldGodHost == "" {
				allSession[i].Close()
			}
			break
		}

		// 4.2.1 如果没有连接信息，则建立连接
		if godNodeAddrStr == "" || oldGodHost == "" {
			session, success := this.AddAddrWhiteListWithSessIndexBack(ip, uint16(port))
			if !success || session == nil {
				// utils.Log.Warn().Msgf("qlw-----设置上帝节点地址, 添加白名单错误!!!")
				// 还原之前的上帝节点信息
				this.GodHost = oldGodHost
				this.SessionEngine.SetGodAddr(this.AreaName[:], oldIp, oldPort)

				return false, "", false
			}

			sessionIndex = session.GetIndex()
			godNodeAddrStr = utils.Bytes2string([]byte(session.GetName()))
		}

		// 5. 关闭所有非上帝节点的节点连接信息
		this.NodeManager.CloseNotGodNodesUseSessIndex(sessionIndex)

		// 6. 构建上帝节点地址的base58字符串的值
		godAddr = nodeStore.AddressNet([]byte(godNodeAddrStr)).B58String()

		// 7. 设置连接成功标识
		connRes = true

		// 8. 设置超级代理节点信息
		godId := nodeStore.AddressNet([]byte(godNodeAddrStr))
		this.GodID = &godId
	}

	return true, godAddr, connRes
}

/*
 * 设置本节点的机器id
 *
 * @param	machineID		string		机器id
 * @return	success			bool		是否设置成功
 */
func (this *Area) SetMachineID(machineID string) bool {
	if this.NodeManager == nil {
		utils.Log.Error().Msgf("还没有初始化NodeManager, 无法进行[设置机器id]操作!!!!")
		return false
	}

	// 设置长度检测
	if len(machineID) > config.MaxMachineIDLen {
		utils.Log.Error().Msgf("设置的机器Id长度过长, 最大长度为:%d, 当前设置长度为:%d!!!!!", config.MaxMachineIDLen, len(machineID))
		return false
	}

	// 已经设置过machineid，则不允许再设置
	if this.NodeManager.GetMachineID() != "" {
		return false
	}

	// 如果节点已经在线，也不允许再进行设置
	if this.MessageCenter.CheckOnline() {
		utils.Log.Error().Msgf("节点已经启动, 无法再进行设置设备机器Id操作!!!!!")
		return false
	}

	// 设置机器id值
	this.NodeManager.SetMachineID(machineID)
	return true
}

/*
 * 获取本节点的机器id
 *
 * @return	machineID		string		机器id
 */
func (this *Area) GetMachineID() string {
	if this.NodeManager == nil {
		utils.Log.Error().Msgf("还没有初始化NodeManager, 无法进行[获取机器id]操作!!!!")
		return ""
	}
	return this.NodeManager.GetMachineID()
}

/*
*删除本节点虚拟节点
 */
func (this *Area) DelSelfVnodeByAddress(addr nodeStore.AddressNet) {
	//1.检查是否自己创建过本虚拟地址
	if !this.Vm.FindInVnodeSelf(virtual_node.AddressNetExtend(addr)) ||
		bytes.Equal(this.NodeManager.NodeSelf.IdInfo.Id, addr) {
		return
	}

	//2.发送消息给其他节点通知删除虚拟节点
	content := []byte(addr)
	sNode := this.NodeManager.GetAllNodesAddr()
	for i, _ := range sNode {
		go this.MessageCenter.SendNeighborMsg(config.MSGID_del_vnode, &sNode[i], &content)
	}

	//3.删除本虚拟节点及自己其他节点up down相关信息
	this.Vm.DelSelfVnodeByAddr(addr)

}

/*
 * 设置本节点的机器id
 *
 * @param	connType		string		p2p连接类型: all tcp udp
 * @return	success			bool		是否设置成功
 */
func (this *Area) SetP2pConnType(connType string) bool {
	if connType == "" {
		utils.Log.Error().Msgf("请输入有效的连接限定字符串!!!!!")
		return false
	}

	// 如果节点已经在线，也不允许再进行设置
	if this.MessageCenter.CheckOnline() || this.levelDB != nil {
		utils.Log.Error().Msgf("节点已经启动, 无法再进行设置p2p连接类型操作!!!!!")
		return false
	}

	// 统一转换为小写
	connType = strings.ToLower(connType)

	// 根据设置信息更新连接方式
	if connType == "tcp" {
		this.connLimitType = config.CONN_TYPE_TCP
	} else if connType == "quic" {
		this.connLimitType = config.CONN_TYPE_QUIC
	} else if connType == "all" {
		this.connLimitType = config.CONN_TYPE_ALL
	} else {
		utils.Log.Error().Msgf("请输入有效的连接方式: all tcp quic")
		return false
	}

	return true
}

// 检查是否支持tcp连接
func (this *Area) CheckSupportTcpConn() bool {
	return this.connLimitType&config.CONN_TYPE_TCP != 0
}

// 检查是否支持quic连接
func (this *Area) CheckSupportQuicConn() bool {
	return this.connLimitType&config.CONN_TYPE_QUIC != 0
}
