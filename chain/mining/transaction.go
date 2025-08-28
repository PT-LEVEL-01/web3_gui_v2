package mining

import (
	"encoding/hex"
	"errors"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"runtime"
	"strconv"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

const (
// BlockTx_Gas       = "gas"
// BlockTx_Hash      = "hash"
// BlockTx_Vout      = "vout"
// BlockTx_Vout_Tx   = "tx"
// BlockTx_Blockhash = "blockhash"
)

type TxItr interface {
	Class() uint64    //交易类型
	SetClass(uint64)  //设置交易类型
	BuildHash()       //构建交易hash
	GetHash() *[]byte //获得交易hash
	// GetHashStr() string                                                                               //获取交易hash字符串
	CheckLockHeight(lockHeight uint64) error //检查锁定高度是否合法
	// CheckFrozenHeight(frozenHeight uint64, frozenTime int64) error //检查余额冻结高度
	CheckSign() error     //检查交易是否合法
	CheckHashExist() bool //检查hash在数据库中是否已经存在
	// CheckStatus()bool{}//交易上链前，初步去除错误的
	GetSpend() uint64                  //获取交易花费的余额
	CheckRepeatedTx(txs ...TxItr) bool //验证同一区块（或者同一组区块）中是否有相同的交易，比如重复押金。
	// Json() (*[]byte, error)                                                           //将交易格式化成json字符串
	Proto() (*[]byte, error)                                  //将交易格式化成proto字节
	Serialize() *[]byte                                       //将需要签名的字段序列化
	GetVin() *[]*Vin                                          //
	GetVout() *[]*Vout                                        //
	GetGas() uint64                                           //
	GetVoutSignSerialize(voutIndex uint64) *[]byte            //获取交易输出序列化
	GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte //获取签名
	// SetBlockHash(bs []byte)                                                           //设置本交易所属的区块hash
	// GetBlockHash() *[]byte                                                            //
	GetLockHeight() uint64                //获取锁定区块高度
	SetSign(index uint64, bs []byte) bool //修改签名
	SetPayload(bs []byte)                 //修改备注
	GetPayload() []byte                   //获取备注内容
	GetBlockHash() *[]byte                //获取区块hash
	GetVOJSON() interface{}               //用于地址和txid格式化显示
	// CheckSign() bool                         //检查签名是否正确，异步验证
	// CheckRepeated(tx TxItr) bool                                                       //验证是否双重支付，押金是否重复，使用同步验证
	// Balance() *sync.Map                                                               //查询交易输出，统计输出地址余额key:utils.Multihash=收款地址;value:TxItem=地址余额;
	// SetTxid(index uint64, txid *[]byte) error //这个交易输出被使用之后，需要把UTXO输出标记下
	// UnSetTxid(bs *[]byte, index uint64) error                                         //区块回滚，把之前标记为已经使用过的交易的标记去掉
	// CountTxItems(height uint64) *TxItemCount       //统计可用余额
	CountTxItemsNew(height uint64) *TxItemCountMap //统计可用余额
	CountTxHistory(height uint64)                  //统计交易记录
	GetGasUsed() uint64
	SetGasUsed(uint64)
	CheckDomain() bool //验证域名

	GetBloom() []byte   //获取bloom
	SetBloom(bs []byte) //设置bloom

	CheckAddressBind() error  //检查地址绑定
	CheckAddressFrozen() bool //检查地址冻结

	GetComment() []byte   //获取comment
	SetComment(bs []byte) //设置comment
}

/*
交易
*/
type TxBase struct {
	Hash       []byte  `json:"h"`     //本交易hash，不参与区块hash，只用来保存
	Type       uint64  `json:"t"`     //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
	Vin_total  uint64  `json:"vin_t"` //输入交易数量
	Vin        []*Vin  `json:"vin"`   //交易输入
	Vout_total uint64  `json:"vot_t"` //输出交易数量
	Vout       []*Vout `json:"vot"`   //交易输出
	Gas        uint64  `json:"g"`     //交易手续费，此字段不参与交易hash
	LockHeight uint64  `json:"l_h"`   //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
	Payload    []byte  `json:"p"`     //备注信息,计划仅虚拟机使用
	BlockHash  []byte  `json:"bh"`    //本交易属于的区块hash，不参与区块hash，只用来保存
	GasUsed    uint64  //gasused
	Comment    []byte  `json:"c"` //交易备注信息
}

/*
交易
*/
type TxBaseVO struct {
	Hash       string    `json:"hash"`        //本交易hash，不参与区块hash，只用来保存
	Type       uint64    `json:"type"`        //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
	Vin_total  uint64    `json:"vin_total"`   //输入交易数量
	Vin        []*VinVO  `json:"vin"`         //交易输入
	Vout_total uint64    `json:"vout_total"`  //输出交易数量
	Vout       []*VoutVO `json:"vout"`        //交易输出
	Gas        uint64    `json:"gas"`         //交易手续费，此字段不参与交易hash
	LockHeight uint64    `json:"lock_height"` //本交易锁定在低于这个高度的块中，超过这个高度，块将不被打包到区块中。
	Payload    string    `json:"payload"`     //备注信息
	BlockHash  string    `json:"blockhash"`   //本交易属于的区块hash，不参与区块hash，只用来保存
	Timestamp  uint64    `json:"timestamp"`   // 区块时间
	Reward     uint64    `json:"reward"`      // 区块奖励
	Comment    string    `json:"comment"`     //交易备注信息
}

/*
转化为VO对象
*/
func (this *TxBase) ConversionVO() TxBaseVO {
	vins := make([]*VinVO, 0)
	for _, one := range this.Vin {
		vins = append(vins, one.ConversionVO())
	}

	vouts := make([]*VoutVO, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, one.ConversionVO())
	}

	mineReward := uint64(0)
	//if this.Class() == uint64(config.Wallet_tx_type_mining) {
	//bs, err := db.GetTxToBlockHash(&this.Hash)
	//if err == nil {
	//bh, err := LoadBlockHeadByHash(bs)
	//if err == nil {
	//rewardv0, _ := getMineTx(*this.GetHash(), bh.Height)
	//for _, reward := range rewardv0 {
	//	vin := *this.GetVin()
	//	addr := vin[0].GetPukToAddr()
	//	if strings.EqualFold(reward.Into, addr.B58String()) && reward.From == "" {
	//		mineReward = reward.Reward.Uint64()
	//	}
	//}
	//}
	//}
	//}

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
		Reward:     mineReward,
		Comment:    string(this.Comment),
	}
}

