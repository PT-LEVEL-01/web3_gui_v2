package main

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"time"

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
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")

	StartAllPeer()
}

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("qlw tps waiwang test!"))
	keyPwd   = "123456789"
	// host     = "127.0.0.1"
	// host     = "172.28.0.141"
	host     = "8.212.2.156"
	basePort = 19960
	resChan  = make(chan bool, 100)
)

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 6
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

	for i := range areas {
		showLogicNodes(areas[i])
	}

	utils.Log.Info().Msgf("Game over------------")

	utils.Log.Info().Msgf("")

	for _, area := range areas {
		utils.Log.Info().Msgf("%s --> %d", area.NodeManager.NodeSelf.IdInfo.Id.B58String(), mapCount[area.NodeManager.NodeSelf.IdInfo.Id.B58String()])
	}

	utils.Log.Info().Msgf("------------------------------------")

	select {}
}

var mapCount = make(map[string]int)

func showLogicNodes(area *libp2parea.Area) {
	utils.Log.Info().Msgf("--------------")
	utils.Log.Info().Msgf("节点:%s", area.NodeManager.NodeSelf.IdInfo.Id.B58String())
	for _, one := range area.NodeManager.GetLogicNodes() {
		utils.Log.Info().Msgf("--逻辑节点:%s", one.B58String())
		if _, exist := mapCount[one.B58String()]; exist {
			mapCount[one.B58String()]++
		} else {
			mapCount[one.B58String()] = 1
		}
	}
	utils.Log.Info().Msgf("--------------")
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

	// area.SetAreaGodAddr(host, basePort)
	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))
	// area.CloseVnode()
	// area.Vm.AddVnodeByIndex(100)
	// area.Vm.AddVnodeByIndex(200)

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
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

const msg_id_p2p_with_res = 1010
const msg_id_p2p_with_res_recv = 1011 //加密消息

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

	area.Register_p2p(msg_id_p2p_with_res, this.RecvP2PMsgWithResHandler)
	area.Register_p2p(msg_id_p2p_with_res_recv, this.RecvP2PMsgWithResHandler_recv)
}

func (this *TestPeer) RecvP2PMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	if !bytes.Equal(this.area.GetNetId(), *message.Head.RecvId) {
		utils.Log.Error().Msgf("收到错误p2p消息self:%s recv:%s", this.area.GetNetId().B58String(), message.Head.RecvId.B58String())
		return
	}
	this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, nil)
}

func (this *TestPeer) RecvP2PMsgHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func (this *TestPeer) SearchSuperHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到SearchSuper消息 from:%s self:%s", message.Head.Sender.B58String(), this.area.GetNetId().B58String())
	err := this.area.SendSearchSuperReplyMsg(message, msg_id_searchSuper_recv, nil)
	if err != nil {
		utils.Log.Error().Msgf("SendSearchSuperReplyMsg error:%s", err.Error())
	}
}

func (this *TestPeer) SearchSuperHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到SearchSuper消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

var nIndex int32

func (this *TestPeer) RecvMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	if this.area.Vm.FindInVnodeSelf(*message.Head.RecvVnode) {
		utils.Log.Info().Msgf("收到vnode p2p消息self:%s from:%s to:%s", this.area.GetNetId().B58String(),
			message.Head.Sender.B58String(), message.Head.RecvVnode.B58String())
	} else {
		utils.Log.Error().Msgf("收到错误vnode p2p消息self:%s from:%s to:%s", this.area.GetNetId().B58String(),
			message.Head.Sender.B58String(), message.Head.RecvVnode.B58String())
		return
	}
	content := []byte(message.Head.RecvVnode.B58String() + " -> " + message.Head.SenderVnode.B58String())
	err := this.area.SendVnodeP2pReplyMsgHE(message, msg_id_vnode_p2p_recv, &content)
	if err != nil {
		utils.Log.Error().Msgf("SendVnodeP2pReplyMsgHE error:%s", err.Error())
	}
}

func (this *TestPeer) RecvMsgHEHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	contentValue := message.Body.Content
	utils.Log.Warn().Msgf("收到vnode p2p消息返回 cur:%s from:%s deal:%s content:%s", this.area.NodeManager.NodeSelf.IdInfo.Id.B58String(), message.Head.Sender.B58String(), message.Head.SenderVnode.B58String(), contentValue)
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func (this *TestPeer) SearchVnodeHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到 VnodeSearch 消息 self:%s from:%s", this.area.GetNetId().B58String(), message.Head.Sender.B58String())
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

func (this *TestPeer) RecvP2PMsgWithResHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	if !bytes.Equal(this.area.GetNetId(), *message.Head.RecvId) {
		utils.Log.Error().Msgf("收到错误p2p消息self:%s recv:%s", this.area.GetNetId().B58String(), message.Head.RecvId.B58String())
		return
	}
	this.area.SendP2pReplyMsg(message, msg_id_p2p_with_res_recv, nil)
}

func (this *TestPeer) RecvP2PMsgWithResHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// utils.Log.Info().Msgf("收到P2P消息返回 from:%s", message.Head.Sender.B58String())
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
	resChan <- true
}
