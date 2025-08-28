package keystore

import (
	"testing"
	"time"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

func TestStoreMem(t *testing.T) {
	w := NewWallet("", addrPre)
	w.SetStore(NewStoreMem())
	ERR := w.AddKeystoreRand(pwd)
	if ERR.CheckFail() {
		utils.Log.Error().Str("添加一个随机数钱包报错", ERR.String()).Send()
		return
	}

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
	world, ERR := keys.ExportMnemonic(pwd)
	if ERR.CheckFail() {
		panic(ERR.String())
	}
	utils.Log.Info().Str("导出助记词", world).Send()

}

/*
保存到文件
*/
type StoreMem struct {
	bs *[]byte
}

func NewStoreMem() *StoreMem {
	return &StoreMem{}
}

/*
从文件中读取
*/
func (this *StoreMem) Load() ([]byte, utils.ERROR) {
	if this.bs == nil {
		return nil, utils.NewErrorSuccess()
	}
	return *this.bs, utils.NewErrorSuccess()
}

/*
保存到文件中
*/
func (this *StoreMem) Save(bs []byte) utils.ERROR {
	//utils.Log.Info().Str("保存文件路径", this.filePath).Send()
	//err := utils.SaveFile(this.filePath, &bs)
	this.bs = &bs
	return utils.NewErrorSuccess()
}
