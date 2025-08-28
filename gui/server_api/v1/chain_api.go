package server_api

import (
	"encoding/hex"
	"math/big"
	"sort"
	"web3_gui/chain/config"
	"web3_gui/chain/mining"
	"web3_gui/chain/mining/name"
	"web3_gui/chain/mining/tx_name_in"
	"web3_gui/chain/mining/tx_name_out"
	"web3_gui/chain/rpc"
	"web3_gui/keystore/v1"
	"web3_gui/keystore/v1/crypto"
	"web3_gui/libp2parea/v1/nodeStore"
	"web3_gui/libp2parea/v1/sdk/jsonrpc2/model"
	"web3_gui/utils"
)

/*
检查IM模块信息
*/
func (a *SdkApi) Chain_GetInfo() *rpc.Getinfo {
	// utils.Log.Info().Msgf("%s", config.NetAddr)

	value, valuef, valuelockup := mining.FindBalanceValue()

	//tbs := token.FindTokenBalanceForAll()
	//tbVOs := make([]token.TokenBalanceVO, 0)
	//for i, one := range tbs {
	//	tbs[i].TokenId = one.TokenId // hex.EncodeToString([]byte(one.TokenId))
	//	tbVO := token.TokenBalanceVO{
	//		TokenId:       hex.EncodeToString([]byte(one.TokenId)),
	//		Name:          one.Name,
	//		Symbol:        one.Symbol,
	//		Supply:        one.Supply.Text(10),
	//		Balance:       one.Balance.Text(10),
	//		BalanceFrozen: one.BalanceFrozen.Text(10),
	//		BalanceLockup: one.BalanceLockup.Text(10),
	//	}
	//	tbVOs = append(tbVOs, tbVO)
	//}

	currentBlock := uint64(0)
	startBlock := uint64(0)
	heightBlock := uint64(0)
	pulledStates := uint64(0)
	startBlockTime := uint64(0)

	chain := mining.GetLongChain()
	if chain != nil {
		currentBlock = chain.GetCurrentBlock()
		startBlock = chain.GetStartingBlock()
		heightBlock = mining.GetHighestBlock()
		pulledStates = chain.GetPulledStates()
		startBlockTime = chain.GetStartBlockTime()
	}

	info := rpc.Getinfo{
		// Netid:          []byte(config.AddrPre),   //
		TotalAmount:    config.Mining_coin_total, //
		Balance:        value,                    //
		BalanceFrozen:  valuef,                   //
		BalanceLockup:  valuelockup,              //
		Testnet:        true,                     //
		Blocks:         currentBlock,             //
		Group:          0,                        //
		StartingBlock:  startBlock,               //区块开始高度
		StartBlockTime: startBlockTime,           //
		HighestBlock:   heightBlock,              //所链接的节点的最高高度
		CurrentBlock:   currentBlock,             //已经同步到的区块高度
		PulledStates:   pulledStates,             //正在同步的区块高度
		//BlockTime:      config.Mining_block_time,       //出块时间
		BlockTime:      uint64(config.Mining_block_time.Nanoseconds()), //出块时间 pl time
		LightNode:      config.Mining_light_min,                        //轻节点押金数量
		CommunityNode:  config.Mining_vote,                             //社区节点押金数量
		WitnessNode:    config.Mining_deposit,                          //见证人押金数量
		NameDepositMin: config.Mining_name_deposit_min,                 //
		AddrPre:        config.AddrPre,                                 //
		//TokenBalance:   tbVOs,                                          //
	}
	return &info
}

/*
导出助记词
@pwd              string    密码
*/
func (a *SdkApi) Chain_ExportKey(pwd string) (result map[string]interface{}) {
	resultMap := make(map[string]interface{})
	str, err := config.Area.Keystore.ExportMnemonic(pwd)
	if err != nil {
		if err.Error() == config.ERROR_wallet_password_fail.Error() {
			// utils.Log.Info().Msgf("创建转账交易错误 222222222222")
			resultMap["code"] = model.FailPwd
			resultMap["error"] = "pwd"
			return resultMap
		}
		resultMap["error"] = err.Error()
		resultMap["code"] = model.Nomarl
		return resultMap
	}
	resultMap["code"] = model.Success
	resultMap["keys"] = str
	return resultMap
}

