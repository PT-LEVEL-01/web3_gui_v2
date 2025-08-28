package rpc

import (
	"github.com/shopspring/decimal"
	"sort"
	"web3_gui/chain/mining"
	"web3_gui/chain_boot/chain_plus"
	chainbootconfig "web3_gui/chain_boot/config"
	"web3_gui/chain_boot/model"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func CommunityList(params *map[string]interface{}, page, total uint64) engine.PostResult {
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

	balanceMgr := chain.GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()

	communityAddrs := []crypto.AddressCoin{}
	wmc.Range(func(key, value any) bool {
		communityAddrs = append(communityAddrs, value.([]crypto.AddressCoin)...)
		return true
	})

	comms := make([]model.CommunityList, 0)
	for _, commAddr := range communityAddrs {
		if comminfo := balanceMgr.GetDepositCommunity(&commAddr); comminfo != nil {
			ratio := balanceMgr.GetDepositRate(&comminfo.SelfAddr)
			comms = append(comms, model.CommunityList{
				Addr:        comminfo.SelfAddr.B58String(),
				WitnessAddr: comminfo.WitnessAddr.B58String(),
				Payload:     comminfo.Name,
				Score:       comminfo.Value,
				Vote:        balanceMgr.GetCommunityVote(&commAddr),
				RewardRatio: float64(ratio),
				DisRatio:    float64(ratio),
			})
		}
	}

	sort.Slice(comms, func(i, j int) bool {
		if comms[i].Vote > comms[j].Vote {
			return true
		}
		return false
	})

	count := len(comms)

	//data := map[string]interface{}{}
	//data["count"] = count
	start, end, ok := helppager(count, int(page), int(pageSizeInt))
	if ok {
		out := make([]model.CommunityList, end-start)
		copy(out, comms[start:end])
		//data["data"] = out
		pr.Data["data"] = out
	} else {
		//data["data"] = []interface{}{}
		pr.Data["data"] = []interface{}{}
	}

	//res, err = model.Tojson(data)
	pr.Data["count"] = count
	//pr.Data["data"] = data
	return *pr
}

func CommunityInfo(params *map[string]interface{}, address string, page, total uint64) engine.PostResult {
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

	balanceMgr := chain.GetBalance()
	commInfo := balanceMgr.GetDepositCommunity(&addr)
	if commInfo == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	cml := balanceMgr.GetCommunityMapLights()
	lightCount := uint64(0)
	lightAddrs := []crypto.AddressCoin{}
	if items, ok := cml.Load(utils.Bytes2string(addr)); ok {
		lightAddrs = items.([]crypto.AddressCoin)
		lightCount = uint64(len(lightAddrs))
	}
	rewardRatio := balanceMgr.GetDepositRate(&addr)
	reward := balanceMgr.GetAddrReward(addr)

	witnessName := ""
	witnessAddr := ""
	if witInfo := balanceMgr.GetDepositWitness(&commInfo.WitnessAddr); witInfo != nil {
		witnessName = witInfo.Name
		witnessAddr = witInfo.SelfAddr.B58String()
	}

	commFrozenReward, _ := balanceMgr.CalculateCommunityRewardAndLightReward(addr)
	contract := ""
	//if config.EVM_Reward_Enable {
	//	contract = precompiled.RewardContract.B58String()
	//}
	data := model.CommunityInfoV0{
		Deposit:      commInfo.Value,
		Vote:         balanceMgr.GetCommunityVote(&addr),
		LightCount:   lightCount,
		RewardRatio:  float64(rewardRatio),
		Reward:       reward.Uint64(),
		WitnessName:  witnessName,
		WitnessAddr:  witnessAddr,
		StartHeight:  commInfo.Height,
		FrozenReward: commFrozenReward.Uint64(),
		Contract:     contract,
		Name:         commInfo.Name,
		LightNode:    []model.LightNode{},
	}

	cvote := balanceMgr.GetCommunityVote(&commInfo.SelfAddr)
	lightNodes := []model.LightNode{}
	for _, lAddr := range lightAddrs {
		reward := balanceMgr.GetAddrReward(lAddr)
		lightvoteinfo := balanceMgr.GetDepositVote(&lAddr)

		lightName := ""
		lightDeposit := uint64(0)
		if lightinfo := balanceMgr.GetDepositLight(&lAddr); lightinfo != nil {
			lightName = lightinfo.Name
			lightDeposit = lightinfo.Value
		}
		//兼容老版本,老版本放大了100000倍
		ratio := decimal.NewFromInt(int64(lightvoteinfo.Value)).DivRound(decimal.NewFromInt(int64(cvote)), 3).Mul(decimal.NewFromInt(100000))
		ln := model.LightNode{
			Addr:            lAddr.B58String(),
			Reward:          reward.Uint64(),
			RewardRatio:     ratio.InexactFloat64(),
			VoteNum:         lightvoteinfo.Value,
			Name:            lightName,
			Deposit:         lightDeposit,
			LastVotedHeight: balanceMgr.GetLastVoteOp(lAddr),
		}
		lightNodes = append(lightNodes, ln)
	}
	sort.Sort(model.LightNodeSort(lightNodes))

	if start, end, ok := helppager(len(lightNodes), int(page), int(pageSizeInt)); ok {
		out := make([]model.LightNode, end-start)
		copy(out, lightNodes[start:end])
		data.LightNode = out
	} else {
		data.LightNode = []model.LightNode{}
	}

	//res, err = model.Tojson(data)
	pr.Data["data"] = data
	return *pr
}

func CommunityDepositIn(params *map[string]interface{}, address, witness string, rate, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()

	voterItr := (*params)["address1"]
	voter := crypto.AddressCoin(voterItr.(coin_address.AddressCoin))

	voteToItr := (*params)["witness1"]
	voteTo := crypto.AddressCoin(voteToItr.(coin_address.AddressCoin))

	ERR := chain_plus.CreateTxCommunityDepositIn(rate, voteTo, voter, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	return *pr
}

func CommunityDepositOut(params *map[string]interface{}, address string, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()

	voterItr := (*params)["address1"]
	voter := voterItr.(coin_address.AddressCoin)

	ERR := chain_plus.CreateTxCommunityDepositOut(voter, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	return *pr
}

func CommunityShowRewardPool(params *map[string]interface{}, address string) engine.PostResult {
	pr := engine.NewPostResult()
	communityItr := (*params)["address1"]
	addr := communityItr.(coin_address.AddressCoin)

	rewardTotal, ERR := chain_plus.GetCommunityRewardPool(addr)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["data"] = rewardTotal
	return *pr
}

func CommunityDistributeReward(params *map[string]interface{}, address string, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	tx, ERR := chain_plus.CreateTxVoteRewardByCommunity(addr, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["data"] = tx
	return *pr
}

func LightList(params *map[string]interface{}, page, total uint64) engine.PostResult {
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

	//vss := mining.GetLightListSortNew(pageInt, pageSizeInt)

	//lightTotal := precompiled.GetLightTotal(config.Area.Keystore.GetCoinbase().Addr)
	balanceMgr := chain.GetBalance()
	lightItems := balanceMgr.GetAllLights()

	lights := make([]mining.VoteScoreOut, 0)
	for _, lightitem := range lightItems {
		if lightinfo := balanceMgr.GetDepositVote(&lightitem.SelfAddr); lightinfo != nil {
			lights = append(lights, mining.VoteScoreOut{
				Witness: lightinfo.WitnessAddr.B58String(),
				Addr:    lightinfo.SelfAddr.B58String(),
				Payload: lightinfo.Name,
				Name:    lightinfo.Name,
				Score:   lightitem.Value,
				Vote:    lightinfo.Value,
			})
		}
	}

	sort.Slice(lights, func(i, j int) bool {
		if lights[i].Vote > lights[j].Vote {
			return true
		}
		return false
	})

	count := len(lights)
	//data := map[string]interface{}{}
	//data["count"] = count
	start, end, ok := helppager(count, int(page), int(pageSizeInt))
	if ok {
		out := make([]mining.VoteScoreOut, end-start)
		copy(out, lights[start:end])
		//data["data"] = out
		pr.Data["data"] = out
	} else {
		//data["data"] = []interface{}{}
		pr.Data["data"] = []interface{}{}
	}
	//res, err = model.Tojson(data)
	pr.Data["count"] = count
	return *pr
}

func LightInfo(params *map[string]interface{}, address string) engine.PostResult {
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

	balanceMgr := chain.GetBalance()
	lightinfo := balanceMgr.GetDepositLight(&addr)
	if lightinfo == nil {
		ERR := utils.NewErrorBus(chainbootconfig.ERROR_CODE_CHAIN_not_found, "")
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}

	data := make(map[string]interface{})
	data["community_addr"] = ""
	data["community_name"] = ""
	data["vote"] = uint64(0)
	data["last_voted_height"] = uint64(0)
	data["reward"] = uint64(0)
	data["voteAddress"] = ""
	data["frozen_reward"] = uint64(0)
	if lightvoteinfo := balanceMgr.GetDepositVote(&addr); lightvoteinfo != nil {
		if commInfo := balanceMgr.GetDepositCommunity(&lightvoteinfo.WitnessAddr); commInfo != nil {
			data["community_name"] = commInfo.Name
		}
		data["community_addr"] = lightvoteinfo.WitnessAddr.B58String()
		lastVoteOpHeight := balanceMgr.GetLastVoteOp(lightvoteinfo.SelfAddr)
		reward := balanceMgr.GetAddrReward(lightvoteinfo.SelfAddr)
		data["vote"] = lightvoteinfo.Value
		data["last_voted_height"] = lastVoteOpHeight
		data["reward"] = reward.Uint64()

		_, lightrewards := balanceMgr.CalculateCommunityRewardAndLightReward(lightvoteinfo.WitnessAddr)
		if v, ok := lightrewards[utils.Bytes2string(lightvoteinfo.SelfAddr)]; ok {
			data["frozen_reward"] = v.Uint64()
		}

		data["voteAddress"] = lightvoteinfo.WitnessAddr.B58String()
	}

	data["start_height"] = balanceMgr.GetLastVoteOp(lightinfo.SelfAddr)
	data["light_addr"] = addr
	data["name"] = lightinfo.Name
	data["contract"] = ""
	//if config.EVM_Reward_Enable {
	//	data["contract"] = precompiled.RewardContract.B58String()
	//}
	data["deposit"] = lightinfo.Value
	//data["start_height"] =
	//data["frozen_reward"] = precompiled.GetMyFrozenReward(addrCoin)
	//data["frozen_reward"] = precompiled.GetMyLightFrozenReward(addrCoin)

	//if cAddr != "" {
	//	cAddress := crypto.AddressFromB58String(cAddr)
	//	cRewardPool := precompiled.GetCommunityRewardPool(cAddress)
	//	if cRewardPool.Cmp(big.NewInt(0)) > 0 {
	//		cRate, err := precompiled.GetRewardRatio(cAddress)
	//		if err != nil {
	//			return model.Errcode(SystemError, "address")
	//		}
	//		lTotal := new(big.Int).Quo(new(big.Int).Mul(cRewardPool, big.NewInt(int64(cRate))), new(big.Int).SetInt64(100))
	//		v := new(big.Int).Mul(new(big.Int).SetUint64(mining.GetLongChain().GetBalance().GetDepositVote(&addrCoin).Value), new(big.Int).SetInt64(1e8))
	//		ratio := new(big.Int).Quo(v, cVote)
	//		lBigValue := new(big.Int).Quo(new(big.Int).Mul(lTotal, ratio), new(big.Int).SetInt64(1e8))
	//		frozenReward = lBigValue.Uint64()
	//	}
	//}

	//res, err = model.Tojson(data)
	//pr.Data["data"] = data
	pr.Data = data
	return *pr
}

func LightDepositIn(params *map[string]interface{}, address string, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()

	voterItr := (*params)["address1"]
	voter := crypto.AddressCoin(voterItr.(coin_address.AddressCoin))

	ERR := chain_plus.CreateTxLightDepositIn(voter, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	return *pr
}

func LightDepositOut(params *map[string]interface{}, address string, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()

	voterItr := (*params)["address1"]
	voter := voterItr.(coin_address.AddressCoin)

	ERR := chain_plus.CreateTxLightDepositOut(voter, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	return *pr
}

func LightVoteIn(params *map[string]interface{}, address, community string, amount, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()

	voterItr := (*params)["address1"]
	voter := voterItr.(coin_address.AddressCoin)

	voteToItr := (*params)["community1"]
	voteTo := voteToItr.(coin_address.AddressCoin)

	ERR := chain_plus.CreateTxLightVoteIn(voter, voteTo, amount, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	return *pr
}

func LightVoteOut(params *map[string]interface{}, address string, amount, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()

	voterItr := (*params)["address1"]
	voter := voterItr.(coin_address.AddressCoin)

	ERR := chain_plus.CreateTxLightVoteOut(voter, amount, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	return *pr
}

func LightDistributeReward(params *map[string]interface{}, address string, gas uint64, pwd, comment string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	tx, ERR := chain_plus.CreateTxVoteRewardByLight(addr, gas, pwd, comment)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["data"] = tx
	return *pr
}
