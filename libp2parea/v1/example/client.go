package main

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"web3_gui/libp2parea/v1/nodeStore"

	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/message_center/flood"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

func main() {
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")
	StartAllPeer()
}

var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	host       = "127.0.0.1"
	basePort   = 19960
)

/*
启动所有节点
*/
func StartAllPeer() {
	n := 9

	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		areaPeers = append(areaPeers, area)
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
	for _, one := range areaPeers {
		utils.Log.Info().Msgf("本节点地址:%s", one.area.GetNetId().B58String())
		logicNodes := one.area.Vm.FindLogicInVnodeSelf(one.area.Vm.GetVnodeSelf()[0].Vid)
		for _, one := range logicNodes {
			utils.Log.Info().Msgf("逻辑节点:%s %d %s", one.Nid.B58String(), one.Index, one.Vid.B58String())
		}
	}

	//--------------------------
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始发送消息")
	//异步发消息
	group := new(sync.WaitGroup)
	group.Add(len(areaPeers))
	// <-time.NewTimer(time.Second * 5).C
	for i, _ := range areaPeers {
		one := areaPeers[i]
		one.sendMsgLoop(one.area, areaPeers, group)
	}
	group.Wait()
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("发送消息完成")

	addrs := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(&areaPeers[0].area.Vm.GetVnodeDiscover().Vnode.Vid)
	for i, _ := range addrs {
		logicNetid, _ := areaPeers[0].area.SearchVnodeId(addrs[i])
		utils.Log.Info().Msgf("查找的逻辑节点地址:%s", logicNetid.B58String())
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("查找逻辑节点地址完成")

	for {
		time.Sleep(time.Second * 20)
		countNum := 1
		for k, v := range MsgCount {
			utils.Log.Info().Msgf("%d节点：%s 收到:%d", countNum, k, v)
			countNum++
		}
		utils.Log.Info().Msgf("-----------------------")
	}

	// select {}

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
	// area.SetNetTypeToTest()

	// area.CloseVnode()

	area.SetDiscoverPeer(serverHost + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	// area.Vm.AddVnode()
	area.Vm.AddVnodeByIndex(100)
	area.Vm.AddVnodeByIndex(200)
	area.Vm.AddVnodeByIndex(300)
	area.Vm.AddVnodeByIndex(400)

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
}

func sendMsgOne(area *libp2parea.Area, toArea *libp2parea.Area) error {
	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	toNetid := toArea.GetNetId()
	_, _, _, err := area.SendP2pMsgWaitRequest(msg_id_p2p, &toNetid, nil, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
	}

	fromNetid := area.GetNetId()
	logicids := nodeStore.GetQuarterLogicAddrNetByAddrNet(&fromNetid)
	for _, one := range logicids {
		_, err := area.SendSearchSuperMsgWaitRequest(msg_id_searchSuper, one, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
			return err
		}
	}

	utils.Log.Info().Msgf("发送虚拟节点id:%s", area.Vm.GetVnodeDiscover().Vnode.Vid.B58String())
	toVnodes := area.Vm.GetVnodeSelf()
	for _, one := range toVnodes {
		_, err := area.SendVnodeP2pMsgHEWaitRequest(msg_id_vnode_p2p, &area.Vm.GetVnodeDiscover().Vnode.Vid, &one.Vid, nil, time.Second*10)
		if err != nil {
			utils.Log.Error().Msgf("发送VnodeP2p消息失败:%s", err.Error())
			return err
		}
	}

	addrs := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(&area.Vm.GetVnodeDiscover().Vnode.Vid)
	for _, one := range addrs {
		utils.Log.Info().Msgf("===--------- start")
		// utils.Log.Info().Msgf("先打印逻辑节点")
		// logicNodes := area.Vm.FindLogicInVnodeSelf(area.Vm.GetVnodeSelf()[0].Vid)
		// for _, one := range logicNodes {
		// 	utils.Log.Info().Msgf("逻辑节点:%s %d %s", one.Nid.B58String(), one.Index, one.Vid.B58String())
		// }
		_, err := area.SendVnodeSearchMsgWaitRequest(msg_id_vnode_search, &area.Vm.GetVnodeSelf()[0].Vid, one, nil, time.Second*10)
		if err != nil {
			utils.Log.Info().Msgf("发送VnodeSearchMsg error:%s", err.Error())
			return err
		}
		utils.Log.Info().Msgf("===--------- end")
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

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_searchSuper, this.SearchSuperHandler)
	area.Register_p2p(msg_id_searchSuper_recv, this.SearchSuperHandler_recv)

	area.Register_vnode_p2pHE(msg_id_vnode_p2p, this.RecvMsgHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_p2p_recv, this.RecvMsgHEHandler)

	area.Register_vnode_search(msg_id_vnode_search, this.SearchVnodeHandler)
	area.Register_vnode_p2pHE(msg_id_vnode_search_recv, this.SearchVnodeHandler_recv)
}

var MsgCountLock = new(sync.Mutex)
var MsgCount = make(map[string]uint64)

func (this *TestPeer) RecvP2PMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	selfId := message.Head.RecvId.B58String()
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
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
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
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
}

func (this *TestPeer) RecvMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// selfId := message.Head.RecvId.B58String()
	// utils.Log.Info().Msgf("收到vnode p2p消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
	err := this.area.SendVnodeP2pReplyMsgHE(message, msg_id_vnode_p2p_recv, nil)
	if err != nil {
		utils.Log.Info().Msgf("SendVnodeP2pReplyMsgHE error:%s", err.Error())
	} else {

	}
}

func (this *TestPeer) RecvMsgHEHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到vnode p2p消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
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
	// selfId := message.Head.RecvId.B58String()
	// MsgCountLock.Lock()
	// count, ok := MsgCount[selfId]
	// if ok {
	// 	MsgCount[selfId] = count + 1
	// } else {
	// 	MsgCount[selfId] = 1
	// }
	// MsgCountLock.Unlock()
}