/*
导入助记词
@pwd              string    密码
*/
func (a *SdkApi) Chain_ImportKey(words, pwd string) (result map[string]interface{}) {
	resultMap := make(map[string]interface{})
	config.ParseConfig()
	k := keystore.NewKeystore(config.KeystoreFileAbsPath, config.AddrPre)
	// utils.Log.Info().Msgf("密钥路径:%s %s", config.KeystoreFileAbsPath, config.AddrPre)
	err := k.ImportMnemonic(words, pwd, pwd, pwd)
	if err != nil {
		if err.Error() == config.ERROR_wallet_password_fail.Error() {
			// utils.Log.Info().Msgf("创建转账交易错误 222222222222")
			resultMap["code"] = model.FailPwd
			resultMap["error"] = "pwd"
			return resultMap
		}
		resultMap["error"] = err.Error()
		resultMap["code"] = model.Nomarl
		return resultMap
	}
	resultMap["code"] = model.Success
	return resultMap
}

/*
获取收款地址列表
@token_id    string    币种ID，用于查询token余额
@page        uint      分页查询的页数
@count       uint      每页显示数量
*/
func (a *SdkApi) Chain_GetCoinAddress(token_id string, page, count int) *[]rpc.AccountVO {
	if page <= 0 {
		page = 1
	}
	if count <= 0 {
		count = 10000
	}
	pageSizeInt := count
	vos := make([]rpc.AccountVO, 0)
	list := config.Area.Keystore.GetAddr()
	total := len(config.Area.Keystore.GetAddr())
	start := (page - 1) * pageSizeInt
	end := start + pageSizeInt
	if start > total {
		return nil
	}
	if end > total {
		end = total
	}

	for i, val := range list[start:end] {
		var ba, fba, baLockup uint64
		if token_id == "" {
			ba, fba, baLockup = mining.GetBalanceForAddrSelf(val.Addr)
		} else {

		}

		// ba, _ := basMap[utils.Bytes2string(val.Addr)]
		// fba, _ := fbasMap[utils.Bytes2string(val.Addr)]
		// baLockup, _ := baLockupMap[utils.Bytes2string(val.Addr)]
		vo := rpc.AccountVO{
			Index:       i + start,
			AddrCoin:    val.GetAddrStr(),
			Type:        mining.GetAddrState(val.Addr),
			Value:       ba,       //可用余额
			ValueFrozen: fba,      //冻结余额
			ValueLockup: baLockup, //
		}
		vos = append(vos, vo)
	}
	return &vos
}

/*
创建新地址
@pwd    string    创建新地址
*/
func (a *SdkApi) Chain_NewCoinAddress(password string) (code int, newAddr *model.GetNewAddress) {
	addr, err := config.Area.Keystore.GetNewAddr(password, password)
	if err != nil {
		if err.Error() == keystore.ERROR_wallet_password_fail.Error() {
			return model.FailPwd, nil
		}
		return model.Nomarl, nil
	}
	getnewadress := model.GetNewAddress{Address: addr.B58String()}
	return model.Success, &getnewadress
}

/*
获取收款地址列表
@srcAddr          string    扣款地址
@dstAddr          string    收款地址
@amount           uint64    转账金额
@gas              uint64    手续费
@frozenHeight     uint64    有效高度
@pwd              string    密码
@comment          string    备注
*/
func (a *SdkApi) Chain_pay(srcAddr, dstAddr string, amount, gas, frozenHeight uint64, pwd string, comment string) (code int, result map[string]interface{}) {
	var src crypto.AddressCoin
	if srcAddr != "" {
		src = crypto.AddressFromB58String(srcAddr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, src) {
			// res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")

			return rpc.ContentIncorrectFormat, nil
		}
		_, ok := config.Area.Keystore.FindAddress(src)
		if !ok {
			// res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
			// return
			return rpc.ContentIncorrectFormat, nil
		}
	}

	dst := crypto.AddressFromB58String(dstAddr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		// res, err = model.Errcode(ContentIncorrectFormat, "address")
		return rpc.ContentIncorrectFormat, nil
	}

	total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, amount+gas)
	if total < amount+gas {
		//资金不够
		// res, err = model.Errcode(BalanceNotEnough)
		return rpc.BalanceNotEnough, nil
	}

	txpay, err := mining.SendToAddress(&src, &dst, amount, gas, frozenHeight, pwd, comment, "", 0)
	// utils.Log.Info().Msgf("转账耗时 %s", time.Now().Sub(startTime))
	if err != nil {
		// utils.Log.Info().Msgf("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// utils.Log.Info().Msgf("创建转账交易错误 222222222222")
			// res, err = model.Errcode(model.FailPwd)
			return model.FailPwd, nil
		}
		// utils.Log.Info().Msgf("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			// res, err = model.Errcode(AmountIsZero, "amount")
			return rpc.AmountIsZero, nil
		}
		// res, err = model.Errcode(model.Nomarl, err.Error())
		return model.Nomarl, nil
	}

	result, err = utils.ChangeMap(txpay)
	if err != nil {
		// res, err = model.Errcode(model.Nomarl, err.Error())
		return model.Nomarl, nil
		// return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())
	return model.Success, result
}