// 验证域名
func (this *TxBase) CheckDomain() bool {
	for _, v := range *this.GetVout() {
		if v.Domain == nil || len(v.Domain) == 0 {
			continue
		}

		//是erc20转账
		//transfer  0xa9059cbb
		//transferFrom  0x23b872dd
		if erc20Info := db.GetErc20Info(v.GetAddrStr()); erc20Info.Address != "" {
			if _, toAddr, _ := precompiled.UnpackErc20Payload(this.Payload); toAddr != nil {
				vin := *this.GetVin()
				if !ens.CheckDomainResolve(vin[0].GetPukToAddr().B58String(), string(v.Domain), toAddr.B58String(), new(big.Int).SetUint64(v.DomainType)) {
					return false
				}

				return true
			}
		}

		vin := *this.GetVin()
		if !ens.CheckDomainResolve(vin[0].GetPukToAddr().B58String(), string(v.Domain), v.GetAddrStr(), new(big.Int).SetUint64(v.DomainType)) {
			return false
		}
	}
	return true
}

// 修改签名
func (this *TxBase) SetSign(index uint64, bs []byte) bool {
	this.Vin[index].Sign = bs
	return true
}

/*
设置本交易所属的区块hash
*/
func (this *TxBase) SetBlockHash(bs []byte) {
	this.BlockHash = bs
}

/*
设置本交易所属的区块hash
*/
func (this *TxBase) GetBlockHash() *[]byte {
	return &this.BlockHash
}

/*
获取锁定区块高度
*/
func (this *TxBase) GetLockHeight() uint64 {
	return this.LockHeight
}

/*
将需要hash的字段序列化
*/
func (this *TxBase) Serialize() *[]byte {
	length := 0
	var vinSs []*[]byte
	if this.Vin != nil {
		vinSs = make([]*[]byte, 0, len(this.Vin))
		for _, one := range this.Vin {
			bsOne := one.SerializeVin()
			vinSs = append(vinSs, bsOne)
			length += len(*bsOne)
		}
	}
	var voutSs []*[]byte
	if this.Vout != nil {
		voutSs = make([]*[]byte, 0, len(this.Vout))
		for _, one := range this.Vout {
			bsOne := one.Serialize()
			voutSs = append(voutSs, bsOne)
			length += len(*bsOne)
		}
	}
	length += 8 + 8 + 8 + 8 + len(this.Payload) + len(this.Comment)
	bs := make([]byte, 0, length)

	bs = append(bs, utils.Uint64ToBytes(this.Type)...)
	bs = append(bs, utils.Uint64ToBytes(this.Vin_total)...)
	if vinSs != nil {
		for _, one := range vinSs {
			bs = append(bs, *one...)
		}
	}
	bs = append(bs, utils.Uint64ToBytes(this.Vout_total)...)
	if voutSs != nil {
		for _, one := range voutSs {
			bs = append(bs, *one...)
		}
	}
	bs = append(bs, utils.Uint64ToBytes(this.Gas)...)
	bs = append(bs, utils.Uint64ToBytes(this.LockHeight)...)
	bs = append(bs, this.Payload...)
	bs = append(bs, this.Comment...)
	return &bs

}

