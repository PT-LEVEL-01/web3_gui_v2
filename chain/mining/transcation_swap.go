package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"runtime"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

type Swap_Step int8

const (
	Swap_Step_Promoter Swap_Step = iota
	Swap_Step_Receiver
)

type TxSwap struct {
	TxBase

	TokenTxidOut      []byte   //卖掉token 如：ICOM
	TokenTxidIn       []byte   //购买token 如：USDT
	AmountOut         *big.Int //卖掉总量
	AmountIn          *big.Int //购买总量
	LockhightPromoter uint64   //挂单过期时间
	PukPromoter       []byte
	SignPromoter      []byte //发起者签名

	AmountDeal    *big.Int //成交总量，如：USDT
	TokenVinRecv  []*Vin   //吃单者签名
	TokenVoutRecv []*Vout  //吃单者将会得到的ICOM，通过（AmountOut，AmountIn）被动计算结果

	//TokenVinPromoter  []*Vin  //挂单者签名
	//TokenVoutPromoter []*Vout //挂单这将会得到的USDT，通过（AmountOut，AmountIn）被动计算结果
}

type TxSwap_VO struct {
	TxBaseVO

	TokenTxidOut      string `json:"token_txid_out"`     //卖掉token 如：ICOM
	TokenTxidIn       string `json:"token_txid_in"`      //购买token 如：USDT
	AmountOut         string `json:"amount_out"`         //卖掉总量
	AmountIn          string `json:"amount_in"`          //购买总量
	LockhightPromoter uint64 `json:"lockhight_promoter"` //挂单过期时间
	PukPromoter       string `json:"puk_promoter"`
	SignPromoter      string `json:"sign_promoter"` //发起者签名

	AmountDeal    string    `json:"amount_deal"`     //成交总量，如：USDT
	TokenVinRecv  []*VinVO  `json:"token_vin_recv"`  //吃单者签名
	TokenVoutRecv []*VoutVO `json:"token_vout_recv"` //吃单者将会得到的ICOM，通过（AmountOut，AmountIn）被动计算结果

	Surplus *string `json:"surplus"`
}

func init() {
	tpc := new(SwapTxController)
	RegisterTransaction(config.Wallet_tx_type_swap, tpc)
}

// 上链序列化
func (this *TxSwap) Serialize() *[]byte {
	return this.SerializeStep(Swap_Step_Receiver)
}

/*
获取锁定区块高度
*/
func (this *TxSwap) GetLockHeight() uint64 {
	return this.LockHeight
}

func (this *TxSwap) SerializeStep(step Swap_Step) *[]byte {

	bs := new([]byte)
	switch step {
	case Swap_Step_Promoter:
		*bs = make([]byte, 0)
		buf := bytes.NewBuffer(*bs)
		this.serializePromoter(buf)
		*bs = buf.Bytes()
	case Swap_Step_Receiver:
		bs = this.TxBase.Serialize()
		buf := bytes.NewBuffer(*bs)
		this.serializeReceiver(buf)
		*bs = buf.Bytes()
	}
	return bs
}

func (this *TxSwap) serializePromoter(buf *bytes.Buffer) {
	buf.Write(this.TokenTxidOut)
	buf.Write(this.TokenTxidIn)
	buf.Write(this.AmountOut.Bytes())
	buf.Write(this.AmountIn.Bytes())
	buf.Write(utils.Uint64ToBytes(this.LockhightPromoter))
	buf.Write(this.PukPromoter)
	buf.Write(this.SignPromoter)
}

func (this *TxSwap) serializeReceiver(buf *bytes.Buffer) {
	this.serializePromoter(buf)

	buf.Write(this.AmountDeal.Bytes())

	if this.TokenVinRecv != nil {
		for _, one := range this.TokenVinRecv {
			bs := one.SerializeVin()
			buf.Write(*bs)
		}
	}

	if this.TokenVoutRecv != nil {
		for _, one := range this.TokenVoutRecv {
			bs := one.Serialize()
			buf.Write(*bs)
		}
	}
}

// 上链签名
func (this *TxSwap) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	return this.GetSignStep(key, vinIndex, Swap_Step_Receiver)
}

func (this *TxSwap) GetSignStep(key *ed25519.PrivateKey, vinIndex uint64, step Swap_Step) *[]byte {

	signDst := new([]byte)
	switch step {
	case Swap_Step_Promoter:
		this.getSignPromoter(signDst)
	case Swap_Step_Receiver:
		signDst = this.GetSignSerialize(nil, vinIndex)
		this.getSignReceiver(signDst)
	}

	sign := keystore.Sign(*key, *signDst)
	return &sign
}

