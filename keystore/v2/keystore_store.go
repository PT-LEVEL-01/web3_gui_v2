package keystore

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/keystore/v2/config"
	"web3_gui/keystore/v2/protobuf/go_protobuf"
	"web3_gui/utils"
)

/*
加载keystore
*/
func (this *Keystore) Load() utils.ERROR {
	bs, ERR := this.store.Load()
	if ERR.CheckFail() {
		return ERR
	}
	ks, err := ParseKeystore(bs)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	this.addrPre = ks.GetAddrPre()        //
	this.Nickname = ks.Nickname           //密钥库昵称
	this.CryptedSeed = ks.CryptedSeed     //加密保存的种子
	this.Salt = ks.Salt                   //盐
	this.CheckHash = ks.CheckHash         //主私钥和链编码加密验证hash值，困难级别
	this.Rounds = ks.Rounds               //pbkdf2迭代次数，困难级别
	this.CheckHashTemp = ks.CheckHashTemp //主私钥和链编码加密验证hash值，容易级别
	this.RoundsTemp = ks.RoundsTemp       //pbkdf2迭代次数，容易级别
	this.MnemonicLang = ks.MnemonicLang   //助记词语言
	this.Addrs = ks.Addrs                 //已经生成的地址列表
	//this.pukMap = ks.pukMap               //key:string=公钥;value:*CoinAddressInfo=地址密钥等信息;
	this.NetAddrs = ks.NetAddrs     //已经生成的网络地址列表
	this.DHAddrs = ks.DHAddrs       //DH密钥列表
	this.AddrsOther = ks.AddrsOther //已生成其他币种的地址列表
	if !this.CheckIntact() {
		return utils.NewErrorBus(config.ERROR_code_wallet_incomplete, "")
	}
	for j, one := range this.Addrs {
		one.keystore = this
		addrInfo := this.Addrs[j]
		this.addrMap.Store(utils.Bytes2string(one.Addr.Bytes()), addrInfo)
		//this.pukMap.Store(utils.Bytes2string(one.Puk), addrInfo)
	}
	for _, one := range this.NetAddrs {
		one.keystore = this
	}
	for _, one := range this.DHAddrs {
		one.keystore = this
	}
	for _, one := range this.AddrsOther {
		one.keystore = this
	}
	//utils.Log.Info().Int("收款地址数量", len(this.Addrs)).Send()
	//utils.Log.Info().Int("收款地址数量", len(this.NetAddrs)).Send()
	//utils.Log.Info().Int("收款地址数量", len(this.DHAddrs)).Send()
	return utils.NewErrorSuccess()
}

/*
从磁盘文件加载keystore
*/
func (this *Keystore) Save() utils.ERROR {
	if this.wallet != nil {
		return this.wallet.Save()
	}
	bs, err := this.Proto()
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	ERR := this.store.Save(*bs)
	return ERR
}

/*
转化
*/
func (this *Keystore) Conver() *go_protobuf.Keystore {
	//this.lock.RLock()
	//defer this.lock.RUnlock()
	ksProto := go_protobuf.Keystore{}
	ksProto.AddrPre = this.GetAddrPre()
	ksProto.Nickname = this.Nickname
	ksProto.CryptedSeed = this.CryptedSeed
	ksProto.Salt = this.Salt
	ksProto.Rounds = this.Rounds
	ksProto.CheckHash = this.CheckHash
	ksProto.Version = this.Version
	ksProto.CoinAddr = make([]*go_protobuf.CoinAddress, 0)
	for _, one := range this.Addrs {
		addrInfo := one.Conver()
		ksProto.CoinAddr = append(ksProto.CoinAddr, addrInfo)
	}
	ksProto.NetAddr = make([]*go_protobuf.CoinAddress, 0)
	for _, one := range this.NetAddrs {
		addrInfo := one.Conver()
		ksProto.NetAddr = append(ksProto.NetAddr, addrInfo)
	}
	ksProto.DHAddr = make([]*go_protobuf.CoinAddress, 0)
	for _, one := range this.DHAddrs {
		addrInfo := one.Conver()
		ksProto.DHAddr = append(ksProto.DHAddr, addrInfo)
	}
	ksProto.CoinAddrOther = make([]*go_protobuf.CoinAddress, 0)
	for _, one := range this.AddrsOther {
		addrInfo := one.Conver()
		ksProto.CoinAddrOther = append(ksProto.CoinAddrOther, addrInfo)
	}
	return &ksProto
}

/*
格式化成[]byte
*/
func (this *Keystore) Proto() (*[]byte, error) {
	ks := this.Conver()
	bs, err := ks.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func ParseKeystore(bs []byte) (*Keystore, error) {
	ksProto := new(go_protobuf.Keystore)
	err := proto.Unmarshal(bs, ksProto)
	if err != nil {
		return nil, err
	}
	ks := ConverProtoKeystore(ksProto)
	return ks, nil
}

func ConverProtoKeystore(ksProto *go_protobuf.Keystore) *Keystore {
	ks := newKeystore(ksProto.AddrPre)
	//ks.addrPre = ksProto.AddrPre
	ks.Nickname = ksProto.Nickname
	ks.CryptedSeed = ksProto.CryptedSeed
	ks.Salt = ksProto.Salt
	ks.Rounds = ksProto.Rounds
	ks.CheckHash = ksProto.CheckHash
	ks.Version = ksProto.Version
	for _, one := range ksProto.CoinAddr {
		coinAddr := ConverProtoCoinAddr(one)
		ks.Addrs = append(ks.Addrs, coinAddr)
		ks.addrMap.Store(utils.Bytes2string(coinAddr.Addr.Bytes()), coinAddr)
	}
	for _, one := range ksProto.NetAddr {
		coinAddr := ConverProtoCoinAddr(one)
		ks.NetAddrs = append(ks.NetAddrs, coinAddr)
	}
	for _, one := range ksProto.DHAddr {
		coinAddr := ConverProtoCoinAddr(one)
		ks.DHAddrs = append(ks.DHAddrs, coinAddr)
	}
	for _, one := range ksProto.CoinAddrOther {
		coinAddr := ConverProtoCoinAddr(one)
		ks.AddrsOther = append(ks.AddrsOther, coinAddr)
	}
	return ks
}
