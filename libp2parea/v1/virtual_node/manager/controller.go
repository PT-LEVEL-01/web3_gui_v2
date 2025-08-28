package manager

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"sync"
	"time"

	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/protobuf/go_protobuf"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

type VnodeCenter struct {
	sessionEngine     *engine.Engine                //
	nodeManager       *nodeStore.NodeManager        //
	messageCenter     *message_center.MessageCenter //
	vm                *virtual_node.VnodeManager    //
	contextRoot       context.Context               //
	isClose           bool                          //是否关闭虚拟节点功能
	getVnodeinfoChan  chan bool                     //
	sendVnodeInfoChan chan bool                     //发送本节点的所有虚拟节点地址
}

func NewVnodeCenter(messageCenterParam *message_center.MessageCenter,
	nodeManagerParam *nodeStore.NodeManager, sessionEngineParam *engine.Engine, vm *virtual_node.VnodeManager, c context.Context) *VnodeCenter {
	vc := VnodeCenter{
		sessionEngine:     sessionEngineParam,
		nodeManager:       nodeManagerParam,
		messageCenter:     messageCenterParam,
		vm:                vm,
		contextRoot:       c,
		getVnodeinfoChan:  make(chan bool, 1),
		sendVnodeInfoChan: make(chan bool, 1),
	}
	vc.messageCenter.Register_p2pHE(config.MSGID_vnode_getstate, vc.GetVnodeOpenState, true)                  //获取节点的虚拟节点开通状态
	vc.messageCenter.Register_p2pHE(config.MSGID_vnode_getstate_recv, vc.GetVnodeOpenState_recv, true)        //获取节点的虚拟节点开通状态 返回
	vc.messageCenter.Register_p2pHE(config.MSGID_vnode_getNearSuperIP, vc.GetNearSuperAddr, true)             //获取相邻节点的Vnode地址
	vc.messageCenter.Register_p2pHE(config.MSGID_vnode_getNearSuperIP_recv, vc.GetNearSuperAddr_recv, true)   //获取相邻节点的Vnode地址 返回
	vc.messageCenter.Register_vnode_search(config.MSGID_vnode_searchID, vc.searchVnodeLogicID, true)          //查询虚拟节点的逻辑节点id
	vc.messageCenter.Register_vnode_p2pHE(config.MSGID_vnode_searchID_recv, vc.searchVnodeLogicID_recv, true) //查询虚拟节点的逻辑节点id 返回

	go vc.LoopGetVnodeinfo()
	go vc.FindNearVnode()
	go vc.CloseVnodeDeal()
	go vc.LoopSendVnodeInfo()
	return &vc
}

/*
查询邻居节点
*/
func (this *VnodeCenter) FindNearVnode() {
	c := this.vm.GetFindVnodeChan()
	// utils.Log.Info().Msgf("开始接收查询邻居节点消息信号")
	var one *[]*virtual_node.FindVnodeVO
	for {
		select {
		case <-this.contextRoot.Done():
			return
		case one = <-c:
			// utils.Log.Info().Msgf("1111111111")
		}
		if this.isClose {
			// utils.Log.Info().Msgf("FindNearVnode is close")
			continue
		}
		// utils.Log.Info().Msgf("1111111111")
		this.GetRemoteVnodeinfo(one)
	}
	// utils.Log.Info().Msgf("停止接收查询邻居节点消息信号")
}

