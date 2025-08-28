/*
保存历史转账纪录
*/
package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/mining/snapshot"
	"web3_gui/chain/protos/go_protos"

	"github.com/gogo/protobuf/proto"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
)

var balanceHistoryManager = NewBalanceHistory()

func init() {
	snapshot.Add(balanceHistoryManager)
}

/*
历史记录管理
*/
type BalanceHistory struct {
	GenerateMaxId           *big.Int  //自增长最高id，保存最新生成的id，可以直接来拿使用
	GenerateMaxIdForAccount *sync.Map //账号自增长最高id
	ForkNo                  uint64    //分叉链id
}

type HistoryItem struct {
	GenerateId *big.Int              //
	IsIn       bool                  //资金转入转出方向，true=转入;false=转出;
	Type       uint64                //交易类型
	InAddr     []*crypto.AddressCoin //输入地址
	OutAddr    []*crypto.AddressCoin //输出地址
	Value      uint64                //交易金额
	Txid       []byte                //交易id
	Height     uint64                //区块高度
	//OutIndex   uint64                //交易输出index，从0开始
	Payload    []byte //
	isCustomTx bool   // 是否自定义伪造交易
	Timestamp  int64  // 时间戳
	BlockHash  []byte
}

func (this *HistoryItem) Proto() ([]byte, error) {

	inaddrs := make([][]byte, 0)
	for _, one := range this.InAddr {
		inaddrs = append(inaddrs, *one)
	}

	outaddrs := make([][]byte, 0)
	for _, one := range this.OutAddr {
		outaddrs = append(outaddrs, *one)
	}

	hip := go_protos.HistoryItem{
		GenerateId: this.GenerateId.Bytes(),
		IsIn:       this.IsIn,
		Type:       this.Type,
		InAddr:     inaddrs,
		OutAddr:    outaddrs,
		Value:      this.Value,
		Txid:       this.Txid,
		Height:     this.Height,
		Payload:    this.Payload,
	}
	return hip.Marshal()
}

func ParseHistoryItem(bs *[]byte) (*HistoryItem, error) {
	if bs == nil {
		return nil, nil
	}
	hip := new(go_protos.HistoryItem)
	err := proto.Unmarshal(*bs, hip)
	if err != nil {
		return nil, err
	}

	inaddrs := make([]*crypto.AddressCoin, 0)
	for _, one := range hip.InAddr {
		addrOne := crypto.AddressCoin(one)
		inaddrs = append(inaddrs, &addrOne)
	}

	outaddrs := make([]*crypto.AddressCoin, 0)
	for _, one := range hip.OutAddr {
		addrOne := crypto.AddressCoin(one)
		outaddrs = append(outaddrs, &addrOne)
	}
	hi := HistoryItem{
		GenerateId: new(big.Int).SetBytes(hip.GenerateId), //
		IsIn:       hip.IsIn,                              //资金转入转出方向，true=转入;false=转出;
		Type:       hip.Type,                              //交易类型
		InAddr:     inaddrs,                               //输入地址
		OutAddr:    outaddrs,                              //输出地址
		Value:      hip.Value,                             //交易金额
		Txid:       hip.Txid,                              //交易id
		Height:     hip.Height,                            //区块高度
		Payload:    hip.Payload,                           //
	}
	return &hi, nil
}

/*
添加一个交易历史记录
*/
func (this *BalanceHistory) Add(hi HistoryItem) error {
	if hi.Height > 1 {
		if hi.isCustomTx {
			for _, outAddr := range hi.OutAddr {
				if err := this.addAccountTxToDb(outAddr, hi.Txid); err != nil {
					engine.Log.Warn("add vout custom tx err:%s", err)
				}
			}
		} else {
			txitr, _, _ := FindTxJsonV1(hi.Txid)
			err := this.addTx(txitr)
			if err != nil {
				engine.Log.Warn("add tx err:%s", err)
			}
		}
	}

	if hi.Type != 1 || !bytes.Equal(*hi.OutAddr[0], precompiled.RewardContract) {
		if hi.GenerateId == nil {
			hi.GenerateId = this.GenerateMaxId
			//创世区块奖励合约部署交易不记录
			if hi.Height > 1 || !bytes.Equal(*hi.OutAddr[0], precompiled.RewardContract) {
				this.GenerateMaxId = new(big.Int).Add(this.GenerateMaxId, big.NewInt(1))
			}
			//this.GenerateMaxId = new(big.Int).Add(this.GenerateMaxId, big.NewInt(1))
		} else {
			if hi.GenerateId.Cmp(this.GenerateMaxId) == 0 {
				if hi.Height > 1 || !bytes.Equal(*hi.OutAddr[0], precompiled.RewardContract) {
					this.GenerateMaxId = new(big.Int).Add(this.GenerateMaxId, big.NewInt(1))
				}
			}
		}

		bs, err := hi.Proto()
		// bs, err := json.Marshal(hi)
		if err != nil {
			return err
		}

		//key := []byte(config.LEVELDB_Head_history_balance + strconv.Itoa(int(this.ForkNo)) + "_" + hi.GenerateId.String())
		key := append(config.LEVELDB_Head_history_balance, []byte(strconv.Itoa(int(this.ForkNo))+"_")...)
		key = append(key, hi.GenerateId.Bytes()...)
		// fmt.Println("key", string(key), "\n", string(bs))
		return db.LevelTempDB.Save(key, &bs)
	}

	return nil
}