func (this *TxSwap) getSignPromoter(signDst *[]byte) {
	*signDst = append(*signDst, this.TokenTxidOut...)
	*signDst = append(*signDst, this.TokenTxidIn...)
	*signDst = append(*signDst, this.AmountOut.Bytes()...)
	*signDst = append(*signDst, this.AmountIn.Bytes()...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.LockhightPromoter)...)
	*signDst = append(*signDst, this.PukPromoter...)
}

func (this *TxSwap) getSignReceiver(signDst *[]byte) {
	this.getSignPromoter(signDst)

	//发起人签名
	*signDst = append(*signDst, this.SignPromoter...)
	*signDst = append(*signDst, this.AmountDeal.Bytes()...)

	if len(this.TokenVoutRecv) != 0 {
		for _, one := range this.TokenVoutRecv {
			*signDst = append(*signDst, *one.Serialize()...)
		}
	}
}

// 上链验证签名
func (this *TxSwap) CheckSign() error {
	return this.CheckSignStep(Swap_Step_Receiver)
}

func (this *TxSwap) CheckSignStep(step Swap_Step) error {
	switch step {
	case Swap_Step_Promoter:
		return this.checkSignPromoter()
	case Swap_Step_Receiver:
		if len(this.Vin) != 1 {
			return config.ERROR_pay_vin_too_much
		}
		return this.checkSignReceiver()
	}

	return nil
}

// 签名1验签
func (this *TxSwap) checkSignPromoter() error {
	//sign := this.GetSignSerialize(nil, uint64(i))
	sign := new([]byte)
	this.getSignPromoter(sign)

	puk := ed25519.PublicKey(this.PukPromoter)
	if config.Wallet_print_serialize_hex {
		engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
	}
	if !ed25519.Verify(puk, *sign, this.SignPromoter) {
		return config.ERROR_sign_fail
	}
	return nil
}

// 签名2验签
func (this *TxSwap) checkSignReceiver() error {
	if len(this.Vin) != 1 && len(this.TokenVinRecv) != 1 {
		return config.ERROR_pay_vin_too_much
	}

	if err := this.checkSignPromoter(); err != nil {
		return err
	}

	sign := this.GetSignSerialize(nil, 0)
	this.getSignReceiver(sign)

	puk := ed25519.PublicKey(this.Vin[0].Puk)
	if config.Wallet_print_serialize_hex {
		engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
	}

	if !ed25519.Verify(puk, *sign, this.Vin[0].Sign) {
		return config.ERROR_sign_fail
	}
	return nil
}

/*
构建hash值得到交易id
*/
func (this *TxSwap) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_swap)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
构建hash值得到交易id
*/
func (this *TxSwap) BuildHashPromoter() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.SerializeStep(Swap_Step_Promoter)

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_swap)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

func (this *TxSwap) buildHashPromoter() []byte {
	bs := this.SerializeStep(Swap_Step_Promoter)

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_swap)

	return append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
构建hash值得到交易id
*/
func (this *TxSwap) BuildHashReceiver() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.SerializeStep(Swap_Step_Receiver)

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_swap)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
获取本交易总共花费的余额
*/
func (this *TxSwap) GetSpend() uint64 {
	return this.Gas
}

/*
检查重复的交易
*/
func (this *TxSwap) CheckRepeatedTx(txs ...TxItr) bool {
	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_swap {
			continue
		}
	}
	return true
}

/*
用于地址和txid格式化显示
*/
func (this *TxSwap) GetVOJSON() interface{} {
	recvVins := make([]*VinVO, 0)
	for _, one := range this.TokenVinRecv {
		recvVins = append(recvVins, one.ConversionVO())
	}
	recvVouts := make([]*VoutVO, 0)
	for _, one := range this.TokenVoutRecv {
		recvVouts = append(recvVouts, one.ConversionVO())
	}

	return TxSwap_VO{
		TxBaseVO:          this.TxBase.ConversionVO(),
		TokenTxidOut:      hex.EncodeToString(this.TokenTxidOut),
		TokenTxidIn:       hex.EncodeToString(this.TokenTxidIn),
		AmountOut:         this.AmountOut.String(),
		AmountIn:          this.AmountIn.String(),
		LockhightPromoter: this.LockhightPromoter,
		PukPromoter:       hex.EncodeToString(this.PukPromoter),
		SignPromoter:      hex.EncodeToString(this.SignPromoter),
		AmountDeal:        this.AmountDeal.String(),
		TokenVinRecv:      recvVins,
		TokenVoutRecv:     recvVouts,
	}
}

