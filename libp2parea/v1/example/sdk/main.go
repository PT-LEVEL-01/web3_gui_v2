package main

import (
	"github.com/astaxie/beego"
	"web3_gui/keystore/v1"
	"web3_gui/libp2parea/v1"
	"web3_gui/libp2parea/v1/engine"
	"web3_gui/libp2parea/v1/example/sdk/module_a/config"
	"web3_gui/libp2parea/v1/example/sdk/module_area_rpc"
	"web3_gui/libp2parea/v1/sdk/web"
	"web3_gui/utils"
)

func main() {
	StartUP("123456789")
}

func StartUP(passwd string) {
	engine.SetLogPath("logs/log.txt")
	config.Step()

	// if passwd != "" {
	// 	config.Wallet_keystore_default_pwd = passwd
	// }

	key := keystore.NewKeystore(config.KeystoreFileAbsPath, config.AddrPre)
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

	area, err := libp2parea.NewArea(config.AreaName, key, passwd)
	if err != nil {
		panic(err.Error())
	}
	if config.NetType != config.NetType_release {
		area.SetNetTypeToTest()
	}
	area.SetMachineID(config.MachineId)

	area.StartUP(config.Init, config.Init_LocalIP, config.Init_LocalPort)

	web.Start(config.SetLibp2pareaConfig())

	//启动web模块
	go beego.Run()

	// NOTE: 在此添加不同的模块
	module_area_rpc.Start(area)

	<-utils.GetStopService()
}
