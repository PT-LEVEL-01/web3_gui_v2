package main

import (
	"crypto/sha256"
	"path/filepath"
	"strconv"
	"time"

	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
)

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("nihaoa a a!"))
	// areaName   = sha256.Sum256([]byte("icom"))
	keyPwd     = "123456789"
	serverHost = "124.221.170.43"
	clientHost = "127.0.0.1"
	basePort   = 19960

	server_test_reaName    = sha256.Sum256([]byte("icom_im_test"))
	serverAddr_test        = "8.216.128.167:19975"
	server_release_reaName = sha256.Sum256([]byte("icom"))
	serverAddr_release     = "cdnode.91qpb.cn:19965"
)

func main() {
	utils.Log.Info().Msgf("start")
	engine.SetLogPath("log.txt")
	area := StartOnePeer(0)

	{
		// go func() {
		// 	for {
		// 		time.Sleep(time.Second * 2)
		// 		logicNodes := area.GetNodeManager().GetLogicNodes()
		// 		for _, one := range logicNodes {
		// 			utils.Log.Info().Msgf("连接数量:%d %s", len(logicNodes), hex.EncodeToString(one))
		// 		}
		// 	}
		// }()

		// time.Sleep(time.Second * 5)

		// area.CloseNet()
	}

	utils.Log.Info().Msgf("StartOnePeer 5555555555")
	area.WaitAutonomyFinish()

	utils.Log.Info().Msgf("StartOnePeer 66666666666")
	utils.Log.Info().Msgf("end")

	for i := 0; i < 300; i++ {
		utils.Log.Info().Msgf("连接数量:%d", len(area.GetNodeManager().GetLogicNodes()))
		time.Sleep(time.Second * 1)
	}
	// time.Sleep(time.Second * 10)
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

	// utils.Log.Info().Msgf("StartOnePeer 1111111111")
	area, err := libp2parea.NewArea(server_test_reaName, key1, keyPwd)
	// utils.Log.Info().Msgf("StartOnePeer 2222222222")
	area.CloseVnode()
	// utils.Log.Info().Msgf("StartOnePeer 333333333")
	// area.SetNetTypeToTest()
	area.SetPhoneNode()

	// area.SetDiscoverPeer(serverHost + ":" + strconv.Itoa(basePort))
	area.SetDiscoverPeer(serverAddr_test)
	area.StartUP(false, clientHost, uint16(basePort+i))
	// utils.Log.Info().Msgf("StartOnePeer 444444444444")
	// InitHandler(area)
	return area
}

const msg_id_text = 1000
const msg_id_searchSuper = 1001

func sendMsg(area *libp2parea.Area, toAddr nodeStore.AddressNet) {
	utils.Log.Info().Msgf("start sendMsg")
	content := []byte("你好")

	msg, sendOk, isSelf, err := area.SendP2pMsg(msg_id_text, &toAddr, &content)
	if err != nil {
		utils.Log.Info().Msgf("发送失败:%s", err.Error())
		return
	}
	utils.Log.Info().Msgf("发送消息%v %t %t", msg, sendOk, isSelf)
	netid := area.GetNetId()
	addrs := nodeStore.GetQuarterLogicAddrNetByAddrNet(&netid)
	for i, v := range addrs {
		if i == 0 {
			continue
		}
		_, err := area.SendSearchSuperMsg(msg_id_searchSuper, v, &content)
		if err != nil {
			utils.Log.Info().Msgf("发送错误:%s", err.Error())
		} else {
			utils.Log.Info().Msgf("发送search super msg成功")
		}
	}
	return

}
