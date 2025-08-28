package libp2parea

import (
	"bytes"
	"errors"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"web3_gui/libp2parea/v1/config"
	gconfig "web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	proxyrec "web3_gui/libp2parea/v1/proxy_rec"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

type MessageVO struct {
	Name     string //消息记录name
	Id       string //发送消息者id
	Index    int64  //unix时间排序
	Time     string //接收时间
	Content  string //消息内容
	Path     string //图片、文件路径
	FileName string //文件名
	Size     int64  //文件大小
	Cate     int    //消息类型
	DBID     int64  //数据量id
}

func (this *Area) RegisterCoreMsg() {
	this.MessageCenter.RegisterMsgVersion()

	this.MessageCenter.Register_search_super(gconfig.MSGID_checkNodeOnline, this.findSuperID, true)                    //检查节点是否在线
	this.MessageCenter.Register_p2p(gconfig.MSGID_checkNodeOnline_recv, this.findSuperID_recv, true)                   //检查节点是否在线_返回
	this.MessageCenter.Register_neighbor(gconfig.MSGID_getNearSuperIP, this.GetNearSuperAddr, true)                    //从邻居节点得到自己的逻辑节点
	this.MessageCenter.Register_neighbor(gconfig.MSGID_getNearSuperIP_recv, this.GetNearSuperAddr_recv, true)          //从邻居节点得到自己的逻辑节点_返回
	this.MessageCenter.Register_neighbor(gconfig.MSGID_getNearConnectVnode, this.GetNearConnectVnode, true)            //从邻居节点得到自己连接的虚拟节点
	this.MessageCenter.Register_neighbor(gconfig.MSGID_getNearConnectVnode_recv, this.GetNearConnectVnode_recv, true)  //从邻居节点得到自己连接的虚拟节点
	this.MessageCenter.Register_multicast(gconfig.MSGID_multicast_online_recv, this.MulticastOnline_recv, true)        //接收节点上线广播
	this.MessageCenter.Register_neighbor(gconfig.MSGID_ask_close_conn_recv, this.AskCloseConn_recv, true)              //询问关闭连接
	this.MessageCenter.Register_neighbor(gconfig.MSGID_del_vnode, this.DelVnodeInSelf, true)                           //删除本节点存的up down虚拟节点
	this.MessageCenter.Register_p2p(gconfig.MSGID_SearchAddr, this.MessageCenter.SearchAddress, true)                  //搜索节点，返回节点真实地址
	this.MessageCenter.Register_p2p(gconfig.MSGID_SearchAddr_recv, this.MessageCenter.SearchAddress_recv, true)        //搜索节点，返回节点真实地址_返回
	this.MessageCenter.Register_p2p(gconfig.MSGID_security_create_pipe, this.MessageCenter.CreatePipe, true)           //创建加密通道
	this.MessageCenter.Register_p2p(gconfig.MSGID_security_create_pipe_recv, this.MessageCenter.CreatePipe_recv, true) //创建加密通道_返回
	this.MessageCenter.Register_p2pHE(gconfig.MSGID_security_pipe_error, this.MessageCenter.Pipe_error, true)          //解密错误
	this.MessageCenter.Register_p2p(gconfig.MSGID_search_node, this.SearchNode, true)                                  //查询一个节点是否在线
	this.MessageCenter.Register_p2p(gconfig.MSGID_search_node_recv, this.SearchNode_recv, true)                        //查询一个节点是否在线_返回

	this.MessageCenter.Register_multicast(gconfig.MSGID_multicast_offline_recv, this.MulticastOffline_recv, true)            //接收节点下线广播
	this.MessageCenter.Register_multicast(gconfig.MSGID_multicast_vnode_offline_recv, this.MulticastVnodeOffline_recv, true) //接收虚拟节点下线广播
	this.MessageCenter.Register_p2p(gconfig.MSGID_checkAddrOnline, this.checkAddrOnline, true)                               //检查节点是否在线
	this.MessageCenter.Register_p2p(gconfig.MSGID_checkAddrOnline_recv, this.checkAddrOnline_recv, true)                     //检查节点是否在线_返回
	this.MessageCenter.Register_p2p(gconfig.MSGID_checkVnodeAddrOnline, this.checkVnodeAddrOnline, true)                     //检查虚拟节点是否在线
	this.MessageCenter.Register_p2p(gconfig.MSGID_checkVnodeAddrOnline_recv, this.checkVnodeAddrOnline_recv, true)           //检查虚拟节点是否在线_返回
	this.MessageCenter.Register_p2p(gconfig.MsGID_recv_router_err, this.recvRouterErr, true)

	this.MessageCenter.Register_p2p(gconfig.MSGID_sync_proxy, this.syncProxyRec, true)                  // 同步代理处理
	this.MessageCenter.Register_p2p(gconfig.MSGID_sync_proxy_recv, this.syncProxyRec_recv, true)        // 同步代理返回
	this.MessageCenter.Register_p2p(gconfig.MSGID_search_addr_proxy, this.getAddrProxy, true)           // 查询地址对应的代理信息
	this.MessageCenter.Register_p2p(gconfig.MSGID_search_addr_proxy_recv, this.getAddrProxy_recv, true) // 查询地址对应的代理信息_返回

	this.MessageCenter.Register_multicast(gconfig.MSGID_multicast_send_vnode_recv, this.MulticastSendVnode_recv, true) // 发送本节点的所有虚拟节点信息广播
}

/*
查询一个节点是否在线
*/
func (this *Area) SearchNode(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if this.CheckRepeatHash(message.Body.Hash) {
		return
	}
	//回复消息
	data := []byte(this.NodeManager.NodeSelf.MachineIDStr)
	this.MessageCenter.SendP2pReplyMsg(message, config.MSGID_search_node_recv, &data)
}

func (this *Area) SearchNode_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if this.CheckRepeatHash(message.Body.ReplyHash) {
		return
	}
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
接收上线的广播
*/
func (this *Area) MulticastOnline_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	newNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	if err != nil {
		return
	}

	// 1. 先判断本地是否存在上帝地址
	if this.GodHost != "" {
		// 1.1 存在上帝地址信息，则判断对方地址如果不是上帝地址，则直接return
		nodeHost := newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort))
		if this.GodHost != nodeHost {
			// utils.Log.Warn().Msgf("qlw-----不是上帝节点的地址, 得知节点上线也不予处理!!!")
			return
		}
	}

	if !newNode.IsApp && newNode.GetIsSuper() {
		this.ProxyDetail.NodeOnlineDeal(newNode)
	}

	// this.Vc.AddNewNode(newNode.IdInfo.Id)
	//查询是否已经有这个连接了，有了就不连接
	//避免地址和node真实地址不对应的情况
	// session := this.SessionEngine.GetSessionByHost(newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort)))
	// if session != nil {
	// 	return
	// }
	//TODO 当真实P2P网络地址和ip端口不对应的情况，容易漏连接某些节点
	//检查是否需要这个逻辑节点
	// ok, _ := this.NodeManager.CheckNeedNode(&newNode.IdInfo.Id)
	// if !ok {
	// 	//		fmt.Println("不需要这个逻辑节点")
	// 	return
	// }
	// //	fmt.Println("需要这个节点")
	// // utils.Log.Info().Msgf("GetNearSuperAddr_recv need: %s", newNode.IdInfo.Id.B58String())
	// //检查是否有这个连接
	// // _, ok = engine.GetSession(newNode.IdInfo.Id.B58String())
	// _, ok = this.SessionEngine.GetSession(utils.Bytes2string(newNode.IdInfo.Id))
	// if !ok {
	// 	//查询是否已经有这个连接了，有了就不连接
	// 	//避免地址和node真实地址不对应的情况
	// 	_, err := this.SessionEngine.AddClientConn(this.AreaName[:], newNode.Addr, uint32(newNode.TcpPort), false, engine.BothMod)
	// 	if err != nil {
	// 		return
	// 	}
	// } else {
	// }
	// //非超级节点判断超级节点是否改变
	// if !this.NodeManager.NodeSelf.GetIsSuper() {
	// 	nearId := this.NodeManager.FindNearInSuper(&this.NodeManager.NodeSelf.IdInfo.Id, nil, false, nil)
	// 	if nearId == nil || bytes.Equal(*nearId, *this.NodeManager.GetSuperPeerId()) {
	// 		return
	// 	}
	// 	// utils.Log.Info().Msgf("更换超级节点id:%s", nearId.B58String())
	// 	this.NodeManager.SetSuperPeerId(nearId)
	// }
}

/*
查询一个id最近的超级节点id
*/
func (this *Area) findSuperID(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	data, _ := this.NodeManager.NodeSelf.Proto()
	this.MessageCenter.SendSearchSuperReplyMsg(message, gconfig.MSGID_checkNodeOnline_recv, &data)

}