func (this *TxBase) GetVin() *[]*Vin {
	return &this.Vin
}
func (this *TxBase) GetVout() *[]*Vout {
	return &this.Vout
}

func (this *TxBase) GetGas() uint64 {
	return this.Gas
}

func (this *TxBase) GetGasUsed() uint64 {
	return this.GasUsed
}

func (this *TxBase) GetHash() *[]byte {
	return &this.Hash
}

func (this *TxBase) Class() uint64 {
	return this.Type
}

func (this *TxBase) SetClass(classType uint64) {
	this.Type = classType
}

func (this *TxBase) SetGasUsed(g uint64) {
	this.GasUsed = g
}

/*
修改备注
*/
func (this *TxBase) SetPayload(bs []byte) {
	this.Payload = bs
}

/*
获取备注内容
*/
func (this *TxBase) GetPayload() []byte {
	return this.Payload
}

/*
获取输出序列化
[UTXO输入引用的块hash]+[UTXO输入引用的块交易hash]+[UTXO输入引用的输出index(uint64)]+
[UTXO输入引用的输出序列化]
*/
func (this *TxBase) GetVoutSignSerialize(voutIndex uint64) *[]byte {
	if voutIndex > uint64(len(this.Vout)) {
		return nil
	}
	vout := this.Vout[voutIndex]
	voutBs := vout.Serialize()
	bs := make([]byte, 0, len(*voutBs)+8)
	bs = append(bs, utils.Uint64ToBytes(voutIndex)...)
	bs = append(bs, *voutBs...)
	return &bs
}

/*
获取本交易用作签名的序列化
[上一个交易GetVoutSignSerialize()返回]+[本交易类型]+[本交易输入总数]+[本交易输入index]+
[本交易输出总数]+[vouts序列化]+[锁定区块高度]
@voutBs    *[]byte    上一个交易GetVoutSignSerialize()返回
*/
func (this *TxBase) GetSignSerialize(voutBs *[]byte, vinIndex uint64) *[]byte {
	if vinIndex > uint64(len(this.Vin)) {
		return nil
	}

	voutBssLenght := 0
	voutBss := make([]*[]byte, 0, len(this.Vout))
	for _, one := range this.Vout {
		voutBsOne := one.Serialize()
		voutBss = append(voutBss, voutBsOne)
		voutBssLenght += len(*voutBsOne)
	}

	var bs []byte
	if voutBs == nil {
		bs = make([]byte, 0, 8+8+8+8+voutBssLenght+8+len(this.Payload)+len(this.Comment))
	} else {
		bs = make([]byte, 0, len(*voutBs)+8+8+8+8+voutBssLenght+8+len(this.Payload)+len(this.Comment))
		bs = append(bs, *voutBs...)
	}

	bs = append(bs, utils.Uint64ToBytes(this.Type)...)
	bs = append(bs, utils.Uint64ToBytes(this.Vin_total)...)
	bs = append(bs, utils.Uint64ToBytes(vinIndex)...)
	bs = append(bs, utils.Uint64ToBytes(this.Vout_total)...)
	for _, one := range voutBss {
		bs = append(bs, *one...)
	}
	bs = append(bs, utils.Uint64ToBytes(this.Gas)...)
	bs = append(bs, utils.Uint64ToBytes(this.LockHeight)...)
	bs = append(bs, this.Payload...)
	bs = append(bs, this.Comment...)
	return &bs

}

/*
获取签名
*/
func (this *TxBase) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)

	// engine.Log.Info("签名字符序列化 耗时 %s", config.TimeNow().Sub(start))
	// fmt.Println("签名前的字节", len(*signDst), hex.EncodeToString(*signDst), "\n")
	// fmt.Printf("签名前的字节 len=%d signDst=%s key=%s \n", len(*signDst), hex.EncodeToString(*signDst), hex.EncodeToString(*key))
	sign := keystore.Sign(*key, *signDst)

	return &sign
}

