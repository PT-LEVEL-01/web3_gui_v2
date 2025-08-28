/*
出块奖励
*/
package mining

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
出块奖励交易
*/
type Tx_reward struct {
	TxBase
	//Index     uint64 //前index见证人数量
	AllReward uint64 //区块总奖励
	Bloom     []byte `json:"bloom"` //bloom过滤器
}

/*
出块奖励交易
*/
type Tx_reward_VO struct {
	TxBaseVO
	//Index uint64 `json:"Index"` //前index见证人数量
	AllReward uint64 //区块总奖励
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_reward) GetVOJSON() interface{} {
	txReward := Tx_reward_VO{
		TxBaseVO: this.ConversionVO(),
		//Index:    this.Index, //前index见证人数量
		AllReward: this.AllReward,
	}
	return txReward
}

/*
转化为VO对象
*/
func (this *Tx_reward) ConversionVO() TxBaseVO {
	vins := make([]*VinVO, 0)
	for _, one := range this.Vin {
		vins = append(vins, one.ConversionVO())
	}

	vouts := make([]*VoutVO, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, one.ConversionVO())
	}

	return TxBaseVO{
		Hash:       hex.EncodeToString(this.Hash),      //本交易hash，不参与区块hash，只用来保存
		Type:       this.Type,                          //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  this.Vin_total,                     //输入交易数量
		Vin:        vins,                               //交易输入
		Vout_total: this.Vout_total,                    //输出交易数量
		Vout:       vouts,                              //交易输出
		Gas:        this.Gas,                           //交易手续费，此字段不参与交易hash
		LockHeight: this.LockHeight,                    //本交易锁定在低于这个高度的块中，超过这个高度，块将不被打包到区块中。
		Payload:    string(this.Payload),               //备注信息
		BlockHash:  hex.EncodeToString(this.BlockHash), //本交易属于的区块hash，不参与区块hash，只用来保存
		Reward:     this.AllReward,
		Comment:    string(this.Comment), //备注信息
	}
}

/*
构建hash值得到交易id
*/
func (this *Tx_reward) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_mining)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_reward) Proto() (*[]byte, error) {
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
	txPay := go_protos.TxReward{
		TxBase: &txBase,
		//Index:  this.Index,
		AllReward: this.AllReward,
	}
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
序列化
*/
func (this *Tx_reward) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	//buf.Write(utils.Uint64ToBytes(this.Index))
	buf.Write(utils.Uint64ToBytes(this.AllReward))
	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *Tx_reward) GetWaitSign(vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	//*signDst = append(*signDst, utils.Uint64ToBytes(this.Index)...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.AllReward)...)
	return signDst
}

/*
获取签名
*/
func (this *Tx_reward) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	//*signDst = append(*signDst, utils.Uint64ToBytes(this.Index)...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.AllReward)...)
	sign := keystore.Sign(*key, *signDst)
	return &sign
}

/*
检查交易是否合法
*/
func (this *Tx_reward) CheckSign() error {
	// start := config.TimeNow()
	// engine.Log.Info("开始验证交易合法性 Tx_reward")
	//检查输入输出是否对等，还有手续费
	if this.Vin == nil || len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) != 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}

	one := this.Vin[0]
	signDst := this.GetSignSerialize(nil, uint64(0))
	//*signDst = append(*signDst, utils.Uint64ToBytes(this.Index)...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.AllReward)...)
	// engine.Log.Info("开始验证交易合法性 Tx_reward 2222222222222222 %s", config.TimeNow().Sub(start))
	puk := ed25519.PublicKey(one.Puk)
	// engine.Log.Info("开始验证交易合法性 Tx_reward 3333333333333333 %s", config.TimeNow().Sub(start))
	if config.Wallet_print_serialize_hex {
		engine.Log.Info("sign serialize:%s", hex.EncodeToString(*signDst))
	}
	if !ed25519.Verify(puk, *signDst, one.Sign) {
		return config.ERROR_sign_fail
	}

	// engine.Log.Info("开始验证交易合法性 Tx_reward 4444444444444444 %s", config.TimeNow().Sub(start))
	outTotal := uint64(0)
	for _, one := range this.Vout {
		outTotal = outTotal + one.Value
	}

	return nil
}

