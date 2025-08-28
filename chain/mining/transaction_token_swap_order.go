package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

func init() {
	tpc := new(TokenSwapOrderController)
	RegisterTransaction(config.Wallet_tx_type_token_swap_order, tpc)
}

/*
token订单
*/
type TxTokenSwapOrder struct {
	TxBase
	OrderID1     []byte   `json:"oid1"` //对手订单ID
	OrderID2     []byte   `json:"oid2"` //订单ID
	TokenAAmount *big.Int `json:"ta_a"` //TokenA成交量
	TokenBAmount *big.Int `json:"tb_a"` //TokenB成交量
}

/*
token订单
*/
type TxTokenSwapOrder_VO struct {
	TxBaseVO
	OrderID1     string `json:"order_id1"`     //对手订单ID
	OrderID2     string `json:"order_id2"`     //订单ID
	TokenAAmount uint64 `json:"tokena_amount"` //TokenA成交量
	TokenBAmount uint64 `json:"tokenb_amount"` //TokenB成交量
}

/*
用于地址和txid格式化显示
*/
func (this *TxTokenSwapOrder) GetVOJSON() interface{} {
	return TxTokenSwapOrder_VO{
		TxBaseVO:     this.TxBase.ConversionVO(),
		OrderID1:     hex.EncodeToString(this.OrderID1),
		OrderID2:     hex.EncodeToString(this.OrderID2),
		TokenAAmount: this.TokenAAmount.Uint64(),
		TokenBAmount: this.TokenBAmount.Uint64(),
	}
}

/*
构建hash值得到交易id
*/
func (this *TxTokenSwapOrder) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_token_swap_order)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
构建hash值得到交易id
*/
func (this *TxTokenSwapOrder) BuildSwapTxKey() string {
	buf := bytes.NewBuffer(nil)
	buf.Write(this.OrderID1)
	buf.Write(this.OrderID2)
	buf.Write(this.TokenAAmount.Bytes())
	buf.Write(this.TokenBAmount.Bytes())
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_token_swap_order)
	return utils.Bytes2string(append(id, utils.Hash_SHA3_256(buf.Bytes())...))
}

/*
格式化成[]byte
*/
func (this *TxTokenSwapOrder) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vins = append(vins, &go_protos.Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
			Nonce: one.Nonce.Bytes(),
			Puk:   one.Puk,
			Sign:  one.Sign,
		})
	}
	vouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	txBase := go_protos.TxBase{
		Hash:       this.Hash,
		Type:       this.Type,
		VinTotal:   this.Vin_total,
		Vin:        vins,
		VoutTotal:  this.Vout_total,
		Vout:       vouts,
		Gas:        this.Gas,
		LockHeight: this.LockHeight,
		Payload:    this.Payload,
		BlockHash:  this.BlockHash,
		Comment:    this.Comment,
	}

	txTokenOrder := go_protos.TxTokenSwapOrder{
		TxBase:       &txBase,
		OrderID1:     this.OrderID1,
		OrderID2:     this.OrderID2,
		TokenAAmount: this.TokenAAmount.Bytes(),
		TokenBAmount: this.TokenBAmount.Bytes(),
	}

	// txTokenOrder.Marshal()
	bs, err := txTokenOrder.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
格式化成json字符串
*/
func (this *TxTokenSwapOrder) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)

	buf.Write(this.OrderID1)
	buf.Write(this.OrderID2)
	buf.Write(this.TokenAAmount.Bytes())
	buf.Write(this.TokenBAmount.Bytes())

	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *TxTokenSwapOrder) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)

	buf := bytes.NewBuffer(*signDst)
	buf.Write(this.OrderID1)
	buf.Write(this.OrderID2)
	buf.Write(this.TokenAAmount.Bytes())
	buf.Write(this.TokenBAmount.Bytes())

	bs := buf.Bytes()
	sign := keystore.Sign(*key, bs)

	return &sign

}

/*
把相同地址的交易输出合并在一起
*/
func (this *TxTokenSwapOrder) MergeTokenVout() {
}

/*
检查交易是否合法
*/
func (this *TxTokenSwapOrder) CheckSign() error {
	if len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	//if len(this.Vin[0].Nonce.Bytes()) == 0 {
	//	// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
	//	return config.ERROR_pay_nonce_is_nil
	//}
	if this.Vout_total != 0 {
		return config.ERROR_pay_vout_too_much
	}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费;3.输入不能重复。
	// vinMap := make(map[string]int)
	// inTotal := uint64(0)
	for i, one := range this.Vin {
		sign := this.GetSignSerialize(nil, uint64(i))
		buf := bytes.NewBuffer(*sign)
		buf.Write(this.OrderID1)
		buf.Write(this.OrderID2)
		buf.Write(this.TokenAAmount.Bytes())
		buf.Write(this.TokenBAmount.Bytes())
		bs := buf.Bytes()

		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		if !ed25519.Verify(puk, bs, one.Sign) {
			return config.ERROR_sign_fail
		}
	}

	return nil
}

