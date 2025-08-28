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
	areaName = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19300
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

	// for {
	// 	time.Sleep(time.Second * 10)
	// }

	toNetId := nodeStore.AddressFromB58String("2y3QG2boRAxfK2bwf18X3jxt6wwZSqqdUquDDmF4UXq7")
	for {
		sendMsgOne(areaPeers[0].area, &toNetId)
		time.Sleep(time.Second * 30)
	}

	// utils.Log.Info().Msgf("Game over !!!!")

	// select {}
}

func sendMsgOne(area *libp2parea.Area, toNetid *nodeStore.AddressNet) error {
	//发送p2p消息
	utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toNetid.B58String())
	senderProxy := nodeStore.AddressFromB58String("7gx5kDeZ9WVQNx8iTcxmfSKj37SQJVhM586EJdTV3Gsw")
	recvProxy := nodeStore.AddressFromB58String("8gAMb3sHx4qyGUWmkKbYt4DRmtYiPscReZHQrNfiViEU")
	_, _, _, err := area.SendP2pMsgProxyWaitRequest(msg_id_p2p, toNetid, &recvProxy, &senderProxy, nil, time.Second*10)
	if err != nil {
		utils.Log.Error().Msgf("发送P2p消息失败:%s", err.Error())
		return err
	}

	// recvProxy2 := nodeStore.AddressFromB58String("4bWPEi2AwpYty4DgYxEYCBM357pp2BHtUFNSXFr3Xy5r")
	// _, _, _, err = area.SendP2pMsgProxyWaitRequest(msg_id_p2p, toNetid, &recvProxy2, &senderProxy, nil, time.Second*10)
	// if err != nil {
	// 	utils.Log.Error().Msgf("发送P2p2消息失败:%s", err.Error())
	// 	return err
	// }

	//发送广播消息
	// utils.Log.Info().Msgf("节点:%s 发送消息给:%s", area.GetNetId().B58String(), toArea.GetNetId().B58String())
	err = area.SendMulticastMsg(msg_id_multicast, nil)
	if err != nil {
		utils.Log.Error().Msgf("发送广播消息失败:%s", err.Error())
	}

	return nil
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

	area.SetMachineID("area_machine_id_client3_" + strconv.Itoa(i))

	// area.OpenVnode()

	area.SetAreaGodAddr("127.0.0.1", 19962)
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
	utils.Log.Info().Msgf("[%s] 收到p2p消息 self:%s sender:%s msgHash:%s", this.area.GetMachineID(), this.area.GetNetId().B58String(), message.Head.Sender.B58String(), msgHash)
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
