package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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
	tpc := new(TokenCancelOrderController)
	RegisterTransaction(config.Wallet_tx_type_token_cancel_order, tpc)
}

/*
token订单
*/
type TxTokenCancelOrder struct {
	TxBase
	TokenVin *Vin     `json:"t_vin"` //持有者签名
	OrderIDs [][]byte `json:"ids"`
}

/*
token订单
*/
type TxTokenCancelOrder_VO struct {
	TxBaseVO
	TokenVin *VinVO   `json:"token_vin"` //持有者签名
	OrderIDs []string `json:"ids"`
}

/*
用于地址和txid格式化显示
*/
func (this *TxTokenCancelOrder) GetVOJSON() interface{} {
	orderIDs := []string{}
	for _, id := range this.OrderIDs {
		orderIDs = append(orderIDs, hex.EncodeToString(id))
	}
	return TxTokenCancelOrder_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		TokenVin: this.TokenVin.ConversionVO(),
		OrderIDs: orderIDs,
	}
}

/*
构建hash值得到交易id
*/
func (this *TxTokenCancelOrder) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_token_cancel_order)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	对整个交易签名
*/
//func (this *Tx_vote_in) Sign(key *keystore.Address, pwd string) (*[]byte, error) {
//	bs := this.SignSerialize()
//	return key.Sign(*bs, pwd)
//}

/*
格式化成[]byte
*/
func (this *TxTokenCancelOrder) Proto() (*[]byte, error) {
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

	tokenVin := &go_protos.Vin{
		Nonce: this.TokenVin.Nonce.Bytes(),
		Puk:   this.TokenVin.Puk,
		Sign:  this.TokenVin.Sign,
	}

	txTokenOrder := go_protos.TxTokenCancelOrder{
		TxBase:   &txBase,
		TokenVin: tokenVin,
		OrderIds: this.OrderIDs,
	}

	bs, err := txTokenOrder.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
格式化成json字符串
*/
func (this *TxTokenCancelOrder) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)

	buf.Write(this.TokenVin.Puk)
	for _, id := range this.OrderIDs {
		buf.Write(id)
	}

	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *TxTokenCancelOrder) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)

	sign := keystore.Sign(*key, *signDst)

	return &sign

}

/*
把相同地址的交易输出合并在一起
*/
func (this *TxTokenCancelOrder) MergeTokenVout() {
}

/*
检查交易是否合法
*/
func (this *TxTokenCancelOrder) CheckSign() error {
	if len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) == 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}
	if this.Vout_total != 0 {
		return config.ERROR_pay_vout_too_much
	}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费;3.输入不能重复。
	// vinMap := make(map[string]int)
	// inTotal := uint64(0)
	for i, one := range this.Vin {
		sign := this.GetSignSerialize(nil, uint64(i))

		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		if !ed25519.Verify(puk, *sign, one.Sign) {
			return config.ERROR_sign_fail
		}
	}

	//取消订单,验证TokenVin 和 拥有者
	tokenVinBs := this.Serialize()
	tokenVinPuk := ed25519.PublicKey(this.TokenVin.Puk)
	if !ed25519.Verify(tokenVinPuk, *tokenVinBs, this.TokenVin.Sign) {
		return config.ERROR_sign_fail
	}

	key := config.BuildDBKeyTokenOrder(this.OrderIDs[0])
	if item, ok := oBook.Orders.Load(utils.Bytes2string(*key)); ok {
		if !bytes.Equal(item.(*go_protos.OrderInfo).Address, *this.TokenVin.GetPukToAddr()) {
			return config.ERROR_token_invalid_order_id
		}
	}

	return nil
}

/*
验证是否合法
*/
func (this *TxTokenCancelOrder) GetWitness() *crypto.AddressCoin {
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
func (this *TxTokenCancelOrder) GetSpend() uint64 {
	return this.Gas
}

/*
检查重复的交易
*/
func (this *TxTokenCancelOrder) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }

	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_token_cancel_order {
			continue
		}
		// ta := one.(*TxTokenCancelOrder)
		// if bytes.Equal(ta.Account, this.Account) {
		// 	return false
		// }
	}
	return true
}

/*
统计交易余额
*/
func (this *TxTokenCancelOrder) CountTxItemsNew(height uint64) *TxItemCountMap {
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

//======================= TokenCancelOrderController =======================

type TokenCancelOrderController struct {
}

func (this *TokenCancelOrderController) Factory() interface{} {
	return new(TxTokenCancelOrder)
}

func (this *TokenCancelOrderController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxTokenCancelOrder)
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

	tokenVin := &Vin{
		Puk:  txProto.TokenVin.Puk,
		Sign: txProto.TokenVin.Sign,
	}
	tx := &TxTokenCancelOrder{
		TxBase:   txBase,
		TokenVin: tokenVin,
		OrderIDs: txProto.OrderIds,
	}
	return tx, nil
}

/*
统计新订单
*/
func (this *TokenCancelOrderController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {
	//锁定金额集合
	unlockedTokens := []*AddrToken{}
	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_token_cancel_order {
			continue
		}

		tokenOrder := txItr.(*TxTokenCancelOrder)
		//处理新订单
		unlockedToken := oBook.CancelOrder(tokenOrder)
		unlockedTokens = append(unlockedTokens, unlockedToken...)
	}

	// 处理Token取消订单锁定金额
	this.handleOrderToken(unlockedTokens)
}

