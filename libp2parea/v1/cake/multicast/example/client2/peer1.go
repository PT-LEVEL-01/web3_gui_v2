package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
	areaName = sha256.Sum256([]byte("nihaoa a a proxy test!"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19200
)

/*
启动所有节点
*/
func StartAllPeer() {
	nsm := nodeStore.NodeSimulationManager{IDdepth: 32 * 8}

	n := 1
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i)
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

	sleepTime := time.Second * 30
	utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
	time.Sleep(sleepTime)

	// area := areaPeers[0].area

	// go func() {
	// 	areaAddr := area.GetNetId()
	// 	for {
	// 		utils.Log.Warn().Msgf("开始查询地址代理信息!!!!!!")
	// 		proxyInfoes, err := area.GetAddrProxy(&areaAddr, "")
	// 		if err == nil && proxyInfoes != nil && len(proxyInfoes) != 0 {
	// 			for i := range proxyInfoes {
	// 				utils.Log.Error().Msgf("[节点%s] %d 查询到的代理信息 代理:%s mid:%s version:%d", areaAddr.B58String(), i, proxyInfoes[i].ProxyId.B58String(), proxyInfoes[i].MachineId, proxyInfoes[i].Version)
	// 			}
	// 		}

	// 		time.Sleep(time.Second * 20)
	// 	}
	// }()

	// dstPort := 0
	// for {
	// 	time.Sleep(time.Second * 20)

	// 	dstPort++
	// 	if dstPort >= 6 {
	// 		dstPort = 0
	// 	}
	// 	utils.Log.Warn().Msgf("开始切换代理服务器")
	// 	area.SetAreaGodAddr(host, 19960+dstPort)
	// }

	select {}
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
	area.SetPhoneNode()

	area.SetMachineID("area_machine_id_client2_" + strconv.Itoa(i))

	// area.OpenVnode()

	area.SetAreaGodAddr("127.0.0.1", 19960)
	// area.SetDiscoverPeer("127.0.0.1:19960")
	area.StartUP(false, host, uint16(basePort+i))

	peer := TestPeer{
		area: area,
	}

	peer.InitHandler(area)
	return &peer
}

const msg_id_p2p = 1001
const msg_id_p2p_recv = 1002 //加密消息

const msg_id_multicast = 1009 //

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_multicast(msg_id_multicast, this.MulticastMsgHandler)
}

func (this *TestPeer) RecvP2PMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	if !bytes.Equal(this.area.GetNetId(), *message.Head.RecvId) {
		utils.Log.Error().Msgf("[%s] 收到错误p2p消息self:%s recv:%s", this.area.GetMachineID(), this.area.GetNetId().B58String(), message.Head.RecvId.B58String())
		return
	}

	var msgHash string
	if message.Body != nil && message.Body.Hash != nil {
		msgHash = hex.EncodeToString(message.Body.Hash)
	}
	var content string
	if message.Body != nil && message.Body.Content != nil {
		content = string(*message.Body.Content)
	}
	utils.Log.Info().Msgf("[%s] 收到p2p消息 self:%s sender:%s sMID:%s content:%s msgHash:%s", this.area.GetMachineID(), this.area.GetNetId().B58String(), message.Head.Sender.B58String(), message.Head.SenderMachineID, content, msgHash)
	// utils.Log.Info().Msgf("[%s] 收到p2p消息 self:%s sender:%s msgHash:%s", this.area.GetMachineID(), this.area.GetNetId().B58String(), message.Head.Sender.B58String(), msgHash)
	this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, nil)
}

func (this *TestPeer) RecvP2PMsgHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	var msgHash string
	if message.Body != nil && message.Body.Hash != nil {
		msgHash = hex.EncodeToString(message.Body.Hash)
	}
	utils.Log.Info().Msgf("[%s] 收到P2P消息返回 from:%s msgHash:%s", this.area.GetMachineID(), message.Head.Sender.B58String(), msgHash)
	flood.ResponseBytes(utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

func (this *TestPeer) MulticastMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	utils.Log.Info().Msgf("[%s] 收到广播消息:%s", this.area.GetMachineID(), this.area.GetNetId().B58String())
}
