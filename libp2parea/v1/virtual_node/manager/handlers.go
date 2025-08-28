package manager

import (
	"bytes"
	"encoding/binary"

	"github.com/gogo/protobuf/proto"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

/*
获取节点的虚拟节点开通状态
*/
func (this *VnodeCenter) GetVnodeOpenState(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("-----------------获取节点的虚拟节点开通状态")
	n := len(this.vm.GetVnodeNumber())
	bs := utils.Uint64ToBytes(uint64(n))
	this.messageCenter.SendP2pReplyMsgHE(message, config.MSGID_vnode_getstate_recv, &bs)

}

/*
获取节点的虚拟节点开通状态 返回
*/
func (this *VnodeCenter) GetVnodeOpenState_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(message_center.CLASS_vnode_getstate, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
获取相邻节点的Vnode地址
*/
func (this *VnodeCenter) GetNearSuperAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if this.isClose {
		// utils.Log.Info().Msgf("GetNearSuperAddr is close")
		this.messageCenter.SendP2pReplyMsgHE(message, config.MSGID_vnode_getNearSuperIP_recv, nil)
		return
	}

	// utils.Log.Info().Msgf("-----------------接收到邻居节点虚拟节点查询消息")
	// utils.Log.Info().Msgf("打印content:%d %s", len(*message.Body.Content), hex.EncodeToString(*message.Body.Content))
	findVnodeVO, err := virtual_node.ParseFindVnodeVO(*message.Body.Content)
	if err != nil {
		utils.Log.Info().Msgf("proto格式化错误:%s", err.Error())
		return
	}

	//验证消息发送方节点id
	if !findVnodeVO.Self.Check() {
		utils.Log.Info().Msgf("验证发送方节点不合法")
		return
	}

	//验证Target参数是否在自己的节点中
	have := false
	if findVnodeVO.Target.Nid == nil || len(findVnodeVO.Target.Nid) == 0 {
		// utils.Log.Info().Msgf("11111111111")
		//当target为空时，是自己的邻居节点
		if this.nodeManager.FindNode(&findVnodeVO.Self.Nid) != nil {
			// utils.Log.Info().Msgf("2222222")
			have = true
		}
		have = true
	} else {
		// utils.Log.Info().Msgf("3333333333333:%s", findVnodeVO.Target.Nid.B58String())
		// vnodeinfo := this.vm.GetVnodeSelf()
		// for _, one := range vnodeinfo {
		// 	// utils.Log.Info().Msgf("本节点id:%s", one.Vid.B58String())
		// 	if bytes.Equal(one.Vid, findVnodeVO.Target.Vid) {
		// 		have = true
		// 		break
		// 	}
		// }

		if this.vm.FindInVnodeSelf(findVnodeVO.Target.Vid) {
			have = true
		}

	}
	//检查发送目标节点是否在本节点中，不在本节点中，不处理
	if !have {
		utils.Log.Info().Msgf("不在本节点中self %s :%s", this.nodeManager.NodeSelf.IdInfo.Id.B58String(), findVnodeVO.Self.Nid.B58String())
		return
	}

	//将查询节点保存到client中
	this.vm.AddClientVnodeinfo(findVnodeVO.Self)

	//将对方节点保存到自己的逻辑节点中，此协议用于节点发现。
	this.vm.AddLogicVnodeinfo(findVnodeVO.Self)

	//获取本节点保存的所有id
	idsMap := make(map[string]nodeStore.AddressNet)
	vnodeinfos := this.vm.GetVnodeAll()
	//添加自己的虚拟节点中的逻辑节点
	for _, one := range vnodeinfos {
		temp := nodeStore.AddressNet(one.Vid)
		idsMap[utils.Bytes2string(temp)] = temp
	}

	//添加连接自己的client节点
	for _, one := range this.vm.GetClientVnodeinfo() {
		temp := nodeStore.AddressNet(one.Vid)
		idsMap[utils.Bytes2string(temp)] = temp
	}

	//添加Discover节点
	// this.vm.DiscoverVnodes.Vnode.Nid
	// idsMap

	//添加自己的虚拟节点
	// for _, one := range vnodeinfo {
	// 	temp := nodeStore.AddressNet(one.Vid)
	// 	idsMap[utils.Bytes2string(temp)] = temp
	// }
	// fmt.Println("添加自己的虚拟节点之后", idsMap)
	// //包含自己真实节点id，必须要包含真实节点，否则网络会形成孤岛。
	// for _, one := range nodeStore.GetAllNodes() {
	// 	idsMap[one.B58String()] = *one
	// }
	//不包括自己
	// delete(idsMap, findVnodeVO.Target.B58String())
	//不包括消息发送方id
	delete(idsMap, utils.Bytes2string(findVnodeVO.Self.Vid))

	//查找对方节点所需要的id
	selfVid := nodeStore.AddressNet(findVnodeVO.Self.Vid)
	idsm := nodeStore.NewIds(selfVid, nodeStore.NodeIdLevel)
	for _, one := range idsMap {
		idsm.AddId(one)
	}
	ids := idsm.GetIds()

	vinfos := make([]virtual_node.Vnodeinfo, 0)
	//这里需要返回查询参数，方便对方把查询和接收一一对应
	//这里规定index=0为消息发送方vid，index=1为发送发查询的逻辑节点vid。
	vinfos = append(vinfos, findVnodeVO.Self)
	vinfos = append(vinfos, findVnodeVO.Target)
	//再添加返回的逻辑节点
	for _, one := range ids {
		temp := virtual_node.AddressNetExtend(one)
		vinfo := this.vm.FindVnodeinfo(temp)
		if vinfo == nil {
			continue
		}
		vinfos = append(vinfos, *vinfo)
	}
	vrs := go_protobuf.VnodeinfoRepeated{
		Vnodes: make([]*go_protobuf.Vnodeinfo, 0),
	}
	for _, one := range vinfos {
		vnodeOne := &go_protobuf.Vnodeinfo{
			Nid:   one.Nid,
			Index: one.Index,
			Vid:   one.Vid,
		}
		vrs.Vnodes = append(vrs.Vnodes, vnodeOne)
		// utils.Log.Info().Msgf("要返回给对方的逻辑节点:%s", one.Vid.B58String())
	}

	// lvidStr :=""
	// for _,one := range vinfos{
	// 	lvidStr +=" "+one.Vid.B58String()
	// }

	// havevidStr := ""
	// for _, one := range idsMap {
	// 	havevidStr += " " + one.B58String()
	// }
	// // headbs, _ := message.Head.JSON()
	// // utils.Log.Info().Msgf("message json:%s", string(headbs))
	// utils.Log.Info().Msgf("查询逻辑节点self:%s 查询:%s havevid:%s", message.Head.RecvId.B58String(), selfVid.B58String(), havevidStr)

	bs, _ := vrs.Marshal() //json.Marshal(vinfos)

	this.messageCenter.SendP2pReplyMsgHE(message, config.MSGID_vnode_getNearSuperIP_recv, &bs)

}

/*
获取相邻节点的Vnode地址 返回
*/
func (this *VnodeCenter) GetNearSuperAddr_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// utils.Log.Info().Msgf("-----------------接收到邻居节点虚拟节点查询消息 返回")
	// vnodeinfos := make([]virtual_node.Vnodeinfo, 0)
	if message.Body.Content == nil {
		return
	}
	vrp := new(go_protobuf.VnodeinfoRepeated)
	err := proto.Unmarshal(*message.Body.Content, vrp)
	if err != nil {
		utils.Log.Info().Msgf("不能解析proto错误:%s", err.Error())
		return
	}

	vnodeinfos := make([]virtual_node.Vnodeinfo, 0)
	for _, one := range vrp.Vnodes {
		vnodeOne := virtual_node.Vnodeinfo{
			Nid:   one.Nid,
			Index: one.Index,
			Vid:   one.Vid,
		}
		vnodeinfos = append(vnodeinfos, vnodeOne)
		// utils.Log.Info().Msgf("解析到邻居的逻辑节点:%s", vnodeOne.Vid.B58String())
	}
	if len(vnodeinfos) < 2 {
		return
	}

	key := ""
	if vnodeinfos[1].Vid == nil || len(vnodeinfos[1].Vid) == 0 {
		key = utils.Bytes2string(vnodeinfos[0].Vid)
	} else {
		key = utils.Bytes2string(append(vnodeinfos[0].Vid, vnodeinfos[1].Vid...))
	}
	// utils.Log.Info().Msgf("返回等待key:%s", hex.EncodeToString([]byte(key)))
	flood.ResponseItr(key, &vnodeinfos)

	// for _, one := range vnodeinfos {
	// 	// fmt.Println("添加一个虚拟节点")
	// 	this.vm.AddLogicVnodeinfo(one)
	// }
	// key := utils.Bytes2string(append(vrp.Vnodes[0].Vid, vrp.Vnodes[1].Vid...))
	// flood.ResponseBytes(key, nil)

}

