package mining

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
转账交易
*/
type Tx_Pay struct {
	TxBase
}

// 交易平均Gas
type TxAveGas struct {
	mux    *sync.RWMutex
	Index  uint64      //当前游标
	AllGas [100]uint64 //最新的100个gas费用
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_Pay) GetVOJSON() interface{} {
	return this.TxBase.ConversionVO()
}

/*
构建hash值得到交易id
*/
func (this *Tx_Pay) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}

	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_pay)
	// jsonBs, _ := this.Json()

	// fmt.Println("序列化输出 111", string(*jsonBs))
	// fmt.Println("序列化输出 222", len(*bs), hex.EncodeToString(*bs))
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
验证是否合法
*/
func (this *Tx_Pay) CheckSign() error {
	// fmt.Println("开始验证交易合法性 Tx_Pay")
	if len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) == 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}
	if this.Vout_total > config.Mining_pay_vout_max {
		return config.ERROR_pay_vout_too_much
	}
	if err := this.TxBase.CheckBase(); err != nil {
		return err
	}
	for _, vout := range this.Vout {
		if vout.Value <= 0 {
			return config.ERROR_amount_zero
		}
	}

	return nil
}

/*
获取本交易总共花费的余额
*/
func (this *Tx_Pay) GetSpend() uint64 {
	spend := this.Gas
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_Pay) CheckRepeatedTx(txs ...TxItr) bool {
	//判断是否出现双花
	// return this.MultipleExpenditures(txs...)
	return true
}

/*
	统计交易余额
*/
// func (this *Tx_Pay) CountTxItems(height uint64) *TxItemCount {
// 	itemCount := TxItemCount{
// 		Additems: make([]*TxItem, 0),
// 		SubItems: make([]*TxSubItems, 0),
// 	}
// 	//将之前的UTXO标记为已经使用，余额中减去。
// 	for _, vin := range this.Vin {
// 		// engine.Log.Info("查看vin中的状态 %d", vin.PukIsSelf)
// 		ok := vin.CheckIsSelf()
// 		if !ok {
// 			continue
// 		}
// 		// engine.Log.Info("统单易1耗时 %s %s", txItr.GetHashStr(), config.TimeNow().Sub(start))
// 		//查找这个地址的余额列表，没有则创建一个
// 		itemCount.SubItems = append(itemCount.SubItems, &TxSubItems{
// 			Txid:      vin.Txid, //utils.Bytes2string(vin.Txid), //  vin.GetTxidStr(),
// 			VoutIndex: vin.Vout,
// 			Addr:      *vin.GetPukToAddr(), // utils.Bytes2string(*vin.GetPukToAddr()), // vin.GetPukToAddrStr(),
// 		})
// 	}

// 	//生成新的UTXO收益，保存到列表中
// 	for voutIndex, vout := range this.Vout {
// 		// if voutIndex == 0 {
// 		// 	continue
// 		// }
// 		//找出需要统计余额的地址
// 		//和自己无关的地址
// 		ok := vout.CheckIsSelf()
// 		if !ok {
// 			continue
// 		}

// 		// engine.Log.Info("统单易5耗时 %s %s", txItr.GetHashStr(), config.TimeNow().Sub(start))
// 		txItem := TxItem{
// 			Addr: &(this.Vout)[voutIndex].Address, //  &vout.Address,
// 			// AddrStr: vout.GetAddrStr(),                      //
// 			Value: vout.Value,      //余额
// 			Txid:  *this.GetHash(), //交易id
// 			// TxidStr:      txHashStr,                              //
// 			VoutIndex:    uint64(voutIndex), //交易输出index，从0开始
// 			Height:       height,            //
// 			LockupHeight: vout.FrozenHeight, //锁仓高度
// 		}

// 		//计入余额列表
// 		// this.notspentBalance.AddTxItem(txItem)
// 		itemCount.Additems = append(itemCount.Additems, &txItem)

// 		//保存到缓存
// 		// engine.Log.Info("开始统计交易余额 区块高度 %d 保存到缓存", bhvo.BH.Height)
// 		// TxCache.AddTxInTxItem(txHashStr, txItr)
// 		TxCache.AddTxInTxItem(*this.GetHash(), this)

// 	}
// 	return &itemCount
// }

/*
统计交易余额
*/
func (this *Tx_Pay) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}
	totalValue := this.Gas
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

	//余额中减去。
	from := this.Vin[0].GetPukToAddr()
	itemCount.Nonce[utils.Bytes2string(*from)] = this.Vin[0].Nonce
	frozenMap, ok := itemCount.AddItems[utils.Bytes2string(*from)]
	if ok {
		oldValue, ok := (*frozenMap)[0]
		if ok {
			oldValue -= int64(totalValue)
			(*frozenMap)[0] = oldValue
		} else {
			(*frozenMap)[0] = (0 - int64(totalValue))
		}
	} else {
		frozenMap := make(map[uint64]int64, 0)
		frozenMap[0] = (0 - int64(totalValue))
		itemCount.AddItems[utils.Bytes2string(*from)] = &frozenMap
	}
	return &itemCount
}

// var paytotal = 0

/*
创建一个转款交易
*/
func CreateTxPay(srcAddress, address *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, domain string, domainType uint64) (*Tx_Pay, error) {
	// paytotal++
	//engine.Log.Info("start CreateTxPay")
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}
	// start := config.TimeNow()

	chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()

	currentHeight := chain.GetCurrentBlock()

	//engine.Log.Info("currentHeight:%d", currentHeight)

	// total = uint64(0)
	// if srcAddress == nil {
	// 	srcAddress, total = FindNotSpendBalance(amount + gas)
	// } else {
	// 	total = GetNotSpendBalance(srcAddress)
	// }

	//查找余额
	vins := make([]*Vin, 0)
	// total := uint64(0)

	total, item := chain.Balance.BuildPayVinNew(srcAddress, amount+gas)
	// chain.Balance.

	// engine.Log.Info("查找余额耗时 %s", config.TimeNow().Sub(start))

	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }

	// for _, item := range items {
	// engine.Log.Info("use item:%s_%d %d %d", hex.EncodeToString(item.Txid), item.VoutIndex, item.LockupHeight, item.Value)
	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)
	// if paytotal > 4 {
	// 	nonce = *new(big.Int).Add(&nonce, big.NewInt(1))
	// }
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	//engine.Log.Info("新交易nonce:%d", vin.Nonce.Uint64())
	vins = append(vins, &vin)
	// }

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)
	//找零
	// changeAddr := change
	// if changeAddr == nil || len(*changeAddr) <= 0 {
	// 	changeAddr = item.Addr
	// }
	// //TODO 将剩余款项转入新的地址，保证资金安全
	// if total > amount+gas {
	// 	vout := Vout{
	// 		Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
	// 		Address: *changeAddr,          //找零地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	// engine.Log.Info("构建输入输出 耗时 %s", config.TimeNow().Sub(start))

	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay,                       //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
			Comment:    []byte{},
		}
		pay = &Tx_Pay{
			TxBase: base,
		}

		// pay.MergeVout()

		// startTime := config.TimeNow()

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
			sign := pay.GetSign(&prk, uint64(i))
			pay.Vin[i].Sign = *sign
		}

		// engine.Log.Info("给输出签名 耗时 %d %s", i, config.TimeNow().Sub(startTime))

		pay.BuildHash()
		// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	//engine.Log.Info("新交易nonce:%d hash:%s", vin.Nonce.Uint64(), hex.EncodeToString(pay.Hash))

	// engine.Log.Info("交易签名 耗时 %s", config.TimeNow().Sub(start))
	//engine.Log.Info("start CreateTxPay %s", hex.EncodeToString(*pay.GetHash()))
	utils.Log.Info().Interface("普通转账交易", pay).Send()
	// chain.Balance.Frozen(items, pay)
	chain.Balance.AddLockTx(pay)
	return pay, nil
}

/*
	合并多个items的转款交易
*/
// func MergeTxPay(items []*TxItem, address *crypto.AddressCoin, gas, frozenHeight uint64, pwd, comment string) (*Tx_Pay, error) {
// 	commentbs := []byte{}
// 	if comment != "" {
// 		commentbs = []byte(comment)
// 	}
// 	// start := config.TimeNow()

// 	chain := forks.GetLongChain()
// 	// _, block := chain.GetLastBlock()
// 	currentHeight := chain.GetCurrentBlock()
// 	//查找余额
// 	vins := make([]*Vin, 0)
// 	// total := uint64(0)
// 	var total = uint64(0)

// 	for _, item := range items {
// 		total += item.Value
// 	}
// 	// engine.Log.Info("查找余额耗时 %s", config.TimeNow().Sub(start))
// 	if total < gas {
// 		//资金不够
// 		return nil, config.ERROR_not_enough
// 	}
// 	amount := total - gas
// 	if len(items) > config.Mining_pay_vin_max {
// 		return nil, config.ERROR_pay_vin_too_much
// 	}

// 	for _, item := range items {
// 		puk, ok := keystore.GetPukByAddr(*item.Addr)
// 		if !ok {
// 			return nil, config.ERROR_public_key_not_exist
// 		}
// 		vin := Vin{
// 			Txid: item.Txid,      //UTXO 前一个交易的id
// 			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 			Puk:  puk,            //公钥
// 		}
// 		vins = append(vins, &vin)
// 		// engine.Log.Info("item:%+v", item)
// 	}

// 	//构建交易输出
// 	vouts := make([]*Vout, 0)
// 	vout := Vout{
// 		Value:        amount,       //输出金额 = 实际金额 * 100000000
// 		Address:      *address,     //钱包地址
// 		FrozenHeight: frozenHeight, //
// 	}
// 	vouts = append(vouts, &vout)
// 	//找零
// 	if total > amount+gas {
// 		vout := Vout{
// 			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
// 			Address: *items[0].Addr,       //找零地址
// 		}
// 		vouts = append(vouts, &vout)
// 	}

// 	// engine.Log.Info("构建输入输出 耗时 %s", config.TimeNow().Sub(start))

// 	var pay *Tx_Pay
// 	for i := uint64(0); i < 10000; i++ {
// 		//没有输出
// 		base := TxBase{
// 			Type:       config.Wallet_tx_type_pay,                       //交易类型
// 			Vin_total:  uint64(len(vins)),                               //输入交易数量
// 			Vin:        vins,                                            //交易输入
// 			Vout_total: uint64(len(vouts)),                              //输出交易数量
// 			Vout:       vouts,                                           //交易输出
// 			Gas:        gas,                                             //交易手续费
// 			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
// 			Payload:    commentbs,                                       //
// 		}
// 		pay = &Tx_Pay{
// 			TxBase: base,
// 		}

// 		// pay.MergeVout()

// 		// startTime := config.TimeNow()

// 		//给输出签名，防篡改
// 		for i, one := range pay.Vin {
// 			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
// 			if err != nil {
// 				return nil, err
// 			}
// 			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
// 			sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
// 			// engine.Log.Info("sign :%v", sign)
// 			pay.Vin[i].Sign = *sign
// 		}

// 		// engine.Log.Info("给输出签名 耗时 %d %s", i, config.TimeNow().Sub(startTime))

// 		pay.BuildHash()
// 		// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
// 		if pay.CheckHashExist() {
// 			pay = nil
// 			continue
// 		} else {
// 			break
// 		}
// 	}
// 	// engine.Log.Info("交易签名 耗时 %s", config.TimeNow().Sub(start))

// 	chain.Balance.Frozen(items, pay)
// 	return pay, nil
// }

