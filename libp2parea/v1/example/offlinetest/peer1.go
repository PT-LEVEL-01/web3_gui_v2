package main

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hyahm/golog"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

func main() {
	utils.PprofMem(time.Minute * 2)

	golog.InitLogger("logother.txt", 0, true)
	golog.Infof("start %s", "log")
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")

	StartAllPeer()
}

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19960
)

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 10
	areas := make([]*libp2parea.Area, 0, n)
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		areaPeers = append(areaPeers, area)
		areas = append(areas, area.area)
		nsm.AddNodeSuperIDs(area.area.GetNetId())
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始等待节点自治")
	//等待各个节点都准备好
	for _, one := range areaPeers {
		one.area.WaitAutonomyFinish()
		one.area.WaitAutonomyFinishVnode()
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("节点自治完成，打印逻辑节点")

	sleepTime := time.Second * 10
	utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	time.Sleep(sleepTime)

	//测试磁力节点消息到达是否准确
	SearchPeerTest(areas, nsm)
	SearchVnodeTest(areas)

	utils.Log.Info().Msgf("qlw-----sleep 10 seconds")
	time.Sleep(time.Second * 10)

	//
	utils.Log.Info().Msgf("qlw-----delete area test-----begin")
	index := 0
	for {
		index++
		SearchVnodeTest(areas)
		utils.Log.Info().Msgf("qlw-----新一轮的信息展示: %d", index)
		time.Sleep(time.Second * 20)

		if index == 3 {
			if _, exist := areaPeers[0].area.Vm.VnodeMap[100]; exist {
				areaPeers[0].area.Vm.DelVnodeByIndex(100)
			}
		}

		if index >= 10 {
			break
		}
	}

	select {}

}

func SearchPeerTest(areas []*libp2parea.Area, nsm nodeStore.NodeSimulationManager) {
	utils.Log.Info().Msgf("Simulation start")
	//构建标准的逻辑节点
	nsm.BuildNodeLogicIDs()
	//模拟节点发现过程
	nsm.Discover()
	//打印各个自定义节点的逻辑节点
	nsm.PrintlnLogicNodesNew(true)
	//发送P2P消息
	// nsm.MsgPingNodesP2P()
	utils.Log.Info().Msgf("---------------------------")
	//发送搜索磁力节点消息
	// utils.Log.Info().Msgf("qlw------------仿真发送消息---start")
	// nsm.MsgPingNodesSearch()
	// utils.Log.Info().Msgf("qlw------------仿真发送消息---end")

	//对比标准节点和自定义节点，各方保存的逻辑节点差异
	nodeStore.EqualNodes(&nsm)

	utils.Log.Info().Msgf("Simulation end")

	//对比仿真的逻辑节点和真实的逻辑节点是否相同
	utils.Log.Info().Msgf("对比仿真逻辑节点和真实逻辑节点是否相同")
	nodeMap := nsm.GetCustomNodes()
	for _, areaOne := range areas {
		ns, ok := nodeMap[utils.Bytes2string(areaOne.GetNetId())]
		if !ok {
			utils.Log.Info().Msgf("未找到节点")
			panic("未找到节点！")
		}
		logicNodes := areaOne.NodeManager.GetLogicNodes()
		ids := make([][]byte, 0, len(logicNodes))
		for _, one := range logicNodes {
			ids = append(ids, one)
		}
		isChange, _ := nodeStore.EqualIds(ids, ns.Logic)
		if isChange {
			utils.Log.Info().Msgf("--------------------------------------")
			utils.Log.Info().Msgf("逻辑节点不相等:%s", areaOne.GetNetId().B58String())
			for _, one := range ns.Logic {
				utils.Log.Info().Msgf("仿真逻辑节点:%s", nodeStore.AddressNet(one).B58String())
			}
			utils.Log.Info().Msgf("")
			for _, one := range logicNodes {
				utils.Log.Info().Msgf("真实逻辑节点:%s", one.B58String())
			}
		}
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("对比逻辑节点结束")
}

func SearchVnodeTest(areas []*libp2parea.Area) {
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("虚拟节点：打印每个节点的虚拟节点地址")
	for _, area := range areas {
		vnodeinfos := area.Vm.GetVnodeSelf()
		for _, vnode := range vnodeinfos {
			utils.Log.Info().Msgf("虚拟地址:%s %d", vnode.Vid.B58String(), vnode.Index)
		}
		utils.Log.Info().Msgf("--")
	}

	utils.Log.Info().Msgf("--------------------------")
	// utils.Log.Info().Msgf("虚拟节点：打印标准逻辑节点")
	// nsm.PrintlnStandardLogicNodesNew(true)

	// //多次启动来对比各个节点的逻辑节点差异
	// // unit.EquleLogicID(areas)
	for _, area := range areas {
		vnodes := area.Vm.GetVnodeSelf()
		for _, one := range vnodes {
			utils.Log.Info().Msgf("本虚拟节点地址:%s", one.Vid.B58String())
			vnode := area.Vm.FindVnodeInSelf(one.Vid)
			vnode.LogicalNode.Range(func(k, v interface{}) bool {
				vnodeInfo := v.(virtual_node.Vnodeinfo)
				utils.Log.Info().Msgf("  逻辑虚拟节点:%s, nid:%s", vnodeInfo.Vid.B58String(), vnodeInfo.Nid.B58String())
				return true
			})
			vnode.LogicalNodeIndexNID.Range(func(k, v interface{}) bool {
				nidMap := v.(*sync.Map)
				nidMap.Range(func(k1, v1 interface{}) bool {
					vnodeInfo := v1.(*virtual_node.Vnodeinfo)
					utils.Log.Info().Msgf("  逻辑虚拟节点2:%s, nid:%s", vnodeInfo.Vid.B58String(), vnodeInfo.Nid.B58String())
					return true
				})
				return true
			})
		}

		utils.Log.Info().Msgf("---开始打印客户端节点-----")
		clientVnodes := area.Vm.GetClientVnodeinfo()
		for _, one := range clientVnodes {
			utils.Log.Info().Msgf("虚拟地址2:%s, nid:%s, index:%d", one.Vid.B58String(), one.Nid.B58String(), one.Index)
		}

		utils.Log.Info().Msgf("本虚拟节点地址:end --------------------------------")
	}

	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始发送虚拟节点磁力消息")
}

type TestPeer struct {
	area *libp2parea.Area
}

func StartOnePeer(i int) *TestPeer {
	keyPath1 := filepath.Join("conf", "keystore"+strconv.Itoa(i)+".key")

	key1 := keystore.NewKeystore(keyPath1, addrPre)
	err := key1.Load()
	if err != nil {
		//没有就创建
		err = key1.CreateNewKeystore(keyPwd)
		if err != nil {
			panic("创建key1错误:" + err.Error())
		}
	}

	if key1.NetAddr == nil {
		_, _, err = key1.CreateNetAddr(keyPwd, keyPwd)
		if err != nil {
			panic("创建NetAddr错误:" + err.Error())
		}
	}
	if len(key1.GetAddr()) < 1 {
		_, err = key1.GetNewAddr(keyPwd, keyPwd)
		if err != nil {
			panic("创建Addr错误:" + err.Error())
		}
	}
	if len(key1.GetDHKeyPair().SubKey) < 1 {
		_, err = key1.GetNewDHKey(keyPwd, keyPwd)
		if err != nil {
			panic("创建Addr错误:" + err.Error())
		}
	}

	area, err := libp2parea.NewArea(areaName, key1, keyPwd)
	if err != nil {
		panic(err.Error())
	}
	area.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	area.SetNetTypeToTest()

	area.OpenVnode()

	//serverHost
	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))

	area.Vm.AddVnodeByIndex(100)
	area.Vm.AddVnodeByIndex(200)
	area.Vm.AddVnodeByIndex(300)
	area.Vm.AddVnodeByIndex(400)
	// area.Vm.AddVnodeByIndex(500)

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
}

