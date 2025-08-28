package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"golang.org/x/sync/errgroup"
	"math/big"
	"runtime"
	"sync/atomic"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/precompiled/ens"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/snapshot"
	"web3_gui/chain_boot/chain_plus"
	chainBootConfig "web3_gui/chain_boot/config"
	"web3_gui/chain_boot/model"
	"web3_gui/chain_boot/object_beans"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func GetInfo(params *map[string]interface{}) engine.PostResult {
	value, valuef, valuelockup := mining.FindBalanceValue()

	//tbs := mining.ListTokenInfos()
	//tbVOs := make([]*mining.TokenInfoV0, 0)
	//for _, one := range tbs {
	//	tbVO := mining.ToTokenInfoV0(one)
	//	tbVOs = append(tbVOs, tbVO)
	//}

	currentBlock := uint64(0)
	startBlock := uint64(0)
	heightBlock := uint64(0)
	pulledStates := uint64(0)
	startBlockTime := uint64(0)
	syncBlockFinish := false
	chain := mining.GetLongChain()
	if chain != nil {
		currentBlock = chain.GetCurrentBlock()
		startBlock = chain.GetStartingBlock()
		heightBlock = mining.GetHighestBlock()
		pulledStates = chain.GetPulledStates()
		startBlockTime = chain.GetStartBlockTime()
		syncBlockFinish = chain.SyncBlockFinish
	}

	info := model.Getinfo{
		//Netid:          []byte(config.AddrPre),   //
		TotalAmount:    config.Mining_coin_total, //
		Balance:        value,                    //
		BalanceFrozen:  valuef,                   //
		BalanceLockup:  valuelockup,              //
		BalanceVote:    GetBalanceVote(),
		Testnet:        true,                                           //
		Blocks:         currentBlock,                                   //
		Group:          0,                                              //
		StartingBlock:  startBlock,                                     //区块开始高度
		StartBlockTime: startBlockTime,                                 //
		HighestBlock:   heightBlock,                                    //所链接的节点的最高高度
		CurrentBlock:   currentBlock,                                   //已经同步到的区块高度
		PulledStates:   pulledStates,                                   //正在同步的区块高度
		SnapshotHeight: snapshot.Height(),                              //快照高度
		BlockTime:      uint64(config.Mining_block_time.Nanoseconds()), //出块时间 pl time
		LightNode:      config.Mining_light_min,                        //轻节点押金数量
		CommunityNode:  config.Mining_vote,                             //社区节点押金数量
		WitnessNode:    config.Mining_deposit,                          //见证人押金数量
		NameDepositMin: config.Mining_name_deposit_min,                 //
		AddrPre:        config.AddrPre,                                 //
		//TokenInfos:      tbVOs,                                          //
		SyncBlockFinish: syncBlockFinish,
		ContractAddress: precompiled.RewardContract.B58String(),
	}
	result := engine.NewPostResult()
	result.Data["info"] = info
	return *result
}

/*
获取区块高度
*/
func BlockHeight(params *map[string]interface{}) engine.PostResult {
	pr := engine.NewPostResult()
	pr.Data["CurrentHeight"] = mining.GetCurrentHeight()
	pr.Data["HighestBlock"] = mining.GetHighestBlock()
	pr.Code = utils.ERROR_CODE_success
	return *pr
}

/*
查询节点的当前累计的见证人+社区+轻节点投票奖励池
*/
func GetBalanceVote(addrs ...coin_address.AddressCoin) uint64 {
	wlist := []coin_address.AddressCoin{} //见证人地址集
	clist := []coin_address.AddressCoin{} //社区地址集
	llist := []coin_address.AddressCoin{} //轻节点地址集
	if len(addrs) != 0 {
		for _, one := range addrs {
			addrType := mining.GetAddrState(crypto.AddressCoin(one))
			switch addrType {
			case 1: // 见证人
				wlist = append(wlist, one)
			case 2: // 社区
				clist = append(clist, one)
			case 3: // 轻节点
				llist = append(llist, one)
			}
		}
	} else {
		for _, one := range config.Area.Keystore.GetAddrAll() {
			addrType := mining.GetAddrState(one.Addr)
			switch addrType {
			case 1: // 见证人
				wlist = append(wlist, coin_address.AddressCoin(one.Addr))
			case 2: // 社区
				clist = append(clist, coin_address.AddressCoin(one.Addr))
			case 3: // 轻节点
				llist = append(llist, coin_address.AddressCoin(one.Addr))
			}
		}
	}

	eg := errgroup.Group{}
	eg.SetLimit(runtime.NumCPU())
	wPoolValue := atomic.Uint64{} //见证人奖励池
	cPoolValue := atomic.Uint64{} //社区奖励池
	lPoolValue := atomic.Uint64{} //轻节点奖励池

	chain := mining.GetLongChain()
	if chain == nil {
		return wPoolValue.Load() + cPoolValue.Load() + lPoolValue.Load()
	}

	balanceMgr := chain.Balance

	for _, one := range wlist {
		addr := one
		eg.Go(func() error {
			witRewardPools := balanceMgr.GetWitnessRewardPool()
			if witRewardPool, ok := witRewardPools.Load(utils.Bytes2string(addr)); ok {
				value, _ := balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(crypto.AddressCoin(addr), witRewardPool.(*big.Int))
				wPoolValue.Add(value.Uint64())
			}
			return nil
		})
	}

	for _, one := range clist {
		addr := one
		eg.Go(func() error {
			value, _ := balanceMgr.CalculateCommunityRewardAndLightReward(crypto.AddressCoin(addr))
			cPoolValue.Add(value.Uint64())
			return nil
		})
	}

	for _, one := range llist {
		addr := one
		eg.Go(func() error {
			addrOne := crypto.AddressCoin(addr)
			if lightvoteinfo := balanceMgr.GetDepositVote(&addrOne); lightvoteinfo != nil {
				_, lightrewards := balanceMgr.CalculateCommunityRewardAndLightReward(lightvoteinfo.WitnessAddr)
				if value, ok := lightrewards[utils.Bytes2string(lightvoteinfo.SelfAddr)]; ok {
					lPoolValue.Add(value.Uint64())
				}
			}
			return nil
		})
	}

	eg.Wait()

	return wPoolValue.Load() + cPoolValue.Load() + lPoolValue.Load()
}

