package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	mc "web3_gui/libp2parea/v1/message_center"
	"web3_gui/libp2parea/v1/virtual_node"
	"web3_gui/utils"
)

func main() {
	fmt.Println("start")
	engine.SetLogPath("log.txt")
	StartAllPeer()
}

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("localtest"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19960
)

/*
启动所有节点
*/
func StartAllPeer() {
	n := 2

	areaPeers := make([]*libp2parea.Area, 0, n)
	for i := 0; i < n; i++ {
		utils.Log.Info().Msgf("启动一个节点 start")
		area := StartOnePeer(i)
		areaPeers = append(areaPeers, area)
		utils.Log.Info().Msgf("启动一个节点 end")
	}

	utils.Log.Info().Msgf("等待各个节点自治完成...")

	//等待各个节点都准备好
	for _, one := range areaPeers {
		one.WaitAutonomyFinishVnode()
	}

	utils.Log.Info().Msgf("各个节点自治完成!!!")

	//异步发消息
	group := new(sync.WaitGroup)
	group.Add(len(areaPeers))
	// <-time.NewTimer(time.Second * 20).C
	for i, _ := range areaPeers {
		one := areaPeers[i]
		go sendMsgLoop(one, areaPeers, group)
	}
	group.Wait()

	//同步发消息
	// <-time.NewTimer(time.Second * 5).C
	// for i := 0; i < 100; i++ {
	// 	for _, one := range areaPeers {
	// 		for _, two := range areaPeers {
	// 			if bytes.Equal(one.GetNetId(), two.GetNetId()) {
	// 				continue
	// 			}
	// 			sendMsgOne(one, two.GetNetId())
	// 			time.Sleep(time.Second / 1000)
	// 		}
	// 	}
	// }

	for {
		time.Sleep(time.Second * 5)
		countNum := 1
		for k, v := range MsgCount {
			utils.Log.Info().Msgf("%d节点：%s 收到:%d", countNum, k, v)
			countNum++
		}
		utils.Log.Info().Msgf("-----------------------")
	}

	// select {}

}

func StartOnePeer(i int) *libp2parea.Area {
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
	area.SetNetTypeToTest()
	area.OpenVnode()
	area.SetLeveldbPath(filepath.Join("conf", "db"+strconv.Itoa(i)))

	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	err = area.StartUP(false, "127.0.0.1", uint16(basePort+i))
	if err != nil {
		utils.Log.Error().Msgf("%s", err.Error())
	}

	InitHandler(area)
	return area
}

func sendMsgLoop(area *libp2parea.Area, toAddrs []*libp2parea.Area, group *sync.WaitGroup) {
	for _, one := range toAddrs {
		if bytes.Equal(area.GetNetId(), one.GetNetId()) {
			continue
		}
		done := false
		for !done {
			err := sendMsgOne(area, one)
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

func sendMsgOne(area *libp2parea.Area, toArea *libp2parea.Area) error {
	// utils.Log.Info().Msgf("start sendMsg")

	toAddr := toArea.GetNetId()

	content := []byte("你好")
	_, _, _, err := area.SendP2pMsg(msg_id_p2pHE, &toAddr, &content)
	if err != nil {
		utils.Log.Error().Msgf("发送失败:%s", err.Error())
		return err
	}
	// utils.Log.Info().Msgf("发送消息%v %t %t", msg, sendOk, isSelf)

	utils.Log.Info().Msgf("发送加密消息:")
	_, _, _, err = area.SendP2pMsgHE(msg_id_p2pHE_recv, &toAddr, &content)
	if err != nil {
		utils.Log.Error().Msgf("加密消息发送失败:%s", err.Error())
		return err
	}

	fromAddrVnode := area.GetVnodeId()
	toAddrVnode := toArea.GetVnodeId()

	toAddrVnode = virtual_node.AddressNetExtend(append(toAddrVnode, []byte("1")...))
	utils.Log.Info().Msgf("要查找的vnodeID:%s", toAddrVnode.B58String())
	area.SendVnodeP2pMsgHE(msg_id_p2pHE, &fromAddrVnode, &toAddrVnode, &toAddr, nil)

	return nil
}

const msg_id_p2pHE = 1001
const msg_id_p2pHE_recv = 1002 //

func InitHandler(area *libp2parea.Area) {
	area.Register_p2pHE(msg_id_p2pHE, RecvMsgHandler)
	area.Register_p2pHE(msg_id_p2pHE_recv, RecvMsgHEHandler)
}

var MsgCountLock = new(sync.Mutex)
var MsgCount = make(map[string]uint64)

func RecvMsgHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	selfId := message.Head.RecvId.B58String()
	utils.Log.Info().Msgf("收到加密消息 from:%s say:%s", message.Head.Sender.B58String(), string(*message.Body.Content))
	MsgCountLock.Lock()
	count, ok := MsgCount[selfId]
	if ok {
		MsgCount[selfId] = count + 1
	} else {
		MsgCount[selfId] = 1
	}
	MsgCountLock.Unlock()

}

func RecvMsgHEHandler(c engine.Controller, msg engine.Packet, message *mc.Message) {
	utils.Log.Info().Msgf("收到加密消息 from:%s say:%s", message.Head.Sender.B58String(), string(*message.Body.Content))
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