func (this *Area) findSuperID_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	newNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	if err != nil {
		return
	}
	node := this.NodeManager.FindNode(&newNode.IdInfo.Id)
	node.FlashOnlineTime()
}

/*
获取相邻节点的超级节点地址
*/
func (this *Area) GetNearSuperAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("GetNearSuperAddr start self:%s", this.GetNetId().B58String())

	nodes := make([]nodeStore.Node, 0)
	ns := this.NodeManager.GetLogicNodes()
	ns = append(ns, this.NodeManager.GetNodesClient()...)
	ns = append(ns, this.NodeManager.GetOneByOneNodes()...)
	allUpDown := this.SessionEngine.GetAllDownUp(this.AreaName[:])

	for i := 0; i < len(allUpDown); i++ {
		//utils.Log.Info().Msgf("111GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet([]byte(allUpDown[i])).B58String(), message.Head.Sender.B58String())
		ns = append(ns, nodeStore.AddressNet([]byte(allUpDown[i])))
	}
	//idsm := nodeStore.NewIds(*message.Head.Sender, nodeStore.NodeIdLevel)
	// for _, one := range ns {
	// 	if bytes.Equal(*message.Head.Sender, one) {
	// 		continue
	// 	}
	// 	utils.Log.Info().Msgf("1122GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), one.B58String(), message.Head.Sender.B58String())
	// 	//idsm.AddId(one)
	// }
	//ids := idsm.GetIds()
	for _, one := range ns {
		if bytes.Equal(*message.Head.Sender, one) {
			continue
		}
		addrNet := nodeStore.AddressNet(one)
		//utils.Log.Info().Msgf("222GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), message.Head.Sender.B58String())
		node := this.NodeManager.FindNode(&addrNet)
		if node != nil {
			//			utils.Log.Info().Msgf("333GetNearSuperAddr self :%s UpDown :%s to :%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), addrNet.B58String(), message.Head.Sender.B58String())
			nodes = append(nodes, *node)
		} else {
		}
	}
	nodeRepeated := go_protobuf.NodeRepeated{
		Nodes: make([]*go_protobuf.Node, 0),
	}
	for _, one := range nodes {
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

	data, _ := nodeRepeated.Marshal()
	this.MessageCenter.SendNeighborReplyMsg(message, gconfig.MSGID_getNearSuperIP_recv, &data, msg.Session)
	// utils.Log.Info().Msgf("GetNearSuperAddr end self:%s", this.GetNetId().B58String())
}

func (this *Area) GetNearConnectVnode(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	vrs := go_protobuf.VnodeinfoSRepeated{
		Vnodes: make([]*go_protobuf.VnodeinfoS, 0),
	}

	this.Vm.RLock()
	for _, one := range this.Vm.VnodeMap {
		//返回index为0的虚拟节点的up down Vnode
		if one.Vnode.Index == 0 {
			one.Lock.RLock()
			zVnode := one.GetDownVnodeInfo()
			zVnode = append(zVnode, one.GetUpVnodeInfo()...)
			one.Lock.RUnlock()
			for i := 0; i < len(zVnode); i++ {
				if bytes.Equal(*message.Head.Sender, zVnode[i].Nid) {
					continue
				}
				vrs.Vnodes = append(vrs.Vnodes, &go_protobuf.VnodeinfoS{
					Nid:      zVnode[i].Nid,
					Vid:      zVnode[i].Vid,
					Index:    zVnode[i].Index,
					Addr:     zVnode[i].Addr,
					TcpPort:  zVnode[i].TcpPort,
					QuicPort: zVnode[i].QuicPort,
				})
			}
			continue
		}
		vnodeifos := one.GetOnebyoneVnodeInfo()
		//添加本虚拟节点
		vrs.Vnodes = append(vrs.Vnodes, &go_protobuf.VnodeinfoS{
			Nid:      one.Vnode.Nid,
			Vid:      one.Vnode.Vid,
			Index:    one.Vnode.Index,
			Addr:     this.NodeManager.NodeSelf.Addr,
			TcpPort:  uint64(this.NodeManager.NodeSelf.TcpPort),
			QuicPort: uint64(this.NodeManager.NodeSelf.QuicPort),
		})
		for i := 0; i < len(vnodeifos); i++ {
			if bytes.Equal(*message.Head.Sender, vnodeifos[i].Nid) {
				continue
			}
			//utils.Log.Info().Msgf("GetNearConnectVnode_ addr:%s port:%d", vnodeifos[i].Addr, vnodeifos[i].TcpPort)
			vrs.Vnodes = append(vrs.Vnodes, &go_protobuf.VnodeinfoS{
				Nid:      vnodeifos[i].Nid,
				Vid:      vnodeifos[i].Vid,
				Index:    vnodeifos[i].Index,
				Addr:     vnodeifos[i].Addr,
				TcpPort:  vnodeifos[i].TcpPort,
				QuicPort: vnodeifos[i].QuicPort,
			})
		}
	}
	this.Vm.RUnlock()

	vs, err := vrs.Marshal()
	if err != nil {
		utils.Log.Warn().Msgf("GetNearConnectVnode vnodes proto err:%s", err)
		return
	}
	this.MessageCenter.SendNeighborReplyMsg(message, gconfig.MSGID_getNearConnectVnode_recv, &vs, msg.Session)
}

/*
获取相邻节点的超级节点地址返回
*/
func (this *Area) GetNearSuperAddr_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("GetNearSuperAddr_recv start self:%s target:%s", this.GetNetId().B58String(), message.Head.Sender.B58String())
	addrNet := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	node := this.NodeManager.FindNode(&addrNet)
	if node == nil {
		return
	} else {
		if node.Type == nodeStore.Node_type_proxy {
			return
		}
	}
	ok := flood.GroupWaitRecv.ResponseBytes(strconv.Itoa(int(msg.Session.GetIndex())), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	// 如果设置了超级代理节点，跳过recvNearLogicNodes
	var godAutonomyFinish bool
	if this.GodHost != "" {
		allSessions := this.SessionEngine.GetAllSession(this.AreaName[:])
		for i := range allSessions {
			if allSessions[i].GetRemoteHost() != this.GodHost {
				continue
			}

			godAutonomyFinish = true
			break
		}
	}
	if !ok && !godAutonomyFinish {
		//如果没有监听，则是邻居节点主动推送
		this.recvNearLogicNodes(message.Body.Content, message.Head.Sender)
	}
	// utils.Log.Info().Msgf("GetNearSuperAddr_recv end self:%s", this.GetNetId().B58String())
}

func (this *Area) GetNearConnectVnode_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	ok := flood.GroupWaitRecv.ResponseBytes(strconv.Itoa(int(msg.Session.GetIndex())), utils.Bytes2string(message.Body.Hash), message.Body.Content)
	if !ok {
		//如果没有监听，则是邻居节点主动推送
		//this.recvNearLogicNodes(message.Body.Content)
	}
}

/*
 * 首尾模式下的session控制
 * config.IsHeadTailModl参数控制
 * 真实节点onebyone规则首尾模式的节点连接控制
 */