/*
获取本节点总余额
*/
func GetBalance(params *map[string]interface{}) engine.PostResult {
	pr := engine.NewPostResult()
	value, valuef, valuelockup := mining.FindBalanceValue()
	type Getinfo struct {
		Balance       uint64 `json:"balance"`       //可用余额
		BalanceFrozen uint64 `json:"BalanceFrozen"` //冻结的余额
		BalanceLockup uint64 `json:"BalanceLockup"` //锁仓的余额
	}
	info := Getinfo{
		Balance:       value,       //
		BalanceFrozen: valuef,      //
		BalanceLockup: valuelockup, //
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["data"] = info
	return *pr
}

/*
获取本节点地址列表
*/
func AddressList(params *map[string]interface{}, startAddr string, total uint64) engine.PostResult {
	pr := engine.NewPostResult()
	if total == 0 {
		total = 100
	}
	if total > 100 {
		total = 100
	}

	var addrStart *coin_address.AddressCoin
	if startAddr != "" {
		addrItr := (*params)["startAddr1"]
		addr := addrItr.(coin_address.AddressCoin)
		addrStart = &addr
	}

	list := config.Area.Keystore.GetAddr()
	if addrStart == nil {
		start := 0
		end := total
		if total > uint64(len(list)) {
			end = uint64(len(list))
		}
		list = list[start:end]
	} else {
		for i, one := range list {
			if addrStart != nil {
				if bytes.Equal(one.Addr, *addrStart) {
					list = list[i:]
					break
				}
			}
		}
		if uint64(len(list)) > total {
			list = list[:total]
		}
	}

	vos := make([]model.AccountVO, 0)
	for i, one := range list {
		ba, fba, baLockup := mining.GetBalanceForAddrSelf(one.Addr)
		addrType := mining.GetAddrState(one.Addr)
		//特殊处理因黑名单踢出的见证人
		//首地址
		if i == 0 && addrType == 4 {
			if depositVal := mining.GetDepositWitnessAddr(&one.Addr); depositVal > 0 {
				addrType = 1
			}
		}

		mainAddr := ""
		if bs, err := db.LevelDB.Get(config.BuildAddressTxBindKey(one.Addr)); err == nil {
			a := crypto.AddressCoin(bs)
			mainAddr = a.B58String()
		}

		subAddrs := []string{}
		if pairs, err := db.LevelDB.HGetAll(one.Addr); err == nil {
			for _, pair := range pairs {
				a := crypto.AddressCoin(pair.Field)
				subAddrs = append(subAddrs, a.B58String())
			}
		}

		vo := model.AccountVO{
			Index:               i,
			Name:                one.Nickname,
			AddrCoin:            one.GetAddrStr(),
			MainAddrCoin:        mainAddr,
			SubAddrCoins:        subAddrs,
			Type:                addrType,
			Value:               ba,       //可用余额
			ValueFrozen:         fba,      //冻结余额
			ValueLockup:         baLockup, //
			BalanceVote:         GetBalanceVote(coin_address.AddressCoin(one.Addr)),
			AddressFrozenStatus: mining.CheckAddressFrozenStatus(one.Addr),
		}
		vos = append(vos, vo)
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["data"] = vos
	return *pr
}

/*
创建一个地址
*/
func AddressCreate(params *map[string]interface{}, nickname, seedPassword, addrPassword string) engine.PostResult {
	utils.Log.Info().Str("创建地址", "").Send()
	pr := engine.NewPostResult()

	keyst, ERR := chain_plus.GetKeystore()
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	addrInfo, ERR := keyst.CreateCoinAddr(nickname, seedPassword, addrPassword)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//addr, err := keystAapter.GetNewAddr(seedPassword, addrPassword)
	//if err != nil {
	//	ERR := utils.NewErrorSysSelf(err)
	//	pr.Code = ERR.Code
	//	//pr.Data["addr"] = addr.B58String()
	//	return *pr
	//}
	ERR = utils.NewErrorSuccess()
	//addr, ERR := keys.CreateCoinAddr(nickname, seedPassword, addrPassword)
	pr.Code = ERR.Code
	pr.Data["addr"] = addrInfo.B58String()
	return *pr
}

func AddressNonce(params *map[string]interface{}, address string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)
	caddr := crypto.AddressCoin(addr)
	nonceInt, err := mining.GetAddrNonce(&caddr)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["nonce"] = nonceInt.Uint64()
	return *pr
}

func AddressNonceMore(params *map[string]interface{}, addressMore []string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["addressMore1"]
	addrs := addrItr.([]coin_address.AddressCoin)

	nonces := make([]uint64, len(addrs))
	for i, addr := range addrs {
		addrOne := crypto.AddressCoin(addr)
		nonceInt, err := mining.GetAddrNonce(&addrOne)
		if err != nil {
			ERR := utils.NewErrorSysSelf(err)
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		nonces[i] = nonceInt.Uint64()
	}
	//nonceInt, ERR := mining.GetAddressNonceMore(addrs)
	//if ERR.CheckFail() {
	//	pr.Code = ERR.Code
	//	pr.Msg = ERR.Msg
	//	return *pr
	//}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["nonces"] = nonces
	return *pr
}

func AddressValidate(params *map[string]interface{}, address string) engine.PostResult {
	pr := engine.NewPostResult()
	pr.Code = utils.ERROR_CODE_success
	return *pr
}

func AddressInfo(params *map[string]interface{}, address string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)
	caddr := crypto.AddressCoin(addr)
	ba, fba, baLockup := uint64(0), uint64(0), uint64(0)

	depositIn := uint64(0)
	//社区地址
	ty := mining.GetAddrState(caddr)
	//特殊处理因黑名单踢出的见证人
	//首地址
	if ty == 4 && bytes.Equal(config.Area.Keystore.GetCoinbase().Addr, addr) {
		if depositVal := mining.GetDepositWitnessAddr(&caddr); depositVal > 0 {
			ty = 1
		}
	}

	balanceMgr := mining.GetLongChain().GetBalance()

	var wValue, cValue, lValue uint64 = 0, 0, 0
	switch ty {
	case 1:
		depositIn = config.Mining_deposit
		witRewardPools := balanceMgr.GetWitnessRewardPool()
		if witRewardPool, ok := witRewardPools.Load(utils.Bytes2string(addr)); ok {
			wreward, _ := balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(caddr, witRewardPool.(*big.Int))
			wValue = wreward.Uint64()
		}

	case 2:
		depositIn = config.Mining_vote
		creward, _ := balanceMgr.CalculateCommunityRewardAndLightReward(caddr)
		cValue = creward.Uint64()
	case 3:
		depositIn = config.Mining_light_min
		if itemInfo := balanceMgr.GetDepositVote(&caddr); itemInfo != nil {
			cAddr := itemInfo.WitnessAddr
			_, lightRewards := balanceMgr.CalculateCommunityRewardAndLightReward(cAddr)
			if v, ok := lightRewards[utils.Bytes2string(caddr)]; ok {
				lValue = v.Uint64()
			}
		}
	}

	mainAddr := ""
	if bs, err := db.LevelDB.Get(config.BuildAddressTxBindKey(addr)); err == nil {
		a := crypto.AddressCoin(bs)
		mainAddr = a.B58String()
	}

	subAddrs := []string{}
	if pairs, err := db.LevelDB.HGetAll(addr); err == nil {
		for _, pair := range pairs {
			a := crypto.AddressCoin(pair.Field)
			subAddrs = append(subAddrs, a.B58String())
		}
	}

	keys, ERR := chain_plus.GetKeystore()
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//keys := gconfig.Node.Keystore.(*keystore.Keystore)
	coinAddrInfo := keys.FindCoinAddr(&addr)

	vo := AccountInfo{
		//Name:                 coinAddrInfo.Nickname,
		AddrCoin:             address,
		MainAddrCoin:         mainAddr,
		SubAddrCoins:         subAddrs,
		Value:                ba,       //可用余额
		ValueFrozen:          fba,      //冻结余额
		ValueLockup:          baLockup, //
		BalanceVote:          wValue + cValue + lValue,
		ValueFrozenWitness:   wValue,
		ValueFrozenCommunity: cValue,
		ValueFrozenLight:     lValue,
		DepositIn:            depositIn,
		AddressFrozenStatus:  mining.CheckAddressFrozenStatus(caddr),
		Type:                 ty,
	}

	if coinAddrInfo != nil {
		vo.Name = coinAddrInfo.Nickname
	}

	pr.Code = utils.ERROR_CODE_success
	pr.Data["data"] = vo
	return *pr
}

type AccountInfo struct {
	Name                 string   //名称
	AddrCoin             string   //收款地址
	MainAddrCoin         string   //主地址收款地址
	SubAddrCoins         []string //从地址收款地址
	Value                uint64   //可用余额
	ValueFrozen          uint64   //冻结余额
	ValueLockup          uint64   //
	BalanceVote          uint64   //当前奖励总金额
	ValueFrozenWitness   uint64   //见证人节点冻结奖励
	ValueFrozenCommunity uint64   //社区节点冻结奖励
	ValueFrozenLight     uint64   //轻节点冻结奖励
	DepositIn            uint64   //质押数量
	AddressFrozenStatus  bool     //地址绑定冻结状态
	Type                 int      //1=见证人;2=社区节点;3=轻节点;4=什么也不是
}

func AddressBalance(params *map[string]interface{}, address string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)
	caddr := crypto.AddressCoin(addr)
	_, ba := db.GetNotSpendBalance(&caddr)
	pr.Data["balance"] = ba
	return *pr
}