func (this *BalanceHistory) addTxReward(hi TxItr) error {
	outs := *hi.GetVout()
	if len(outs) < 1 {
		return errors.New("vouts error" + hex.EncodeToString(*hi.GetHash()))
	}
	outAddr := outs[0].Address
	//解析日志，遍历事件，
	rewardLogs := precompiled.GetRewardHistoryLog(hi.GetLockHeight(), *hi.GetHash())
	for _, v := range rewardLogs {
		//记录自己的即可
		if bytes.Equal(outAddr, crypto.AddressFromB58String(v.Into)) {
			continue
		}
		_, ok := Area.Keystore.FindAddress(crypto.AddressFromB58String(v.Into))
		if !ok {
			continue
		}
		intoAddr := crypto.AddressFromB58String(v.Into)
		if !GetLongChain().WitnessBackup.haveWitness(&intoAddr) {
			continue
		}
		engine.Log.Info("161行%s,%d", v.Into, v.Reward)
		err := this.addAccountTxToDb(&intoAddr, *hi.GetHash())
		if err != nil {
			engine.Log.Warn("add tx to db err:%s", err)
		}
	}
	return nil
}
func (this *BalanceHistory) addTx(hi TxItr) error {
	if hi == nil {
		return errors.New("nil txitr")
	}

	ins := *hi.GetVin()
	if len(ins) < 1 {
		return errors.New("vin error" + hex.EncodeToString(*hi.GetHash()))
	}
	inAddr := ins[0].GetPukToAddr()

	// 记录账号交易记录
	//engine.Log.Error("%d %s %s %s", hi.Class(), inAddr.B58String(), hex.EncodeToString(*hi.GetHash()), (*hi.GetVout())[0].GetAddrStr())
	if err := this.addAccountTxToDb(inAddr, *hi.GetHash()); err != nil {
		return err
	}

	//isErc20 := false
	//outs := *hi.GetVout()
	//if hi.Class() == config.Wallet_tx_type_contract && len(outs) > 0 {
	//	outAddr := outs[0].Address.B58String()
	//	if info := db.GetErc20Info(outAddr); info.Address != "" {
	//		isErc20 = true
	//	}
	//}

	//addrs := make([]crypto.AddressCoin, 0)
	//addrs = append(addrs, inAddr)
	//if isErc20 {
	//	// 记录代币交易记录
	//	payload := hi.GetPayload()
	//	if txClass, toAddr, _ := precompiled.UnpackErc20Payload(payload); txClass != 0 {
	//		err := this.addAccountTxToDb(&toAddr, *hi.GetHash())
	//		if err != nil {
	//			return err
	//		}
	//	}
	//} else {
	for _, out := range *hi.GetVout() {
		if !bytes.Equal(out.Address, *inAddr) {
			err := this.addAccountTxToDb(&out.Address, *hi.GetHash())
			if err != nil {
				engine.Log.Warn("add tx to db err:%s", err)
			}
			//addrs = append(addrs, out.Address)
		}
	}
	//}

	return nil
}

