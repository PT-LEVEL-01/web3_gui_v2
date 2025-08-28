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

	area := areaPeers[0].area

	recvAddr := nodeStore.AddressFromB58String("3De2hUzoutTkRZTrqpU9NuMh5AthGuc6cwFEoEtiB7X8")

	// SendSearchSuperMsgAuto 测试
	// for {
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("---------------------------------------")
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("发送p2p search消息")
	// 	msg, err := area.SendSearchSuperMsgAuto(msg_id_search, &recvAddr, nil)
	// 	if err != nil {
	// 		utils.Log.Error().Msgf("SendSearchSuperMsgAuto err:%s", err)
	// 	}

	// 	strTag := string(msg.Body.Hash)
	// 	flood.RegisterRequest(strTag)
	// 	_, err = flood.WaitResponse(strTag, time.Second*10)
	// 	if err != nil {
	// 		utils.Log.Error().Msgf("flood.WaitResponse err:%s", err)
	// 	} else {
	// 		utils.Log.Warn().Msgf("flood.WaitResponse success")
	// 	}

	// 	time.Sleep(time.Second * 15)
	// }

	// SendSearchSuperMsgAutoWaitRequest 测试
	// for {
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("---------------------------------------")
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("发送p2p search消息")
	// 	_, err := area.SendSearchSuperMsgAutoWaitRequest(msg_id_search, &recvAddr, nil, time.Second*10)
	// 	if err != nil {
	// 		utils.Log.Error().Msgf("SendSearchSuperMsgAutoWaitRequest err:%s", err)
	// 	}

	// 	utils.Log.Warn().Msgf("flood.WaitResponse success")

	// 	time.Sleep(time.Second * 15)
	// }

	// SendP2pMsgAuto 测试
	// for {
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("---------------------------------------")
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("发送p2p消息")
	// 	res := area.SendP2pMsgAuto(msg_id_p2p, &recvAddr, "area_machine_id_client2_0", nil, false)
	// 	for i := range res {
	// 		if res[i].Err != nil {
	// 			utils.Log.Error().Msgf("SendP2pMsgAuto err:%s", res[i].Err)
	// 			continue
	// 		}

	// 		var recvId, recvProxy, recvMID, sendProxy, msgHash string
	// 		if res[i].Msg.Head.RecvId != nil {
	// 			recvId = res[i].Msg.Head.RecvId.B58String()
	// 		}
	// 		if res[i].Msg.Head.RecvProxyId != nil {
	// 			recvProxy = res[i].Msg.Head.RecvProxyId.B58String()
	// 		}
	// 		if res[i].Msg.Head.RecvMachineID != "" {
	// 			recvMID = res[i].Msg.Head.RecvMachineID
	// 		}
	// 		if res[i].Msg.Head.SenderProxyId != nil {
	// 			sendProxy = res[i].Msg.Head.SenderProxyId.B58String()
	// 		}
	// 		if res[i].Msg.Body != nil && res[i].Msg.Body.Hash != nil {
	// 			msgHash = hex.EncodeToString(res[i].Msg.Body.Hash)
	// 		}
	// 		utils.Log.Warn().Msgf("发送数据成功, recvId:%s, recvProxy:%s, recvMID:%s, sendProxy:%s, msgHash:%s", recvId, recvProxy, recvMID, sendProxy, msgHash)
	// 	}

	// 	time.Sleep(time.Second * 15)
	// }

	// SendP2pMsgAuto 测试
	// for {
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("---------------------------------------")
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("发送p2p消息")
	// 	res := area.SendP2pMsgAutoWaitRequest(msg_id_p2p, &recvAddr, "area_machine_id_client2_0", nil, time.Second*10, false)
	// 	for i := range res {
	// 		if res[i].Err != nil {
	// 			utils.Log.Error().Msgf("SendP2pMsgAutoWaitRequest err:%s", res[i].Err)
	// 			continue
	// 		}

	// 		utils.Log.Warn().Msgf("发送数据成功!!!!!!")
	// 	}

	// 	time.Sleep(time.Second * 15)
	// }

	// SendP2pMsgAuto 测试
	// for {
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("---------------------------------------")
	// 	utils.Log.Info().Msgf("")
	// 	utils.Log.Info().Msgf("发送p2p消息")
	// 	res := area.SendP2pMsgHEAuto(msg_id_p2p, &recvAddr, "", nil, false)
	// 	for i := range res {
	// 		if res[i].Err != nil {
	// 			utils.Log.Error().Msgf("SendP2pMsgHEAuto err:%s", res[i].Err)
	// 			continue
	// 		}

	// 		utils.Log.Warn().Msgf("发送数据成功!!!!!!")
	// 	}

	// 	time.Sleep(time.Second * 15)
	// }

	// SendP2pMsgHEAutoWaitRequest 测试
	for {
		utils.Log.Info().Msgf("")
		utils.Log.Info().Msgf("---------------------------------------")
		utils.Log.Info().Msgf("")
		utils.Log.Info().Msgf("发送p2p消息")

		// 检查节点是否在线
		isOnline := area.CheckProxyClientIsOnline(&recvAddr, "area_machine_id_client1_0")
		utils.Log.Error().Msgf("client:%s isOnline:%v", recvAddr.B58String(), isOnline)

		time.Sleep(time.Second * 2)

		utils.Log.Info().Msgf("")

		// 发送消息
		res := area.SendP2pMsgHEAutoWaitRequest(msg_id_p2p, &recvAddr, "", nil, time.Second*10, false)
		for i := range res {
			if res[i].Err != nil {
				utils.Log.Error().Msgf("SendP2pMsgHEAuto err:%s mid:%s", res[i].Err, res[i].MachineID)
				continue
			}

			utils.Log.Warn().Msgf("发送数据成功!!!!!! mid:%s", res[i].MachineID)
		}

		time.Sleep(time.Second * 15)
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
	area.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	area.SetNetTypeToTest()
	area.SetPhoneNode()

	area.SetMachineID("area_machine_id_client3_" + strconv.Itoa(i))

	// area.OpenVnode()

	area.SetAreaGodAddr("127.0.0.1", 19964)
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

const msg_id_search = 1003
const msg_id_search_recv = 1004

const msg_id_multicast = 1009 //

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.RecvP2PMsgHandler)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_search, this.SearchSuperHandler)
	area.Register_p2p(msg_id_search_recv, this.SearchSuperHandler_recv)

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
	utils.Log.Info().Msgf("[%s] 收到p2p消息 self:%s sender:%s sMID:%s msgHash:%s", this.area.GetMachineID(), this.area.GetNetId().B58String(), message.Head.Sender.B58String(), message.Head.SenderMachineID, msgHash)
	this.area.SendP2pReplyMsg(message, msg_id_p2p_recv, nil)
}

func (this *TestPeer) RecvP2PMsgHandler_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	var msgHash string
	if message.Body != nil && message.Body.Hash != nil {
		msgHash = hex.EncodeToString(message.Body.Hash)
	}
	var senderProxy string
	if message.Head.SenderProxyId != nil {
		senderProxy = message.Head.SenderProxyId.B58String()
	}
	utils.Log.Info().Msgf("[%s] 收到P2P消息返回 from:%s sMID:%s senderProxy:%s msgHash:%s", this.area.GetMachineID(), message.Head.Sender.B58String(), message.Head.SenderMachineID, senderProxy, msgHash)
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
	utils.Log.Info().Msgf("[%s] 收到广播消息:%s", this.area.GetMachineID(), this.area.GetNetId().B58String())
}