func (this *Area) onebyoneRecvNodeHeadTail(newNode nodeStore.Node) {
	var oneByoneNeed bool
	var kadNeed bool
	var delSess []engine.Session
	var allS []engine.Session
	oneByoneNeed = true

	if bytes.Equal(newNode.IdInfo.Id, this.NodeManager.NodeSelf.IdInfo.Id) {
		return
	}
	//utils.Log.Info().Msgf(" 111self:%s  comein %s", this.GetNetId().B58String(), newNode.IdInfo.Id.B58String())
	allDownSess := this.SessionEngine.GetAllDownSession(this.AreaName[:])
	allUpSess := this.SessionEngine.GetAllUpSession(this.AreaName[:])
	allS = append(allS, allDownSess...)
	allS = append(allS, allUpSess...)
	for i := 0; i < len(allS); i++ {
		if bytes.Equal(nodeStore.AddressNet(allS[i].GetName()), newNode.IdInfo.Id) {
			return
		}
	}

	if len(allDownSess) >= config.GreaterThanSelfMaxConn/2 && len(allUpSess) >= config.GreaterThanSelfMaxConn/2 {
		//utils.Log.Info().Msgf("模式转换回去", this.NodeManager.NodeSelf.IdInfo.Id.B58String())
		this.NodeManager.IsHeadTailModl = false
		return
	}

	if len(allDownSess) >= len(allUpSess) {
		//本节点比发现的节点大
		//在down比up多的时候，如果是大于本节点的，直接去连，反之本节点比newnode大就需处理
		if new(big.Int).SetBytes(this.GetNetId()).Cmp(new(big.Int).SetBytes([]byte(newNode.IdInfo.Id))) == 1 {
			//取偶数方便后面计算
			downCap := (config.GreaterThanSelfMaxConn + config.GreaterThanSelfMaxConn - len(allUpSess)) / 2 * 2
			//只有allDownSess数量大于等于downCap数量时候才做插入并删除一个的操作，没到就直接添加
			if downCap <= len(allDownSess) {
				sortedSes, _ := message_center.GetSortSession(allDownSess, this.GetNetId())
				if new(big.Int).SetBytes(newNode.IdInfo.Id).Cmp(new(big.Int).SetBytes([]byte(sortedSes[len(allDownSess)-1].GetName()))) == -1 {
					delSess = append(delSess, sortedSes[len(allDownSess)-1])
				} else if new(big.Int).SetBytes(newNode.IdInfo.Id).Cmp(new(big.Int).SetBytes([]byte(sortedSes[downCap/2].GetName()))) == -1 {
					delSess = append(delSess, sortedSes[downCap/2])
				}
				if new(big.Int).SetBytes(newNode.IdInfo.Id).Cmp(new(big.Int).SetBytes([]byte(sortedSes[downCap/2-1].GetName()))) == 1 {
					delSess = append(delSess, sortedSes[downCap/2-1])
				}

				if len(delSess) == 0 {
					oneByoneNeed = false
				}
				if len(allDownSess) > downCap && len(delSess) == 0 {

					del := len(allDownSess) - downCap
					//utils.Log.Info().Msgf("***************", downCap, del, delSess)
					for i := 0; i < del; i++ {
						delSess = append(delSess, sortedSes[downCap/2+i])
					}
				}
			}
		}
	} else {
		//本节点比发现的节点小
		//在up比down多的时候，如果是小于本节点的，直接去连，反之本节点比newnode小就需处理
		if new(big.Int).SetBytes(this.GetNetId()).Cmp(new(big.Int).SetBytes([]byte(newNode.IdInfo.Id))) == -1 {
			//取偶数方便后面计算
			upCap := (config.GreaterThanSelfMaxConn + config.GreaterThanSelfMaxConn - len(allDownSess)) / 2 * 2
			//只有allUpSess数量大于等于upCap数量时候才做插入并删除一个的操作，没到就直接添加
			if upCap <= len(allUpSess) {
				sortedSes, _ := message_center.GetSortSession(allUpSess, this.GetNetId())
				if new(big.Int).SetBytes(newNode.IdInfo.Id).Cmp(new(big.Int).SetBytes([]byte(sortedSes[0].GetName()))) == 1 {
					delSess = append(delSess, sortedSes[0])
				} else if new(big.Int).SetBytes(newNode.IdInfo.Id).Cmp(new(big.Int).SetBytes([]byte(sortedSes[upCap/2-1].GetName()))) == 1 {
					delSess = append(delSess, sortedSes[upCap/2-1])
				}
				if new(big.Int).SetBytes(newNode.IdInfo.Id).Cmp(new(big.Int).SetBytes([]byte(sortedSes[upCap/2].GetName()))) == -1 {
					delSess = append(delSess, sortedSes[upCap/2])
				}
				if len(delSess) == 0 {
					oneByoneNeed = false
				}
				if len(allUpSess) > upCap && len(delSess) == 0 {
					del := len(allUpSess) - upCap
					//utils.Log.Info().Msgf("***************", upCap, del, delSess)
					for i := 0; i < del; i++ {
						delSess = append(delSess, sortedSes[upCap/2+i])
					}
				}
			}
		}
	}

	if len(delSess) != 0 {
		//utils.Log.Info().Msgf("删除特定选中节点 %s -> %s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet([]byte(delSess[0].GetName())).B58String())
		for i := 0; i < len(delSess); i++ {
			if this.NodeManager.IsLogicNode(nodeStore.AddressNet(delSess[i].GetName())) ||
				this.Vm.IsSelfVnodeNeed(nodeStore.AddressNet(delSess[i].GetName())) {
				if len(allDownSess) >= len(allUpSess) {
					this.SessionEngine.DelNodeDownSession(this.AreaName[:], delSess[i])
				} else {
					this.SessionEngine.DelNodeUpSession(this.AreaName[:], delSess[i])
				}
			} else {
				this.SessionEngine.RemoveSession(this.AreaName[:], delSess[i])
				delSess[i].Close()
			}
		}
	}

	ok, _ := this.NodeManager.CheckNeedNode(&newNode.IdInfo.Id)
	if ok {
		kadNeed = true
	}
	// node := this.NodeManager.FindNode(&newNode.IdInfo.Id)
	// if node != nil {
	// 	kadNeed = false
	// }
	if Session, ok := this.SessionEngine.GetSessionAll(this.AreaName[:], utils.Bytes2string(newNode.IdInfo.Id)); ok {
		if len(Session) != 0 {
			if oneByoneNeed {
				if this.SessionEngine.CompareAddressNet(engine.AddressNet(newNode.IdInfo.Id)) {
					this.SessionEngine.AddDownSession(this.AreaName[:], Session[0])
				} else {
					this.SessionEngine.AddUpSession(this.AreaName[:], Session[0])
				}
				return
			}
			kadNeed = false
		}
	}

	var err error
	if kadNeed && oneByoneNeed {
		//节点同时符合kad协议和真实地址onebyone协议
		_, err = this.connNetByMethod(newNode.Addr, uint32(newNode.QuicPort), uint32(newNode.TcpPort), false, engine.BothMod)
	} else if !kadNeed && oneByoneNeed {
		//节点不符合kad协议 只符合真实地址onebyone协议
		_, err = this.connNetByMethod(newNode.Addr, uint32(newNode.QuicPort), uint32(newNode.TcpPort), false, engine.OnebyoneMod)
	} else if kadNeed && !oneByoneNeed {
		//节点只符合kad协议 不符合真实地址onebyone协议
		_, err = this.connNetByMethod(newNode.Addr, uint32(newNode.QuicPort), uint32(newNode.TcpPort), false, engine.KadMod)
	}
	if err != nil {
		//utils.Log.Info().Msgf("oneByone connecting_Delete error %s, self:%s", err.Error(), this.NodeManager.NodeSelf.IdInfo.Id.B58String())
		return
	}

	if kadNeed {
		//非超级节点判断超级节点是否改变
		if !this.NodeManager.NodeSelf.GetIsSuper() {
			nearId := this.NodeManager.FindNearInSuper(&this.NodeManager.NodeSelf.IdInfo.Id, nil, false, nil)
			//判断是否需要替换超级节点
			if nearId == nil || bytes.Equal(*nearId, *this.NodeManager.GetSuperPeerId()) {
				return
			}
			// utils.Log.Info().Msgf("Replace supernode id:%s", nearId.B58String())
			this.NodeManager.SetSuperPeerId(nearId)
		}
	}
}

/*
 * 删除node的up/down 多余sesssion使用
 */
func (this *Area) delMoreSess(new []engine.Session, isUpSess bool) {
	if del := len(new) - config.GreaterThanSelfMaxConn; del > 0 {
		sortedSes, _ := message_center.GetSortSession(new, this.GetNetId())
		for i := 0; i < del; i++ {
			var delOne engine.Session
			if isUpSess {
				delOne = sortedSes[i]
			} else {
				delOne = sortedSes[len(new)-1-i]
			}
			if this.NodeManager.IsLogicNode(nodeStore.AddressNet(delOne.GetName())) ||
				this.Vm.IsSelfVnodeNeed(nodeStore.AddressNet(delOne.GetName())) {
				//utils.Log.Info().Msgf("删除多余 self:%s 需要删除:%s %d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet(delOne.GetName()).B58String(), delOne.GetIndex())
				this.SessionEngine.DelNodeDownSession(this.AreaName[:], delOne)
				this.SessionEngine.DelNodeUpSession(this.AreaName[:], delOne)
			} else {
				this.SessionEngine.RemoveSession(this.AreaName[:], delOne)
				//utils.Log.Info().Msgf("删除多余 self:%s 需要删除:%s %d", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet(delOne.GetName()).B58String(), delOne.GetIndex())
				delOne.Close()
			}
		}
	}
}

/*
 * 真实节点onebyone规则正常模式的节点连接控制
 */