/*
查询自己是否是见证人
*/
func (a *SdkApi) Chain_GetWitnessInfo() *rpc.WitnessInfo {
	winfo := rpc.WitnessInfo{}

	chain := mining.GetLongChain()
	if chain == nil {
		return nil
	}
	var witnessAddr crypto.AddressCoin
	winfo.IsCandidate, winfo.IsBackup, winfo.IsKickOut, witnessAddr, winfo.Value = mining.GetWitnessStatus()
	winfo.Addr = witnessAddr.B58String()

	addr := config.Area.Keystore.GetCoinbase()
	winfo.Payload = mining.FindWitnessName(addr.Addr)
	return &winfo
}

/*
缴纳押金，成为见证人
@amount    uint64    押金
@gas       uint64    手续费
@pwd       uint64    密码
@payload   string    见证人名称
@rate      uint16    分配比例
*/
func (a *SdkApi) Chain_WitnessDepositIn(amount, gas uint64, pwd, payload string, rate uint16) (code int, errorStr string) {
	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= config.Wallet_vote_start_height {
		return rpc.VoteNotOpen, ""
	}
	//查询余额是否足够
	value, _, _ := mining.FindBalanceValue()
	if amount > value {
		return rpc.BalanceNotEnough, ""
	}
	err := mining.DepositIn(amount, gas, pwd, payload, rate)
	if err != nil {
		utils.Log.Info().Msgf("%s", err.Error())
		if err.Error() == config.ERROR_password_fail.Error() {
			return model.FailPwd, ""
		} else if err.Error() == config.ERROR_witness_deposit_exist.Error() {
			return rpc.WitnessDepositExist, ""
		} else if err.Error() == config.ERROR_witness_deposit_less.Error() {
			return rpc.WitnessDepositLess, ""
		}
		return model.Nomarl, err.Error()
	}
	config.SubmitDepositin = true
	return model.Success, ""
}

/*
获取域名列表
*/
func (a *SdkApi) Chain_GetNames() (result map[string]interface{}) {
	nameinfoVOs := make([]rpc.NameinfoVO, 0)
	names := name.GetNameList()
	for _, one := range names {
		nets := make([]string, 0)
		for _, two := range one.NetIds {
			nets = append(nets, two.B58String())
		}
		addrs := make([]string, 0)
		for _, two := range one.AddrCoins {
			addrs = append(addrs, two.B58String())
		}
		voOne := rpc.NameinfoVO{
			Name:           one.Name,              //域名
			Owner:          one.Owner.B58String(), //
			NetIds:         nets,                  //节点地址
			AddrCoins:      addrs,                 //钱包收款地址
			Height:         one.Height,            //注册区块高度，通过现有高度计算出有效时间
			NameOfValidity: one.NameOfValidity,    //有效块数量
			Deposit:        one.Deposit,
		}
		nameinfoVOs = append(nameinfoVOs, voOne)
	}

	var list struct {
		List []rpc.NameinfoVO
	}
	list.List = nameinfoVOs

	resultMap := make(map[string]interface{})
	result, e := utils.ChangeMap(list)
	if e != nil {
		resultMap["code"] = model.Nomarl
		resultMap["error"] = e.Error()
		return resultMap
	}
	result["code"] = model.Success
	return result
}

