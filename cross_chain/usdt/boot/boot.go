package boot

import (
	"web3_gui/cross_chain/usdt/config"
	"web3_gui/cross_chain/usdt/db"
	"web3_gui/cross_chain/usdt/tron/full_node"
	"web3_gui/cross_chain/usdt/wallet"
	"web3_gui/keystore/v2"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func Start() {
	//配置文件，没有就创建
	ERR := CheckCreateConfigJSON()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	//连接数据库
	_, ERR = db.ConnLevelDB()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	//初始化钱包
	ERR = InitWallet()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	//加载钱包地址
	ERR = wallet.LoadWalletAddress()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	//启动扫链程序->波场
	ERR = full_node.ScanTron()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	//启动rpc接口
	ERR = StartServer()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
}

/*
不存在则创建config.json文件
*/
func CheckCreateConfigJSON() utils.ERROR {
	exist, err := utils.PathExists(config.PATH_config_json)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if exist {
		return utils.NewErrorSuccess()
	}
	return config.SaveConfig()
}

/*
初始化钱包
*/
func InitWallet() utils.ERROR {
	wallet, ERR := keystore.LoadOrSaveWallet(config.PATH_Wallet, config.WALLET_address_pre, config.WALLET_password)
	if ERR.CheckFail() {
		utils.Log.Info().Str("加载和初始化钱包错误", ERR.String()).Send()
		return ERR
	}
	keyst, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return ERR
	}
	config.Wallet_keystore = keyst
	return utils.NewErrorSuccess()
}

/*
启动http服务器
*/
func StartServer() utils.ERROR {
	addrInfo, ERR := config.Wallet_keystore.GetNetAddrInfo(config.WALLET_password)
	if ERR.CheckFail() {
		return ERR
	}
	sessionEngine := engine.NewEngine(addrInfo.GetAddrStr())
	ERR = RegisterRPC(sessionEngine)
	if ERR.CheckFail() {
		return ERR
	}
	return sessionEngine.ListenOnePort(config.HTTP_PORT, false)
}