func (this *Area) oneByoneRecvNode(newNode nodeStore.Node) {
	var delSess []engine.Session
	var allS []engine.Session
	var oneByoneNeed bool
	var kadNeed bool
	oneByoneNeed = true

	if bytes.Equal(newNode.IdInfo.Id, this.NodeManager.NodeSelf.IdInfo.Id) {
		return
	}

	//utils.Log.Info().Msgf(" 111self:%s  comein %s", this.GetNetId().B58String(), newNode.IdInfo.Id.B58String())
	allDownSess := this.SessionEngine.GetAllDownSession(this.AreaName[:])
	allUpSess := this.SessionEngine.GetAllUpSession(this.AreaName[:])
	allS = append(allS, allDownSess...)
	allS = append(allS, allUpSess...)
	for i := 0; i < len(allS); i++ {
		if bytes.Equal(nodeStore.AddressNet(allS[i].GetName()), newNode.IdInfo.Id) {
			return
		}
	}

	//utils.Log.Info().Msgf(" 222self:%s  comein %s", this.GetNetId().B58String(), newNode.IdInfo.Id.B58String())
	if new(big.Int).SetBytes(this.GetNetId()).Cmp(new(big.Int).SetBytes([]byte(newNode.IdInfo.Id))) == -1 {
		//接入节点比自己大的情况
		sortedSes, _ := message_center.GetSortSession(allUpSess, this.GetNetId())
		if len(allUpSess) != 0 {
			// upSess已经满了 并且 新node大于upSess的最大地址，直接丢弃
			if len(allUpSess) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(newNode.IdInfo.Id)).Cmp(new(big.Int).SetBytes([]byte(sortedSes[0].GetName()))) == 1 {
				oneByoneNeed = false
			}
			//直接插入 后删除最大
			if len(allUpSess) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(newNode.IdInfo.Id)).Cmp(new(big.Int).SetBytes([]byte(sortedSes[0].GetName()))) == -1 {
				delSess = append(delSess, sortedSes[0])
			}
			if len(allUpSess) > config.GreaterThanSelfMaxConn {
				deltimes := len(allUpSess) - config.GreaterThanSelfMaxConn
				for i := 0; i < deltimes; i++ {
					delSess = append(delSess, sortedSes[i])
				}
			}
		}
	} else {
		//接入节点比自己小的情况
		sortedSes, _ := message_center.GetSortSession(allDownSess, this.GetNetId())
		if len(allDownSess) != 0 {
			// downSess已经满了 并且 新node小于downSess的最小地址，直接丢弃
			if len(allDownSess) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(newNode.IdInfo.Id)).Cmp(new(big.Int).SetBytes([]byte(sortedSes[config.GreaterThanSelfMaxConn-1].GetName()))) == -1 {
				oneByoneNeed = false
			}
			//直接插入 后删除最小
			if len(allDownSess) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(newNode.IdInfo.Id)).Cmp(new(big.Int).SetBytes([]byte(sortedSes[len(sortedSes)-1].GetName()))) == 1 {
				delSess = append(delSess, sortedSes[len(sortedSes)-1])
			}
			if len(allDownSess) > config.GreaterThanSelfMaxConn {
				deltimes := len(allDownSess) - config.GreaterThanSelfMaxConn
				for i := 0; i < deltimes; i++ {
					delSess = append(delSess, sortedSes[len(sortedSes)-1-i])
				}
			}
		}
	}

	if len(delSess) != 0 {
		for i := 0; i < len(delSess); i++ {
			//utils.Log.Info().Msgf("删除特定选中节点 %s -> %s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet([]byte(delSess[i].GetName())).B58String())
			if this.NodeManager.IsLogicNode(nodeStore.AddressNet(delSess[i].GetName())) ||
				this.Vm.IsSelfVnodeNeed(nodeStore.AddressNet(delSess[i].GetName())) {
				if len(allDownSess) >= len(allUpSess) {
					this.SessionEngine.DelNodeDownSession(this.AreaName[:], delSess[i])
				} else {
					this.SessionEngine.DelNodeUpSession(this.AreaName[:], delSess[i])
				}
			} else {
				this.SessionEngine.RemoveSession(this.AreaName[:], delSess[i])
				delSess[i].Close()
			}
		}
	}

	ok, _ := this.NodeManager.CheckNeedNode(&newNode.IdInfo.Id)
	if ok {
		kadNeed = true
	}
	if Session, ok := this.SessionEngine.GetSessionAll(this.AreaName[:], utils.Bytes2string(newNode.IdInfo.Id)); ok {
		if len(Session) != 0 {
			if oneByoneNeed {
				if this.SessionEngine.CompareAddressNet(engine.AddressNet(newNode.IdInfo.Id)) {
					this.SessionEngine.AddDownSession(this.AreaName[:], Session[0])
				} else {
					this.SessionEngine.AddUpSession(this.AreaName[:], Session[0])
				}
				oneByoneNeed = false
			}
			kadNeed = false
		}
	}

	var err error
	if kadNeed && oneByoneNeed {
		//同时符合kad协议和相邻地址一个连一个协议
		// aa := ""
		// for v := 0; v < len(allUpSess); v++ {
		// 	nn := fmt.Sprintf("\n upSession : %s  %d\n", nodeStore.AddressNet([]byte(allUpSess[v].GetName())).B58String(), allUpSess[v].GetIndex())
		// 	aa += nn
		// }
		// for v := 0; v < len(allDownSess); v++ {
		// 	nn := fmt.Sprintf("\n downSession : %s  %d\n", nodeStore.AddressNet([]byte(allDownSess[v].GetName())).B58String(), allDownSess[v].GetIndex())
		// 	aa += nn
		// }
		// utils.Log.Info().Msgf("22222排序位次 : ziji :%s in :%s  \n----> %s, ", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), newNode.IdInfo.Id.B58String(), aa)
		_, err = this.connNetByMethod(newNode.Addr, uint32(newNode.QuicPort), uint32(newNode.TcpPort), false, engine.BothMod)
	} else if !kadNeed && oneByoneNeed {
		//不符合kad协议 只符合相邻地址一个连一个协议
		// aa := ""
		// for v := 0; v < len(allUpSess); v++ {
		// 	nn := fmt.Sprintf("\n upSession : %s  %d\n", nodeStore.AddressNet([]byte(allUpSess[v].GetName())).B58String(), allUpSess[v].GetIndex())
		// 	aa += nn
		// }
		// for v := 0; v < len(allDownSess); v++ {
		// 	nn := fmt.Sprintf("\n downSession : %s  %d\n", nodeStore.AddressNet([]byte(allDownSess[v].GetName())).B58String(), allDownSess[v].GetIndex())
		// 	aa += nn
		// }
		// utils.Log.Info().Msgf("22222排序位次 : ziji :%s in :%s  \n----> %s, ", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), newNode.IdInfo.Id.B58String(), aa)
		_, err = this.connNetByMethod(newNode.Addr, uint32(newNode.QuicPort), uint32(newNode.TcpPort), false, engine.OnebyoneMod)
	} else if kadNeed && !oneByoneNeed {
		//只符合kad协议 不符合相邻地址一个连一个协议
		// utils.Log.Info().Msgf("3333 from:%s target:%s", this.GetNetId().B58String(), newNode.IdInfo.Id.B58String())
		// aa := ""
		// for v := 0; v < len(allUpSess); v++ {
		// 	nn := fmt.Sprintf("\n upSession : %s  %d\n", nodeStore.AddressNet([]byte(allUpSess[v].GetName())).B58String(), allUpSess[v].GetIndex())
		// 	aa += nn
		// }
		// for v := 0; v < len(allDownSess); v++ {
		// 	nn := fmt.Sprintf("\n downSession : %s  %d\n", nodeStore.AddressNet([]byte(allDownSess[v].GetName())).B58String(), allDownSess[v].GetIndex())
		// 	aa += nn
		// }
		// utils.Log.Info().Msgf("22222排序位次 :\nziji :%s  \n----> %s, ", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), aa)
		_, err = this.connNetByMethod(newNode.Addr, uint32(newNode.QuicPort), uint32(newNode.TcpPort), false, engine.KadMod)
	}

	if err != nil {
		//utils.Log.Info().Msgf("oneByone connecting_Delete error ", err.Error())
		return
	}

	this.delMoreSess(this.SessionEngine.GetAllDownSession(this.AreaName[:]), false)
	this.delMoreSess(this.SessionEngine.GetAllUpSession(this.AreaName[:]), true)

	if kadNeed {
		//非超级节点判断超级节点是否改变
		if !this.NodeManager.NodeSelf.GetIsSuper() {
			nearId := this.NodeManager.FindNearInSuper(&this.NodeManager.NodeSelf.IdInfo.Id, nil, false, nil)
			//判断是否需要替换超级节点
			if nearId == nil || bytes.Equal(*nearId, *this.NodeManager.GetSuperPeerId()) {
				return
			}
			// utils.Log.Info().Msgf("Replace supernode id:%s", nearId.B58String())
			this.NodeManager.SetSuperPeerId(nearId)
		}
	}

}

/*
 * 获取邻居节点的虚拟节点和邻居节点连接的虚拟节点信息
 */
