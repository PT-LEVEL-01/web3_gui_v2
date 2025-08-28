package mining

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
转账交易
*/
type Tx_Vote_Reward struct {
	TxBase
	StartHeight uint64 //快照开始高度
	EndHeight   uint64 //快照结束高度
}

type Tx_Vote_Reward_VO struct {
	TxBaseVO
	StartHeight uint64 `json:"StartHeight"` //快照开始高度
	EndHeight   uint64 `json:"EndHeight"`   //快照结束高度
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_Vote_Reward) GetVOJSON() interface{} {
	return Tx_Vote_Reward_VO{
		TxBaseVO:    this.TxBase.ConversionVO(),
		StartHeight: this.StartHeight, //快照开始高度
		EndHeight:   this.EndHeight,   //快照结束高度
	}
}

/*
构建hash值得到交易id
*/
func (this *Tx_Vote_Reward) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_voting_reward)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_Vote_Reward) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0, len(this.Vin))
	for _, one := range this.Vin {
		vins = append(vins, &go_protos.Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
			Puk:   one.Puk,
			Sign:  one.Sign,
			Nonce: one.Nonce.Bytes(),
		})
	}
	vouts := make([]*go_protos.Vout, 0, len(this.Vout))
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
		GasUsed:    this.GasUsed,
		Comment:    this.Comment,
	}

	txPay := go_protos.TxVoteReward{
		TxBase:      &txBase,
		StartHeight: this.StartHeight,
		EndHeight:   this.EndHeight,
	}
	// txPay.Marshal()
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
格式化成json字符串
*/
func (this *Tx_Vote_Reward) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(utils.Uint64ToBytes(this.StartHeight))
	buf.Write(utils.Uint64ToBytes(this.EndHeight))
	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *Tx_Vote_Reward) GetWaitSign(vinIndex uint64) *[]byte {

	// txItr, err := LoadTxBase(txid)
	// // txItr, err := FindTxBase(txid)

	// // bs, err := db.Find(txid)
	// // if err != nil {
	// // 	return nil
	// // }
	// // txItr, err := ParseTxBase(ParseTxClass(txid), bs)
	// if err != nil {
	// 	return nil
	// }

	// blockhash, err := db.GetTxToBlockHash(&txid)
	// if err != nil || blockhash == nil {
	// 	return nil
	// }

	// voutBs := txItr.GetVoutSignSerialize(voutIndex)
	// signBs := make([]byte, 0, len(*blockhash)+len(*txItr.GetHash())+len(*voutBs))
	// signBs = append(signBs, *blockhash...)
	// signBs = append(signBs, *txItr.GetHash()...)
	// signBs = append(signBs, *voutBs...)

	// buf := bytes.NewBuffer(nil)
	// //上一个交易 所属的区块hash
	// buf.Write(*blockhash)
	// // buf.Write(*txItr.GetBlockHash())
	// //上一个交易的hash
	// buf.Write(*txItr.GetHash())
	// //上一个交易的指定输出序列化
	// buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	// //本交易类型输入输出数量等信息和所有输出
	// signBs := buf.Bytes()
	signDst := this.GetSignSerialize(nil, vinIndex)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.StartHeight)...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.EndHeight)...)

	//sign := keystore.Sign(*key, *signDst)
	return signDst

	// bs = txItr.GetVoutSignSerialize(voutIndex)
	// bs = this.GetSignSerialize(bs, vinIndex)
	// *bs = append(*bs, this.Vote...)

	// *bs = keystore.Sign(*key, *bs)
	// return bs
}

/*
获取签名
*/
func (this *Tx_Vote_Reward) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	// txItr, err := LoadTxBase(txid)
	// // txItr, err := FindTxBase(txid)

	// if err != nil {
	// 	return nil
	// }

	// blockhash, err := db.GetTxToBlockHash(&txid)
	// if err != nil || blockhash == nil {
	// 	return nil
	// }

	// if txItr.GetBlockHash() == nil {
	// 	txItr = GetRemoteTxAndSave(txid)
	// 	if txItr.GetBlockHash() == nil {
	// 		return nil
	// 	}
	// }

	// voutBs := txItr.GetVoutSignSerialize(voutIndex)
	// signBs := make([]byte, 0, len(*blockhash)+len(*txItr.GetHash())+len(*voutBs))
	// signBs = append(signBs, *blockhash...)
	// signBs = append(signBs, *txItr.GetHash()...)
	// signBs = append(signBs, *voutBs...)

	// buf := bytes.NewBuffer(nil)
	// //上一个交易 所属的区块hash
	// buf.Write(*blockhash)
	// // buf.Write(*txItr.GetBlockHash())
	// //上一个交易的hash
	// buf.Write(*txItr.GetHash())
	// //上一个交易的指定输出序列化
	// buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	// //本交易类型输入输出数量等信息和所有输出
	// signBs := buf.Bytes()
	signDst := this.GetSignSerialize(nil, vinIndex)

	*signDst = append(*signDst, utils.Uint64ToBytes(this.StartHeight)...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.EndHeight)...)

	sign := keystore.Sign(*key, *signDst)

	return &sign

	// bs = txItr.GetVoutSignSerialize(voutIndex)
	// bs = this.GetSignSerialize(bs, vinIndex)
	// *bs = append(*bs, this.Vote...)

	// *bs = keystore.Sign(*key, *bs)
	// return bs
}

