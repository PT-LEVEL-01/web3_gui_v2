package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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
	tpc := new(TokenPaymentController)
	//tpc.ActiveVoutIndex = new(sync.Map)
	RegisterTransaction(config.Wallet_tx_type_token_payment, tpc)
}

/*
token转账交易
*/
type TxTokenPay struct {
	TxBase
	Token_txid       []byte  `json:"token_txid"`       //合约地址
	Token_Vin_total  uint64  `json:"token_vin_total"`  //输入交易数量
	Token_Vin        []*Vin  `json:"token_vin"`        //交易输入
	Token_Vout_total uint64  `json:"token_vout_total"` //输出交易数量
	Token_Vout       []*Vout `json:"token_vout"`       //交易输出
	// Token_publish_txid     []byte         `json:"token_publish_txid"` //token的合约地址
	// Token_publish_txid_str string `json:"-"` //
}

/*
 */
type TxTokenPay_VO struct {
	TxBaseVO
	Token_name         string    `json:"token_name"`         //名称
	Token_symbol       string    `json:"token_symbol"`       //单位
	Token_supply       uint64    `json:"token_supply"`       //发行总量
	Token_txid         string    `json:"token_txid"`         //合约地址
	Token_Vin_total    uint64    `json:"token_vin_total"`    //输入交易数量
	Token_Vin          []*VinVO  `json:"token_vin"`          //交易输入
	Token_Vout_total   uint64    `json:"token_vout_total"`   //输出交易数量
	Token_Vout         []*VoutVO `json:"token_vout"`         //交易输出
	Token_publish_txid string    `json:"token_publish_txid"` //token的合约地址

}

/*
用于地址和txid格式化显示
*/
func (this *TxTokenPay) GetVOJSON() interface{} {

	vins := make([]*VinVO, 0)
	for _, one := range this.Token_Vin {
		vins = append(vins, one.ConversionVO())
	}
	vouts := make([]*VoutVO, 0)
	for _, one := range this.Token_Vout {
		vouts = append(vouts, one.ConversionVO())
	}

	return TxTokenPay_VO{
		TxBaseVO:         this.TxBase.ConversionVO(),
		Token_txid:       hex.EncodeToString(this.Token_txid), //
		Token_Vin_total:  this.Token_Vin_total,                //输入交易数量
		Token_Vin:        vins,                                //交易输入
		Token_Vout_total: this.Token_Vout_total,               //输出交易数量
		Token_Vout:       vouts,                               //交易输出
		// Token_publish_txid: hex.EncodeToString(this.Token_publish_txid), //token的合约地址
	}
}