func (this *Area) recvNearVnodes(bv *[]byte, node *nodeStore.Node) {
	var vnodeinfos []virtual_node.VnodeinfoS
	vrp := new(go_protobuf.VnodeinfoSRepeated)

	if bv == nil {
		return
	}
	err := proto.Unmarshal(*bv, vrp)
	if err != nil {
		utils.Log.Warn().Msgf("recvNearVnodes Unmarshal fail, err:", err)
		return
	}

	for _, v := range vrp.Vnodes {
		if v.Index == 0 {
			continue
		}
		var addr string
		var port uint64
		var quicPort uint64
		if v.Addr == "" && v.TcpPort == 0 {
			addr = node.Addr
			port = uint64(node.TcpPort)
			quicPort = uint64(node.QuicPort)
		} else {
			addr = v.Addr
			port = v.TcpPort
			quicPort = v.QuicPort
		}
		//utils.Log.Info().Msgf("recvNearVnodes n:%s v:%s Addr:%s, port:%d", nodeStore.AddressNet(v.Nid).B58String(), nodeStore.AddressNet(v.Vid).B58String(), addr, port)
		vnodeinfos = append(vnodeinfos, virtual_node.VnodeinfoS{
			Nid:      v.Nid,
			Vid:      v.Vid,
			Index:    v.Index,
			Addr:     addr,
			TcpPort:  port,
			QuicPort: quicPort,
		})
	}

	this.Vm.RLock()
	for _, m := range this.Vm.VnodeMap {
		if m.Vnode.Index == 0 {
			continue
		}
		vnodeS := virtual_node.VnodeinfoS{
			Nid:   m.Vnode.Nid,
			Index: m.Vnode.Index,
			Vid:   m.Vnode.Vid,
		}
		vnodeinfos = append(vnodeinfos, vnodeS)
	}

	group := new(sync.WaitGroup)
	delm := new(sync.Map)

	group.Add(len(this.Vm.VnodeMap))
	var tmpArr []uint64
	for i, _ := range this.Vm.VnodeMap {
		tmpArr = append(tmpArr, i)
	}
	this.Vm.RUnlock()
	for _, i := range tmpArr {
		if v, ok := this.Vm.VnodeMap[i]; ok {
			if v.HeadTail {
				//utils.Log.Info().Msgf("self : %s mVode.HeadTail : %t", this.Vm.VnodeMap[i].Vnode.Vid.B58String(), this.Vm.VnodeMap[i].HeadTail)
				go this.recvVnodeHeadTailHandler(vnodeinfos, v, group, delm)
			} else {
				go this.recvVnodeHandler(vnodeinfos, v, group, delm)
			}
		}
	}
	group.Wait()
	delm.Range(func(key, value interface{}) bool {
		delk, ok := key.(string)
		if !ok {
			return false
		}
		delV, ok := value.(virtual_node.VnodeinfoS)
		if !ok {
			return false
		}
		if delk != "" && !this.Vm.IsSelfVnodeNeed(delV.Nid) &&
			!this.NodeManager.IsLogicNode(delV.Nid) &&
			!this.SessionEngine.IsNodeUpdownSess(this.AreaName[:], engine.AddressNet(delV.Nid)) {
			if ss, ok := this.SessionEngine.GetSession(this.AreaName[:], delk); ok {
				//utils.Log.Info().Msgf("shanchu_shanchu self:%s target:%s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), delV.Nid.B58String())
				ss.Close()
			}
		}
		return true
	})
}

/*
接收邻居节点的逻辑节点和虚拟节点并添加
*/
func (this *Area) recvNearLogicNodes(bs *[]byte, recvid *nodeStore.AddressNet) {
	// this.GetNearSuperAddr_recvLock.Lock()
	// defer this.GetNearSuperAddr_recvLock.Unlock()
	nodes, err := nodeStore.ParseNodesProto(bs)
	if err != nil {
		return
	}

	for i, _ := range nodes {
		newNode := nodes[i]
		//utils.Log.Info().Msgf(" 收到self:%s  comein %s", this.GetNetId().B58String(), newNode.IdInfo.Id.B58String())
		connectingKeyQuic := newNode.Addr + ":" + strconv.Itoa(int(newNode.QuicPort))
		_, ok := this.connecting.LoadOrStore(connectingKeyQuic, 0)
		if ok {
			continue
		}

		connectingKeyTcp := newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort))
		if newNode.TcpPort != newNode.QuicPort {
			_, ok = this.connecting.LoadOrStore(connectingKeyTcp, 0)
			if ok {
				continue
			}
		}

		if this.NodeManager.IsHeadTailModl {
			//utils.Log.Info().Msgf("使用头尾模式进行判断 : %s", this.GetNetId().B58String())
			this.onebyoneRecvNodeHeadTail(newNode)
		} else {
			//utils.Log.Info().Msgf("使用普通模式进行判断 : %s", this.GetNetId().B58String())
			this.oneByoneRecvNode(newNode)
		}
		this.connecting.Delete(connectingKeyQuic)
		if newNode.TcpPort != newNode.QuicPort {
			this.connecting.Delete(connectingKeyTcp)
		}
	}

	bv, err := this.MessageCenter.SendNeighborMsgWaitRequest(gconfig.MSGID_getNearConnectVnode, recvid, nil, time.Second*8)
	if err != nil {
		//utils.Log.Warn().Msgf("MSGID_getNearConnectVnode error:%s self:%s target:%s ", err.Error(), this.GetNetId().B58String(), recvid.B58String())
		return
	}
	node := this.NodeManager.FindNode(recvid)
	if node == nil {
		return
	}
	this.recvNearVnodes(bv, node)
}

/*
询问关闭这个链接
当双方都没有这个链接的引用时，就关闭这个链接
*/
func (this *Area) AskCloseConn_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	mh := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// utils.Log.Info().Msgf("AdkClose1:%s", mh.B58String())
	// utils.Log.Info().Msgf("AdkClose2:%v", this.NodeManager.FindNodeInLogic(&mh))
	// utils.Log.Info().Msgf("AdkClose3:%t", this.NodeManager.FindWhiteList(&mh))
	if this.NodeManager.FindNodeInLogic(&mh) == nil && !this.NodeManager.FindWhiteList(&mh) {
		//自己也没有这个连接的引用，则关闭这个链接
		// utils.Log.Info().Msgf("Close this session:%s", mh.B58String())
		msg.Session.Close()
	}
}

/*
 * 接收真实节点下线的广播
 */
func (this *Area) MulticastOffline_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("qlw---MulticastOffline_recv: 接收到消息")
	if message.Body.Content == nil || len(*message.Body.Content) == 0 {
		return
	}

	// 解析下线节点信息
	offlineNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	if err != nil {
		return
	}

	// 1. 获取需要查询的节点真实地址信息
	nid := offlineNode.IdInfo.Id
	// utils.Log.Error().Msgf("下线节点地址为: %s machineId:%s", nid.B58String(), offlineNode.MachineIDStr)

	// 2. 判断自己是不是该节点，如果是，则直接返回
	if bytes.Equal(nid, this.Vm.DiscoverVnodes.Vnode.Nid) {
		return
	}

	// 3. 检查自己是否含有该节点，如果不存在则直接返回
	var bNeedDealNode bool
	if this.Vm.CheckNodeinfoExistInSelf(nid) {
		bNeedDealNode = true
	}

	// 4. 判断是否需要处理下线的客户端
	var bNeedDealProxy bool
	if this.ProxyDetail.CheckIsSaveAddr(&nid) {
		bNeedDealProxy = true
	}

	// 5. 如果即不是自己的逻辑节点, 也不是代理的存储节点, 则不用处理
	if !bNeedDealNode && !bNeedDealProxy {
		return
	}

	// 6. 存在的话，先尝试询问对象是否在线
	_, _, _, err = this.SendP2pMsgWaitRequest(gconfig.MSGID_checkAddrOnline, &nid, nil, 3*time.Second)
	// 6.1 在线，则直接返回，无需操作
	if err == nil {
		return
	}

	// 不在线，进行相应的处理
	// 7. 处理逻辑节点相关操作
	if bNeedDealNode {
		// 7.1 删除发现节点的对应信息
		this.Vm.DiscoverVnodes.DeleteNid(nid)
		// 7.2 删除虚拟节点的对应信息
		this.Vm.RLock()
		for _, one := range this.Vm.VnodeMap {
			one.DeleteNid(nid)
		}
		this.Vm.RUnlock()
		// 7.3. 删除连接对应的加密管道
		if this.MessageCenter != nil && this.MessageCenter.RatchetSession != nil {
			this.MessageCenter.RatchetSession.RemoveSendPipe(nid, offlineNode.MachineIDStr)
		}

		// 7.4 触发查找新的节点操作
		this.findNearNodeTimer.Release()
	}

	// 8. 处理代理的存储节点操作
	if bNeedDealProxy {
		// 删除对应的存储节点信息
		this.ProxyDetail.NodeOfflineDeal(&nid)
	}

	// utils.Log.Info().Msgf("处理节点下线成功: nodeid:%s, self:%s", nid.B58String(), this.Vm.DiscoverVnodes.Vnode.Vid.B58String())
}

/*
 * 接受处理其他节点消息，删除本节点up down排序中虚拟节点
 */
func (this *Area) DelVnodeInSelf(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if message.Body.Content == nil && len(*message.Body.Content) == 0 {
		return
	}
	vid := nodeStore.AddressNet(*message.Body.Content)
	this.Vm.DelUpdownVnodeByAddr(virtual_node.AddressNetExtend(vid))
}

