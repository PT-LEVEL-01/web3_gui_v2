package main

import (
	"crypto/sha256"
	"path/filepath"
	"strconv"

	"github.com/astaxie/beego"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/cake/update_version"
	"web3_gui/libp2parea/v1/config"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/sdk/web"
	"web3_gui/utils"
)

var (
	addrPre  = "SELF"
	areaName = sha256.Sum256([]byte("nihaoa a a!"))
	keyPwd   = "123456789"
	host     = "127.0.0.1"
	basePort = 19960
)

func main() {
	StartUP(keyPwd)
}
func StartUP(passwd string) {
	engine.SetLogPath("logs/log.txt")

	keyPath1 := filepath.Join("conf", "keystore.key")
	//	config.Step()
	key := keystore.NewKeystore(keyPath1, addrPre)
	err := key.Load()
	if err != nil {
		//没有就创建
		err = key.CreateNewKeystore(passwd)
		if err != nil {
			panic("创建key错误:" + err.Error())
		}
	}

	if key.NetAddr == nil {
		_, _, err := key.CreateNetAddr(passwd, passwd)
		if err != nil {
			panic("创建NetAddr错误:" + err.Error())
		}
	}

	if len(key.GetAddr()) < 1 {
		_, err = key.GetNewAddr(passwd, passwd)
		if err != nil {
			panic("创建Addr错误:" + err.Error())
		}
	}
	if len(key.GetDHKeyPair().SubKey) < 1 {
		_, err = key.GetNewDHKey(passwd, passwd)
		if err != nil {
			panic("创建DHKey错误:" + err.Error())
		}
	}

	area, err := libp2parea.NewArea(areaName, key, passwd)
	if err != nil {
		panic(err.Error())
	}

	area.SetLeveldbPath(config.Path_leveldb)
	area.SetNetTypeToTest()

	area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	area.StartUP(true, host, uint16(basePort))

	web.Start("{}")

	//启动web模块
	go beego.Run()

	// NOTE: 在此添加不同的模块

	uv := update_version.NewUpdateVersion(area, nodeStore.AddressFromB58String("47yKUvBfKBQ9g6L9wPwAZPVMkD2X1JF7DM7Y4d2Frw2q"), "win", 1001, 1002, 1003, 1004, 1005, 1006)
	uv.TickerUpdateVersion()
	update_version.Start(uv)

	<-utils.GetStopService()
}
