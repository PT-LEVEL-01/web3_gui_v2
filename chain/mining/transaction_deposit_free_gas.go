package mining

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
质押免gas费
*/
type Tx_DepositFreeGas struct {
	TxBase
	DepositAddress crypto.AddressCoin `json:"d_a"` //质押地址
	Deposit        uint64             `json:"d"`   //质押金额
}

/*
质押免gas费
*/
type Tx_DepositFreeGas_VO struct {
	TxBaseVO
	DepositAddress string `json:"deposit_address"` //质押地址
	Deposit        uint64 `json:"deposit"`         //质押金额
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_DepositFreeGas) GetVOJSON() interface{} {
	txv0 := Tx_DepositFreeGas_VO{
		TxBaseVO:       this.ConversionVO(),
		DepositAddress: this.DepositAddress.B58String(),
		Deposit:        this.Deposit,
	}
	return txv0
}

/*
转化为VO对象
*/
func (this *Tx_DepositFreeGas) ConversionVO() TxBaseVO {
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
func (this *Tx_DepositFreeGas) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_deposit_free_gas)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_DepositFreeGas) Proto() (*[]byte, error) {
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
	txPay := go_protos.TxDepositFreeGas{
		TxBase:         &txBase,
		DepositAddress: this.DepositAddress,
		Deposit:        this.Deposit,
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
func (this *Tx_DepositFreeGas) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(this.DepositAddress)
	buf.Write(utils.Uint64ToBytes(this.Deposit))
	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *Tx_DepositFreeGas) GetWaitSign(vinIndex uint64) *[]byte {
	buf := bytes.NewBuffer(*this.GetSignSerialize(nil, 0))
	buf.Write(this.DepositAddress)
	buf.Write(utils.Uint64ToBytes(this.Deposit))
	bs := buf.Bytes()
	return &bs
}

/*
获取签名
*/
func (this *Tx_DepositFreeGas) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	buf := this.GetWaitSign(vinIndex)
	sign := keystore.Sign(*key, *buf)
	return &sign
}

/*
检查交易是否合法
*/
func (this *Tx_DepositFreeGas) CheckSign() error {
	// start := config.TimeNow()
	// engine.Log.Info("开始验证交易合法性 Tx_DepositFreeGas")
	//检查输入输出是否对等，还有手续费
	if this.Vin == nil || len(this.Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) == 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}

	for i, one := range this.Vin {
		bs := this.GetWaitSign(uint64(i))
		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*bs))
		}

		if !ed25519.Verify(puk, *bs, one.Sign) {
			return config.ERROR_sign_fail
		}
	}

	return nil
}

/*
获取本交易总共花费的余额
*/
func (this *Tx_DepositFreeGas) GetSpend() uint64 {
	spend := this.Gas
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_DepositFreeGas) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

/*
统计交易余额
*/
func (this *Tx_DepositFreeGas) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}

	totalValue := this.Gas + this.Deposit

	//余额中减去。
	from := this.Vin[0].GetPukToAddr()
	itemCount.Nonce[utils.Bytes2string(*from)] = this.Vin[0].Nonce

	frozenMap := make(map[uint64]int64, 0)
	frozenMap[0] = (0 - int64(totalValue))
	itemCount.AddItems[utils.Bytes2string(*from)] = &frozenMap

	return &itemCount
}

func (this *Tx_DepositFreeGas) CountTxHistory(height uint64) {
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

/*
转换结构
*/
func (this *Tx_DepositFreeGas) ConvertFrom(txProto *go_protos.TxDepositFreeGas) error {
	if txProto.TxBase.Type != config.Wallet_tx_type_deposit_free_gas {
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

	this.TxBase = txBase
	this.DepositAddress = txProto.DepositAddress
	this.Deposit = txProto.Deposit
	return nil
}

// 创建质押交易
func BuildDepositFreeGasTx(src crypto.AddressCoin, gas uint64, pwd string, depositAddress crypto.AddressCoin, deposit uint64) (*Tx_DepositFreeGas, error) {
	chain := forks.GetLongChain()
	currentHeight := chain.GetCurrentBlock()

	if gas < config.Wallet_tx_gas_min {
		return nil, config.ERROR_tx_gas_too_little
	}

	vins := make([]*Vin, 0)

	total, item := chain.Balance.BuildPayVinNew(&src, deposit+gas)
	if total < deposit+gas {
		return nil, config.ERROR_not_enough
	}
	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
		Puk:   puk, //公钥
	}
	vins = append(vins, &vin)

	//构建交易输出
	vouts := make([]*Vout, 0)

	var pay *Tx_DepositFreeGas
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_deposit_free_gas, //交易类型
			Vin_total:  uint64(len(vins)),                      //输入交易数量
			Vin:        vins,                                   //交易输入
			Vout_total: uint64(len(vouts)),                     //输出交易数量
			Vout:       vouts,                                  //交易输出
			Gas:        gas,
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    []byte{},                                        //
			Comment:    []byte{},
		}
		pay = &Tx_DepositFreeGas{
			TxBase:         base,
			DepositAddress: depositAddress,
			Deposit:        deposit,
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
		// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}

	chain.Balance.AddLockTx(pay)

	// engine.Log.Info("create tx finish!")
	if err := chain.TransactionManager.AddTx(pay); err != nil {
		chain.Balance.DelLockTx(pay)
		return nil, errors.Wrap(err, "add tx fail!")
	}

	//广播
	MulticastTx(pay)

	return pay, nil
}