/*
 * 接收虚拟节点下线的广播
 */
func (this *Area) MulticastVnodeOffline_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// 1. 获取需要查询的虚拟节点信息
	vnodeInfo, err := virtual_node.ParseVnodeinfo(*message.Body.Content)
	if err != nil {
		// 解析失败，直接返回，不做处理
		// utils.Log.Info().Msgf("parse vnode info err:%s", err)
		return
	}
	// 1.1 检查虚拟节点参数的有效性
	if vnodeInfo == nil || vnodeInfo.Nid == nil || len(vnodeInfo.Nid) == 0 || vnodeInfo.Vid == nil || len(vnodeInfo.Vid) == 0 {
		// utils.Log.Info().Msgf("recv vnode info incorrect")
		return
	}

	// 2. 检查自己是否含有该节点，如果不存在则直接返回
	if this.Vm.FindVnodeinfo(vnodeInfo.Vid) == nil {
		return
	}

	// 3. 存在的话，先尝试询问对象是否在线
	if this.Vc.CheckVnodeIsOnline(vnodeInfo) {
		// 3.1 在线，则直接返回，无需操作
		return
	}
	// 3.2 不在线，删除对应的虚拟节点信息
	// 3.2.1 删除发现节点中的虚拟节点信息
	this.Vm.DiscoverVnodes.DeleteVid(vnodeInfo)
	// 3.2.2 删除虚拟节点中的虚拟节点信息
	this.Vm.RLock()
	for _, one := range this.Vm.VnodeMap {
		one.DeleteVid(vnodeInfo)
	}
	this.Vm.RUnlock()

	// 4. 触发查找新的节点操作
	this.findNearNodeTimer.Release()
}

/*
 * 查询一个地址是否在线
 */
func (this *Area) checkAddrOnline(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	this.MessageCenter.SendSearchSuperReplyMsg(message, gconfig.MSGID_checkAddrOnline_recv, nil)
}

/*
 * 查询一个地址是否在线的返回
 */
func (this *Area) checkAddrOnline_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseBytes(utils.Bytes2string(*&message.Body.Hash), nil)
}

/*
 * 查询一个虚拟地址是否在线
 */
func (this *Area) checkVnodeAddrOnline(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// 1. 获取需要查询的虚拟节点地址信息
	vid := virtual_node.AddressNetExtend(*message.Body.Content)

	// 2. 检查自己是否含有该节点
	vnodeCheckResult := make([]byte, 1)
	if this.Vm.FindVnodeInSelf(vid) == nil {
		// 2.1 不存在该虚拟节点，则告知已下线
		vnodeCheckResult[0] = byte(config.VNodeIdResult_offline)
	} else {
		// 2.2 存在虚拟节点，则告知在线
		vnodeCheckResult[0] = byte(config.VNodeIdResult_online)
	}

	// 3. 返回对应的消息
	this.SendP2pReplyMsg(message, gconfig.MSGID_checkVnodeAddrOnline_recv, &vnodeCheckResult)
}

/*
 * 查询一个虚拟地址是否在线的返回
 */
func (this *Area) checkVnodeAddrOnline_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// 直接返回是否在线标识信息
	flood.ResponseBytes(utils.Bytes2string(*&message.Body.Hash), message.Body.Content)
}

/*
 * 接收真实节点所有虚拟节点信息的广播
 *
 * @author: qlw
 */
func (this *Area) MulticastSendVnode_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("-----------------接收真实节点所有虚拟节点信息的广播 接收处")
	// 1. 解析虚拟节点
	if message.Body.Content == nil {
		return
	}
	vrp := new(go_protobuf.VnodeinfoRepeated)
	err := proto.Unmarshal(*message.Body.Content, vrp)
	if err != nil {
		utils.Log.Info().Msgf("不能解析proto错误:%s", err.Error())
		return
	}

	// 2. 提取所有的虚拟节点地址信息
	vnodeinfos := make([]virtual_node.Vnodeinfo, 0)
	for _, one := range vrp.Vnodes {
		// 只处理虚拟节点信息
		if one.Index == 0 {
			continue
		}

		vnodeOne := virtual_node.Vnodeinfo{
			Nid:   one.Nid,
			Index: one.Index,
			Vid:   one.Vid,
		}
		vnodeinfos = append(vnodeinfos, vnodeOne)
		// utils.Log.Info().Msgf("解析到虚拟节点地址为:%s", vnodeOne.Vid.B58String())
	}
	if len(vnodeinfos) == 0 {
		return
	}

	// 3. 判断自己是不是该节点，如果是，则直接返回
	nid := vnodeinfos[0].Nid
	if bytes.Equal(nid, this.Vm.DiscoverVnodes.Vnode.Nid) {
		return
	}

	// 4. 询问对象是否在线
	_, _, _, err = this.SendP2pMsgWaitRequest(gconfig.MSGID_checkAddrOnline, &nid, nil, 3*time.Second)
	if err != nil {
		return
	}

	// 5. 尝试把虚拟节点添加到逻辑节点中
	this.Vm.AddLogicVnodeinfo(vnodeinfos...)
}

/*
 * 通过获取到邻居节点保存的虚拟节点信息，onebyone规则头尾模式排序
 */
func (this *Area) recvVnodeHeadTailHandler(vnodeinfos []virtual_node.VnodeinfoS, mVode *virtual_node.Vnode, group *sync.WaitGroup, dm *sync.Map) {
	mVode.Lock.Lock()
	defer mVode.Lock.Unlock()
	for _, one := range vnodeinfos {
		//utils.Log.Info().Msgf("OneHandler self:%s in:%s", one.Vid.B58String(), mVode.Vnode.Vid.B58String())
		this.onebyoneRecvVnodeHeadTail(one, mVode, dm)
	}
	group.Done()
}

/*
 * onebyone规则首尾模式下排序虚拟节点
 */
