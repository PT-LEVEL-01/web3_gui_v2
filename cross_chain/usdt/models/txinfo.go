package models

import (
	"github.com/gogo/protobuf/proto"
	"web3_gui/cross_chain/usdt/protobuf/go_protos"
)

type TxInfo struct {
	NodeAddr        string //节点地址
	WalletAddr      string //钱包地址
	ChainName       string //链名称
	ChainCoinType   uint32 //链编号
	BlockHeight     uint64 //区块高度
	BlockHash       string //区块hash
	TxId            string //交易id
	TimeBlock       int64  //交易时间
	Status          string //状态
	From            string //转账地址
	To              string // 目标地址
	ContractAddress string //
	Fee             int64  // 交易手续费
	FeeCny          string // 交易手续费
	FeeUsd          string // 交易手续费
	Confirm         int    //交易确认数
	Amount          int64  // 交易数量
	TokenName       string // 代币名称
	Send            bool   // 是否是发送
	TimeStamp       int64  //更新时间
	DocId           int
	Type            string
	Nonce           uint64
	Extra           interface{}
	Unique          string
	UtxoInput       []*UtxoAmount // btc链存在多个输入和多个输出，btc等utxo模型链使用
	UtxoOutput      []*UtxoAmount // btc链存在多个输入和多个输出，btc等utxo模型链使用
}

type UtxoAmount struct {
	Address string
	Amount  uint64
}

/*
序列化
*/
func (this *TxInfo) Proto() (*[]byte, error) {
	ccai := go_protos.CrossChainTxInfo{
		NodeAddr:        this.NodeAddr,        //节点地址
		WalletAddr:      this.WalletAddr,      //钱包地址
		ChainName:       this.ChainName,       //链名称
		ChainCoinType:   this.ChainCoinType,   //链编号
		BlockHeight:     this.BlockHeight,     //区块高度
		BlockHash:       this.BlockHash,       //区块hash
		TxId:            this.TxId,            //交易id
		TimeBlock:       this.TimeBlock,       //交易时间
		Status:          this.Status,          //状态
		From:            this.From,            //转账地址
		To:              this.To,              // 目标地址
		ContractAddress: this.ContractAddress, //
		Fee:             this.Fee,             // 交易手续费
		FeeCny:          this.FeeCny,          // 交易手续费
		FeeUsd:          this.FeeUsd,          // 交易手续费
		Confirm:         int64(this.Confirm),  //交易确认数
		Amount:          this.Amount,          // 交易数量
		TokenName:       this.TokenName,       // 代币名称
		Send:            this.Send,            // 是否是发送
		TimeStamp:       this.TimeStamp,       //更新时间
		DocId:           int64(this.DocId),
		Type:            this.Type,
		Nonce:           this.Nonce,
		Unique:          this.Unique,
		UtxoInput:       make([]*go_protos.CrossChainUtxoAmount, 0, len(this.UtxoInput)),  // btc链存在多个输入和多个输出，btc等utxo模型链使用
		UtxoOutput:      make([]*go_protos.CrossChainUtxoAmount, 0, len(this.UtxoOutput)), // btc链存在多个输入和多个输出，btc等utxo模型链使用
	}
	for _, one := range this.UtxoInput {
		utxo := go_protos.CrossChainUtxoAmount{
			Address: one.Address,
			Amount:  one.Amount,
		}
		ccai.UtxoInput = append(ccai.UtxoInput, &utxo)
	}
	for _, one := range this.UtxoOutput {
		utxo := go_protos.CrossChainUtxoAmount{
			Address: one.Address,
			Amount:  one.Amount,
		}
		ccai.UtxoOutput = append(ccai.UtxoOutput, &utxo)
	}
	bs, err := ccai.Marshal()
	return &bs, err
}

/*
解析用户基本信息
*/
func ParseTxInfo(bs *[]byte) (*TxInfo, error) {
	if bs == nil {
		return nil, nil
	}
	bhp := new(go_protos.CrossChainTxInfo)
	err := proto.Unmarshal(*bs, bhp)
	if err != nil {
		return nil, err
	}
	bh := TxInfo{
		NodeAddr:        bhp.NodeAddr,        //节点地址
		WalletAddr:      bhp.WalletAddr,      //钱包地址
		ChainName:       bhp.ChainName,       //链名称
		ChainCoinType:   bhp.ChainCoinType,   //链编号
		BlockHeight:     bhp.BlockHeight,     //区块高度
		BlockHash:       bhp.BlockHash,       //区块hash
		TxId:            bhp.TxId,            //交易id
		TimeBlock:       bhp.TimeBlock,       //交易时间
		Status:          bhp.Status,          //状态
		From:            bhp.From,            //转账地址
		To:              bhp.To,              // 目标地址
		ContractAddress: bhp.ContractAddress, //
		Fee:             bhp.Fee,             // 交易手续费
		FeeCny:          bhp.FeeCny,          // 交易手续费
		FeeUsd:          bhp.FeeUsd,          // 交易手续费
		Confirm:         int(bhp.Confirm),    //交易确认数
		Amount:          bhp.Amount,          // 交易数量
		TokenName:       bhp.TokenName,       // 代币名称
		Send:            bhp.Send,            // 是否是发送
		TimeStamp:       bhp.TimeStamp,       //更新时间
		DocId:           int(bhp.DocId),
		Type:            bhp.Type,
		Nonce:           bhp.Nonce,
		Unique:          bhp.Unique,
		UtxoInput:       make([]*UtxoAmount, 0, len(bhp.UtxoInput)),  // btc链存在多个输入和多个输出，btc等utxo模型链使用
		UtxoOutput:      make([]*UtxoAmount, 0, len(bhp.UtxoOutput)), // btc链存在多个输入和多个输出，btc等utxo模型链使用
	}
	for _, one := range bhp.UtxoInput {
		utxo := UtxoAmount{
			Address: one.Address,
			Amount:  one.Amount,
		}
		bh.UtxoInput = append(bh.UtxoInput, &utxo)
	}
	for _, one := range bhp.UtxoOutput {
		utxo := UtxoAmount{
			Address: one.Address,
			Amount:  one.Amount,
		}
		bh.UtxoOutput = append(bh.UtxoOutput, &utxo)
	}
	return &bh, nil
}
