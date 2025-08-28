package keystore

import (
	"golang.org/x/crypto/ed25519"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2"
	"web3_gui/utils"
)

type KeystoreInterface interface {
	GetNetAddrKeyPair(password string) (puk []byte, prk []byte, ERR utils.ERROR)
	GetDhAddrKeyPair(password string) (*keystore.KeyPair, utils.ERROR)
	GetCoinAddrAll() []keystore.CoinAddressInfo

	GetDHKeyPair() DHKeyPair
	CreateNetAddr(password, netAddressPassword string) (ed25519.PrivateKey, ed25519.PublicKey, error)
	GetNetAddr(pwd string) (ed25519.PrivateKey, ed25519.PublicKey, error)
	GetNewAddr(password, newAddressPassword string) (crypto.AddressCoin, error)
	GetAddr() []*AddressInfo
	FindPuk(puk []byte) (addrInfo AddressInfo, ok bool)
	GetCoinbase() *AddressInfo
	FindAddress(addr crypto.AddressCoin) (addrInfo AddressInfo, ok bool)
	GetAddrAll() []*AddressInfo
	GetPukByAddr(addr crypto.AddressCoin) (puk ed25519.PublicKey, ok bool)
	GetKeyByPuk(puk []byte, password string) (rand []byte, prk ed25519.PrivateKey, err error)
	GetKeyByAddr(addr crypto.AddressCoin, password string) (rand []byte, prk ed25519.PrivateKey, puk ed25519.PublicKey, err error)

	//下面是不用实现的接口，之前rpc接口使用的
	GetNewAddrByName(name string, password string, newAddressPassword string) (crypto.AddressCoin, error)
	UpdatePwd(oldpwd, newpwd string) (ok bool, err error)
	UpdateNetAddrPwd(oldpwd, newpwd string) (ok bool, err error)
	UpdateAddrPwd(addr, oldpwd, newpwd string) (ok bool, err error)
	ImportMnemonic(words, pwd, firstCoinAddressPassword, netAddressAndDHkeyPassword string) error
	ExportMnemonic(pwd string) (string, error)
	ImportMnemonicEncry(words, wordsPwd, seedPwd, firstCoinAddressPassword, netAddressAndDHkeyPassword string) error
	ExportMnemonicEncry(seedPwd, wordsPwd string) (string, error)
}

type DHKeyPair struct {
	Index     uint64           `json:"index"`     //棘轮数量
	KeyPair   keystore.KeyPair `json:"keypair"`   //
	CheckHash []byte           `json:"checkhash"` //主私钥和链编码加密验证hash值
	SubKey    []byte           `json:"subKey"`    //子密钥
}