func AddressBalanceMore(params *map[string]interface{}, addressMore []string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["addressMore1"]
	addrs := addrItr.([]coin_address.AddressCoin)
	values, ERR := db.FindNotSpendBalances(addrs)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//查询锁定余额
	valuesFrozen := make([]uint64, len(addressMore))
	for i, addr := range addressMore {
		addrOne := crypto.AddressCoin(addr)
		frozenValue := mining.GetAddrFrozenValue(&addrOne)
		valuesFrozen[i] = frozenValue
	}
	pr.Data["BalanceNotSpend"] = values
	pr.Data["BalanceFrozen"] = valuesFrozen
	return *pr
}

func AddressAllBalanceRange(params *map[string]interface{}, startAddr string, total uint64) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr, ok := (*params)["startAddr1"]
	if ok {

	}
	addrs := addrItr.([]coin_address.AddressCoin)
	values, ERR := db.FindNotSpendBalances(addrs)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["BalanceNotSpend"] = values
	return *pr
}

func AddressVoteInfo(params *map[string]interface{}, addressMore []string) engine.PostResult {
	pr := engine.NewPostResult()

	addrItr := (*params)["addressMore1"]
	addrs := addrItr.([]coin_address.AddressCoin)

	addrInfos := []*model.AddrVoteInfo{}
	communityAddrs := []crypto.AddressCoin{}
	lightAddrs := []crypto.AddressCoin{}
	addrIndex := make(map[string]int)
	for i, addrOne := range addrs {
		addrCoin := crypto.AddressCoin(addrOne)
		// 查余额
		value, valueFrozen, _ := mining.GetNotspendByAddrOther(mining.GetLongChain(), addrCoin)
		// 查询地址角色
		role := mining.GetAddrState(addrCoin)
		addrInfo := &model.AddrVoteInfo{
			Balance:       value,
			BalanceFrozen: valueFrozen,
			Role:          role,
		}

		switch role {
		case 1: // 见证人
			if mining.GetDepositWitnessAddr(&addrCoin) > 0 {
				role = 1
				addrInfo.DepositIn = config.Mining_deposit
			}
		case 2: // 社区
			communityAddrs = append(communityAddrs, addrCoin)
		case 3: // 轻节点
			lightAddrs = append(lightAddrs, addrCoin)
		default: // 无
		}
		addrIndex[addrCoin.B58String()] = i
		addrInfos = append(addrInfos, addrInfo)
	}

	balanceMgr := mining.GetLongChain().GetBalance()
	//社区节点
	if len(communityAddrs) > 0 {
		for _, addr := range communityAddrs {
			itemInfo := balanceMgr.GetDepositCommunity(&addr)
			if itemInfo != nil {
				i := addrIndex[addr.B58String()]
				addrInfos[i].DepositIn = config.Mining_vote
				addrInfos[i].VoteAddr = itemInfo.WitnessAddr.B58String()
				addrInfos[i].VoteIn = itemInfo.Value
			}
		}
	}

	//轻节点
	if len(lightAddrs) > 0 {
		for _, addr := range lightAddrs {
			itemInfo := balanceMgr.GetDepositVote(&addr)
			if itemInfo != nil {
				i := addrIndex[addr.B58String()]
				addrInfos[i].DepositIn = config.Mining_light_min
				addrInfos[i].VoteAddr = itemInfo.WitnessAddr.B58String()
				addrInfos[i].VoteIn = itemInfo.Value
			}
		}
	}
	pr.Data["Infos"] = addrInfos
	return *pr
}

func TransactionRecord(params *map[string]interface{}, address string, page, total uint64) engine.PostResult {
	pr := engine.NewPostResult()
	count, datas := chain_plus.TransactionRecord(address, int(page), int(total))
	pr.Data["count"] = count
	pr.Data["data"] = datas
	return *pr
}

func MnemonicImport(params *map[string]interface{}, words, newSeedPassword string) engine.PostResult {
	pr := engine.NewPostResult()

	//keystAapter := gconfig.Node.Keystore.(*kva.Keystore)
	//keys, ERR := keystAapter.GetV2Keystore()
	//if ERR.CheckFail() {
	//	pr.Code = ERR.Code
	//	pr.Msg = ERR.Msg
	//	return *pr
	//}

	wallet := keystore.NewWallet(config.KeystoreFileAbsPath, config.AddrPre)
	ERR := wallet.ImportMnemonic(words, newSeedPassword, newSeedPassword, newSeedPassword, newSeedPassword)
	pr.Code = ERR.Code
	pr.Msg = ERR.Msg
	return *pr
}

func MnemonicExport(params *map[string]interface{}, seedPassword string) engine.PostResult {
	pr := engine.NewPostResult()

	keys, ERR := chain_plus.GetKeystore()
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	words, ERR := keys.ExportMnemonic(seedPassword)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//words, err := config.Area.Keystore.ExportMnemonic(seedPassword)
	//if err != nil {
	//	ERR := utils.NewErrorSysSelf(err)
	//	pr.Code = ERR.Code
	//	pr.Msg = ERR.Msg
	//	return *pr
	//}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["words"] = words
	return *pr
}

func MnemonicImportEncry(params *map[string]interface{}, words, wordsPassword, newSeedPassword string) engine.PostResult {
	pr := engine.NewPostResult()
	err := config.Area.Keystore.ImportMnemonicEncry(words, wordsPassword, newSeedPassword, newSeedPassword, newSeedPassword)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["words"] = words
	return *pr
}

func MnemonicExportEncry(params *map[string]interface{}, seedPassword, wordsPassword string) engine.PostResult {
	pr := engine.NewPostResult()
	words, err := config.Area.Keystore.ExportMnemonicEncry(seedPassword, wordsPassword)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["words"] = words
	return *pr
}