/*
创建多个转款交易
*/
func CreateTxsPay(srcAddr *crypto.AddressCoin, address []PayNumber, gas uint64, pwd, comment string) (*Tx_Pay, error) {

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]*Vin, 0)

	total, item := chain.Balance.BuildPayVinNew(srcAddr, amount+gas)
	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }

	// for _, item := range items {
	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk, //公钥
		//					Sign: *sign,           //签名
	}
	vins = append(vins, &vin)
	// }

	//构建交易输出
	vouts := make([]*Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:        one.Amount,       //输出金额 = 实际金额 * 100000000
			Address:      one.Address,      //钱包地址
			FrozenHeight: one.FrozenHeight, //
		}
		vouts = append(vouts, &vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	// if total > amount+gas {
	// 	vout := Vout{
	// 		Value:   total - amount - gas,       //输出金额 = 实际金额 * 100000000
	// 		Address: keystore.GetAddr()[0].Addr, //钱包地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	vouts = CleanZeroVouts(&vouts)
	vouts = MergeVouts(&vouts)

	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay,                       //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			// CreateTime: config.TimeNow().Unix(),         //创建时间
			Payload: commentbs, //
			Comment: []byte{},
		}
		pay = &Tx_Pay{
			TxBase: base,
		}
		// pay.CleanZeroVout()

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

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
	// engine.Log.Info("冻结交易 %+v", items)
	// engine.Log.Info("创建多人转账交易 %d %+v", block.Height, pay)
	// chain.Balance.Frozen(items, pay)
	chain.Balance.AddLockTx(pay)
	return pay, nil
}

/*
多人转账
*/
type PayNumber struct {
	Address      crypto.AddressCoin //转账地址
	Amount       uint64             //转账金额
	FrozenHeight uint64             //冻结高度
}