/*
域名注册，修改，续期
@srcAddr          string    扣款地址
@dstAddr          string    收款地址
@amount           uint64    转账金额
@gas              uint64    手续费
@frozenHeight     uint64    有效高度
@pwd              string    密码
@comment          string    备注
@name             string    域名
@netIds           []string  网络地址
@addrCoins        []string  收款地址
*/
func (a *SdkApi) Chain_NameIn(srcAddr, dstAddr string, amount, gas, frozenHeight uint64, pwd, comment string,
	name string, netIds []string, addrCoins []string) (result map[string]interface{}) {

	resultMap := make(map[string]interface{})
	var src crypto.AddressCoin
	if srcAddr != "" {
		src = crypto.AddressFromB58String(srcAddr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, src) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "srcAddr"
			return resultMap
		}
		_, ok := config.Area.Keystore.FindAddress(src)
		if !ok {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "srcAddr"
			return resultMap
		}
	}
	var dst crypto.AddressCoin
	if dstAddr != "" {
		dst = crypto.AddressFromB58String(dstAddr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, dst) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "dstAddr"
			return resultMap
		}
	}

	//域名解析的节点地址参数
	ids := make([]nodeStore.AddressNet, 0)
	for _, netidOne := range netIds {
		if netidOne == "" {
			continue
		}
		idOne := nodeStore.AddressFromB58String(netidOne)
		ids = append(ids, idOne)
	}

	//收款地址参数
	coins := make([]crypto.AddressCoin, 0)
	for _, addrcoinOne := range addrCoins {
		if addrcoinOne == "" {
			continue
		}
		idOne := crypto.AddressFromB58String(addrcoinOne)
		if !crypto.ValidAddr(config.AddrPre, idOne) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "addrCoins"
			return resultMap
		}
		coins = append(coins, idOne)
	}
	txpay, err := tx_name_in.NameIn(nil, &dst, amount, gas, 0, pwd, comment, mining.NameInActionReg, name, ids, coins)
	if err == nil {
		result, e := utils.ChangeMap(txpay.GetVOJSON())
		if e != nil {
			resultMap["code"] = model.Nomarl
			resultMap["error"] = err.Error()
			return resultMap
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())
		result["code"] = model.Success
		return result
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		// return model.FailPwd, nil
		resultMap["code"] = model.FailPwd
		resultMap["error"] = ""
		return resultMap
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		// return rpc.BalanceNotEnough, nil
		resultMap["code"] = rpc.BalanceNotEnough
		resultMap["error"] = ""
		return resultMap
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		// return rpc.NameExist, nil
		resultMap["code"] = rpc.NameExist
		resultMap["error"] = ""
		return resultMap
	}
	resultMap["error"] = err.Error()
	resultMap["code"] = model.Nomarl
	return resultMap
}

/*
域名注销
@srcAddr          string    扣款地址
@dstAddr          string    收款地址
@amount           uint64    转账金额
@gas              uint64    手续费
@frozenHeight     uint64    有效高度
@pwd              string    密码
@comment          string    备注
@name             string    域名
@netIds           []string  网络地址
@addrCoins        []string  收款地址
*/
func (a *SdkApi) Chain_NameOut(srcAddr, dstAddr string, amount, gas, frozenHeight uint64, pwd, comment string,
	name string) (code int, result map[string]interface{}) {
	var src crypto.AddressCoin
	if srcAddr != "" {
		src = crypto.AddressFromB58String(srcAddr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, src) {
			return rpc.ContentIncorrectFormat, nil
		}
		_, ok := config.Area.Keystore.FindAddress(src)
		if !ok {
			return rpc.ContentIncorrectFormat, nil
		}
	}
	dst := crypto.AddressFromB58String(dstAddr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		return rpc.ContentIncorrectFormat, nil
	}

	resultMap := make(map[string]interface{})
	txpay, err := tx_name_out.NameOut(nil, &dst, 0, gas, 0, pwd, comment, name)
	if err == nil {
		result, e := utils.ChangeMap(txpay)
		if e != nil {
			resultMap["error"] = err.Error()
			return model.Nomarl, resultMap
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())
		return model.Success, result
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		return model.FailPwd, nil
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		return rpc.BalanceNotEnough, nil
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		return rpc.NameExist, nil
	}
	resultMap["error"] = err.Error()
	return model.Nomarl, resultMap
}