/*
转账
为了提高性能，接口不对地址正确性验证，前缀正确性验证，请提前验证了再请求
*/
func SendToAddress(params *map[string]interface{}, srcaddress, address string, amount, gas, frozen_height uint64,
	pwd, comment string, domain string, domain_type uint64) engine.PostResult {
	pr := engine.NewPostResult()
	var src coin_address.AddressCoin
	if srcaddress != "" {
		srcAddrItr := (*params)["srcaddress1"]
		src = srcAddrItr.(coin_address.AddressCoin)
	}
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	var commentBs []byte
	if comment != "" {
		commentBs = []byte(comment)
	}
	txPay, ERR := chain_plus.CreateTxPay(&src, &addr, amount, gas, frozen_height, pwd, commentBs, domain, domain_type)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	ERR = chain_plus.CurrentLimitingAndMulticastTx(txPay)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["data"] = txPay.GetVOJSON()
	return *pr
}

/*
转账
为了提高性能，接口不对地址正确性验证，前缀正确性验证，请提前验证了再请求
*/
func SendToAddressMore(params *map[string]interface{}, srcAddress string, addressMore []string, amountMore, frozenHeightMore []uint64,
	domainMore []string, domainTypeMore []uint64, gas uint64, pwd string, comment string) engine.PostResult {
	pr := engine.NewPostResult()
	var src coin_address.AddressCoin
	if srcAddress != "" {
		srcAddrItr := (*params)["srcAddress1"]
		src = srcAddrItr.(coin_address.AddressCoin)
	}
	addrItr := (*params)["addressMore1"]
	addrs := addrItr.([]coin_address.AddressCoin)

	var commentBs []byte
	if comment != "" {
		commentBs = []byte(comment)
	}

	txPay, ERR := chain_plus.CreateTxPayMore(src, addrs, amountMore, frozenHeightMore, domainMore, domainTypeMore, gas, pwd, commentBs)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	ERR = chain_plus.CurrentLimitingAndMulticastTx(txPay)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["data"] = txPay.GetVOJSON()
	return *pr
}

func PayOrder(params *map[string]interface{}, srcaddress, serverAddr, orderId16 string, amount, gas uint64, pwd string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	var src coin_address.AddressCoin
	if srcaddress != "" {
		srcAddrItr := (*params)["srcaddress1"]
		src = srcAddrItr.(coin_address.AddressCoin)
	}
	addrItr := (*params)["serverAddr1"]
	dst := addrItr.(coin_address.AddressCoin)

	orderItr := (*params)["orderId161"]
	orderId := orderItr.([]byte)

	order := object_beans.NewCommonOrder(orderId)
	bs, err := order.Proto()
	if err != nil {
		pr.Code = utils.ERROR_CODE_system_error_self
		pr.Msg = err.Error()
		return *pr
	}
	txPay, ERR := chain_plus.CreateTxPay(&src, &dst, amount, gas, 0, pwd, *bs, "", 0)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	ERR = chain_plus.CurrentLimitingAndMulticastTx(txPay)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["orderId16"] = orderId16
	pr.Data["data"] = txPay.GetVOJSON()
	return *pr
}

/*
给多人转账
*/
func PushTxProto64(params *map[string]interface{}, base64StdStr string, checkBalance bool) engine.PostResult {
	pr := engine.NewPostResult()
	txItr, ERR := chain_plus.CheckTxBase64(base64StdStr, checkBalance)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	err := mining.AddTx(txItr)
	if err != nil {
		pr.Code = utils.ERROR_CODE_system_error_self
		pr.Msg = err.Error()
		return *pr
	}
	pr.Data["info"] = txItr
	return *pr
}

func TxStatusOnChain(params *map[string]interface{}, txHash16 string) engine.PostResult {
	pr := engine.NewPostResult()
	txHashItr := (*params)["txHash161"]
	txHash := txHashItr.([]byte)

	_, code, _, blockHeight, timestamp := mining.FindTxJsonVo(txHash)

	tx, err := mining.LoadTxBase(txHash)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	pr.Code = utils.ERROR_CODE_success
	status := TxStatus{
		TxInfo:      tx,                                     //
		Bloom:       hex.EncodeToString(tx.GetBloom()),      //
		UpChainCode: code,                                   //1=未上链;2=成功上链;3=上链失败;
		BlockHeight: blockHeight,                            //
		BlockHash:   hex.EncodeToString(*tx.GetBlockHash()), //
		Timestamp:   timestamp,                              //
	}
	pr.Data["TxStatus"] = status
	return *pr
}

func TxStatusOnChainMore(params *map[string]interface{}, txHashMore16 []string) engine.PostResult {
	pr := engine.NewPostResult()
	txHashItr := (*params)["txHashMore161"]
	txHashMore := txHashItr.([][]byte)
	txStatus := make([]TxStatus, 0, len(txHashMore))
	for _, txHash := range txHashMore {
		_, code, _, blockHeight, timestamp := mining.FindTxJsonVo(txHash)
		tx, err := mining.LoadTxBase(txHash)
		if err != nil {
			ERR := utils.NewErrorSysSelf(err)
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}

		pr.Code = utils.ERROR_CODE_success
		status := TxStatus{
			TxInfo:      tx,                                     //
			Bloom:       hex.EncodeToString(tx.GetBloom()),      //
			UpChainCode: code,                                   //1=未上链;2=成功上链;3=上链失败;
			BlockHeight: blockHeight,                            //
			BlockHash:   hex.EncodeToString(*tx.GetBlockHash()), //
			Timestamp:   timestamp,                              //
		}
		txStatus = append(txStatus, status)
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["TxStatusMore"] = txStatus
	return *pr
}

type TxStatus struct {
	TxInfo      mining.TxItr //
	Bloom       string       //
	UpChainCode uint64       //1=未上链;2=成功上链;3=上链失败;
	BlockHeight uint64       //
	BlockHash   string       //
	Timestamp   int64        //
}

func TxProto64ByHash16(params *map[string]interface{}, txHash16 string) engine.PostResult {
	pr := engine.NewPostResult()
	txHashItr := (*params)["txHash161"]
	txHash := txHashItr.([]byte)

	tx, err := mining.LoadTxBase(txHash)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	txbs, err := tx.Proto()
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["Tx"] = base64.StdEncoding.EncodeToString(*txbs)
	return *pr
}

func TxJsonByHash16(params *map[string]interface{}, txHash16 string) engine.PostResult {
	pr := engine.NewPostResult()
	txHashItr := (*params)["txHash161"]
	txHash := txHashItr.([]byte)

	tx, err := mining.LoadTxBase(txHash)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["Tx"] = tx.GetVOJSON()
	return *pr
}

func BlockProto64ByHash16(params *map[string]interface{}, blockHash16 string) engine.PostResult {
	pr := engine.NewPostResult()
	blockHashItr := (*params)["blockHash161"]
	blockHash := blockHashItr.([]byte)

	bh, err := mining.LoadBlockHeadByHash(&blockHash)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	if bh == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	txbs, err := bh.Proto()
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["BlockProto64"] = base64.StdEncoding.EncodeToString(*txbs)
	return *pr
}

func BlockJsonByHash16(params *map[string]interface{}, blockHash16 string) engine.PostResult {
	pr := engine.NewPostResult()
	blockHashItr := (*params)["blockHash161"]
	blockHash := blockHashItr.([]byte)

	bh, err := mining.LoadBlockHeadByHash(&blockHash)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	if bh == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	reward := config.ClacRewardForBlockHeightFun(bh.Height)
	//reward := toUint64(config.BLOCK_TOTAL_REWARD)
	gas := uint64(0)
	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
		txItr, _ := mining.LoadTxBase(one)
		if txItr != nil {
			gas += txItr.GetGas()
		}
	}

	reward += gas

	extSigns := make([]string, len(bh.ExtSign))
	for k, v := range bh.ExtSign {
		extSigns[k] = hex.EncodeToString(v)
	}

	//查询区块的bloom
	bloom := ""
	if bloomBs, err := mining.GetBlockHeadBloom(bh.Height); err == nil {
		bloom = hex.EncodeToString(bloomBs)
	}

	bhvo := BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		ExtSign:           extSigns,
		Reward:            reward,
		//Gas:               reward - toUint64(config.BLOCK_TOTAL_REWARD),
		Gas:     gas,
		Destroy: 0,
		Bloom:   bloom,
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["BlockJson"] = bhvo
	return *pr
}