var rand = utils.GetHashForDomain(utils.GetRandomDomain())
var magneticIds []nodeStore.AddressNet
var magneticVnodeIds []virtual_node.AddressNetExtend

func sendMsgOne(area *libp2parea.Area, toArea *libp2parea.Area) error {
	//发送p2p消息
	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	toNetid := toArea.GetNetId()
	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toNetid, nil, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
	}

	//发送虚拟节点p2p消息
	toVnodes := toArea.Vm.GetVnodeSelf()
	for _, one := range toVnodes {
		utils.Log.Info().Msgf("发送虚拟节点id:%s to:%s", area.Vm.GetVnodeDiscover().Vnode.Vid.B58String(), one.Vid.B58String())
		_, err := area.SendVnodeP2pMsgHEWaitRequest(msg_id_vnode_p2p, &area.Vm.GetVnodeDiscover().Vnode.Vid, &one.Vid, &one.Nid, nil, time.Second*2)
		if err != nil {
			utils.Log.Error().Msgf("发送VnodeP2p消息失败:%s", err.Error())
			return err
		}
	}

	//发送广播消息
	// utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	err = area.SendMulticastMsg(msg_id_multicast, nil)
	if err != nil {
		utils.Log.Error().Msgf("发送广播消息失败:%s", err.Error())
	}

	return nil
}

func (this *TestPeer) sendMsgLoop(area *libp2parea.Area, toAddrs []*TestPeer, group *sync.WaitGroup) {
	utils.Log.Info().Msgf("开始一个节点的发送")
	for _, one := range toAddrs {
		if bytes.Equal(area.GetNetId(), one.area.GetNetId()) {
			continue
		}
		done := false
		for !done {
			err := sendMsgOne(area, one.area)
			if err != nil {
				//当网络发送失败，等待1秒钟后，继续发送
				time.Sleep(time.Second)
			} else {
				done = true
				time.Sleep(time.Second / 100)
			}
		}
	}
	group.Done()
	utils.Log.Info().Msgf("发送完一个节点")
}

