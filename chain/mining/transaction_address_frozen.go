package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
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

const (
	TxAddressFrozenTypeFrozen uint64 = iota
	TxAddressFrozenTypeUnFrozen
)

type TxAddressFrozen struct {
	TxBase

	//0 frozen ,1 unfrozen
	FrozenType uint64
	FrozenAddr crypto.AddressCoin
}

type TxAddressFrozenVO struct {
	TxBaseVO

	//0 frozen ,1 unfrozen
	FrozenType uint64 `json:"frozen_type"`
	FrozenAddr string `json:"frozen_addr"`
}

func init() {
	tpc := new(TxAddressFrozenController)
	RegisterTransaction(config.Wallet_tx_type_address_frozen, tpc)
}

func (this *TxAddressFrozen) GetVOJSON() interface{} {
	tx := TxAddressFrozenVO{
		TxBaseVO:   this.TxBase.ConversionVO(),
		FrozenType: this.FrozenType,
	}
	if this.FrozenAddr != nil && len(this.FrozenAddr) > 0 {
		tx.FrozenAddr = this.FrozenAddr.B58String()
	}
	return tx
}

func (this *TxAddressFrozen) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_address_frozen)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

func (this *TxAddressFrozen) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vins = append(vins, &go_protos.Vin{
			Puk:   one.Puk,
			Sign:  one.Sign,
			Nonce: one.Nonce.Bytes(),
		})
	}
	vouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
			Domain:       one.Domain,
			DomainType:   one.DomainType,
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
	txPay := go_protos.TxAddressFrozen{
		TxBase:     &txBase,
		FrozenType: this.FrozenType,
		FrozenAddr: this.FrozenAddr,
	}
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *TxAddressFrozen) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(utils.Uint64ToBytes(this.FrozenType))
	buf.Write(this.FrozenAddr)
	*bs = buf.Bytes()
	return bs
}

func (this *TxAddressFrozen) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.FrozenType)...)
	*signDst = append(*signDst, this.FrozenAddr...)
	sign := keystore.Sign(*key, *signDst)
	return &sign
}

func (this *TxAddressFrozen) CheckSign() error {
	if this.Vin == nil || len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) == 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}

	for i, one := range this.Vin {
		signDst := this.GetSignSerialize(nil, uint64(i))
		//本交易特有信息
		*signDst = append(*signDst, utils.Uint64ToBytes(this.FrozenType)...)
		*signDst = append(*signDst, this.FrozenAddr...)
		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*signDst))
		}
		if !ed25519.Verify(puk, *signDst, one.Sign) {
			return config.ERROR_sign_fail
		}
	}
	return nil
}

func (this *TxAddressFrozen) GetSpend() uint64 {
	return this.Gas
}

func (this *TxAddressFrozen) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

func (this *TxAddressFrozen) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}
	totalValue := this.Gas
	//余额中减去。
	from := this.Vin[0].GetPukToAddr()
	itemCount.Nonce[utils.Bytes2string(*from)] = this.Vin[0].Nonce
	frozenMap := make(map[uint64]int64, 0)
	frozenMap[0] = (0 - int64(totalValue))
	itemCount.AddItems[utils.Bytes2string(*from)] = &frozenMap

	//持久化冻结/解冻
	if err := this.storeFrozen(); err != nil {
		engine.Log.Error("记录保存地址绑定失败：%s", err.Error())
	}

	return &itemCount
}

func (tm *TxAddressFrozen) checkFrozenStatus() error {
	if !db.CheckAddressBind(*tm.Vin[0].GetPukToAddr(), tm.FrozenAddr) {
		return errors.New("bind address mismatch")
	}

	switch tm.FrozenType {
	case TxAddressFrozenTypeFrozen:
		if CheckAddressFrozenStatus(tm.FrozenAddr) {
			return errors.New("already frozen")
		}
	case TxAddressFrozenTypeUnFrozen:
		if !CheckAddressFrozenStatus(tm.FrozenAddr) {
			return errors.New("already unfrozen")
		}
	}

	return nil
}

