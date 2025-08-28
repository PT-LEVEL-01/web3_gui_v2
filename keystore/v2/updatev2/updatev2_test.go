package updatev2

import (
	"fmt"
	"testing"
	"web3_gui/keystore/v1"
	ks2 "web3_gui/keystore/v2"
)

func TestUpdate(t *testing.T) {
	addrPre := "TEST"
	ks2.SetCoinTypeCustom(123456789, addrPre) //CoinType_coin = 123456789
	//ks2.CoinType_dh = ks2.CoinType_coin + 1
	ks2.SetCoinTypeCustomDH(123456789 + 1)
	oldFilePath := "D:\\test\\temp/keystore.key"
	newFilePath := "D:\\test\\temp/wallet.bin"
	pwd := "123456789"
	ERR := UpdateV2(addrPre, oldFilePath, newFilePath, pwd)
	fmt.Println("结果", ERR.String())

	//打印钱包之前的收款地址
	//加载旧版key文件
	keysOld := keystore.NewKeystore(oldFilePath, addrPre)
	err := keysOld.Load()
	if err != nil {
		//没有就返回
		fmt.Println("错误", err.Error())
		panic(err)
	}
	for i, one := range keysOld.GetAddr() {
		fmt.Println("旧版地址", i, one.GetAddrStr())
	}

	wallet := ks2.NewWallet(newFilePath, addrPre)
	ERR = wallet.Load()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	keys, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	for i, one := range keys.GetCoinAddrAll() {
		fmt.Println("新版地址", i, one.GetAddrStr())
	}

	for i, _ := range keys.GetCoinAddrAll() {
		ERR = keys.ConvertOldVersionAddress(i, pwd)
		if ERR.CheckFail() {
			panic(ERR.String())
		}
	}

	for i, one := range keys.GetCoinAddrAll() {
		fmt.Println("更改后地址", i, one.GetAddrStr())
	}

	for i, _ := range keys.GetCoinAddrAll() {
		ERR = keys.ConvertNewVersionAddress(i, pwd)
		if ERR.CheckFail() {
			panic(ERR.String())
		}
	}

	for i, one := range keys.GetCoinAddrAll() {
		fmt.Println("还原为新版地址", i, one.GetAddrStr())
	}

	//升级的旧版不能添加地址
	_, ERR = keys.CreateCoinAddr("", pwd, pwd)
	if ERR.CheckSuccess() {
		panic(ERR.String())
	}
}
