package keystore

import (
	"encoding/hex"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"path/filepath"
	"strconv"
	"testing"
	"time"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

var (
	keystorePath = filepath.Join("D:/test/keystore", "keystore"+strconv.Itoa(0)+".key")
)

func TestKeystore(t *testing.T) {
	utils.Log.Info().Str("start", "1").Send()
	CreateKeystore()
	utils.Log.Info().Str("start", "2").Send()
	LoadKeystore()
	utils.Log.Info().Str("end", "1").Send()
	//MnemonicExample()
}

func CreateKeystore() {
	//utils.Log.Info().Str("开始创建密钥库", "1111111").Send()
	kst := NewKeystoreSingle(keystorePath, addrPre)
	ERR := kst.CreateRand(pwd, pwd, pwd, pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("创建key文件错误:", ERR.String()).Send()
		return
	}
	//utils.Log.Info().Str("开始创建密钥库", "1111111").Send()
	//获取收款地址列表
	coinAddrList := kst.GetCoinAddrAll()
	if len(coinAddrList) == 0 {
		utils.Log.Info().Str("收款地址数量为 0", "").Send()
	}
	for _, one := range coinAddrList {
		utils.Log.Info().Str("收款地址", one.GetAddrStr()).Send()
	}
	//utils.Log.Info().Str("开始创建密钥库", "1111111").Send()
	//获取网络地址
	netAddr, ERR := kst.GetNetAddrInfo(pwd)
	if ERR.CheckFail() {
		utils.Log.Info().Str("ERR", ERR.String()).Send()
	} else {
		utils.Log.Info().Str("网络地址", netAddr.GetAddrStr()).Send()
	}
	//获取dh公钥
	dhKey, ERR := kst.GetDHAddrInfo(pwd)
	if ERR.CheckFail() {
		utils.Log.Info().Str("ERR", ERR.String()).Send()
		return
	}
	keyPair, ERR := dhKey.GetDHKeyPair(pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("解密公钥错误:", ERR.String()).Send()
	} else {
		puk := keyPair.GetPublicKey()
		utils.Log.Info().Hex("DH公钥", puk[:]).Send()
	}
}

func LoadKeystore() {
	kst := NewKeystoreSingle(keystorePath, addrPre)
	ERR := kst.Load()
	if ERR.CheckFail() {
		utils.Log.Error().Str("加载key文件错误:", ERR.String()).Send()
		return
	}
	//utils.Log.Info().Int("收款地址数量", len(kst.Addrs)).Send()
	//utils.Log.Info().Int("收款地址数量", len(kst.NetAddrs)).Send()
	//utils.Log.Info().Int("收款地址数量", len(kst.DHAddrs)).Send()
	//获取收款地址列表
	coinAddrList := kst.GetCoinAddrAll()
	if len(coinAddrList) == 0 {
		utils.Log.Info().Str("收款地址数量为 0", "").Send()
	}
	for _, one := range coinAddrList {
		utils.Log.Info().Str("收款地址", one.GetAddrStr()).Send()
	}
	//获取网络地址
	netAddr, ERR := kst.GetNetAddrInfo(pwd)
	if ERR.CheckFail() {
		utils.Log.Info().Str("ERR", ERR.String()).Send()
	} else {
		utils.Log.Info().Str("网络地址", netAddr.GetAddrStr()).Send()
	}
	//获取dh公钥
	dhKey, ERR := kst.GetDHAddrInfo(pwd)
	if ERR.CheckFail() {
		utils.Log.Info().Str("ERR", ERR.String()).Send()
		return
	}
	keyPair, ERR := dhKey.GetDHKeyPair(pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("解密公钥错误:", ERR.String()).Send()
	} else {
		puk := keyPair.GetPublicKey()
		utils.Log.Info().Hex("DH公钥", puk[:]).Send()
	}
}

func TestMnemonicExample(t *testing.T) {
	SetCoinTypeCustom(0x800055aa, "TEST")
	SetCoinTypeCustomDH(0x800055aa + 1)
	keystore := NewKeystoreSingle(keystorePath, addrPre)
	if ERR := keystore.Load(); ERR.CheckFail() {
		//keystore CreateNewKeystore(pwd)
	}
	words := "salt slush prepare manual labor van snack person achieve recycle clever join"
	fmt.Println("打印助记词:", words)
	keystore = NewKeystoreSingle(keystorePath, addrPre)
	ERR := keystore.ImportMnemonic(words, pwd, pwd, pwd, pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	//fmt.Println("", keystore.GetCoinAddrAll())
	fmt.Println("第一个地址", keystore.GetCoinAddrAll()[0].GetAddrStr())
	words, ERR = keystore.ExportMnemonic(pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	fmt.Println("打印助记词:", words)

}

/*
生成各币种地址
*/
func TestHdExample(t *testing.T) {
	keystore := NewKeystoreSingle(keystorePath, addrPre)
	if ERR := keystore.Load(); ERR.CheckFail() {
		//keystore CreateNewKeystore(pwd)
	}
	//"range sheriff try enroll deer over ten level bring display stamp recycle"
	words := "range sheriff try enroll deer over ten level bring display stamp recycle"
	fmt.Println("打印助记词:", words)
	keystore = NewKeystoreSingle(keystorePath, addrPre)
	ERR := keystore.ImportMnemonic(words, pwd, pwd, pwd, pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	words, ERR = keystore.ExportMnemonic(pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	fmt.Println("打印助记词:", words)
	addrInfo, ERR := keystore.CreateAddressCoinType("", pwd, pwd, config.COINTYPE_BTC)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	//BTC:  1AwEPfoojHnKrhgt1vfuZAhrvPrmz7Rh44
	utils.Log.Info().Str("BTC 地址", addrInfo.GetAddrStr()).Send()
	addrInfo, ERR = keystore.CreateAddressCoinType("", pwd, pwd, config.COINTYPE_TRX)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	//TRX:  TXaMXTQgtdV6iqxtmQ7HNnqzXRoJKXfFAz
	utils.Log.Info().Str("TRX 地址", addrInfo.GetAddrStr()).Send()

	//LTC:  LLCaMFT8AKjDTvz1Ju8JoyYXxuug4PZZmS

}

func TestHdTest(t *testing.T) {
	words := "range sheriff try enroll deer over ten level bring display stamp recycle"
	var seed []byte
	var err error
	start := time.Now()
	for range 100 {
		seed, err = bip39.NewSeedWithErrorChecking(words, "")
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("花费时间", time.Now().Sub(start))
	fmt.Println("种子", hex.EncodeToString(seed))

	length := 12
	entropy, err := bip39.NewEntropy(length / 3 * 32)
	if err != nil {
		panic(err)
	}
	fmt.Println("新种子", hex.EncodeToString(entropy))
	words, err = bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}
	fmt.Println("新助记词", words)
	seed, err = bip39.EntropyFromMnemonic(words) //12位助记词  原生的seed 这个seed是没有加密过的
	if err != nil {
		panic(err)
	}
	fmt.Println("新种子", hex.EncodeToString(seed))

}
