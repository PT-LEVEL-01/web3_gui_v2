package im

import (
	"context"
	"os"
	"runtime"
	"time"
	chainconfig "web3_gui/chain/config"
	"web3_gui/chain_boot/boot/test"
	"web3_gui/chain_orders"
	"web3_gui/config"
	"web3_gui/file_transfer"
	"web3_gui/gui/tray"
	"web3_gui/im/db"
	"web3_gui/im/model"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2"
	"web3_gui/libp2parea/v2/cake/transfer_manager"
	"web3_gui/libp2parea/v2/cake/update_version"
	"web3_gui/storage"
	"web3_gui/utils"
)

var Node *libp2parea.Node

func StartUP(passwd string, init bool, ctx context.Context) utils.ERROR {

	if init {
		config.InitNode = true
		config.ParamJSONstr = `{"init":true}`
	} else {
		init = config.ParseHaveInitFlag()
	}

	//utils.Log.Info().Msgf("地址前缀:%s", config.AddrPre)
	utils.Log.Info().Msgf("key文件路径:%s", config.KeystoreFileAbsPath)

	wallet, ERR := keystore.LoadOrSaveWallet(config.KeystoreFileAbsPath, config.AddrPre, passwd)
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Msgf("地址前缀:%s", wallet.GetAddrPre())
	keyst, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return ERR
	}

	//utils.Log.Info().Msgf("地址前缀:%s", keyst.GetAddrPre())

	node, ERR := libp2parea.NewNode(config.AreaName, config.AddrPre, keyst, passwd)
	if ERR.CheckFail() {
		return ERR
	}

	ERR = node.SetDiscoverPeer(config.DiscoverPeerAddr)
	if ERR.CheckFail() {
		return ERR
	}

	//utils.Log.Info().Uint16("打印端口", config.Init_LocalPort).Send()
	ERR = node.StartUP(config.Init_LocalPort)
	if ERR.CheckFail() {
		return ERR
	}

	config.Wallet_keystore_default_pwd = passwd
	chainconfig.Wallet_keystore_default_pwd = passwd

	//utils.Log.Info().Msgf("55555555555555555555")
	if !init {
		node.WaitAutonomyFinish()
	}
	//utils.Log.Info().Msgf("666666666666666666666666")
	ERR = StartModuls(node, passwd, config.ConfigJSONStr, ctx)
	//utils.Log.Info().Msgf("77777777777777777")
	return ERR
}

// 启动其他模块
func StartModuls(area *libp2parea.Node, password, configJSONStr string, ctx context.Context) utils.ERROR {
	config.Node = area
	//utils.Log.Info().Msgf("启动参数:%s", configJSONStr)
	//启动文件传输模块
	StartFileTransfer(area, configJSONStr)
	//utils.Log.Info().Msgf("11111111111111")
	//启动im模块
	ERR := StartIM(area, configJSONStr, "", ctx)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("11111111111111")
	//启动版本升级管理模块
	StartVersionUpdate(area)
	//utils.Log.Info().Msgf("11111111111111")
	//启动云存储服务
	storage.StartStorageServer(area)
	//utils.Log.Info().Msgf("11111111111111")
	//启动区块链节点
	go test.StartChainWithNode(area, password, configJSONStr, config.InitNode)
	go TestOther()
	//启动订单模块
	chain_orders.Start()
	//utils.Log.Info().Msgf("11111111111111")
	return utils.NewErrorSuccess()
}

// 启动IM模块
func StartIM(area *libp2parea.Node, configJSONStr, params string, ctx context.Context) utils.ERROR {
	//utils.Log.Info().Msgf("11111111111111")
	err := config.ParseConfig(configJSONStr, params)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Msgf("11111111111111")
	Node = area
	config.NetAddr = area.GetNetId().B58String()
	config.LevelDB = area.GetLevelDB()
	//初始化rpc接口
	//RegisterRPC()
	//utils.Log.Info().Msgf("11111111111111")
	//注册p2p消息
	RegisterHandlers()
	//utils.Log.Info().Msgf("11111111111111")
	//第一次启动，初始化个人信息
	ERR := InitUserInfoSelf(ctx)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//utils.Log.Info().Msgf("11111111111111")
	//启动兴趣圈广播发送和接收系统
	InitCircleClassMultcast()
	//utils.Log.Info().Msgf("11111111111111")
	//启动IM代理模块
	ERR = StartIMProxy(area, configJSONStr, params, ctx)
	//utils.Log.Info().Msgf("11111111111111")
	MulticastOnline()
	return ERR
}

// 启动IM代理模块
func StartIMProxy(area *libp2parea.Node, configJSONStr, params string, ctx context.Context) utils.ERROR {
	defer utils.PrintPanicStack(nil)
	//utils.Log.Info().Msgf("11111111111111")
	//初始化代理节点信息
	ERR := InitImProxySetup()
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Msgf("11111111111111")
	//启动定时广播
	InitImProxyMultcast()
	//utils.Log.Info().Msgf("11111111111111")
	StaticProxyClientManager, ERR = CreateProxyClientManager(area.GetLevelDB())
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("创建CreateProxyClientManager 错误:%s", ERR.String())
		return ERR
	}
	//utils.Log.Info().Msgf("11111111111111")
	StaticProxyServerManager, ERR = CreateProxyServerManager(area.GetLevelDB())
	if !ERR.CheckSuccess() {
		utils.Log.Info().Msgf("创建CreateProxyServerManager 错误:%s", ERR.String())
		return ERR
	}
	StaticProxyServerManager.groupKnitManager = StaticProxyClientManager.groupKnitManager
	//utils.Log.Info().Msgf("11111111111111")
	return utils.NewErrorSuccess()
}