/*
验证是否合法
*/
func (this *TxTokenSwapOrder) GetWitness() *crypto.AddressCoin {
	witness := crypto.BuildAddr(config.AddrPre, this.Vin[0].Puk)
	// witness, err := keystore.ParseHashByPubkey(this.Vin[0].Puk)
	// if err != nil {
	// 	return nil
	// }
	return &witness
}

/*
获取本交易总共花费的余额
*/
func (this *TxTokenSwapOrder) GetSpend() uint64 {
	return this.Gas
}

/*
检查重复的交易
*/
func (this *TxTokenSwapOrder) CheckRepeatedTx(txs ...TxItr) bool {
	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_token_swap_order {
			continue
		}

		if bytes.Equal(this.Hash, *one.GetHash()) {
			return false
		}
	}
	return true
}

/*
统计交易余额
*/
func (this *TxTokenSwapOrder) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}

	totalValue := this.Gas

	//余额中减去。
	from := this.Vin[0].GetPukToAddr()
	itemCount.Nonce[utils.Bytes2string(*from)] = this.Vin[0].Nonce

	frozenMap := make(map[uint64]int64, 0)
	frozenMap[0] = (0 - int64(totalValue))
	itemCount.AddItems[utils.Bytes2string(*from)] = &frozenMap

	return &itemCount
}

//======================= TokenSwapOrderController =======================

type TokenSwapOrderController struct {
}

func (this *TokenSwapOrderController) Factory() interface{} {
	return new(TxTokenSwapOrder)
}