/*
添加账户交易记录
*/
func (this *BalanceHistory) addAccountTxToDb(inAddr *crypto.AddressCoin, hash []byte) error {
	// 记录地址最大索引
	index := uint64(1)
	addBlockIndex, ok := this.GenerateMaxIdForAccount.Load(strings.ToLower(inAddr.B58String()))
	if ok {
		index = addBlockIndex.(uint64) + 1
		this.GenerateMaxIdForAccount.Store(strings.ToLower(inAddr.B58String()), index)
	} else {
		this.GenerateMaxIdForAccount.Store(strings.ToLower(inAddr.B58String()), uint64(1))
	}

	//engine.Log.Info("add history tx %s %s", inAddr.B58String(), hex.EncodeToString(hash))

	//记录地址交易hash和索引
	//addrkey := []byte(config.Address_history_tx + "_" + strings.ToLower(inAddr.B58String()))
	addrkey := append(config.Address_history_tx, []byte("_"+strings.ToLower(inAddr.B58String()))...)
	indexBs := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBs, index)
	_, err := db.LevelTempDB.HSet(addrkey, indexBs, hash)
	if err != nil {
		return err
	}
	return err
}

/*
获取交易历史记录
*/
func (this *BalanceHistory) Get(start *big.Int, total int) []HistoryItem {
	if total == 0 {
		total = config.Wallet_balance_history
	}
	if start == nil {
		start = new(big.Int).Sub(this.GenerateMaxId, big.NewInt(1))
	}
	his := make([]HistoryItem, 0)

	//key := config.LEVELDB_Head_history_balance + strconv.Itoa(int(this.ForkNo)) + "_"
	key := append(config.LEVELDB_Head_history_balance, []byte(strconv.Itoa(int(this.ForkNo))+"_")...)
	for i := 0; i < total; i++ {
		//keyOne := key + new(big.Int).Sub(start, big.NewInt(int64(i))).String()
		id := new(big.Int).Sub(start, big.NewInt(int64(i)))
		if id.Int64() < 0 {
			break
		}
		keyOne := append(key, id.Bytes()...)
		bs, err := db.LevelTempDB.Find([]byte(keyOne))
		if err != nil {
			continue
		}

		hi, err := ParseHistoryItem(bs)
		// hi := new(HistoryItem)

		// // err = json.Unmarshal(*bs, hi)
		// decoder := json.NewDecoder(bytes.NewBuffer(*bs))
		// decoder.UseNumber()
		// err = decoder.Decode(hi)

		if err != nil {
			continue
		}
		his = append(his, *hi)
	}
	return his
}

/*
倒叙获取交易历史记录
*/
func (this *BalanceHistory) GetDesc(start uint64, total int) []HistoryItem {
	if total == 0 {
		total = config.Wallet_balance_history
	}

	his := make([]HistoryItem, 0)

	//key := config.LEVELDB_Head_history_balance + strconv.Itoa(int(this.ForkNo)) + "_"
	key := append(config.LEVELDB_Head_history_balance, []byte(strconv.Itoa(int(this.ForkNo))+"_")...)
	start1 := new(big.Int).Sub(this.GenerateMaxId, big.NewInt(int64(start)))
	for i := 1; i <= total; i++ {
		//keyOne := key + new(big.Int).Sub(start1, big.NewInt(int64(i))).String()
		id := new(big.Int).Sub(start1, big.NewInt(int64(i)))
		if id.Int64() < 0 {
			break
		}
		keyOne := append(key, id.Bytes()...)
		bs, err := db.LevelTempDB.Find([]byte(keyOne))
		if err != nil || bs == nil {
			continue
		}
		hi, err := ParseHistoryItem(bs)

		if err != nil {
			continue
		}
		his = append(his, *hi)
	}
	return his
}

// 额外记录交易细化事件
func AddCustomTxEvent(txid []byte) {
}