/*
创建多个地址转账交易，带签名
*/
func CreateTxsPayByPayload(address []PayNumber, gas uint64, pwd string, cs *CommunitySign) (*Tx_Pay, error) {
	chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]*Vin, 0)

	total, item := chain.Balance.BuildPayVinNew(nil, amount+gas)
	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }
	// for _, item := range items {
	// engine.Log.Info("打印地址 %s", item.Addr.B58String())
	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		// engine.Log.Error("异常：未找到公钥")
		return nil, config.ERROR_public_key_not_exist
	}
	// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk, //公钥
		//					Sign: *sign,           //签名
	}
	vins = append(vins, &vin)
	// }

	//构建交易输出
	vouts := make([]*Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:        one.Amount,       //输出金额 = 实际金额 * 100000000
			Address:      one.Address,      //钱包地址
			FrozenHeight: one.FrozenHeight, //
		}
		vouts = append(vouts, &vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas,            //输出金额 = 实际金额 * 100000000
			Address: Area.Keystore.GetAddr()[0].Addr, //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	vouts = CleanZeroVouts(&vouts)
	vouts = MergeVouts(&vouts)

	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay,                       //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			// CreateTime: config.TimeNow().Unix(),         //创建时间
			Comment: []byte{},
		}
		// pay.CleanZeroVout()
		var txItr TxItr = &Tx_Pay{
			TxBase: base,
		}

		//给payload签名
		if cs != nil {
			addr := crypto.BuildAddr(config.AddrPre, cs.Puk)
			_, prk, _, err := Area.Keystore.GetKeyByAddr(addr, pwd)
			if err != nil {
				return nil, err
			}
			txItr = SignPayload(txItr, cs.Puk, prk, cs.StartHeight, cs.EndHeight)
		}

		pay = txItr.(*Tx_Pay)

		//给输出签名，防篡改
		for i, one := range *pay.GetVin() {
			_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

			sign := pay.GetSign(&prk, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
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
	// chain.Balance.Frozen(items, pay)
	chain.Balance.AddLockTx(pay)
	return pay, nil
}

/*
	创建离线交易
*/

func CreateOfflineTx(keyStorePath, srcaddress, address, pwd, comment string, amount, gas, frozenHeight uint64, nonce uint64,
	currentHeight uint64, domain string, domainType uint64) (tx, hash string, err error) {

	srcaddr := srcaddress
	src := crypto.AddressFromB58String(srcaddr)
	//判断地址前缀是否正确
	//if !crypto.ValidAddr(config.AddrPre, src) {
	//	return
	//}

	addr := address
	dst := crypto.AddressFromB58String(addr)
	//if !crypto.ValidAddr(config.AddrPre, dst) {
	//	return
	//}

	key := InitKeyStore(keyStorePath, pwd)

	txpay, err := CreateTxPayV2(key, &src, &dst, amount, gas, frozenHeight, pwd, comment, nonce, currentHeight, domain, domainType)
	if err != nil {
		return
	}
	paybs, _ := txpay.Proto()

	//engine.Log.Error("交易hash: %s", fmt.Sprintf("%x", txpay.Hash))

	hash = fmt.Sprintf("%x", txpay.Hash)

	//data := make(map[string]interface{})
	//data["tx"] =
	tx = base64.StdEncoding.EncodeToString(*paybs)
	return

}

/*
*
初始化keystore 临时测试
*/
func InitKeyStore(path, pwd string) *keystore.Keystore {
	if path == "" {
		path = "conf/keystore.key"
	}
	//addrPre := "IM"
	if config.AddrPre == "" {
		config.AddrPre = "IM"
	}
	addrPre := config.AddrPre

	//定义本地keystore保存位置，地址前缀
	/*key := keystore.NewKeystore(path, addrPre)

	//加载本地的keystore
	err := key.Load()
	if err != nil {
		//本地没有就创建
		err = key.CreateNewWallet(config.Wallet_keystore_default_pwd)
		if err != nil {
			panic("创建keystore错误:" + err.Error())
		}
		err = key.Save()
		if err != nil {
			panic("保存keystore错误:" + err.Error())
		}
	}*/
	//panic("创建Addr错误 888881:" + fmt.Sprintf("path:%s", path))

	//bbb := `{"coinbase":0,"dhindex":0,"seed":"jX7DqhbsSP17kdCpVbm/sjdR4OqgYhlBJI8mguB3LVI=","checkhash":"fqUFgE/EMsVA3g5ayd19R7w4OxLo20DQ+wzXUkLTw0k=","addrs":[{"index":0,"nickname":"","addr":"SU2P3EzuFRFBxH8I/JAmgfxnhqH8oX1EsVUC","puk":"a2HtDsSlipIp2dZ37syaixsGr3FaSFhFYyjDMW9Aypw=","subKey":"v2Uy5Dt51fJURfagEgaj5gCIXJr9YNx5msvoxLss2oPj5W8Dp1ds6RqNBda3qfvXnek/I6ytYIHyugNUBA07TzcR6njc2wKsRa+/gPqxHCc=","checkhash":"UHN/uAyH8BIrRuke57sYlmPtmKJBXLQcAf2MmRtq7aU=","version":4},{"index":1,"nickname":"","addr":"SU33uCUrLqg769G2xy4QGS1csnPsKtGENnkC","puk":"Hlt1ae2QfxyxMHJUj6YM4BVuqnvWY0vfDLQ6PcCnLFc=","subKey":"OzcMS9RHK452wIyS/A2Djz+7gDzBEQVjb1FUFcLyU2v+UQ1+wy3vC/bCV+FM/DpHxnV9UC864v7D9GzRZAtcLb/uykbCWdwl7DBj2sFd2LM=","checkhash":"nIm4qyi3uprfkpp8MqcfbDFLCWzbAK6271Sdx02kFyo=","version":4},{"index":2,"nickname":"","addr":"SU37Vem4zF8AeMDoxGvPcRCNjhffeqwiW/kC","puk":"JxgTj9eIT45i5CiV7jyMxZOzCqrtvdRm8UUoVhkUYmE=","subKey":"BfkM+VvuD8vMTI1ri3Ne/+9lcmUNRX8nZfNTalYsdAqQ+hCjzjqod2tKZj6Aj0vUyea7Z2uue/5p4ffxqfxrBydR5LY0xfRxWk10uRwFSvA=","checkhash":"je1plueSJ9pP4dzdFMNk1UJjrA206zUiAWkmZnZBYrs=","version":4},{"index":3,"nickname":"","addr":"SU19fkn/Al3MM+BPWJ2BVSVboVdB04rsjLEC","puk":"AnG3PptAlluDdyume805XiSoKh4OvWbHyXqQLdVbuvk=","subKey":"z8lFDaHLyw0gyLfHSA4EVgSEZFvJoTSJEmo8ccz4Qe+jIniVEJj5u74dfvp6NhVj6nAghoELC6kCDysVMz6uRlUMWw6YFgOHPwBXytsNJEY=","checkhash":"5y318p2aK1xtotkIdu3+po6/3VpDlQoJWb5oKf+zboU=","version":4},{"index":4,"nickname":"","addr":"SU0Yo5rCwA2ojFfZacBFMX/kV3zhFIGi/xoC","puk":"lTxiTfMJ9w+Y2WhR0q4qhcxNZPAKXuURFDWHWPYUZbY=","subKey":"qGPaaParj2rxdH1SbBsDLtOHmdSub4qYVNvWNlAU/OFqnvG73FoFN0EBv9RqpaN7Wx9sbraXb68Xa6pj+O46On1CnpndFZtrBRkLLLRz5KM=","checkhash":"XNWkeamcPL8zNO5FDhiwmZk1k48EJ5qFWY+VYvJFytU=","version":4},{"index":5,"nickname":"","addr":"SU1rOJGIwMQ4L52YKSFJoaDUjX0R296gHKoC","puk":"VgjhOxNxYt7F/wal04toQMG19FgXkTukKN3r0N9/h5Y=","subKey":"500ukZ/oQIuTIpaJ8CKVaefIqwRujsmDQ3DbdhJUgd7g7yPLwRyVIKq/GANnqxguEO9NWJWlP+b+0l+QvjA+ceyWNOj/35fgQwdQ8rD3t+g=","checkhash":"LVLfSL0dKWkOh8Dm2SNNNwE3hjgaGnsSli8gjEze3/0=","version":4},{"index":6,"nickname":"","addr":"SU1RNlqcTYxppmbK4DX7/xhFP8qv4N/qt7MC","puk":"v45i67LV1/X4jWv6BoqvGRjKw9C1SHS9/k/JIa+oktY=","subKey":"Ng4hNSR2WAArcDt4RTEgX4/kSMsuNB4E1N6sdfTI++JNViVGN5AbjBBL5W96der1nh0QBPBtz8sBHj49JigL3y8dR9zwrvGcvtfMqCtsG8I=","checkhash":"loNT7j8LxpB7BLbPl8BRTYyzUY0iyLldsDZfZshDGA4=","version":4},{"index":7,"nickname":"","addr":"SU2ggmeBuC8bX3Kvdgwd+UnlHsC8wMKtcE8C","puk":"SZ2d+Z9SGOQWHiESeU5fOAC0I+4kFHD1P4E0j6fHQq4=","subKey":"0JxmtMFcsyCt+wda+RszKK1h2n/rR8GVY45poIxiHvtPNyYkNF240UL8UuOOA166Q1f/eHtw78jHyKUqh37FVr43IK6lF2Ednv3jPy00M2s=","checkhash":"ZfMowchSps87CvSNqzexVagWokOYANsSR6oiN2yxHD0=","version":4},{"index":8,"nickname":"","addr":"SU2jTimCFYa9QHc5A/9MdfWuwGumjbstKlIC","puk":"yMdzf/Djz1rqnNmGZfbJNHAdQ16OO154FUJE5Y4+Xo0=","subKey":"OhSHydAdU/UpuNl8hsO137s064nhsrKf2ctBZFuugq5pjaGqKPLqNBHnyRa0+BhMPMifURgx2E5IqYBpDTqbpAfX7mDa9EiZ+HyzNzfytBo=","checkhash":"MXgGBAFToYTCJ6sHHNEs7PZptI77C4Col6SxTwVxVoQ=","version":4},{"index":9,"nickname":"","addr":"SU2JdQNXgU6pDZCnHOK3MNVQbrx56/6i3QEC","puk":"m4wM8nv12JH1Y6DXSAO5T0Lu4/VrEJFZlo0aZTBoPtM=","subKey":"SYA+sVwr6uyyFfhqZcPOaRzUtWEiim6dRvmdkbTlnulJmGktSBxU5KS7H0s1lbV7fm3XiqGJs19iySbKrByNiMedkppvKno/8veZcUltKTo=","checkhash":"8vkj8McFq72jpwOrJPGrPiMoUpMA8l1qnzoh3LT3tmQ=","version":4},{"index":10,"nickname":"","addr":"SU0//GToh3OkPAjeAX2H2UBFjUAzl8N8OXEC","puk":"XaV+bY2go5zxOzxoXdbpjWOJ19uSj9i7Mir5YXSqGq0=","subKey":"d3/n8PypAC4ZZlnTJVQqWmGAhEYh9/s3mDJ+YCbEQ4w91wIB/4Pt7kdLtJegmCbenGhF0BVzMNi19ft6QLRhplnMkRPhm1/J6z14WYFjFws=","checkhash":"+PDMqTcGXF6Re/mZFZfbVZfNpuzfxAWONRvKVPzhgM4=","version":4},{"index":11,"nickname":"","addr":"SU1nVcMYSVhVBCXcQltslhHiBrvVWowe/+4C","puk":"ovEYNEROUFieh8vHxx2K23874mSREVlf43pPglNyFCo=","subKey":"3pie62CXlOa7RbgetfT47gOo5SxUHI/5uvGJIqzlywywXY9o4KQU+jSgwgv2c7oUQ7h6KHGjmIr24Ko//aBqIFOrJ6PZirAwtZzT0+vVGgE=","checkhash":"Jr+wjopja1vgtQ+p5FWhLGJ8M9UFGhe1jUtEoZfAdLg=","version":4},{"index":12,"nickname":"","addr":"SU3HcBP0L3l5XwPtK3++aD/fVU3i0zQga+cC","puk":"wgdcH7QHLxcUXBq/5mdns6i/lT4mwKiuu893AvNqTaQ=","subKey":"oaL0ErL3koaV8BysaugVcXz5W6xeSuKmImpFqx30PEZ9yaQAHh0rrM0OAFjJXG/X2npA5cPT9/E/0TAV3q49Oap+c+bm2W1NPLQd6ZnLDGE=","checkhash":"6r8JW1LW7F2SJCCdSQ+c8DKXS3OcQ03rYkl2LwEo4HY=","version":4},{"index":13,"nickname":"","addr":"SU3+xBA07ey/I1J/jZTDNYPaspVp9TmQE/0C","puk":"5CbmEoUNDb9z+nmtJ9buCkOWkK1T6hQkUj7k0D6GgDs=","subKey":"k5FBK1Z5iRGnDuXj+Cmvn9+oQLir3piK1Yta83/hNf1lnc9yWzmio1YbpjqRUe2WHjM2X1Gts1EWXIRc/YXGOy6PKPIjXOQqDzqMvbf4Oqw=","checkhash":"ek6Mu6fhG4/ABbk3CzrWeqL6O+8NKrRF8U0HERbvzMA=","version":4},{"index":14,"nickname":"","addr":"SU0Z09ZCRIubNlRrjit1dLAQBt2qbsdjPscC","puk":"0BuFk5As7/9tGxCaqr+nEc+6hyv9wd5J/3Enuz1i/QA=","subKey":"aPleXTa9TjjxQneoCeQIfwRowDQ5CuWSFXQiQ5GL97FhIPQzPEoWLjVp1ZpKLA8ksZUHpkiq2J+s6G84dmNj9/Isy31EK8/aqBVwngIiy14=","checkhash":"esuBRFDrWS0RvlYmGV1XLN7RIi3TsN4V29HEfIjLK7Y=","version":4},{"index":15,"nickname":"","addr":"SU1C52oBpAvpSerifWPfmy5AxeEDMrLQ9P0C","puk":"QVVJr7V7QbgTWiwkC/Ks8t/Cbh2F6FzNrUiPO2aJdTU=","subKey":"hAe1rqB3pTg1r0aDAXDp24qq89P9QDAKIpkMVXXAUI4F+ZNN0QxrMWsyllVX5bh7FJjXPIFpDWvB6mbo15o5d9TsCfpKvqU5Xj7Vcw+mjRs=","checkhash":"8jRlJ90+nlpctAwdPPE1fyFo93KdAFoa7ZcEdRH45Q8=","version":4},{"index":16,"nickname":"","addr":"SU2rv95HGImgsQFDK2g53OlNUaJzjoUvagoC","puk":"xVkHM+Tx/EKyMD5f4mPoQuYLysw09f2XvdaxKNuK7HQ=","subKey":"hAHpiEPR8Jq4eM+gOIdvgK+RH46BUZRFXLt01N6DjUfB9GjatEIujePA90LiTFQQbfIWMdoJrbZ9STejOm9h2/wDsdZrIoiDP9QTT7lEmwo=","checkhash":"KQ5z9DdkGXQ7W/zDxxpdvGlwIxmUaScBlCmHSAJ1wmw=","version":4},{"index":17,"nickname":"","addr":"SU2O++WRX/5DN7f6MlkioMEVqlp4F0Q0t4kC","puk":"oaaTOzWLUgGfLsRKYaB+T/JJt7YbBqOLJC+QCiq+Caw=","subKey":"266XXSAIc1S41Bxw4UNx3fckjI37E867y3bsw50nkI3UVx+CgFbPm3xywrwTDRsSXpxsDAFVN0YcmEyV0Hlgm46YwFikq/SbchZPwV6ftlc=","checkhash":"WyjXjpa6KDMa5KDRGJrSEiQITG99GVSVWpa3zLWptks=","version":4},{"index":18,"nickname":"","addr":"SU3RDKpn+GaWxZ+fdtkcKxbWpzj/CJ+3jmUC","puk":"291PV4Ic/kMGSSh+UVOpNhLbhOBjdvt1M4HniwUMafQ=","subKey":"p8ca3k79+0OcX+wdTRABKizogZUZx7ooDazDRrqdiph9YHMcCSnf/hVBUBdwXMJ9G4Ncngl6nDG6DofMOcdMD0IWZCStnm4Qt9CV7LCMYds=","checkhash":"hqCzpvfXcMjei3/0KOv5vBtDNvlGBiF1u8AM+IIpiRc=","version":4},{"index":19,"nickname":"","addr":"SU1Qg4+Q95vs8sL8v9HcqONgC6+iObMQxKAC","puk":"l+2OZ2zxkcHtg38PBGZL0+fENpaSSWCbIfZgq6nWCEs=","subKey":"wSyU+/25D4OqKY2SGFUIpWRHc+zITKqmczQ0wAHVSiHDH0pK8R/g7XYq/uxP2qL/lkDZh0BvdW5J6WBhu1mgxafZuQWb8Qut+rGAI58Se/Q=","checkhash":"ggS2aQQ1OK7255X0ie+2iZTycaHzas6Q2MdeoEh16OY=","version":4},{"index":20,"nickname":"","addr":"SU3Jwr8e+48LZL3P8GyNjEf3rAR7djABJ9sC","puk":"vwhmaJ0VmlR3JPuCndwkj8h+V+Sln+wVmpcqj00M5cs=","subKey":"Lk97QKIEBPcjNgynxjMzpwKH5zOSkG4F7MDXWJc8Nb+fhvcloT8eqqt61bqWxAhkwRwwYyRGgP0BA8ZffJR8uGOFa06D9i6bRbeYdLlVTs4=","checkhash":"qI4J6In2DROSzK8783ys4hMnxdTXhoHzONrn2jcqdgo=","version":4},{"index":21,"nickname":"","addr":"SU1+W1nj3G1b/C81ofL3DF7SQ0dFSH3EYZIC","puk":"KsLwDzzRNxqxhkhzmCWCnM23Xxyl14HjHWE8YL6FGdk=","subKey":"4bTSsq8n+LKltvH63xIjVJBhCn9fXxTWM29onFhHj2Kfu+aaPg33HJxVyGJ032KmtKVcn/1rCuftKvdJLqQKnxHLfnkrhj6RHS609dX0XOc=","checkhash":"dNZxaRhTBXAj6P7UkzKYaytolzTemwkBumRbTtk7Yh4=","version":4},{"index":22,"nickname":"","addr":"SU08o26Jiw1+UMiJUQA2b/R+4xc2eZHzokIC","puk":"r+kN5M/dlIMWnGoeGgvBdNMUc31N0YRBveBx+SB4Poc=","subKey":"2QiMM4mLFLPNkGu4wVtvrXChFb6ykf5Za1kGQs0Uj1DpUMRAkuzpFD4rQYz9r2rqBPuAsomaugoQEBrmbtlU2XX0StgkTrsWgMtxDuWkxZc=","checkhash":"h4SkWao1PPbnTDy58kS9RZ0tycCZRJZcMnctSyoJ3SM=","version":4},{"index":23,"nickname":"","addr":"SU3cHMZPjjswTsWmb5UC74vwOkz9tmyWr9MC","puk":"Op6Gig4gk+UrAXRU6B94e8I4cqmDNabXzUKHc67iYKg=","subKey":"rfulQ9aajlbv6Fuwwr04baKeuSdpk/qEl48sy9aKxOLyChuwx+499KhPBj+zkIjAwFtPhX3452bXDQxCIogq9o7kiSqqntGZUhj/czm+93c=","checkhash":"AfBjKHM7L/Qp0+FhDq+gy18l2F1X1kOfRiMxpVZ9PyM=","version":4},{"index":24,"nickname":"","addr":"SU1wke9XiZGC/VhMvKQI8gyjJoFqjW3fTDgC","puk":"hH0OrbBOPvEvqSryAdkK46+HgsiJeo86lfwFg9NTWFE=","subKey":"ouGbwGVq6sFVWSG/3CURQSogzpaxGkV81VqacQesPVIIEMaGJ4xlrG+iUjmgRmC/i3XYe4dN7k5MzCYdcdlBECa+qcVKH5wQgP2UD3GnjGg=","checkhash":"kno4/7wzqjE5kAaOpiMn4m70S6S7A0eoNgKFSn8iu6Q=","version":4},{"index":25,"nickname":"","addr":"SU1SyiFTSSsJkoN/dPUbWN9XV2mdy6b1KpsC","puk":"PVf46fOfRTdN+HIxrs4vLe4rq1OptyPJLJEh0rI9xMQ=","subKey":"J0vBgto9libA3CVNIBntx3qt/5Hw1nyJ0UCKiYrXaVul8kc0ouuFzBwcM3Cun3At/uyU5DwmrR66xzaNe83k33Up/f9TJcqdQuEkmdZ9uYk=","checkhash":"vAOVlmVw0tgyqQsgbzNIwaRAauBe0IMHgKAmx/UOQms=","version":4},{"index":26,"nickname":"","addr":"SU0AQ3ZNhpzXajGx0itCIXKrTPHDji7QyOoC","puk":"mr8SNaZsg1WUzVkrx+2MpKyRh1KcPkfiNv6QOobuE1o=","subKey":"uc2rEsIsQeE7YveCoxcm3aJhLYl/I883pty3D+TVdmiMWy/Wqp2XbmLFRqDvZ1MLJld7eR75+hsdhmXuji/jvGBFCowG7dPTP34NKlk4jqs=","checkhash":"IBMJ1rXwL5Hs11jqgZOgbvbyO338ZXi4MsCDGBja1u0=","version":4},{"index":27,"nickname":"","addr":"SU2wVLO8M5lAejNSxeuGRJ5p7z3wQOqX7eMC","puk":"vi6hboi8F1rzdLIsX1RvIlKUK8f6+7L+ttAwiIERu5o=","subKey":"C4kvoKNtJE4pwZItyCzYv4D3Am7UFeqDuc/E/x+zjzmOr0XVy2zUe4NSpuHv3C5CSARLahyldb24N2rt0BrWFWrV0zoHuJcKVuHlKCzTSoU=","checkhash":"yw8Ux3KfFt7WNP3kTpfI+EmqAsz8UexrPAp876FXSp8=","version":4},{"index":28,"nickname":"","addr":"SU0Q4lHiNBuUVbI8Duq3gxilFB2P2A8XPDAC","puk":"ArSwTNp5bkwoqS2iX4wTwhQzOLKYNyNLHhOxPq/Ng9A=","subKey":"kB3JQIA/KWWhhqBmemVdG2CGdRetBXK3H4qTwCMcnwdJncRLagkz9Kz9y7Eq2Wzg4y6ksIDKHZDuB7B6hCu5TRrC0Mx6WcY/wGRYvaLluuU=","checkhash":"3JQV/5J38KtRyk8rN06KYzeuqlZw/+GL6SpFrJX4ZkE=","version":4},{"index":29,"nickname":"","addr":"SU12lsfOLHLWoePP5vhVrEWGPwa10Mr9XJQC","puk":"LnsG41ucDIboN2LSF9dsPajp1gMA211AA18NjSdH1AQ=","subKey":"gdy1wxNFxaCPqpwmPiQ0J6pdt9vC9rhU+cMUmyLgNEeTDyBBIt4a+TyXD/GD09A2UoHod21PN0wChY7mT0vZP43Y/ooqnDKvjJp6a9+KTxs=","checkhash":"oikDFDQAPwPIqPtqs4eJH2WA9qExsblz++i12yuuGwQ=","version":4},{"index":30,"nickname":"","addr":"SU2B2s9eSLWXadzvc8o54YrT7Y6M4MToNSAC","puk":"IWzumUCXlpKF6PW3v3o1LVRCUVCuS+wEJzvYAUHM3ns=","subKey":"0+cS3lqm4EeOzzhIG4Nuark+nlpKTe3CZZLugYeHVhCZsdVpOyNmE2AOqIyqNMgfOIRuPoeVeWd84mH6kuXZ0ZQVh03yTA3uDkyMIaIQcvM=","checkhash":"rkFdMDqrFFk+Rxug2duMgj92+TH9NkmiUxRzYMxaVQE=","version":4},{"index":31,"nickname":"","addr":"SU1f5tfgDOXP5BTs4yS5JaSAu71Kqqvld1MC","puk":"NazwOpCesjAYI53uD7mUUd8JHYipXrneqEm7+w6X+YU=","subKey":"A4s/uCmmAZaFXJFqDOCKlwph+Z1p5wocRedec5Ka/+kjt3Ig7DkqYThQtrobrZ2aXML75gSbZ0YCssntS9bxbS0+pDxPx/Aqh72z6DOTTFU=","checkhash":"nKp189XwUQD6OFPXf3fm5ePV0UyJlSxVas6wZNpyWqo=","version":4},{"index":32,"nickname":"","addr":"SU2KJINjBuFSkl8+dKSFn080YUwZHP1R0gUC","puk":"6qLrUumqWVaSQ6KVwzJXWQ57/EFG718/OXk9av9H54s=","subKey":"oQzWQV5aKEvSZCdpwEgt8z3JGnIEqgs2lWrK+uHKppsf0V8ajhDIhkkPe786o0o4HJijiR3z+Ia1Me623TpUO4hXeIFTvi3rDWGuDO7HsFU=","checkhash":"MAj0dbKAHzKsCpZSlWFRR3jDzswV5m7WMp0fe4DXAhM=","version":4},{"index":33,"nickname":"","addr":"SU2LJQidyjujLvA876Q0HUwtNRKGeWS5GJMC","puk":"Rhuo5qzQnHZGpUtz8+v+moS1VyUVHiAgBpdK9qoOgMA=","subKey":"3WLIeJarfF3KPKV3XS7T6wZ7dgC3UJFFHSUQQjyGz0m7sum27+4+hKhNZ+dRtVbsawLqYuCUrLjfX+bKGqTp/iTYETjI5FKF4GJQXyhB7hA=","checkhash":"QeJzSJvN8nCNeHlQgl650vSYIubOyM3uZpG1PfUsJjQ=","version":4},{"index":34,"nickname":"","addr":"SU1SkAseMq9mncIf1GO7AwOx80PgIeQ8gpAC","puk":"CAE4Qjw7jXM6WY+hsWhowjLaxkNFU1qgX4XXyrtq5VU=","subKey":"QczfMI93NUjne/JAT2D64VBur4gxerKkEH9xVbf+b3cDDXHxiXGt7JQ3utLiVYmro+H240uvi5UquOHHICgHQ0vsF5MT0gLzbs+wFalm9yo=","checkhash":"VZpgWw0Grx1kItLKwK8PRblxaoUh6pPrtRdfIxedLZI=","version":4},{"index":35,"nickname":"","addr":"SU3lb1HQivByC7We2PlRLZkTM1tRCr1ZsjkC","puk":"73CF4SkbITUrysG0Wbw/4/0UKmOPnojXu3IN6NqDXf0=","subKey":"EK8IkHnG5Tx1dKFlxa5kTkccbf8Riu8AWS/HPRZtjx/81KFUhqYGrzoIDA8MRfoY/9wmyUlEfxa1T4Dt+n/4duHYaVcfCFR138XXwYW8/nY=","checkhash":"0hKGU2+DyIlPsAYm56f3PR1UYoBCQB6iYQ1li0YclhM=","version":4},{"index":36,"nickname":"","addr":"SU1MHq167M4UWx+N4VKDilzNkICJxd7zxTsC","puk":"6MGMq18I5pm7xyBBtVr+J1QWNjOAhsBGTqiwjAoWMH8=","subKey":"mJT/mIMC6I0kzPvAH90XKPRScxeZvsxaKoBc2E8Jae78y8r91I9UgjWytrv4GoELeojoXDQ+332OQzkhK1jhrbjq6aw/e35RWoOShCZyTwQ=","checkhash":"pkuWFqrXpHPAp7GisnXoufnPwTGfDztlVTSYX/glmik=","version":4},{"index":37,"nickname":"","addr":"SU0jn7fHbgGJzRKSxG20ThH3c3yfJGZs6e8C","puk":"I7bJJWEzEqEY1X73inYlj+AN73cGOyfFdLSk4vPm/VA=","subKey":"26MViJoy3aNL2hKX8nfA8O67vc9aR8L4/DjznYPrIrFclJ0EgUZyH+xhzWI7lYoFBHwH68oX5Nw2Wz8WCJSfxT7dcXkY7Zywyx9Fo0kiIqQ=","checkhash":"k6XEeQH9swj3jJFQVprNhH8aCI4pgh9K+CiGH9kkFc4=","version":4},{"index":38,"nickname":"","addr":"SU3xkto35AH18G5jckoegdTZuXyRTPrprJwC","puk":"YbHMAilnX0BmMWRpA/XX40jYBMUfHgIKz6xrVPlnpOY=","subKey":"CwIivC/MzDkROeOwa2mlI4YFRwsbrTupjgRKbUtaqYnBeuA9/cwk5vZdvwd8+QHCnAUNFk8kFDl7z+5Tp8isNwJQYO8kzwTyHnsDNkfUlPQ=","checkhash":"PyLUzwnL+sR8dCt9fm1T7/Ukf7CTIO5N58eZTkSJCz0=","version":4},{"index":39,"nickname":"","addr":"SU2h0TlKcZgkqkeOsPN7mhVrqBUsMfPASRsC","puk":"cfkkErhG5odNK9X2E7iw7wqdQ17wczbHvZhvrCvFTOY=","subKey":"bKRXFAmlzFGjHi3FWFcCmWk5tNT2F9ti3Ci12lKsxjGqkwHKb9BycphzCkZMX16NH9/pJFQ0lU3lceCoqwR+9UVfc1/gszGqCKC91TTW/so=","checkhash":"rG50NcExdj5VDgWWhuA+phKZc21fXk0CE0iFXJPXXM8=","version":4},{"index":40,"nickname":"","addr":"SU0YKIEjf/fKHq5FrW2uYhymU+IWVztEAQsC","puk":"KgktXLmbC91hOS/pY8zGq4KGLwIYxbbIWfD4dOvGinc=","subKey":"hZNV+teT9HWSfBykq9haxgQ4wNMMf6QZ360GVayuP1OERG/dHfaGGogt3kwB3+NkT0puUiUmDz/4qZaHeJ5T7WvlCiScj3FZvwNp3NVvWDg=","checkhash":"SZs/Fjb82NcZRE1rKRG3MnZ5DKQYkdTqeMTJUcyZxr0=","version":4}],"dhkey":[{"index":2,"keypair":{"PublicKey":[166,247,224,252,53,186,222,181,121,155,88,80,250,233,78,50,55,222,178,42,51,42,42,80,85,77,242,219,86,55,65,82],"PrivateKey":[145,250,238,61,197,114,179,32,1,114,180,152,83,248,42,222,213,38,5,89,143,249,191,193,206,65,46,225,162,170,198,120]},"checkhash":"Ne25tY73lnvoFPknBnc+GL0TyjsLV5Dlp7+OK5a/uuk=","subKey":"ASvMQjAIrQHV0RFhjsiG3AtND0nfgsjN+aqBBFqKEolnYor7iIeCZvPFgle2xn4R"}],"netaddr":{"puk":"dQbtOgDOLX0UltLCKtJIPVOr2qlC1Po8bZakef3E2kg=","subKey":"g+Q4QqZLPiVR4Ttt5LEi+9tk7PdNm/LlNdg2P32MerWrOQUXKVuZ3Yckx+CBwegB","checkhash":"l95nHtzaLsZLGCn/s3cxeneJrfvSOYQPRLeJhefg5B4="}}`

	key := keystore.NewKeystore(path, addrPre)
	//if path == "ttt" {
	//	err := key.Load2()
	//	if err != nil {
	//		panic("创建key错误:" + err.Error())
	//	}
	//	return key
	//}

	err := key.Load()
	if err != nil {
		//没有就创建
		err = key.CreateNewKeystore(pwd)
		if err != nil {
			panic("创建key错误:" + err.Error())
		}
	}

	if key.NetAddr == nil {
		_, _, err = key.CreateNetAddr(pwd, pwd)
		if err != nil {
			panic("创建网络地址错误：" + err.Error())
		}
	}

	if len(key.GetAddr()) < 1 {
		//panic("创建Addr错误 88888:" + fmt.Sprintf("path:%s,len:%d,addr:%s", path, len(key.GetAddr()), key.GetCoinbase().Addr.B58String()) + err.Error())
		_, err = key.GetNewAddr(pwd, pwd)
		if err != nil {
			panic("创建Addr错误:" + err.Error())
		}
	}
	if len(key.GetDHKeyPair().SubKey) < 1 {
		_, err = key.GetNewDHKey(pwd, pwd)
		if err != nil {
			panic("创建DHKey错误:" + err.Error())
		}
	}

	return key
}

/*
创建一个转款交易
*/
func CreateTxPayV2(key *keystore.Keystore, srcAddress, address *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, nonceInt uint64, currentHeight uint64, domain string, domainType uint64) (*Tx_Pay, error) {

	//this.GetKeyByAddr(this.Wallets[i].GetAddr(), "pass")

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	//查找余额
	vins := make([]*Vin, 0)

	puk, ok := key.GetPukByAddr(*srcAddress)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := big.NewInt(int64(nonceInt))

	// if paytotal > 4 {
	//	nonce = *new(big.Int).Add(&nonce, big.NewInt(1))
	// }

	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	//engine.Log.Info("新交易nonce:%d", vin.Nonce.Uint64())
	vins = append(vins, &vin)
	// }

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	var pay *Tx_Pay
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_pay, //交易类型
		Vin_total:  uint64(len(vins)),         //输入交易数量
		Vin:        vins,                      //交易输入
		Vout_total: uint64(len(vouts)),        //输出交易数量
		Vout:       vouts,                     //交易输出
		Gas:        gas,                       //交易手续费
		//LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &Tx_Pay{
		TxBase: base,
	}

	// pay.MergeVout()

	// startTime := config.TimeNow()

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return nil, err
		}
		// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	// engine.Log.Info("给输出签名 耗时 %d %s", i, config.TimeNow().Sub(startTime))

	pay.BuildHash()
	// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
	//if pay.CheckHashExist() {
	//	pay = nil
	//	continue
	//} else {
	//	break
	//}
	//}
	//engine.Log.Info("新交易nonce:%d hash:%s", vin.Nonce.Uint64(), hex.EncodeToString(pay.Hash))

	// engine.Log.Info("交易签名 耗时 %s", config.TimeNow().Sub(start))
	//engine.Log.Info("start CreateTxPay %s", hex.EncodeToString(*pay.GetHash()))
	// chain.Balance.Frozen(items, pay)
	//chain.Balance.AddLockTx(pay)
	return pay, nil
}