func (this *Area) onebyoneRecvVnodeHeadTail(vi virtual_node.VnodeinfoS, mVnode *virtual_node.Vnode, dm *sync.Map) {
	//排除本虚拟节点
	if bytes.Equal(vi.Vid, mVnode.Vnode.Vid) {
		return
	}

	var oneByoneNeed bool
	var delVnode []virtual_node.VnodeinfoS
	var allVnode []virtual_node.VnodeinfoS
	oneByoneNeed = true

	//utils.Log.Info().Msgf(" 111self:%s  comein %s", mVnode.Vnode.Vid.B58String(), vi.Vid.B58String())
	dVnode := mVnode.GetDownVnodeInfo()
	uVnode := mVnode.GetUpVnodeInfo()
	allVnode = append(allVnode, dVnode...)
	allVnode = append(allVnode, uVnode...)

	if len(dVnode) >= config.GreaterThanSelfMaxConn/2 && len(uVnode) >= config.GreaterThanSelfMaxConn/2 {
		//utils.Log.Info().Msgf("模式转换回去 self : %s", mVnode.Vnode.Vid.B58String())
		mVnode.HeadTail = false
		return
	}

	for _, one := range allVnode {
		if bytes.Equal(one.Vid, vi.Vid) {
			return
		}
	}

	if len(dVnode) >= len(uVnode) {
		//本节点比发现的节点大
		//在down比up多的时候，如果是大于本节点的，直接去连，反之本节点比newnode大就需处理
		if new(big.Int).SetBytes(mVnode.Vnode.Vid).Cmp(new(big.Int).SetBytes([]byte(vi.Vid))) == 1 {
			//取偶数方便后面计算
			downCap := (config.GreaterThanSelfMaxConn + config.GreaterThanSelfMaxConn - len(uVnode)) / 2 * 2
			//只有allDownSess数量大于等于downCap数量时候才做插入并删除一个的操作，没到就直接添加
			if downCap <= len(dVnode) {
				// sortedSes, _ := message_center.GetSortSession(allDownSess, this.GetNetId())
				sortedSes, _ := mVnode.GetSortAddressNetExtend(dVnode, mVnode.Vnode.Vid)
				//utils.Log.Info().Msgf("==== %d %v", downCap, sortedSes)
				if new(big.Int).SetBytes(vi.Vid).Cmp(new(big.Int).SetBytes([]byte(sortedSes[len(dVnode)-1].Vid))) == -1 {
					delVnode = append(delVnode, sortedSes[len(dVnode)-1])
				} else if new(big.Int).SetBytes(vi.Vid).Cmp(new(big.Int).SetBytes([]byte(sortedSes[downCap/2].Vid))) == -1 {
					delVnode = append(delVnode, sortedSes[downCap/2])
				}
				if new(big.Int).SetBytes(vi.Vid).Cmp(new(big.Int).SetBytes([]byte(sortedSes[downCap/2-1].Vid))) == 1 {
					delVnode = append(delVnode, sortedSes[downCap/2-1])
				}

				if len(delVnode) == 0 {
					oneByoneNeed = false
				}
				if len(dVnode) > downCap && len(delVnode) == 0 {

					del := len(dVnode) - downCap
					for i := 0; i < del; i++ {
						delVnode = append(delVnode, sortedSes[downCap/2+i])
					}
				}
			}
		}
	} else {
		//本节点比发现的节点小
		//在up比down多的时候，如果是小于本节点的，直接去连，反之本节点比newnode小就需处理
		if new(big.Int).SetBytes(mVnode.Vnode.Vid).Cmp(new(big.Int).SetBytes([]byte(vi.Vid))) == -1 {
			//取偶数方便后面计算
			upCap := (config.GreaterThanSelfMaxConn + config.GreaterThanSelfMaxConn - len(dVnode)) / 2 * 2
			//只有allUpSess数量大于等于upCap数量时候才做插入并删除一个的操作，没到就直接添加
			if upCap <= len(uVnode) {
				sortedSes, _ := mVnode.GetSortAddressNetExtend(uVnode, mVnode.Vnode.Vid)
				//utils.Log.Info().Msgf("==== %d %v", upCap, sortedSes)
				if new(big.Int).SetBytes(vi.Vid).Cmp(new(big.Int).SetBytes([]byte(sortedSes[0].Vid))) == 1 {
					delVnode = append(delVnode, sortedSes[0])
				} else if new(big.Int).SetBytes(vi.Vid).Cmp(new(big.Int).SetBytes([]byte(sortedSes[upCap/2-1].Vid))) == 1 {
					delVnode = append(delVnode, sortedSes[upCap/2-1])
				}
				if new(big.Int).SetBytes(vi.Vid).Cmp(new(big.Int).SetBytes([]byte(sortedSes[upCap/2].Vid))) == -1 {
					delVnode = append(delVnode, sortedSes[upCap/2])
				}
				if len(delVnode) == 0 {
					oneByoneNeed = false
				}
				if len(uVnode) > upCap && len(delVnode) == 0 {
					del := len(uVnode) - upCap
					for i := 0; i < del; i++ {
						delVnode = append(delVnode, sortedSes[upCap/2+i])
					}
				}
			}
		}
	}

	if len(delVnode) != 0 {
		//utils.Log.Info().Msgf("删除特定选中节点 %s -> %s", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), nodeStore.AddressNet([]byte(delSess[0].Vid)).B58String())
		for i := 0; i < len(delVnode); i++ {
			mVnode.DownVnodeInfo.Delete(utils.Bytes2string(delVnode[i].Vid))
			mVnode.UpVnodeInfo.Delete(utils.Bytes2string(delVnode[i].Vid))
			dm.Store(utils.Bytes2string(delVnode[i].Nid), delVnode[i])
		}

	}

	if oneByoneNeed {
		if _, ok := this.SessionEngine.GetSession(this.AreaName[:], utils.Bytes2string(vi.Nid)); !ok {
			//节点连接
			if vi.Addr != "" {
				var bConnSuccess bool
				if this.CheckSupportQuicConn() {
					connectingKey := vi.Addr + ":" + strconv.Itoa(int(vi.QuicPort))
					_, ok := this.connecting.LoadOrStore(connectingKey, 0)
					if ok {
						return
					}
					//utils.Log.Info().Msgf("连接连接 self %s to %s target %s", mVnode.Vnode.Vid.B58String(), connectingKey, vi.Nid.B58String())
					_, err := this.SessionEngine.AddClientQuicConn(this.AreaName[:], vi.Addr, uint32(vi.QuicPort), false, engine.VnodeMod)
					this.connecting.Delete(connectingKey)
					if err == nil {
						bConnSuccess = true
					}
				}
				if !bConnSuccess && this.CheckSupportTcpConn() {
					connectingKey := vi.Addr + ":" + strconv.Itoa(int(vi.TcpPort))
					_, ok := this.connecting.LoadOrStore(connectingKey, 0)
					if ok {
						return
					}
					_, err := this.SessionEngine.AddClientConn(this.AreaName[:], vi.Addr, uint32(vi.TcpPort), false, engine.VnodeMod)
					this.connecting.Delete(connectingKey)
					if err != nil {
						//utils.Log.Warn().Msgf("%s 添加地址:%s 失败 error : %s", mVnode.Vnode.Vid.B58String(), connectingKey, err.Error())
						return
					}
					bConnSuccess = true
				}
				if !bConnSuccess {
					return
				}
			}
		}
		if new(big.Int).SetBytes([]byte(vi.Vid)).Cmp(new(big.Int).SetBytes([]byte(mVnode.Vnode.Vid))) == 1 {
			//utils.Log.Info().Msgf("添加up节点 self : %s store : %s", mVode.Vnode.Vid.B58String(), vi.Vid.B58String())
			mVnode.UpVnodeInfo.Store(utils.Bytes2string(vi.Vid), vi)
		} else {
			//utils.Log.Info().Msgf("添加down节点 self : %s store : %s", mVode.Vnode.Vid.B58String(), vi.Vid.B58String())
			mVnode.DownVnodeInfo.Store(utils.Bytes2string(vi.Vid), vi)
		}
	}
}

/*
 * 通过获取到邻居节点保存的虚拟节点信息，onebyone规则正常模式排序
 */
func (this *Area) recvVnodeHandler(vnodeinfos []virtual_node.VnodeinfoS, mVode *virtual_node.Vnode, group *sync.WaitGroup, dm *sync.Map) {
	mVode.Lock.Lock()
	defer mVode.Lock.Unlock()

	for _, one := range vnodeinfos {
		this.oneByoneRecvVnode(one, mVode, dm)
	}
	group.Done()
}

/*
 * onebyone规则正常模式下排序虚拟节点
 */