const msg_id_p2p = 1001
const msg_id_p2p_recv = 1002 //加密消息

const msg_id_searchSuper = 1003
const msg_id_searchSuper_recv = 1004 //加密消息

const msg_id_vnode_p2p = 1005
const msg_id_vnode_p2p_recv = 1006 //加密消息

const msg_id_vnode_search = 1007      //搜索节点消息
const msg_id_vnode_search_recv = 1008 //搜索节点消息 返回

const msg_id_multicast = 1009 //

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_searchSuper, this.SearchSuperHandler)
	area.Register_p2p(msg_id_searchSuper_recv, this.SearchSuperHandler_recv)

	area.Register_vnode_p2pHE(msg_id_vnode_p2p, this.RecvMsgHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_p2p_recv, this.RecvMsgHEHandler)

	area.Register_vnode_search(msg_id_vnode_search, this.SearchVnodeHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_search_recv, this.SearchVnodeHandler_recv)

	area.Register_multicast(msg_id_multicast, this.MulticastMsgHandler)
}

var MsgCountLock = new(sync.Mutex)
var MsgCount = make(map[string]uint64)

func (this *TestPeer) RecvP2PMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	selfId := message.Head.RecvId.B58String()
	if !bytes.Equal(this.area.GetNetId(), *message.Head.RecvId) {
		utils.Log.Error().Msgf("收到错误p2p消息self:%s recv:%s", this.area.GetNetId().B58String(), message.Head.RecvId.B58String())
		return
	}
	// utils.Log.Info().Msgf("收到P2P消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	MsgCountLock.Lock()
	count, ok := MsgCount[selfId]
	if ok {
		MsgCount[selfId] = count + 1
	} else {
		MsgCount[selfId] = 1
	}
	MsgCountLock.Unlock()
	this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, nil)
}

func (this *TestPeer) RecvP2PMsgHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func (this *TestPeer) SearchSuperHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	selfId := message.Head.RecvId.B58String()
	// utils.Log.Info().Msgf("收到SearchSuper消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	MsgCountLock.Lock()
	count, ok := MsgCount[selfId]
	if ok {
		MsgCount[selfId] = count + 1
	} else {
		MsgCount[selfId] = 1
	}
	MsgCountLock.Unlock()
	err := this.area.SendSearchSuperReplyMsg(message, msg_id_searchSuper_recv, nil)
	if err != nil {
		utils.Log.Error().Msgf("SendSearchSuperReplyMsg error:%s", err.Error())
	}
}

func (this *TestPeer) SearchSuperHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到SearchSuper消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func (this *TestPeer) RecvMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// selfId := message.Head.RecvId.B58String()
	// vnodeinfo := vnc.FindInVnodeinfoSelf(*this.Head.RecvVnode)
	if this.area.Vm.FindInVnodeSelf(*message.Head.RecvVnode) {
		utils.Log.Info().Msgf("收到vnode p2p消息self:%s from:%s to:%s", this.area.GetNetId().B58String(),
			message.Head.Sender.B58String(), message.Head.RecvVnode.B58String())
	} else {
		utils.Log.Error().Msgf("收到错误vnode p2p消息self:%s from:%s to:%s", this.area.GetNetId().B58String(),
			message.Head.Sender.B58String(), message.Head.RecvVnode.B58String())
		return
	}
	err := this.area.SendVnodeP2pReplyMsgHE(message, msg_id_vnode_p2p_recv, nil)
	if err != nil {
		utils.Log.Info().Msgf("SendVnodeP2pReplyMsgHE error:%s", err.Error())
	} else {

	}
}

func (this *TestPeer) RecvMsgHEHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到vnode p2p消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func (this *TestPeer) SearchVnodeHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// selfId := message.Head.RecvId.B58String()
	selfId := message.Head.RecvVnode.B58String()
	// utils.Log.Info().Msgf("收到 VnodeSearch 消息 self:%s from:%s", this.area.GetNetId().B58String(), message.Head.Sender.B58String())
	MsgCountLock.Lock()
	count, ok := MsgCount[selfId]
	if ok {
		MsgCount[selfId] = count + 1
	} else {
		MsgCount[selfId] = 1
	}
	MsgCountLock.Unlock()
	err := this.area.SendVnodeSearchReplyMsg(message, msg_id_vnode_search_recv, nil)
	if err != nil {
		utils.Log.Info().Msgf("回复消息错误:%s", err.Error())
	}
}

func (this *TestPeer) SearchVnodeHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到 VnodeSearch 消息返回 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func (this *TestPeer) MulticastMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
}