/*
// 额外记录社区分账合约交易事件，因为合约内部有循环，交易细节只能通过事件获取
func AddCustomTxEvent(txid []byte, hiIn HistoryItem, hiOut HistoryItem) {
	txItr, code, blockHash := FindTx(txid)
	if blockHash == nil {
		return
	}

	if code != 2 {
		return
	}

	//获取区块信息
	bh, err := LoadBlockHeadByHash(blockHash)
	if err != nil {
		return
	}

	//需要排除已经添加的 地址->交易
	addrm := make(map[string]struct{})
	if len(hiOut.InAddr) > 0 {
		addrm[utils.Bytes2string(*hiOut.InAddr[0])] = struct{}{}
		if len(hiOut.OutAddr) > 0 {
			addrm[utils.Bytes2string(*hiOut.OutAddr[0])] = struct{}{}
		}
	} else if len(hiIn.OutAddr) > 0 {
		addrm[utils.Bytes2string(*hiIn.OutAddr[0])] = struct{}{}
		if len(hiIn.InAddr) > 0 {
			addrm[utils.Bytes2string(*hiIn.InAddr[0])] = struct{}{}
		}
	}

	//1.奖励事件
	rewardLogs := precompiled.GetRewardHistoryLog(bh.Height, txid)
	for _, v := range rewardLogs {
		if v.Utype != 2 && v.Utype != 3 {
			continue
		}

		//addr := crypto.AddressCoin{}
		intos := []*crypto.AddressCoin{}
		if v.Into != "" {
			into := crypto.AddressFromB58String(v.Into)
			//排除已经添加的 地址->交易
			if _, ok := addrm[utils.Bytes2string(into)]; ok {
				continue
			}
			intos = append(intos, &into)
			//addr = into
		}

		froms := []*crypto.AddressCoin{}
		from := precompiled.RewardContract
		froms = append(froms, &from)

		txType := uint64(0)
		switch v.Utype {
		case 2:
			txType = config.Wallet_tx_type_reward_C
		case 3:
			txType = config.Wallet_tx_type_reward_L
		}

		item := HistoryItem{
			IsIn:       false,     //资金转入转出方向，true=转入;false=转出;
			Type:       txType,    //交易类型
			InAddr:     froms,     //输入地址,对应合约from
			OutAddr:    intos,     //输出地址,对应合约into
			Txid:       txid,      //交易id
			Height:     bh.Height, //
			Value:      v.Reward.Uint64(),
			isCustomTx: true,
			Timestamp:  bh.Time,
			BlockHash:  *blockHash,
		}

		_ = balanceHistoryManager.Add(item)
	}

	//2.ERC20转账事件
	if txClass, toAddr, value := precompiled.UnpackErc20Payload(txItr.GetPayload()); txClass != 0 && txItr.Class() == config.Wallet_tx_type_contract {
		intos := []*crypto.AddressCoin{}
		into := crypto.AddressFromB58String(toAddr.B58String())
		intos = append(intos, &into)

		froms := []*crypto.AddressCoin{}

		item := HistoryItem{
			IsIn:       false,     //资金转入转出方向，true=转入;false=转出;
			Type:       txClass,   //交易类型
			InAddr:     froms,     //输入地址,对应合约from
			OutAddr:    intos,     //输出地址,对应合约into
			Txid:       txid,      //交易id
			Height:     bh.Height, //
			Value:      value,
			isCustomTx: true,
			Timestamp:  bh.Time,
			BlockHash:  *blockHash,
		}

		_ = balanceHistoryManager.Add(item)
	}

	//3.域名提现事件
	if txClass, value := precompiled.UnpackEnsPayload(txItr.GetPayload()); txClass == 30 && txItr.Class() == config.Wallet_tx_type_contract {
		ensEventLogs := precompiled.GetEnsHistoryLog(bh.Height, txid)
		for _, v := range ensEventLogs {
			intos := []*crypto.AddressCoin{}
			into := crypto.AddressFromB58String(v.Into)
			intos = append(intos, &into)

			froms := []*crypto.AddressCoin{}
			//from := crypto.AddressFromB58String(v.From)
			//froms = append(froms, &from)

			item := HistoryItem{
				IsIn:       false,     //资金转入转出方向，true=转入;false=转出;
				Type:       txClass,   //交易类型
				InAddr:     froms,     //输入地址,对应合约from
				OutAddr:    intos,     //输出地址,对应合约into
				Txid:       txid,      //交易id
				Height:     bh.Height, //
				Value:      value,
				isCustomTx: true,
				Timestamp:  bh.Time,
				BlockHash:  *blockHash,
			}

			_ = balanceHistoryManager.Add(item)
		}
	}
}
*/