/*
查询虚拟节点的逻辑节点id
*/
func (this *VnodeCenter) searchVnodeLogicID(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	utils.Log.Info().Msgf("-----------------查询虚拟节点的逻辑节点id onebyone:%t", message.Head.OneByOne)
	var vnode *virtual_node.Vnode
	var getNodeCount uint16

	if message.Body.Content != nil {
		getNodeCount = binary.LittleEndian.Uint16(*message.Body.Content)
	}

	vrs := go_protobuf.VnodeinfoRepeated{
		Vnodes: make([]*go_protobuf.Vnodeinfo, 0),
	}

	if message.Head.OneByOne {
		// if message.Head.RecvId != nil {
		// 	utils.Log.Info().Msgf("searchVnodeLogicID %d  RecvId :%s ", getNodeCount, message.Head.RecvId.B58String())
		// }
		this.vm.RLock()
		for i, _ := range this.vm.VnodeMap {
			if this.vm.VnodeMap[i].Vnode.Index == 0 {
				continue
			}
			if bytes.Equal(this.vm.VnodeMap[i].Vnode.Vid, *message.Head.RecvVnode) {
				vnode = this.vm.VnodeMap[i]
			}
		}
		this.vm.RUnlock()

		if vnode != nil {
			var sendVnode []virtual_node.VnodeinfoS
			vinfo := vnode.GetOnebyoneVnodeInfo()
			if len(vinfo) == 1 {
				has := vinfo[0].Vid
				this.vm.RLock()
				for i, v := range this.vm.VnodeMap {
					if i == 0 || bytes.Equal(v.Vnode.Vid, has) {
						continue
					}
					vinfo = append(vinfo, virtual_node.VnodeinfoS{
						Nid:   v.Vnode.Nid,
						Vid:   v.Vnode.Vid,
						Index: v.Vnode.Index,
					})
				}
				this.vm.RUnlock()
			}
			sortedVnode, times := vnode.GetSortAddressNetExtend(vinfo, vnode.Vnode.Vid)
			//优先把本vnode放进去
			sendVnode = append(sendVnode, sortedVnode[times])
			// if vnode.HeadTail {
			// 	if times < config.GreaterThanSelfMaxConn/2 {
			// 		//虚拟节点在首部
			// 		dVnode := vnode.GetDownVnodeInfo()
			// 		maxTail := len(dVnode) / 2
			// 		mark := 1
			// 		for i := 1; i < len(sortedVnode); i++ {
			// 			//找upvnode，由近及远
			// 			if u := times - i; u >= 0 {
			// 				sendVnode = append(sendVnode, sortedVnode[u])
			// 			} else if mark <= maxTail {
			// 				sendVnode = append(sendVnode, sortedVnode[len(sortedVnode)-mark])
			// 				mark++
			// 			}
			// 			if d := times + i; d < len(sendVnode)-maxTail {
			// 				sendVnode = append(sendVnode, sortedVnode[d])
			// 			}
			// 		}

			// 	} else {
			// 		//虚拟节点在尾部
			// 		uVnode := vnode.GetUpVnodeInfo()
			// 		maxHead := len(uVnode) / 2
			// 		mark := 1
			// 		for i := 1; i < len(sortedVnode); i++ {
			// 			//找upvnode，由近及远
			// 			if u := times - i; u >= maxHead {
			// 				sendVnode = append(sendVnode, sortedVnode[u])
			// 			}
			// 			if d := times + i; d < len(sendVnode)-maxTail {
			// 				sendVnode = append(sendVnode, sortedVnode[d])
			// 			}
			// 		}
			// 	}
			// } else {
			for i := 1; i < len(sortedVnode); i++ {
				//找upvnode，由近及远
				if u := times - i; u >= 0 {
					sendVnode = append(sendVnode, sortedVnode[u])
				}
				if d := times + i; d < len(sortedVnode) {
					sendVnode = append(sendVnode, sortedVnode[d])
				}
			}
			// }
			for i, _ := range sendVnode {
				vrs.Vnodes = append(vrs.Vnodes, &go_protobuf.Vnodeinfo{
					Nid:   sendVnode[i].Nid,
					Vid:   sendVnode[i].Vid,
					Index: sendVnode[i].Index,
				})
			}

			if !(getNodeCount <= 0 || int(getNodeCount) >= len(sendVnode)) {
				vrs.Vnodes = vrs.Vnodes[:getNodeCount]
			}
		}
	}

	// for _, one := range vrs.Vnodes {
	// 	utils.Log.Info().Msgf(" ------ %s", nodeStore.AddressNet(one.Vid).B58String())
	// }
	bs, err := vrs.Marshal()
	if err != nil {
		utils.Log.Info().Msgf("searchVnodeLogicID vnodes proto err:%s", err)
		return
	}

	this.messageCenter.SendVnodeSearchReplyMsg(message, config.MSGID_vnode_searchID_recv, &bs, message.Head.OneByOne)
}

/*
查询虚拟节点的逻辑节点id 返回
*/
func (this *VnodeCenter) searchVnodeLogicID_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(message_center.CLASS_vnode_getstate, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	//utils.Log.Info().Msgf("-----------------接到虚拟节点的逻辑节点id")
	//utils.Log.Info().Msgf("-----------------收到虚拟节点的逻辑节点id Onebyone:%t", message.Head.OneByOne)
	//
	var bs []byte
	if message.Head.OneByOne {
		if message.Body.Content != nil && len(*message.Body.Content) != 0 {
			bs = *message.Body.Content
			//utils.Log.Info().Msgf("2222222", bs)
		}
	} else {
		bs = []byte(*message.Head.SenderVnode)
	}

	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), &bs)
}
