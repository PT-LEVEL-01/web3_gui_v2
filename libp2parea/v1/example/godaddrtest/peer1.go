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
	"web3_gui/utils"
)

func main() {
	utils.PprofMem(time.Minute * 2)

	golog.InitLogger("logother.txt", 0, true)
	golog.Infof("start %s", "log")
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")

	// unit.Simulation()
	// return
	StartAllPeer()
}

var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	host       = "8.210.59.108"
	basePort   = 19960
)

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 5
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

	sleepTime := time.Second * 30
	utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	time.Sleep(sleepTime)

	for {
		for i := range areas {
			showLogicNode(areas[i])
			utils.Log.Info().Msgf("-----------------------------")
		}
		// sendMsgOne(areas[0], &targetID)
		utils.Log.Info().Msgf("qlw----休息20秒后,将再次尝试获取目标地址")
		utils.Log.Info().Msgf("")
		utils.Log.Info().Msgf("")
		time.Sleep(time.Second * 20)
		// utils.Log.Info().Msgf("")
	}
}

func showLogicNode(area *libp2parea.Area) {
	utils.Log.Info().Msgf("show nodes cur:%s------------------------------begin", area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	logicNodes := area.NodeManager.GetAllNodesAddr()
	for i := range logicNodes {
		utils.Log.Info().Msgf("qlw----logic node:%s", logicNodes[i].B58String())
	}
	utils.Log.Info().Msgf("show nodes cur:%s------------------------------end", area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	utils.Log.Info().Msgf("")
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
	// area.SetNetTypeToTest()

	// area.OpenVnode()

	//serverHost
	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))
	// area.CloseVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnodeByIndex(100)
	// area.Vm.AddVnodeByIndex(200)
	// area.Vm.AddVnodeByIndex(300)
	// area.Vm.AddVnodeByIndex(400)

	// vnodeinfo := area.Vm.AddVnodeByIndex(100)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// vnodeinfo = area.Vm.AddVnodeByIndex(200)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// vnodeinfo = area.Vm.AddVnodeByIndex(300)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// vnodeinfo = area.Vm.AddVnodeByIndex(400)
	// utils.Log.Info().Msgf("添加虚拟节点self:%s vid:%s", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())
	// golog.Infof("添加虚拟节点self:%s vid:%s\n", vnodeinfo.Nid.B58String(), vnodeinfo.Vid.B58String())

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
}

func sendMsgOne(area *libp2parea.Area, toArea *libp2parea.Area) error {
	//发送p2p消息
	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	toNetid := toArea.GetNetId()
	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toNetid, nil, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
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

	area.Register_nodeClosedCallback(this.NodeCloseCallbackFunc)
	area.Register_nodeBeenGodAddrCallback(this.NodeBeenGodAddrCallbackFunc)
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
	utils.Log.Info().Msgf("收到广播消息:%s", this.area.GetNetId().B58String())
}

func (this *TestPeer) NodeCloseCallbackFunc(addr nodeStore.AddressNet, machineID string) {
	utils.Log.Info().Msgf("qlw----- 节点关闭连接回调 cur:%s, node:%s, machineID:%s", this.area.GetNetId().B58String(), addr.B58String(), machineID)
}

func (this *TestPeer) NodeBeenGodAddrCallbackFunc(addr nodeStore.AddressNet, machineID string) {
	utils.Log.Info().Msgf("qlw----- 节点被设置为上帝地址回调 cur:%s, node:%s, machineID:%s", this.area.GetNetId().B58String(), addr.B58String(), machineID)
}