/*
获取待签名数据
*/
func (this *TxBase) GetWaitSign(vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	return signDst
}

/*
检查锁定高度是否合法
判断交易中锁定高度值是否大于等于参数值
@localHeight    uint64    交易锁定高度
*/
func (this *TxBase) CheckLockHeight(lockHeight uint64) error {
	// engine.Log.Info("对比锁定区块高度 %d %d", this.GetLockHeight(), lockHeight)
	if lockHeight < config.Mining_block_start_height+config.Mining_block_start_height_jump {
		return nil
	}

	if this.GetLockHeight() < lockHeight {
		// engine.Log.Warn("对比锁定区块高度 失败 %d %d %s", this.GetLockHeight(), lockHeight, hex.EncodeToString(*this.GetHash()))
		//engine.Log.Warn("Failed to compare lock block height: LockHeight=%d %d %s", this.GetLockHeight(), lockHeight, hex.EncodeToString(*this.GetHash()))
		return config.ERROR_tx_lockheight
	}
	return nil
}

/*
检查交易是否合法
@localHeight    uint64    交易锁定高度
*/
func (this *TxBase) CheckBase() error {
	if len(this.Vin) > 1 {
		return config.ERROR_pay_vin_too_much
	}
	// fmt.Println("开始验证交易合法性 Tx_deposit_in")
	//判断vin是否太多
	// if len(this.Vin) > config.Mining_pay_vin_max {
	// 	return config.ERROR_pay_vin_too_much
	// }

	//不能出现余额为0的转账
	for i, one := range this.Vout {
		if i != 0 && one.Value <= 0 {
			return config.ERROR_amount_zero
		}
	}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费;3.输入不能重复。
	// vinMap := make(map[string]int)
	// inTotal := uint64(0)
	for i, one := range this.Vin {

		sign := this.GetSignSerialize(nil, uint64(i))

		// engine.Log.Debug("CheckBase 44444444444444 %s", config.TimeNow().Sub(start))

		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		// fmt.Printf("txid:%x puk:%x sign:%x", md5.Sum(one.Txid), md5.Sum(one.Puk), md5.Sum(one.Sign))
		if !ed25519.Verify(puk, *sign, one.Sign) {
			// engine.Log.Debug("ERROR_sign_fail 222 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			// engine.Log.Debug("ed25519.Verify: puk: %x; waitSignData: %x; sign: %x\n", puk, *sign, one.Sign)
			// engine.Log.Debug("ERROR_sign_fail 222 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			return config.ERROR_sign_fail
		}
		// engine.Log.Debug("CheckBase 5555555555555555 %s", config.TimeNow().Sub(start))

	}
	return nil
}

/*
检查这个交易hash是否在数据库中已经存在
*/
func (this *TxBase) CheckHashExist() bool {
	txhashkey := config.BuildBlockTx(this.Hash)
	ok, _ := db.LevelDB.CheckHashExist(txhashkey)
	return ok
}

