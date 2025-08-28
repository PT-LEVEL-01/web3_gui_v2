package mining

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
	"math"
	"math/big"
	"sort"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/base58"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
多签交易
*/
type Tx_Multsign_Addr struct {
	TxBase
	MultAddress crypto.AddressCoin `json:"m_a"` //多签地址
	RandNum     *big.Int           `json:"r_n"` //随机数
	MultVins    []*Vin             `json:"m_v"` //多签集合,保证顺序,验证 2f/3 + 1
}

/*
多签交易
*/
type Tx_Multsign_Addr_VO struct {
	TxBaseVO
	MultAddress string   `json:"mult_address"` //多签地址
	MultPuk     string   `json:"mult_puk"`     //多签虚拟公钥
	MultVins    []*VinVO `json:"mult_vins"`    //多签集合,保证顺序,验证 2f/3 + 1
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_Multsign_Addr) GetVOJSON() interface{} {
	multVins := []*VinVO{}
	puks := [][]byte{}
	for _, vin := range this.MultVins {
		multVins = append(multVins, vin.ConversionVO())
		puks = append(puks, vin.Puk)
	}

	multPukBytes := mergePublicKeys(this.RandNum, puks...)
	multPuk := string(base58.Encode(multPukBytes))
	txMultsign := Tx_Multsign_Addr_VO{
		TxBaseVO:    this.ConversionVO(),
		MultAddress: this.MultAddress.B58String(),
		MultPuk:     multPuk,
		MultVins:    multVins,
	}
	return txMultsign
}

