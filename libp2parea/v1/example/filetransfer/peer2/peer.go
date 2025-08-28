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

	n := 1
	areas := make([]*libp2parea.Area, 0, n)
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area := StartOnePeer(i + 1)
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

	utils.Log.Info().Msgf("qlw-----sleep 10 seconds")
	time.Sleep(time.Second * 10)

	//for {
	showLogicNode(areas[0])
	//utils.Log.Info().Msgf("")
	//utils.Log.Info().Msgf("每10秒钟打印一次logic nodes info___")
	//time.Sleep(10 * time.Second)
	//utils.Log.Info().Msgf("")
	//}

	select {}
}

func showLogicNode(area *libp2parea.Area) {
	logicNodeAddrs := area.NodeManager.GetLogicNodes()
	for i := range logicNodeAddrs {
		utils.Log.Info().Msgf("logicNodeInfoes cur:%s, logic:%s", area.NodeManager.NodeSelf.IdInfo.Id.B58String(), logicNodeAddrs[i].B58String())

		//path, _ := os.Getwd()
		fObj := NewFileInfo("./files/xmind.rar")
		err := fObj.FindFile(area, logicNodeAddrs[i])
		if err != nil {
			utils.Log.Info().Msgf("查找接收方的文件时出错误:%v", err.Error())
			break
		}

		ok := fObj.SendFile(area, logicNodeAddrs[i])
		utils.Log.Info().Msgf("最后报告, ok:%v", ok)
		//panic(ok)
	}
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
		_, _, err := key1.CreateNetAddr(keyPwd, keyPwd)
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
			panic("创建DHKey错误:" + err.Error())
		}
	}

	area, err := libp2parea.NewArea(areaName, key1, keyPwd)
	if err != nil {
		panic(err.Error())
	}
	area.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	area.SetNetTypeToTest()

	//area.OpenVnode()

	//serverHost
	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(false, host, uint16(basePort+i))

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

	return nil
}

const msg_id_p2p = 1001
const msg_id_p2p_recv = 1002 //加密消息

const msg_id_searchSuper = 1003
const msg_id_searchSuper_recv = 1004 //加密消息

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_p2p, this.FileMsg)
	area.Register_p2p(msg_id_p2p_recv, this.RecvP2PMsgHandler_recv)

	area.Register_p2p(msg_id_searchSuper, this.FindFileMsg)
	area.Register_p2p(msg_id_searchSuper_recv, this.SearchSuperHandler_recv)
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