/*
统计交易余额
*/
func (this *TxSwap) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}

	promoter := crypto.BuildAddr(config.AddrPre, this.PukPromoter)
	receiver := this.Vin[0].GetPukToAddr()

	//交易中icom余额变动
	if len(this.TokenTxidOut) == 0 {
		//计算目标数量（token0）
		amountDst := new(big.Int).SetUint64(this.TokenVoutRecv[0].Value)
		frozenMapProm, ok := itemCount.AddItems[utils.Bytes2string(promoter)]
		if ok {
			oldValue, ok := (*frozenMapProm)[0]
			if ok {
				oldValue -= amountDst.Int64()
				(*frozenMapProm)[0] = oldValue
			} else {
				(*frozenMapProm)[0] = 0 - amountDst.Int64()
			}
		} else {
			frozenMapProm := make(map[uint64]int64, 0)
			frozenMapProm[0] = 0 - amountDst.Int64()
			itemCount.AddItems[utils.Bytes2string(promoter)] = &frozenMapProm
		}

		frozenMapRecv, ok := itemCount.AddItems[utils.Bytes2string(*receiver)]
		if ok {
			oldValue, ok := (*frozenMapRecv)[0]
			if ok {
				oldValue += amountDst.Int64()
				(*frozenMapRecv)[0] = oldValue
			} else {
				(*frozenMapRecv)[0] = amountDst.Int64()
			}
		} else {
			frozenMapRecv := make(map[uint64]int64, 0)
			frozenMapRecv[0] = amountDst.Int64()
			itemCount.AddItems[utils.Bytes2string(*receiver)] = &frozenMapRecv
		}
	}

	if len(this.TokenTxidIn) == 0 {
		frozenMapRecv, ok := itemCount.AddItems[utils.Bytes2string(*receiver)]
		if ok {
			oldValue, ok := (*frozenMapRecv)[0]
			if ok {
				oldValue -= this.AmountDeal.Int64()
				(*frozenMapRecv)[0] = oldValue
			} else {
				(*frozenMapRecv)[0] = 0 - this.AmountDeal.Int64()
			}
		} else {
			frozenMapRecv := make(map[uint64]int64, 0)
			frozenMapRecv[0] = 0 - this.AmountDeal.Int64()
			itemCount.AddItems[utils.Bytes2string(*receiver)] = &frozenMapRecv
		}

		frozenMapProm, ok := itemCount.AddItems[utils.Bytes2string(promoter)]
		if ok {
			oldValue, ok := (*frozenMapProm)[0]
			if ok {
				oldValue += this.AmountDeal.Int64()
				(*frozenMapProm)[0] = oldValue
			} else {
				(*frozenMapProm)[0] = this.AmountDeal.Int64()
			}
		} else {
			frozenMapProm := make(map[uint64]int64, 0)
			frozenMapProm[0] = this.AmountDeal.Int64()
			itemCount.AddItems[utils.Bytes2string(promoter)] = &frozenMapProm
		}
	}

	//余额中减去。
	itemCount.Nonce[utils.Bytes2string(*receiver)] = this.Vin[0].Nonce

	frozenMapRecv, ok := itemCount.AddItems[utils.Bytes2string(*receiver)]
	if ok {
		oldValue, ok := (*frozenMapRecv)[0]
		if ok {
			oldValue -= int64(this.Gas)
			(*frozenMapRecv)[0] = oldValue
		} else {
			(*frozenMapRecv)[0] = int64(this.Gas)
		}
	} else {
		frozenMapRecv := make(map[uint64]int64, 0)
		frozenMapRecv[0] = 0 - int64(this.Gas)
		itemCount.AddItems[utils.Bytes2string(*receiver)] = &frozenMapRecv
	}

	return &itemCount
}

