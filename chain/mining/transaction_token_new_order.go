package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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
	tpc := new(TokenNewOrderController)
	RegisterTransaction(config.Wallet_tx_type_token_new_order, tpc)
}

/*
token订单
*/
type TxTokenNewOrder struct {
	TxBase
	TokenAID     []byte   `json:"ta_id"` //代币ID
	TokenAAmount *big.Int `json:"ta_a"`  //代币金额
	TokenBID     []byte   `json:"tb_id"`
	TokenBAmount *big.Int `json:"tb_a"`
	TokenVin     *Vin     `json:"t_vin"` //持有者签名
	Price        float64  //价格
	Buy          bool     `json:"buy"` //买=true;卖=false
}

/*
token订单
*/
type TxTokenNewOrder_VO struct {
	TxBaseVO
	TokenAID     string  `json:"tokena_id"`     //代币ID
	TokenAAmount string  `json:"tokena_amount"` //代币金额
	TokenBID     string  `json:"tokenb_id"`
	TokenBAmount string  `json:"tokenb_amount"`
	TokenVin     *VinVO  `json:"token_vin"` //持有者签名
	Price        float64 //价格
	Buy          bool    `json:"buy"` //买=true;卖=false
}

/*
用于地址和txid格式化显示
*/
func (this *TxTokenNewOrder) GetVOJSON() interface{} {
	return TxTokenNewOrder_VO{
		TxBaseVO:     this.TxBase.ConversionVO(),
		TokenAID:     hex.EncodeToString(this.TokenAID),
		TokenAAmount: this.TokenAAmount.Text(10),
		TokenBID:     hex.EncodeToString(this.TokenBID),
		TokenBAmount: this.TokenBAmount.Text(10),
		Price:        this.Price,
		Buy:          this.Buy,
		TokenVin:     this.TokenVin.ConversionVO(),
	}
}

/*
构建hash值得到交易id
*/
func (this *TxTokenNewOrder) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_token_new_order)
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
func (this *TxTokenNewOrder) Proto() (*[]byte, error) {
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
	txTokenOrder := go_protos.TxTokenNewOrder{
		TxBase:       &txBase,
		TokenAID:     this.TokenAID,
		TokenAAmount: this.TokenAAmount.Bytes(),
		TokenBID:     this.TokenBID,
		TokenBAmount: this.TokenBAmount.Bytes(),
		TokenVin:     tokenVin,
		Price:        this.Price,
		Buy:          this.Buy,
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
func (this *TxTokenNewOrder) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)

	buf.Write(this.TokenAID)
	buf.Write(this.TokenAAmount.Bytes())
	buf.Write(this.TokenBID)
	buf.Write(this.TokenBAmount.Bytes())
	buf.Write(this.TokenVin.Puk)
	var b bytes.Buffer
	binary.Write(&b, binary.BigEndian, this.Price)
	buf.Write(b.Bytes())
	binary.Write(buf, binary.BigEndian, this.Buy)

	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *TxTokenNewOrder) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)

	sign := keystore.Sign(*key, *signDst)

	return &sign
}

/*
把相同地址的交易输出合并在一起
*/
func (this *TxTokenNewOrder) MergeTokenVout() {
}

/*
检查交易是否合法
*/
func (this *TxTokenNewOrder) CheckSign() error {
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

	//创建订单,验证TokenVin
	tokenVinBs := this.Serialize()
	tokenVinPuk := ed25519.PublicKey(this.TokenVin.Puk)
	if !ed25519.Verify(tokenVinPuk, *tokenVinBs, this.TokenVin.Sign) {
		return config.ERROR_sign_fail
	}

	return nil
}

/*
验证是否合法
*/
func (this *TxTokenNewOrder) GetWitness() *crypto.AddressCoin {
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
func (this *TxTokenNewOrder) GetSpend() uint64 {
	return this.Gas
}

/*
检查重复的交易
*/
func (this *TxTokenNewOrder) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }

	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_token_new_order {
			continue
		}
		// ta := one.(*TxTokenNewOrder)
		// if bytes.Equal(ta.Account, this.Account) {
		// 	return false
		// }
	}
	return true
}

