package mining

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
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
多签交易
*/
type Tx_Multsign_Pay struct {
	TxBase
	DefaultMultTx
}

/*
多签交易
*/
type Tx_Multsign_Pay_VO struct {
	TxBaseVO
	MultAddress string   `json:"mult_address"` //多签地址
	MultVins    []*VinVO `json:"mult_vins"`    //多签集合,保证顺序,验证 2f/3 + 1
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_Multsign_Pay) GetVOJSON() interface{} {
	multVins := []*VinVO{}
	for _, vin := range this.GetMultVins() {
		multVins = append(multVins, vin.ConversionVO())
	}
	multAddress := this.GetMultAddress()
	txMultsign := Tx_Multsign_Pay_VO{
		TxBaseVO:    this.ConversionVO(),
		MultAddress: multAddress.B58String(),
		MultVins:    multVins,
	}
	return txMultsign
}

/*
转化为VO对象
*/
func (this *Tx_Multsign_Pay) ConversionVO() TxBaseVO {
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
		Comment:    string(this.Comment),               //备注信息
	}
}

/*
构建hash值得到交易id
*/
func (this *Tx_Multsign_Pay) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_multsign_pay)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_Multsign_Pay) Proto() (*[]byte, error) {
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
	multVins := []*go_protos.Vin{}
	for _, one := range this.GetMultVins() {
		multVins = append(multVins, &go_protos.Vin{
			Puk:  one.Puk,
			Sign: one.Sign,
		})
	}
	txPay := go_protos.TxMultsignPay{
		TxBase:      &txBase,
		MultAddress: this.GetMultAddress(),
		MultVins:    multVins,
	}
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
转换结构
*/
func (this *Tx_Multsign_Pay) ConvertFrom(txProto *go_protos.TxMultsignPay) error {
	if txProto.TxBase.Type != config.Wallet_tx_type_multsign_pay {
		return errors.New("tx type error")
	}
	vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
	for _, one := range txProto.TxBase.Vin {
		nonce := new(big.Int).SetBytes(one.Nonce)
		vins = append(vins, &Vin{
			Puk:   one.Puk,
			Sign:  one.Sign,
			Nonce: *nonce,
		})
	}
	vouts := make([]*Vout, 0, len(txProto.TxBase.Vout))
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
	txBase.GasUsed = txProto.TxBase.GasUsed
	txBase.Comment = txProto.TxBase.Comment

	multVins := []*Vin{}
	for _, one := range txProto.MultVins {
		nonce := new(big.Int).SetBytes(one.Nonce)
		multVins = append(multVins, &Vin{
			Puk:   one.Puk,
			Sign:  one.Sign,
			Nonce: *nonce,
		})
	}

	this.TxBase = txBase
	this.DefaultMultTx = NewDefaultMultTxImpl(txProto.MultAddress, multVins)
	return nil
}

/*
序列化
*/
func (this *Tx_Multsign_Pay) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(this.GetMultAddress())
	for _, vin := range this.GetMultVins() {
		buf.Write(vin.Puk)
	}
	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *Tx_Multsign_Pay) GetWaitSign(vinIndex uint64) *[]byte {
	buf := bytes.NewBuffer(*this.GetSignSerialize(nil, 0))
	buf.Write(this.GetMultAddress())
	for _, vin := range this.GetMultVins() {
		buf.Write(vin.Puk)
	}
	buf.Write(utils.Uint64ToBytes(vinIndex))
	bs := buf.Bytes()
	return &bs
}

/*
获取签名
*/
func (this *Tx_Multsign_Pay) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	buf := this.GetWaitSign(vinIndex)
	sign := keystore.Sign(*key, *buf)
	return &sign
}

