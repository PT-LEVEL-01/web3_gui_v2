package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

func init() {
	tpc := new(TokenPublishController)
	RegisterTransaction(config.Wallet_tx_type_token_publish, tpc)
}

/*
交押金，注册一个域名
域名为抢注式，只需要一点点手续费就能注册。注册时间为一年到期，快要到期时续费。
注册域名需要押金，押金不管多少。
域名可以转让，可以修改解析的网络地址和收款账号。不能注销，到期自动失效。
*/
type TxTokenPublish struct {
	TxBase
	Token_name       string   `json:"token_name"`       //名称
	Token_symbol     string   `json:"token_symbol"`     //单位
	Token_supply     *big.Int `json:"token_supply"`     //发行总量
	Token_accuracy   uint64   `json:"token_accuracy"`   //小数点之后位数
	Token_Vout_total uint64   `json:"token_vout_total"` //输出交易数量
	Token_Vout       []Vout   `json:"token_vout"`       //交易输出
}

/*
 */
type TxTokenPublish_VO struct {
	TxBaseVO
	Token_name       string    `json:"token_name"`       //名称
	Token_symbol     string    `json:"token_symbol"`     //单位
	Token_supply     string    `json:"token_supply"`     //发行总量
	Token_accuracy   uint64    `json:"token_accuracy"`   //小数点之后位数
	Token_Vout_total uint64    `json:"token_vout_total"` //输出交易数量
	Token_Vout       []*VoutVO `json:"token_vout"`       //交易输出
}

/*
用于地址和txid格式化显示
*/
func (this *TxTokenPublish) GetVOJSON() interface{} {
	vouts := make([]*VoutVO, 0)
	for _, one := range this.Token_Vout {
		vouts = append(vouts, one.ConversionVO())
	}
	return TxTokenPublish_VO{
		TxBaseVO:         this.TxBase.ConversionVO(),
		Token_name:       this.Token_name,            //名称
		Token_symbol:     this.Token_symbol,          //单位
		Token_supply:     this.Token_supply.Text(10), //发行总量
		Token_accuracy:   this.Token_accuracy,        //小数点之后位数
		Token_Vout_total: this.Token_Vout_total,      //输出交易数量
		Token_Vout:       vouts,                      //交易输出
	}
}

/*
构建hash值得到交易id
*/
func (this *TxTokenPublish) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_token_publish)

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
	格式化成json字符串
*/
// func (this *TxToken) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
格式化成[]byte
*/
func (this *TxTokenPublish) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vins = append(vins, &go_protos.Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
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
	tokenVouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Token_Vout {
		tokenVouts = append(tokenVouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	txPay := go_protos.TxTokenPublish{
		TxBase:          &txBase,
		TokenName:       this.Token_name,
		TokenSymbol:     this.Token_symbol,
		TokenSupply:     this.Token_supply.Bytes(),
		TokenAccuracy:   this.Token_accuracy,
		Token_VoutTotal: this.Token_Vout_total,
		Token_Vout:      tokenVouts,
	}
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
格式化成json字符串
*/
func (this *TxTokenPublish) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)

	buf.Write([]byte(this.Token_name))
	buf.Write([]byte(this.Token_symbol))
	buf.Write(this.Token_supply.Bytes())
	buf.Write(utils.Uint64ToBytes(this.Token_accuracy))
	buf.Write(utils.Uint64ToBytes(this.Token_Vout_total))
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			bs := one.Serialize()
			buf.Write(*bs)
		}
	}
	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *TxTokenPublish) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)

	*signDst = append(*signDst, this.Token_name...)
	*signDst = append(*signDst, this.Token_symbol...)
	*signDst = append(*signDst, this.Token_supply.Bytes()...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_accuracy)...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vout_total)...)
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			*signDst = append(*signDst, *one.Serialize()...)
		}
	}
	// fmt.Println("签名前的字节", len(*bs), hex.EncodeToString(*bs), "\n")
	sign := keystore.Sign(*key, *signDst)
	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))
	return &sign
}

// func (this *TxTokenPublish) GetTokenVoutSignSerialize(voutIndex uint64) *[]byte {
// 	bufVout := bytes.NewBuffer(nil)
// 	//上一个交易的指定输出序列化
// 	bufVout.Write(utils.Uint64ToBytes(voutIndex))
// 	vout := this.Token_Vout[voutIndex]
// 	bs := vout.Serialize()
// 	bufVout.Write(*vout.Serialize())
// 	*bs = bufVout.Bytes()
// 	return bs
// }