func (this *TokenSwapOrderController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxTokenSwapOrder)
	err := txProto.Unmarshal(*bs)
	if err != nil {
		return nil, err
	}
	vins := make([]*Vin, 0)
	for _, one := range txProto.TxBase.Vin {
		nonce := new(big.Int).SetBytes(one.Nonce)
		vins = append(vins, &Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
			Puk:   one.Puk,
			Sign:  one.Sign,
			Nonce: *nonce,
		})
	}
	vouts := make([]*Vout, 0)
	for _, one := range txProto.TxBase.Vout {
		vouts = append(vouts, &Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	txBase := TxBase{}
	txBase.Hash = txProto.TxBase.Hash
	txBase.Type = txProto.TxBase.Type
	txBase.Vin_total = txProto.TxBase.VinTotal
	txBase.Vin = vins
	txBase.Vout_total = txProto.TxBase.VoutTotal
	txBase.Vout = vouts
	txBase.Gas = txProto.TxBase.Gas
	txBase.LockHeight = txProto.TxBase.LockHeight
	txBase.Payload = txProto.TxBase.Payload
	txBase.BlockHash = txProto.TxBase.BlockHash
	txBase.Comment = txProto.TxBase.Comment

	tx := &TxTokenSwapOrder{
		TxBase:       txBase,
		OrderID1:     txProto.OrderID1,
		OrderID2:     txProto.OrderID2,
		TokenAAmount: new(big.Int).SetBytes(txProto.TokenAAmount),
		TokenBAmount: new(big.Int).SetBytes(txProto.TokenBAmount),
	}
	return tx, nil
}

/*
5.统计
更新订单:
订单1: TokenAAmount=Old-amount; TokenBAmount=Old-(amount*price1)
订单2: TokenAAmount=Old-amount; TokenBAmount=Old-(amount*price)
更新账户:
订单1:锁定金额(B)=Old-(amount*price1); TokenBRefund=amount*(price1-price)
订单2:锁定金额(A)=Old-amount,TokenBAmount=Old-(amount*price)
*/
func (this *TokenSwapOrderController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {
	//锁定金额集合
	updateTokens := []*AddrToken{}
	unlockedTokens := []*AddrToken{}

	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_token_swap_order { //撮合订单统计
			continue
		}

		swapTx := txItr.(*TxTokenSwapOrder)
		//更新订单薄
		updateToken, unlockedToken := oBook.UpdateOrderV3(swapTx)
		updateTokens = append(updateTokens, updateToken...)
		unlockedTokens = append(unlockedTokens, unlockedToken...)

		//移除撮合交易缓存
		oBook.RemoveTxCache(swapTx) //统计移除
	}

	this.handleOrderToken(updateTokens, unlockedTokens)
}

// 处理Token撮合订单锁定金额和可用余额
func (this *TokenSwapOrderController) handleOrderToken(updateTokens, unlockedTokens []*AddrToken) {
	//合并重复可用金额
	mergeUpdateTokens := make(map[string]*AddrToken)
	mergeUpdateICom := make(map[string]*AddrToken)
	for _, one := range updateTokens {
		if one.TokenId == nil || len(one.TokenId) == 0 {
			if item, ok := mergeUpdateICom[utils.Bytes2string(one.Addr)]; ok {
				item.Value.Add(item.Value, one.Value)
			} else {
				mergeUpdateICom[utils.Bytes2string(one.Addr)] = &AddrToken{
					Addr:  one.Addr,
					Value: new(big.Int).Set(one.Value),
				}
			}
		} else {
			key := config.BuildDBKeyTokenAddrValue(one.TokenId, one.Addr)
			if item, ok := mergeUpdateTokens[utils.Bytes2string(*key)]; ok {
				item.Value.Add(item.Value, one.Value)
			} else {
				mergeUpdateTokens[utils.Bytes2string(*key)] = &AddrToken{
					TokenId: one.TokenId,
					Addr:    one.Addr,
					Value:   new(big.Int).Set(one.Value),
				}
			}
		}
	}

	//合并重复锁定金额
	mergeUnlockedTokens := make(map[string]*AddrToken)
	for _, one := range unlockedTokens {
		key := config.BuildDBKeyTokenAddrLockedValue(one.TokenId, one.Addr)
		if item, ok := mergeUnlockedTokens[utils.Bytes2string(*key)]; ok {
			item.Value.Add(item.Value, one.Value)
		} else {
			mergeUnlockedTokens[utils.Bytes2string(*key)] = &AddrToken{
				TokenId: one.TokenId,
				Addr:    one.Addr,
				Value:   new(big.Int).Set(one.Value),
			}
		}
	}

	for key, one := range mergeUnlockedTokens {
		lockKey := []byte(key)
		val := new(big.Int)
		bs, err := db.LevelDB.Find(lockKey)
		if err == nil {
			//更新金额
			val = val.SetBytes(*bs)
		}
		val.Sub(val, one.Value)

		if val.Cmp(big.NewInt(0)) <= 0 {
			db.LevelDB.Del(lockKey)
		} else {
			data := val.Bytes()
			db.LevelDB.Save(lockKey, &data)
		}
	}

	//代币
	for key, one := range mergeUpdateTokens {
		//更新代币可用金额
		updateKey := []byte(key)
		updateTokenNotSpendBalance(&updateKey, one.Value)
	}

	//主链币
	for _, one := range mergeUpdateICom {
		_, oldValue := db.GetNotSpendBalance(&one.Addr)
		newValue := new(big.Int).Set(one.Value)
		newValue.Add(newValue, new(big.Int).SetUint64(oldValue))
		db.SetNotSpendBalance(&one.Addr, newValue.Uint64())
	}
}

func (this *TokenSwapOrderController) CheckMultiplePayments(txItr TxItr) error {
	return nil
}

func (this *TokenSwapOrderController) SyncCount() {

}

func (this *TokenSwapOrderController) RollbackBalance() {
	// return new(Tx_account)
}

// 释放Token锁定金额
func (this *TokenSwapOrderController) handleOrderLockedTokenRelease(orderId []byte) {
}

/*
构建Token撮合订单交易
*/
func (this *TokenSwapOrderController) BuildTx(_ *sync.Map, _, _ *crypto.AddressCoin,
	_, _, _ uint64, _, _ string, params ...interface{}) (TxItr, error) {
	return nil, nil
}

// 构建撮合交易,不广播,放入缓存池
func BuildSwapTokenOrderByWitness(
	orderId1 []byte,
	orderId2 []byte,
	tokenAAmount *big.Int,
	tokenBAmount *big.Int,
) (*TxTokenSwapOrder, error) {
	//======================= 撮合交易 ===============================
	//查找余额
	vins := make([]*Vin, 0)
	baseCoinAddr := Area.Keystore.GetCoinbase()
	vin := Vin{
		Puk: baseCoinAddr.Puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	emptyVouts := make([]*Vout, 0)

	chain := GetLongChain()
	_, block := chain.GetLastBlock()
	//var txin *TxToken
	var txin *TxTokenSwapOrder
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_token_swap_order,         //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(emptyVouts)),                        //输出交易数量
			Vout:       emptyVouts,                                     //
			Gas:        0,                                              //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    []byte{},                                       //
		}

		txin = &TxTokenSwapOrder{
			TxBase:       base,
			OrderID1:     orderId1,
			OrderID2:     orderId2,
			TokenAAmount: tokenAAmount,
			TokenBAmount: tokenBAmount,
		}

		//给输出签名，防篡改
		for j, one := range txin.Vin {
			_, prk, err := config.Area.Keystore.GetKeyByPuk(one.Puk, config.Wallet_keystore_default_pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

			sign := txin.GetSign(&prk, uint64(j))
			txin.Vin[j].Sign = *sign
		}

		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}

	return txin, nil
}

/*
计算撮合订单结果
tokenAAmount=双方最小TokenA
tokenBAmount=按卖单价计算TokenB
tokenBRefund=买单返还TokenB

return TokenA成交量,TokenB成交量,买单返还,err
*/
//func CalcSwapTokenOrderAmount(bTokenAAmount, bTokenBAmount, sTokenAAmount, sTokenBAmount *big.Int) (*big.Int, *big.Int, *big.Int, error) {
//	tokenAAmount := new(big.Int).Set(bTokenAAmount)
//	if sTokenAAmount.Cmp(tokenAAmount) == -1 {
//		tokenAAmount = new(big.Int).Set(sTokenAAmount)
//	}
//
//	if tokenAAmount.Cmp(big.NewInt(0)) <= 0 {
//		return nil, nil, nil, config.ERROR_token_swap_tx_amount_zero
//	}
//
//	//买单>=卖单
//	if new(big.Int).Mul(bTokenBAmount, sTokenAAmount).Cmp(new(big.Int).Mul(sTokenBAmount, bTokenAAmount)) < 0 {
//		return nil, nil, nil, config.ERROR_token_swap_tx_invalid
//	}
//
//	tokenBAmount := big.NewInt(0)
//	tokenBAmount.Mul(tokenAAmount, sTokenBAmount)
//	tokenBAmount.Div(tokenBAmount, sTokenAAmount)
//
//	if tokenBAmount.Cmp(big.NewInt(0)) <= 0 {
//		return nil, nil, nil, config.ERROR_token_swap_tx_amount_zero
//	}
//
//	tmpTokenBAmount := big.NewInt(0)
//	tmpTokenBAmount.Mul(tokenAAmount, bTokenBAmount)
//	tmpTokenBAmount.Div(tmpTokenBAmount, bTokenAAmount)
//
//	tokenBRefund := big.NewInt(0).Sub(tmpTokenBAmount, tokenBAmount)
//
//	return tokenAAmount, tokenBAmount, tokenBRefund, nil
//}

// buyPrice: true=买单是对手;false=卖单是对手
// order1是对手单,order2是吃单
func CalcSwapTokenOrderAmountV3(order1, order2 *go_protos.OrderInfo) (*big.Int, *big.Int, error) {
	if order1.Buy == order2.Buy {
		return nil, nil, errors.New("buy/sell type")
	}

	var buy *go_protos.OrderInfo
	var sell *go_protos.OrderInfo
	if order1.Buy {
		buy = order1
		sell = order2
	} else {
		buy = order2
		sell = order1
	}

	//买单>=卖单
	if !(buy.Price >= sell.Price) {
		return nil, nil, errors.New("price")
	}

	bTokenAAmount := new(big.Int).SetBytes(buy.PendingTokenAAmount)
	//bTokenBAmount := new(big.Int).SetBytes(buy.PendingTokenBAmount)
	sTokenAAmount := new(big.Int).SetBytes(sell.PendingTokenAAmount)
	//sTokenBAmount := new(big.Int).SetBytes(sell.PendingTokenBAmount)
	//验证订单余额
	if bTokenAAmount.Cmp(big.NewInt(0)) <= 0 ||
		//bTokenBAmount.Cmp(big.NewInt(0)) <= 0 ||
		sTokenAAmount.Cmp(big.NewInt(0)) <= 0 {
		//sTokenBAmount.Cmp(big.NewInt(0)) <= 0 {
		return nil, nil, errors.New("pending amount 0")
	}

	//TokenA最大成交量
	tokenAAmount := new(big.Int).Set(bTokenAAmount)
	if tokenAAmount.Cmp(sTokenAAmount) == 1 {
		tokenAAmount.Set(sTokenAAmount)
	}

	if tokenAAmount.Cmp(big.NewInt(0)) <= 0 {
		return nil, nil, errors.New("tokena 0")
	}

	//tokenBAmount=对手价*tokenAAmount
	tokenBAmountDec := decimal.NewFromBigInt(tokenAAmount, 0).Mul(decimal.NewFromFloat(order1.Price))
	tokenBAmount := tokenBAmountDec.BigInt()
	if tokenBAmount.Cmp(big.NewInt(0)) <= 0 {
		return nil, nil, errors.New("tokenb 0")
	}

	return tokenAAmount, tokenBAmount, nil
}
