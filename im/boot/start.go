package boot

import (
	"path/filepath"
	"web3_gui/chain_boot/boot"
	"web3_gui/config"
	"web3_gui/im/im"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/v2"
	"web3_gui/storage"
	"web3_gui/utils"
)

/*
启动rpc模块
*/
func StartRPCModule() {
	engine.LogClose()
	err := boot.SaveConfigFile(filepath.Join(config.Path_configDir, config.Path_config))
	if err != nil {
		panic("保存config.json文件错误:" + err.Error())
	}
	configJSONStr, err := config.LoadConfigJSONString(config.FILEPATH_config)
	if err != nil {
		panic("加载config.json文件错误:" + err.Error())
	}
	err = config.ParseConfig(configJSONStr, "")
	if err != nil {
		panic("解析config.json文件错误:" + err.Error())
	}
	//engine.SetLogPath("logs/log.txt")
	utils.LogBuildDefaultFile("logs/log.txt")
	//web.Start(config.ConfigJSONStr)
	//routers.RegisterRpc()
	//启动web模块
	//go beego.Run()

	//im.RegisterBootRPC()

	<-utils.GetStopService()
}

/*
启动IM模块
@passwd    string    种子密码
@params    string    配置中是否有init参数，有则用初始节点方式启动
*/
func StartUP(passwd, params string) utils.ERROR {
	configJSONStr, err := config.LoadConfigJSONString(config.FILEPATH_config)
	if err != nil {
		return utils.NewErrorSysSelf(err)
		//return errors.New("加载config.json文件错误:" + err.Error())
	}
	err = config.ParseConfig(configJSONStr, params)
	if err != nil {
		return utils.NewErrorSysSelf(err)
		//panic("解析config.json文件错误:" + err.Error())
	}
	utils.LogBuildDefaultFile("logs/log.txt")
	//engine.SetLogPath("logs/log.txt")
	wallet, ERR := keystore.LoadOrSaveWallet(config.KeystoreFileAbsPath, config.AddrPre, passwd)
	if ERR.CheckFail() {
		return ERR
	}
	keyst, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return ERR
	}
	node, ERR := libp2parea.NewNode(config.AreaName, config.AddrPre, keyst, passwd)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = node.StartUP(config.Init_LocalPort)
	if ERR.CheckFail() {
		return ERR
	}

	//area, err := libp2parea.NewArea(config.AreaName, key, passwd)
	//if err != nil {
	//	return errors.New("启动Area错误:" + err.Error())
	//}
	//if config.NetType != config.NetType_release {
	//	area.SetNetTypeToTest()
	//}
	// area.SetDiscoverPeer(host + ":" + strconv.Itoa(basePort))
	// utils.Log.Info().Msgf("端口:%d", config.Init_LocalPort)

	//area.StartUP(config.ParseHaveInitFlag(), config.Init_LocalIP, config.Init_LocalPort)
	//web.Start(config.ConfigJSONStr)
	//routers.RegisterRpc()
	//启动web模块
	//go beego.Run()
	StartModuls(node, configJSONStr, params)
	<-utils.GetStopService()
	return utils.NewErrorSuccess()
}

// 启动其他模块
func StartModuls(area *libp2parea.Node, configJSONStr, params string) {
	im.StartIM(area, configJSONStr, params, nil)
	im.StartFileTransfer(area, configJSONStr)
	storage.StartStorageServer(area)
}