/*
	创建离线合约交易
*/

func CreateOfflineContractTx(keyStorePath, srcaddress, address, pwd, comment string, amount, gas, frozenHeight, gasPrice uint64, nonce uint64, currentHeight uint64, domain string, domainType uint64, abi, source string) (tx, hash, addressContract string, err error) {

	srcaddr := srcaddress
	src := crypto.AddressFromB58String(srcaddr)
	//判断地址前缀是否正确
	//if !crypto.ValidAddr(config.AddrPre, src) {
	//	return "", "", errors.New("地址前缀不对")
	//}

	var txpay *Tx_Contract
	var paybs *[]byte

	addr := address
	var dst crypto.AddressCoin

	key := InitKeyStore(keyStorePath, pwd)

	//engine.Log.Error(key.GetAddr()[0].Addr.B58String())

	if addr != "" {
		dst = crypto.AddressFromB58String(addr)
		//if !crypto.ValidAddr(config.AddrPre, dst) {
		//	return "", "", errors.New("地址前缀不对1")
		//}
		//engine.Log.Error(dst.B58String())

		txpay, err, addressContract = CreateTxContractNewV2(key, &src, &dst, amount, gas, frozenHeight, pwd, comment, source, abi, 0, gasPrice, nonce, currentHeight, domain, domainType)
		if err != nil {
			return
		}
	} else {
		txpay, err, addressContract = CreateTxContractNewV2(key, &src, nil, amount, gas, frozenHeight, pwd, comment, source, abi, 0, gasPrice, nonce, currentHeight, domain, domainType)

		if err != nil {
			return
		}
	}
	paybs, err = txpay.Proto()
	if err != nil {
		return
	}

	//engine.Log.Error("交易hash: %s", fmt.Sprintf("%x", txpay.Hash))

	hash = fmt.Sprintf("%x", txpay.Hash)

	tx = base64.StdEncoding.EncodeToString(*paybs)
	return

}

