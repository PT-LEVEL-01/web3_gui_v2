package libp2parea

import (
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/libp2parea/v2/message_center"
	nodeStore "web3_gui/libp2parea/v2/node_store"
)

func (this *Node) RegisterCoreMsg() {
	this.MessageCenter.RegisterMsgVersion()

	//this.MessageCenter.Register_neighbor(gconfig.MSGID_getNearSuperIP, this.GetNearSuperAddr)                    //从邻居节点得到自己的逻辑节点
	//this.MessageCenter.Register_neighbor(gconfig.MSGID_getNearSuperIP_recv, this.GetNearSuperAddr_recv)          //从邻居节点得到自己的逻辑节点_返回
	//this.MessageCenter.Register_multicast(gconfig.MSGID_multicast_online_recv, this.MulticastOnline_recv)        //接收节点上线广播
	//this.MessageCenter.Register_neighbor(gconfig.MSGID_ask_close_conn_recv, this.AskCloseConn_recv)              //询问关闭连接
	//this.MessageCenter.Register_p2p(gconfig.MSGID_SearchAddr, this.MessageCenter.SearchAddress)                  //搜索节点，返回节点真实地址
	//this.MessageCenter.Register_p2p(gconfig.MSGID_SearchAddr_recv, this.MessageCenter.SearchAddress_recv)        //搜索节点，返回节点真实地址_返回
	//this.MessageCenter.Register_p2p(gconfig.MSGID_security_create_pipe, this.MessageCenter.CreatePipe)           //创建加密通道
	//this.MessageCenter.Register_p2p(gconfig.MSGID_security_create_pipe_recv, this.MessageCenter.CreatePipe_recv) //创建加密通道_返回
	//this.MessageCenter.Register_p2pHE(gconfig.MSGID_security_pipe_error, this.MessageCenter.Pipe_error)          //解密错误
	//this.MessageCenter.Register_p2p(gconfig.MSGID_search_node, this.SearchNode)                                  //查询一个节点是否在线
	//this.MessageCenter.Register_p2p(gconfig.MSGID_search_node_recv, this.SearchNode_recv)                        //查询一个节点是否在线_返回
	//
	//this.MessageCenter.Register_multicast(gconfig.MSGID_multicast_offline_recv, this.MulticastOffline_recv)            //接收节点下线广播
	//this.MessageCenter.Register_multicast(gconfig.MSGID_multicast_vnode_offline_recv, this.MulticastVnodeOffline_recv) //接收虚拟节点下线广播
	//this.MessageCenter.Register_p2p(gconfig.MSGID_checkAddrOnline, this.checkAddrOnline)                               //检查节点是否在线
	//this.MessageCenter.Register_p2p(gconfig.MSGID_checkAddrOnline_recv, this.checkAddrOnline_recv)                     //检查节点是否在线_返回
	//this.MessageCenter.Register_p2p(gconfig.MsGID_recv_router_err, this.recvRouterErr)
	//
	//this.MessageCenter.Register_p2p(gconfig.MSGID_sync_proxy, this.syncProxyRec)                  // 同步代理处理
	//this.MessageCenter.Register_p2p(gconfig.MSGID_sync_proxy_recv, this.syncProxyRec_recv)        // 同步代理返回
	//this.MessageCenter.Register_p2p(gconfig.MSGID_search_addr_proxy, this.getAddrProxy)           // 查询地址对应的代理信息
	//this.MessageCenter.Register_p2p(gconfig.MSGID_search_addr_proxy_recv, this.getAddrProxy_recv) // 查询地址对应的代理信息_返回

}

/*
查询一个节点是否在线
*/
func (this *Node) SearchNode(message *message_center.MessageBase) {
	//if this.CheckRepeatHash(message.Body.Hash) {
	//	return
	//}
	////回复消息
	//data := []byte(this.NodeManager.NodeSelf.MachineIDStr)
	//this.MessageCenter.SendP2pReplyMsg(message, config.MSGID_search_node_recv, &data)
}

