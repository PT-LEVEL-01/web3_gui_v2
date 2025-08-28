package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"golang.org/x/crypto/ed25519"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
nft铸造和转让
*/
type Tx_nft struct {
	TxBase
	NFT_ID        []byte             //NFT唯一id，转让的时候使用
	NFT_Owner     crypto.AddressCoin //拥有者
	NFT_name      string             //名称
	NFT_symbol    string             //单位
	NFT_resources string             //URI外部资源
}

type Tx_nft_VO struct {
	TxBaseVO
	NFT_ID        string `json:"NFT_ID"`        //NFT唯一id，转让的时候使用
	NFT_Owner     string `json:"NFT_Owner"`     //拥有者
	NFT_name      string `json:"NFT_name"`      //名称
	NFT_symbol    string `json:"NFT_symbol"`    //单位
	NFT_resources string `json:"NFT_resources"` //URI外部资源
}

/*
用于地址和txid格式化显示
*/
func (this *Tx_nft) GetVOJSON() interface{} {
	nft := Tx_nft_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		// NFT_ID:        hex.EncodeToString(this.NFT_ID), //
		// NFT_Owner:     this.NFT_Owner.B58String(),      //拥有者
		NFT_name:      this.NFT_name,      //名称
		NFT_symbol:    this.NFT_symbol,    //单位
		NFT_resources: this.NFT_resources, //URI外部资源
	}
	if this.NFT_ID != nil && len(this.NFT_ID) > 0 {
		nft.NFT_ID = hex.EncodeToString(this.NFT_ID)
	}
	if this.NFT_Owner != nil && len(this.NFT_Owner) > 0 {
		nft.NFT_Owner = this.NFT_Owner.B58String()
	}
	return nft
}

/*
构建hash值得到交易id
*/
func (this *Tx_nft) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_nft)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
格式化成[]byte
*/
func (this *Tx_nft) Proto() (*[]byte, error) {
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
	txPay := go_protos.TxNFT{
		TxBase:    &txBase,
		ID:        this.NFT_ID,
		Owner:     this.NFT_Owner,
		Name:      []byte(this.NFT_name),
		Symbol:    []byte(this.NFT_symbol),
		Resources: []byte(this.NFT_resources),
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
func (this *Tx_nft) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(this.NFT_ID)
	buf.Write(this.NFT_Owner)
	buf.Write([]byte(this.NFT_name))
	buf.Write([]byte(this.NFT_symbol))
	buf.Write([]byte(this.NFT_resources))
	*bs = buf.Bytes()
	return bs
}

/*
获取签名
*/
func (this *Tx_nft) GetWaitSign(vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	*signDst = append(*signDst, this.NFT_ID...)
	*signDst = append(*signDst, this.NFT_Owner...)
	*signDst = append(*signDst, []byte(this.NFT_name)...)
	*signDst = append(*signDst, []byte(this.NFT_symbol)...)
	*signDst = append(*signDst, []byte(this.NFT_resources)...)
	return signDst
}

/*
获取签名
*/
func (this *Tx_nft) GetSign(key *ed25519.PrivateKey, vinIndex uint64) *[]byte {
	signDst := this.GetSignSerialize(nil, vinIndex)
	*signDst = append(*signDst, this.NFT_ID...)
	*signDst = append(*signDst, this.NFT_Owner...)
	*signDst = append(*signDst, []byte(this.NFT_name)...)
	*signDst = append(*signDst, []byte(this.NFT_symbol)...)
	*signDst = append(*signDst, []byte(this.NFT_resources)...)
	// engine.Log.Info("验证签名前的字节:%d %s", len(*signDst), hex.EncodeToString(*signDst))
	sign := keystore.Sign(*key, *signDst)
	return &sign
}

/*
检查交易是否合法
*/
func (this *Tx_nft) CheckSign() error {

	if this.Vin == nil || len(this.Vin) != 1 {
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
	// inTotal := uint64(0)
	for i, one := range this.Vin {
		signDst := this.GetSignSerialize(nil, uint64(i))
		//本交易特有信息
		*signDst = append(*signDst, this.NFT_ID...)
		*signDst = append(*signDst, this.NFT_Owner...)
		*signDst = append(*signDst, []byte(this.NFT_name)...)
		*signDst = append(*signDst, []byte(this.NFT_symbol)...)
		*signDst = append(*signDst, []byte(this.NFT_resources)...)
		// engine.Log.Info("验证签名前的字节:%d %s", len(*signDst), hex.EncodeToString(*signDst))
		puk := ed25519.PublicKey(one.Puk)
		// fmt.Printf("sign后:puk:%x signdst:%x sign:%x", md5.Sum(puk), md5.Sum(*signDst), md5.Sum(one.Sign))
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*signDst))
		}
		if !ed25519.Verify(puk, *signDst, one.Sign) {
			return config.ERROR_sign_fail
		}

	}
	return nil
}