/**
构建合约交易
*/
// 多了三个参数 gasPrice,class（合约类型），source
func CreateTxContractNewV2(key *keystore.Keystore, srcAddress, address *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment, source, abi string, class uint64, gasPrice uint64, nonceInt uint64, currentHeight uint64, domain string, domainType uint64) (*Tx_Contract, error, string) {
	commentbs := []byte{}
	if comment != "" {
		commentbs = common.Hex2Bytes(comment)
		//commentbs = []byte(comment)
	}
	sourceBytes := []byte{}
	if source != "" {
		//sourceBytes, _ = common.ZipBytes(common.Hex2Bytes(source))
		sourceBytes, _ = common.ZipBytes([]byte(source))
	}
	abiBytes := []byte{}
	if source != "" {
		//sourceBytes, _ = common.ZipBytes(common.Hex2Bytes(source))
		abiBytes, _ = common.ZipBytes([]byte(abi))
	}

	//压缩源代码

	//chain := forks.GetLongChain()

	//currentHeight := chain.GetCurrentBlock()

	vins := make([]*Vin, 0)

	//total, item := chain.Balance.BuildPayVinNew(srcAddress, amount+gas)
	//
	//if total < amount+gas {
	//	//资金不够
	//	return nil, config.ERROR_not_enough
	//}

	// 注释掉获取公钥
	puk, ok := key.GetPukByAddr(*srcAddress)
	if !ok {
		//engine.Log.Info("公钥找不到")
		return nil, config.ERROR_public_key_not_exist, ""
	}
	nonce := big.NewInt(int64(nonceInt))
	//
	//engine.Log.Error("77777777777777777777777777777777777777777777")
	//fmt.Println(puk)

	// 自定义公钥私钥用于测试
	//var pubZ ed25519.PublicKey
	//pubZ = []byte{180, 211, 198, 89, 105, 185, 123, 32, 49, 186, 84, 49, 241, 184, 62, 151, 29, 152, 5, 221, 73, 125, 220, 121, 168, 245, 203, 205, 194, 148, 189, 253}
	//
	//var prkZ ed25519.PrivateKey
	//prkZ = []byte{62, 20, 158, 78, 199, 253, 176, 153, 33, 255, 51, 189, 172, 7, 176, 103, 187, 145, 121, 16, 219, 134, 46, 200, 158, 24, 98, 210, 197, 210, 237, 108, 180, 211, 198, 89, 105, 185, 123, 32, 49, 186, 84, 49, 241, 184, 62, 151, 29, 152, 5, 221, 73, 125, 220, 121, 168, 245, 203, 205, 194, 148, 189, 253}

	//fmt.Println("nonce", nonce)
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		//Nonce: *nonce,
		Puk: puk, //公钥
	}
	// engine.Log.Info("新交易nonce:%d", vin.Nonce.Uint64())
	vins = append(vins, &vin)

	// }
	//默认为合约调用
	action := "call"
	//构建交易输出
	vouts := make([]*Vout, 0)
	if address == nil {
		// 合约地址生成方式待选择
		data := append(*srcAddress, vin.Nonce.Bytes()...)
		addCoin := crypto.BuildAddr(config.AddrPre, data)
		action = "create"
		//addCoin, _ := config.Area.Keystore.GetNewAddr(*config.WalletPwd)
		address = &addCoin
	}

	vout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	var pay *Tx_Contract

	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_contract,              //交易类型
		Vin_total:  uint64(len(vins)),                           //输入交易数量
		Vin:        vins,                                        //交易输入
		Vout_total: uint64(len(vouts)),                          //输出交易数量
		Vout:       vouts,                                       //交易输出
		Gas:        gas,                                         //交易手续费
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &Tx_Contract{
		TxBase:        base,
		Action:        action,
		GzipSource:    sourceBytes,
		GzipAbi:       abiBytes,
		ContractClass: class,
		GasPrice:      gasPrice,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {

		//注释掉了获取私钥 用于测试
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)

		if err != nil {
			return nil, errors.New(err.Error() + "密码：" + pwd), ""
		}
		//engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()
	//}

	//// 验证合约
	//gasUsed, err := pay.PreExecNew()
	//if err != nil {
	//	return nil, err
	//}
	//pay.GasUsed = gasUsed

	return pay, nil, address.B58String()
}