/*
检查交易是否合法
*/
func (this *TxTokenPublish) CheckSign() error {
	// fmt.Println("开始验证交易合法性")
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

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费
	for i, one := range this.Vin {

		sign := this.GetSignSerialize(nil, uint64(i))

		*sign = append(*sign, this.Token_name...)
		*sign = append(*sign, this.Token_symbol...)
		*sign = append(*sign, this.Token_supply.Bytes()...)
		*sign = append(*sign, utils.Uint64ToBytes(this.Token_accuracy)...)
		*sign = append(*sign, utils.Uint64ToBytes(this.Token_Vout_total)...)
		if this.Token_Vout != nil {
			for _, one := range this.Token_Vout {
				*sign = append(*sign, *one.Serialize()...)
			}
		}

		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		if !ed25519.Verify(puk, *sign, one.Sign) {
			return config.ERROR_sign_fail
		}
	}
	return nil
}

/*
验证是否合法
*/
func (this *TxTokenPublish) GetWitness() *crypto.AddressCoin {
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
func (this *TxTokenPublish) GetSpend() uint64 {
	return this.Gas
}

/*
检查重复的交易
*/
func (this *TxTokenPublish) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }

	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_token_publish {
			continue
		}
		// ta := one.(*TxToken)
		// if bytes.Equal(ta.Account, this.Account) {
		// 	return false
		// }
	}
	return true
}

/*
统计交易余额
*/
func (this *TxTokenPublish) CountTxItemsNew(height uint64) *TxItemCountMap {
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

/*
发布token
@nameStr      string                Token名称全称
@symbolStr    string                Token单位，符号
@supply       *big.Int              发行总量
@owner        crypto.AddressCoin    Token所有者
*/
func CreateTxTokenPublish(srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string,
	nameStr, symbolStr string, supply *big.Int, owner crypto.AddressCoin) (*TxTokenPublish, error) {

	var commentbs []byte
	if comment != "" {
		commentbs = []byte(comment)
	}

	//验证发行总量最少
	if supply.Cmp(config.Witness_token_supply_min) < 0 {
		return nil, config.ERROR_token_min_fail //
	}
	// if supply < config.Witness_token_supply_min {
	// 	return nil, config.ERROR_token_min_fail //
	// }

	//所有人为空，则设置所有人为本钱包coinbase
	if owner == nil {
		owner = config.Area.Keystore.GetCoinbase().Addr
	}

	tokenVout := make([]Vout, 0)
	voutOne := Vout{
		Value:   supply.Uint64(), //输出金额 = 实际金额 * 100000000
		Address: owner,           //钱包地址
	}
	tokenVout = append(tokenVout, voutOne)

	//
	chain := GetLongChain()

	//查找余额
	vins := make([]*Vin, 0)

	total, item := chain.GetBalance().BuildPayVinNew(srcAddr, gas)
	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }
	// for _, item := range items {
	puk, ok := config.Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)
	// }

	//余额不够给手续费
	// if total < (amount + gas) {
	// 	//余额不够
	// 	// _, e := model.Errcode(model.NotEnough)
	// 	return nil, config.ERROR_not_enough
	// }

	//押金冻结存放地址
	// var dstAddr crypto.AddressCoin
	// if addr == nil {
	// 	dstAddr = keystore.GetCoinbase().Addr
	// } else {
	// 	dstAddr = *addr
	// }
	//构建交易输出
	vouts := make([]*Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	// vout := mining.Vout{
	// 	Value:   amount,  //输出金额 = 实际金额 * 100000000
	// 	Address: dstAddr, //钱包地址
	// }
	// vouts = append(vouts, &vout)
	//找回零钱
	// if total > amount+gas {
	// 	vout := mining.Vout{
	// 		Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
	// 		Address: *items[0].Addr,       // keystore.GetAddr()[0].Addr, //钱包地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()
	var txin *TxTokenPublish
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_token_publish,             //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
			Comment:    []byte{},
		}
		txin = &TxTokenPublish{
			TxBase:           base,
			Token_name:       nameStr,                //名称
			Token_symbol:     symbolStr,              //单位
			Token_supply:     supply,                 //发行总量
			Token_Vout_total: uint64(len(tokenVout)), //输出交易数量
			Token_Vout:       tokenVout,              //交易输出
		}

		//这个交易存在手续费为0的情况，所以不合并vout
		// txin.MergeVout()
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := config.Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			sign := txin.GetSign(&prk, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
			txin.Vin[i].Sign = *sign
		}

		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	// chain.GetBalance().Frozen(items, txin)
	chain.GetBalance().AddLockTx(txin)
	return txin, nil
}

