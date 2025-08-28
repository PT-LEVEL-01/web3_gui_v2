package keystore

import (
	"encoding/hex"
	"golang.org/x/crypto/ed25519"
	"web3_gui/keystore/adapter/crypto"
)

type AddressInfo struct {
	Index     uint64             `json:"index"`     //棘轮数量
	Nickname  string             `json:"nickname"`  //地址昵称
	Addr      crypto.AddressCoin `json:"addr"`      //收款地址
	Puk       ed25519.PublicKey  `json:"puk"`       //公钥
	SubKey    []byte             `json:"subKey"`    //子密钥
	AddrStr   string             `json:"-"`         //
	PukStr    string             `json:"-"`         //
	CheckHash []byte             `json:"checkhash"` //主私钥和链编码加密验证hash值
	Version   int                `json:"version"`   //地址版本
}

func (this *AddressInfo) GetAddrStr() string {
	if this.AddrStr == "" {
		this.AddrStr = this.Addr.B58String()
	}
	return this.AddrStr
}

func (this *AddressInfo) GetPukStr() string {
	if this.PukStr == "" {
		this.PukStr = hex.EncodeToString(this.Puk)
	}
	return this.PukStr
}