func (tm *TxAddressFrozen) storeFrozen() error {
	switch tm.FrozenType {
	case TxAddressFrozenTypeFrozen:
		if err := db.SaveAddressFrozen(tm.FrozenAddr); err != nil {
			engine.Log.Error("保存地址冻结失败：%s %s", tm.FrozenAddr.B58String(), err.Error())
			return err
		}
	case TxAddressFrozenTypeUnFrozen:
		if err := db.SaveAddressUnFrozen(tm.FrozenAddr); err != nil {
			engine.Log.Error("保存地址解冻失败：%s %s", tm.FrozenAddr.B58String(), err.Error())
			return err
		}
	}
	return nil
}

func AddressFrozen(srcAddr, frozenAddr *crypto.AddressCoin, frozenType, gas, frozenHeight uint64, pwd string, comment string, domain string, domainType uint64) (*TxAddressFrozen, error) {
	unlock := createTxMutex.Lock(srcAddr.B58String())

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

	txpay, err := BuildAddressFrozen(srcAddr, frozenAddr, frozenType, gas, frozenHeight, pwd, comment, domain, domainType)
	unlock()
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}

	if txpay == nil {
		return nil, fmt.Errorf("tx exists")
	}
	txpay.BuildHash()

	if txCtrl := GetTransactionCtrl(txpay.Class()); txCtrl != nil {
		if err = txCtrl.CheckMultiplePayments(txpay); err != nil {
			DelCacheTxAndLimit(hex.EncodeToString(*txpay.GetHash()), txpay)
			forks.GetLongChain().Balance.DelLockTx(txpay)
			return nil, err
		}
	}

	if err := forks.GetLongChain().TransactionManager.AddTx(txpay); err != nil {
		return nil, errors.Wrap(err, "add tx fail!")
	}

	MulticastTx(txpay)

	return txpay, nil
}

func BuildAddressFrozen(srcAddress, frozenAddress *crypto.AddressCoin, frozenType, gas, frozenHeight uint64, pwd string, comment string, domain string, domainType uint64) (*TxAddressFrozen, error) { // paytotal++
	if !db.CheckAddressBind(*srcAddress, *frozenAddress) {
		return nil, errors.New("bind address mismatch")
	}

	//commentbs := []byte{}
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	chain := forks.GetLongChain()

	currentHeight := chain.GetCurrentBlock()

	//查找余额
	vins := make([]*Vin, 0)

	total, item := chain.Balance.BuildPayVinNew(srcAddress, gas)

	if total < gas {
		return nil, config.ERROR_not_enough
	}

	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        0, //输出金额 = 实际金额 * 100000000
		Address:      *frozenAddress,
		FrozenHeight: frozenHeight,
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)
	var pay *TxAddressFrozen
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_address_frozen,            //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
		}
		pay = &TxAddressFrozen{
			TxBase:     base,
			FrozenType: frozenType,
			FrozenAddr: *frozenAddress,
		}

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			sign := pay.GetSign(&prk, uint64(i))
			pay.Vin[i].Sign = *sign
		}

		pay.BuildHash()
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

type TxAddressFrozenController struct{}

func (this *TxAddressFrozenController) Factory() interface{} {
	return new(TxAddressTransfer)
}

func (this *TxAddressFrozenController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxAddressFrozen)
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
			Domain:       one.Domain,
			DomainType:   one.DomainType,
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

	tx := &TxAddressFrozen{
		TxBase:     txBase,
		FrozenType: txProto.FrozenType,
		FrozenAddr: txProto.FrozenAddr,
	}
	return tx, nil
}

/*
统计余额
将已经注册的域名保存到数据库
将自己注册的域名保存到内存
*/
func (this *TxAddressFrozenController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {
}

func (this *TxAddressFrozenController) CheckMultiplePayments(txItr TxItr) error {
	//验证冻结和解冻
	tx := txItr.(*TxAddressFrozen)

	return tx.checkFrozenStatus()
}

func (this *TxAddressFrozenController) SyncCount() {
}

func (this *TxAddressFrozenController) RollbackBalance() {
}

/*
注册域名交易，域名续费交易，修改域名的网络地址交易
@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *TxAddressFrozenController) BuildTx(deposit *sync.Map, srcAddr, addr *crypto.AddressCoin,
	amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {
	return nil, nil
}

// 查询地址是否被冻结
func CheckAddressFrozenStatus(addr crypto.AddressCoin) bool {
	return db.AddressIsFrozen(addr)
}