/*
统计交易余额
*/
func (this *TxTokenNewOrder) CountTxItemsNew(height uint64) *TxItemCountMap {
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

//======================= TokenNewOrderController =======================

type TokenNewOrderController struct {
}

func (this *TokenNewOrderController) Factory() interface{} {
	return new(TxTokenNewOrder)
}

func (this *TokenNewOrderController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxTokenNewOrder)
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
	tx := &TxTokenNewOrder{
		TxBase:       txBase,
		TokenAID:     txProto.TokenAID,
		TokenAAmount: new(big.Int).SetBytes(txProto.TokenAAmount),
		TokenBID:     txProto.TokenBID,
		TokenBAmount: new(big.Int).SetBytes(txProto.TokenBAmount),
		Price:        txProto.Price,
		Buy:          txProto.Buy,
		TokenVin:     tokenVin,
	}
	return tx, nil
}

/*
统计新订单
*/
func (this *TokenNewOrderController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {
	//锁定金额集合
	lockedTokens := []*AddrToken{}
	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_token_new_order {
			continue
		}

		tokenOrder := txItr.(*TxTokenNewOrder)
		//处理新订单
		lockedToken := oBook.NewOrder(tokenOrder)
		lockedTokens = append(lockedTokens, lockedToken...)
	}

	// 处理Token新订单锁定金额和可用余额
	this.handleOrderToken(lockedTokens)
}

// 处理Token新订单锁定金额和可用余额
func (this *TokenNewOrderController) handleOrderToken(lockedTokens []*AddrToken) {
	for _, addrToken := range lockedTokens {
		lockKey := config.BuildDBKeyTokenAddrLockedValue(addrToken.TokenId, addrToken.Addr)

		val := new(big.Int)
		bs, err := db.LevelDB.Find(*lockKey)
		if err == nil {
			//更新金额
			val = val.SetBytes(*bs)
		}
		val.Add(val, addrToken.Value)

		if val.Cmp(big.NewInt(0)) <= 0 {
			db.LevelDB.Del(*lockKey)
		} else {
			data := val.Bytes()
			db.LevelDB.Save(*lockKey, &data)
		}
	}
}

func (this *TokenNewOrderController) CheckMultiplePayments(txItr TxItr) error {
	txToken := txItr.(*TxTokenNewOrder)

	srcAddr := txToken.Vin[0].PukToAddr
	addr := txToken.TokenVin.PukToAddr

	if txToken.Buy {
		notSpend := uint64(0)
		if txToken.TokenBID == nil || len(txToken.TokenBID) == 0 { //主链币
			notSpend, _, _ = GetBalanceForAddrSelf(addr)
			if bytes.Equal(srcAddr, addr) {
				notSpend += txToken.GetSpend()
			}
		} else {
			notSpend, _, _ = GetTokenNotSpendAndLockedBalance(txToken.TokenBID, addr)
		}
		if notSpend < txToken.TokenBAmount.Uint64() {
			return config.ERROR_token_not_enough
		}
	} else {
		notSpend := uint64(0)
		if txToken.TokenAID == nil || len(txToken.TokenAID) == 0 { //主链币
			notSpend, _, _ = GetBalanceForAddrSelf(addr)
			if bytes.Equal(srcAddr, addr) {
				notSpend += txToken.GetSpend()
			}
		} else {
			notSpend, _, _ = GetTokenNotSpendAndLockedBalance(txToken.TokenAID, addr)
		}
		if notSpend < txToken.TokenAAmount.Uint64() {
			return config.ERROR_token_not_enough
		}
	}

	return nil
}

func (this *TokenNewOrderController) SyncCount() {

}

func (this *TokenNewOrderController) RollbackBalance() {
	// return new(Tx_account)
}

// 释放Token锁定金额
func (this *TokenNewOrderController) handleOrderLockedTokenRelease(orderId []byte) {
}

type paramsCreateTokenOrder struct {
	TokenAID     []byte              `json:"ta_id"` //代币ID
	TokenAAmount uint64              `json:"ta_a"`  //代币金额
	TokenBID     []byte              `json:"tb_id"`
	TokenBAmount uint64              `json:"tb_a"`
	Buy          bool                `json:"buy"` //买=true;卖=false
	TokenAddr    *crypto.AddressCoin //代币交易地址
}