/*
	验证是否合法
*/
// func (this *Tx_nft) GetWitness() *crypto.AddressCoin {
// 	witness := crypto.BuildAddr(config.AddrPre, this.Vin[0].Puk)
// 	// witness, err := keystore.ParseHashByPubkey(this.Vin[0].Puk)
// 	// if err != nil {
// 	// 	return nil
// 	// }
// 	return &witness
// }

/*
	设置投票地址
*/
// func (this *Tx_nft) SetVoteAddr(addr crypto.AddressCoin) {
// 	// this.VoteAddr = addr
// 	// bs, err := this.Json()
// 	bs, err := this.Proto()
// 	if err != nil {
// 		return
// 	}
// 	// TxCache.FlashTxInCache(hex.EncodeToString(*this.GetHash()), this)
// 	// TxCache.FlashTxInCache(this.GetHashStr(), this)
// 	db.LevelDB.Save(*this.GetHash(), bs)
// }

/*
获取本交易总共花费的余额
*/
func (this *Tx_nft) GetSpend() uint64 {
	return this.Gas
}

/*
检查重复的交易
*/
func (this *Tx_nft) CheckRepeatedTx(txs ...TxItr) bool {

	//如果是铸造，则无冲突
	if this.NFT_ID == nil || len(this.NFT_ID) == 0 {
		return true
	}

	for _, one := range txs {
		if one.Class() == config.Wallet_tx_type_nft {
			//判断重复转让
			nftOne := one.(*Tx_nft)
			if nftOne.NFT_ID == nil || len(nftOne.NFT_ID) == 0 {
				continue
			}
			if bytes.Equal(nftOne.NFT_ID, this.NFT_ID) {
				return false
			}
		}
		if one.Class() == config.Wallet_tx_type_nft_exchange {
			//判断重复交换
			nftOne := one.(*Tx_nft_exchange)
			if bytes.Equal(nftOne.NFT_ID_Sponsor, this.NFT_ID) {
				return false
			}
			if bytes.Equal(nftOne.NFT_ID_Recipient, this.NFT_ID) {
				return false
			}
		}
		if one.Class() == config.Wallet_tx_type_nft_destroy {
			nftOne := one.(*Tx_nft_destroy)
			if bytes.Equal(nftOne.NFT_ID, this.NFT_ID) {
				return false
			}
		}
	}
	return true
}

/*
	统计交易余额
*/
// func (this *Tx_vote_in) CountTxItems(height uint64) *TxItemCount {
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
// 		if voutIndex == 0 {
// 			continue
// 		}
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
func (this *Tx_nft) CountTxItemsNew(height uint64) *TxItemCountMap {
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, len(this.Vout)+len(this.Vin)),
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

func (this *Tx_nft) CountTxHistory(height uint64) {
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
	}
	//转入历史记录
	// hiIn := HistoryItem{
	// 	IsIn:    true,                           //资金转入转出方向，true=转入;false=转出;
	// 	Type:    this.Class(),                   //交易类型
	// 	InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
	// 	OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
	// 	// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
	// 	// Value:  amount,          //交易金额
	// 	Txid:   *this.GetHash(), //交易id
	// 	Height: height,          //
	// 	// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
	// }
	//
	addrCoin := make(map[string]bool)
	for _, vin := range this.Vin {
		addrInfo, isSelf := Area.Keystore.FindPuk(vin.Puk)
		// hiIn.InAddr = append(hiIn.InAddr, &addrInfo.Addr)
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
	for voutIndex, vout := range this.Vout {
		if voutIndex != 0 {
			continue
		}
		hiOut.OutAddr = append(hiOut.OutAddr, &vout.Address)
		hiOut.Value += vout.Value
		_, ok := Area.Keystore.FindAddress(vout.Address)
		if !ok {
			continue
		}
		// hiIn.Value += vout.Value
		if _, ok := addrCoin[utils.Bytes2string(vout.Address)]; ok {
			continue
		} else {
			addrCoin[utils.Bytes2string(vout.Address)] = false
		}
		// hiIn.OutAddr = append(hiIn.OutAddr, &vout.Address)
	}
	if len(hiOut.InAddr) > 0 {
		balanceHistoryManager.Add(hiOut)
	}
	// if len(hiIn.OutAddr) > 0 {
	// 	balanceHistoryManager.Add(hiIn)
	// }
}