/*
发布一个token
@srcAddr          string    扣款地址
@ownerAddr        string    拥有者
@gas              uint64    手续费
@pwd              string    密码
@comment          string    备注
@name             string    token名称
@symbol           string    token单位
@supply           string    token发行总量
*/
func (a *SdkApi) Chain_TokenPublish(srcAddr, ownerAddr string, gas uint64, pwd, comment string,
	name, symbol string, supply string) (result map[string]interface{}) {
	resultMap := make(map[string]interface{})

	var src crypto.AddressCoin
	if srcAddr != "" {
		src = crypto.AddressFromB58String(srcAddr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, src) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "srcAddr"
			return resultMap
		}
		_, ok := config.Area.Keystore.FindAddress(src)
		if !ok {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "srcAddr"
			return resultMap
		}
	}
	var owner crypto.AddressCoin
	if ownerAddr != "" {
		owner = crypto.AddressFromB58String(ownerAddr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, owner) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "ownerAddr"
			return resultMap
		}
	}

	supplyBig, ok := new(big.Int).SetString(supply, 10)
	if !ok {
		resultMap["code"] = rpc.ContentIncorrectFormat
		resultMap["error"] = "supply"
		return resultMap
	}

	txItr, err := mining.PublishToken(&src, nil, 0, gas, 0, pwd, comment, name, symbol, supplyBig, 8, owner)
	if err == nil {
		result, e := utils.ChangeMap(txItr.GetVOJSON())
		if e != nil {
			resultMap["code"] = model.Nomarl
			resultMap["error"] = err.Error()
			return resultMap
		}
		result["hash"] = hex.EncodeToString(*txItr.GetHash())
		result["code"] = model.Success
		return result
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		// return model.FailPwd, nil
		resultMap["code"] = model.FailPwd
		resultMap["error"] = ""
		return resultMap
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		// return rpc.BalanceNotEnough, nil
		resultMap["code"] = rpc.BalanceNotEnough
		resultMap["error"] = ""
		return resultMap
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		// return rpc.NameExist, nil
		resultMap["code"] = rpc.NameExist
		resultMap["error"] = ""
		return resultMap
	}
	resultMap["error"] = err.Error()
	resultMap["code"] = model.Nomarl
	return resultMap
}

/*
@srcAddr          string    扣款地址
@dstAddr          string    收款地址
@amount           uint64    转账金额
@gas              uint64    手续费
@frozenHeight     uint64    有效高度
@pwd              string    密码
@comment          string    备注
@tokenID          string    tokenID
*/
func (a *SdkApi) Chain_TokenPay(srcAddr, dstAddr string, amount, gas, frozenHeight uint64, pwd, comment, tokenID string) (result map[string]interface{}) {
	resultMap := make(map[string]interface{})

	var src crypto.AddressCoin
	if srcAddr != "" {
		src = crypto.AddressFromB58String(srcAddr)
		//判断地址前缀是否正确
		if !crypto.ValidAddr(config.AddrPre, src) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "srcAddr"
			return resultMap
		}
		_, ok := config.Area.Keystore.FindAddress(src)
		if !ok {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "srcAddr"
			return resultMap
		}
	}
	dst := crypto.AddressFromB58String(dstAddr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		resultMap["code"] = rpc.ContentIncorrectFormat
		resultMap["error"] = "dstAddr"
		return resultMap
	}

	txItr, err := mining.TokenPay(&src, &dst, amount, gas, frozenHeight, pwd, comment, tokenID)
	if err == nil {
		result, e := utils.ChangeMap(txItr.GetVOJSON())
		if e != nil {
			resultMap["code"] = model.Nomarl
			resultMap["error"] = err.Error()
			return resultMap
		}
		result["hash"] = hex.EncodeToString(*txItr.GetHash())
		result["code"] = model.Success
		return result
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		// return model.FailPwd, nil
		resultMap["code"] = model.FailPwd
		resultMap["error"] = ""
		return resultMap
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		// return rpc.BalanceNotEnough, nil
		resultMap["code"] = rpc.BalanceNotEnough
		resultMap["error"] = ""
		return resultMap
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		// return rpc.NameExist, nil
		resultMap["code"] = rpc.NameExist
		resultMap["error"] = ""
		return resultMap
	}
	resultMap["error"] = err.Error()
	resultMap["code"] = model.Nomarl
	return resultMap
}

