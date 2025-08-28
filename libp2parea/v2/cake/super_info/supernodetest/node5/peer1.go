package main

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	superinfo "web3_gui/libp2parea/v2/cake/super_info"
	"web3_gui/libp2parea/v2/config"
	"web3_gui/libp2parea/v2/engine"
	mc "web3_gui/libp2parea/v2/message_center"
	"web3_gui/libp2parea/v2/message_center/flood"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

func main() {
	utils.Log.Info().Msgf("start")

	StartAllPeer()
}

var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("nihaoa a ip addr test!"))
	keyPwd     = "123456789"
	host       = "127.0.0.1"
	basePort   = 35000
	serverHost = "127.0.0.1"
	serverPort = 31000
)

/*
启动所有节点
*/
func StartAllPeer() {
	n := 1
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
		si := superinfo.NewSuperInfo(area.area, area.SuperNodeOnlineCallbackFunc, area.SuperNodeOfflineCallbackFunc)
		area.superInfo = si
		areaPeers = append(areaPeers, area)
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("开始等待节点自治")
	//等待各个节点都准备好
	for _, one := range areaPeers {
		one.area.WaitAutonomyFinish()
	}
	utils.Log.Info().Msgf("--------------------------")
	utils.Log.Info().Msgf("节点自治完成，打印逻辑节点")

	sleepTime := time.Second * 30
	utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	time.Sleep(sleepTime)

	// 定时显示所有连接信息
	go func() {
		area := areaPeers[0].area
		si := areaPeers[0].superInfo
		for {
			utils.Log.Info().Msgf("")
			utils.Log.Info().Msgf("-------------------------------------------------")
			valueCnt := 0
			si.SuperNodes.Range(func(key, value any) bool {
				valueCnt++
				node := value.(*nodeStore.Node)
				utils.Log.Error().Msgf("[%s] ip:%sport%d", node.IdInfo.Id.B58String(), node.Addr, node.TcpPort)
				return true
			})
			if valueCnt < 5 {
				allNodes := area.NodeManager.GetAllNodes()
				utils.Log.Info().Msgf("数据不足, 查看详情!!!!!!!!!!!")
				for i := range allNodes {
					utils.Log.Error().Msgf("节点 [%s] ip:%sport%d", allNodes[i].IdInfo.Id.B58String(), allNodes[i].Addr, allNodes[i].TcpPort)
				}
				utils.Log.Info().Msgf("查看所有的session信息!!!!!!!")
				allSession := area.SessionEngine.GetAllSession(area.AreaName[:])
				for i := range allSession {
					utils.Log.Error().Msgf("session [%s] ip:%s", nodeStore.AddressNet(allSession[i].GetName()).B58String(), allSession[i].GetRemoteHost())
				}
			}
			utils.Log.Info().Msgf("-------------------------------------------------")
			utils.Log.Info().Msgf("")

			time.Sleep(time.Second * 50)
		}
	}()

	select {}
}

type TestPeer struct {
	area      *libp2parea.Area
	superInfo *superinfo.SuperInfo
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

	// area.OpenVnode()

	//serverHost
	area.SetDiscoverPeer(serverHost + ":" + strconv.Itoa(serverPort))
	area.StartUP(false, host, uint16(basePort+i))

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

const msg_id_multicast = 1009 //

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_searchSuper, this.SearchSuperHandler)
	area.Register_p2p(msg_id_searchSuper_recv, this.SearchSuperHandler_recv)

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

func (this *TestPeer) MulticastMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	utils.Log.Info().Msgf("收到广播消息:%s", this.area.GetNetId().B58String())
}

func (this *TestPeer) SuperNodeOnlineCallbackFunc(addr nodeStore.AddressNet, machineID string) {
	utils.Log.Info().Msgf("qlw----- 超级节点上线回调 cur:%s, node:%s, machineID:%s", this.area.GetNetId().B58String(), addr.B58String(), machineID)
}

func (this *TestPeer) SuperNodeOfflineCallbackFunc(addr nodeStore.AddressNet, machineID string) {
	utils.Log.Error().Msgf("111111111111111111111111111111111111")
	utils.Log.Error().Msgf("222222222222222222222222222222222222")
	utils.Log.Error().Msgf("333333333333333333333333333333333333")
	utils.Log.Error().Msgf("444444444444444444444444444444444444")
	utils.Log.Error().Msgf("qlw----- 超级节点下线回调 cur:%s, node:%s, machineID:%s", this.area.GetNetId().B58String(), addr.B58String(), machineID)
}