/*
检查交易是否合法
*/
func (this *Tx_Multsign_Pay) CheckSign() error {
	// start := config.TimeNow()
	// engine.Log.Info("开始验证交易合法性 Tx_Multsign_Pay")
	//检查输入输出是否对等，还有手续费
	if this.Vin == nil || len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) == 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}

	verifyCount := 0
	total := len(this.GetMultVins())
	for i, vin := range this.GetMultVins() {
		puk := ed25519.PublicKey(vin.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(this.Hash))
		}
		bs := this.GetWaitSign(uint64(i))
		if ed25519.Verify(puk, *bs, vin.Sign) {
			verifyCount++
		}
	}

	//验证通过的数量>=2f/3 + 1; 即验证通过
	if !(verifyCount >= config.MultsignMajorityPrinciple(total)) {
		engine.Log.Warn("Check Mult-Signature: %s Total:%d Pass:%d", hex.EncodeToString(this.Hash), total, verifyCount)
		return config.ERROR_check_sign_not_pass
	} else {
		engine.Log.Warn("Check Mult-Signature: %s Total:%d Pass:%d", hex.EncodeToString(this.Hash), total, verifyCount)
	}

	return nil
}

/*
获取本交易总共花费的余额
*/
func (this *Tx_Multsign_Pay) GetSpend() uint64 {
	spend := this.Gas
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_Multsign_Pay) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

// 检查锁定高度
func (this *Tx_Multsign_Pay) CheckLockHeight(lockHeight uint64) error {
	if this.GetLockHeight() < lockHeight {
		//处理多签
		key := config.BuildMultsignRequestTx(this.Hash)
		db.LevelDB.Remove(key)
		return config.ERROR_tx_lockheight
	}
	return nil
}

/*
统计交易余额
仅处理创世块交易,奖励统计查看distributeReward方法
*/
func (this *Tx_Multsign_Pay) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}

	//处理多签燃料费
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

func (this *Tx_Multsign_Pay) CountTxHistory(height uint64) {
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

// 创建多签请求签名交易,仅广播
func BuildRequestMultsignPayTx(multAddress, address crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd string, comment string) (*Tx_Multsign_Pay, error) {
	chain := forks.GetLongChain()
	currentHeight := chain.GetCurrentBlock()

	//查找余额
	vins := make([]*Vin, 0)

	total, _ := chain.Balance.BuildPayVinNew(&multAddress, amount)
	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}

	//查询多签
	pukSet, ok := GetMultsignAddrSet(multAddress)
	if !ok {
		return nil, config.ERROR_tx_multaddr_not_found
	}

	sumpuk, multVins := GetMultAddrPublicKey(pukSet)
	nonce := chain.GetBalance().FindNonce(&multAddress)
	vin := Vin{
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   sumpuk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      address,      //钱包地址
		FrozenHeight: frozenHeight, //
	}
	vouts = append(vouts, &vout)

	var mtx *Tx_Multsign_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_multsign_pay,                       //交易类型
			Vin_total:  uint64(len(vins)),                                        //输入交易数量
			Vin:        vins,                                                     //交易输入
			Vout_total: uint64(len(vouts)),                                       //输出交易数量
			Vout:       vouts,                                                    //交易输出
			Gas:        gas,                                                      //交易手续费
			LockHeight: currentHeight + config.Wallet_multsign_tx_lockHeight + i, //锁定高度
			Payload:    []byte{},                                                 //
			Comment:    []byte(comment),
		}
		mtx = &Tx_Multsign_Pay{
			TxBase:        base,
			DefaultMultTx: NewDefaultMultTxImpl(multAddress, multVins),
		}

		//若自己是其中之一,就先签名
		for j, one := range mtx.GetMultVins() {
			if _, prk, err := Area.Keystore.GetKeyByPuk(one.Puk, pwd); err == nil {
				sign := mtx.GetSign(&prk, uint64(j))
				mtx.GetMultVins()[j].Sign = *sign
			}
		}

		mtx.BuildHash()
		// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*mtx.GetHash()))
		if mtx.CheckHashExist() {
			mtx = nil
			continue
		} else {
			break
		}
	}

	//处理多签请求
	if err := mtx.DefaultMultTx.RequestMultTx(mtx); err != nil {
		engine.Log.Warn("Multsign Request: %s %s", hex.EncodeToString(*mtx.GetHash()), err.Error())
		return nil, err
	}

	return mtx, nil
}
