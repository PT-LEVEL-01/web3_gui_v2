package rpc

import (
	"math/big"
	"sort"
	chaninconfig "web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain_boot/chain_plus"
	chainbootconfig "web3_gui/chain_boot/config"
	"web3_gui/chain_boot/model"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func WitnessDepositIn(params *map[string]interface{}, nickname string, gas, rate uint64, pwd string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	txPay, ERR := chain_plus.CreateTxWitnessDepositIn(gas, pwd, nickname, uint16(rate))
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

func WitnessDepositOut(params *map[string]interface{}, gas uint64, pwd string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	txPay, ERR := chain_plus.CreateTxWitnessDepositOut(gas, pwd)
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

func WitnessFeatureList(params *map[string]interface{}) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	wbg := mining.GetWitnessListSort()
	balanceMgr := mining.GetLongChain().GetBalance()
	wvos := make([]model.WitnessInfoVO, 0)
	for _, one := range wbg.Witnesses {
		name := mining.FindWitnessName(*one.Addr)
		// engine.Log.Info("查询到的见证人名称:%s", name)
		//vote := precompiled.GetWitnessVote(*one.Addr)
		vote := balanceMgr.GetWitnessVote(one.Addr)
		wvo := model.WitnessInfoVO{
			Addr:            one.Addr.B58String(), //见证人地址
			Payload:         name,                 //
			Score:           one.Score,            //押金
			Vote:            vote,                 //      voteValue,            //投票票数
			CreateBlockTime: one.CreateBlockTime,  //预计出块时间
		}
		wvos = append(wvos, wvo)
	}
	pr.Data["list"] = wvos
	return *pr
}

func WitnessFeatureInfo(params *map[string]interface{}, address string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	srcAddrItr := (*params)["address1"]
	addr := crypto.AddressCoin(srcAddrItr.(coin_address.AddressCoin))
	witness := chain.WitnessChain.FindWitnessByAddr(addr)
	if witness == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
		//return model.Errcode(model.Nomarl, "not found witness")
	}
	score := witness.Score
	addBlockCount, addBlockReward := chain.Balance.GetAddBlockNum(address)

	balanceMgr := chain.GetBalance()
	voteNum := balanceMgr.GetWitnessVote(&addr)
	ratio := balanceMgr.GetDepositRate(&addr)
	wvo := model.WitnessInfoList{
		Addr:           address,                      //见证人地址
		Payload:        mining.FindWitnessName(addr), //名字
		Score:          score,                        //质押量
		Vote:           voteNum,                      // 总票数
		AddBlockCount:  addBlockCount,
		AddBlockReward: addBlockReward,
		Ratio:          float64(ratio),
	}
	//res, err = model.Tojson(wvo)
	pr.Data["data"] = wvo
	return *pr
}

func WitnessCandidateList(params *map[string]interface{}) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	wits := chain.WitnessBackup.GetAllWitness()
	wvos := []model.WitnessInfoList{}
	for _, v := range wits {
		addBlockCount, addBlockReward := chain.Balance.GetAddBlockNum(v.Addr.B58String())
		wvo := model.WitnessInfoList{
			Addr:           v.Addr.B58String(),                   //见证人地址
			Payload:        mining.FindWitnessName(*v.Addr),      //名字
			Score:          mining.GetDepositWitnessAddr(v.Addr), //质押量
			Vote:           chain.Balance.GetWitnessVote(v.Addr), // 总票数
			AddBlockCount:  addBlockCount,
			AddBlockReward: addBlockReward,
			Ratio:          float64(chain.Balance.GetDepositRate(v.Addr)),
		}
		wvos = append(wvos, wvo)
	}
	// 按投票数排序
	sort.Sort(model.WitnessListSort(wvos))
	pr.Data["list"] = wvos
	return *pr
}

func WitnessCandidateInfo(params *map[string]interface{}, address string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	srcAddrItr := (*params)["address1"]
	addr := crypto.AddressCoin(srcAddrItr.(coin_address.AddressCoin))

	witness := chain.WitnessChain.FindWitnessByAddr(addr)
	if witness == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	score := witness.Score
	addBlockCount, addBlockReward := chain.Balance.GetAddBlockNum(address)
	balanceMgr := chain.GetBalance()
	voteNum := balanceMgr.GetWitnessVote(&addr)
	ratio := balanceMgr.GetDepositRate(&addr)

	wvo := model.WitnessInfoList{
		Addr:           address,                      //见证人地址
		Payload:        mining.FindWitnessName(addr), //名字
		Score:          score,                        //质押量
		Vote:           voteNum,                      // 总票数
		AddBlockCount:  addBlockCount,
		AddBlockReward: addBlockReward,
		Ratio:          float64(ratio),
	}
	pr.Data["data"] = wvo
	return *pr
}