// 处理Token取消订单锁定金额
func (this *TokenCancelOrderController) handleOrderToken(unlockedTokens []*AddrToken) {
	for _, one := range unlockedTokens {
		lockKey := config.BuildDBKeyTokenAddrLockedValue(one.TokenId, one.Addr)
		val := new(big.Int)
		bs, err := db.LevelDB.Find(*lockKey)
		if err == nil {
			//更新金额
			val = val.SetBytes(*bs)
		}
		val.Sub(val, one.Value)

		if val.Cmp(big.NewInt(0)) <= 0 {
			db.LevelDB.Del(*lockKey)
		} else {
			data := val.Bytes()
			db.LevelDB.Save(*lockKey, &data)
		}
	}
}

func (this *TokenCancelOrderController) CheckMultiplePayments(txItr TxItr) error {
	//txToken := txItr.(*TxTokenCancelOrder)

	//vinAddr := crypto.AddressCoin{}
	//for _, vin := range txToken.Token_Vin {
	//	vinAddr = *vin.GetPukToAddr()
	//}

	//frozenToken := uint64(0)
	//for _, vout := range txToken.Token_Vout {
	//	frozenToken += vout.Value
	//}

	////在未打包交易中获取该地址的token花费
	//notSpendToken, lockToken := GetTokenNotSpendAndLockedBalance(txToken.Token_txid, vinAddr)
	//if notSpendToken < lockToken+frozenToken {
	//	return config.ERROR_token_not_enough
	//}

	return nil
}

func (this *TokenCancelOrderController) SyncCount() {

}

func (this *TokenCancelOrderController) RollbackBalance() {
	// return new(Tx_account)
}

// 释放Token锁定金额
func (this *TokenCancelOrderController) handleOrderLockedTokenRelease(orderId []byte) {
}

type paramsCancelTokenOrder struct {
	Address  *crypto.AddressCoin //代币交易地址
	OrderIDs [][]byte            //订单ID
}

/*
构建Token创建订单交易
*/
func (this *TokenCancelOrderController) BuildTx(deposit *sync.Map, srcAddr, addr *crypto.AddressCoin,
	amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {

	if len(params) < 1 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}

	tokenParams := params[0].(*paramsCancelTokenOrder)
	//==================== 构建取消订单交易 =======================
	for _, id := range tokenParams.OrderIDs {
		key := config.BuildDBKeyTokenOrder(id)
		if item, ok := oBook.Orders.Load(utils.Bytes2string(*key)); ok {
			if item.(*go_protos.OrderInfo).State != OrderStateNormal {
				return nil, config.ERROR_token_invalid_order_id
			}
		}
	}

	puk, ok := config.Area.Keystore.GetPukByAddr(*tokenParams.Address)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
	chain := GetLongChain()
	//nonce := chain.GetBalance().FindNonce(item.Addr)
	tokenVin := &Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk: puk, //公钥
		//					Sign: *sign,           //签名
	}

	//查找余额
	vins := make([]*Vin, 0)
	total, item := chain.GetBalance().BuildPayVinNew(srcAddr, gas)
	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}

	// for _, item := range items {
	puk, ok = config.Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk: puk, //公钥
		//					Sign: *sign,           //签名
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)
	// }

	//构建交易输出
	emptyVouts := make([]*Vout, 0)

	var commentbs []byte
	if comment != "" {
		commentbs = []byte(comment)
	}

	_, block := chain.GetLastBlock()
	//var txin *TxToken
	var txin *TxTokenCancelOrder
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_token_cancel_order,       //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(emptyVouts)),                        //输出交易数量
			Vout:       emptyVouts,                                     //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                      //
		}
		txin = &TxTokenCancelOrder{
			TxBase:   base,
			TokenVin: tokenVin,
			OrderIDs: tokenParams.OrderIDs,
		}

		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := config.Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

			sign := txin.GetSign(&prk, uint64(i))
			txin.Vin[i].Sign = *sign
		}

		//给输出签名，防篡改
		_, prk, err := config.Area.Keystore.GetKeyByPuk(txin.TokenVin.Puk, pwd)
		if err != nil {
			return nil, err
		}

		// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

		tokenVinBs := txin.Serialize()
		tokenVinSign := keystore.Sign(prk, *tokenVinBs)
		txin.TokenVin.Sign = tokenVinSign

		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}

	//把txitem冻结起来
	// FrozenToken(txidBs, tokenTxItems, txin)

	// chain.GetBalance().Frozen(items, txin)
	chain.GetBalance().AddLockTx(txin)
	return txin, nil
}

func CancelTokenOrder(srcAddr *crypto.AddressCoin, addr *crypto.AddressCoin, orderIds [][]byte, gas uint64, pwd string) (TxItr, error) {
	tokenParams := &paramsCancelTokenOrder{
		Address:  addr,
		OrderIDs: orderIds,
	}
	txItr, err := GetLongChain().GetBalance().BuildOtherTx(config.Wallet_tx_type_token_cancel_order, srcAddr,
		nil, 0, gas, 0, pwd, "", tokenParams)
	if err != nil {
		return nil, err
	}

	return txItr, err
}