type BlockHeadVO struct {
	Hash              string   //区块头hash
	Height            uint64   //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
	GroupHeight       uint64   //矿工组高度
	GroupHeightGrowth uint64   //组高度增长量。默认0为自动计算增长量（兼容之前的区块）,最少增量为1
	Previousblockhash string   //上一个区块头hash
	Nextblockhash     string   //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
	NTx               uint64   //交易数量
	MerkleRoot        string   //交易默克尔树根hash
	Tx                []string //本区块包含的交易id
	Time              int64    //出块时间，unix时间戳
	Witness           string   //此块见证人地址
	Sign              string   //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	ExtSign           []string //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	Reward            uint64   //出块奖励
	Gas               uint64   //燃料费
	Destroy           uint64   //销毁
	Bloom             string   //bloom过滤器
}

func BlockProto64ByHeight(params *map[string]interface{}, height uint64) engine.PostResult {
	pr := engine.NewPostResult()
	bh := mining.LoadBlockHeadByHeight(height)
	if bh == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	bhvo := mining.BlockHeadVO{}
	bhvo.BH = bh
	bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))
	for _, one := range bh.Tx {
		txItr, e := mining.LoadTxBase(one)
		if e != nil {
			ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_found, "交易:"+hex.EncodeToString(one))
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		bhvo.Txs = append(bhvo.Txs, txItr)
	}
	bs, e := bhvo.Proto()
	if e != nil {
		ERR := utils.NewErrorSysSelf(e)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["BlockProto64"] = base64.StdEncoding.EncodeToString(*bs)
	return *pr
}

func BlockJsonByHeight(params *map[string]interface{}, height uint64) engine.PostResult {
	pr := engine.NewPostResult()
	bh := mining.LoadBlockHeadByHeight(height)
	if bh == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	reward := config.ClacRewardForBlockHeightFun(bh.Height)
	//reward := toUint64(config.BLOCK_TOTAL_REWARD)
	gas := uint64(0)
	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
		txItr, _ := mining.LoadTxBase(one)
		if txItr != nil {
			gas += txItr.GetGas()
		}
	}

	reward += gas

	extSigns := make([]string, len(bh.ExtSign))
	for k, v := range bh.ExtSign {
		extSigns[k] = hex.EncodeToString(v)
	}

	//查询区块的bloom
	bloom := ""
	if bloomBs, err := mining.GetBlockHeadBloom(bh.Height); err == nil {
		bloom = hex.EncodeToString(bloomBs)
	}

	bhvo := BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		ExtSign:           extSigns,
		Reward:            reward,
		//Gas:               reward - toUint64(config.BLOCK_TOTAL_REWARD),
		Gas:     gas,
		Destroy: 0,
		Bloom:   bloom,
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["BlockJson"] = bhvo
	return *pr
}

func BlocksProto64ByHeightRange(params *map[string]interface{}, startHeight, total uint64) engine.PostResult {
	pr := engine.NewPostResult()
	endHeight := startHeight + total

	//待返回的区块
	bhvos := make([]*[]byte, 0, endHeight-startHeight+1)

	for i := startHeight; i <= endHeight; i++ {
		bhvo := mining.BlockHeadVO{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}
		bhvo.BH = bh
		bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))

		for _, one := range bh.Tx {
			txItr, e := mining.LoadTxBase(one)
			// txItr, e := mining.FindTxBase(one)
			if e != nil {
				ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_found, "交易:"+hex.EncodeToString(one))
				pr.Code = ERR.Code
				pr.Msg = ERR.Msg
				return *pr
			}
			bhvo.Txs = append(bhvo.Txs, txItr)
		}
		bs, e := bhvo.Proto()
		if e != nil {
			ERR := utils.NewErrorSysSelf(e)
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		bhvos = append(bhvos, bs)
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["BlocksProto64"] = bhvos
	return *pr
}