/*
格式化成[]byte
*/
func (this *TxBase) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vinOne := &go_protos.Vin{
			// Txid: one.Txid,
			// Vout: one.Vout,
			Puk:  one.Puk,
			Sign: one.Sign,
			// Nonce: one.Nonce.Bytes(),
		}
		// if len(one.Nonce.Bytes()) == 0 {
		vinOne.Nonce = one.Nonce.Bytes()
		// }
		vins = append(vins, vinOne)
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
		GasUsed:    this.GasUsed,
		Comment:    this.Comment,
	}

	txPay := go_protos.TxPay{
		TxBase: &txBase,
	}
	// txPay.Marshal()
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *TxBase) CountTxHistory(height uint64) {
	//转出历史记录
	hiOut := HistoryItem{
		IsIn:    false,                          //资金转入转出方向，true=转入;false=转出;
		Type:    this.Class(),                   //交易类型
		InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
		OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
		// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
		// Value:  amount,          //交易金额
		Txid:   *this.GetHash(), //交易id
		Height: height,          //
		// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
		// Payload:
	}
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
	//只有转账才保存备注信息
	if this.Class() == config.Wallet_tx_type_pay {
		hiOut.Payload = this.Payload
		hiIn.Payload = this.Payload
	}

	addrCoin := make(map[string]bool)
	for _, vin := range this.Vin {
		addrInfo, isSelf := Area.Keystore.FindPuk(vin.Puk)
		hiIn.InAddr = append(hiIn.InAddr, &addrInfo.Addr)
		if !isSelf {
			continue
		}
		if _, ok := addrCoin[utils.Bytes2string(addrInfo.Addr)]; ok {
			continue
		} else {
			addrCoin[utils.Bytes2string(addrInfo.Addr)] = false
		}
		hiOut.InAddr = append(hiOut.InAddr, &addrInfo.Addr)
	}

	//生成新的UTXO收益，保存到列表中
	addrCoin = make(map[string]bool)
	for _, vout := range this.Vout {
		hiOut.OutAddr = append(hiOut.OutAddr, &vout.Address)
		hiOut.Value += vout.Value
		_, ok := Area.Keystore.FindAddress(vout.Address)
		if !ok {
			continue
		}
		hiIn.Value += vout.Value
		if _, ok := addrCoin[utils.Bytes2string(vout.Address)]; ok {
			continue
		} else {
			addrCoin[utils.Bytes2string(vout.Address)] = false
		}
		hiIn.OutAddr = append(hiIn.OutAddr, &vout.Address)
	}

	// 额外记录合约交易事件
	//AddCustomTxEvent(*this.GetHash(), hiIn, hiOut)

	if len(hiOut.InAddr) > 0 {
		balanceHistoryManager.Add(hiOut)
		return
	}
	if len(hiIn.OutAddr) > 0 {
		balanceHistoryManager.Add(hiIn)
	}
}

/*
将输出为0的vout删除
*/
func CleanZeroVouts(vs *[]*Vout) []*Vout {
	vouts := make([]*Vout, 0)
	for _, one := range *vs {
		if one.Value == 0 {
			continue
		}
		vouts = append(vouts, one)
	}
	return vouts
}

/*
将输出为0的vout删除，并且合并相同地址的余额
*/
func MergeVouts(vs *[]*Vout) []*Vout {
	voutMap := make(map[string]*Vout)
	for _, one := range *vs {
		if one.Value == 0 {
			continue
		}
		v, ok := voutMap[utils.Bytes2string(one.Address)+strconv.Itoa(int(one.FrozenHeight))]
		if ok {
			v.Value = v.Value + one.Value
			continue
		}
		//voutMap[utils.Bytes2string(one.Address)+strconv.Itoa(int(one.FrozenHeight))] = (*vs)[i]
		voutMap[utils.Bytes2string(one.Address)+strconv.Itoa(int(one.FrozenHeight))] = &Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
			AddrIsSelf:   one.AddrIsSelf,
			AddrStr:      one.AddrStr,
			Domain:       one.Domain,
			DomainType:   one.DomainType,
		}
	}
	vouts := make([]*Vout, 0)
	for _, v := range voutMap {
		vouts = append(vouts, v)
	}
	return vouts
}

/*
把相同地址的交易输出合并在一起，删除余额为0的输出
*/
func (this *TxBase) MergeVout() {
	this.Vout = MergeVouts(this.GetVout())
	this.Vout_total = uint64(len(this.Vout))
}

/*
删除余额为0的输出
*/
func (this *TxBase) CleanZeroVout() {
	this.Vout = CleanZeroVouts(this.GetVout())
	this.Vout_total = uint64(len(this.Vout))
}