/*
构建hash值得到交易id
*/
func (this *TxTokenPay) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_token_payment)
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
func (this *TxTokenPay) Proto() (*[]byte, error) {
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
	tokenVins := make([]*go_protos.Vin, 0)
	for _, one := range this.Token_Vin {
		tokenVins = append(tokenVins, &go_protos.Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
			Nonce: one.Nonce.Bytes(),
			Puk:   one.Puk,
			Sign:  one.Sign,
		})
	}
	tokenVouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Token_Vout {
		tokenVouts = append(tokenVouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}

	txPay := go_protos.TxTokenPay{
		TxBase:          &txBase,
		Token_Txid:      this.Token_txid,
		Token_VinTotal:  this.Token_Vin_total,
		Token_Vin:       tokenVins,
		Token_VoutTotal: this.Token_Vout_total,
		Token_Vout:      tokenVouts,
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
func (this *TxTokenPay) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)

	buf.Write([]byte(this.Token_txid))
	// buf.Write([]byte(this.Token_symbol))
	// buf.Write(utils.Uint64ToBytes(this.Token_supply))

	buf.Write(utils.Uint64ToBytes(this.Token_Vin_total))
	if this.Token_Vin != nil {
		for _, one := range this.Token_Vin {
			bs := one.SerializeVin()
			buf.Write(*bs)
		}
	}
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
func (this *TxTokenPay) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {

	// txItr, err := mining.LoadTxBase(txid)
	// // txItr, err := mining.FindTxBase(txid)
	// if err != nil {
	// 	return nil
	// }

	// blockhash, err := db.GetTxToBlockHash(&txid)
	// if err != nil || blockhash == nil {
	// 	return nil
	// }

	// if txItr.GetBlockHash() == nil {
	// 	txItr = mining.GetRemoteTxAndSave(txid)
	// 	if txItr.GetBlockHash() == nil {
	// 		return nil
	// 	}
	// }
	// txItr.SetBlockHash([]byte{})
	// fmt.Println("000000000000000000", len(*txItr.GetBlockHash()))

	// buf := bytes.NewBuffer(nil)
	//上一个交易 所属的区块hash
	// buf.Write(*blockhash)
	// buf.Write(*txItr.GetBlockHash())
	//上一个交易的hash
	// buf.Write(*txItr.GetHash())
	//上一个交易的指定输出序列化
	// buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	//本交易类型输入输出数量等信息和所有输出
	// signBs := buf.Bytes()
	//signDst := this.GetSignSerialize(nil, vinIndex)
	//*signDst = append(*signDst, this.Token_txid...)

	//// engine.Log.Info("222222 %s", hex.EncodeToString(*signDst))

	//*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vin_total)...)
	//if this.Token_Vin != nil {
	//	for _, one := range this.Token_Vin {
	//		*signDst = append(*signDst, *one.SerializeVin()...)
	//	}
	//}
	//*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vout_total)...)
	//if this.Token_Vout != nil {
	//	for _, one := range this.Token_Vout {
	//		*signDst = append(*signDst, *one.Serialize()...)
	//	}
	//}

	// engine.Log.Info("签名之前序列化字符串 %d %s", len(*signDst), hex.EncodeToString(*signDst))
	signDst := this.GetWaitSignSerialize(vinIndex)
	sign := keystore.Sign(*key, *signDst)

	return &sign
}

/*
获取Token签名
*/
func (this *TxTokenPay) GetWaitSignSerialize(vinIndex uint64) *[]byte {
	signDst := this.TxBase.GetSignSerialize(nil, vinIndex)

	*signDst = append(*signDst, this.Token_txid...)
	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vin_total)...)
	if this.Token_Vin != nil {
		for _, one := range this.Token_Vin {
			*signDst = append(*signDst, one.Puk...)
		}
	}
	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vout_total)...)
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			*signDst = append(*signDst, *one.Serialize()...)
		}
	}

	return signDst
}

/*
把相同地址的交易输出合并在一起
*/
func (this *TxTokenPay) MergeTokenVout() {
	this.Token_Vout = MergeVouts(&this.Token_Vout)
	this.Token_Vout_total = uint64(len(this.Token_Vout))
}