/*
获取comment
*/
func GetComment(tag string, jsonData map[string]interface{}) (comment string, err error) {

	switch tag {
	// 域名投放 1
	case "LaunchDomain":
		if !TypeConversion("float64", jsonData["len"]) {
			return "", errors.New("len 类型错误")
		}
		if !TypeConversion("float64", jsonData["price"]) {
			return "", errors.New("price 类型错误")
		}
		if !TypeConversion("float64", jsonData["openTime"]) {
			return "", errors.New("openTime 类型错误")
		}
		if !TypeConversion("float64", jsonData["foreverPrice"]) {
			return "", errors.New("foreverPrice 类型错误")
		}
		if !TypeConversion("float64", jsonData["reNewPrice"]) {
			return "", errors.New("reNewPrice 类型错误")
		}
		length := big.NewInt(int64(jsonData["len"].(float64)))
		price := big.NewInt(int64(jsonData["price"].(float64)))
		openTime := big.NewInt(int64(jsonData["openTime"].(float64)))
		foreverPrice := big.NewInt(int64(jsonData["foreverPrice"].(float64)))
		reNewPrice := big.NewInt(int64(jsonData["reNewPrice"].(float64)))
		comment = ens.BuildLaunchDomainLenInput(length, price, openTime, foreverPrice, reNewPrice)
	// 获取部署注册器的输入,ens注册表合约地址，baseName 根节点名称
	case "DelayBaseRegistar":
		if !TypeConversion("string", jsonData["ens"]) {
			return "", errors.New("ens 类型错误")
		}
		if !TypeConversion("string", jsonData["ensPool"]) {
			return "", errors.New("ensPool 类型错误")
		}
		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		if !TypeConversion("[]interface{}", jsonData["stocks"]) {
			return "", errors.New("stocks 类型错误")
		}

		addr := jsonData["ens"].(string)
		enspoolAddr := jsonData["ensPool"].(string)

		//判断地址前缀是否正确
		//if !crypto.ValidAddr("iCom", crypto.AddressFromB58String(addr)) {
		//	return "", errors.New("ens 地址错误")
		//}

		match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,60}$", addr)
		if !match {
			return "", errors.New("ens 地址错误")
		}

		match, _ = regexp.MatchString("^[a-zA-Z0-9]{20,60}$", enspoolAddr)
		if !match {
			return "", errors.New("enspoolAddr 地址错误")
		}

		name := jsonData["name"].(string)
		stocksItr := jsonData["stocks"].([]interface{})
		stocks := make([]ens.Stock, 0)
		for _, v := range stocksItr {
			tempv := v.([]interface{})
			tempAddr := crypto.AddressFromB58String(tempv[0].(string))
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, tempAddr) {
				return "", errors.New("stocks 地址错误")
			}
			to := common.Address(evmutils.AddressCoinToAddress(tempAddr))
			stocks = append(stocks, ens.Stock{To: to, Ratio: uint16(tempv[1].(float64))})
		}

		if len(stocks) == 0 {

			if !TypeConversion("string", jsonData["src"]) {
				return "", errors.New("src 类型错误")
			}

			src, ok := jsonData["src"]
			if !ok {
				return "", errors.New("src 参数")
			}

			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(src.(string))) {
				return "", errors.New("src 地址错误")
			}

			to := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(src.(string))))
			stocks = append(stocks, ens.Stock{To: to, Ratio: uint16(100)})
		}

		comment = ens.BuildDelayRegistrarInput(addr, enspoolAddr, name, stocks)
	// 创建部署平台解析器合约交易
	case "BuildDelayPublicResolver":

		ensAddr := evmutils.AddressToAddressCoin(common.HexToAddress(config.ENS_CONTRACT_ADDR).Bytes())

		comment = ens.BuildDelayResolverInput(ensAddr.B58String())
	//添加域名输入 1
	case "AddDomain":

		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		if !TypeConversion("float64", jsonData["price"]) {
			return "", errors.New("price 类型错误")
		}
		if !TypeConversion("float64", jsonData["open_time"]) {
			return "", errors.New("open_time 类型错误")
		}
		if !TypeConversion("float64", jsonData["forever_price"]) {
			return "", errors.New("forever_price 类型错误")
		}
		if !TypeConversion("float64", jsonData["reNewPrice"]) {
			return "", errors.New("reNewPrice 类型错误")
		}

		name := jsonData["name"].(string)
		price := int64(jsonData["price"].(float64))
		openTime := int64(jsonData["open_time"].(float64))
		foreverPrice := int64(jsonData["forever_price"].(float64))
		reNewPrice := int64(jsonData["reNewPrice"].(float64))

		comment = ens.BuildLaunchDomainNameInput(name, big.NewInt(price), big.NewInt(openTime), big.NewInt(foreverPrice), big.NewInt(reNewPrice))
	// 获取设置节点为合约的输入node [32]byte, label [32]byte, owner common.Address  1
	case "SetDomainManger":

		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		if !TypeConversion("string", jsonData["registar"]) {
			return "", errors.New("registar 类型错误")
		}

		name := jsonData["name"].(string)
		registar := jsonData["registar"].(string)

		//判断地址前缀是否正确
		//if !crypto.ValidAddr("iCom", crypto.AddressFromB58String(registar)) {
		//	return "", errors.New("registar 地址错误")
		//}

		match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,60}$", registar)
		if !match {
			return "", errors.New("registar 地址错误")
		}

		if name == "" {
			comment = ens.BuildNodeOwnerInput(name, registar)
		} else {
			comment = ens.BuildSubNodeRecordInput("", name, registar)
		}
	// 注册域名 1
	case "RegisterDomain":

		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		if !TypeConversion("string", jsonData["owner"]) {
			return "", errors.New("owner 类型错误")
		}
		if !TypeConversion("bool", jsonData["forever"]) {
			return "", errors.New("forever 类型错误")
		}
		if !TypeConversion("float64", jsonData["duration"]) {
			return "", errors.New("duration 类型错误")
		}

		name := jsonData["name"].(string)
		owner := jsonData["owner"].(string)
		forever := jsonData["forever"].(bool)
		duration := uint64(jsonData["duration"].(float64))

		// 如果不是永久 时间不能小于当前结果
		if !forever && duration < 365*24*60*60 {
			//res, err = model.Errcode(model.Nomarl, "min duration is 31536000")
			return "", errors.New("min duration is 31536000")
		}
		if !forever && duration%31536000 != 0 {
			//res, err = model.Errcode(model.Nomarl, "duration must multiple 31536000")
			return "", errors.New("duration must multiple 31536000")
		}

		if forever && duration != 0 {
			return "", errors.New("forever is true,duration is must 0")
		}
		//如果是永久，时间设置为1万年
		//if forever {
		//	duration = 31536000 * 10000
		//}

		comment = ens.BuildRegisterInputV2(name, owner, big.NewInt(int64(duration)), forever)
	//  续费域名  1
	case "ReNewDomain":

		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		if !TypeConversion("bool", jsonData["forever"]) {
			return "", errors.New("forever 类型错误")
		}
		if !TypeConversion("float64", jsonData["duration"]) {
			return "", errors.New("duration 类型错误")
		}

		name := jsonData["name"].(string)
		forever := jsonData["forever"].(bool)
		duration := uint64(jsonData["duration"].(float64))

		if duration < 365*24*60*60 {
			//res, err = model.Errcode(model.Nomarl, "min duration is 31536000")
			return "", errors.New("min duration is 31536000")
		}
		if !forever && duration%31536000 != 0 {
			//res, err = model.Errcode(model.Nomarl, "duration must multiple 31536000")
			return "", errors.New("duration must multiple 31536000")
		}
		//
		if forever {
			return "", errors.New("renew forever is must false")
		}

		comment = ens.BuildReNewInputV2(name, big.NewInt(int64(duration)))

		// 设置其他币种解析
	case "SetDomainOtherResolver":

		if !TypeConversion("string", jsonData["other_address"]) {
			return "", errors.New("other_address 类型错误")
		}
		if !TypeConversion("string", jsonData["root"]) {
			return "", errors.New("root 类型错误")
		}
		if !TypeConversion("string", jsonData["sub"]) {
			return "", errors.New("sub 类型错误")
		}
		if !TypeConversion("float64", jsonData["coin_type"]) {
			return "", errors.New("coin_type 类型错误")
		}

		imAddr := jsonData["other_address"].(string)

		sub := jsonData["sub"].(string)
		root := jsonData["root"].(string)

		if root == "" && sub == "" {
			return "", errors.New("root 和 sub不能为空")
		}

		node := sub + "." + root
		if root == "" {
			node = sub
		} else if sub == "" {
			node = root
		}

		coinType := int64(jsonData["coin_type"].(float64))

		//判断地址前缀是否正确
		//if coinType <= 10 && !crypto.ValidAddr(config.AddrPre, crypto.AddressFromB58String(imAddr)) {
		//	return "", errors.New("other_address 地址错误")
		//}
		if imAddr != "" {
			match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,100}$", imAddr)
			if !match {
				return "", errors.New("other_address 地址错误")
			}

		}

		comment = ens.BuildMultiResolverInput(node, imAddr, big.NewInt(coinType))
	//根域名持有人和平台持有人提现
	case "DomainWithDraw":

		comment = ens.BuildWithDrawInput()
	// 1
	case "DomainTransfer":
		if !TypeConversion("string", jsonData["to"]) {
			return "", errors.New("to 类型错误")
		}
		to := jsonData["to"].(string)

		//判断地址前缀是否正确
		//if !crypto.ValidAddr("iCom", crypto.AddressFromB58String(to)) {
		//	return "", errors.New("to 地址错误")
		//}

		match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,60}$", to)
		if !match {
			return "", errors.New("to 地址错误")
		}

		comment = ens.BuildTransferInput(to)
	case "DomainTransferSub":
		if !TypeConversion("string", jsonData["to"]) {
			return "", errors.New("to 类型错误")
		}
		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		to := jsonData["to"].(string)
		name := jsonData["name"].(string)

		//判断地址前缀是否正确
		//if !crypto.ValidAddr("iCom", crypto.AddressFromB58String(to)) {
		//	return "", errors.New("to 地址错误")
		//}

		match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,60}$", to)
		if !match {
			return "", errors.New("to 地址错误")
		}

		comment = ens.BuildTransferEnsInput(to, name)

		// 域名锁定
	case "SetLockDomain":
		if !TypeConversion("[]interface{}", jsonData["names"]) {
			return "", errors.New("names 类型错误")
		}

		namesItr := jsonData["names"].([]interface{})
		names := make([]ens.LockName, 0)
		for _, v := range namesItr {
			tempv := v.([]interface{})
			names = append(names, ens.LockName{Name: tempv[0].(string), Tag: tempv[1].(string)})
		}

		comment = ens.BuildLockNameInputV1(names)
	case "DelDomainImResolver":

		if !TypeConversion("string", jsonData["sub"]) {
			return "", errors.New("sub 类型错误")
		}
		if !TypeConversion("string", jsonData["root"]) {
			return "", errors.New("root 类型错误")
		}
		if !TypeConversion("float64", jsonData["coin_type"]) {
			return "", errors.New("coin_type 类型错误")
		}

		sub := jsonData["sub"].(string)

		root := jsonData["root"].(string)

		if root == "" && sub == "" {
			return "", errors.New("root 和 sub不能为空")
		}

		node := sub + "." + root
		if root == "" {
			node = sub
		} else if sub == "" {
			node = root
		}

		if _, ok := jsonData["coin_type"]; !ok {
			return "", errors.New("缺少coin_type 参数")
		}
		coinType := int64(jsonData["coin_type"].(float64))

		comment = ens.BuildDelResolverInput(node, big.NewInt(coinType))
	//域名解锁
	case "UnLockDomain":

		bs, errJ := json.Marshal(jsonData["names"])
		if errJ != nil {
			//res, err = model.Errcode(model.TypeWrong, "names")
			return "", errors.New("names 类型错误")
		}
		names := make([]string, 0)
		decoder := json.NewDecoder(bytes.NewBuffer(bs))
		err = decoder.Decode(&names)
		if err != nil {
			return "", errors.New("names 类型错误")

		}

		comment = ens.BuildUnLockNameInput(names)

		//
	case "ModifyLaunchDomain":

		if !TypeConversion("float64", jsonData["len"]) {
			return "", errors.New("len 类型错误")
		}
		if !TypeConversion("float64", jsonData["price"]) {
			return "", errors.New("price 类型错误")
		}
		if !TypeConversion("float64", jsonData["open_time"]) {
			return "", errors.New("open_time 类型错误")
		}
		if !TypeConversion("float64", jsonData["forever_price"]) {
			return "", errors.New("forever_price 类型错误")
		}
		if !TypeConversion("float64", jsonData["reNewPrice"]) {
			return "", errors.New("reNewPrice 类型错误")
		}

		length := int64(jsonData["len"].(float64))
		price := int64(jsonData["price"].(float64))
		openTime := int64(jsonData["open_time"].(float64))
		foreverPrice := int64(jsonData["forever_price"].(float64))
		reNewPrice := int64(jsonData["reNewPrice"].(float64))

		comment = ens.BuildModifyLaunchDomainLenInput(big.NewInt(length), big.NewInt(price), big.NewInt(openTime), big.NewInt(foreverPrice), big.NewInt(reNewPrice))

		//
	case "AbortLaunchDomain":
		if !TypeConversion("float64", jsonData["len"]) {
			return "", errors.New("len 类型错误")
		}

		length := int64(jsonData["len"].(float64))

		comment = ens.BuildAbortLaunchDomainLenInput(big.NewInt(length))

		//
	case "ModifyAddDomain":

		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		if !TypeConversion("float64", jsonData["price"]) {
			return "", errors.New("price 类型错误")
		}
		if !TypeConversion("float64", jsonData["open_time"]) {
			return "", errors.New("open_time 类型错误")
		}
		if !TypeConversion("float64", jsonData["forever_price"]) {
			return "", errors.New("forever_price 类型错误")
		}
		if !TypeConversion("float64", jsonData["reNewPrice"]) {
			return "", errors.New("reNewPrice 类型错误")
		}

		name := jsonData["name"].(string)
		price := int64(jsonData["price"].(float64))
		openTime := int64(jsonData["open_time"].(float64))
		foreverPrice := int64(jsonData["forever_price"].(float64))
		reNewPrice := int64(jsonData["reNewPrice"].(float64))

		comment = ens.BuildModifyLaunchDomainNameInput(name, big.NewInt(price), big.NewInt(openTime), big.NewInt(foreverPrice), big.NewInt(reNewPrice))

		//
	case "AbortAddDomain":

		if !TypeConversion("string", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		name := jsonData["name"].(string)

		comment = ens.BuildAbortLaunchDomainNameInput(name)
	case "TransferStock":
		if !TypeConversion("[]interface{}", jsonData["stocks"]) {
			return "", errors.New("stocks 类型错误")
		}

		stocksItr := jsonData["stocks"].([]interface{})
		stocks := make([]ens.Stock, 0)
		for _, v := range stocksItr {
			tempv := v.([]interface{})

			addrTemp := crypto.AddressFromB58String(tempv[0].(string))
			//判断地址前缀是否正确
			//if !crypto.ValidAddr("iCom", addrTemp) {
			//	return "", errors.New("stocks 地址错误")
			//}

			match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,60}$", tempv[0].(string))
			if !match {
				return "", errors.New("stocks 地址错误")
			}

			to := common.Address(evmutils.AddressCoinToAddress(addrTemp))
			stocks = append(stocks, ens.Stock{To: to, Ratio: uint16(tempv[1].(float64))})
		}

		if len(stocks) == 0 {
			if !TypeConversion("string", jsonData["src"]) {
				return "", errors.New("src 类型错误")
			}
			src, ok := jsonData["src"]
			if !ok {
				return "", errors.New("src 参数")
			}

			//判断地址前缀是否正确
			//if !crypto.ValidAddr("iCom", crypto.AddressFromB58String(src.(string))) {
			//	return "", errors.New("src 地址错误")
			//}

			match, _ := regexp.MatchString("^[a-zA-Z0-9]{20,60}$", src.(string))
			if !match {
				return "", errors.New("src 地址错误")
			}

			to := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(src.(string))))
			stocks = append(stocks, ens.Stock{To: to, Ratio: uint16(100)})
		}
		comment = ens.BuildTransferStockInput(stocks)
	// 初始化类型投放 类型和名字
	case "SetTypeName":

		if !TypeConversion("string", jsonData["typeName"]) {
			return "", errors.New("typeName 类型错误")
		}
		typeName := jsonData["typeName"].(string)

		if !TypeConversion("[]interface{}", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		nameItr := jsonData["name"].([]interface{})

		var nameTemp []string
		for _, v := range nameItr {
			nameTemp = append(nameTemp, v.(string))
		}

		comment = ens.BuildSetTypeNameInput(typeName, nameTemp)
	case "LaunchDomainType":
		if !TypeConversion("string", jsonData["typeName"]) {
			return "", errors.New("typeName 类型错误")
		}
		if !TypeConversion("float64", jsonData["price"]) {
			return "", errors.New("price 类型错误")
		}
		if !TypeConversion("float64", jsonData["openTime"]) {
			return "", errors.New("openTime 类型错误")
		}
		if !TypeConversion("float64", jsonData["foreverPrice"]) {
			return "", errors.New("foreverPrice 类型错误")
		}
		if !TypeConversion("float64", jsonData["reNewPrice"]) {
			return "", errors.New("reNewPrice 类型错误")
		}

		typeName := jsonData["typeName"].(string)
		price := big.NewInt(int64(jsonData["price"].(float64)))
		openTime := big.NewInt(int64(jsonData["openTime"].(float64)))
		foreverPrice := big.NewInt(int64(jsonData["foreverPrice"].(float64)))
		reNewPrice := big.NewInt(int64(jsonData["reNewPrice"].(float64)))

		comment = ens.BuildLaunchDomainTypeInput(typeName, price, openTime, foreverPrice, reNewPrice)
	case "LaunchDomainTypeAbort":

		if !TypeConversion("string", jsonData["typeName"]) {
			return "", errors.New("typeName 类型错误")
		}
		typeName := jsonData["typeName"].(string)

		comment = ens.BuildLaunchDomainTypeAbortInput(typeName)
	case "LaunchDomainTypeModify":
		if !TypeConversion("string", jsonData["typeName"]) {
			return "", errors.New("typeName 类型错误")
		}
		if !TypeConversion("float64", jsonData["price"]) {
			return "", errors.New("price 类型错误")
		}
		if !TypeConversion("float64", jsonData["openTime"]) {
			return "", errors.New("openTime 类型错误")
		}
		if !TypeConversion("float64", jsonData["foreverPrice"]) {
			return "", errors.New("foreverPrice 类型错误")
		}
		if !TypeConversion("float64", jsonData["reNewPrice"]) {
			return "", errors.New("reNewPrice 类型错误")
		}

		typeName := jsonData["typeName"].(string)
		price := big.NewInt(int64(jsonData["price"].(float64)))
		openTime := big.NewInt(int64(jsonData["openTime"].(float64)))
		foreverPrice := big.NewInt(int64(jsonData["foreverPrice"].(float64)))
		reNewPrice := big.NewInt(int64(jsonData["reNewPrice"].(float64)))

		comment = ens.BuildLaunchDomainTypeModifyInput(typeName, price, openTime, foreverPrice, reNewPrice)
	// 初始化官方锁定 tag和名字
	case "SetMasterLock":

		if !TypeConversion("string", jsonData["tag"]) {
			return "", errors.New("tag 类型错误")
		}
		tagItr := jsonData["tag"].(string)

		if !TypeConversion("[]interface{}", jsonData["name"]) {
			return "", errors.New("name 类型错误")
		}
		nameItr := jsonData["name"].([]interface{})

		var nameTemp []string
		for _, v := range nameItr {
			nameTemp = append(nameTemp, v.(string))
		}

		comment = ens.BuildAddTypeNameInput(tagItr, nameTemp)
	case "LockMasterDomain":

		if !TypeConversion("[]interface{}", jsonData["tags"]) {
			return "", errors.New("tags 类型错误")
		}
		tagsItr := jsonData["tags"].([]interface{})

		var tagsTemp []string
		for _, v := range tagsItr {
			tagsTemp = append(tagsTemp, v.(string))
		}

		comment = ens.BuildLockMasterInput(tagsTemp)
	case "UnLockMasterDomain":

		if !TypeConversion("[]interface{}", jsonData["tags"]) {
			return "", errors.New("tags 类型错误")
		}
		tagsItr := jsonData["tags"].([]interface{})

		var tagsTemp []string
		for _, v := range tagsItr {
			tagsTemp = append(tagsTemp, v.(string))
		}

		comment = ens.BuildUnLockMasterInput(tagsTemp)
	case "SetOperator":
		if !TypeConversion("[]interface{}", jsonData["addrs"]) {
			return "", errors.New("addrs 类型错误")
		}
		addrsItr := jsonData["addrs"].([]interface{})
		if len(addrsItr) == 0 {
			return "", errors.New("addrs 不能为空")
		}

		var addrs []common.Address
		for _, v := range addrsItr {
			t := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(v.(string))))
			if t == evmutils.ZeroAddress {
				return "", errors.New("addrs 地址类型错误")
			}
			addrs = append(addrs, t)
		}

		comment = ens.BuildSetOperatorInput(addrs)
	case "DelOperator":
		if !TypeConversion("[]interface{}", jsonData["addrs"]) {
			return "", errors.New("addrs 类型错误")
		}
		addrsItr := jsonData["addrs"].([]interface{})
		if len(addrsItr) == 0 {
			return "", errors.New("addrs 不能为空")
		}

		var addrs []common.Address
		for _, v := range addrsItr {
			t := common.Address(evmutils.AddressCoinToAddress(crypto.AddressFromB58String(v.(string))))
			if t == evmutils.ZeroAddress {
				return "", errors.New("addrs 地址类型错误")
			}
			addrs = append(addrs, t)
		}

		comment = ens.BuildDelOperatorInput(addrs)

	// 设置折扣
	case "ResetTiers":
		if !TypeConversion("bool", jsonData["isRenew"]) {
			return "", errors.New("forever 类型错误")
		}
		if !TypeConversion("[]interface{}", jsonData["tiers"]) {
			return "", errors.New("tiers 类型错误")
		}
		isRenew := jsonData["isRenew"].(bool)
		tiersItr := jsonData["tiers"].([]interface{})
		tiers := make([]ens.RateTier, 0)
		for _, v := range tiersItr {
			tempv := v.([]interface{})
			tiers = append(tiers, ens.RateTier{Level: uint16(tempv[0].(float64)), Rate: uint16(tempv[1].(float64))})
		}
		comment = ens.BuildSetTiers(isRenew, tiers)
	}

	return

}