/*
构建Token创建订单交易
*/
func (this *TokenNewOrderController) BuildTx(deposit *sync.Map, srcAddr, addr *crypto.AddressCoin,
	amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {

	if len(params) < 1 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}

	tokenParams := params[0].(*paramsCreateTokenOrder)
	//==================== 构建新订单交易 =======================
	puk, ok := config.Area.Keystore.GetPukByAddr(*tokenParams.TokenAddr)
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

	//当Buy=true代表买入A，vin用于证明B资产。
	//当Buy=false代表卖出A，vin用于证明A资产。
	if tokenParams.Buy {
		notSpend := uint64(0)
		if tokenParams.TokenBID == nil || len(tokenParams.TokenBID) == 0 { //主链币
			notSpend, _, _ = GetBalanceForAddrSelf(*tokenVin.GetPukToAddr())
			if bytes.Equal(*srcAddr, *tokenVin.GetPukToAddr()) {
				notSpend += gas
			}
		} else {
			notSpend, _, _ = GetTokenNotSpendAndLockedBalance(tokenParams.TokenBID, *tokenVin.GetPukToAddr())
		}
		if notSpend < tokenParams.TokenBAmount {
			return nil, config.ERROR_token_not_enough
		}
	} else {
		notSpend := uint64(0)
		if tokenParams.TokenAID == nil || len(tokenParams.TokenAID) == 0 { //主链币
			notSpend, _, _ = GetBalanceForAddrSelf(*tokenVin.GetPukToAddr())
			if bytes.Equal(*srcAddr, *tokenVin.GetPukToAddr()) {
				notSpend += gas
			}
		} else {
			notSpend, _, _ = GetTokenNotSpendAndLockedBalance(tokenParams.TokenAID, *tokenVin.GetPukToAddr())
		}
		if notSpend < tokenParams.TokenAAmount {
			return nil, config.ERROR_token_not_enough
		}
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

	round := int32(2)
	if tokenInfo, err := FindTokenInfo(tokenParams.TokenBID); err == nil {
		round = int32(tokenInfo.TokenAccuracy)
	}
	priceDec := decimal.NewFromBigInt(new(big.Int).SetUint64(tokenParams.TokenBAmount), 0).Div(decimal.NewFromBigInt(new(big.Int).SetUint64(tokenParams.TokenAAmount), 0))
	price := priceDec.Round(round).InexactFloat64()

	_, block := chain.GetLastBlock()
	//var txtno *TxToken
	var txtno *TxTokenNewOrder
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_token_new_order,          //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(emptyVouts)),                        //输出交易数量
			Vout:       emptyVouts,                                     //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                      //
		}
		txtno = &TxTokenNewOrder{
			TxBase:       base,
			TokenAID:     tokenParams.TokenAID,
			TokenAAmount: new(big.Int).SetUint64(tokenParams.TokenAAmount),
			TokenBID:     tokenParams.TokenBID,
			TokenBAmount: new(big.Int).SetUint64(tokenParams.TokenBAmount),
			Price:        price,
			Buy:          tokenParams.Buy,
			TokenVin:     tokenVin,
		}

		//给输出签名，防篡改
		for i, one := range txtno.Vin {
			_, prk, err := config.Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

			sign := txtno.GetSign(&prk, uint64(i))
			txtno.Vin[i].Sign = *sign
		}

		//给输出签名，防篡改
		_, prk, err := config.Area.Keystore.GetKeyByPuk(txtno.TokenVin.Puk, pwd)
		if err != nil {
			return nil, err
		}

		// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
		//tokenVin签名
		tokenVinBs := txtno.Serialize()
		tokenVinSign := keystore.Sign(prk, *tokenVinBs)
		txtno.TokenVin.Sign = tokenVinSign

		txtno.BuildHash()
		if txtno.CheckHashExist() {
			txtno = nil
			continue
		} else {
			break
		}
	}

	//把txitem冻结起来
	// FrozenToken(txidBs, tokenTxItems, txtno)

	// chain.GetBalance().Frozen(items, txtno)
	chain.GetBalance().AddLockTx(txtno)
	return txtno, nil
}

func CreateTokenOrder(srcAddr *crypto.AddressCoin, tokenAddr *crypto.AddressCoin, tokenAID []byte, tokenAAmount uint64, tokenBID []byte, tokenBAmount uint64, buy bool, gas uint64, pwd string) (TxItr, error) {
	tokenParams := &paramsCreateTokenOrder{
		TokenAID:     tokenAID,
		TokenAAmount: tokenAAmount,
		TokenBID:     tokenBID,
		TokenBAmount: tokenBAmount,
		Buy:          buy,
		TokenAddr:    tokenAddr,
	}
	txItr, err := GetLongChain().GetBalance().BuildOtherTx(config.Wallet_tx_type_token_new_order, srcAddr,
		nil, 0, gas, 0, pwd, "", tokenParams)
	if err != nil {
		return nil, err
	}

	return txItr, err
}