/*
解析交易
*/
func ParseTxBaseProto(txtype uint64, bs *[]byte) (TxItr, error) {
	if bs == nil {
		return nil, nil
	}
	// engine.Log.Info("解析交易 %d", txtype)
	// timeNow := config.TimeNow()
	if txtype == 0 {
		txPay := go_protos.TxPay{}
		err := proto.Unmarshal(*bs, &txPay)
		if err != nil {
			return nil, err
		}
		txtype = txPay.TxBase.Type
	}
	// engine.Log.Info("耗时 111 %s", config.TimeNow().Sub(timeNow))
	// engine.Log.Info("解析交易 %d", txtype)
	var tx interface{}
	switch txtype {
	case config.Wallet_tx_type_mining: //挖矿所得

		txProto := new(go_protos.TxReward)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}

		if txProto.TxBase.Type != config.Wallet_tx_type_mining {
			return nil, errors.New("tx type error")
		}

		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
		for _, one := range txProto.TxBase.Vin {
			// nonce := new(big.Int).SetBytes(one.Nonce)
			vins = append(vins, &Vin{
				// Txid: one.Txid,
				// Vout: one.Vout,
				Puk:  one.Puk,
				Sign: one.Sign,
				// Nonce: nonce,
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
		tx = &Tx_reward{
			TxBase: txBase,
			//Index:  txProto.Index,
			AllReward: txProto.AllReward,
		}

	case config.Wallet_tx_type_deposit_in: //投票参与挖矿输入，余额锁定

		txProto := new(go_protos.TxDepositIn)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}

		if txProto.TxBase.Type != config.Wallet_tx_type_deposit_in {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_deposit_in{
			TxBase: txBase,
			Puk:    txProto.Puk,
			Rate:   uint16(txProto.Rate),
		}

	case config.Wallet_tx_type_deposit_out: //投票参与挖矿输出，余额解锁
		// tx = new(Tx_deposit_out)
		txProto := new(go_protos.TxDepositOut)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		if txProto.TxBase.Type != config.Wallet_tx_type_deposit_out {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_deposit_out{
			TxBase: txBase,
		}
	case config.Wallet_tx_type_pay: //普通支付
		txProto := new(go_protos.TxPay)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		if txProto.TxBase.Type != config.Wallet_tx_type_pay {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_Pay{
			TxBase: txBase,
		}
	case config.Wallet_tx_type_vote_in: //
		txProto := new(go_protos.TxVoteIn)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		if txProto.TxBase.Type != config.Wallet_tx_type_vote_in {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_vote_in{
			TxBase:   txBase,
			Vote:     txProto.Vote,
			VoteType: uint16(txProto.VoteType),
			// VoteAddr: txProto.VoteAddr,
			Rate: uint16(txProto.Rate),
		}

	case config.Wallet_tx_type_vote_out: //
		txProto := new(go_protos.TxVoteOut)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		if txProto.TxBase.Type != config.Wallet_tx_type_vote_out {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_vote_out{
			TxBase:   txBase,
			Vote:     txProto.Vote,             //见证人地址
			VoteType: uint16(txProto.VoteType), //投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
		}
	case config.Wallet_tx_type_voting_reward: //
		txProto := new(go_protos.TxVoteReward)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		if txProto.TxBase.Type != config.Wallet_tx_type_voting_reward {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_Vote_Reward{
			TxBase:      txBase,
			StartHeight: txProto.StartHeight,
			EndHeight:   txProto.EndHeight,
		}
	case config.Wallet_tx_type_nft: //
		txProto := new(go_protos.TxNFT)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		if txProto.TxBase.Type != config.Wallet_tx_type_nft {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_nft{
			TxBase:        txBase,
			NFT_ID:        txProto.ID,
			NFT_Owner:     crypto.AddressCoin(txProto.Owner),
			NFT_name:      string(txProto.Name),      //名称
			NFT_symbol:    string(txProto.Symbol),    //单位
			NFT_resources: string(txProto.Resources), //URI外部资源
		}
	case config.Wallet_tx_type_contract: //合约
		txProto := new(go_protos.TxContract)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		if txProto.TxBase.Type != config.Wallet_tx_type_contract {
			return nil, errors.New("tx type error")
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
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
		tx = &Tx_Contract{
			TxBase:        txBase,
			Action:        txProto.Action,
			GzipSource:    txProto.GzipSource,
			ContractClass: txProto.ContractClass,
			GasPrice:      txProto.GasPrice,
			GzipAbi:       txProto.GzipAbi,
		}
	case config.Wallet_tx_type_multsign_addr: //创建多签地址
		txProto := new(go_protos.TxMultsignAddr)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		mtx := &Tx_Multsign_Addr{}
		if err := mtx.ConvertFrom(txProto); err != nil {
			return nil, err
		}
		return mtx, nil
	case config.Wallet_tx_type_multsign_pay: //多签支付
		txProto := new(go_protos.TxMultsignPay)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		mtx := &Tx_Multsign_Pay{}
		if err := mtx.ConvertFrom(txProto); err != nil {
			return nil, err
		}
		return mtx, nil
	case config.Wallet_tx_type_multsign_name: //多签域名
		txProto := new(go_protos.TxMultsignName)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		mtx := &Tx_Multsign_Name{}
		if err := mtx.ConvertFrom(txProto); err != nil {
			return nil, err
		}
		return mtx, nil
	case config.Wallet_tx_type_deposit_free_gas: //质押免gas
		txProto := new(go_protos.TxDepositFreeGas)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		tx := &Tx_DepositFreeGas{}
		if err := tx.ConvertFrom(txProto); err != nil {
			return nil, err
		}
		return tx, nil
	default:
		tx = GetNewTransaction(txtype, bs)
		if tx == nil {
			//未知交易类型
			engine.Log.Info("Unknown transaction type:%d", txtype)
			return nil, errors.New("Unknown transaction type")
		}
	}
	return tx.(TxItr), nil
}

/*
UTXO输入
*/
type Vin struct {
	// Txid []byte `json:"txid"` //UTXO 前一个交易的id
	// // TxidStr string `json:"-"`    //
	// Vout uint64 `json:"vout"` //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	Puk   []byte  `json:"puk"` //公钥
	Nonce big.Int `json:"n"`   //转出账户nonce
	// PukStr       string             `json:"-"`    //缓存公钥字符串
	PukIsSelf int                `json:"-"` //缓存此公钥是否属于自己钱包。0=未检查;1=不属于自己;2=是属于自己;
	PukToAddr crypto.AddressCoin `json:"-"` //缓存此公钥的地址
	// PukToAddrStr string             `json:"-"`    //缓存此公钥的地址字符串
	Sign []byte `json:"sign"` //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	//	VoutSign []byte `json:"voutsign"` //对本交易的输出签名
}

type VinVO struct {
	// Txid string `json:"txid"` //UTXO 前一个交易的id
	// Vout uint64 `json:"vout"` //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	Addr  string `json:"addr"`
	Puk   string `json:"puk"`  //公钥
	Nonce string `json:"n"`    //转出账户nonce
	Sign  string `json:"sign"` //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	//	VoutSign []byte `json:"voutsign"` //对本交易的输出签名
}

/*
 */
func (this *Vin) ConversionVO() *VinVO {
	addr := crypto.BuildAddr(config.AddrPre, this.Puk)
	vinvo := &VinVO{
		// Txid: hex.EncodeToString(this.Txid), // this.GetTxidStr(),             //UTXO 前一个交易的id
		// Vout: this.Vout,                     //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
		Addr: addr.B58String(),
		Puk:  hex.EncodeToString(this.Puk), // this.GetPukStr(),              //公钥
		// Nonce: hex.EncodeToString(this.Nonce.Bytes()), //转出账户nonce
		Sign: hex.EncodeToString(this.Sign), //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
	// if len(this.Nonce.Bytes()) == 0 {
	vinvo.Nonce = this.Nonce.Text(10) // Format(), //hex.EncodeToString(this.Nonce.Bytes())
	// }
	return vinvo
}

/*
检查公钥是否属于自己
*/
func (this *Vin) CheckIsSelf() bool {
	// engine.Log.Info("判断地址 111 %d", this.PukIsSelf)
	if this.PukIsSelf == 0 {
		// engine.Log.Info("从新判断vin中地址是否属于自己 %d", this.PukIsSelf)
		_, ok := Area.Keystore.FindPuk(this.Puk)
		if ok {
			this.PukIsSelf = 2
			// this.PukToAddrStr = addrInfo.AddrStr
		} else {
			this.PukIsSelf = 1
		}
	}
	// engine.Log.Info("判断地址 222 %d", this.PukIsSelf)
	if this.PukIsSelf == 1 {
		return false
	} else {
		return true
	}
}

/*
获取公钥对应的地址
*/
func (this *Vin) GetPukToAddr() *crypto.AddressCoin {
	if this.PukToAddr == nil || len(this.PukToAddr) <= 0 {
		addr := crypto.BuildAddr(config.AddrPre, this.Puk)
		this.PukToAddr = addr
	}
	return &this.PukToAddr
}

/*
给轻节点使用，初始化AddPre
*/
func InitAddPreForLightNode(addPre string) {
	config.AddrPre = addPre
}

/*
将需要签名的字段序列化
*/
func (this *Vin) SerializeVin() *[]byte {
	bs := make([]byte, 0, len(this.Puk)+len(this.Sign)+len(this.Nonce.Bytes()))
	bs = append(bs, this.Puk...)
	bs = append(bs, this.Sign...)
	bs = append(bs, this.Nonce.Bytes()...)
	return &bs
}

/*
验证地址是否属于自己
*/
func (this *Vin) ValidateAddr() (*crypto.AddressCoin, bool) {
	addr := crypto.BuildAddr(config.AddrPre, this.Puk)
	_, ok := Area.Keystore.FindAddress(addr)
	if !ok {
		return &addr, false
	}
	return &addr, true
}

/*
UTXO输出
*/
type Vout struct {
	Value        uint64             `json:"value"`         //输出金额 = 实际金额 * 100000000
	Address      crypto.AddressCoin `json:"address"`       //钱包地址
	FrozenHeight uint64             `json:"frozen_height"` //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
	AddrIsSelf   int                `json:"-"`             //缓存地址是否属于自己钱包。0=未检查;1=不属于自己;2=是属于自己;
	AddrStr      string             `json:"-"`             //缓存地址字符串
	Domain       []byte             `json:"domain"`
	DomainType   uint64             `json:"domain_type"`
	// Txids        []byte             `json:"txid"`          //本输出被使用后的交易id
}

type VoutVO struct {
	Value        uint64 `json:"value"`         //输出金额 = 实际金额 * 100000000
	Address      string `json:"address"`       //钱包地址
	FrozenHeight uint64 `json:"frozen_height"` //冻结高度。在冻结高度以下，未花费的交易余额不能使用
	// Txids        string `json:"txid"`          //本输出被使用后的交易id
}

/*
 */
func (this *Vout) ConversionVO() *VoutVO {
	return &VoutVO{
		Value:        this.Value,               //输出金额 = 实际金额 * 100000000
		Address:      this.Address.B58String(), //钱包地址
		FrozenHeight: this.FrozenHeight,        //
		// Txids:        hex.EncodeToString(this.Txids), //本输出被使用后的交易id
	}
}

/*
检查地址是否属于自己
*/
func (this *Vout) CheckIsSelf() bool {
	// return true
	if this.AddrIsSelf == 0 {
		_, ok := Area.Keystore.FindAddress(this.Address)
		if ok {
			this.AddrIsSelf = 2
		} else {
			this.AddrIsSelf = 1
		}
	}
	if this.AddrIsSelf == 1 {
		return false
	} else {
		return true
	}
}

/*
获取地址字符串
*/
func (this *Vout) GetAddrStr() string {
	if this.AddrStr == "" {
		this.AddrStr = this.Address.B58String()
	}
	return this.AddrStr
}

/*
将需要签名的字段序列化
*/
func (this *Vout) Serialize() *[]byte {
	bs := make([]byte, 0, len(this.Address)+8+8)
	bs = append(bs, utils.Uint64ToBytes(this.Value)...)
	bs = append(bs, this.Address...)
	bs = append(bs, utils.Uint64ToBytes(this.FrozenHeight)...)
	bs = append(bs, utils.Uint64ToBytes(this.DomainType)...)
	bs = append(bs, []byte(this.Domain)...)
	return &bs
}

/*
全网广播交易
*/
func MulticastTx(txItr TxItr) {
	//		engine.NLog.Debug(engine.LOG_console, "是超级节点发起投票")
	//		log.Println("是超级节点发起投票")
	utils.Go(func() {
		goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
		_, file, line, _ := runtime.Caller(0)
		engine.AddRuntime(file, line, goroutineId)
		defer engine.DelRuntime(file, line, goroutineId)
		// bs, err := txItr.Json()
		bs, err := txItr.Proto()
		if err != nil {
			// engine.Log.Warn("交易json格式化错误，取消广播 %s", txItr.GetHashStr())
			return
		}
		Area.SendMulticastMsg(config.MSGID_multicast_transaction, bs)
	}, nil)

}

/*
通过交易hash解析交易类型
*/
func ParseTxClass(txid []byte) uint64 {
	classBs := txid[:8]
	return utils.BytesToUint64(classBs)
}

/*
*
获取交易的gasused
*/
func GetTxGasUsed(txItr TxItr) (uint64, error) {
	var gasUsed uint64

	var err error
	switch txItr.Class() {
	case config.Wallet_tx_type_contract:
		gasUsed, err = txItr.(*Tx_Contract).PreExecNew()
		if err != nil {
			return 0, err
		}
		break
	default:
		gasUsed = config.DefaultTxToken
	}

	return gasUsed, nil
}

// 获取bloom
func (this *TxBase) GetBloom() []byte {
	return nil
}

// 设置bloom
func (this *TxBase) SetBloom(bs []byte) {

}

func (this *TxBase) CheckAddressBind() error {
	return nil
}

func (this *TxBase) CheckAddressFrozen() bool {
	return CheckAddressFrozenStatus(*this.Vin[0].GetPukToAddr())
}

// 获取comment
func (this *TxBase) GetComment() []byte {
	return this.Comment
}

// 设置comment
func (this *TxBase) SetComment(comment []byte) {
	this.Comment = comment
}
