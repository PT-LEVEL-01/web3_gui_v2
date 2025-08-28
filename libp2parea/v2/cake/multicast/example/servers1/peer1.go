package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/cake/country"
	"web3_gui/libp2parea/v2/cake/multicast"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	mc "web3_gui/libp2parea/v2/message_center"
	"web3_gui/libp2parea/v2/message_center/flood"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func main() {
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")

	StartAllPeer()
}

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("nihaoa a a proxy test!"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19960
)

var areaInfo = `{
	"address": {
		"47.109.16.70:19965": "GCS5wMJaBmMHCYZY6sY5E8bw7MGnHPFPfmJ9ojUkCU4s",
		"47.109.16.70:19967": "AWqtLYee34Y6XXZfs1Tr2aRp8KRh7FKxA4CTm1XUGA7G",
		"47.109.16.70:19968": "6ih9Yyg8xociXPdbBoFHaPscowFMktcCC8SJMuWp2qw8",
		"47.109.16.70:19969": "4nKnkyyM6dqu5u9sciuWSWn78Ah4LaCvJcAtz8q9zJjz",
		"47.109.16.70:19970": "CFKJRh64pD1nuEcQPfX1RRcLWVaHcJamK41izvDLL3e1"
	},
	"country": {
		"CN": "CN",
		"HK": "HK",
		"UK":"UK"
	},
	"nodes2": {
		"CN": {
			"FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n": "47.109.16.70:19965",
			"4MTCfHur4XPmRhpUByGC7Z9ve3pbezDyiYuYM6cA5Lmj": "47.109.16.70:19966",
			"6Wrco6JKC5c7pg8qp5Z5mNnayHqdW1eyLExHboihCY4c": "47.109.16.70:19967"
			
		},
		"HK": {
			"qUX2PsQATzSaofU6Q7sGmipu1g2V1H2KwgdzcqZsbN7": "47.109.16.70:19969",
			"GQHsVDddC44AM6z4LmS6f1brkJn1cNYhrms75ZEwchAG": "47.109.16.70:19970"
		},
		"UK":{
			"5o1oZJwByLAgT36BXXbBRDJ8UDhUUY4JLJbjR3aG6ZEp": "47.109.16.70:19968"
		},
		"default": {
			"FNHWnFV8puymF1xFLMuPxPDgDxPk4rSBaSuJ7HYpjG8n": "47.109.16.70:19965"
		}
	},
	"nodes": {
		"CN": "47.109.16.70:19965,47.109.16.70:19966",
		"MY": "47.109.16.70:19967,47.109.16.70:19968",
		"SG": "47.109.16.70:19969,47.109.16.70:19970",
		"default": "47.109.16.70:19965"
	},
	"version": 23
}`

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 1
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		ct := country.NewAreaCountry(area.area, i == 0)
		area.ct = ct
		areaPeers = append(areaPeers, area)
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

	// 启动大区, 初始化multicastmsg
	for i := range areaPeers {
		areaPeers[i].ct.Start()
		if i == 0 {
			areaPeers[i].ct.SetData(areaInfo)
		}
		gm := multicast.NewMulticastMsg(areaPeers[i].area, areaPeers[i].ct)
		areaPeers[i].gm = gm
	}

	sleepTime := time.Second * 30
	utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	time.Sleep(sleepTime)

	var nodeMachineInfoes []*multicast.NodeMachineInfo
	nodeMachineInfoes = append(nodeMachineInfoes, &multicast.NodeMachineInfo{NodeId: nodeStore.AddressFromB58String("3De2hUzoutTkRZTrqpU9NuMh5AthGuc6cwFEoEtiB7X8")})
	nodeMachineInfoes = append(nodeMachineInfoes, &multicast.NodeMachineInfo{NodeId: nodeStore.AddressFromB58String("BjS4ZvVL6c7oZNWybyDHxcji9mFRZbpBT3FDMAos8ZsK"), MachineID: "area_machine_id_client3_0"})

	cnt := 0
	for {
		time.Sleep(time.Second * 30)

		utils.Log.Info().Msgf("")
		utils.Log.Info().Msgf("--------------------------------------------")
		utils.Log.Info().Msgf("")

		cnt++
		content := fmt.Sprintf("[%d]I'm %s", cnt, areaPeers[0].area.GetNetId().B58String())
		areaPeers[0].gm.SendMulticastMsg(msg_id_p2p, nodeMachineInfoes, content, time.Second*10)
	}

	// select {}
}

type TestPeer struct {
	area *libp2parea.Area
	ct   *country.AreaCountry
	gm   *multicast.MulticastMsg
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

	area.SetMachineID("area_machine_id_1")

	// area.OpenVnode()

	area.SetDiscoverPeer(host + ":" + strconv.Itoa(19960))
	area.StartUP(false, host, uint16(basePort+i))

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
}

const msg_id_p2p = 1001
const msg_id_p2p_recv = 1002 //加密消息

const msg_id_search = 1003
const msg_id_search_recv = 1004

const msg_id_multicast = 1009 //

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_search, this.SearchSuperHandler)
	area.Register_p2p(msg_id_search_recv, this.SearchSuperHandler_recv)

	area.Register_multicast(msg_id_multicast, this.MulticastMsgHandler)

	area.Register_nodeClosedCallback(this.NodeCloseCallbackFunc)
	area.Register_nodeBeenGodAddrCallback(this.NodeBeenGodAddrCallbackFunc)
}

var MsgCountLock = new(sync.Mutex)
var MsgCount = make(map[string]uint64)

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
	utils.Log.Info().Msgf("[%s] 收到p2p search消息 self:%s sender:%s sMID:%s recvId:%s", this.area.GetMachineID(), this.area.GetNetId().B58String(), message.Head.Sender.B58String(), message.Head.SenderMachineID, message.Head.RecvId.B58String())

	err := this.area.SendSearchSuperReplyMsg(message, msg_id_search_recv, nil)
	if err != nil {
		utils.Log.Error().Msgf("SendSearchSuperReplyMsg error:%s", err.Error())
	}
}

func (this *TestPeer) SearchSuperHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
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
