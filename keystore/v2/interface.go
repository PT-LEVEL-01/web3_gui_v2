package keystore

import (
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/utils"
)

type KeystoreInterface interface {
	GetNetAddrKeyPair(password string) (puk []byte, prk []byte, ERR utils.ERROR)
	GetDhAddrKeyPair(password string) (*KeyPair, utils.ERROR)
	GetCoinAddrAll() []CoinAddressInfo
}

func GetCoinTypeCustom() (uint32, string) {
	return coin_address.GetCoinTypeCustom()
}

func SetCoinTypeCustom(coinType uint32, coinName string) {
	coin_address.SetCoinTypeCustom(coinType, coinName)
}

func GetCoinTypeCustomDH() uint32 {
	return coin_address.GetCoinTypeCustomDH()
}

func SetCoinTypeCustomDH(coinType uint32) {
	coin_address.SetCoinTypeCustomDH(coinType)
}