/*
获取本交易总共花费的余额
*/
func (this *Tx_reward) GetSpend() uint64 {
	spend := this.Gas
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_reward) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

/*
	统计交易余额
*/
// func (this *Tx_reward) CountTxItems(height uint64) *TxItemCount {
// 	itemCount := TxItemCount{
// 		Additems: make([]*TxItem, 0),
// 		SubItems: make([]*TxSubItems, 0),
// 	}
// 	//将之前的UTXO标记为已经使用，余额中减去。
// 	// for _, vin := range this.Vin {
// 	// 	// engine.Log.Info("查看vin中的状态 %d", vin.PukIsSelf)
// 	// ok := vin.CheckIsSelf()
// 	// if !ok {
// 	// 	continue
// 	// }
// 	// 	// engine.Log.Info("统单易1耗时 %s %s", txItr.GetHashStr(), config.TimeNow().Sub(start))
// 	// 	//查找这个地址的余额列表，没有则创建一个
// 	// 	itemCount.SubItems = append(itemCount.SubItems, &TxSubItems{
// 	// 		Txid:      vin.Txid, //utils.Bytes2string(vin.Txid), //  vin.GetTxidStr(),
// 	// 		VoutIndex: vin.Vout,
// 	// 		Addr:      *vin.GetPukToAddr(), // utils.Bytes2string(*vin.GetPukToAddr()), // vin.GetPukToAddrStr(),
// 	// 	})
// 	// }

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
仅处理创世块交易,奖励统计查看distributeReward方法
*/
func (this *Tx_reward) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)),
		Nonce:    make(map[string]big.Int),
	}

	//处理创世块交易
	if height == config.Mining_block_start_height {
		// totalValue := this.Gas
		for _, vout := range this.Vout {
			// totalValue += vout.Value
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
	}

	//余额中减去。
	// from := this.Vin[0].GetPukToAddr()
	// frozenMap, ok := itemCount.AddItems[utils.Bytes2string(*from)]
	// if ok {
	// 	oldValue, ok := (*frozenMap)[0]
	// 	if ok {
	// 		oldValue -= int64(totalValue)
	// 		(*frozenMap)[0] = oldValue
	// 	} else {
	// 		(*frozenMap)[0] = (0 - int64(totalValue))
	// 	}
	// } else {
	// 	frozenMap := make(map[uint64]int64, 0)
	// 	frozenMap[0] = (0 - int64(totalValue))
	// 	itemCount.AddItems[utils.Bytes2string(*from)] = &frozenMap
	// }
	return &itemCount
}

func (this *Tx_reward) CountTxHistory(height uint64) {
	//转入历史记录
	hiIn := HistoryItem{
		IsIn:    true,                           //资金转入转出方向，true=转入;false=转出;
		Type:    this.Class(),                   //交易类型
		InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
		OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
		// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
		// Value:  amount,          //交易金额
		Txid:   *this.GetHash(), //交易id
		Height: height,          //
		// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
	}

	//生成新的UTXO收益，保存到列表中
	for _, vout := range this.Vout {
		_, ok := Area.Keystore.FindAddress(vout.Address)
		if !ok {
			continue
		}
		//hiIn.Value += vout.Value
		hiIn.OutAddr = append(hiIn.OutAddr, &vout.Address)
	}
	if len(hiIn.OutAddr) > 0 {
		balanceHistoryManager.Add(hiIn)
	}
}

// 获取bloom
func (this *Tx_reward) GetBloom() []byte {
	key := append(config.DBKEY_tx_bloom, this.Hash...)
	bs, err := db.LevelDB.Find(key)
	if err != nil {
		return nil
	}
	return *bs
}

// 设置bloom
func (this *Tx_reward) SetBloom(bs []byte) {
	key := append(config.DBKEY_tx_bloom, this.Hash...)
	db.LevelDB.Save(key, &bs)
}