/*
获取候选见证人列表
*/
func (a *SdkApi) Chain_GetCandidateList() (resultMap map[string]interface{}) {
	resultMap = make(map[string]interface{})

	chain := mining.GetLongChain()
	if chain == nil {
		// return nil, 0, errors.New("get chain failed")
		resultMap["error"] = "get chain failed"
		resultMap["code"] = model.Nomarl
		return resultMap
	}

	wits := chain.WitnessBackup.GetAllWitness()

	wvos := []rpc.WitnessList{}
	// total := len(wits)

	for _, v := range wits {
		addBlockCount, addBlockReward := rpc.GetAddressAddBlockReward(v.Addr.B58String())
		wvo := rpc.WitnessList{
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
	sort.Sort(rpc.WitnessListSort(wvos))

	// currWitness := mining.GetLongChain().WitnessChain.GetCurrGroupLastWitness()
	// wbg := currWitness.WitnessBigGroup

	// allWiness := append(wbg.Witnesses, wbg.WitnessBackup...)
	// addrs := make([]common.Address, len(allWiness))
	// for k, one := range allWiness {
	// 	addrs[k] = common.Address(evmutils.AddressCoinToAddress(*one.Addr))
	// }

	// from := config.Area.Keystore.GetCoinbase().Addr
	// _, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(from, addrs)
	// if err != nil {
	// 	resultMap["error"] = "address"
	// 	resultMap["code"] = model.Nomarl
	// 	return resultMap
	// }
	// wvos := make([]rpc.WitnessVO, 0)
	// for k, one := range allWiness {
	// 	wvo := rpc.WitnessVO{
	// 		Addr:            one.Addr.B58String(),              //见证人地址
	// 		Payload:         mining.FindWitnessName(*one.Addr), //
	// 		Score:           one.Score,                         //押金
	// 		Vote:            votes[k].Uint64(),                 //      voteValue,            //投票票数
	// 		CreateBlockTime: one.CreateBlockTime,               //预计出块时间
	// 	}
	// 	wvos = append(wvos, wvo)
	// }
	resultMap["list"] = wvos
	resultMap["code"] = model.Success
	return resultMap
}

/*
获取社区节点列表
*/
func (a *SdkApi) Chain_GetCommunityList() (resultMap map[string]interface{}) {
	resultMap = make(map[string]interface{})

	balanceMgr := mining.GetLongChain().GetBalance()
	wmc := balanceMgr.GetWitnessMapCommunitys()

	communityAddrs := []crypto.AddressCoin{}
	wmc.Range(func(key, value any) bool {
		communityAddrs = append(communityAddrs, value.([]crypto.AddressCoin)...)
		return true
	})

	comms := make([]rpc.CommunityList, 0)
	for _, commAddr := range communityAddrs {
		if comminfo := balanceMgr.GetDepositCommunity(&commAddr); comminfo != nil {
			ratio := balanceMgr.GetDepositRate(&comminfo.SelfAddr)
			comms = append(comms, rpc.CommunityList{
				Addr:        comminfo.SelfAddr.B58String(),
				WitnessAddr: comminfo.WitnessAddr.B58String(),
				Payload:     comminfo.Name,
				Score:       config.Mining_vote,
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

	// count := len(comms)

	// vss := mining.GetCommunityListSort()
	resultMap["list"] = comms
	resultMap["code"] = model.Success
	return resultMap
}

/*
获取自己的投票列表
获得自己给哪些见证人/社区投过票的列表
@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
*/
func (a *SdkApi) Chain_GetVoteList(voteType int) (resultMap map[string]interface{}) {
	resultMap = make(map[string]interface{})

	var items []*mining.DepositInfo
	switch voteType {
	case mining.VOTE_TYPE_community:
		items = mining.GetDepositCommunityList()
	case mining.VOTE_TYPE_light:
		items = mining.GetDepositLightList()
	case mining.VOTE_TYPE_vote:
		items = mining.GetDepositVoteList()
	}
	infos := make([]rpc.VoteInfoVO, 0, len(items))
	for _, item := range items {
		var name string
		if voteType == mining.VOTE_TYPE_community {
			name = mining.FindWitnessName(item.WitnessAddr)
		} else {
			name = item.Name
		}

		viVO := rpc.VoteInfoVO{
			// Txid:        hex.EncodeToString(ti.Txid), //
			WitnessAddr: item.WitnessAddr.B58String(), //见证人地址
			Value:       item.Value,                   //投票数量
			// Height:      item.Height,           //区块高度
			AddrSelf: item.SelfAddr.B58String(), //自己投票的地址
			Payload:  name,                      //
		}
		infos = append(infos, viVO)
	}
	vinfos := rpc.NewVinfos(infos)
	sort.Stable(&vinfos)
	resultMap["list"] = vinfos.GetInfos()
	resultMap["code"] = model.Success
	return resultMap
}

/*
成为社区节点押金，成为轻节点押金，投票
@voteType    int          投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
@address     string       记票地址
@witness     string       给谁投票
@rate        int          见证人奖励比例 1-100
@amount      uint64       投票数量
@gas         uint64       手续费
@gasPrice    uint64       手续费
@pwd         string       密码
@payload     string       备注
*/
func (a *SdkApi) Chain_VoteIn(voteType uint16, address, witness string, rate uint16, amount, gas, gasPrice uint64, pwd, payload string) (resultMap map[string]interface{}) {
	resultMap = make(map[string]interface{})

	// var payAddress crypto.AddressCoin
	// if payAddr != "" {
	// 	payAddress = crypto.AddressFromB58String(payAddr)
	// 	if !crypto.ValidAddr(config.AddrPre, payAddress) {
	// 		resultMap["code"] = rpc.ContentIncorrectFormat
	// 		resultMap["error"] = "payAddr"
	// 		return resultMap
	// 	}
	// } else {
	// 	payAddress = config.Area.Keystore.GetCoinbase().Addr
	// }
	// total, _ := mining.GetLongChain().Balance.BuildPayVinNew(&payAddress, amount+gas)
	// if total < amount+gas {
	// 	//资金不够
	// 	resultMap["code"] = rpc.BalanceNotEnough
	// 	resultMap["error"] = "amount"
	// 	return resultMap
	// }

	var voter, voteTo crypto.AddressCoin
	if address != "" {
		voter = crypto.AddressFromB58String(address)
		if !crypto.ValidAddr(config.AddrPre, voter) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "address"
			return resultMap
		}
	}

	if witness != "" {
		voteTo = crypto.AddressFromB58String(witness)
		if !crypto.ValidAddr(config.AddrPre, voteTo) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "witness"
			return resultMap
		}
	}

	switch voteType {
	case mining.VOTE_TYPE_community:
		if voter == nil {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "address"
			return resultMap
		}
		if voteTo == nil {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "witness"
			return resultMap
		}
	case mining.VOTE_TYPE_vote:
		if voter == nil {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "address"
			return resultMap
		}
		if voteTo == nil {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "witness"
			return resultMap
		}
	case mining.VOTE_TYPE_light:
		voteTo = nil
		if voter == nil {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "address"
			return resultMap
		}
	default:
		resultMap["code"] = rpc.SystemError
		resultMap["error"] = "votetype"
		return resultMap
	}

	if rate > 100 {
		resultMap["code"] = rpc.DistributeRatioTooBig
		resultMap["error"] = "分配比例不能大于100"
		return resultMap
	}

	if gas < config.Wallet_tx_gas_min {
		resultMap["code"] = rpc.GasTooLittle
		resultMap["error"] = "gas"
		return resultMap
	}

	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= config.Wallet_vote_start_height {
		resultMap["error"] = "vote not open"
		resultMap["code"] = rpc.VoteNotOpen
		return resultMap
	}

	//查询余额是否足够
	value, _, _ := mining.FindBalanceValue()
	if amount > value {
		resultMap["error"] = "BalanceNotEnough"
		resultMap["code"] = rpc.BalanceNotEnough
		return resultMap
	}

	err := mining.VoteIn(voteType, rate, voteTo, voter, amount, gas, pwd, payload, gasPrice)
	if err == nil {
		// result, e := utils.ChangeMap(txItr.GetVOJSON())
		// if e != nil {
		// 	resultMap["code"] = model.Nomarl
		// 	resultMap["error"] = err.Error()
		// 	return resultMap
		// }
		// result["hash"] = hex.EncodeToString(*txItr.GetHash())
		resultMap["code"] = model.Success
		return resultMap
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		resultMap["code"] = model.FailPwd
		resultMap["error"] = "pwd"
		return resultMap
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		resultMap["code"] = rpc.BalanceNotEnough
		resultMap["error"] = "amount"
		return resultMap
	}
	resultMap["error"] = err.Error()
	resultMap["code"] = model.Nomarl
	return resultMap
}

/*
取消投票，取消押金
@voteType    int          投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
@address     string       记票地址
@amount      uint64       投票数量
@gas         uint64       手续费
@gasPrice    uint64       手续费
@pwd         string       密码
@payload     string       备注
*/
func (a *SdkApi) Chain_VoteOut(voteType uint16, address string, amount, gas, gasPrice uint64, pwd, payload string) (resultMap map[string]interface{}) {
	resultMap = make(map[string]interface{})

	var voter crypto.AddressCoin
	if address != "" {
		voter = crypto.AddressFromB58String(address)
		if !crypto.ValidAddr(config.AddrPre, voter) {
			resultMap["code"] = rpc.ContentIncorrectFormat
			resultMap["error"] = "address"
			return resultMap
		}
	}
	if voter == nil {
		resultMap["code"] = rpc.ContentIncorrectFormat
		resultMap["error"] = "address"
		return resultMap
	}

	switch voteType {
	case mining.VOTE_TYPE_community:
	case mining.VOTE_TYPE_vote:
	case mining.VOTE_TYPE_light:
	default:
		resultMap["code"] = rpc.SystemError
		resultMap["error"] = "votetype"
		return resultMap
	}

	if gas < config.Wallet_tx_gas_min {
		resultMap["code"] = rpc.GasTooLittle
		resultMap["error"] = "gas"
		return resultMap
	}

	err := mining.VoteOut(voteType, voter, amount, gas, pwd, payload, gasPrice)
	// err := mining.VoteIn(voteType, rate, voteTo, voter, amount, gas, pwd, payload, gasPrice)
	if err == nil {
		// result, e := utils.ChangeMap(txItr.GetVOJSON())
		// if e != nil {
		// 	resultMap["code"] = model.Nomarl
		// 	resultMap["error"] = err.Error()
		// 	return resultMap
		// }
		// result["hash"] = hex.EncodeToString(*txItr.GetHash())
		resultMap["code"] = model.Success
		return resultMap
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		resultMap["code"] = model.FailPwd
		resultMap["error"] = "pwd"
		return resultMap
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		resultMap["code"] = rpc.BalanceNotEnough
		resultMap["error"] = "amount"
		return resultMap
	}
	//投票已经存在
	if err.Error() == config.ERROR_vote_exist.Error() {
		resultMap["code"] = rpc.VoteExist
		resultMap["error"] = ""
		return resultMap
	}
	resultMap["error"] = err.Error()
	resultMap["code"] = model.Nomarl
	return resultMap
}

/*
社区向轻节点分账
@srcAddress  string       是社区节点/轻节点地址
@gas         uint64       手续费
@gasPrice    uint64       手续费
@pwd         string       密码
@payload     string       备注
*/
func (a *SdkApi) Chain_CommunityDistribute(srcAddress string, gas, gasPrice uint64, pwd, payload string) (resultMap map[string]interface{}) {
	resultMap = make(map[string]interface{})

	if srcAddress == "" {
		resultMap["code"] = rpc.ContentIncorrectFormat
		resultMap["error"] = "srcAddress"
		return resultMap
	}
	voter := crypto.AddressFromB58String(srcAddress)
	if !crypto.ValidAddr(config.AddrPre, voter) {
		resultMap["code"] = rpc.ContentIncorrectFormat
		resultMap["error"] = "srcAddress"
		return resultMap
	}

	role := mining.GetAddrState(voter)
	if !(role == 2 || role == 3) {
		resultMap["code"] = model.Nomarl
		resultMap["error"] = "not community or light address"
		return resultMap
	}

	if gas < config.Wallet_tx_gas_min {
		resultMap["code"] = rpc.GasTooLittle
		resultMap["error"] = "gas"
		return resultMap
	}

	_, err := mining.CreateTxVoteRewardNew(&voter, gas, pwd)
	// if err != nil {
	// 	if err.Error() == config.ERROR_password_fail.Error() {
	// 		res, err = model.Errcode(model.FailPwd)
	// 		return
	// 	}
	// 	res, err = model.Errcode(model.Nomarl, err.Error())
	// 	return
	// }

	// err := mining.VoteOut(voteType, voter, amount, gas, pwd, payload, gasPrice)
	// err := mining.VoteIn(voteType, rate, voteTo, voter, amount, gas, pwd, payload, gasPrice)
	if err == nil {
		// result, e := utils.ChangeMap(txItr.GetVOJSON())
		// if e != nil {
		// 	resultMap["code"] = model.Nomarl
		// 	resultMap["error"] = err.Error()
		// 	return resultMap
		// }
		// result["hash"] = hex.EncodeToString(*txItr.GetHash())
		resultMap["code"] = model.Success
		return resultMap
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		resultMap["code"] = model.FailPwd
		resultMap["error"] = "pwd"
		return resultMap
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		resultMap["code"] = rpc.BalanceNotEnough
		resultMap["error"] = "amount"
		return resultMap
	}
	resultMap["error"] = err.Error()
	resultMap["code"] = model.Nomarl
	return resultMap
}