// 返回交易类型,vin,vout
// 通过实际高度解析历史交易
func DealTxInfo(tx mining.TxItr, addr string, blockheight uint64) (uint64, interface{}, interface{}) {
	if tx == nil {
		return 0, nil, nil
	}

	vouts := tx.GetVout()
	out := *vouts
	vins := tx.GetVin()
	vin := *vins

	if addr == "" {
		if len(vin) > 0 {
			addr = vin[0].PukToAddr.B58String()
		}
	}

	//如果是和奖励合约交互的交易
	if bytes.Equal(out[0].Address, precompiled.RewardContract) {
		payload := tx.GetPayload()
		//解析
		txClass, params := precompiled.UnpackPayload(payload)
		//如果类型是1的话，为见证人奖励交易记录，从事件中获取奖励的金额,
		// if tx.Class() == config.Wallet_tx_type_mining {
		// 	rewardLogs := precompiled.GetRewardHistoryLog(blockheight, *tx.GetHash())
		// 	for _, v := range rewardLogs {
		// 		if v.Into == addr {
		// 			returnVins := []mining.VinVO{}
		// 			returnVins = append(returnVins, mining.VinVO{
		// 				Addr: precompiled.RewardContract.B58String(),
		// 			})
		// 			returnVouts := []mining.VoutVO{}
		// 			returnVouts = append(returnVouts, mining.VoutVO{
		// 				Address: addr,
		// 				Value:   v.Reward.Uint64(),
		// 			})
		// 			return config.Wallet_tx_type_reward_W, returnVins, returnVouts
		// 		}
		// 	}
		// }

		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_community_in:
				//社区节点质押,只需要变更类型
				return config.Wallet_tx_type_community_in, nil, nil
			case config.Wallet_tx_type_community_out:
				//社区节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_vote,
				})
				return config.Wallet_tx_type_community_out, returnVins, returnVouts
			case config.Wallet_tx_type_vote_in:
				//轻节点投票
				return config.Wallet_tx_type_vote_in, nil, nil
			case config.Wallet_tx_type_vote_out:
				//轻节点取消投票
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_vote_out, returnVins, returnVouts
			case config.Wallet_tx_type_light_in:
				//轻节点质押
				return config.Wallet_tx_type_light_in, nil, nil
			case config.Wallet_tx_type_light_out:
				//轻节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_light_min,
				})
				return config.Wallet_tx_type_light_out, returnVins, returnVouts
			case config.Wallet_tx_type_reward_C:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_C, returnVins, returnVouts
			case config.Wallet_tx_type_reward_L:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_L, returnVins, returnVouts
			case config.Wallet_tx_type_reward_W:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_W, returnVins, returnVouts
			case config.Wallet_tx_type_community_distribute: //社区分账
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_community_distribute, returnVins, returnVouts
			}
		}

	}

	// 根域名注册合约交互的交易
	if bytes.Equal(out[0].Address, ens.GetRegisterAddr()) {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_domain_register: // 域名注册
				return config.Wallet_tx_type_domain_register, nil, nil
			case config.Wallet_tx_type_domain_renew: // 域名续费
				return config.Wallet_tx_type_domain_renew, nil, nil
			case config.Wallet_tx_type_domain_withdraw: // 域名提现
				rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
				for _, v := range rewardLogs {
					if v.Into == addr {
						returnVins := []mining.VinVO{}
						returnVins = append(returnVins, mining.VinVO{
							Addr: precompiled.RewardContract.B58String(),
						})
						returnVouts := []mining.VoutVO{}
						returnVouts = append(returnVouts, mining.VoutVO{
							Address: addr,
							Value:   v.Amount.Uint64(),
						})

						return config.Wallet_tx_type_domain_withdraw, returnVins, returnVouts
					}
				}
			}
		}
	}

	// 剩下的可能是子域名合约的交易
	if tx.Class() == config.Wallet_tx_type_contract {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		switch txClass {
		case config.Wallet_tx_type_domain_register: // 子域名注册
			return config.Wallet_tx_type_subdomain_register, nil, nil
		case config.Wallet_tx_type_domain_renew: // 子域名续费
			return config.Wallet_tx_type_subdomain_renew, nil, nil
		case config.Wallet_tx_type_domain_withdraw: // 子域名提现
			rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
			for _, v := range rewardLogs {
				if v.Into == addr {
					returnVins := []mining.VinVO{}
					returnVins = append(returnVins, mining.VinVO{
						Addr: precompiled.RewardContract.B58String(),
					})
					returnVouts := []mining.VoutVO{}
					returnVouts = append(returnVouts, mining.VoutVO{
						Address: addr,
						Value:   v.Amount.Uint64(),
					})

					return config.Wallet_tx_type_subdomain_withdraw, returnVins, returnVouts
				}
			}
		}
	}

	return 0, nil, nil
}

// 返回交易类型,vin,vout
// 通过实际高度解析历史交易
// NOTE: 特殊处理了域名提现
func DealTxInfoV3(tx mining.TxItr, addr string, blockheight uint64) (uint64, interface{}, interface{}, bool) {
	if tx == nil {
		return 0, nil, nil, false
	}

	vouts := tx.GetVout()
	out := *vouts
	vins := tx.GetVin()
	vin := *vins

	if addr == "" {
		if len(vin) > 0 {
			addr = vin[0].PukToAddr.B58String()
		}
	}

	//如果是和奖励合约交互的交易
	if bytes.Equal(out[0].Address, precompiled.RewardContract) {
		payload := tx.GetPayload()
		//解析
		txClass, params := precompiled.UnpackPayload(payload)
		//如果类型是1的话，为见证人奖励交易记录，从事件中获取奖励的金额,
		if tx.Class() == config.Wallet_tx_type_mining {
			returnVins := []mining.VinVO{}
			returnVins = append(returnVins, mining.VinVO{
				// Addr: precompiled.RewardContract.B58String(),
				Addr: "",
			})
			return config.Wallet_tx_type_mining, returnVins, nil, false
		}

		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_community_in:
				//社区节点质押,只需要变更类型
				return config.Wallet_tx_type_community_in, nil, nil, false
			case config.Wallet_tx_type_community_out:
				//社区节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_vote,
				})
				return config.Wallet_tx_type_community_out, returnVins, returnVouts, false
			case config.Wallet_tx_type_vote_in:
				//轻节点投票
				return config.Wallet_tx_type_vote_in, nil, nil, false
			case config.Wallet_tx_type_vote_out:
				//轻节点取消投票
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_vote_out, returnVins, returnVouts, false
			case config.Wallet_tx_type_light_in:
				//轻节点质押
				return config.Wallet_tx_type_light_in, nil, nil, false
			case config.Wallet_tx_type_light_out:
				//轻节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   config.Mining_light_min,
				})
				return config.Wallet_tx_type_light_out, returnVins, returnVouts, false
			case config.Wallet_tx_type_reward_C:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_C, returnVins, returnVouts, false
			case config.Wallet_tx_type_reward_L:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_L, returnVins, returnVouts, false
			case config.Wallet_tx_type_reward_W:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_reward_W, returnVins, returnVouts, false
			case config.Wallet_tx_type_community_distribute: //社区分账
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   params,
				})
				return config.Wallet_tx_type_community_distribute, returnVins, returnVouts, false
			}
		}

	}

	// 根域名注册合约交互的交易
	if bytes.Equal(out[0].Address, ens.GetRegisterAddr()) {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			switch txClass {
			case config.Wallet_tx_type_domain_register: // 域名注册
				return config.Wallet_tx_type_domain_register, nil, nil, true
			case config.Wallet_tx_type_domain_renew: // 域名续费
				return config.Wallet_tx_type_domain_renew, nil, nil, true
			case config.Wallet_tx_type_domain_withdraw: // 域名提现
				rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
				//for _, v := range rewardLogs {
				//	if v.Into == addr {
				//		returnVins := []mining.VinVO{}
				//		returnVins = append(returnVins, mining.VinVO{
				//			Addr: precompiled.RewardContract.B58String(),
				//		})
				//		returnVouts := []mining.VoutVO{}
				//		returnVouts = append(returnVouts, mining.VoutVO{
				//			Address: addr,
				//			Value:   v.Amount.Uint64(),
				//		})

				//		return config.Wallet_tx_type_domain_withdraw, returnVins, returnVouts, true
				//	}
				//}
				returnVins := []mining.VinVO{}
				returnVouts := []mining.VoutVO{}
				for _, v := range rewardLogs {
					returnVins = append(returnVins, mining.VinVO{
						Addr: v.From,
					})
					returnVouts = append(returnVouts, mining.VoutVO{
						Address: v.Into,
						Value:   v.Amount.Uint64(),
					})
				}

				return config.Wallet_tx_type_domain_withdraw, returnVins, returnVouts, true
			}
		}
	}

	// 剩下的可能是子域名合约的交易
	if tx.Class() == config.Wallet_tx_type_contract {
		payload := tx.GetPayload()
		//解析
		txClass, _ := precompiled.UnpackPayload(payload)
		switch txClass {
		case config.Wallet_tx_type_domain_register: // 子域名注册
			return config.Wallet_tx_type_subdomain_register, nil, nil, true
		case config.Wallet_tx_type_domain_renew: // 子域名续费
			return config.Wallet_tx_type_subdomain_renew, nil, nil, true
		case config.Wallet_tx_type_domain_withdraw: // 子域名提现
			rewardLogs := precompiled.GetEnsHistoryLog(blockheight, *tx.GetHash())
			//for _, v := range rewardLogs {
			//	if v.Into == addr {
			//		returnVins := []mining.VinVO{}
			//		returnVins = append(returnVins, mining.VinVO{
			//			Addr: precompiled.RewardContract.B58String(),
			//		})
			//		returnVouts := []mining.VoutVO{}
			//		returnVouts = append(returnVouts, mining.VoutVO{
			//			Address: addr,
			//			Value:   v.Amount.Uint64(),
			//		})

			//		return config.Wallet_tx_type_subdomain_withdraw, returnVins, returnVouts, true
			//	}
			//}
			returnVins := []mining.VinVO{}
			returnVouts := []mining.VoutVO{}
			for _, v := range rewardLogs {
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: addr,
					Value:   v.Amount.Uint64(),
				})
			}
			return config.Wallet_tx_type_subdomain_withdraw, returnVins, returnVouts, true
		}
	}

	return tx.Class(), nil, nil, false
}