/*
验证是否合法
*/
func (this *Tx_Vote_Reward) CheckSign() error {
	if this.Vin == nil || len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) == 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}
	if this.Vout_total > config.Mining_pay_vout_max {
		return config.ERROR_pay_vout_too_much
	}

	//for _, vout := range this.Vout {
	//	if vout.Value <= 0 {
	//		return config.ERROR_amount_zero
	//	}
	//}

	// fmt.Println("开始验证交易合法性 Tx_Vote_Reward")
	// if err := this.TxBase.CheckBase(); err != nil {
	// 	return err
	// }

	// return nil

	// fmt.Println("开始验证交易合法性 Tx_deposit_in")
	//判断vin是否太多
	// if len(this.Vin) > config.Mining_pay_vin_max {
	// 	return config.ERROR_pay_vin_too_much
	// }

	//扣款地址必须一样
	var puk *[]byte

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费
	// inTotal := uint64(0)
	for i, one := range this.Vin {
		if i == 0 {
			puk = &one.Puk
		} else {
			if !bytes.Equal(*puk, one.Puk) {
				return config.ERROR_vote_reward_addr_disunity
			}
		}

		// txItr, err := LoadTxBase(one.Txid)
		// if err != nil {
		// 	return config.ERROR_tx_format_fail
		// }

		// blockhash, err := db.GetTxToBlockHash(&one.Txid)
		// if err != nil || blockhash == nil {
		// 	return config.ERROR_tx_format_fail
		// }

		// vout := (*txItr.GetVout())[one.Vout]
		// //如果这个交易已经被使用，则验证不通过，否则会出现双花问题。
		// inTotal = inTotal + vout.Value

		// //验证公钥是否和地址对应
		// addr := crypto.BuildAddr(config.AddrPre, one.Puk)
		// if !bytes.Equal(addr, (*txItr.GetVout())[one.Vout].Address) {
		// 	return config.ERROR_public_and_addr_notMatch
		// }

		// voutBs := txItr.GetVoutSignSerialize(one.Vout)
		// signBs := make([]byte, 0, len(*blockhash)+len(*txItr.GetHash())+len(*voutBs))
		// signBs = append(signBs, *blockhash...)
		// signBs = append(signBs, *txItr.GetHash()...)
		// signBs = append(signBs, *voutBs...)

		//验证签名
		signDst := this.GetSignSerialize(nil, uint64(i))
		//本交易特有信息
		*signDst = append(*signDst, utils.Uint64ToBytes(this.StartHeight)...)
		*signDst = append(*signDst, utils.Uint64ToBytes(this.EndHeight)...)
		// fmt.Println("验证签名前的字节3", len(*signDst), hex.EncodeToString(*signDst))
		puk := ed25519.PublicKey(one.Puk)
		// fmt.Printf("sign后:puk:%x signdst:%x sign:%x", md5.Sum(puk), md5.Sum(*signDst), md5.Sum(one.Sign))
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*signDst))
		}
		if !ed25519.Verify(puk, *signDst, one.Sign) {
			return config.ERROR_sign_fail
		}

	}
	//判断输入输出是否相等
	// outTotal := uint64(0)
	// for _, one := range this.Vout {
	// 	outTotal = outTotal + one.Value
	// }
	// // fmt.Println("这里的手续费是否正确", outTotal, inTotal, this.Gas)
	// if outTotal > inTotal {
	// 	return config.ERROR_tx_fail
	// }
	// this.Gas = inTotal - outTotal

	return nil
}

/*
获取本交易总共花费的余额
*/
func (this *Tx_Vote_Reward) GetSpend() uint64 {
	spend := this.Gas
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_Vote_Reward) CheckRepeatedTx(txs ...TxItr) bool {
	//判断是否出现双花
	// return this.MultipleExpenditures(txs...)
	for _, txOne := range txs {
		if txOne.Class() != config.Wallet_tx_type_voting_reward {
			continue
		}
		if bytes.Equal(this.Vin[0].Puk, (*txOne.GetVin())[0].Puk) {
			return false
		}
	}

	return true
}

