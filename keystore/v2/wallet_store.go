package keystore

import (
	"github.com/gogo/protobuf/proto"
	"sync"
	"sync/atomic"
	"web3_gui/keystore/v2/config"
	"web3_gui/keystore/v2/protobuf/go_protobuf"
	"web3_gui/utils"
)

/*
加载keystore
*/
func (this *Wallet) Load() utils.ERROR {
	bs, ERR := this.store.Load()
	if ERR.CheckFail() {
		return ERR
	}
	wallet, err := ParseWallet(bs)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	this.addrPre = wallet.addrPre
	this.UseKeystore = wallet.UseKeystore
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	this.Keystore = wallet.Keystore
	for _, one := range this.Keystore {
		one.wallet = this
	}
	if !this.CheckIntact() {
		return utils.NewErrorBus(config.ERROR_code_wallet_incomplete, "")
	}
	return utils.NewErrorSuccess()
}

/*
从磁盘文件加载keystore
*/
func (this *Wallet) Save() utils.ERROR {
	bs, err := this.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	//utils.Log.Info().Interface("钱包保存方法", this.store).Send()
	ERR := this.store.Save(*bs)
	return ERR
}

/*
格式化成[]byte
*/
func (this *Wallet) Proto() (*[]byte, error) {
	wallet := go_protobuf.Wallet{
		AddrPre:     this.GetAddrPre(),
		Keystore:    make([]*go_protobuf.Keystore, 0, len(this.Keystore)),
		UseKeystore: uint64(this.UseKeystore),
	}
	for _, one := range this.Keystore {
		ksOne := one.Conver()
		wallet.Keystore = append(wallet.Keystore, ksOne)
	}
	bs, err := wallet.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func ParseWallet(bs []byte) (*Wallet, error) {
	wallet := new(go_protobuf.Wallet)
	err := proto.Unmarshal(bs, wallet)
	if err != nil {
		return nil, err
	}
	newWallet := &Wallet{
		addrPre:     wallet.AddrPre,          //
		lock:        new(sync.RWMutex),       //
		Keystore:    make([]*Keystore, 0),    //保存的多个密钥
		UseKeystore: int(wallet.UseKeystore), //上次使用的密钥库索引
		CheckPwd:    new(atomic.Bool),        //是否
	}
	for _, one := range wallet.Keystore {
		ksOne := ConverProtoKeystore(one)
		ksOne.wallet = newWallet
		for _, addrOne := range ksOne.Addrs {
			addrOne.keystore = ksOne
		}
		for _, addrOne := range ksOne.NetAddrs {
			addrOne.keystore = ksOne
		}
		for _, addrOne := range ksOne.DHAddrs {
			addrOne.keystore = ksOne
		}
		for _, addrOne := range ksOne.AddrsOther {
			addrOne.keystore = ksOne
		}
		newWallet.Keystore = append(newWallet.Keystore, ksOne)
	}
	return newWallet, nil
}