func (this *Area) oneByoneRecvVnode(vi virtual_node.VnodeinfoS, mVode *virtual_node.Vnode, dm *sync.Map) {
	var oneByoneNeed bool
	var delVnode []virtual_node.VnodeinfoS
	var allVnode []virtual_node.VnodeinfoS
	oneByoneNeed = true

	//排除当前虚拟节点
	if bytes.Equal(vi.Vid, mVode.Vnode.Vid) {
		return
	}

	dVnode := mVode.GetDownVnodeInfo()
	uVnode := mVode.GetUpVnodeInfo()
	allVnode = append(allVnode, dVnode...)
	allVnode = append(allVnode, uVnode...)

	for _, one := range allVnode {
		if bytes.Equal(one.Vid, vi.Vid) {
			return
		}
	}

	if new(big.Int).SetBytes(mVode.Vnode.Vid).Cmp(new(big.Int).SetBytes([]byte(vi.Vid))) == -1 {
		//接入节点比自己大的情况
		sortedVnode, _ := mVode.GetSortAddressNetExtend(uVnode, mVode.Vnode.Vid)
		if len(uVnode) != 0 {
			// upSess已经满了 并且 新node大于upSess的最大地址，直接丢弃
			if len(uVnode) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(vi.Vid)).Cmp(new(big.Int).SetBytes([]byte(sortedVnode[0].Vid))) == 1 {
				oneByoneNeed = false
			}
			//直接插入 后删除最大
			if len(uVnode) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(vi.Vid)).Cmp(new(big.Int).SetBytes([]byte(sortedVnode[0].Vid))) == -1 {
				delVnode = append(delVnode, sortedVnode[0])
			}
			//utils.Log.Info().Msgf(" 444self:%s  comein %s, oneByoneNeed %t needDelMax %t MaxforceDel %t", this.GetNetId().B58String(), newNode.IdInfo.Id.B58String(), oneByoneNeed, needDelMax, MaxforceDel)
			if len(uVnode) > config.GreaterThanSelfMaxConn {
				deltimes := len(uVnode) - config.GreaterThanSelfMaxConn
				for i := 0; i < deltimes; i++ {
					delVnode = append(delVnode, sortedVnode[i])
				}
			}
		}
	} else {
		//接入节点比自己小的情况
		sortedVnode, _ := mVode.GetSortAddressNetExtend(dVnode, mVode.Vnode.Vid)
		if len(dVnode) != 0 {
			// downSess已经满了 并且 新node小于downSess的最小地址，直接丢弃
			if len(dVnode) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(vi.Vid)).Cmp(new(big.Int).SetBytes([]byte(sortedVnode[config.GreaterThanSelfMaxConn-1].Vid))) == -1 {
				oneByoneNeed = false
			}
			//直接插入 后删除最小
			if len(dVnode) >= config.GreaterThanSelfMaxConn &&
				new(big.Int).SetBytes([]byte(vi.Vid)).Cmp(new(big.Int).SetBytes([]byte(sortedVnode[len(sortedVnode)-1].Vid))) == 1 {
				delVnode = append(delVnode, dVnode[len(sortedVnode)-1])
			}
			if len(dVnode) > config.GreaterThanSelfMaxConn {
				deltimes := len(dVnode) - config.GreaterThanSelfMaxConn
				for i := 0; i < deltimes; i++ {
					delVnode = append(delVnode, sortedVnode[len(sortedVnode)-1-i])
				}
			}
		}
	}

	if len(delVnode) != 0 {
		for i := 0; i < len(delVnode); i++ {
			mVode.DownVnodeInfo.Delete(utils.Bytes2string(delVnode[i].Vid))
			mVode.UpVnodeInfo.Delete(utils.Bytes2string(delVnode[i].Vid))
			dm.Store(utils.Bytes2string(delVnode[i].Nid), delVnode[i])
		}
	}

	if oneByoneNeed {
		//先连接节点再添加进up down，连接失败不添加
		if _, ok := this.SessionEngine.GetSession(this.AreaName[:], utils.Bytes2string(vi.Nid)); !ok {
			if vi.Addr != "" {
				connectingKey := vi.Addr + ":" + strconv.Itoa(int(vi.QuicPort))
				_, ok := this.connecting.LoadOrStore(connectingKey, 0)
				if ok {
					// utils.Log.Info().Msgf("%s 解析节点地址 %s %s %d 44444444", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), newNode.IdInfo.Id.B58String(), newNode.Addr, newNode.TcpPort)
					return
				}
				_, err := this.SessionEngine.AddClientQuicConn(this.AreaName[:], vi.Addr, uint32(vi.QuicPort), false, engine.VnodeMod)
				this.connecting.Delete(connectingKey)
				if err != nil {
					utils.Log.Warn().Msgf("%s 添加地址:%s 失败 error : %s", mVode.Vnode.Vid.B58String(), connectingKey, err.Error())

					connectingKey = vi.Addr + ":" + strconv.Itoa(int(vi.TcpPort))
					_, ok := this.connecting.LoadOrStore(connectingKey, 0)
					if ok {
						// utils.Log.Info().Msgf("%s 解析节点地址 %s %s %d 44444444", this.NodeManager.NodeSelf.IdInfo.Id.B58String(), newNode.IdInfo.Id.B58String(), newNode.Addr, newNode.TcpPort)
						return
					}
					_, err := this.SessionEngine.AddClientConn(this.AreaName[:], vi.Addr, uint32(vi.TcpPort), false, engine.VnodeMod)
					this.connecting.Delete(connectingKey)
					if err != nil {
						utils.Log.Warn().Msgf("%s 添加地址:%s 失败 error : %s", mVode.Vnode.Vid.B58String(), connectingKey, err.Error())
						return
					}
				}
			}
		}
		if new(big.Int).SetBytes([]byte(vi.Vid)).Cmp(new(big.Int).SetBytes([]byte(mVode.Vnode.Vid))) == 1 {
			//utils.Log.Info().Msgf("添加up节点 self : %s store : %s", mVode.Vnode.Vid.B58String(), vi.Vid.B58String())
			mVode.UpVnodeInfo.Store(utils.Bytes2string(vi.Vid), vi)
		} else {
			//utils.Log.Info().Msgf("添加down节点 self : %s store : %s", mVode.Vnode.Vid.B58String(), vi.Vid.B58String())
			mVode.DownVnodeInfo.Store(utils.Bytes2string(vi.Vid), vi)
		}
	}
	this.delMoreVnode(mVode, dm)

	if mVode.CheckAutonomyFinish() {
		if (len(mVode.GetDownVnodeInfo()) < config.GreaterThanSelfMaxConn/2 && len(mVode.GetUpVnodeInfo()) >= config.GreaterThanSelfMaxConn/2) ||
			(len(mVode.GetUpVnodeInfo()) < config.GreaterThanSelfMaxConn/2 && len(mVode.GetDownVnodeInfo()) >= config.GreaterThanSelfMaxConn/2) {
			mVode.HeadTail = true
		}
	}
}

/*
 * 维护虚拟节点上下连接数
 */
func (this *Area) delMoreVnode(mVode *virtual_node.Vnode, dm *sync.Map) {
	dV := mVode.GetDownVnodeInfo()
	uV := mVode.GetUpVnodeInfo()
	dsort, _ := mVode.GetSortAddressNetExtend(dV, mVode.Vnode.Vid)
	usort, _ := mVode.GetSortAddressNetExtend(uV, mVode.Vnode.Vid)

	if del := len(uV) - config.GreaterThanSelfMaxConn; del > 0 {
		for i := 0; i < del; i++ {
			mVode.UpVnodeInfo.Delete(utils.Bytes2string(usort[i].Vid))
			dm.Store(utils.Bytes2string(usort[i].Nid), usort[i])
		}
	}

	if del := len(dV) - config.GreaterThanSelfMaxConn; del > 0 {
		for i := len(dV) - 1; i > len(dV)-del-1; i-- {
			mVode.DownVnodeInfo.Delete(utils.Bytes2string(dsort[i].Vid))
			dm.Store(utils.Bytes2string(dsort[i].Nid), dsort[i])
		}
	}
}

/*
 * 接受中转错误返回
 */
func (this *Area) recvRouterErr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//utils.Log.Info().Msgf("recvRouterErr recvRouterErr recvRouterErr")
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
 * 接收同步代理信息
 */
func (this *Area) syncProxyRec(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("syncProxyRec start self:%s from:%s", this.GetNetId().B58String(), message.Head.Sender.B58String())
	// 1. 判断参数的合法性
	if message.Body == nil || message.Body.Content == nil || message.Head.Sender == nil {
		utils.Log.Error().Msgf("参数的不合法!!!!!!")
		return
	}

	// 2. 解析代理信息
	proxyInfoes, err := proxyrec.ParseProxyesProto(message.Body.Content)
	if err != nil {
		utils.Log.Error().Msgf("代理信息解析失败!!!!!!")
		return
	}

	// 3. 处理代理信息
	this.ProxyData.AddOrUpdateProxyRec(proxyInfoes, message.Head.Sender, message.Head.RecvId)

	// 4. 返回对应的消息
	this.SendP2pReplyMsg(message, gconfig.MSGID_sync_proxy_recv, nil)
}

/*
 * 接收同步代理信息的返回
 */
func (this *Area) syncProxyRec_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
 * 获取地址对应的代理信息
 */
func (this *Area) getAddrProxy(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("getAddrProxy start self:%s from:%s", this.GetNetId().B58String(), message.Head.Sender.B58String())
	// 1. 判断参数的合法性
	if message.Body == nil || message.Body.Content == nil || message.Head.Sender == nil {
		utils.Log.Error().Msgf("参数的不合法!!!!!!")
		this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
		return
	}

	// 2. 解析代理信息
	proxyInfo, err := proxyrec.ParseProxyProto(message.Body.Content)
	if err != nil {
		utils.Log.Error().Msgf("代理信息解析失败!!!!!!")
		this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
		return
	}

	// 3. 处理代理信息
	exist, proxyInfoes := this.ProxyData.GetNodeIdProxy(proxyInfo)
	if !exist || len(proxyInfoes) == 0 {
		// utils.Log.Warn().Msgf("不存在代理信息!!!! clientId:%s machineId:%s", proxyInfo.NodeId.B58String(), proxyInfo.MachineId)
		this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
		return
	}

	// 4. 返回对应的消息
	var res go_protobuf.ProxyRepeated
	res.Proxys = make([]*go_protobuf.ProxyInfo, 0)
	for i := range proxyInfoes {
		var proxy go_protobuf.ProxyInfo
		proxy.Id = *proxyInfoes[i].NodeId
		proxy.ProxyId = *proxyInfoes[i].ProxyId
		proxy.MachineID = proxyInfoes[i].MachineId
		proxy.Version = proxyInfoes[i].Version

		res.Proxys = append(res.Proxys, &proxy)
	}
	bs, err := res.Marshal()
	if err != nil {
		utils.Log.Error().Msgf("ProxyInfo marshal err:%s", err)
		this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, nil)
		return
	}
	this.SendP2pReplyMsg(message, gconfig.MSGID_search_addr_proxy_recv, &bs)
}

/*
 * 获取地址对应的代理信息的返回
 */
func (this *Area) getAddrProxy_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
 * 指定连接方式进行连接
 */
func (this *Area) connNetByMethod(ip string, quicPort, tcpPort uint32, powerful bool, onlyAddOneSess string) (ss engine.Session, err error) {
	var bConnSuccess bool
	// 先尝试quic连接
	if this.CheckSupportQuicConn() {
		_, err = this.SessionEngine.AddClientQuicConn(this.AreaName[:], ip, quicPort, powerful, onlyAddOneSess)
		if err == nil {
			bConnSuccess = true
		}
	}

	// 再尝试tcp连接
	if !bConnSuccess && this.CheckSupportTcpConn() {
		_, err = this.SessionEngine.AddClientConn(this.AreaName[:], ip, tcpPort, powerful, onlyAddOneSess)
		if err == nil {
			bConnSuccess = true
		}
	}

	if !bConnSuccess && err == nil {
		err = errors.New("无法进行任何连接处理")
	}

	return
}