// 返回交易类型，vin,vout
func DealTxInfoV2(tx mining.TxItr) (uint64, interface{}, interface{}) {
	//tx, _, _ := mining.FindTx(txid)
	vouts := tx.GetVout()
	out := *vouts
	vins := tx.GetVin()
	vin := *vins
	//如果是和奖励合约交互的交易
	if bytes.Equal(out[0].Address, precompiled.RewardContract) {

		//如果类型是1的话，为见证人奖励交易记录，从事件中获取奖励的金额,
		//if tx.Class() == config.Wallet_tx_type_mining {
		//	returnVins := []mining.VinVO{}
		//	returnVins = append(returnVins, mining.VinVO{
		//		Addr: precompiled.RewardContract.B58String(),
		//	})
		//	returnVouts := []mining.VoutVO{}
		//	rewardLogs := precompiled.GetRewardHistoryLog(tx.GetLockHeight(), *tx.GetHash())
		//	for _, v := range rewardLogs {
		//		if v.From == "" {
		//			returnVouts = append(returnVouts, mining.VoutVO{
		//				Address: v.Into,
		//				Value:   v.Reward.Uint64(),
		//			})
		//		}
		//	}
		//	return config.Wallet_tx_type_reward_W, returnVins, returnVouts
		//}
		//如果类型是20的话，为投票、质押、取消质押、取消投票、提现，从payload中解析即可
		if tx.Class() == config.Wallet_tx_type_contract {
			payload := tx.GetPayload()
			//解析
			txClass, params := precompiled.UnpackPayload(payload)
			switch txClass {
			case config.Wallet_tx_type_community_in:
				//社区节点质押,只需要变更类型
				return config.Wallet_tx_type_community_in, nil, nil
			case config.Wallet_tx_type_community_out:
				//社区节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   config.Mining_vote,
				})
				return config.Wallet_tx_type_community_out, returnVins, returnVouts
			case config.Wallet_tx_type_vote_in:
				//轻节点投票
				return config.Wallet_tx_type_vote_in, nil, nil
			case config.Wallet_tx_type_vote_out:
				//轻节点取消投票
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_vote_out, returnVins, returnVouts
			case config.Wallet_tx_type_light_in:
				//轻节点质押
				return config.Wallet_tx_type_light_in, nil, nil
			case config.Wallet_tx_type_light_out:
				//轻节点取消质押
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   config.Mining_light_min,
				})
				return config.Wallet_tx_type_light_out, returnVins, returnVouts
			case config.Wallet_tx_type_reward_C:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_reward_C, returnVins, returnVouts
			case config.Wallet_tx_type_reward_L:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_reward_L, returnVins, returnVouts
			case config.Wallet_tx_type_reward_W:
				//提现奖励
				returnVins := []mining.VinVO{}
				returnVins = append(returnVins, mining.VinVO{
					Addr: precompiled.RewardContract.B58String(),
				})
				returnVouts := []mining.VoutVO{}
				returnVouts = append(returnVouts, mining.VoutVO{
					Address: vin[0].PukToAddr.B58String(),
					Value:   params,
				})
				return config.Wallet_tx_type_reward_W, returnVins, returnVouts
			}
		}
	}
	return 0, nil, nil
}

