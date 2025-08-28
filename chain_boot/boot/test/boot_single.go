package test

import (
	"path/filepath"
	"web3_gui/chain/boot/startup_web"
	"web3_gui/chain/config"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/restart"
	"web3_gui/chain/rpc/limiter"
	"web3_gui/chain_boot/boot"
	"web3_gui/chain_boot/boot/pprof"
	"web3_gui/chain_boot/rpc"
	gconfig "web3_gui/config"
	"web3_gui/keystore/adapter"
	ks2 "web3_gui/keystore/v2"
	"web3_gui/keystore/v2/updatev2"
	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
单独运行一个区块链节点实例
*/
func StartChain() utils.ERROR {
	//InitConfig()
	err := boot.SaveConfigFile(filepath.Join(config.Path_configDir, config.Path_config))
	if err != nil {
		panic(err)
	}
	configJSONStr, err := gconfig.LoadConfigJSONString(gconfig.FILEPATH_config)
	if err != nil {
		panic("加载config.json文件错误:" + err.Error())
	}
	err = gconfig.ParseConfig(configJSONStr, "")
	if err != nil {
		panic("解析config.json文件错误:" + err.Error())
	}
	
	err = utils.LogBuildDefaultFile("logs/log.txt")
	if err != nil {
		panic(err)
	}
	//utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()

	// NOTE 监控Cpu/Mem
	pprof.StartPprof()
	config.ParseConfig()
	//utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()
	if !config.EnableStartupWeb && config.EnableStartInputPassword {
		pwd, err := boot.ParseWalletPassword(Path_keystore_old, config.KeystoreFileAbsPath, config.AddrPre)
		if err != nil {
			utils.Log.Error().Err(err).Send()
			//fmt.Println(err)
			return utils.NewErrorSysSelf(err)
		}
		config.Wallet_keystore_default_pwd = pwd
	}

	//utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()
	// utils.PprofMem(time.Minute * 10)

	//engine.SetLogPath("logs/log.txt")
	//utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()
	config.Step()

	//utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()
	config.ParseInitFlag()
	//utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()
	//重启主链命令
	if config.ReStartNode {
		if err := restart.Restart(); err != nil {
			engine.Log.Error("主链重启失败:%v", err)
			//重启失败就退出
			panic(err)
		}
	}
	//utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()

	//startup web
	if config.EnableStartupWeb {
		utils.Log.Info().Bool("11111111", config.EnableStartupWeb).Send()
		startup_web.Start("")
	}
	//utils.Log.Info().Str("11111111", "").Send()

	//if passwd != "" {
	//	config.Wallet_keystore_default_pwd = passwd
	//}

	//升级keystore
	ERR := updatev2.UpdateV2(config.AddrPre, Path_keystore_old, config.KeystoreFileAbsPath, config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		panic("升级密钥库错误:" + ERR.String())
	}
	wallet, ERR := ks2.LoadOrSaveWallet(config.KeystoreFileAbsPath, config.AddrPre, config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	keyItr, ERR := keystore.NewKeystoreItr(wallet, config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	//utils.Log.Info().Interface("keyItr ")
	// engine.Log.Info("启动的area name:%s", string(config.AreaName[:]))
	area, ERR := libp2parea.NewArea(config.AreaName, config.AddrPre, keyItr, config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		return ERR
	}
	node := area.GetNode()
	ERR = node.SetDiscoverPeer(gconfig.DiscoverPeerAddr)
	if ERR.CheckFail() {
		return ERR
	}
	ERR = area.StartUP(config.Init_LocalPort)
	if ERR.CheckFail() {
		return ERR
	}

	rpc.RegisterRPC(area.GetNode())

	//启动web模块
	go func() {
		mining.ChainInitialized()
		//beego.Run()
	}()

	//启动区块链模块
	Start(area)

	//启动grpc服务
	//rpcServer := rpc_server.NewRpcServer()
	//rpcServer.Start()

	<-utils.GetStopService()
	return utils.NewErrorSuccess()
}

/*
区块链服务启动
*/
func Start(area *libp2parea.Area) {
	config.Area = area

	//rpc.RegisterChainWitnessVoteRPC()
	//rpc.RegisterErrorCode()
	//初始化rpc请求速率，交易速率
	config.RegisterRateHandle()
	limiter.InitRpcLimiter()

	evm.InitRewardContract()
	precompiled.InitRewardContract()
	//启动区块链模块
	err := Register(area)
	if err != nil {
		engine.Log.Error(err.Error())
	}
	//全节点启动成功，广播地址给轻节点
	//node := area.GetNetId()
	//bytes := []byte(node)
	//area.SendMulticastMsg(config.FUll_NODE_INLINE_MUTIL, &bytes)
	// rpcServer := rpc_server.NewRpcServer()
	// rpcServer.Start()
	//启动区块链 RPC 模块
	// NOTE 链同步完成之后启动RPC模块
	// routers.RegisterRpc()
}

/*
通过命令行启动，要解析命令行传入的参数
*/
func Register(area *libp2parea.Area) error {
	// 非创世节点等待网络自治完成
	config.ParseInitFlag()
	//Area = area
	return register(area)
}
