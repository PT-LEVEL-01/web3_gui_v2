package mining

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/mining/name"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/utils"
)

const (
	NameActionNone       = iota
	NameInActionReg      //域名注册
	NameInActionTransfer //域名转让
	NameInActionRenew    //域名续费
	NameInActionUpdate   //域名更新
	NameOutAction        //域名注销
)

/*
多签域名交易
*/
type Tx_Multsign_Name struct {
	TxBase
	DefaultMultTx
	Account             []byte                 `json:"a"`   //账户名称
	NetIds              []nodeStore.AddressNet `json:"n"`   //网络地址列表
	NetIdsMerkleHash    []byte                 `json:"nmh"` //网络地址默克尔树hash
	AddrCoins           []crypto.AddressCoin   `json:"as"`  //网络地址列表
	AddrCoinsMerkleHash []byte                 `json:"amh"` //网络地址默克尔树hash
	NameActionType      int                    `json:"nt"`
}

/*
多签域名交易
*/
type Tx_Multsign_Name_VO struct {
	TxBaseVO
	MultAddress         string   `json:"mult_address"`          //多签地址
	MultVins            []*VinVO `json:"mult_vins"`             //多签集合,保证顺序,验证 2f/3 + 1
	Account             string   `json:"account"`               //账户名称
	NetIds              []string `json:"netids"`                //网络地址列表
	NetIdsMerkleHash    string   `json:"netids_merkle_hash"`    //网络地址默克尔树hash
	AddrCoins           []string `json:"addrcoins"`             //网络地址列表
	AddrCoinsMerkleHash string   `json:"addrcoins_merkle_hash"` //网络地址默克尔树hash
	NameActionType      int      `json:"name_action_type"`
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_Multsign_Name) GetVOJSON() interface{} {
	multVins := []*VinVO{}
	for _, vin := range this.GetMultVins() {
		multVins = append(multVins, vin.ConversionVO())
	}
	multAddress := this.GetMultAddress()

	netids := make([]string, 0)
	for _, one := range this.NetIds {
		netids = append(netids, one.B58String())
	}
	addrs := make([]string, 0)
	for _, one := range this.AddrCoins {
		addrs = append(addrs, one.B58String())
	}

	txMultsign := Tx_Multsign_Name_VO{
		TxBaseVO:            this.ConversionVO(),
		MultAddress:         multAddress.B58String(),
		MultVins:            multVins,
		Account:             string(this.Account),
		NetIds:              netids,
		NetIdsMerkleHash:    hex.EncodeToString(this.NetIdsMerkleHash), //网络地址默克尔树hash
		AddrCoins:           addrs,
		AddrCoinsMerkleHash: hex.EncodeToString(this.AddrCoinsMerkleHash),
		NameActionType:      this.NameActionType,
	}
	return txMultsign
}