/*
获取MutlDeal
*/
func MultDeal(tag string, jsonData map[string]interface{}, keyStorePath, srcaddress, address, pwd string, amount, gas, frozenHeight, gasPrice uint64, nonce uint64, currentHeight uint64, domain string, domainType uint64) (hash string, err error) {

	// 获取comment
	comment, err := GetComment(tag, jsonData)
	if err != nil {
		return
	}

	// 构建离线交易
	tx, hash, _, err := CreateOfflineContractTx(keyStorePath, srcaddress, address, pwd, comment, amount, gas, frozenHeight, gasPrice, nonce, currentHeight, domain, domainType, "", "")
	if err != nil {
		return
	}

	// pushContractTx
	_, err = PushContractTx(tx)
	if err != nil {
		return
	}

	return
}

/*
PushContractTx
*/
func PushContractTx(txjson string) (res []byte, err error) {

	txjsonBs, e := base64.StdEncoding.DecodeString(txjson)
	if e != nil {
		engine.Log.Info("DecodeString fail:%s", e.Error())
		return
	}

	// mining.ParseTxBaseProto()

	// engine.Log.Info("txjson:%s", string(txjsonBs))
	txItr, e := ParseTxBaseProto(0, &txjsonBs)
	// txItr, err := mining.ParseTxBase(0, &txjsonBs)
	if e != nil {
		engine.Log.Info("ParseTxBaseProto fail:%s", e.Error())
		return
	}
	// engine.Log.Info("rpc transaction received %s", hex.EncodeToString(*txItr.GetHash()))
	if e := txItr.CheckSign(); e != nil {
		engine.Log.Info("transaction check fail:%s", e.Error())
		return
	}

	tx := txItr.(*Tx_Contract)
	//验证合约
	gasUsed, err := tx.PreExecNew()
	if err != nil {
		engine.Log.Error("sssssssssssssssss %s", err.Error())
		return nil, err
	}
	tx.GasUsed = gasUsed

	tx.BuildHash()

	e = AddTx(txItr)
	if e != nil {
		return
	}
	return
}

/*
类型断言
*/
func TypeConversion(t string, i interface{}) bool {
	a := ""
	switch i.(type) {
	case string:
		a = "string"
	case float64:
		a = "float64"
	case bool:
		a = "bool"
	case []interface{}:
		a = "[]interface{}"
	case []string:
		a = "[]string"
	default:
		a = "no type error"
	}
	if a != t {
		return false
	}
	return true
}

/*
*
创建离线奖励交易
*/
func CreateOfflineRewardTx(keyStorePath, srcaddress, address, voteTag string, voteType, rate uint16, pwd, comment string, amount, gas, frozenHeight uint64, nonce uint64, currentHeight uint64, domain string, domainType uint64) (tx, hash string, err error) {

	srcaddr := srcaddress
	src := crypto.AddressFromB58String(srcaddr)
	//判断地址前缀是否正确
	//if !crypto.ValidAddr(config.AddrPre, src) {
	//	return
	//}

	addr := address
	dst := crypto.AddressFromB58String(addr)
	//if !crypto.ValidAddr(config.AddrPre, dst) {
	//	return
	//}

	key := InitKeyStore(keyStorePath, pwd)

	var paybs *[]byte
	switch voteTag {
	case "votein":
		txpay, err := CreateTxRewardVoteinPay(key, &src, &dst, voteType, rate, amount, gas, frozenHeight, pwd, comment, nonce, currentHeight, domain, domainType)
		if err != nil {
			return "", "", err
		}
		paybs, _ = txpay.Proto()
		hash = fmt.Sprintf("%x", txpay.Hash)
	case "voteout":
		txpay, err := CreateTxRewardVoteoutPay(key, &src, &dst, voteType, rate, amount, gas, frozenHeight, pwd, comment, nonce, currentHeight, domain, domainType)
		if err != nil {
			return "", "", err
		}
		paybs, _ = txpay.Proto()
		hash = fmt.Sprintf("%x", txpay.Hash)
	case "reward":
		txpay, err := CreateTxReward(key, &src, &dst, voteType, rate, amount, gas, frozenHeight, pwd, comment, nonce, currentHeight, domain, domainType)
		if err != nil {
			return "", "", err
		}
		paybs, _ = txpay.Proto()
		hash = fmt.Sprintf("%x", txpay.Hash)
	}

	//engine.Log.Error("交易hash: %s", fmt.Sprintf("%x", txpay.Hash))

	//data := make(map[string]interface{})
	//data["tx"] =
	tx = base64.StdEncoding.EncodeToString(*paybs)
	return

}