func WitnessZoneDeposit(params *map[string]interface{}, addressMore []string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	addrsItr := (*params)["addressMore1"]
	addrMore := addrsItr.([]coin_address.AddressCoin)

	addrs := []crypto.AddressCoin{}
	for _, addr := range addrMore {
		addrs = append(addrs, crypto.AddressCoin(addr))
	}

	if len(addrs) <= 0 {
		keys, ERR := chain_plus.GetKeystore()
		if ERR.CheckFail() {
			pr.Code = ERR.Code
			pr.Msg = ERR.Msg
			return *pr
		}
		allAddrs := keys.GetCoinAddrAll()
		for _, addr := range allAddrs {
			addrs = append(addrs, crypto.AddressCoin(addr.Addr.Bytes()))
		}
	}

	// depositAmount := toUint64(0)
	witCount := uint64(0)
	communityCount := uint64(0)
	lightCount := uint64(0)
	for _, addr := range addrs {
		role := mining.GetAddrState(addr)
		switch role {
		case 1: // 见证人
			witCount++
		case 2: // 社区
			communityCount++
		case 3: // 轻节点
			lightCount++
		}

		// depositAmount += precompiled.GetMyDeposit(addr.Addr)
	}

	// if config.ParseInitFlag() {
	// 	depositAmount += config.Mining_deposit
	// }

	info := map[string]interface{}{
		"deposit": witCount*chaninconfig.Mining_deposit + communityCount*chaninconfig.Mining_vote + lightCount*chaninconfig.Mining_light_min,
		// "total":   depositAmount,
	}
	//res, err = model.Tojson(info)

	pr.Data["data"] = info
	return *pr
}

func WitnessList(params *map[string]interface{}, page, total uint64) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	pageSizeInt := total
	if pageSizeInt > uint64(chainbootconfig.PageSizeLimit) {
		pageSizeInt = uint64(chainbootconfig.PageSizeLimit)
	}
	if pageSizeInt == 0 {
		pageSizeInt = chainbootconfig.PageTotalDefault
	}
	start := (page) * pageSizeInt

	witnessOrigin := chain.WitnessBackup.GetAllWitness()
	wvos, count, ERR := chain_plus.WitnessesListWithRangeV1(witnessOrigin, int(start), int(start+pageSizeInt))
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	pr.Data["count"] = count
	pr.Data["data"] = wvos
	return *pr
}

func WitnessInfo(params *map[string]interface{}, address string, page, total uint64) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()
	chain := mining.GetLongChain()
	if chain == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_ready, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	addrItr := (*params)["address1"]
	addr := crypto.AddressCoin(addrItr.(coin_address.AddressCoin))

	pageSizeInt := total
	if pageSizeInt > uint64(chainbootconfig.PageSizeLimit) {
		pageSizeInt = uint64(chainbootconfig.PageSizeLimit)
	}

	// 获得质押量
	depositAmount := mining.GetDepositWitnessAddr(&addr)
	if depositAmount <= 0 {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	addBlockCount, addBlockReward := mining.GetLongChain().Balance.GetAddBlockNum(address)
	frozenReward := big.NewInt(0)
	balanceMgr := mining.GetLongChain().GetBalance()
	ratio := balanceMgr.GetDepositRate(&addr)
	totalReward := balanceMgr.GetAddrReward(addr)
	wrp := balanceMgr.GetWitnessRewardPool()
	if val, ok := wrp.Load(utils.Bytes2string(addr)); ok {
		rewardpool := val.(*big.Int)
		frozenReward, _ = balanceMgr.CalculateWitnessRewardAndCommunityRewardPools(addr, rewardpool)
	}

	witnessName := mining.FindWitnessName(addr)
	data := model.WitnessInfoV0{
		Deposit:        depositAmount,
		AddBlockCount:  addBlockCount,
		AddBlockReward: addBlockReward,
		RewardRatio:    float64(ratio),
		TotalReward:    totalReward.Uint64(),
		FrozenReward:   frozenReward.Uint64(),
		Name:           witnessName,
		CommunityNode:  []model.CommunityNode{},
		DestroyNum:     0,
	}

	// 总票数
	voteTotal := balanceMgr.GetWitnessVote(&addr)

	wmc := balanceMgr.GetWitnessMapCommunitys()
	cml := balanceMgr.GetCommunityMapLights()
	if items, ok := wmc.Load(utils.Bytes2string(addr)); ok {
		communityAddrs := items.([]crypto.AddressCoin)
		data.CommunityCount = uint64(len(communityAddrs))

		cs := make([]model.CommunityNode, 0)
		// 获取社区节点详情
		for _, cAddr := range communityAddrs {
			reward := balanceMgr.GetAddrReward(cAddr)
			commInfo := balanceMgr.GetDepositCommunity(&cAddr)
			lightNum := uint64(0)
			if lights, ok := cml.Load(utils.Bytes2string(cAddr)); ok {
				lightNum = uint64(len(lights.([]crypto.AddressCoin)))
			}
			rewardRatio := balanceMgr.GetDepositRate(&addr)
			cs = append(cs, model.CommunityNode{
				Name:        commInfo.Name,
				Addr:        commInfo.SelfAddr.B58String(),
				Deposit:     commInfo.Value,
				Reward:      reward.Uint64(),
				LightNum:    lightNum,
				VoteNum:     balanceMgr.GetCommunityVote(&cAddr),
				RewardRatio: float64(rewardRatio),
			})
		}

		sort.Sort(model.CommunityNodeSort(cs))
		if start, end, ok := helppager(len(cs), int(page), int(pageSizeInt)); ok {
			out := make([]model.CommunityNode, end-start)
			copy(out, cs[start:end])
			data.CommunityNode = out
		} else {
			data.CommunityNode = []model.CommunityNode{}
		}
	}

	data.Vote = voteTotal

	//return model.Tojson(data)
	pr.Data["data"] = data
	return *pr
}

// 帮助分页
func helppager(totallen, page, pageSize int) (int, int, bool) {
	if page <= 0 || pageSize <= 0 {
		return 0, 0, false
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start > totallen {
		return 0, 0, false
	}

	if end > totallen {
		end = totallen
	}

	return start, end, true
}