//======================= TokenPublishController =======================

type TokenPublishController struct {
}

func (this *TokenPublishController) Factory() interface{} {
	return new(TxTokenPublish)
}

func (this *TokenPublishController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxTokenPublish)
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

	tokenVouts := make([]Vout, 0)
	for _, one := range txProto.Token_Vout {
		tokenVouts = append(tokenVouts, Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	tx := &TxTokenPublish{
		TxBase:           txBase,
		Token_name:       txProto.TokenName,
		Token_symbol:     txProto.TokenSymbol,
		Token_supply:     new(big.Int).SetBytes(txProto.TokenSupply),
		Token_accuracy:   txProto.TokenAccuracy,
		Token_Vout_total: txProto.Token_VoutTotal,
		Token_Vout:       tokenVouts,
	}
	return tx, nil
}

/*
统计余额
将已经注册的域名保存到数据库
将自己注册的域名保存到内存
*/
func (this *TokenPublishController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {

	txItemCounts := make(map[string]*map[string]*big.Int, 0)

	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_token_publish {
			continue
		}
		txToken := txItr.(*TxTokenPublish)
		//添加一个token信息
		SaveTokenInfo(txToken.Token_Vout[0].Address, *txToken.GetHash(), txToken.Token_name, txToken.Token_symbol, txToken.Token_accuracy, txToken.Token_supply)

		txid := utils.Bytes2string(*txToken.GetHash())

		tokenMap, ok := txItemCounts[txid]
		if !ok {
			newm := make(map[string]*big.Int, 0)
			tokenMap = &newm
			txItemCounts[txid] = &newm
		}

		// totalValue := uint64(0)
		for _, vout := range txToken.Token_Vout {
			// totalValue += vout.Value
			value, ok := (*tokenMap)[utils.Bytes2string(vout.Address)]
			if !ok {
				value = new(big.Int)
			}
			value.Add(value, new(big.Int).SetUint64(vout.Value))
			//value += int64(vout.Value)
			(*tokenMap)[utils.Bytes2string(vout.Address)] = value
		}
	}

	//统计主链上的余额
	// balance.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)

	//统计token的余额
	CountToken(&txItemCounts)
}

func (this *TokenPublishController) CheckMultiplePayments(txItr TxItr) error {
	return nil
}

func (this *TokenPublishController) SyncCount() {

}

func (this *TokenPublishController) RollbackBalance() {
	// return new(Tx_account)
}

/*
注册域名交易，域名续费交易，修改域名的网络地址交易
@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *TokenPublishController) BuildTx(deposit *sync.Map, srcAddr,
	addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string,
	params ...interface{}) (TxItr, error) {

	if len(params) < 5 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}

	// @name    string    Token名称全称
	nameStr := params[0].(string)
	// @symbol    string    Token单位，符号
	symbolStr := params[1].(string)
	// @supply    uint64    发行总量
	supply := params[2].(*big.Int)
	// @accuracy    uint64    精度
	accuracy := params[3].(uint64)
	// @owner    crypto.AddressCoin    所有者
	owner := params[4].(crypto.AddressCoin)

	var commentbs []byte
	if comment != "" {
		commentbs = []byte(comment)
	}

	//验证发行总量最少
	if supply.Cmp(config.Witness_token_supply_min) < 0 {
		return nil, config.ERROR_token_min_fail //
	}
	// if supply < config.Witness_token_supply_min {
	// 	return nil, config.ERROR_token_min_fail //
	// }

	//所有人为空，则设置所有人为本钱包coinbase
	if owner == nil {
		owner = config.Area.Keystore.GetCoinbase().Addr
	}

	tokenVout := make([]Vout, 0)
	voutOne := Vout{
		Value:   supply.Uint64(), //输出金额 = 实际金额 * 100000000
		Address: owner,           //钱包地址
	}
	tokenVout = append(tokenVout, voutOne)

	//
	chain := GetLongChain()

	//查找余额
	vins := make([]*Vin, 0)

	total, item := chain.GetBalance().BuildPayVinNew(srcAddr, gas)
	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }
	// for _, item := range items {
	puk, ok := config.Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk:   puk, //公钥
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)),
	}
	vins = append(vins, &vin)
	// }

	//余额不够给手续费
	// if total < (amount + gas) {
	// 	//余额不够
	// 	// _, e := model.Errcode(model.NotEnough)
	// 	return nil, config.ERROR_not_enough
	// }

	//押金冻结存放地址
	// var dstAddr crypto.AddressCoin
	// if addr == nil {
	// 	dstAddr = keystore.GetCoinbase().Addr
	// } else {
	// 	dstAddr = *addr
	// }
	//构建交易输出
	vouts := make([]*Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	// vout := Vout{
	// 	Value:   amount,  //输出金额 = 实际金额 * 100000000
	// 	Address: dstAddr, //钱包地址
	// }
	// vouts = append(vouts, &vout)
	//找回零钱
	// if total > amount+gas {
	// 	vout := Vout{
	// 		Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
	// 		Address: *items[0].Addr,       // keystore.GetAddr()[0].Addr, //钱包地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()
	var txin *TxTokenPublish
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_token_publish,             //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
			Comment:    []byte{},
		}
		txin = &TxTokenPublish{
			TxBase:           base,
			Token_name:       nameStr,                //名称
			Token_symbol:     symbolStr,              //单位
			Token_supply:     supply,                 //发行总量
			Token_accuracy:   accuracy,               //精度
			Token_Vout_total: uint64(len(tokenVout)), //输出交易数量
			Token_Vout:       tokenVout,              //交易输出
		}

		//这个交易存在手续费为0的情况，所以不合并vout
		// txin.MergeVout()
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := config.Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			sign := txin.GetSign(&prk, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
			txin.Vin[i].Sign = *sign
		}

		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	// chain.GetBalance().Frozen(items, txin)
	chain.GetBalance().AddLockTx(txin)
	return txin, nil
}

// func (this *TokenPaymentController) Check(txItr TxItr, lockHeight uint64) error {
// 	txAcc := txItr.(*TxToken)
// 	return txAcc.Check(lockHeight)
// }

// /*
// 	检查域名是否过期
// 	@return    bool    域名是否存在
// 	@return    bool    域名是否过期
// */
// func CheckName(nameStr string) (bool, bool, error) {
// 	//判断域名是否已经注册
// 	txid, err := db.Find(append([]byte(config.Name), []byte(nameStr)...))
// 	if err != nil {
// 		if err == leveldb.ErrNotFound {
// 			return false, true, errors.New("域名账号不存在")
// 		}
// 		return false, true, err
// 	}

// 	bs, err := db.Find(*txid)
// 	if err != nil {
// 		return false, true, err
// 	}

// 	//域名已经存在，检查之前的域名是否过期，检查是否是续签
// 	existTx, err := ParseTxBase(bs)
// 	if err != nil {
// 		return false, true, errors.New("checkname 解析域名注册交易出错")
// 	}
// 	//检查区块高度，查看是否过期
// 	blockBs, err := db.Find(*existTx.GetBlockHash())
// 	if err != nil {
// 		//TODO 可能是数据库损坏或数据被篡改出错
// 		return false, true, errors.New("查找域名注册交易对应的区块出错")
// 	}
// 	bh, err := ParseBlockHead(blockBs)
// 	if err != nil {
// 		return false, true, errors.New("解析域名注册交易对应的区块出错")
// 	}
// 	//检查是否过期
// 	if GetHighestBlock() > (bh.Height + name.NameOfValidity) {
// 		//域名已经存在
// 		return true, true, nil
// 	} else {
// 		return true, false, nil
// 	}

// }

/*
		发布一种Token
		@addr    *crypto.AddressCoin    收款地址
		@amount    uint64    转账金额
		@gas    uint64    手续费
		@pwd    string    支付密码
		@name    string    Token名称全称
		@symbol    string    Token单位，符号
		@supply    uint64    发行总量
	    @owner    crypto.AddressCoin    所有者
*/
func PublishToken(srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string,
	name, symbol string, supply *big.Int, accuracy uint64, owner crypto.AddressCoin) (TxItr, error) {

	//缴纳押金注册一个名称
	txItr, err := GetLongChain().GetBalance().BuildOtherTx(config.Wallet_tx_type_token_publish,
		srcAddr, addr, amount, gas, frozenHeight, pwd, comment, name, symbol, supply, accuracy, owner)
	if err != nil {
		// fmt.Println("缴纳域名押金失败", err)
	} else {
		// fmt.Println("缴纳域名押金完成")
	}
	return txItr, err
}
