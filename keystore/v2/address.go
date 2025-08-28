package keystore

import (
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/keystore/v2/config"
	"web3_gui/utils"
)

func AddressFromB58String(str string) coin_address.AddressCoin {
	return coin_address.AddressFromB58String(str)
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pre string, pubKey []byte) (coin_address.AddressCoin, utils.ERROR) {
	if config.ShortAddress {
		return coin_address.BuildAddr_old(pre, pubKey)
	}
	return coin_address.BuildAddr(pre, pubKey)
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr_Old(pre string, pubKey []byte) (coin_address.AddressCoin, utils.ERROR) {
	return coin_address.BuildAddr_old(pre, pubKey)
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddrByData(pre string, data []byte) (coin_address.AddressCoin, utils.ERROR) {
	return coin_address.BuildAddrByData(pre, data)
}

/*
通过公钥生成地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr_old(pre string, pubKey []byte) (coin_address.AddressCoin, utils.ERROR) {
	return coin_address.BuildAddr_old(pre, pubKey)
}

/*
解析前缀
*/
func ParseAddrPrefix(addr coin_address.AddressCoin) string {
	return coin_address.ParseAddrPrefix(addr)
}

/*
判断有效地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func ValidAddrCoin(pre string, addr coin_address.AddressCoin) bool {
	return coin_address.ValidAddrCoin(pre, addr)
}

/*
判断有效地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func ValidAddrNet(pre string, addr coin_address.AddressCoin) bool {
	return coin_address.ValidLongAddrCoin(pre, addr)
}

/*
判断有效地址
@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func ValidAddrCoin_old(pre string, addr coin_address.AddressCoin) bool {
	return coin_address.ValidAddrCoin_old(pre, addr)
}

/*
检查公钥生成的网络地址是否一样
@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddrNet(pre string, pubKey []byte, addr coin_address.AddressCoin) (bool, utils.ERROR) {
	return coin_address.CheckPukAddrNet(pre, pubKey, addr)
}
