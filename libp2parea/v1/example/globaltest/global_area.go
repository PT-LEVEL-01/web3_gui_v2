package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/global"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/nodeStore"
)

func main() {
	fmt.Println("start")
	engine.SetLogPath("log.txt")
	StartAllPeer()
}

var (
	addrPre   = "SELF"
	areaName  = sha256.Sum256([]byte("nihaoa a a!"))
	areaName2 = sha256.Sum256([]byte("nihaoa a b!"))
	areaName3 = sha256.Sum256([]byte("nihaoa a c!"))
	keyPwd    = "123456789"
	host      = "127.0.0.1"
	basePort  = 19600
)

/*
启动所有节点
*/
func StartAllPeer() {
	n := 1

	areas := make([]*libp2parea.Area, 0, n)
	areaPeers := make([]*TestPeer, 0, n)
	for i := 0; i < n; i++ {
		area1, area2, area3 := StartThreePeer(i)
		areas = append(areas, area1.area)
		areas = append(areas, area2.area)
		areas = append(areas, area3.area)
		areaPeers = append(areaPeers, area1)
		areaPeers = append(areaPeers, area2)
		areaPeers = append(areaPeers, area3)
	}

	{
		sleepTime := time.Second * 10
		utils.Log.Info().Msgf("等待%s后打印关系", sleepTime)
		time.Sleep(sleepTime)

		for _, one := range areaPeers {
			allNodes := one.area.NodeManager.GetAllNodes()
			utils.Log.Info().Msgf("[%s] cur:%s logic nodes-------------begin", hex.EncodeToString(one.area.AreaName[:]), one.area.GetNetId().B58String())
			for i := range allNodes {
				utils.Log.Info().Msgf("logic node id:%s", allNodes[i].IdInfo.Id.B58String())
			}
			utils.Log.Info().Msgf("[%s] cur:%s logic nodes-------------end", hex.EncodeToString(one.area.AreaName[:]), one.area.GetNetId().B58String())
			utils.Log.Info().Msgf("")
		}
		utils.Log.Info().Msgf("")
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

	for _, one := range areaPeers {
		allNodes := one.area.NodeManager.GetAllNodes()
		utils.Log.Info().Msgf("cur:%s logic nodes-------------begin", hex.EncodeToString(one.area.AreaName[:]))
		for i := range allNodes {
			utils.Log.Info().Msgf("logic node id:%s", allNodes[i].IdInfo.Id.B58String())
		}
		utils.Log.Info().Msgf("cur:%s logic nodes-------------end", hex.EncodeToString(one.area.AreaName[:]))
		utils.Log.Info().Msgf("")
	}
	utils.Log.Info().Msgf("")

	//异步发消息
	group := new(sync.WaitGroup)
	group.Add(len(areaPeers))
	<-time.NewTimer(time.Second * 20).C
	for i := range areaPeers {
		one := areaPeers[i]
		go sendMsgLoop(one.area, areaPeers, group)
	}
	group.Wait()

	dstPort := 0
	for {
		dstPort++
		if dstPort >= 3 {
			dstPort = 0
		}
		utils.Log.Warn().Msgf("开始切换代理服务器")
		areaPeers[0].area.SetAreaGodAddr(host, dstPort+19100)

		time.Sleep(time.Second * 30)
	}
}

func StartThreePeer(i int) (peer1, peer2, peer3 *TestPeer) {
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

	global := global.NewGlobal(key1, keyPwd)

	area, err := libp2parea.NewAreaByGlobal(areaName, key1, keyPwd, global)
	if err != nil {
		panic(err.Error())
	}
	// area.SetNetTypeToTest()
	area.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	// area.SetDiscoverPeer(host + ":19100")
	area.SetAreaGodAddr(host, 19100)
	area.SetPhoneNode()
	utils.Log.Error().Msgf("节点1 areaName为: %s", hex.EncodeToString(areaName[:]))
	area.StartUPGlobal(false, global, host, uint16(basePort+i))
	peer1 = &TestPeer{
		area: area,
	}
	peer1.InitHandler(area)

	area2, err := libp2parea.NewAreaByGlobal(areaName2, key1, keyPwd, global)
	if err != nil {
		panic(err.Error())
	}
	// area.SetNetTypeToTest()
	area2.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	// area2.SetDiscoverPeer(host + ":19200")
	area2.SetAreaGodAddr(host, 19200)
	area2.SetPhoneNode()
	utils.Log.Error().Msgf("节点2 areaName为: %s", hex.EncodeToString(areaName2[:]))
	area2.StartUPGlobal(false, global, host, uint16(basePort+i))
	peer2 = &TestPeer{
		area: area2,
	}
	peer2.InitHandler(area2)

	area3, err := libp2parea.NewAreaByGlobal(areaName3, key1, keyPwd, global)
	if err != nil {
		panic(err.Error())
	}
	// area.SetNetTypeToTest()
	area3.SetLeveldbPath(config.Path_leveldb + strconv.Itoa(i))
	// area3.SetDiscoverPeer(host + ":19300")
	area3.SetAreaGodAddr(host, 19300)
	area3.SetPhoneNode()
	utils.Log.Error().Msgf("节点3 areaName为: %s", hex.EncodeToString(areaName3[:]))
	area3.StartUPGlobal(false, global, host, uint16(basePort+i))
	peer3 = &TestPeer{
		area: area3,
	}
	peer3.InitHandler(area3)

	return
}

var sendIndex int
var sendLock sync.Mutex

func sendMsgOne(area *libp2parea.Area, toAddr nodeStore.AddressNet) error {
	// utils.Log.Info().Msgf("start sendMsg")
	sendLock.Lock()
	sendIndex++
	msg := fmt.Sprintf("[%s] cur:%s send msg to %s index:%d", hex.EncodeToString(area.AreaName[:]), area.NodeManager.NodeSelf.IdInfo.Id.B58String(), toAddr.B58String(), sendIndex)
	sendLock.Unlock()
	content := []byte(msg)
	_, _, _, err := area.SendP2pMsg(msg_id_text, &toAddr, &content)
	if err != nil {
		utils.Log.Error().Msgf("发送失败:%s", err.Error())
		return err
	} else {
		utils.Log.Warn().Msgf(msg)
	}
	// utils.Log.Info().Msgf("发送消息%v %t %t", msg, sendOk, isSelf)

	// _, _, _, err = area.SendP2pMsgHE(msg_id_text_he, &toAddr, &content)
	// if err != nil {
	// 	utils.Log.Error().Msgf("加密消息发送失败:%s", err.Error())
	// 	return err
	// }

	return nil
}

func sendMsgLoop(area *libp2parea.Area, toAddrs []*TestPeer, group *sync.WaitGroup) {
	for _, one := range toAddrs {
		if bytes.Equal(area.GetNetId(), one.area.GetNetId()) {
			continue
		}
		done := false
		for !done {
			err := sendMsgOne(area, one.area.GetNetId())
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
}

const msg_id_text = 1001
const msg_id_text_recv = 1002 //回执消息

const msg_id_multicast = 1009 //

func (this *TestPeer) InitHandler(area *libp2parea.Area) {
	area.Register_p2p(msg_id_text, this.RecvMsgHandler)

	area.Register_multicast(msg_id_multicast, this.MulticastMsgHandler)
}

type TestPeer struct {
	area *libp2parea.Area
}

func (this *TestPeer) RecvMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	utils.Log.Info().Msgf("[%s] cur:%s from:%s msgHash:%s say:------>>>>%s", hex.EncodeToString(this.area.AreaName[:]), this.area.NodeManager.NodeSelf.IdInfo.Id.B58String(), message.Head.Sender.B58String(), hex.EncodeToString(message.Body.Hash), string(*message.Body.Content))

	if !bytes.Equal(this.area.GetNetId(), *message.Head.RecvId) {
		utils.Log.Error().Msgf("收到错误p2p消息self:%s recv:%s msgHash:%s", this.area.GetNetId().B58String(), message.Head.RecvId.B58String(), hex.EncodeToString(message.Body.Hash))
		return
	}

	content := []byte(fmt.Sprintf("[%s] I catch your msg: cur:%s", hex.EncodeToString(this.area.AreaName[:]), this.area.GetNetId().B58String()))
	this.area.SendP2pReplyMsg(message, msg_id_text_recv, &content)

	// 发送p2p消息
	utils.Log.Info().Msgf("[%s] 节点:%s 发送消息给:%s", hex.EncodeToString(this.area.AreaName[:]), this.area.GetNetId().B58String(), message.Head.Sender.B58String())
	sendValue := fmt.Sprintf("[%s] %s send to %s", hex.EncodeToString(this.area.AreaName[:]), this.area.GetNetId().B58String(), message.Head.Sender.B58String())
	content = []byte(sendValue)
	this.area.SendP2pMsg(msg_id_text, message.Head.Sender, &content)
}

func (this *TestPeer) MulticastMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	utils.Log.Info().Msgf("[%s] 收到广播消息: self:%s content:%s", hex.EncodeToString(this.area.AreaName[:]), this.area.GetNetId().B58String(), string(*message.Body.Content))
}