/*
格式化成[]byte
*/
func (this *TxSwap) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vins = append(vins, &go_protos.Vin{
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
	tokenRecvVins := make([]*go_protos.Vin, 0)
	for _, one := range this.TokenVinRecv {
		tokenRecvVins = append(tokenRecvVins, &go_protos.Vin{
			Nonce: one.Nonce.Bytes(),
			Puk:   one.Puk,
			Sign:  one.Sign,
		})
	}
	tokenRecvVouts := make([]*go_protos.Vout, 0)
	for _, one := range this.TokenVoutRecv {
		tokenRecvVouts = append(tokenRecvVouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}

	amountDeal := new(big.Int)
	if this.AmountDeal != nil {
		amountDeal = amountDeal.Set(this.AmountDeal)
	}

	txPay := go_protos.TxSwap{
		TxBase:            &txBase,
		TokenTxidOut:      this.TokenTxidOut,
		TokenTxidIn:       this.TokenTxidIn,
		AmountOut:         this.AmountOut.Bytes(),
		AmountIn:          this.AmountIn.Bytes(),
		LockhightPromoter: this.LockhightPromoter,
		PukPromoter:       this.PukPromoter,
		SignPromoter:      this.SignPromoter,
		AmountDeal:        amountDeal.Bytes(),
		TokenVinRecv:      tokenRecvVins,
		TokenVoutRecv:     tokenRecvVouts,
		//TokenVinPromoter:  tokenPromoterVins,
		//TokenVoutPromoter: tokenPromoterVouts,
	}
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *TxSwap) MulticastTx() {
	utils.Go(func() {
		goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
		_, file, line, _ := runtime.Caller(0)
		engine.AddRuntime(file, line, goroutineId)
		defer engine.DelRuntime(file, line, goroutineId)
		bs, err := this.Proto()
		if err != nil {
			// engine.Log.Warn("交易json格式化错误，取消广播 %s", txItr.GetHashStr())
			return
		}
		Area.SendMulticastMsg(config.MSGID_multicast_swap_transaction, bs)
	}, nil)
}

func (this *TxSwap) CalcDstAmount() *big.Int {
	return CalcDstAmount(this.AmountOut, this.AmountIn, this.AmountDeal)
}

func CalcDstAmount(out, in, deal *big.Int) *big.Int {
	amountDst := new(big.Int).Mul(out, deal)
	return amountDst.Div(amountDst, in)
}

func SwapTxPromoter(srcAddr *crypto.AddressCoin, tokenTxidOut, tokenTxidIn []byte, amountOut, amountIn *big.Int, lockhightPromoter uint64, pwd string) (*TxSwap, error) {
	unlock := createTxMutex.Lock(srcAddr.B58String())
	txpay, err := BuildSwapTxPromoter(srcAddr, tokenTxidOut, tokenTxidIn, amountOut, amountIn, lockhightPromoter, pwd)
	unlock()

	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	if txpay == nil {
		return nil, fmt.Errorf("hash exists")
	}
	txpay.BuildHashPromoter()

	//添加池子
	if err := forks.GetLongChain().transactionSwapManager.AddTx(txpay, Swap_Step_Promoter); err != nil {
		//GetLongChain().Balance.DelLockTx(txpay)
		return nil, errors.Wrap(err, "add tx fail!")
	}
	//广播
	txpay.MulticastTx()
	return txpay, nil
}

func BuildSwapTxPromoter(srcAddr *crypto.AddressCoin, tokenTxidOut, tokenTxidIn []byte, amountOut, amountIn *big.Int, lockhightPromoter uint64, pwd string) (*TxSwap, error) {
	vins := make([]*Vin, 0)
	puk, ok := config.Area.Keystore.GetPukByAddr(*srcAddr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}

	//构建交易输出
	vouts := make([]*Vout, 0)

	chain := GetLongChain()
	_, block := chain.GetLastBlock()

	var pay *TxSwap
	for i := uint64(0); i < 10000; i++ {
		base := TxBase{
			Type:       config.Wallet_tx_type_swap,
			Vin_total:  uint64(len(vins)),                //输入交易数量
			Vin:        vins,                             //交易输入
			Vout_total: uint64(len(vouts)),               //输出交易数量
			Vout:       vouts,                            //
			Gas:        0,                                //交易手续费
			LockHeight: block.Height + lockhightPromoter, //锁定高度
		}

		pay = &TxSwap{
			TxBase:            base,
			TokenTxidOut:      tokenTxidOut,
			TokenTxidIn:       tokenTxidIn,
			AmountOut:         amountOut,
			AmountIn:          amountIn,
			PukPromoter:       puk,
			LockhightPromoter: block.Height + lockhightPromoter,
		}
		_, prk, err := Area.Keystore.GetKeyByPuk(puk, pwd)
		if err != nil {
			return nil, err
		}
		sign := pay.GetSignStep(&prk, uint64(i), Swap_Step_Promoter)
		pay.SignPromoter = *sign

		pay.BuildHashPromoter()

		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}

	//chain.Balance.AddLockTx(pay)
	return pay, nil
}

func BuildSwapPromoterOfflineTx(key *keystore.Keystore, srcAddr *crypto.AddressCoin, tokenTxidOut, tokenTxidIn []byte, amountOut, amountIn *big.Int, lockhightPromoter uint64, pwd string) (*TxSwap, error) {

	vins := make([]*Vin, 0)
	puk, ok := key.GetPukByAddr(*srcAddr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}

	//构建交易输出
	vouts := make([]*Vout, 0)

	var pay *TxSwap
	for i := uint64(0); i < 10000; i++ {
		base := TxBase{
			Type:       config.Wallet_tx_type_swap,
			Vin_total:  uint64(len(vins)),  //输入交易数量
			Vin:        vins,               //交易输入
			Vout_total: uint64(len(vouts)), //输出交易数量
			Vout:       vouts,              //
			Gas:        0,                  //交易手续费
			LockHeight: lockhightPromoter,  //锁定高度
		}

		pay = &TxSwap{
			TxBase:            base,
			TokenTxidOut:      tokenTxidOut,
			TokenTxidIn:       tokenTxidIn,
			AmountOut:         amountOut,
			AmountIn:          amountIn,
			PukPromoter:       puk,
			LockhightPromoter: lockhightPromoter,
		}

		_, prk, err := key.GetKeyByPuk(puk, pwd)
		if err != nil {
			return nil, err
		}
		sign := pay.GetSignStep(&prk, uint64(i), Swap_Step_Promoter)
		pay.SignPromoter = *sign

		pay.BuildHashPromoter()

		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	return pay, nil
}

func SwapTxReceiver(srcAddr *crypto.AddressCoin, tx *TxSwap, amountDeal, gas uint64, pwd string) (*TxSwap, error) {
	unlock := createTxMutex.Lock(srcAddr.B58String())
	//defer unlock()

	//验证未上链交易数量
	addrStr := utils.Bytes2string(*srcAddr)
	tsrItr, ok := forks.GetLongChain().TransactionManager.unpackedTransaction.addrs.Load(addrStr)
	if ok {
		tsr := tsrItr.(*TransactionsRatio)
		if tsr.TrsLen() >= config.Wallet_addr_tx_count_max {
			unlock()
			return nil, errors.New("unpackedTransaction too more")
		}
	}

	txpay, err := BuildSwapTxReceiver(srcAddr, tx, amountDeal, gas, pwd)
	unlock()
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}

	if txpay == nil {
		return nil, fmt.Errorf("tx exists")
	}
	txpay.BuildHash()

	//添加原交易池子
	if err := forks.GetLongChain().TransactionManager.AddTx(txpay); err != nil {
		//GetLongChain().Balance.DelLockTx(txpay)
		return nil, errors.Wrap(err, "add tx fail!")
	}

	MulticastTx(txpay)

	return txpay, nil
}

func BuildSwapTxReceiver(srcAddr *crypto.AddressCoin, tx *TxSwap, amountDeal, gas uint64, pwd string) (*TxSwap, error) {
	////先检查发布者的余额
	//addrPromoter := crypto.BuildAddr(config.AddrPre, tx.PukPromoter)
	//tokenTotalPromoter, _, _ := GetTokenNotSpendAndLockedBalance(tx.TokenTxidOut, addrPromoter)
	////先只考虑全量兑换的情况
	//if tokenTotalPromoter < tx.AmountOut.Uint64() {
	//	return nil, config.ERROR_token_not_enough
	//}
	//先检查发布者的余额
	swaptx := GetLongChain().GetTransactionSwapManager().GetSwapSwapTransaction(tx.Hash)
	if swaptx.Surplus.Uint64() == 0 {
		return nil, fmt.Errorf("Promoter %s", config.ERROR_token_not_enough.Error())
	}

	//检查一下锁定高度
	if swaptx.Tx.LockHeight <= GetLongChain().GetCurrentBlock() {
		return nil, fmt.Errorf("Receiver %s", config.ERROR_tx_frozenheight.Error())
	}

	chain := GetLongChain()

	//检查结单者的余额
	checkAmount := new(big.Int).SetUint64(amountDeal)
	if len(tx.TokenTxidIn) == 0 {
		checkAmount.Add(checkAmount, new(big.Int).SetUint64(gas))
	} else {
		total, _ := chain.GetBalance().BuildPayVinNew(srcAddr, gas)
		if total < gas {
			return nil, config.ERROR_not_enough
		}

	}
	if _, err := getAndCheckNotSpendAndLockedBalance(tx.TokenTxidIn, *srcAddr, checkAmount); err != nil {
		return nil, fmt.Errorf("Receiver %s", err.Error())
	}

	//算一下目标金额是否足够
	amountDst := new(big.Int).Mul(tx.AmountOut, new(big.Int).SetUint64(amountDeal))
	amountDst = amountDst.Div(amountDst, tx.AmountIn)
	if swaptx.Surplus.Cmp(amountDst) < 0 {
		return nil, fmt.Errorf("Promoter Surplus %s", config.ERROR_token_not_enough.Error())
	}

	vins := make([]*Vin, 0)

	puk, ok := config.Area.Keystore.GetPukByAddr(*srcAddr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(srcAddr)
	vin := Vin{
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)

	tokenVinsRecv := make([]*Vin, 0)
	tokenvin := vin
	tokenVinsRecv = append(tokenVinsRecv, &tokenvin)

	//构建交易输出
	tokenVoutsRecv := make([]*Vout, 0)
	//转账token给目标地址

	tokenVout := Vout{
		Value: amountDst.Uint64(), //输出金额 = 实际金额 * 100000000
	}
	tokenVoutsRecv = append(tokenVoutsRecv, &tokenVout)

	var pay *TxSwap
	for i := uint64(0); i < 10000; i++ {
		base := TxBase{
			Type:       config.Wallet_tx_type_swap,
			Vin_total:  uint64(len(vins)),  //输入交易数量
			Vin:        vins,               //交易输入
			Vout_total: uint64(len(vouts)), //输出交易数量
			Vout:       vouts,              //
			Gas:        gas,                //交易手续费
			LockHeight: tx.LockHeight + i,  //锁定高度
		}

		pay = &TxSwap{
			TxBase:            base,
			TokenTxidOut:      tx.TokenTxidOut,
			TokenTxidIn:       tx.TokenTxidIn,
			AmountOut:         tx.AmountOut,
			AmountIn:          tx.AmountIn,
			PukPromoter:       tx.PukPromoter,
			SignPromoter:      tx.SignPromoter,
			LockhightPromoter: tx.LockhightPromoter,
			AmountDeal:        new(big.Int).SetUint64(amountDeal),
			TokenVinRecv:      tokenVinsRecv,
			TokenVoutRecv:     tokenVoutsRecv,
		}

		for i, one := range pay.Vin {
			_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			sign := pay.GetSignStep(&prk, uint64(i), Swap_Step_Receiver)
			pay.Vin[i].Sign = *sign
			pay.TokenVinRecv[i].Sign = *sign
		}
		pay.BuildHashReceiver()

		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}

	chain.Balance.AddLockTx(pay)
	return pay, nil
}

func BuildSwapTxReceiverOffline(key *keystore.Keystore, srcAddr *crypto.AddressCoin, tx *TxSwap, amountDeal, gas, lockhightReceiver uint64, pwd string) (*TxSwap, error) {

	vins := make([]*Vin, 0)
	puk, ok := key.GetPukByAddr(*srcAddr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	vin := Vin{
		Puk:   puk, //公钥
		Nonce: *new(big.Int).SetUint64(1),
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)

	tokenVinsRecv := make([]*Vin, 0)
	tokenvin := vin
	tokenVinsRecv = append(tokenVinsRecv, &tokenvin)

	//构建交易输出
	tokenVoutsRecv := make([]*Vout, 0)
	//转账token给目标地址
	amountDst := new(big.Int).Mul(tx.AmountOut, new(big.Int).SetUint64(amountDeal))
	amountDst = amountDst.Div(amountDst, tx.AmountIn)

	tokenVout := Vout{
		Value: amountDst.Uint64(), //输出金额 = 实际金额 * 100000000
	}
	tokenVoutsRecv = append(tokenVoutsRecv, &tokenVout)

	var pay *TxSwap
	for i := uint64(0); i < 10000; i++ {
		base := TxBase{
			Type:       config.Wallet_tx_type_swap,
			Vin_total:  uint64(len(vins)),  //输入交易数量
			Vin:        vins,               //交易输入
			Vout_total: uint64(len(vouts)), //输出交易数量
			Vout:       vouts,              //
			Gas:        0,                  //交易手续费
			LockHeight: lockhightReceiver,  //锁定高度
		}

		pay = &TxSwap{
			TxBase:            base,
			TokenTxidOut:      tx.TokenTxidOut,
			TokenTxidIn:       tx.TokenTxidIn,
			AmountOut:         tx.AmountOut,
			AmountIn:          tx.AmountIn,
			PukPromoter:       tx.PukPromoter,
			SignPromoter:      tx.SignPromoter,
			LockhightPromoter: tx.LockhightPromoter,
			AmountDeal:        new(big.Int).SetUint64(amountDeal),
			TokenVinRecv:      tokenVinsRecv,
			TokenVoutRecv:     tokenVoutsRecv,
		}

		for i, one := range pay.Vin {
			_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			sign := pay.GetSignStep(&prk, uint64(i), Swap_Step_Receiver)
			pay.Vin[i].Sign = *sign
			pay.TokenVinRecv[i].Sign = *sign
		}

		pay.BuildHashReceiver()

		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}

	//chain.GetBalance().AddLockTx(pay)
	return pay, nil
}

func ParseSwapTxProto(bs *[]byte) (*TxSwap, error) {
	if bs == nil {
		return nil, nil
	}

	txProto := new(go_protos.TxSwap)
	err := proto.Unmarshal(*bs, txProto)
	if err != nil {
		return nil, err
	}

	txBase := TxBase{}
	txBase.Hash = txProto.TxBase.Hash
	txBase.LockHeight = txProto.TxBase.LockHeight
	txBase.Type = txProto.TxBase.Type
	tx := &TxSwap{
		TxBase:            txBase,
		TokenTxidOut:      txProto.TokenTxidOut,
		TokenTxidIn:       txProto.TokenTxidIn,
		AmountOut:         new(big.Int).SetBytes(txProto.AmountOut),
		AmountIn:          new(big.Int).SetBytes(txProto.AmountIn),
		LockhightPromoter: txProto.LockhightPromoter,
		PukPromoter:       txProto.PukPromoter,
		SignPromoter:      txProto.SignPromoter,
	}

	return tx, nil
}

type SwapTxController struct{}

func (this *SwapTxController) Factory() interface{} {
	return new(TxSwap)
}

func (this *SwapTxController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxSwap)
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

	//txSwap

	tokenVinRecv := make([]*Vin, 0)
	for _, one := range txProto.TokenVinRecv {
		nonce := new(big.Int).SetBytes(one.Nonce)
		tokenVinRecv = append(tokenVinRecv, &Vin{
			Puk:   one.Puk,
			Sign:  one.Sign,
			Nonce: *nonce,
		})
	}

	tokenVoutRecv := make([]*Vout, 0)
	for _, one := range txProto.TokenVoutRecv {
		tokenVoutRecv = append(tokenVoutRecv, &Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	tx := &TxSwap{
		TxBase:            txBase,
		TokenTxidOut:      txProto.TokenTxidOut,
		TokenTxidIn:       txProto.TokenTxidIn,
		AmountOut:         new(big.Int).SetBytes(txProto.AmountOut),
		AmountIn:          new(big.Int).SetBytes(txProto.AmountIn),
		LockhightPromoter: txProto.LockhightPromoter,
		PukPromoter:       txProto.PukPromoter,
		SignPromoter:      txProto.SignPromoter,
		AmountDeal:        new(big.Int).SetBytes(txProto.AmountDeal),
		TokenVinRecv:      tokenVinRecv,
		TokenVoutRecv:     tokenVoutRecv,
	}
	return tx, nil
}

/*
统计余额
将已经注册的域名保存到数据库
将自己注册的域名保存到内存
*/
func (this *SwapTxController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {
	//engine.Log.Error("swap 统计 1111111111111111111111111111")

	txItemCounts := make(map[string]*map[string]*big.Int, 0)

	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_swap {
			continue
		}

		txToken := txItr.(*TxSwap)

		promoter := crypto.BuildAddr(config.AddrPre, txToken.PukPromoter)
		receiver := txToken.Vin[0].GetPukToAddr()

		if len(txToken.TokenTxidOut) != 0 {
			token0 := utils.Bytes2string(txToken.TokenTxidOut)
			token0Map, ok := txItemCounts[token0]
			if !ok {
				newm := make(map[string]*big.Int, 0)
				token0Map = &newm
				txItemCounts[token0] = &newm
			}

			//计算目标数量（token0）
			amountDst := txToken.CalcDstAmount()

			//receiver 添加 token0
			value0Recv, ok := (*token0Map)[utils.Bytes2string(*receiver)]
			if !ok {
				value0Recv = new(big.Int)
			}
			value0Recv.Add(value0Recv, amountDst)
			(*token0Map)[utils.Bytes2string(*receiver)] = value0Recv

			//promoter 扣掉 token0
			value0Prom, ok := (*token0Map)[utils.Bytes2string(promoter)]
			if !ok {
				value0Prom = new(big.Int)
			}
			value0Prom.Sub(value0Prom, amountDst)
			(*token0Map)[utils.Bytes2string(promoter)] = value0Prom
		}

		if len(txToken.TokenTxidIn) != 0 {
			token1 := utils.Bytes2string(txToken.TokenTxidIn)
			token1Map, ok := txItemCounts[token1]
			if !ok {
				newm := make(map[string]*big.Int, 0)
				token1Map = &newm
				txItemCounts[token1] = &newm
			}

			//receiver 扣掉 token1
			value1Recv, ok := (*token1Map)[utils.Bytes2string(*receiver)]
			if !ok {
				value1Recv = new(big.Int)
			}
			value1Recv.Sub(value1Recv, txToken.AmountDeal)
			(*token1Map)[utils.Bytes2string(*receiver)] = value1Recv

			//promoter 添加 token1
			value1Prom, ok := (*token1Map)[utils.Bytes2string(promoter)]
			if !ok {
				value1Prom = new(big.Int)
			}
			value1Prom.Add(value1Prom, txToken.AmountDeal)
			(*token1Map)[utils.Bytes2string(promoter)] = value1Prom
		}

		//更新swap池状态
		forks.GetLongChain().transactionSwapManager.UpdateSwapPoolStatus(txToken.buildHashPromoter(), txToken)
	}

	//统计token的余额
	if len(txItemCounts) > 0 {
		CountToken(&txItemCounts)
	}
}

func (this *SwapTxController) CheckMultiplePayments(txItr TxItr) error {
	//engine.Log.Error("swap 多花 22222222222222222222222222222")
	txToken := txItr.(*TxSwap)

	//只检查挂单
	if len(txToken.TokenVinRecv) == 0 {
		//promoter
		vinAddrProm := crypto.BuildAddr(config.AddrPre, txToken.PukPromoter)
		//promoter 冻结token0
		frozenTokenProm := txToken.AmountOut

		//在未打包交易中获取该地址的token花费 promoter
		if _, err := getAndCheckNotSpendAndLockedBalance(txToken.TokenTxidOut, vinAddrProm, frozenTokenProm); err != nil {
			return err
		}

		return nil
	}

	//只检查吃单
	//receiver
	vinAddrRecv := crypto.AddressCoin{}
	for _, vin := range txToken.Vin {
		vinAddrRecv = *vin.GetPukToAddr()
	}

	//receiver 冻结token1
	frozenTokenRecv := txToken.AmountDeal

	//在未打包交易中获取该地址的token花费 receiver
	if _, err := getAndCheckNotSpendAndLockedBalance(txToken.TokenTxidIn, vinAddrRecv, frozenTokenRecv); err != nil {
		return err
	}

	return nil
}

func (this *SwapTxController) SyncCount() {
}

func (this *SwapTxController) RollbackBalance() {
}

/*
注册域名交易，域名续费交易，修改域名的网络地址交易
@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *SwapTxController) BuildTx(deposit *sync.Map, srcAddr, addr *crypto.AddressCoin,
	amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {
	return nil, nil
}

// 获取和验证可用余额
func getAndCheckNotSpendAndLockedBalance(token []byte, addr crypto.AddressCoin, amount *big.Int) (*big.Int, error) {
	if len(token) > 0 {
		notSpendToken, _, _ := GetTokenNotSpendAndLockedBalance(token, addr)
		if amount.Cmp(new(big.Int).SetUint64(notSpendToken)) > 0 {
			return nil, config.ERROR_token_not_enough
		}
		return new(big.Int).SetUint64(notSpendToken), nil
	}

	totalAll, _ := GetLongChain().Balance.BuildPayVinNew(&addr, 0)
	if amount.Cmp(new(big.Int).SetUint64(totalAll)) > 0 {
		return nil, config.ERROR_not_enough
	}
	return new(big.Int).SetUint64(totalAll), nil
}