// 额外记录奖励合约交易事件，因为合约内部有循环，交易细节只能通过事件获取
func CustomRewardTxEvent(txid []byte) {
	txIter, code, blockHash := FindTx(txid)
	if blockHash == nil {
		return
	}

	if code != 2 {
		return
	}

	//获取区块信息
	bh, err := LoadBlockHeadByHash(blockHash)
	if err != nil {
		return
	}

	vinAddr := crypto.AddressCoin{}
	if vins := *(txIter.GetVin()); len(vins) == 0 {
		return
	} else {
		vinAddr = *(vins[0].GetPukToAddr())
	}

	isRewardTx := false
	rewardLogs := precompiled.GetRewardHistoryLog(bh.Height, txid)
	for _, v := range rewardLogs {
		if v.Utype == 1 {
			if _, ok := Area.Keystore.FindAddress(vinAddr); ok {
				isRewardTx = true
				break
			}
		}
	}

	if !isRewardTx {
		return
	}

	// 记录真实交易
	vins := []*crypto.AddressCoin{}
	for _, vin := range *txIter.GetVin() {
		vins = append(vins, vin.GetPukToAddr())
	}

	vouts := []*crypto.AddressCoin{}
	for _, vout := range *txIter.GetVout() {
		vouts = append(vouts, &vout.Address)
	}

	txType := uint64(config.Wallet_tx_type_mining)
	hiOut := HistoryItem{
		IsIn:       false,  //资金转入转出方向，true=转入;false=转出;
		Type:       txType, //交易类型
		InAddr:     vins,
		OutAddr:    vouts,
		Txid:       txid,      //交易id
		Height:     bh.Height, //
		Value:      0,
		isCustomTx: true,
		Timestamp:  bh.Time,
		BlockHash:  *blockHash,
	}

	_ = balanceHistoryManager.Add(hiOut)

	/*
		rewardLogs := precompiled.GetRewardHistoryLog(bh.Height, txid)
		for _, v := range rewardLogs {
			if v.Utype != 1 {
				continue
			}

			//addr := crypto.AddressCoin{}
			intos := []*crypto.AddressCoin{}
			if v.Into != "" {
				into := crypto.AddressFromB58String(v.Into)
				intos = append(intos, &into)
				//addr = into
			}

			froms := []*crypto.AddressCoin{}
			// from := crypto.AddressFromB58String(v.From)
			from := precompiled.RewardContract
			froms = append(froms, &from)
			engine.Log.Warn("%s", from.B58String())

			txType := uint64(0)
			switch v.Utype {
			case 1:
				txType = config.Wallet_tx_type_reward_W
			case 2:
				txType = config.Wallet_tx_type_reward_C
			case 3:
				txType = config.Wallet_tx_type_reward_L
			}

			hiOut := HistoryItem{
				IsIn:       false,     //资金转入转出方向，true=转入;false=转出;
				Type:       txType,    //交易类型
				InAddr:     froms,     //输入地址,对应合约from
				OutAddr:    intos,     //输出地址,对应合约into
				Txid:       txid,      //交易id
				Height:     bh.Height, //
				Value:      v.Reward.Uint64(),
				isCustomTx: true,
				Timestamp:  bh.Time,
				BlockHash:  *blockHash,
			}

			//_ = balanceHistoryManager.addCustomTx(hiOut)
			_ = balanceHistoryManager.Add(hiOut)
		}
	*/
}

func NewBalanceHistory() *BalanceHistory {
	return &BalanceHistory{
		// ForkNo:        forkNo,
		GenerateMaxId:           big.NewInt(0),
		GenerateMaxIdForAccount: new(sync.Map),
	}
}

// 快照碎片对象名称
func (s *BalanceHistory) SnapshotName() string {
	return "balance_history"
}

// 序列化内存对象
func (s *BalanceHistory) SnapshotSerialize() ([]byte, error) {
	generateMaxIdForAccount := make(map[string]uint64)
	s.GenerateMaxIdForAccount.Range(func(key, value any) bool {
		generateMaxIdForAccount[key.(string)] = value.(uint64)
		return true
	})

	sbh := go_protos.SnapshotBalanceHistory{
		GenerateMaxId:           s.GenerateMaxId.Bytes(),
		GenerateMaxIdForAccount: generateMaxIdForAccount,
		ForkNo:                  s.ForkNo,
	}

	return sbh.Marshal()
}

// 反序列化,还原内存对象
func (s *BalanceHistory) SnapshotDeSerialize(data []byte) error {
	sbh := go_protos.SnapshotBalanceHistory{}
	if err := sbh.Unmarshal(data); err != nil {
		return err
	}

	generateMaxIdForAccount := new(sync.Map)
	for k, v := range sbh.GenerateMaxIdForAccount {
		generateMaxIdForAccount.Store(k, v)
	}
	s.GenerateMaxId = new(big.Int).SetBytes(sbh.GenerateMaxId)
	s.GenerateMaxIdForAccount = generateMaxIdForAccount
	s.ForkNo = sbh.ForkNo
	return nil
}