/*
向逻辑节点发送消息，获取自己的逻辑节点
*/
func (this *VnodeCenter) GetRemoteVnodeinfo(vnodes *[]*virtual_node.FindVnodeVO) {
	var err error
	var bs []byte
	recvLogicVnodes := make([]virtual_node.Vnodeinfo, 0)
	group := new(sync.WaitGroup)
	group.Add(len(*vnodes))
	for _, one := range *vnodes {
		key := ""
		if one.Target.Vid == nil || len(one.Target.Vid) == 0 {
			key = utils.Bytes2string(one.Self.Vid)
		} else {
			key = utils.Bytes2string(append(one.Self.Vid, one.Target.Vid...))
		}
		// utils.Log.Info().Msgf("注册等待key:%s", hex.EncodeToString([]byte(key)))
		flood.RegisterRequest(key)
		bs, err = one.Proto()
		if err != nil {
			// utils.Log.Info().Msgf("1111111111")
			continue
		} else {
			// utils.Log.Info().Msgf("33333333333")
		}
		// utils.Log.Info().Msgf("原文:%d %s", len(bs), hex.EncodeToString(bs))

		if one.Target.Nid == nil || len(one.Target.Nid) == 0 {
			// utils.Log.Info().Msgf("1111111111")
			logicNodes := this.nodeManager.GetLogicNodes()
			for i, _ := range logicNodes {
				two := logicNodes[i]
				// utils.Log.Info().Msgf("开始查询邻居节点 %s", one.B58String())
				_, _, _, err = this.messageCenter.SendP2pMsgHE(config.MSGID_vnode_getNearSuperIP, &two, &bs)
			}
		} else {
			// utils.Log.Info().Msgf("222222222222")
			// utils.Log.Info().Msgf("开始查询邻居节点 %s", one.Target.Nid.B58String())
			_, _, _, err = this.messageCenter.SendP2pMsgHE(config.MSGID_vnode_getNearSuperIP, &one.Target.Nid, &bs)
		}
		if err != nil {
			utils.Log.Info().Msgf("send GetRemoteVnodeinfo error:%s", err.Error())
			group.Done()
			continue
		}
		itr, err := flood.WaitResponseItr(key, time.Second*10)
		if err != nil {
			utils.Log.Info().Msgf("send GetRemoteVnodeinfo error:%s %s", err.Error(), hex.EncodeToString([]byte(key)))
			group.Done()
			continue
		}
		vnodeinfos := itr.(*[]virtual_node.Vnodeinfo)
		temp := (*vnodeinfos)[2:]
		for i, _ := range temp {
			one := temp[i]
			recvLogicVnodes = append(recvLogicVnodes, one)
		}

		// tempStr := ""
		// for _, one := range temp {
		// 	tempStr += " " + one.Vid.B58String()
		// }
		// utils.Log.Info().Msgf("收到逻辑节点地址self:%s vid:%s", (*vnodeinfos)[0].Vid.B58String(), tempStr)

		group.Done()
	}
	group.Wait()

	// 依次检查虚拟节点是否存在
	// 保存有效的逻辑地址节点列表
	validLogicVnodes := make([]virtual_node.Vnodeinfo, 0)
	for _, one := range recvLogicVnodes {
		if this.CheckVnodeIsOnline(&one) {
			validLogicVnodes = append(validLogicVnodes, one)
		}
	}

	this.vm.AddLogicVnodeinfos((*vnodes)[0].Self.Vid, &validLogicVnodes)
}

/*
触发：定时获取邻居节点的虚拟节点地址
*/
func (this *VnodeCenter) TriggerLoopGetVnodeinfo() {
	select {
	case this.getVnodeinfoChan <- false:
	default:
	}
}

/*
定时获取邻居节点的虚拟节点地址
*/
func (this *VnodeCenter) LoopGetVnodeinfo() {
	// newTiker := utils.NewBackoffTimerChan(config.VNODE_get_neighbor_vnode_tiker...)
	// open := false
	for {
		<-this.getVnodeinfoChan

		// if newTiker.Wait(this.contextRoot) == 0 {
		// 	return
		// }
		//功能未开启则退出
		// if len(this.vm.GetVnodeNumber()) <= 0 {
		// 	// open = false
		// 	continue
		// }
		//虚拟节点功能刚打开，重新设置同步时间
		// if !open {
		// 	newTiker.Reset()
		// }
		// open = true
		nodes := this.nodeManager.GetLogicNodes()
		for _, one := range nodes {
			nodeinfo := virtual_node.Vnodeinfo{
				Nid:   one,                                //节点真实网络地址
				Index: 0,                                  //节点第几个空间，从1开始,下标为0的节点为实际节点。
				Vid:   virtual_node.AddressNetExtend(one), //vid，虚拟节点网络地址
			}
			if this.CheckVnodeIsOnline(&nodeinfo) {
				this.vm.AddLogicVnodeinfo(nodeinfo)
			}
		}
	}
}

/*
 * 广播关闭虚拟节点消息
 */
func (this *VnodeCenter) CloseVnodeDeal() {
	c := this.vm.GetCloseVnodeChan()
	// utils.Log.Info().Msgf("开始接收查询邻居节点消息信号")
	var one *virtual_node.Vnodeinfo
	for {
		select {
		case <-this.contextRoot.Done():
			return
		case one = <-c:
			// utils.Log.Info().Msgf("1111111111")
		}
		if this.isClose {
			// utils.Log.Info().Msgf("FindNearVnode is close")
			continue
		}
		// utils.Log.Info().Msgf("1111111111")
		this.BroadcastCloseVnode(one)
	}
}

/*
 * 触发：定时发送本节点的所有虚拟节点地址
 */
