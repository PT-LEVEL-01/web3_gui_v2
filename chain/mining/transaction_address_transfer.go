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

type TxAddressTransfer struct {
	TxBase

	PayAddress crypto.AddressCoin
}

type TxAddressTransferVO struct {
	TxBaseVO

	PayAddress string `json:"pay_address"` //支付地址
}

func init() {
	tpc := new(AddressTransferController)
	RegisterTransaction(config.Wallet_tx_type_address_transfer, tpc)
}

func (this *TxAddressTransfer) GetVOJSON() interface{} {
	tx := TxAddressTransferVO{
		TxBaseVO: this.TxBase.ConversionVO(),
	}
	if this.PayAddress != nil && len(this.PayAddress) > 0 {
		tx.PayAddress = this.PayAddress.B58String()
	}
	return tx
}

func (this *TxAddressTransfer) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_address_transfer)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

func (this *TxAddressTransfer) Proto() (*[]byte, error) {
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
	txPay := go_protos.TxAddressTransfer{
		TxBase:     &txBase,
		PayAddress: this.PayAddress,
	}
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *TxAddressTransfer) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(this.PayAddress)
	*bs = buf.Bytes()
	return bs
}

func (this *TxAddressTransfer) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	*signDst = append(*signDst, this.PayAddress...)
	sign := keystore.Sign(*key, *signDst)
	return &sign
}

func (this *TxAddressTransfer) CheckSign() error {
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
		*signDst = append(*signDst, this.PayAddress...)
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

func (this *TxAddressTransfer) GetSpend() uint64 {
	return this.Gas
}

func (this *TxAddressTransfer) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

func (this *TxAddressTransfer) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}
	var totalValue uint64

	for _, vout := range this.Vout {
		totalValue += vout.Value
		frozenMap, ok := itemCount.AddItems[utils.Bytes2string(vout.Address)]
		if ok {
			oldValue, ok := (*frozenMap)[vout.FrozenHeight]
			if ok {
				oldValue += int64(vout.Value)
				(*frozenMap)[vout.FrozenHeight] = oldValue
			} else {
				(*frozenMap)[vout.FrozenHeight] = int64(vout.Value)
			}
		} else {
			frozenMap := make(map[uint64]int64, 0)
			frozenMap[vout.FrozenHeight] = int64(vout.Value)
			itemCount.AddItems[utils.Bytes2string(vout.Address)] = &frozenMap
		}
	}

	//payAddress余额中减去。
	payFrom := this.PayAddress
	payFrozenMap, ok := itemCount.AddItems[utils.Bytes2string(payFrom)]
	if ok {
		oldValue, ok := (*payFrozenMap)[0]
		if ok {
			oldValue -= int64(totalValue)
			(*payFrozenMap)[0] = oldValue
		} else {
			(*payFrozenMap)[0] = (0 - int64(totalValue))
		}
	} else {
		payFrozenMap := make(map[uint64]int64, 0)
		payFrozenMap[0] = (0 - int64(totalValue))
		itemCount.AddItems[utils.Bytes2string(payFrom)] = &payFrozenMap
	}

	//from余额中减去gas
	from := this.Vin[0].GetPukToAddr()
	itemCount.Nonce[utils.Bytes2string(*from)] = this.Vin[0].Nonce
	frozenMap, ok := itemCount.AddItems[utils.Bytes2string(*from)]
	if ok {
		oldValue, ok := (*frozenMap)[0]
		if ok {
			oldValue -= int64(this.Gas)
			(*frozenMap)[0] = oldValue
		} else {
			(*frozenMap)[0] = (0 - int64(this.Gas))
		}
	} else {
		frozenMap := make(map[uint64]int64, 0)
		frozenMap[0] = (0 - int64(this.Gas))
		itemCount.AddItems[utils.Bytes2string(*from)] = &frozenMap
	}

	return &itemCount
}

func AddressTransfer(srcAddr, addr, payAddr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd string, comment string, domain string, domainType uint64) (*TxAddressTransfer, error) {
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

	txpay, err := BuildAddressTransfer(srcAddr, addr, payAddr, amount, gas, frozenHeight, pwd, comment, domain, domainType)
	unlock()
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}

	if txpay == nil {
		return nil, fmt.Errorf("tx exists")
	}
	txpay.BuildHash()

	if err := forks.GetLongChain().TransactionManager.AddTx(txpay); err != nil {
		return nil, errors.Wrap(err, "add tx fail!")
	}

	MulticastTx(txpay)

	return txpay, nil
}

func BuildAddressTransfer(srcAddress, address, payAddress *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd string, comment string, domain string, domainType uint64) (*TxAddressTransfer, error) { // paytotal++
	if !db.CheckAddressBind(*srcAddress, *payAddress) {
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
		Value:        amount, //输出金额 = 实际金额 * 100000000
		Address:      *address,
		FrozenHeight: frozenHeight,
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)
	var pay *TxAddressTransfer
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_address_transfer,          //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
		}
		pay = &TxAddressTransfer{
			TxBase:     base,
			PayAddress: *payAddress,
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

type AddressTransferController struct{}

func (this *AddressTransferController) Factory() interface{} {
	return new(TxAddressTransfer)
}

func (this *AddressTransferController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxAddressTransfer)
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

	tx := &TxAddressTransfer{
		TxBase:     txBase,
		PayAddress: txProto.PayAddress,
	}
	return tx, nil
}

/*
统计余额
将已经注册的域名保存到数据库
将自己注册的域名保存到内存
*/
func (this *AddressTransferController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {
}

func (this *AddressTransferController) CheckMultiplePayments(txItr TxItr) error {
	//engine.Log.Error("swap 多花 22222222222222222222222222222")
	tx := txItr.(*TxAddressTransfer)

	if !db.CheckAddressBind(*tx.Vin[0].GetPukToAddr(), tx.PayAddress) {
		return errors.New("bind address mismatch")
	}

	return nil
}

func (this *AddressTransferController) SyncCount() {
}

func (this *AddressTransferController) RollbackBalance() {
}

/*
注册域名交易，域名续费交易，修改域名的网络地址交易
@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *AddressTransferController) BuildTx(deposit *sync.Map, srcAddr, addr *crypto.AddressCoin,
	amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {
	return nil, nil
}