/*
	统计交易余额
*/
// func (this *Tx_Vote_Reward) CountTxItems(height uint64) *TxItemCount {
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
func (this *Tx_Vote_Reward) CountTxItemsNew(height uint64) *TxItemCountMap {
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

/*
创建给轻节点分配奖励的交易
*/
func CreateTxVoteReward(addr *crypto.AddressCoin, address []PayNumber, gas uint64, pwd string, startHeight, endHeight uint64) (*Tx_Vote_Reward, error) {
	engine.Log.Info("CreateTxVoteReward start")
	chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]*Vin, 0)

	total := GetCommunityVoteRewardFrozen(addr)
	_, lockValue := chain.GetBalance().FindLockTotalByAddr(addr)
	total -= lockValue
	// engine.Log.Info("社区节点各种余额:total:%d gas:%d amount:%d lockValue:%d", total, gas, amount, lockValue)

	// total, item := chain.Balance.BuildPayVinNew(&addr, gas)
	if total < gas+amount {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }
	// for _, item := range items {
	// engine.Log.Info("打印地址 %s", item.Addr.B58String())
	puk, ok := Area.Keystore.GetPukByAddr(*addr)
	if !ok {
		// engine.Log.Error("异常：未找到公钥")
		return nil, config.ERROR_public_key_not_exist
	}
	// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
	nonce := chain.GetBalance().FindNonce(addr)
	vin := Vin{
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
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

	var pay *Tx_Vote_Reward
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_voting_reward,             //交易类型
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
		var txItr TxItr = &Tx_Vote_Reward{
			TxBase:      base,
			StartHeight: startHeight,
			EndHeight:   endHeight,
		}

		//给payload签名
		// if cs != nil {
		// 	addr := crypto.BuildAddr(config.AddrPre, cs.Puk)
		// 	_, prk, _, err := keystore.GetKeyByAddr(addr, pwd)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	txItr = SignPayload(txItr, cs.Puk, prk, cs.StartHeight, cs.EndHeight)
		// }

		pay = txItr.(*Tx_Vote_Reward)

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
	engine.Log.Info("CreateTxVoteReward end hash:%s", hex.EncodeToString(*pay.GetHash()))
	return pay, nil
}

/*
创建社区/轻节点手动分配奖励的交易
*/
func CreateTxVoteRewardNew(addr *crypto.AddressCoin, gas uint64, pwd, comment string) (*Tx_Vote_Reward, error) {
	chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()
	amount := uint64(0)

	//查找余额
	vins := make([]*Vin, 0)

	total, item := chain.Balance.BuildPayVinNew(addr, amount+gas)
	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}

	//判断奖励>0
	var communityAddr *crypto.AddressCoin
	if commInfo := chain.Balance.GetDepositCommunity(addr); commInfo != nil {
		communityAddr = &commInfo.SelfAddr
	} else if lightInfo := chain.Balance.GetDepositVote(addr); lightInfo != nil {
		communityAddr = &lightInfo.WitnessAddr
	}

	//社区地址不为空,则触发分奖
	if communityAddr == nil {
		return nil, config.ERROR_has_no_reward
	}

	// 通过社区地址,计算社区和轻节点的奖励
	communityReward, _ := chain.Balance.CalculateCommunityRewardAndLightReward(*communityAddr)
	//社区奖励
	if communityReward.Cmp(big.NewInt(0)) == 0 {
		return nil, config.ERROR_has_no_reward
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
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)
	// }

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:   amount, //输出金额 = 实际金额 * 100000000
		Address: *addr,  //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	// if total > amount+gas {
	// 	vout := Vout{
	// 		Value:   total - amount - gas,       //输出金额 = 实际金额 * 100000000
	// 		Address: keystore.GetAddr()[0].Addr, //钱包地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	var pay *Tx_Vote_Reward
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_voting_reward,             //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			//Payload:    []byte{},
			// CreateTime: config.TimeNow().Unix(),         //创建时间
			Comment: []byte(comment),
		}
		// pay.CleanZeroVout()
		pay = &Tx_Vote_Reward{
			TxBase: base,
		}

		//给payload签名
		// if cs != nil {
		// 	addr := crypto.BuildAddr(config.AddrPre, cs.Puk)
		// 	_, prk, _, err := keystore.GetKeyByAddr(addr, pwd)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	txItr = SignPayload(txItr, cs.Puk, prk, cs.StartHeight, cs.EndHeight)
		// }

		//给输出签名，防篡改
		//for i, one := range *pay.GetVin() {
		//	_, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd)
		//	if err != nil {
		//		return nil, err
		//	}

		//	// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

		//	sign := pay.GetSign(&prk, uint64(i))
		//	//				sign := pay.GetVoutsSign(prk, uint64(i))
		//	pay.Vin[i].Sign = *sign
		//}
		//给输出签名，防篡改
		for i, one := range *pay.GetVin() {
			for _, key := range Area.Keystore.GetAddr() {

				puk, ok := Area.Keystore.GetPukByAddr(key.Addr)
				if !ok {
					// fmt.Println("签名出错 1111111111")
					//签名出错
					return nil, config.ERROR_get_sign_data_fail // errors.New("签名出错")
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := Area.Keystore.GetKeyByAddr(key.Addr, pwd)
					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						// fmt.Println("签名出错 2222222222222", err.Error())
						return nil, err
					}
					sign := pay.GetSign(&prk, uint64(i))
					//				sign := txin.GetVoutsSign(prk, uint64(i))
					pay.Vin[i].Sign = *sign
				}
			}
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
	//chain.Balance.AddLockTx(pay)

	MulticastTx(pay)
	if err := chain.GetBalance().txManager.AddTx(pay); err != nil {
		return nil, err
	}

	return pay, nil
}