func (this *VnodeCenter) TriggerLoopSendVnodeinfo() {
	select {
	case this.sendVnodeInfoChan <- false:
	default:
	}
}

/*
 * 广播本节点的所有虚拟节点信息消息
 */
func (this *VnodeCenter) LoopSendVnodeInfo() {
	// utils.Log.Info().Msgf("开始广播本节点的所有虚拟节点信息消息")
	for {
		select {
		case <-this.contextRoot.Done():
			return
		case <-this.sendVnodeInfoChan:
			// utils.Log.Info().Msgf("1111111111")
		}
		if this.isClose {
			// utils.Log.Info().Msgf("LoopSendVnodeInfo is close")
			continue
		}
		// utils.Log.Info().Msgf("1111111111")
		this.BroadcastSelfVnodeInfo()
	}
}

/*
	添加一个新节点，发送消息看这个节点是否开通了虚拟节点
	已开通则添加这个节点，未开通则抛弃。
*/
// func (this *VnodeCenter) AddNewNode(addr nodeStore.AddressNet) {

// 	utils.Go(func() {
// 		// utils.Log.Info().Msgf("添加新节点")
// 		//自己节点没开通虚拟节点，就不需要添加
// 		// if len(this.vm.GetVnodeNumber()) <= 0 {
// 		// 	utils.Log.Info().Msgf("自己未开通虚拟节点功能")
// 		// 	return
// 		// }
// 		//判断这个节点是否已经添加，已经添加则不需要重复添加
// 		vnodeinfoMap := this.vm.GetVnodeLogical()
// 		for _, v := range vnodeinfoMap {
// 			if bytes.Equal(v.Nid, addr) {
// 				// utils.Log.Info().Msgf("1111111111111")
// 				return
// 			}
// 		}
// 		//查询节点是否开通了虚拟节点
// 		bs, ok, isSelf, err := this.messageCenter.SendP2pMsgHEWaitRequest(config.MSGID_vnode_getstate, &addr, nil, time.Second*10)
// 		if err != nil {
// 			// utils.Log.Info().Msgf("22222222222222222")
// 			return
// 		}
// 		if !ok || isSelf {
// 			// utils.Log.Info().Msgf("33333333333333333")
// 			return
// 		}
// 		if bs == nil || len(*bs) <= 0 {
// 			// utils.Log.Info().Msgf("44444444444444444")
// 			return
// 		}
// 		index := utils.BytesToUint64(*bs)

// 		for i := uint64(0); i < index; i++ {
// 			vnodeinfo := virtual_node.BuildNodeinfo(i, addr)
// 			this.vm.AddLogicVnodeinfo(*vnodeinfo)
// 		}
// 	})
// }

/*
通知有新节点加入网络
*/
func (this *VnodeCenter) NoticeAddNode(addr nodeStore.AddressNet) {
	nodeinfo := virtual_node.BuildNodeinfo(0, addr)
	this.vm.AddLogicVnodeinfo(*nodeinfo)
	// this.AddNewNode(addr)
}

/*
通知有节点离线
*/
func (this *VnodeCenter) NoticeRemoveNode(addr nodeStore.AddressNet) {
	this.vm.DiscoverVnodes.DeleteNid(addr)
	this.vm.RLock()
	defer this.vm.RUnlock()
	for _, one := range this.vm.VnodeMap {
		one.DeleteNid(addr)
	}
}

/*
关闭vnode
*/
func (this *VnodeCenter) Close() {
	this.isClose = true
	this.vm.IsClose()
}

/*
打开vnode
*/
func (this *VnodeCenter) Open() {
	this.isClose = false
	this.vm.IsOpen()
}

/*
查找虚拟逻辑节点的真实地址
@nodeId    *AddressNetExtend     要查找的节点
*/
func (this *VnodeCenter) SearchVnodeId(nodeId *virtual_node.AddressNetExtend, recvProxyId, senderProxyId *nodeStore.AddressNet, onebyone bool, num uint16) (*[]byte, error) {
	var content []byte
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		return nil, err
	}
	content = buf.Bytes()
	bs, err := this.messageCenter.SendVnodeSearchMsgWaitRequest(config.MSGID_vnode_searchID, &this.vm.GetVnodeDiscover().Vnode.Vid, nodeId, recvProxyId, senderProxyId, &content, time.Second*10, onebyone)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

/*
 * 检查虚拟节点是否在线
 * 	@param	vnodeInfo 	需要检查的虚拟节点信息
 *	@return bool		是否在线
 */
