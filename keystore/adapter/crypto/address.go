package crypto

import (
	"bytes"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
)

type AddressCoin coin_address.AddressCoin

func (this AddressCoin) B58String() string {
	key := coin_address.AddressCoin([]byte(this))
	return key.B58String()
}
func (this AddressCoin) GetPre() string {
	key := coin_address.AddressCoin(this)
	return key.GetPre()
}

func (this AddressCoin) Data() []byte {
	key := coin_address.AddressCoin([]byte(this))
	return key.Data()
}

func AddressFromB58String(str string) AddressCoin {
	key := keystore.AddressFromB58String(str)
	return AddressCoin(key)
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pre string, pubKey []byte) AddressCoin {
	key, ERR := keystore.BuildAddr(pre, pubKey)
	if ERR.CheckFail() {
		return nil
	}
	return AddressCoin(key)
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddrOld(pre string, pubKey []byte) AddressCoin {
	key, ERR := keystore.BuildAddr_old(pre, pubKey)
	if ERR.CheckFail() {
		return nil
	}
	return AddressCoin(key)
}

func ParseAddrPrefix(addr AddressCoin) string {
	return keystore.ParseAddrPrefix(coin_address.AddressCoin(addr))
}

/*
判断有效地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func ValidAddr(pre string, addr AddressCoin) bool {
	return keystore.ValidAddrCoin(pre, coin_address.AddressCoin(addr))
}

/*
检查公钥生成的地址是否一样
@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddr(pre string, pubKey []byte, addr AddressCoin) bool {
	tagAddr := BuildAddr(pre, pubKey)
	return bytes.Equal(tagAddr, addr)
}
