package rpc

import (
	"bytes"
	"math/big"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/mining"
	"web3_gui/chain_boot/chain_plus"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/v2/coin_address"
	"web3_gui/libp2parea/v2/engine"
	"web3_gui/utils"
)

func TokenPublish(params *map[string]interface{}, name, symbol, supply string, accuracy uint64, owner string,
	gas uint64, pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	var ownerAddr coin_address.AddressCoin
	if owner != "" {
		srcAddrItr := (*params)["owner1"]
		ownerAddr = srcAddrItr.(coin_address.AddressCoin)
	}

	supplyItr := (*params)["supply1"]
	supplyBig := supplyItr.(*big.Int)

	tx, ERR := chain_plus.CreateTxTokenPublish(&ownerAddr, name, symbol, supplyBig, accuracy, gas, pwd, []byte(comment))
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	ERR = chain_plus.CurrentLimitingAndMulticastTx(tx)
	if ERR.CheckFail() {
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["data"] = tx.GetVOJSON()
	return *pr
}

func TokenBalance(params *map[string]interface{}, tokenId16, address string) engine.PostResult {
	pr := engine.NewPostResult()
	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)
	caddr := crypto.AddressCoin(addr)
	ba, fba, baLockup := uint64(0), uint64(0), uint64(0)
	if tokenId16 == "" {
		ba, fba, baLockup = mining.GetBalanceForAddrSelf(caddr)
	} else {
		tokenIdItr := (*params)["tokenId161"]
		tokenId := tokenIdItr.([]byte)
		ba, fba, baLockup = mining.GetTokenNotSpendAndLockedBalance(tokenId, caddr)
	}

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

	if tokenId16 != "" {
		ty = 4
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

func TokenSendToAddress(params *map[string]interface{}, tokenId16, srcaddress, address string, amount, gas, frozen_height uint64,
	pwd, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	tokenIdItr := (*params)["tokenId161"]
	tokenId := tokenIdItr.([]byte)

	var src coin_address.AddressCoin
	if srcaddress != "" {
		srcAddrItr := (*params)["srcaddress1"]
		src = srcAddrItr.(coin_address.AddressCoin)
	}

	addrItr := (*params)["address1"]
	addr := addrItr.(coin_address.AddressCoin)

	txPay, ERR := chain_plus.CreateTxTokenSendToAddress(tokenId, src, addr, amount, gas, frozen_height, pwd, []byte(comment))
	if ERR.CheckFail() {
		utils.Log.Warn().Str("ERR", ERR.String()).Send()
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	ERR = chain_plus.CurrentLimitingAndMulticastTx(txPay)
	if ERR.CheckFail() {
		utils.Log.Warn().Str("ERR", ERR.String()).Send()
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	pr.Data["data"] = txPay.GetVOJSON()
	return *pr
}

func TokenSendToAddressMore(params *map[string]interface{}, tokenId16, srcAddress string, addressMore []string, amountMore,
	frozenHeightMore []uint64, gas uint64, pwd string, comment string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	tokenIdItr := (*params)["tokenId161"]
	tokenId := tokenIdItr.([]byte)

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

	txPay, ERR := chain_plus.CreateTxTokenSendToAddressMore(tokenId, src, addrs, amountMore, frozenHeightMore, gas, pwd, commentBs)
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

func TokenInfo(params *map[string]interface{}, tokenId16 string) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	tokenIdItr := (*params)["tokenId161"]
	tokenId := tokenIdItr.([]byte)

	tokeninfo, err := mining.FindTokenInfo(tokenId)
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	//res, err = model.Tojson(mining.ToTokenInfoV0(tokeninfo))
	pr.Data["data"] = mining.ToTokenInfoV0(tokeninfo)
	return *pr
}

func TokenList(params *map[string]interface{}) engine.PostResult {
	//utils.Log.Info().Uint64("支付订单金额", amount).Send()
	pr := engine.NewPostResult()

	list, err := mining.ListTokenInfos()
	if err != nil {
		ERR := utils.NewErrorSysSelf(err)
		pr.Code = ERR.Code
		pr.Msg = ERR.Msg
		return *pr
	}
	infos := []*mining.TokenInfoV0{}
	for _, l := range list {
		infos = append(infos, mining.ToTokenInfoV0(l))
	}
	pr.Data["data"] = infos
	return *pr
}