func (this *Node) SearchNode_recv(message *message_center.MessageBase) {
	//if this.CheckRepeatHash(message.Body.ReplyHash) {
	//	return
	//}
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收上线的广播
*/
func (this *Node) MulticastOnline_recv(message *message_center.MessageBase) {
	//newNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	//if err != nil {
	//	return
	//}
	//
	//// 1. 先判断本地是否存在上帝地址
	//if this.GodHost != "" {
	//	// 1.1 存在上帝地址信息，则判断对方地址如果不是上帝地址，则直接return
	//	nodeHost := newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort))
	//	if this.GodHost != nodeHost {
	//		// utils.Log.Warn().Msgf("qlw-----不是上帝节点的地址, 得知节点上线也不予处理!!!")
	//		return
	//	}
	//}
	//
	//if !newNode.IsApp && newNode.GetIsSuper() {
	//	this.ProxyDetail.NodeOnlineDeal(newNode)
	//}

}

/*
获取相邻节点的超级节点地址
*/
func (this *Node) GetNearSuperAddr(message *message_center.MessageBase) {
	// utils.Log.Info().Msgf("GetNearSuperAddr start self:%s", this.GetNetId().B58String())
	//
	//nodes := make([]nodeStore.Node, 0)
	//ns := this.NodeManager.GetLogicNodes()
	//ns = append(ns, this.NodeManager.GetNodesClient()...)
	//ns = append(ns, this.NodeManager.GetOneByOneNodes()...)
	//allUpDown := this.SessionEngine.GetAllDownUp(this.AreaName[:])
	//
	//for i := 0; i < len(allUpDown); i++ {
	//	//utils.Log.Info().Msgf("111GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet([]byte(allUpDown[i])).B58String(), message.Head.Sender.B58String())
	//	ns = append(ns, nodeStore.AddressNet([]byte(allUpDown[i])))
	//}
	////idsm := nodeStore.NewKademlia(*message.Head.Sender, nodeStore.NodeIdLevel)
	//// for _, one := range ns {
	//// 	if bytes.Equal(*message.Head.Sender, one) {
	//// 		continue
	//// 	}
	//// 	utils.Log.Info().Msgf("1122GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), one.B58String(), message.Head.Sender.B58String())
	//// 	//idsm.AddId(one)
	//// }
	////ids := idsm.GetIds()
	//for _, one := range ns {
	//	if bytes.Equal(*message.Head.Sender, one) {
	//		continue
	//	}
	//	addrNet := nodeStore.AddressNet(one)
	//	//utils.Log.Info().Msgf("222GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), message.Head.Sender.B58String())
	//	node := this.NodeManager.FindNode(&addrNet)
	//	if node != nil {
	//		//			utils.Log.Info().Msgf("333GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), message.Head.Sender.B58String())
	//		nodes = append(nodes, *node)
	//	} else {
	//	}
	//}
	//nodeRepeated := go_protobuf.NodeRepeated{
	//	Nodes: make([]*go_protobuf.Node, 0),
	//}
	//for _, one := range nodes {
	//	idinfo := go_protobuf.IdInfo{
	//		Id:   one.IdInfo.Id,
	//		EPuk: one.IdInfo.EPuk,
	//		CPuk: one.IdInfo.CPuk[:],
	//		V:    one.IdInfo.V,
	//		Sign: one.IdInfo.Sign,
	//	}
	//	nodeOne := go_protobuf.Node{
	//		IdInfo:       &idinfo,
	//		IsSuper:      one.GetIsSuper(),
	//		Addr:         one.Addr,
	//		TcpPort:      uint32(one.TcpPort),
	//		IsApp:        one.IsApp,
	//		MachineID:    one.MachineID,
	//		Version:      one.Version,
	//		MachineIDStr: one.MachineIDStr,
	//		QuicPort:     uint32(one.QuicPort),
	//	}
	//	nodeRepeated.Nodes = append(nodeRepeated.Nodes, &nodeOne)
	//}
	//
	////增加自己节点信息
	//idinfoSelf := go_protobuf.IdInfo{
	//	Id:   this.NodeManager.NodeSelf.IdInfo.Id,
	//	EPuk: this.NodeManager.NodeSelf.IdInfo.EPuk,
	//	CPuk: this.NodeManager.NodeSelf.IdInfo.CPuk[:],
	//	V:    this.NodeManager.NodeSelf.IdInfo.V,
	//	Sign: this.NodeManager.NodeSelf.IdInfo.Sign,
	//}
	//nodeRepeated.Nodes = append(nodeRepeated.Nodes, &go_protobuf.Node{
	//	IdInfo:       &idinfoSelf,
	//	IsSuper:      this.NodeManager.NodeSelf.GetIsSuper(),
	//	Addr:         this.NodeManager.NodeSelf.Addr,
	//	TcpPort:      uint32(this.NodeManager.NodeSelf.TcpPort),
	//	IsApp:        this.NodeManager.NodeSelf.IsApp,
	//	MachineID:    this.NodeManager.NodeSelf.MachineID,
	//	Version:      this.NodeManager.NodeSelf.Version,
	//	MachineIDStr: this.NodeManager.NodeSelf.MachineIDStr,
	//	QuicPort:     uint32(this.NodeManager.NodeSelf.QuicPort),
	//})
	//
	//data, _ := nodeRepeated.Marshal()
	//this.MessageCenter.SendNeighborReplyMsg(message, gconfig.MSGID_getNearSuperIP_recv, &data, msg.Session)
	//// utils.Log.Info().Msgf("GetNearSuperAddr end self:%s", this.GetNetId().B58String())
}

/*
获取相邻节点的超级节点地址返回
*/
func (this *Node) GetNearSuperAddr_recv(message *message_center.MessageBase) {
	//// utils.Log.Info().Msgf("GetNearSuperAddr_recv start self:%s target:%s", this.GetNetId().B58String(), message.Head.Sender.B58String())
	//addrNet := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	//node := this.NodeManager.FindNode(&addrNet)
	//if node == nil {
	//	return
	//} else {
	//	if node.Type == nodeStore.Node_type_proxy {
	//		return
	//	}
	//}
	//ok := flood.GroupWaitRecv.ResponseBytes(strconv.Itoa(int(msg.Session.GetIndex())), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	//// 如果设置了超级代理节点，跳过recvNearLogicNodes
	//var godAutonomyFinish bool
	//if this.GodHost != "" {
	//	allSessions := this.SessionEngine.GetAllSession(this.AreaName[:])
	//	for i := range allSessions {
	//		if allSessions[i].GetRemoteHost() != this.GodHost {
	//			continue
	//		}
	//
	//		godAutonomyFinish = true
	//		break
	//	}
	//}
	//if !ok && !godAutonomyFinish {
	//	//如果没有监听，则是邻居节点主动推送
	//	this.recvNearLogicNodes(message.Body.Content, message.Head.Sender)
	//}
	//// utils.Log.Info().Msgf("GetNearSuperAddr_recv end self:%s", this.GetNetId().B58String())
}

/*
接收邻居节点的逻辑节点和虚拟节点并添加
*/
func (this *Node) recvNearLogicNodes(bs *[]byte, recvid *nodeStore.AddressNet) {
	//// this.GetNearSuperAddr_recvLock.Lock()
	//// defer this.GetNearSuperAddr_recvLock.Unlock()
	//nodes, err := nodeStore.ParseNodesProto(bs)
	//if err != nil {
	//	return
	//}
	//
	//for i, _ := range nodes {
	//	newNode := nodes[i]
	//	//utils.Log.Info().Msgf(" 收到self:%s  comein %s", this.GetNetId().B58String(), newNode.IdInfo.Id.B58String())
	//	connectingKeyQuic := newNode.Addr + ":" + strconv.Itoa(int(newNode.QuicPort))
	//	_, ok := this.connecting.LoadOrStore(connectingKeyQuic, 0)
	//	if ok {
	//		continue
	//	}
	//
	//	connectingKeyTcp := newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort))
	//	if newNode.TcpPort != newNode.QuicPort {
	//		_, ok = this.connecting.LoadOrStore(connectingKeyTcp, 0)
	//		if ok {
	//			continue
	//		}
	//	}
	//
	//	if this.NodeManager.IsHeadTailModl {
	//		//utils.Log.Info().Msgf("使用头尾模式进行判断 : %s", this.GetNetId().B58String())
	//		this.onebyoneRecvNodeHeadTail(newNode)
	//	} else {
	//		//utils.Log.Info().Msgf("使用普通模式进行判断 : %s", this.GetNetId().B58String())
	//		this.oneByoneRecvNode(newNode)
	//	}
	//	this.connecting.Delete(connectingKeyQuic)
	//	if newNode.TcpPort != newNode.QuicPort {
	//		this.connecting.Delete(connectingKeyTcp)
	//	}
	//}
	//
	//bv, err := this.MessageCenter.SendNeighborMsgWaitRequest(gconfig.MSGID_getNearConnectVnode, recvid, nil, time.Second*8)
	//if err != nil {
	//	//utils.Log.Warn().Msgf("MSGID_getNearConnectVnode error:%s self:%s target:%s ", err.Error(), this.GetNetId().B58String(), recvid.B58String())
	//	return
	//}
	//node := this.NodeManager.FindNode(recvid)
	//if node == nil {
	//	return
	//}
	//this.recvNearVnodes(bv, node)
}

/*
询问关闭这个链接
当双方都没有这个链接的引用时，就关闭这个链接
*/
func (this *Node) AskCloseConn_recv(message *message_center.MessageBase) {
	//mh := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	//// utils.Log.Info().Msgf("AdkClose1:%s", mh.B58String())
	//// utils.Log.Info().Msgf("AdkClose2:%v", this.NodeManager.FindNodeInLogic(&mh))
	//// utils.Log.Info().Msgf("AdkClose3:%t", this.NodeManager.FindWhiteList(&mh))
	//if this.NodeManager.FindNodeInLogic(&mh) == nil && !this.NodeManager.FindWhiteList(&mh) {
	//	//自己也没有这个连接的引用，则关闭这个链接
	//	// utils.Log.Info().Msgf("Close this session:%s", mh.B58String())
	//	msg.Session.Close()
	//}
}

/*
 * 接收真实节点下线的广播
 */
func (this *Node) MulticastOffline_recv(message *message_center.MessageBase) {
	//// utils.Log.Info().Msgf("qlw---MulticastOffline_recv: 接收到消息")
	//if message.Body.Content == nil || len(*message.Body.Content) == 0 {
	//	return
	//}
	//
	//// 解析下线节点信息
	//offlineNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	//if err != nil {
	//	return
	//}
	//
	//// 1. 获取需要查询的节点真实地址信息
	//nid := offlineNode.IdInfo.Id
	//// utils.Log.Error().Msgf("下线节点地址为: %s machineId:%s", nid.B58String(), offlineNode.MachineIDStr)
	//
	//// 2. 判断自己是不是该节点，如果是，则直接返回
	//if bytes.Equal(nid, this.Vm.DiscoverVnodes.Vnode.Nid) {
	//	return
	//}
	//
	//// 3. 检查自己是否含有该节点，如果不存在则直接返回
	//var bNeedDealNode bool
	//if this.Vm.CheckNodeinfoExistInSelf(nid) {
	//	bNeedDealNode = true
	//}
	//
	//// 4. 判断是否需要处理下线的客户端
	//var bNeedDealProxy bool
	//if this.ProxyDetail.CheckIsSaveAddr(&nid) {
	//	bNeedDealProxy = true
	//}
	//
	//// 5. 如果即不是自己的逻辑节点, 也不是代理的存储节点, 则不用处理
	//if !bNeedDealNode && !bNeedDealProxy {
	//	return
	//}
	//
	//// 6. 存在的话，先尝试询问对象是否在线
	//_, _, _, err = this.SendP2pMsgWaitRequest(gconfig.MSGID_checkAddrOnline, &nid, nil, 3*time.Second)
	//// 6.1 在线，则直接返回，无需操作
	//if err == nil {
	//	return
	//}
	//
	//// 不在线，进行相应的处理
	//// 7. 处理逻辑节点相关操作
	//if bNeedDealNode {
	//	// 7.1 删除发现节点的对应信息
	//	this.Vm.DiscoverVnodes.DeleteNid(nid)
	//	// 7.2 删除虚拟节点的对应信息
	//	this.Vm.RLock()
	//	for _, one := range this.Vm.VnodeMap {
	//		one.DeleteNid(nid)
	//	}
	//	this.Vm.RUnlock()
	//	// 7.3. 删除连接对应的加密管道
	//	if this.MessageCenter != nil && this.MessageCenter.RatchetSession != nil {
	//		this.MessageCenter.RatchetSession.RemoveSendPipe(nid, offlineNode.MachineIDStr)
	//	}
	//
	//	// 7.4 触发查找新的节点操作
	//	this.findNearNodeTimer.Release()
	//}
	//
	//// 8. 处理代理的存储节点操作
	//if bNeedDealProxy {
	//	// 删除对应的存储节点信息
	//	this.ProxyDetail.NodeOfflineDeal(&nid)
	//}

	// utils.Log.Info().Msgf("处理节点下线成功: nodeid:%s, self:%s", nid.B58String(), this.Vm.DiscoverVnodes.Vnode.Vid.B58String())
}

/*
 * 接收虚拟节点下线的广播
 */
func (this *Node) MulticastVnodeOffline_recv(message *message_center.MessageBase) {
	//// 1. 获取需要查询的虚拟节点信息
	//vnodeInfo, err := virtual_node.ParseVnodeinfo(*message.Body.Content)
	//if err != nil {
	//	// 解析失败，直接返回，不做处理
	//	// utils.Log.Info().Msgf("parse vnode info err:%s", err)
	//	return
	//}
	//// 1.1 检查虚拟节点参数的有效性
	//if vnodeInfo == nil || vnodeInfo.Nid == nil || len(vnodeInfo.Nid) == 0 || vnodeInfo.Vid == nil || len(vnodeInfo.Vid) == 0 {
	//	// utils.Log.Info().Msgf("recv vnode info incorrect")
	//	return
	//}
	//
	//// 2. 检查自己是否含有该节点，如果不存在则直接返回
	//if this.Vm.FindVnodeinfo(vnodeInfo.Vid) == nil {
	//	return
	//}
	//
	//// 3. 存在的话，先尝试询问对象是否在线
	//if this.Vc.CheckVnodeIsOnline(vnodeInfo) {
	//	// 3.1 在线，则直接返回，无需操作
	//	return
	//}
	//// 3.2 不在线，删除对应的虚拟节点信息
	//// 3.2.1 删除发现节点中的虚拟节点信息
	//this.Vm.DiscoverVnodes.DeleteVid(vnodeInfo)
	//// 3.2.2 删除虚拟节点中的虚拟节点信息
	//this.Vm.RLock()
	//for _, one := range this.Vm.VnodeMap {
	//	one.DeleteVid(vnodeInfo)
	//}
	//this.Vm.RUnlock()
	//
	//// 4. 触发查找新的节点操作
	//this.findNearNodeTimer.Release()
}

/*
 * 查询一个地址是否在线
 */
func (this *Node) checkAddrOnline(message *message_center.MessageBase) {
	//this.MessageCenter.SendSearchSuperReplyMsg(message, gconfig.MSGID_checkAddrOnline_recv, nil)
}

/*
 * 查询一个地址是否在线的返回
 */
func (this *Node) checkAddrOnline_recv(message *message_center.MessageBase) {
	//flood.ResponseBytes(utils.Bytes2string(*&message.Body.Hash), nil)
}

/*
 * 接受中转错误返回
 */
func (this *Node) recvRouterErr(message *message_center.MessageBase) {
	//utils.Log.Info().Msgf("recvRouterErr recvRouterErr recvRouterErr")
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
 * 接收同步代理信息
 */
func (this *Node) syncProxyRec(message *message_center.MessageBase) {
	//// utils.Log.Info().Msgf("syncProxyRec start self:%s from:%s", this.GetNetId().B58String(), message.Head.Sender.B58String())
	//// 1. 判断参数的合法性
	//if message.Body == nil || message.Body.Content == nil || message.Head.Sender == nil {
	//	utils.Log.Error().Msgf("参数的不合法!!!!!!")
	//	return
	//}
	//
	//// 2. 解析代理信息
	//proxyInfoes, err := proxyrec.ParseProxyesProto(message.Body.Content)
	//if err != nil {
	//	utils.Log.Error().Msgf("代理信息解析失败!!!!!!")
	//	return
	//}
	//
	//// 3. 处理代理信息
	//this.ProxyData.AddOrUpdateProxyRec(proxyInfoes, message.Head.Sender, message.Head.RecvId)
	//
	//// 4. 返回对应的消息
	//this.SendP2pReplyMsg(message, gconfig.MSGID_sync_proxy_recv, nil)
}

/*
 * 接收同步代理信息的返回
 */
func (this *Node) syncProxyRec_recv(message *message_center.MessageBase) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
 * 获取地址对应的代理信息
 */
func (this *Node) getAddrProxy(message *message_center.MessageBase) {
	//// utils.Log.Info().Msgf("getAddrProxy start self:%s from:%s", this.GetNetId().B58String(), message.Head.Sender.B58String())
	//// 1. 判断参数的合法性
	//if message.Body == nil || message.Body.Content == nil || message.Head.Sender == nil {
	//	utils.Log.Error().Msgf("参数的不合法!!!!!!")
	//	this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
	//	return
	//}
	//
	//// 2. 解析代理信息
	//proxyInfo, err := proxyrec.ParseProxyProto(message.Body.Content)
	//if err != nil {
	//	utils.Log.Error().Msgf("代理信息解析失败!!!!!!")
	//	this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
	//	return
	//}
	//
	//// 3. 处理代理信息
	//exist, proxyInfoes := this.ProxyData.GetNodeIdProxy(proxyInfo)
	//if !exist || len(proxyInfoes) == 0 {
	//	// utils.Log.Warn().Msgf("不存在代理信息!!!! clientId:%s machineId:%s", proxyInfo.NodeId.B58String(), proxyInfo.MachineId)
	//	this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
	//	return
	//}
	//
	//// 4. 返回对应的消息
	//var res go_protobuf.ProxyRepeated
	//res.Proxys = make([]*go_protobuf.ProxyInfo, 0)
	//for i := range proxyInfoes {
	//	var proxy go_protobuf.ProxyInfo
	//	proxy.Id = *proxyInfoes[i].NodeId
	//	proxy.ProxyId = *proxyInfoes[i].ProxyId
	//	proxy.MachineID = proxyInfoes[i].MachineId
	//	proxy.Version = proxyInfoes[i].Version
	//
	//	res.Proxys = append(res.Proxys, &proxy)
	//}
	//bs, err := res.Marshal()
	//if err != nil {
	//	utils.Log.Error().Msgf("ProxyInfo marshal err:%s", err)
	//	this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
	//	return
	//}
	//this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, &bs)
}

/*
 * 获取地址对应的代理信息的返回
 */
func (this *Node) getAddrProxy_recv(message *message_center.MessageBase) {
	//flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
 * 指定连接方式进行连接
 */
func (this *Node) connNetByMethod(ip string, quicPort, tcpPort uint32, powerful bool, onlyAddOneSess string) (ss engine.Session, err error) {
	//var bConnSuccess bool
	//// 先尝试quic连接
	//if this.CheckSupportQuicConn() {
	//	_, err = this.SessionEngine.AddClientQuicConn(this.AreaName[:], ip, quicPort, powerful, onlyAddOneSess)
	//	if err == nil {
	//		bConnSuccess = true
	//	}
	//}
	//
	//// 再尝试tcp连接
	//if !bConnSuccess && this.CheckSupportTcpConn() {
	//	_, err = this.SessionEngine.AddClientConn(this.AreaName[:], ip, tcpPort, powerful, onlyAddOneSess)
	//	if err == nil {
	//		bConnSuccess = true
	//	}
	//}
	//
	//if !bConnSuccess && err == nil {
	//	err = errors.New("无法进行任何连接处理")
	//}
	//
	return
}
