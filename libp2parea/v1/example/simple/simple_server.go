package main

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/utils"
)

var (
	addrPre    = "SELF"
	areaName   = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd     = "123456789"
	serverHost = "3.8.4.211"
	clientHost = "127.0.0.1"
	basePort   = 19960
)

func main() {
	fmt.Println("start")
	engine.SetLogPath("simple_server_log.txt")
	StartOnePeer()

	select {}
}

func StartOnePeer() *libp2parea.Area {
	keyPath1 := filepath.Join("conf", "keystore_server.key")
	dbpath := filepath.Join("db", "msgcache_server")

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
	area.SetNetTypeToTest()
	area.SetLeveldbPath(dbpath)

	// area.SetDiscoverPeer(serverHost+":"+strconv.Itoa(19981), "7.6.6.1:19981")
	area.StartUP(false, clientHost, uint16(basePort))

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