/*
检查交易是否合法
*/
func (this *TxTokenPay) CheckSign() error {
	if len(this.Vin) != 1 || len(this.Token_Vin) != 1 {
		return config.ERROR_pay_vin_too_much
	}
	if len(this.Vin[0].Nonce.Bytes()) == 0 {
		// engine.Log.Info("txid:%s nonce is nil", txItr.GetHash())
		return config.ERROR_pay_nonce_is_nil
	}
	if this.Vout_total != 0 {
		return config.ERROR_pay_vout_too_much
	}

	for i, _ := range this.Token_Vin {
		this.Token_Vin[i].Sign = nil
	}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费;3.输入不能重复。
	// vinMap := make(map[string]int)
	// inTotal := uint64(0)
	for i, one := range this.Vin {

		//不能有重复的vin
		// key := string(mining.BuildKeyForUnspentTransaction(one.Txid, one.Vout))
		// key := utils.Bytes2string(this.Vin[i].Puk)
		// if _, ok := vinMap[key]; ok {
		// 	return config.ERROR_tx_Repetitive_vin
		// }
		// vinMap[key] = 0

		// txItr, err := mining.LoadTxBase(one.Txid)
		// // txItr, err := mining.FindTxBase(one.Txid)
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
		// 	engine.Log.Error("ERROR_sign_fail")
		// 	return config.ERROR_sign_fail
		// }

		// //验证签名
		// buf := bytes.NewBuffer(nil)
		// //上一个交易 所属的区块hash
		// buf.Write(*blockhash)
		// // buf.Write(*txItr.GetBlockHash())
		// //上一个交易的hash
		// buf.Write(*txItr.GetHash())
		// //上一个交易的指定输出序列化
		// buf.Write(*txItr.GetVoutSignSerialize(one.Vout))
		// //本交易类型输入输出数量等信息和所有输出
		// bs := buf.Bytes()
		//sign := this.GetSignSerialize(nil, uint64(i))

		//*sign = append(*sign, this.Token_txid...)
		//*sign = append(*sign, utils.Uint64ToBytes(this.Token_Vin_total)...)
		//if this.Token_Vin != nil {
		//	for _, vin := range this.Token_Vin {
		//		*sign = append(*sign, *vin.SerializeVin()...)
		//	}
		//}
		//*sign = append(*sign, utils.Uint64ToBytes(this.Token_Vout_total)...)
		//if this.Token_Vout != nil {
		//	for _, vout := range this.Token_Vout {
		//		*sign = append(*sign, *vout.Serialize()...)
		//	}
		//}

		sign := this.GetWaitSignSerialize(uint64(i))
		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}

		if !ed25519.Verify(puk, *sign, one.Sign) {
			return config.ERROR_sign_fail
		}
	}

	// engine.Log.Info("验证token转账耗时 111111111111111 %s", config.TimeNow().Sub(start))

	//判断输入输出是否相等
	// outTotal := uint64(0)
	// for _, one := range this.Vout {
	// 	outTotal = outTotal + one.Value
	// }
	// // engine.Log.Info("这里的手续费是否正确 %d %d %d", outTotal, inTotal, this.Gas)
	// if outTotal > inTotal {
	// 	return config.ERROR_tx_fail
	// }
	// this.Gas = inTotal - outTotal

	//token交易vin和vout额度相等
	//同时判断vin中所有token种类相同
	// var publishTxid *[]byte
	// tokenVoutTotal := uint64(0)
	// for _, one := range this.Token_Vout {
	// 	tokenVoutTotal = tokenVoutTotal + one.Value
	// }

	//输入不能重复。
	// vinMap = make(map[string]int)
	// tokenVinTotal := uint64(0)
	for _, _ = range this.Token_Vin {

		// keyStr := config.TokenPublishTxid + utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))

		//不能有重复的vin
		// if _, ok := vinMap[keyStr]; ok {
		// 	return config.ERROR_tx_Repetitive_vin
		// }
		// vinMap[keyStr] = 0

		// publishTxidBs, err := db.LevelTempDB.Find([]byte(keyStr))
		// if err != nil {
		// 	return err
		// }
		//先判断种类是否相同
		// if vinIndex == 0 {
		// 	publishTxid = publishTxidBs
		// } else {
		// 	if !bytes.Equal(*publishTxid, *publishTxidBs) {
		// 		return config.ERROR_tx_fail
		// 	}
		// }

		// engine.Log.Info("验证token转账耗时 222222222222222 %s", config.TimeNow().Sub(start))

		//查询余额
		// txItr, err := mining.LoadTxBase(vin.Txid)
		// // txItr, err := mining.FindTxBase(vin.Txid)
		// if err != nil {
		// 	return config.ERROR_tx_format_fail
		// }

		// bs, err := json.Marshal(txItr)
		// if err != nil {
		// 	return err
		// }
		// txToken := new(TxToken)
		// decoder := json.NewDecoder(bytes.NewBuffer(bs))
		// decoder.UseNumber()
		// err = decoder.Decode(txToken)
		// if err != nil {
		// 	return err
		// }

		// if mining.ParseTxClass(vin.Txid) == config.Wallet_tx_type_token_publish {
		// 	txToken := txItr.(*publish.TxToken)
		// 	vout := (txToken.Token_Vout)[vin.Vout]
		// 	tokenVinTotal = tokenVinTotal + vout.Value
		// } else if mining.ParseTxClass(vin.Txid) == config.Wallet_tx_type_token_payment {
		// 	txToken := txItr.(*TxToken)
		// 	vout := (txToken.Token_Vout)[vin.Vout]
		// 	tokenVinTotal = tokenVinTotal + vout.Value
		// } else {
		// 	return config.ERROR_tx_fail
		// }

	}
	// if tokenVoutTotal != tokenVinTotal {
	// 	return config.ERROR_tx_fail
	// }

	// engine.Log.Info("验证token转账耗时 3333333333333333 %s", config.TimeNow().Sub(start))

	return nil
}

/*
验证是否合法
*/
func (this *TxTokenPay) GetWitness() *crypto.AddressCoin {
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
func (this *TxTokenPay) GetSpend() uint64 {
	return this.Gas
}

/*
检查重复的交易
*/
func (this *TxTokenPay) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }

	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_token_payment {
			continue
		}
		// ta := one.(*TxTokenPay)
		// if bytes.Equal(ta.Account, this.Account) {
		// 	return false
		// }
	}
	return true
}