func BlocksJsonByHeightRange(params *map[string]interface{}, startHeight, total uint64) engine.PostResult {
	pr := engine.NewPostResult()
	endHeight := startHeight + total

	type BlockHeadOut struct {
		FromBroadcast   bool                     `json:"-"`   //是否来自于广播的区块
		StaretBlockHash []byte                   `json:"sbh"` //创始区块hash
		BH              *mining.BlockHead        `json:"bh"`  //区块
		Txs             []map[string]interface{} `json:"txs"` //交易明细
		//Txs []mining.TxItr `json:"txs"` //交易明细
	}
	//待返回的区块
	bhvos := make([]BlockHeadOut, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {
		bhvo := BlockHeadOut{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		if bh == nil {
			break
		}
		bhvo.BH = bh
		bhvo.Txs = make([]map[string]interface{}, 0, len(bh.Tx))
		for _, one := range bh.Tx {
			txItrJson, code, txItr := mining.FindTxJsonVoV1(one)
			if txItr == nil {
				ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_found, "交易:"+hex.EncodeToString(one))
				pr.Code = ERR.Code
				pr.Msg = ERR.Msg
				return *pr
			}
			item, err := utils.ChangeMap(txItrJson) //JsonMethod(txItrJson)
			if err != nil {
				ERR := utils.NewErrorSysSelf(err)
				pr.Code = ERR.Code
				pr.Msg = ERR.Msg
				return *pr
			}
			txClass, vins, vouts := DealTxInfo(txItr, "", bh.Height)
			if txClass > 0 {
				item["type"] = txClass
				if vins != nil {
					item["vin"] = vins
				}
				if vouts != nil {
					item["vout"] = vouts
					item["vout_total"] = len(vouts.([]mining.VoutVO))
				}

				//if txClass == config.Wallet_tx_type_reward_W {
				//	item["hash"] = item["hash"].(string) + "-1"
				//}
			}
			//合约交易是20
			var (
				gasUsed  uint64
				gasLimit uint64
				gasPrice uint64
			)

			if txItr.Class() == config.Wallet_tx_type_contract {
				tx := txItr.(*mining.Tx_Contract)
				gasUsed = tx.GetGasLimit()
				gasLimit = config.EVM_GAS_MAX
				gasPrice = tx.GasPrice
			} else {
				gasUsed = txItr.GetGas()
				gasLimit = gasUsed
				gasPrice = config.DEFAULT_GAS_PRICE
			}
			item["free"] = txItr.GetGas()
			item["gas_used"] = gasUsed
			item["gas_limit"] = gasLimit
			item["gas_price"] = gasPrice
			item["upchaincode"] = code
			//如果是27，
			if txClass != config.Wallet_tx_type_reward_W {
				bhvo.Txs = append(bhvo.Txs, item)
			} else if txClass > 0 {
				if len(item["vout"].([]mining.VoutVO)) > 0 {
					bhvo.Txs = append(bhvo.Txs, item)
				}
			}
			//
			if txItr.Class() == config.Wallet_tx_type_mining {
				newiTem := make(map[string]interface{})
				for k, v := range item {
					newiTem[k] = v
				}
				returnVouts := []mining.VoutVO{}
				vouts := txItr.GetVout()
				outs := *vouts
				if bytes.Equal(outs[0].Address, precompiled.RewardContract) {
					returnVouts = append(returnVouts, mining.VoutVO{Address: precompiled.RewardContract.B58String(), Value: outs[0].Value})
					newiTem["type"] = config.Wallet_tx_type_mining
					newiTem["vout"] = returnVouts
					newiTem["vout_total"] = 1
					newiTem["vin"] = []mining.VinVO{}
					newiTem["vin_total"] = 0
					newiTem["hash"] = hex.EncodeToString(*txItr.GetHash()) + "-1"
					newTx := append([]map[string]interface{}{newiTem}, bhvo.Txs...)
					bhvo.Txs = newTx
				}

			}

		}

		bhvos = append(bhvos, bhvo)
	}
	pr.Code = utils.ERROR_CODE_success
	pr.Data["BlocksJson"] = bhvos
	return *pr
}

func ChainDepositAll(params *map[string]interface{}) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	wbg := mining.GetWitnessListSort()
	count := len(wbg.Witnesses)
	count1 := len(wbg.WitnessBackup)

	//commNum, count3 := precompiled.GetRoleTotal(config.Area.Keystore.GetCoinbase().Addr)
	commNum := int(0)
	balanceMgr := chain.GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()
	lightitems := balanceMgr.GetAllLights()
	wmc.Range(func(key, value any) bool {
		if commAddrItrs, ok := wmc.Load(key); ok {
			commAddrs := commAddrItrs.([]crypto.AddressCoin)
			commNum += len(commAddrs)
		}
		return true
	})

	depositAmount := uint64(count+count1) * config.Mining_deposit
	depositAmount = depositAmount + (uint64(commNum) * config.Mining_vote)
	depositAmount = depositAmount + (uint64(len(lightitems)) * config.Mining_light_min)

	//depositAmount := uint64(0)
	//balanceMgr := chain.GetBalance()
	//wmc := balanceMgr.GetWitnessMapCommunitys()
	//wmc.Range(func(key, value any) bool {
	//	depositAmount += config.Mining_deposit
	//	comms := value.([]crypto.AddressCoin)
	//	depositAmount += (uint64(len(comms)) * config.Mining_vote)
	//	return true
	//})
	//depositAmount += (uint64(len(lightitems)) * config.Mining_light_min)

	//res, err = model.Tojson(depositAmount)
	pr.Data["data"] = depositAmount
	return *pr
}

func ChainDepositNodeNum(params *map[string]interface{}) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	wbg := mining.GetWitnessListSort()
	count := len(wbg.Witnesses)
	count1 := len(wbg.WitnessBackup)

	//commNum, count3 := precompiled.GetRoleTotal(config.Area.Keystore.GetCoinbase().Addr)
	commNum := int(0)
	balanceMgr := chain.GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()
	lightitems := balanceMgr.GetAllLights()
	wmc.Range(func(key, value any) bool {
		if commAddrItrs, ok := wmc.Load(key); ok {
			commAddrs := commAddrItrs.([]crypto.AddressCoin)
			commNum += len(commAddrs)
		}
		return true
	})

	//data := make(map[string]interface{})
	//data["wit_num"] = count
	//data["back_wit_num"] = count1
	//data["community_num"] = commNum
	//data["light_num"] = len(lightitems)
	//return model.Tojson(data)
	pr.Data["wit_num"] = count
	pr.Data["back_wit_num"] = count1
	pr.Data["community_num"] = commNum
	pr.Data["light_num"] = len(lightitems)
	return *pr
}

func GetTxGas(params *map[string]interface{}) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	txAveGas := chain.Balance.GetTxAveGas()
	count := uint64(0)
	totalGas := uint64(0)
	for _, gas := range txAveGas.AllGas {
		if gas > 0 {
			count++
			totalGas += gas
		}
	}

	normal := uint64(0)
	if count != 0 {
		normal = totalGas / count
	}
	if normal < config.Wallet_tx_gas_min {
		normal = config.Wallet_tx_gas_min
	}

	low := normal / 125 * 100
	if low < config.Wallet_tx_gas_min {
		low = config.Wallet_tx_gas_min
	}

	data := make(map[string]interface{})
	data["low"] = low
	data["normal"] = normal
	data["fast"] = normal * 125 / 100
	//res, err = model.Tojson(data)
	pr.Data = data
	return *pr
}

func TestReceiveCoin(params *map[string]interface{}, address, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	if config.AddrPre != "TEST" {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_test_net, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	keyst, ERR := chain_plus.GetKeystore()
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	src := coin_address.AddressCoin(keyst.GetCoinAddrAll()[0].Addr.Bytes())

	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)
	var commentBs []byte
	if comment != "" {
		commentBs = []byte(comment)
	}

	//src := config.Area.Keystore.GetCoinbase().Addr

	gas := uint64(config.Wallet_tx_gas_min)
	frozenHeight := uint64(0)

	pwd := config.Wallet_keystore_default_pwd
	//comment := "发送测试币"

	runeLength := len([]rune(comment))
	if runeLength > 1024 {
		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_comment_size_too_long, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	//temp := new(big.Int).Mul(big.NewInt(int64(runeLength)), big.NewInt(int64(config.Wallet_tx_gas_min)))
	//temp = new(big.Int).Div(temp, big.NewInt(1024))
	//if gas < temp.Uint64() {
	//	res, err = model.Errcode(GasTooLittle, "gas")
	//	return
	//	ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_comment_size_too_long, "")
	//	pr.Code = ERR.Code
	//	pr.Msg = ERR.Msg
	//	return *pr
	//}
	amount := uint64(config.FAUCET_COIN)

	t1 := mining.GetFaucetTime(address)
	if !config.TimeNow().After(time.Unix(t1+24*60*60, 0)) {
		//res, err = model.Errcode(TestCoinLimit, "lock time has not expired. Please try again later")
		//return

		ERR := utils.NewErrorBus(chainBootConfig.ERROR_CODE_CHAIN_recv_interval_too_short, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	txpay, ERR := chain_plus.CreateTxPay(&src, &addr, amount, gas, frozenHeight, pwd, commentBs, "", 0)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	err := mining.SetFaucetTime(address)
	if err != nil {
		utils.Log.Error().Str("TestReceiveCoin", err.Error()).Send()
	}
	//result, err := utils.ChangeMap(txpay)
	//if err != nil {
	//	res, err = model.Errcode(model.Nomarl, err.Error())
	//	return
	//}
	//result["hash"] = hex.EncodeToString(*txpay.GetHash())

	//res, err = model.Tojson(result)

	pr.Data["hash"] = hex.EncodeToString(*txpay.GetHash())
	pr.Data["data"] = txpay
	return *pr
}