/*
启动文件传输模块
*/
func StartFileTransfer(area *libp2parea.Node, configJSONStr string) {
	file_transfer.ManagerStatic = file_transfer.NewManager(area)
	//云存储使用的传输单元
	tf := file_transfer.ManagerStatic.CreateClass(config.FILE_TRANSFER_CLASS_storage)
	tf.SetAutoReceive(true)
	tf.SetReceiveFilePath(config.FILE_TRANSFER_storage_download_path)
	//im聊天主动发送文件的传输单元
	tf = file_transfer.ManagerStatic.CreateClass(config.FILE_TRANSFER_CLASS_im_send_file)
	tf.SetAutoReceive(false)

	//初始化
	transferManger := transfer_manager.NewTransferManger(area, config.MSGID_file_transfer_11, config.MSGID_file_transfer_12,
		config.MSGID_file_transfer_13, config.MSGID_file_transfer_14, config.MSGID_file_transfer_15, config.MSGID_file_transfer_16) //启动文件传输模块
	transferManger.Load()
	transfer_manager.Start(transferManger)
	//设置自动拉取文件
	transfer_manager.TransferMangerStatic.PullTaskIsAutoSet(true)

	//
	go transferManger.BuildFileHash()
}

/*
设置下载文件块时，临时存放文件夹路径
*/
func SetFileTransferTempDownloadDir(dirPath string) {
	tf := file_transfer.ManagerStatic.GetClass(config.FILE_TRANSFER_CLASS_storage)
	if tf == nil {
		return
	}
	tf.SetReceiveFilePath(dirPath)
}

var UpdateVersionModule *update_version.UpdateVersion

/*
启动版本更新模块
*/
func StartVersionUpdate(area *libp2parea.Node) {
	update_version.P2p_mgs_timeout = time.Second * 10
	update_version.Lenth = 1024 * 10
	update_version.Update_version_expiration_interval = time.Second * 30 //time.Hour * 7
	UpdateVersionModule = update_version.NewUpdateVersion(area, config.VersionUpdateAddress, runtime.GOOS,
		config.MSGID_version_update_1, config.MSGID_version_update_2,
		config.MSGID_version_update_3, config.MSGID_version_update_4,
		config.MSGID_version_update_5, config.MSGID_version_update_6)
	UpdateVersionModule.SetCallback(func(newFilePath string) {
		bs, err := os.ReadFile(newFilePath)
		if err != nil {
			utils.Log.Info().Msgf("读新版本文件错误:%s", err.Error())
			return
		}
		err = utils.SaveFile(string(os.Args[0]), &bs)
		if err != nil {
			utils.Log.Info().Msgf("读新版本文件错误:%s", err.Error())
			return
		}
	})
	UpdateVersionModule.SetPlatform(runtime.GOOS)
	UpdateVersionModule.TickerUpdateVersion()
	update_version.Start(UpdateVersionModule)
}

/*
启动判断IM个人信息是否存在
*/
func InitUserInfoSelf(ctx context.Context) utils.ERROR {
	//utils.Log.Info().Msgf("初始化个人信息")
	utils.Log.Info().Msgf("初始化个人信息:%+v", Node.GetNetId())
	//个人信息，没有就创建
	userinfo, ERR := db.GetSelfInfo(*Node.GetNetId())
	//其他错误
	if !ERR.CheckSuccess() {
		//utils.Log.Info().Msgf("初始化个人信息")
		return ERR
	}
	//utils.Log.Info().Msgf("初始化个人信息")
	//个人信息不存在
	if userinfo == nil {
		//utils.Log.Info().Msgf("初始化个人信息")
		//utils.Log.Info().Msgf("检查个人信息 11111111111")
		userinfo = model.NewUserInfo(*Node.GetNetId())
		userinfo.Nickname = utils.BuildName() // randomname.GenerateName()
		//utils.Log.Info().Msgf("更新个人信息:%+v", userinfo)
		ERR := db.UpdateSelfInfo(*Node.GetNetId(), userinfo)
		if !ERR.CheckSuccess() {
			return ERR
		}
		//utils.Log.Error().Msgf("getuserinfo error:%s", err.Error())
	}
	//utils.Log.Info().Msgf("检查个人信息 11111111111")
	//tray.InitTray(ctx)
	if userinfo.Tray {
		//utils.Log.Info().Msgf("初始化个人信息")
		tray.OpenSystemTray()
		//go StartTray(ctx)
	}
	//SetUserSelf_cache(userinfo)
	//获取一遍好友列表
	GetFriendListAPI()
	return utils.NewErrorSuccess()
}

/*
测试其他内容
*/
func TestOther() {
	//t := go_protos.Test{Size_: 1}
	//bs, _ := t.Marshal()
	//utils.Log.Info().Msgf("测试打印内容：%v", bs)
}

/*
启动托盘
*/
//func StartTray(ctx context.Context) {
//	//time.Sleep(time.Second * 30)
//	tray.Start(ctx)
//}