func (this *VnodeCenter) CheckVnodeIsOnline(vnodeInfo *virtual_node.Vnodeinfo) bool {
	// 虚拟节点信息有误，直接返回
	if vnodeInfo == nil || vnodeInfo.Vid == nil || len(vnodeInfo.Vid) == 0 || vnodeInfo.Nid == nil || len(vnodeInfo.Nid) == 0 {
		return false
	}

	// 发送消息，确认该虚拟节点是否在线
	content := []byte(vnodeInfo.Vid)
	res, _, _, err := this.messageCenter.SendP2pMsgWaitRequest(config.MSGID_checkVnodeAddrOnline, &vnodeInfo.Nid, &content, 5*time.Second)
	if err != nil {
		// utils.Log.Info().Msgf("send CheckVnodeIsOnline vid:%s nid:%s error:%s", vnodeInfo.Vid.B58String(), vnodeInfo.Nid.B58String(), err.Error())
		return false
	}

	if res == nil || len(*res) == 0 {
		// utils.Log.Info().Msgf("CheckVnodeIsOnline: get vnode online result err: vnodeid:%s", vnodeInfo.Vid.B58String())
		return false
	}

	// 根据结果判断虚拟节点是否在线
	if (*res)[0] != config.VNodeIdResult_online {
		// utils.Log.Info().Msgf("qlw---CheckVnodeIsOnline: vnodid:%s, nid:%s offline----", vnodeInfo.Vid.B58String(), vnodeInfo.Nid.B58String())
		return false
	}

	// utils.Log.Info().Msgf("qlw---CheckVnodeIsOnline: vnodid:%s, nid:%s online----", vnodeInfo.Vid.B58String(), vnodeInfo.Nid.B58String())
	return true
}

/*
 * 广播通知虚拟节点关闭信息
 */
func (this *VnodeCenter) BroadcastCloseVnode(vnode *virtual_node.Vnodeinfo) {
	if vnode == nil || vnode.Nid == nil || len(vnode.Nid) == 0 || vnode.Vid == nil || len(vnode.Vid) == 0 {
		return
	}

	// 1. 删除发现节点对应的逻辑节点信息
	// 1.1 删除发现节点中的逻辑节点信息
	this.vm.DiscoverVnodes.DeleteVid(vnode)
	// 1.2 删除其他虚拟节点对应的逻辑节点信息
	this.vm.RLock()
	for _, one := range this.vm.VnodeMap {
		// 如果是自己则不处理，因为所有删除节点的地方都会进行删除操作
		if bytes.Equal(one.Vnode.Vid, vnode.Vid) {
			continue
		}
		one.DeleteVid(vnode)
	}
	this.vm.RUnlock()

	// 2. 广播通知虚拟节点已下线
	// 2.1 把要删除的虚拟节点广播出去
	bs, err := vnode.Proto()
	if err != nil {
		// utils.Log.Info().Msgf("BroadcastCloseVnode vnode proto err:%s", err)
		return
	}
	if err := this.messageCenter.SendMulticastMsg(config.MSGID_multicast_vnode_offline_recv, &bs); err != nil {
		// utils.Log.Info().Msgf("发送虚拟节点下线广播消息 err:%s", err)
	}
}

/*
 * @author qlw
 * 广播通知本节点的所有虚拟节点信息
 */
func (vc *VnodeCenter) BroadcastSelfVnodeInfo() {
	// utils.Log.Info().Msgf("qlw----开始广播通知本节点的所有虚拟节点信息")
	// 1. 获取所有的虚拟节点信息
	vinfos := make([]virtual_node.Vnodeinfo, 0)
	vc.vm.RLock()
	for _, one := range vc.vm.VnodeMap {
		if one.Vnode.Index == 0 {
			continue
		}
		vinfos = append(vinfos, one.Vnode)
	}
	vc.vm.RUnlock()

	// 2. 如果不存在虚拟节点，则不用广播
	if len(vinfos) == 0 {
		// utils.Log.Info().Msgf("qlw---BroadcastSelfVnodeInfo 该节点没有虚拟节点信息, 因此不需要发送广播信息!")
		return
	}

	// 3. 广播通知自己的所有虚拟节点信息
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
	}

	// 4. 构建字节流
	bs, err := vrs.Marshal()
	if err != nil {
		utils.Log.Info().Msgf("BroadcastSelfVnodeInfo vnodes proto err:%s", err)
		return
	}

	// 5. 发送广播信息
	if err := vc.messageCenter.SendMulticastMsg(config.MSGID_multicast_send_vnode_recv, &bs); err != nil {
		utils.Log.Info().Msgf("广播通知本节点的所有虚拟节点信息 err:%s", err)
	}
}