/*
创建一个votein交易
*/
func CreateTxRewardVoteinPay(key *keystore.Keystore, srcAddress, address *crypto.AddressCoin, voteType, rate uint16, amount, gas, frozenHeight uint64, pwd, comment string, nonceInt uint64, currentHeight uint64, domain string, domainType uint64) (*Tx_vote_in, error) {
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	//查找余额
	vins := make([]*Vin, 0)

	puk, ok := key.GetPukByAddr(*srcAddress)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := big.NewInt(int64(nonceInt))

	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value: amount, //输出金额 = 实际金额 * 100000000
		//Address:      *address,     //钱包地址
		Address:      *srcAddress,  //钱包地址
		FrozenHeight: frozenHeight, //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	var pay *Tx_vote_in
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_vote_in,               //交易类型
		Vin_total:  uint64(len(vins)),                           //输入交易数量
		Vin:        vins,                                        //交易输入
		Vout_total: uint64(len(vouts)),                          //输出交易数量
		Vout:       vouts,                                       //交易输出
		Gas:        gas,                                         //交易手续费
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &Tx_vote_in{
		TxBase:   base,
		Vote:     *address,
		VoteType: voteType,
		Rate:     rate,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return nil, err
		}
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()
	//}
	return pay, nil
}

/*
创建一个voteout交易
*/
func CreateTxRewardVoteoutPay(key *keystore.Keystore, srcAddress, address *crypto.AddressCoin, voteType, rate uint16, amount, gas, frozenHeight uint64, pwd, comment string, nonceInt uint64, currentHeight uint64, domain string, domainType uint64) (*Tx_vote_out, error) {
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	//查找余额
	vins := make([]*Vin, 0)
	puk, ok := key.GetPukByAddr(*srcAddress)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := big.NewInt(int64(nonceInt))

	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//engine.Log.Info("取消金额:%d", amount)
	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value: amount, //输出金额 = 实际金额 * 100000000
		//Address:      *address,     //钱包地址
		Address:      *srcAddress,  //钱包地址
		FrozenHeight: frozenHeight, //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	var pay *Tx_vote_out
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_vote_out,              //交易类型
		Vin_total:  uint64(len(vins)),                           //输入交易数量
		Vin:        vins,                                        //交易输入
		Vout_total: uint64(len(vouts)),                          //输出交易数量
		Vout:       vouts,                                       //交易输出
		Gas:        gas,                                         //交易手续费
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &Tx_vote_out{
		TxBase:   base,
		Vote:     *address, //见证人地址
		VoteType: voteType,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return nil, err
		}
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()
	//}
	return pay, nil
}

/*
创建社区/轻节点手动分配奖励的交易
*/
func CreateTxReward(key *keystore.Keystore, srcAddress, address *crypto.AddressCoin, voteType, rate uint16, amount, gas, frozenHeight uint64, pwd, comment string, nonceInt uint64, currentHeight uint64, domain string, domainType uint64) (*Tx_Vote_Reward, error) {
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	//查找余额
	vins := make([]*Vin, 0)
	puk, ok := key.GetPukByAddr(*srcAddress)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := big.NewInt(int64(nonceInt))

	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//engine.Log.Info("取消金额:%d", amount)
	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value: amount, //输出金额 = 实际金额 * 100000000
		//Address:      *address,     //钱包地址
		Address:      *srcAddress,  //钱包地址
		FrozenHeight: frozenHeight, //
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)

	var pay *Tx_Vote_Reward
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_voting_reward,         //交易类型
		Vin_total:  uint64(len(vins)),                           //输入交易数量
		Vin:        vins,                                        //交易输入
		Vout_total: uint64(len(vouts)),                          //输出交易数量
		Vout:       vouts,                                       //交易输出
		Gas:        gas,                                         //交易手续费
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &Tx_Vote_Reward{
		TxBase: base,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return nil, err
		}
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()
	//}
	return pay, nil
}

func CreateOfflineTxBuildAddressBind(keyStorePath, srcaddress, bindaddress string, bindType, gas, frozenHeight, nonceInt, currentHeight uint64, pwd string, domain string, domainType uint64, comment string) (string, string, error) {
	srcAddress := crypto.AddressFromB58String(srcaddress)
	bindAddress := crypto.AddressFromB58String(bindaddress)

	key := InitKeyStore(keyStorePath, pwd)

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	//查找余额
	vins := make([]*Vin, 0)

	puk, ok := key.GetPukByAddr(srcAddress)
	if !ok {
		return "", "", config.ERROR_public_key_not_exist
	}

	nonce := big.NewInt(int64(nonceInt))
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        0, //输出金额 = 实际金额 * 100000000
		Address:      bindAddress,
		FrozenHeight: frozenHeight,
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)
	var pay *TxAddressBind

	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_address_bind,          //交易类型
		Vin_total:  uint64(len(vins)),                           //输入交易数量
		Vin:        vins,                                        //交易输入
		Vout_total: uint64(len(vouts)),                          //输出交易数量
		Vout:       vouts,                                       //交易输出
		Gas:        gas,                                         //交易手续费
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &TxAddressBind{
		TxBase:   base,
		BindType: bindType,
		BindAddr: bindAddress,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return "", "", err
		}
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()

	paybs, _ := pay.Proto()
	return base64.StdEncoding.EncodeToString(*paybs), fmt.Sprintf("%x", pay.Hash), nil
}

func CreateOfflineTxBuildAddressTransfer(keyStorePath, srcaddress, address, payaddress string, amount, gas, frozenHeight, nonceInt, currentHeight uint64, pwd string, domain string, domainType uint64, comment string) (string, string, error) {
	//if !db.CheckAddressBind(*srcAddress, *payAddress) {
	//	return nil, errors.New("bind address mismatch")
	//}
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	srcAddress := crypto.AddressFromB58String(srcaddress)
	addr := crypto.AddressFromB58String(address)
	payAddress := crypto.AddressFromB58String(payaddress)

	key := InitKeyStore(keyStorePath, pwd)

	//查找余额
	vins := make([]*Vin, 0)

	puk, ok := key.GetPukByAddr(srcAddress)
	if !ok {
		return "", "", config.ERROR_public_key_not_exist
	}

	nonce := big.NewInt(int64(nonceInt))
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        amount, //输出金额 = 实际金额 * 100000000
		Address:      addr,
		FrozenHeight: frozenHeight,
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)
	var pay *TxAddressTransfer

	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_address_transfer,      //交易类型
		Vin_total:  uint64(len(vins)),                           //输入交易数量
		Vin:        vins,                                        //交易输入
		Vout_total: uint64(len(vouts)),                          //输出交易数量
		Vout:       vouts,                                       //交易输出
		Gas:        gas,                                         //交易手续费
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &TxAddressTransfer{
		TxBase:     base,
		PayAddress: payAddress,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return "", "", err
		}
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()
	paybs, _ := pay.Proto()
	return base64.StdEncoding.EncodeToString(*paybs), fmt.Sprintf("%x", pay.Hash), nil
}

func CreateOfflineTxBuildAddressFrozen(keyStorePath, srcaddress, frozenaddress string, frozenType, gas, frozenHeight, nonceInt, currentHeight uint64, pwd string, domain string, domainType uint64, comment string) (string, string, error) {
	//if !db.CheckAddressBind(*srcAddress, *frozenAddress) {
	//	return nil, errors.New("bind address mismatch")
	//}

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	srcAddress := crypto.AddressFromB58String(srcaddress)
	frozenAddress := crypto.AddressFromB58String(frozenaddress)
	key := InitKeyStore(keyStorePath, pwd)

	//查找余额
	vins := make([]*Vin, 0)

	puk, ok := key.GetPukByAddr(srcAddress)
	if !ok {
		return "", "", config.ERROR_public_key_not_exist
	}

	nonce := big.NewInt(int64(nonceInt))
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Nonce: *new(big.Int).Add(nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        0, //输出金额 = 实际金额 * 100000000
		Address:      frozenAddress,
		FrozenHeight: frozenHeight,
		Domain:       []byte(domain),
		DomainType:   domainType,
	}
	vouts = append(vouts, &vout)
	var pay *TxAddressFrozen

	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_address_frozen,        //交易类型
		Vin_total:  uint64(len(vins)),                           //输入交易数量
		Vin:        vins,                                        //交易输入
		Vout_total: uint64(len(vouts)),                          //输出交易数量
		Vout:       vouts,                                       //交易输出
		Gas:        gas,                                         //交易手续费
		LockHeight: currentHeight + config.Wallet_tx_lockHeight, //锁定高度
		Payload:    commentbs,                                   //
		Comment:    []byte{},
	}
	pay = &TxAddressFrozen{
		TxBase:     base,
		FrozenType: frozenType,
		FrozenAddr: frozenAddress,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		_, prk, err := key.GetKeyByPuk(one.Puk, pwd)
		if err != nil {
			return "", "", err
		}
		sign := pay.GetSign(&prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}

	pay.BuildHash()

	paybs, _ := pay.Proto()
	return base64.StdEncoding.EncodeToString(*paybs), fmt.Sprintf("%x", pay.Hash), nil
}

// 支付类检查是否免gas
func CheckTxPayFreeGas(txItr TxItr) error {
	amount := uint64(0)
	for _, one := range *txItr.GetVout() {
		amount += one.Value
	}

	gas := txItr.GetGas()
	comment := string(txItr.GetComment())
	vin := (*txItr.GetVin())[0]
	src := vin.GetPukToAddr()
	return CheckTxPayFreeGasWithParams(txItr.Class(), *src, amount, gas, comment)
}

// 支付类检查是否免gas
func CheckTxPayFreeGasWithParams(class uint64, src crypto.AddressCoin, amount, gas uint64, comment string) error {
	if class == config.Wallet_tx_type_pay || class == config.Wallet_tx_type_multsign_pay || class == config.Wallet_tx_type_address_transfer {
		if value, ok := GetLongChain().Balance.freeGasAddrSet.Load(utils.Bytes2string(src)); ok {
			item := value.(*DepositFreeGasItem)
			if item.LimitCount > 0 && item.LimitHeight >= GetLongChain().CurrentBlock {
				return nil
			}
		}
		runeLength := len([]rune(comment))
		if runeLength > 1024 {
			return errors.New("comment")
		}
		if config.EnableFreeGas {
			if amount < config.Wallet_tx_free_gas_min_amount { //<min_amount不免gas费
				temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(int64(config.Wallet_tx_gas_min)))
				temp = new(big.Int).Div(temp, big.NewInt(1024))
				if gas < config.Wallet_tx_gas_min || gas < temp.Uint64() {
					return errors.New("gas too little")
				}
			}
		} else {
			temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(int64(config.Wallet_tx_gas_min)))
			temp = new(big.Int).Div(temp, big.NewInt(1024))
			if gas < config.Wallet_tx_gas_min || gas < temp.Uint64() {
				return errors.New("gas too little")
			}
		}
	} else {
		if gas < config.Wallet_tx_gas_min {
			return errors.New("gas too little")
		}
		return nil
	}
	return nil
}