/*
转化为VO对象
*/
func (this *Tx_Multsign_Name) ConversionVO() TxBaseVO {
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
func (this *Tx_Multsign_Name) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_multsign_name)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_Multsign_Name) Proto() (*[]byte, error) {
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

	netids := make([][]byte, 0)
	for _, one := range this.NetIds {
		netids = append(netids, one)
	}
	addrCoins := make([][]byte, 0)
	for _, one := range this.AddrCoins {
		addrCoins = append(addrCoins, one)
	}

	txPay := go_protos.TxMultsignName{
		TxBase:              &txBase,
		MultAddress:         this.GetMultAddress(),
		MultVins:            multVins,
		Account:             this.Account,
		NetIds:              netids,
		NetIdsMerkleHash:    this.NetIdsMerkleHash,
		AddrCoins:           addrCoins,
		AddrCoinsMerkleHash: this.AddrCoinsMerkleHash,
		NameActionType:      int32(this.NameActionType),
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
func (this *Tx_Multsign_Name) ConvertFrom(txProto *go_protos.TxMultsignName) error {
	if txProto.TxBase.Type != config.Wallet_tx_type_multsign_name {
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
	netids := []nodeStore.AddressNet{}
	for _, one := range txProto.NetIds {
		netids = append(netids, one)
	}
	addrCoins := []crypto.AddressCoin{}
	for _, one := range txProto.AddrCoins {
		addrCoins = append(addrCoins, one)
	}

	this.TxBase = txBase
	this.DefaultMultTx = NewDefaultMultTxImpl(txProto.MultAddress, multVins)
	this.Account = txProto.Account
	this.NetIds = netids
	this.NetIdsMerkleHash = txProto.NetIdsMerkleHash
	this.AddrCoins = addrCoins
	this.AddrCoinsMerkleHash = txProto.AddrCoinsMerkleHash
	this.NameActionType = int(txProto.NameActionType)
	return nil
}

/*
序列化
*/
func (this *Tx_Multsign_Name) Serialize() *[]byte {
	buf := bytes.NewBuffer(*this.TxBase.Serialize())
	buf.Write(this.GetMultAddress())
	for _, vin := range this.GetMultVins() {
		buf.Write(vin.Puk)
	}
	buf.Write(this.Account)
	for _, v := range this.NetIds {
		buf.Write(v)
	}
	buf.Write(this.NetIdsMerkleHash)
	for _, v := range this.AddrCoins {
		buf.Write(v)
	}
	buf.Write(this.AddrCoinsMerkleHash)
	buf.Write(utils.Int64ToBytes(int64(this.NameActionType)))
	bs := buf.Bytes()
	return &bs
}

/*
获取签名
*/
func (this *Tx_Multsign_Name) GetWaitSign(vinIndex uint64) *[]byte {
	buf := bytes.NewBuffer(*this.GetSignSerialize(nil, 0))
	buf.Write(this.GetMultAddress())
	for _, vin := range this.GetMultVins() {
		buf.Write(vin.Puk)
	}
	buf.Write(this.Account)
	for _, v := range this.NetIds {
		buf.Write(v)
	}
	buf.Write(this.NetIdsMerkleHash)
	for _, v := range this.AddrCoins {
		buf.Write(v)
	}
	buf.Write(this.AddrCoinsMerkleHash)
	buf.Write(utils.Int64ToBytes(int64(this.NameActionType)))

	buf.Write(utils.Uint64ToBytes(vinIndex))
	bs := buf.Bytes()
	return &bs
}

/*
获取签名
*/
func (this *Tx_Multsign_Name) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	buf := this.GetWaitSign(vinIndex)
	sign := keystore.Sign(*key, *buf)
	return &sign
}

/*
检查交易是否合法
*/
func (this *Tx_Multsign_Name) CheckSign() error {
	// start := config.TimeNow()
	// engine.Log.Info("开始验证交易合法性 Tx_Multsign_Name")
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
		engine.Log.Info("Check Mult-Signature: %s Total:%d Pass:%d", hex.EncodeToString(this.Hash), total, verifyCount)
	}

	return nil
}

/*
获取本交易总共花费的余额
*/
func (this *Tx_Multsign_Name) GetSpend() uint64 {
	spend := this.Gas
	for _, vout := range this.Vout {
		spend += vout.Value
	}
	return spend
}

/*
是否验证通过
*/
func (this *Tx_Multsign_Name) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

// 检查锁定高度
func (this *Tx_Multsign_Name) CheckLockHeight(lockHeight uint64) error {
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
func (this *Tx_Multsign_Name) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
		Nonce:    make(map[string]big.Int),
	}

	//处理多签燃料费
	totalValue := this.Gas
	switch this.NameActionType {
	case NameInActionReg: //域名注册
		//处理质押
		for _, vout := range this.Vout {
			totalValue += vout.Value
		}
		//case NameInActionTransfer: //域名转让
		//case NameInActionRenew: //域名续费
		//case NameInActionUpdate: //域名更新
		//case NameOutAction: //域名注销
	}

	////处理多签燃料费
	//totalValue := this.Gas
	//for _, vout := range this.Vout {
	//	totalValue += vout.Value
	//	frozenMap, ok := itemCount.AddItems[utils.Bytes2string(vout.Address)]
	//	if ok {
	//		oldValue, ok := (*frozenMap)[vout.FrozenHeight]
	//		if ok {
	//			oldValue += int64(vout.Value)
	//			(*frozenMap)[vout.FrozenHeight] = oldValue
	//		} else {
	//			(*frozenMap)[vout.FrozenHeight] = int64(vout.Value)
	//		}
	//	} else {
	//		frozenMap := make(map[uint64]int64, 0)
	//		frozenMap[vout.FrozenHeight] = int64(vout.Value)
	//		itemCount.AddItems[utils.Bytes2string(vout.Address)] = &frozenMap
	//	}
	//}

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

func (this *Tx_Multsign_Name) CountTxHistory(height uint64) {
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
func BuildRequestMultsignNameTx(multAddress crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd string, account string, netids []nodeStore.AddressNet, addrCoins []crypto.AddressCoin, nameActionType int) (*Tx_Multsign_Name, error) {
	chain := GetLongChain()
	currentHeight := chain.GetCurrentBlock()

	if gas < config.Wallet_tx_gas_min {
		return nil, config.ERROR_tx_gas_too_little
	}

	var vinAddress crypto.AddressCoin
	var voutAddress crypto.AddressCoin
	switch nameActionType {
	case NameInActionReg:
		//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
		if account == "" {
			return nil, config.ERROR_params_fail
		}
		if strings.Contains(account, ".") || strings.Contains(account, " ") {
			return nil, config.ERROR_params_fail
		}
		//判断域名是否已经注册
		nameinfo := name.FindNameToNet(account)
		if nameinfo != nil {
			return nil, config.ERROR_name_exist
		}
		if amount < config.Mining_name_deposit_min {
			return nil, config.ERROR_name_deposit
		}
		//资金不够
		total, _ := chain.Balance.BuildPayVinNew(&multAddress, amount)
		if total < amount+gas {
			return nil, config.ERROR_not_enough
		}
		if len(addrCoins) == 0 {
			return nil, config.ERROR_params_fail
		}
		vinAddress = multAddress
		voutAddress = multAddress

	case NameInActionTransfer:
		//查找域名是否属于自己
		nameInfo := name.FindName(account)
		if nameInfo == nil {
			return nil, config.ERROR_tx_name_not_owner
		}
		//资金不够
		total, _ := chain.Balance.BuildPayVinNew(&nameInfo.Owner, amount)
		if total < gas {
			return nil, config.ERROR_not_enough
		}

		vinAddress = nameInfo.Owner
		voutAddress = multAddress
		multAddress = nameInfo.Owner //确保这个地址是多签地址
		//amount = nameInfo.Deposit
		amount = 0
		netids = nameInfo.NetIds
		addrCoins = nameInfo.AddrCoins

	case NameInActionRenew:
		//查找域名是否属于自己
		nameInfo := name.FindName(account)
		if nameInfo == nil {
			return nil, config.ERROR_tx_name_not_owner
		}
		if !bytes.Equal(nameInfo.Owner, multAddress) {
			return nil, config.ERROR_tx_name_not_owner
		}
		//资金不够
		total, _ := chain.Balance.BuildPayVinNew(&nameInfo.Owner, amount)
		if total < gas {
			return nil, config.ERROR_not_enough
		}

		vinAddress = nameInfo.Owner
		voutAddress = nameInfo.Owner
		//amount = nameInfo.Deposit
		amount = 0
		netids = nameInfo.NetIds
		addrCoins = nameInfo.AddrCoins

	case NameInActionUpdate:
		if len(addrCoins) == 0 {
			return nil, config.ERROR_params_fail
		}
		//查找域名是否属于自己
		nameInfo := name.FindName(account)
		if nameInfo == nil {
			return nil, config.ERROR_tx_name_not_owner
		}
		if !bytes.Equal(nameInfo.Owner, multAddress) {
			return nil, config.ERROR_tx_name_not_owner
		}
		//资金不够
		total, _ := chain.Balance.BuildPayVinNew(&nameInfo.Owner, amount)
		if total < gas {
			return nil, config.ERROR_not_enough
		}

		vinAddress = nameInfo.Owner
		voutAddress = nameInfo.Owner
		//amount = nameInfo.Deposit
		amount = 0
	case NameOutAction:
		//查找域名是否属于自己
		nameInfo := name.FindName(account)
		if nameInfo == nil {
			return nil, config.ERROR_tx_name_not_owner
		}
		if !bytes.Equal(nameInfo.Owner, multAddress) {
			return nil, config.ERROR_tx_name_not_owner
		}
		//资金不够
		total, _ := chain.Balance.BuildPayVinNew(&nameInfo.Owner, amount)
		if total < gas {
			return nil, config.ERROR_not_enough
		}

		vinAddress = nameInfo.Owner
		voutAddress = nameInfo.Owner
		//amount = nameInfo.Deposit
		amount = 0
	default:
		return nil, config.ERROR_params_fail
	}

	//查找余额
	vins := make([]*Vin, 0)

	//查询多签
	pukSet, ok := GetMultsignAddrSet(vinAddress)
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
		Address:      voutAddress,  //钱包地址
		FrozenHeight: frozenHeight, //
	}
	vouts = append(vouts, &vout)

	netidsBs := make([][]byte, 0)
	for _, one := range netids {
		netidsBs = append(netidsBs, one)
	}

	addrCoinBs := make([][]byte, 0)
	for _, one := range addrCoins {
		addrCoinBs = append(addrCoinBs, one)
	}

	var mtx *Tx_Multsign_Name
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_multsign_name,                      //交易类型
			Vin_total:  uint64(len(vins)),                                        //输入交易数量
			Vin:        vins,                                                     //交易输入
			Vout_total: uint64(len(vouts)),                                       //输出交易数量
			Vout:       vouts,                                                    //交易输出
			Gas:        gas,                                                      //交易手续费
			LockHeight: currentHeight + config.Wallet_multsign_tx_lockHeight + i, //锁定高度
			Payload:    []byte{},                                                 //
		}
		mtx = &Tx_Multsign_Name{
			TxBase:              base,
			DefaultMultTx:       NewDefaultMultTxImpl(multAddress, multVins),
			Account:             []byte(account),
			NetIds:              netids,
			NetIdsMerkleHash:    utils.BuildMerkleRoot(netidsBs),
			AddrCoins:           addrCoins,
			AddrCoinsMerkleHash: utils.BuildMerkleRoot(addrCoinBs),
			NameActionType:      nameActionType,
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