/*
	统计交易余额
*/
// func (this *TxTokenPay) CountTxItems(height uint64) *mining.TxItemCount {
// 	itemCount := mining.TxItemCount{
// 		Additems: make([]*mining.TxItem, 0),
// 		SubItems: make([]*mining.TxSubItems, 0),
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
// 		itemCount.SubItems = append(itemCount.SubItems, &mining.TxSubItems{
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
// 		txItem := mining.TxItem{
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
// 		mining.TxCache.AddTxInTxItem(*this.GetHash(), this)

// 	}
// 	return &itemCount
// }

/*
统计交易余额
*/
func (this *TxTokenPay) CountTxItemsNew(height uint64) *TxItemCountMap {
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

//======================= TokenPaymentController =======================

type TokenPaymentController struct {
	//ActiveVoutIndex *sync.Map //活动的交易输出，key:string=[txid]_[vout index];value:=;
}

func (this *TokenPaymentController) Factory() interface{} {
	return new(TxTokenPay)
}

func (this *TokenPaymentController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxTokenPay)
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

	tokenVins := make([]*Vin, 0)
	for _, one := range txProto.Token_Vin {
		nonce := new(big.Int).SetBytes(one.Nonce)
		tokenVins = append(tokenVins, &Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
			Puk:   one.Puk,
			Sign:  one.Sign,
			Nonce: *nonce,
		})
	}

	tokenVouts := make([]*Vout, 0)
	for _, one := range txProto.Token_Vout {
		tokenVouts = append(tokenVouts, &Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	tx := &TxTokenPay{
		TxBase:           txBase,
		Token_txid:       txProto.Token_Txid,
		Token_Vin_total:  txProto.Token_VinTotal,
		Token_Vin:        tokenVins,
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
func (this *TokenPaymentController) CountBalance(deposit *sync.Map, bhvo *BlockHeadVO) {

	// engine.Log.Info("开始统计token余额 1111111")

	txItemCounts := make(map[string]*map[string]*big.Int, 0)

	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_token_payment {
			continue
		}

		txToken := txItr.(*TxTokenPay)
		//余额中减去。
		from := txToken.Token_Vin[0].GetPukToAddr()
		txid := utils.Bytes2string(txToken.Token_txid)

		tokenMap, ok := txItemCounts[txid]
		if !ok {
			newm := make(map[string]*big.Int, 0)
			tokenMap = &newm
			txItemCounts[txid] = &newm
		}

		totalValue := uint64(0)
		for _, vout := range txToken.Token_Vout {
			value, ok := (*tokenMap)[utils.Bytes2string(vout.Address)]
			if !ok {
				value = new(big.Int)
			}
			//value += int64(vout.Value)
			//value.Add(vout.Value)
			value.Add(value, new(big.Int).SetUint64(vout.Value))
			(*tokenMap)[utils.Bytes2string(vout.Address)] = value
			totalValue += vout.Value
		}

		value, ok := (*tokenMap)[utils.Bytes2string(*from)]
		if !ok {
			value = new(big.Int)
		}
		//value -= int64(totalValue)
		value.Sub(value, new(big.Int).SetUint64(totalValue))
		(*tokenMap)[utils.Bytes2string(*from)] = value
	}
	//统计主链上的余额
	// balance.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)

	//统计token的余额
	CountToken(&txItemCounts)
}

func (this *TokenPaymentController) CheckMultiplePayments(txItr TxItr) error {
	txToken := txItr.(*TxTokenPay)

	vinAddr := crypto.AddressCoin{}
	for _, vin := range txToken.Token_Vin {
		vinAddr = *vin.GetPukToAddr()
	}

	totalToken := uint64(0)
	for _, vout := range txToken.Token_Vout {
		totalToken += vout.Value
	}

	//在未打包交易中获取该地址的token花费
	notSpendToken, _, _ := GetTokenNotSpendAndLockedBalance(txToken.Token_txid, vinAddr)
	if notSpendToken < totalToken {
		return config.ERROR_token_not_enough
	}

	return nil
}

func (this *TokenPaymentController) SyncCount() {

}

func (this *TokenPaymentController) RollbackBalance() {
	// return new(Tx_account)
}

/*
注册域名交易，域名续费交易，修改域名的网络地址交易
@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *TokenPaymentController) BuildTx(deposit *sync.Map, srcAddr, addr *crypto.AddressCoin,
	amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {

	if len(params) < 1 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}

	//---------------------开始构建token的交易----------------------
	//发布token的交易id
	txid := params[0].(string)
	txidBs, err := hex.DecodeString(txid)
	if err != nil {
		return nil, config.ERROR_params_fail
	}

	// txidBs, err := hex.DecodeString(txid)
	// if err != nil {
	// 	return nil, config.ERROR_params_fail
	// }

	var commentbs []byte
	if comment != "" {
		commentbs = []byte(comment)
	}

	// srcAddrStr := ""
	// if srcAddr != nil {
	// 	srcAddrStr = srcAddr.B58String()
	// }

	tokenTotal, _, _ := GetTokenNotSpendAndLockedBalance(txidBs, *srcAddr)
	if tokenTotal < amount {
		return nil, config.ERROR_token_not_enough
	}
	tokenVins := make([]*Vin, 0)
	// for _, item := range tokenTxItems {
	puk, ok := config.Area.Keystore.GetPukByAddr(*srcAddr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	tokenvin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk: puk, //公钥
	}
	tokenVins = append(tokenVins, &tokenvin)
	// }

	//构建交易输出
	tokenVouts := make([]*Vout, 0)
	//转账token给目标地址
	tokenVout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *addr,        //钱包地址
		FrozenHeight: frozenHeight, //
	}
	tokenVouts = append(tokenVouts, &tokenVout)
	//找零
	// if tokenTotal > amount {
	// 	tokenVout := Vout{
	// 		Value:   tokenTotal - amount,   //输出金额 = 实际金额 * 100000000
	// 		Address: *tokenTxItems[0].Addr, // keystore.GetAddr()[0].Addr, //钱包地址
	// 	}
	// 	tokenVouts = append(tokenVouts, &tokenVout)
	// }

	//---------------------开始构建主链上的交易----------------------
	//查找余额
	vins := make([]*Vin, 0)
	chain := GetLongChain()
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
	vouts := make([]*Vout, 0)

	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	// if total > gas {
	// 	vout := Vout{
	// 		Value:   total - gas,    //输出金额 = 实际金额 * 100000000
	// 		Address: *items[0].Addr, // keystore.GetAddr()[0].Addr, //钱包地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	_, block := chain.GetLastBlock()
	//var txin *TxToken
	var txin *TxTokenPay
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_token_payment,            //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                      //
			Comment:    []byte{},
		}
		//txin = &TxToken{
		txin = &TxTokenPay{
			TxBase:           base,
			Token_txid:       txidBs,
			Token_Vin_total:  uint64(len(tokenVins)),  //输入交易数量
			Token_Vin:        tokenVins,               //交易输入
			Token_Vout_total: uint64(len(tokenVouts)), //输出交易数量
			Token_Vout:       tokenVouts,              //交易输出
		}

		// txin.MergeVout()
		// txin.MergeTokenVout()

		//给token交易签名 给输出签名，防篡改
		// for i, one := range txin.Token_Vin {
		// 	_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	sign := txin.GetTokenSign(&prk, one.Txid, one.Vout, uint64(i))
		// 	txin.Token_Vin[i].Sign = *sign
		// }

		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := config.Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

			sign := txin.GetSign(&prk, uint64(i))
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

	//把txitem冻结起来
	// FrozenToken(txidBs, tokenTxItems, txin)

	// chain.GetBalance().Frozen(items, txin)
	chain.GetBalance().AddLockTx(txin)
	return txin, nil
}

// 获取代币锁定金额
func GetTokenLockedBalance(tokenId []byte, addr crypto.AddressCoin) *big.Int {
	key := config.BuildDBKeyTokenAddrLockedValue(tokenId, addr)
	val := big.NewInt(0)
	bs, err := db.LevelDB.Get(*key)
	if err != nil {
		return val
	}

	return val.SetBytes(bs)
}

/*
		Token转账
		@addr      *crypto.AddressCoin    收款地址
		@amount    uint64                 转账金额
		@gas       uint64                 手续费
		@pwd       string                 支付密码
	    @txid      string                 发布token的交易id
*/
func TokenPay(srcAddress, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string,
	txid string) (TxItr, error) {

	//缴纳押金注册一个名称
	txItr, err := GetLongChain().GetBalance().BuildOtherTx(config.Wallet_tx_type_token_payment, srcAddress,
		addr, amount, gas, frozenHeight, pwd, comment, txid)
	if err != nil {
		// fmt.Println("缴纳域名押金失败", err)
	} else {
		// fmt.Println("缴纳域名押金完成")
	}
	return txItr, err
}

/*
Token多人转账
*/
func TokenPayMore(srcAddr, tokenSrcAddr crypto.AddressCoin, address []PayNumber, gas uint64, pwd, comment string,
	txid []byte) (TxItr, error) {

	//发布token的交易id
	// txidBs, err := hex.DecodeString(txid)
	// if err != nil {
	// 	return nil, config.ERROR_params_fail
	// }

	//---------------------开始构建token的交易----------------------
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//tokenTxItem, notSpendToken := getTokenNotSpendBalance(&txid, &tokenSrcAddr)
	notSpendToken, _, _ := GetTokenNotSpendAndLockedBalance(txid, tokenSrcAddr)
	if notSpendToken < amount {
		return nil, config.ERROR_token_not_enough
	}
	tokenVins := make([]*Vin, 0)
	// for _, item := range tokenTxItems {
	puk, ok := config.Area.Keystore.GetPukByAddr(tokenSrcAddr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
	tokenVin := Vin{
		// Txid: item.Txid,      //UTXO 前一个交易的id
		// Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk: puk, //公钥
		//					Sign: *sign,           //签名
	}
	tokenVins = append(tokenVins, &tokenVin)
	// }

	//构建交易输出
	tokenVouts := make([]*Vout, 0, len(address))
	for _, one := range address {
		vout := Vout{
			Value:        one.Amount,       //输出金额 = 实际金额 * 100000000
			Address:      one.Address,      //钱包地址
			FrozenHeight: one.FrozenHeight, //
		}
		tokenVouts = append(tokenVouts, &vout)
	}
	// //找零
	// if notSpendToken > amount {
	// 	tokenVout := Vout{
	// 		Value:   notSpendToken - amount,   //输出金额 = 实际金额 * 100000000
	// 		Address: *tokenTxItems[0].Addr, // keystore.GetAddr()[0].Addr, //钱包地址
	// 	}
	// 	tokenVouts = append(tokenVouts, &tokenVout)
	// }

	//---------------------开始构建主链上的交易----------------------
	//查找余额
	vins := make([]*Vin, 0)
	chain := GetLongChain() // forks.GetLongChain()
	total, item := chain.GetBalance().BuildPayVinNew(&srcAddr, gas)

	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	// if len(items) > config.Mining_pay_vin_max {
	// 	return nil, config.ERROR_pay_vin_too_much
	// }

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
	vouts := make([]*Vout, 0)

	//找零
	// if total > gas {
	// 	vout := Vout{
	// 		Value:   total - gas,    //输出金额 = 实际金额 * 100000000
	// 		Address: *items[0].Addr, //钱包地址
	// 	}
	// 	vouts = append(vouts, &vout)
	// }

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	// _, block := chain.GetLastBlock()
	currentHeight := chain.GetCurrentBlock()
	var txin *TxTokenPay
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_token_payment,             //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: uint64(len(vouts)),                              //输出交易数量
			Vout:       vouts,                                           //
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
			// CreateTime: config.TimeNow().Unix(),      //创建时间
			Comment: []byte{},
		}
		txin = &TxTokenPay{
			TxBase:           base,
			Token_txid:       txid,
			Token_Vin_total:  uint64(len(tokenVins)),  //输入交易数量
			Token_Vin:        tokenVins,               //交易输入
			Token_Vout_total: uint64(len(tokenVouts)), //输出交易数量
			Token_Vout:       tokenVouts,              //交易输出
			// Token_publish_txid: txidBs,                  //
		}

		// txin.MergeVout()
		// txin.MergeTokenVout()

		//给token交易签名 给输出签名，防篡改
		// for i, one := range txin.Token_Vin {
		// 	_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	sign := txin.GetTokenSign(&prk, one.Txid, one.Vout, uint64(i))
		// 	txin.Token_Vin[i].Sign = *sign
		// }

		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := config.Area.Keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, config.TimeNow().Sub(startTime))

			sign := txin.GetSign(&prk, uint64(i))
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

	//把txitem冻结起来
	// FrozenToken(txid, tokenTxItems, txin)

	// chain.GetBalance().Frozen(items, txin)
	chain.GetBalance().AddLockTx(txin)
	AddTx(txin)

	return txin, nil
}
