package boot

import (
	"github.com/astaxie/beego"
	"web3_gui/chain/config"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/restart"
	"web3_gui/chain/rpc"
	"web3_gui/chain/rpc/limiter"
	"web3_gui/keystore/adapter"
	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/sdk/web"
	routers "web3_gui/libp2parea/adapter/sdk/web/routers"
	"web3_gui/utils"
)

/*
Start 区块链服务启动
*/
func Start(area *libp2parea.Area) {
	config.Area = area

	rpc.RegisterChainWitnessVoteRPC()
	rpc.RegisterErrorCode()
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
Start 区块链服务启动
*/
func StartWithParams(area *libp2parea.Area, configJSONStr string, init bool) error {
	config.Area = area
	config.InitNode = init
	err := config.ParseConfigWithConfigJSON(configJSONStr, "")
	if err != nil {
		engine.Log.Error(err.Error())
		return err
	}

	rpc.RegisterChainWitnessVoteRPC()
	rpc.RegisterErrorCode()
	//初始化rpc请求速率，交易速率
	config.RegisterRateHandle()
	limiter.InitRpcLimiter()

	evm.InitRewardContract()
	precompiled.InitRewardContract()
	//启动区块链模块
	err = RegisterWithParamsJSON(area, configJSONStr)
	if err != nil {
		engine.Log.Error(err.Error())
		return err
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
	return nil
}

/*
StartWithArea 区块链服务启动,包含了网络模块
*/
func StartWithArea(passwd string) {
	// utils.PprofMem(time.Minute * 10)

	engine.SetLogPath("logs/log.txt")
	config.Step()

	config.ParseInitFlag()
	//重启主链命令
	if config.ReStartNode {
		if err := restart.Restart(); err != nil {
			engine.Log.Error("主链重启失败:%v", err)
			//重启失败就退出
			panic(err)
		}
	}
	if passwd != "" {
		config.Wallet_keystore_default_pwd = passwd
	}

	key := keystore.NewKeystore(config.KeystoreFileAbsPath, config.AddrPre)
	err := key.Load()
	if err != nil {
		//没有就创建
		err = key.CreateNewKeystore(config.Wallet_keystore_default_pwd)
		if err != nil {
			panic("创建key错误:" + err.Error())
		}
	}
	if key.NetAddr == nil {
		_, _, err := key.CreateNetAddr(config.Wallet_keystore_default_pwd, config.Wallet_keystore_default_pwd)
		if err != nil {
			panic("创建NetAddr错误:" + err.Error())
		}
	}
	if len(key.GetAddr()) == 0 {
		_, err = key.GetNewAddr(config.Wallet_keystore_default_pwd, config.Wallet_keystore_default_pwd)
		if err != nil {
			panic("创建Addr错误:" + err.Error())
		}
	}
	if len(key.GetDHKeyPair().SubKey) == 0 {
		_, err = key.GetNewDHKey(config.Wallet_keystore_default_pwd, config.Wallet_keystore_default_pwd)
		if err != nil {
			panic("创建DHKey错误:" + err.Error())
		}
	}

	// engine.Log.Info("启动的area name:%s", string(config.AreaName[:]))
	area, ERR := libp2parea.NewArea(config.AreaName, config.AddrPre, key, config.Wallet_keystore_default_pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	if config.NetType != config.NetType_release {
		area.SetNetTypeToTest()
	}

	ERR = area.StartUP(config.Init_LocalPort)
	if ERR.CheckFail() {
		panic(ERR.String())
	}

	//启动 RPC 模块
	// NOTE 链同步完成之间启动RPC模块
	//启动web模块
	go func() {
		mining.ChainInitialized()
		routers.RegisterRpc()
		if !config.EnableStartupWeb {
			web.Start(config.SetLibp2pareaConfig())
			go beego.Run()
		}
	}()

	//启动区块链模块
	Start(area)

	//启动grpc服务
	//rpcServer := rpc_server.NewRpcServer()
	//rpcServer.Start()

	<-utils.GetStopService()
}
