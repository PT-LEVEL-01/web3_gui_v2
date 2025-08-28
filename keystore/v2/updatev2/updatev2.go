package updatev2

import (
	"sync"
	"web3_gui/keystore/v1"
	ks2 "web3_gui/keystore/v2"
	"web3_gui/utils"
)

var lock = new(sync.Mutex)

func UpdateV2(addrPre, filePathOld, filePathNew, pwd string) utils.ERROR {
	lock.Lock()
	defer lock.Unlock()
	//检查v2路径文件已经存在，则退出
	exist, err := utils.PathExists(filePathNew)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if exist {
		return utils.NewErrorSuccess()
	}
	//加载旧版key文件
	keysOld := keystore.NewKeystore(filePathOld, addrPre)
	err = keysOld.Load()
	if err != nil {
		//没有就返回
		return utils.NewErrorSuccess()
	}
	//导出助记词
	mnemonic, err := keysOld.ExportMnemonic(pwd)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}

	//打印之前钱包地址
	for _, one := range keysOld.GetAddr() {
		utils.Log.Info().Str("原有收款地址", one.GetAddrStr()).Send()
	}

	//fmt.Println("导出助记词", mnemonic)

	//导入新版钱包助记词
	wallet := ks2.NewWallet(filePathNew, addrPre)
	ERR := wallet.ImportMnemonic(mnemonic, pwd, pwd, pwd, pwd)
	if ERR.CheckFail() {
		return ERR
	}

	keys, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		return ERR
	}
	//恢复相同数量收款地址
	for range len(keysOld.GetAddr()) - len(keys.GetCoinAddrAll()) {
		_, ERR := keys.CreateCoinAddr("", pwd, pwd)
		if ERR.CheckFail() {
			return ERR
		}
	}

	//打印新版钱包地址
	for _, one := range keys.GetCoinAddrAll() {
		utils.Log.Info().Str("新版收款地址", one.GetAddrStr()).Send()
	}

	//转换成旧地址
	for i, _ := range keys.GetCoinAddrAll() {
		ERR = keys.ConvertOldVersionAddress(i, pwd)
		if ERR.CheckFail() {
			return ERR
		}
	}

	//打印新版钱包地址
	for _, one := range keys.GetCoinAddrAll() {
		utils.Log.Info().Str("还原成旧版收款地址", one.GetAddrStr()).Send()
	}
	return utils.NewErrorSuccess()
}