/*
铸造一个NFT
NFT_ID        []byte             //NFT唯一id，转让的时候使用
NFT_Owner     crypto.AddressCoin //拥有者
NFT_name      string             //名称
NFT_symbol    string             //单位
NFT_resources string             //URI外部资源
*/
func CreateTxNFT(srcAddress *crypto.AddressCoin, gas uint64, pwd, comment string,
	owner *crypto.AddressCoin, name, symbol, resources string) (*Tx_nft, error) {
	// engine.Log.Info("start CreateTxNFT")
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}
	chain := forks.GetLongChain()
	currentHeight := chain.GetCurrentBlock()

	//查找余额
	vins := make([]*Vin, 0)
	total, item := chain.Balance.BuildPayVinNew(srcAddress, gas)
	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}

	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)
	engine.Log.Info("地址nonce:%v", nonce)
	vin := Vin{
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)), //
		Puk:   puk,                                      //公钥
	}
	// engine.Log.Info("新交易nonce:%d", vin.Nonce.Uint64())
	vins = append(vins, &vin)

	//构建交易输出
	// vouts := make([]*Vout, 0)

	var pay *Tx_nft
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_nft,                       //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: 0,                                               //输出交易数量
			Vout:       nil,                                             //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
			Comment:    []byte{},
		}
		pay = &Tx_nft{
			TxBase: base,
			// NFT_ID        []byte             //NFT唯一id，转让的时候使用
			NFT_Owner:     *owner,    //拥有者
			NFT_name:      name,      //名称
			NFT_symbol:    symbol,    //单位
			NFT_resources: resources, //URI外部资源
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
	// engine.Log.Info("交易签名 耗时 %s", config.TimeNow().Sub(start))
	// chain.Balance.Frozen(items, pay)
	chain.Balance.AddLockTx(pay)
	engine.Log.Info("end CreateTxNFT %s", hex.EncodeToString(*pay.GetHash()))
	return pay, nil
}

/*
转移一个NFT
NFT_ID        []byte             //NFT唯一id，转让的时候使用
NFT_Owner     crypto.AddressCoin //拥有者
NFT_name      string             //名称
NFT_symbol    string             //单位
NFT_resources string             //URI外部资源
*/
func TransferTxNFT(gas uint64, pwd, comment string, nftid []byte, owner *crypto.AddressCoin) (*Tx_nft, error) {
	// engine.Log.Info("start TransferTxNFT")
	//查询这个NFT的所属者
	oldOwner, err := FindNFTOwner(nftid)
	if err != nil {
		return nil, err
	}
	//检查这个nft是否属于自己
	_, ok := Area.Keystore.FindAddress(*oldOwner)
	if !ok {
		return nil, config.ERROR_tx_nft_not_our_own
	}

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}
	chain := forks.GetLongChain()
	currentHeight := chain.GetCurrentBlock()

	//查找余额
	vins := make([]*Vin, 0)
	total, item := chain.Balance.BuildPayVinNew(oldOwner, gas)
	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}

	puk, ok := Area.Keystore.GetPukByAddr(*item.Addr)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}
	nonce := chain.GetBalance().FindNonce(item.Addr)
	vin := Vin{
		Nonce: *new(big.Int).Add(&nonce, big.NewInt(1)), //
		Puk:   puk,                                      //公钥
	}
	// engine.Log.Info("新交易nonce:%d", vin.Nonce.Uint64())
	vins = append(vins, &vin)

	//构建交易输出
	// vouts := make([]*Vout, 0)

	var pay *Tx_nft
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_nft,                       //交易类型
			Vin_total:  uint64(len(vins)),                               //输入交易数量
			Vin:        vins,                                            //交易输入
			Vout_total: 0,                                               //输出交易数量
			Vout:       nil,                                             //交易输出
			Gas:        gas,                                             //交易手续费
			LockHeight: currentHeight + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                       //
			Comment:    []byte{},
		}
		pay = &Tx_nft{
			TxBase:    base,
			NFT_ID:    nftid,  //NFT唯一id，转让的时候使用
			NFT_Owner: *owner, //拥有者
			// NFT_name:      name,      //名称
			// NFT_symbol:    symbol,    //单位
			// NFT_resources: resources, //URI外部资源
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
	// engine.Log.Info("交易签名 耗时 %s", config.TimeNow().Sub(start))
	// chain.Balance.Frozen(items, pay)
	chain.Balance.AddLockTx(pay)
	engine.Log.Info("end TransferTxNFT %s", hex.EncodeToString(*pay.GetHash()))
	return pay, nil
}
