package models

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/cross_chain/usdt/protobuf/go_protos"
)

type WalletAddressInfo struct {
	Purpose    uint32 //固定为 44'，表示 BIP-44 标准。
	CoinType   uint32 //指定加密货币的类型（例如，0' 表示 Bitcoin，2' 表示 Litecoin 等）。
	Account    uint32 //表示账户索引，允许用户在同一钱包中有多个账户。
	Change     uint32 //表示地址类型（0 表示接收地址，1 表示找零地址）。
	Index      uint32 //表示地址索引，用于生成每个账户的具体地址。
	AddressBs  []byte //收款地址链上字节
	AddressStr string //收款地址字符串
}

/*
序列化
*/
func (this *WalletAddressInfo) Proto() (*[]byte, error) {
	ccai := go_protos.CrossChainAddrInfo{
		Purpose:  this.Purpose,
		CoinType: this.CoinType,
		Account:  this.Account,
		Change:   this.Change,
		Index:    this.Index,
		AddrStr:  this.AddressStr,
		AddrBs:   this.AddressBs,
	}
	bs, err := ccai.Marshal()
	return &bs, err
}

/*
解析用户基本信息
*/
func ParseWalletAddressInfo(bs *[]byte) (*WalletAddressInfo, error) {
	if bs == nil {
		return nil, nil
	}
	bhp := new(go_protos.CrossChainAddrInfo)
	err := proto.Unmarshal(*bs, bhp)
	if err != nil {
		return nil, err
	}
	bh := WalletAddressInfo{
		Purpose:    bhp.Purpose,  //固定为 44'，表示 BIP-44 标准。
		CoinType:   bhp.CoinType, //指定加密货币的类型（例如，0' 表示 Bitcoin，2' 表示 Litecoin 等）。
		Account:    bhp.Account,  //表示账户索引，允许用户在同一钱包中有多个账户。
		Change:     bhp.Change,   //表示地址类型（0 表示接收地址，1 表示找零地址）。
		Index:      bhp.Index,    //表示地址索引，用于生成每个账户的具体地址。
		AddressBs:  bhp.AddrBs,   //收款地址链上字节
		AddressStr: bhp.AddrStr,  //收款地址字符串
	}
	return &bh, nil
}
