package keystore

import (
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

const (
	addrPre      = "TEST"
	pwd          = "123456789"
	fristAddrPwd = "123456789"
)

var (
	walletPath = filepath.Join("D:/test/temp/keystore", "wallet"+strconv.Itoa(0)+".key")
)

func TestWallet(t *testing.T) {
	wallet, ERR := LoadOrSaveWallet(walletPath, "TEST", pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	keys, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	_, ERR = keys.CreateCoinAddr("", pwd, pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	_, ERR = keys.CreateAddressCoinType("", pwd, pwd, config.COINTYPE_BTC)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	world, ERR := keys.ExportMnemonic(pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	utils.Log.Info().Str("导出助记词", world).Send()

	for i, one := range keys.GetCoinAddrAll() {
		utils.Log.Info().Str("地址"+strconv.Itoa(i), one.GetAddrStr()).Send()
	}
}

func hdwallet_example() {

}

func TestWallet_wallet_example(t *testing.T) {
	//指定钱包保存路径，以及收款地址前缀
	w := NewWallet(walletPath, addrPre)
	//加载钱包文件
	ERR := w.Load()
	if ERR.CheckFail() {
		//文件不存在
		if ERR.Code == config.ERROR_code_wallet_incomplete {
			utils.Log.Error().Str("钱包文件损坏", walletPath).Send()
			return
		} else if ERR.Code == config.ERROR_code_wallet_file_not_exist {
			utils.Log.Error().Str("钱包文件不存在", walletPath).Send()
		} else {
			utils.Log.Error().Str("加载钱包文件报错", ERR.String()).Send()
			return
		}
	}
	list := w.List()
	if len(list) == 0 {
		ERR = w.AddKeystoreRand(pwd)
		if ERR.CheckFail() {
			utils.Log.Error().Str("添加一个随机数钱包报错", ERR.String()).Send()
			return
		}
		list = w.List()
	}
	_, ERR = w.Use(list[0].Index, pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("使用密钥库错误", ERR.String()).Send()
		return
	}

}

func TestWallet_createOtherAddr(t *testing.T) {
	pwd := "123456789"
	wallet, ERR := LoadOrSaveWallet(walletPath, "", pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	ks, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		panic(ERR.String())
	}

	utils.Log.Info().Int("其他收款地址数量", len(ks.AddrsOther)).Send()

	addr, ERR := ks.CreateAddressCoinType("", pwd, pwd, config.COINTYPE_BTC)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	fmt.Println(addr.GetAddrStr())

}

func TestWalletCreateAddress100000(t *testing.T) {
	wallet, ERR := LoadOrSaveWallet(walletPath, "TEST", pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	keys, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	start := time.Now()
	for range 10 {
		_, ERR = keys.CreateAddressCoinType("", pwd, pwd, config.COINTYPE_BTC)
		if ERR.CheckFail() {
			panic(ERR.String())
		}
	}
	utils.Log.Info().Str("花费时间", time.Now().Sub(start).String()).Send()
	start = time.Now()
	world, ERR := keys.ExportMnemonic(pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	utils.Log.Info().Str("导出助记词 花费时间", time.Now().Sub(start).String()).Send()
	utils.Log.Info().Str("导出助记词", world).Send()

}

func TestWalletUpdateNickname(t *testing.T) {

	words := "salt slush prepare manual labor van snack person achieve recycle clever join"
	wallet := NewWallet(walletPath, addrPre)
	ERR := wallet.ImportMnemonic(words, pwd, pwd, pwd, pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	keyst, ERR := wallet.GetKeystoreUse()
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	for i, one := range keyst.GetCoinAddrAll() {
		utils.Log.Info().Str("address"+strconv.Itoa(i), one.GetAddrStr()).Str("nickname", one.Nickname).Send()
	}
	ERR = keyst.UpdateCoinAddrNickname(coin_address.AddressCoin(keyst.GetCoinAddrAll()[0].Addr.Bytes()), pwd, "hello")
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	for i, one := range keyst.GetCoinAddrAll() {
		utils.Log.Info().Str("address"+strconv.Itoa(i), one.GetAddrStr()).Str("nickname", one.Nickname).Send()
	}
}