/*
转化为VO对象
*/
func (this *Tx_Multsign_Addr) ConversionVO() TxBaseVO {
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
func (this *Tx_Multsign_Addr) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_multsign_addr)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_Multsign_Addr) Proto() (*[]byte, error) {
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
	for _, one := range this.MultVins {
		multVins = append(multVins, &go_protos.Vin{
			Puk:  one.Puk,
			Sign: one.Sign,
		})
	}
	txPay := go_protos.TxMultsignAddr{
		TxBase:      &txBase,
		MultAddress: this.MultAddress,
		MultVins:    multVins,
		RandNum:     this.RandNum.Bytes(),
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
func (this *Tx_Multsign_Addr) ConvertFrom(txProto *go_protos.TxMultsignAddr) error {
	if txProto.TxBase.Type != config.Wallet_tx_type_multsign_addr {
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
	this.MultAddress = txProto.MultAddress
	this.RandNum = new(big.Int).SetBytes(txProto.RandNum)
	this.MultVins = multVins
	return nil
}

/*
序列化
*/
func (this *Tx_Multsign_Addr) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(this.MultAddress)
	buf.Write(this.RandNum.Bytes())
	for _, vin := range this.MultVins {
		buf.Write(vin.Puk)
	}
	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *Tx_Multsign_Addr) GetWaitSign(vinIndex uint64) *[]byte {
	buf := bytes.NewBuffer(*this.GetSignSerialize(nil, 0))
	buf.Write(this.MultAddress)
	buf.Write(this.RandNum.Bytes())
	for _, vin := range this.MultVins {
		buf.Write(vin.Puk)
	}
	buf.Write(utils.Uint64ToBytes(vinIndex))
	bs := buf.Bytes()
	return &bs
}

/*
获取签名
*/
func (this *Tx_Multsign_Addr) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	buf := this.GetWaitSign(vinIndex)
	sign := keystore.Sign(*key, *buf)
	return &sign
}

/*
检查交易是否合法
*/
func (this *Tx_Multsign_Addr) CheckSign() error {
	// start := config.TimeNow()
	// engine.Log.Info("开始验证交易合法性 Tx_Multsign_Addr")
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
func (this *Tx_Multsign_Addr) GetSpend() uint64 {
	spend := this.Gas
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_Multsign_Addr) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

/*
统计交易余额
仅处理创世块交易,奖励统计查看distributeReward方法
*/
func (this *Tx_Multsign_Addr) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}

	//处理多签燃料费
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

func (this *Tx_Multsign_Addr) CountTxHistory(height uint64) {
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

// 创建多签集合交易
func BuildCreateMultsignAddrTx(src crypto.AddressCoin, gas uint64, pwd string, name string, puks ...[]byte) (*Tx_Multsign_Addr, error) {
	chain := forks.GetLongChain()
	currentHeight := chain.GetCurrentBlock()

	if gas < config.Wallet_tx_gas_min {
		return nil, config.ERROR_tx_gas_too_little
	}

	vins := make([]*Vin, 0)

	total, item := chain.Balance.BuildPayVinNew(&src, gas)
	if total < gas {
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

	//生成多签地址
	sortPuks, multAddress, randNum := GenerateMultAddress(puks...)
	multVins := make([]*Vin, len(puks))
	for i, _ := range sortPuks {
		multVins[i] = &Vin{
			Puk: sortPuks[i],
		}
	}

	//构建交易输出
	vouts := make([]*Vout, 0)

	var pay *Tx_Multsign_Addr
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_multsign_addr, //交易类型
			Vin_total:  uint64(len(vins)),                   //输入交易数量
			Vin:        vins,                                //交易输入
			Vout_total: uint64(len(vouts)),                  //输出交易数量
			Vout:       vouts,                               //交易输出
			Gas:        gas,
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    []byte{},                                        //
			Comment:    []byte(name),
		}
		pay = &Tx_Multsign_Addr{
			TxBase:      base,
			MultAddress: multAddress,
			RandNum:     randNum,
			MultVins:    multVins,
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
	if err := forks.GetLongChain().TransactionManager.AddTx(pay); err != nil {
		GetLongChain().Balance.DelLockTx(pay)
		return nil, errors.Wrap(err, "add tx fail!")
	}

	//广播
	MulticastTx(pay)

	return pay, nil
}

// 查询多签地址成员公钥
func GetMultsignAddrSet(multAddress crypto.AddressCoin) (*go_protos.MultsignSet, bool) {
	bs, ok := ExistMultsignAddrSet(multAddress)
	if !ok {
		return nil, false
	}
	pukSet := &go_protos.MultsignSet{}
	if err := pukSet.Unmarshal(*bs); err != nil {
		return nil, false
	}

	return pukSet, true
}

// 保存多签地址成员公钥
func SaveMultsignAddrSet(pukSet *go_protos.MultsignSet) {
	bs, _ := pukSet.Marshal()
	db.LevelDB.Save(config.BuildMultsignAddrSet(pukSet.MultAddress), &bs)
}

// 判断自己是否多签地址成员
func IsMultAddrParter(multAddress crypto.AddressCoin) bool {
	pukSet, _ := GetMultsignAddrSet(multAddress)
	for _, puk := range pukSet.Puks {
		if _, ok := config.Area.Keystore.FindPuk(puk); ok {
			return true
		}
	}
	return false
}

// 查询多签地址是否存在
func ExistMultsignAddrSet(multAddress crypto.AddressCoin) (*[]byte, bool) {
	bs, err := db.LevelDB.Find(config.BuildMultsignAddrSet(multAddress))
	if err != nil || len(*bs) == 0 {
		return nil, false
	}

	return bs, true
}

// rn 随机数
func mergePublicKeys(rn *big.Int, puks ...[]byte) []byte {
	buf := bytes.NewBuffer(rn.Bytes())
	for _, puk := range puks {
		buf.Write(puk)
	}
	sum := sha256.Sum256(buf.Bytes())
	return sum[:]
}

// 获取多签聚合公钥
func GetMultAddrPublicKey(pukSet *go_protos.MultsignSet) ([]byte, []*Vin) {
	multVins := make([]*Vin, len(pukSet.Puks))
	puks := [][]byte{}
	for i, one := range pukSet.Puks {
		multVins[i] = &Vin{
			Puk: one,
		}
		puks = append(puks, one)
	}

	randNum := new(big.Int).SetBytes(pukSet.RandNum)
	return mergePublicKeys(randNum, puks...), multVins
}

// 生成多签地址
func GenerateMultAddress(puks ...[]byte) ([][]byte, crypto.AddressCoin, *big.Int) {
	randNum, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt))
	//异或排序
	base := [32]byte{}
	puksort := pukSort{
		base: new(big.Int).SetBytes(base[:]),
	}
	for i, _ := range puks {
		puksort.puks = append(puksort.puks, new(big.Int).SetBytes(puks[i]))
	}
	sort.Stable(puksort)
	sortPuks := [][]byte{}
	for _, puk := range puksort.puks {
		sortPuks = append(sortPuks, puk.Bytes())
	}
	sumpuk := mergePublicKeys(randNum, sortPuks...)
	multAddress := crypto.BuildAddr(config.AddrPre, sumpuk)
	if _, ok := ExistMultsignAddrSet(multAddress); ok {
		sortPuks, multAddress, randNum = GenerateMultAddress(sortPuks...)
	}

	return sortPuks, multAddress, randNum
}

// 生成多签地址
func GenerateMultAddressWithoutCheck(puks ...[]byte) ([][]byte, crypto.AddressCoin, *big.Int) {
	randNum, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt))
	//异或排序
	base := [32]byte{}
	puksort := pukSort{
		base: new(big.Int).SetBytes(base[:]),
	}
	for i, _ := range puks {
		puksort.puks = append(puksort.puks, new(big.Int).SetBytes(puks[i]))
	}
	sort.Stable(puksort)
	sortPuks := [][]byte{}
	for _, puk := range puksort.puks {
		sortPuks = append(sortPuks, puk.Bytes())
	}
	sumpuk := mergePublicKeys(randNum, sortPuks...)
	multAddress := crypto.BuildAddr(config.AddrPre, sumpuk)

	return sortPuks, multAddress, randNum
}

type pukSort struct {
	base *big.Int
	puks []*big.Int
}

func (this pukSort) Len() int {
	return len(this.puks)
}

func (this pukSort) Less(i, j int) bool {
	a := new(big.Int).Xor(this.base, this.puks[i])
	b := new(big.Int).Xor(this.base, this.puks[j])
	if a.Cmp(b) > 0 {
		return false
	} else {
		return true
	}
}

func (this pukSort) Swap(i, j int) {
	this.puks[i], this.puks[j] = this.puks[j], this.puks[i]
}
