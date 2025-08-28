package rpc

import (
	"encoding/hex"
	"net/http"
	"sort"

	"web3_gui/chain/config"
	"web3_gui/chain/mining"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

func VoteInNew(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	var voter crypto.AddressCoin
	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)
	if addr == "" {
		res, err = model.Errcode(model.NoField, "address")
		return
	} else {
		voter = crypto.AddressFromB58String(addr)
		if !crypto.ValidAddr(config.AddrPre, voter) {
			res, err = model.Errcode(ContentIncorrectFormat, "address")
			return
		}
	}

	var voteTo crypto.AddressCoin
	witnessAddrItr, ok := rj.Get("witness")
	if voteType != mining.VOTE_TYPE_light {
		if !ok {
			res, err = model.Errcode(ContentIncorrectFormat, "witness")
			return
		}
		witnessStr := witnessAddrItr.(string)
		if witnessStr != "" {
			voteTo = crypto.AddressFromB58String(witnessStr)
			if !crypto.ValidAddr(config.AddrPre, voteTo) {
				res, err = model.Errcode(ContentIncorrectFormat, "witness")
				return
			}
		} else {
			res, err = model.Errcode(ContentIncorrectFormat, "witness")
			return
		}
	}

	var rate uint16
	switch voteType {
	case mining.VOTE_TYPE_community:
		rateItr, ok := rj.Get("rate")
		if !ok {
			res, err = model.Errcode(model.NoField, "rate")
			return
		}
		rate = uint16(rateItr.(float64))
		if rate > 100 {
			res, err = model.Errcode(DistributeRatioTooBig, "分配比例不能大于100")
			return
		}
	case mining.VOTE_TYPE_vote:
	case mining.VOTE_TYPE_light:
		voteTo = nil
	default:
		res, err = model.Errcode(SystemError, "votetype")
		return
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := toUint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	if gas < config.Wallet_tx_gas_min {
		res, err = model.Errcode(5002, "gas is less tx_gas_min")
		return
	}
	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	payload := ""
	payloadItr, ok := rj.Get("payload")
	if ok {
		payload = payloadItr.(string)
	}

	//从31万个块高度之后，才开放见证人和社区节点质押
	heightBlock := mining.GetHighestBlock()
	if heightBlock <= config.Wallet_vote_start_height {
		res, err = model.Errcode(VoteNotOpen)
		return
	}

	//查询余额是否足够
	value, _, _ := mining.FindBalanceValue()
	if amount > value {
		res, err = model.Errcode(BalanceNotEnough)
		return
	}

	err = mining.VoteIn(voteType, rate, voteTo, voter, amount, gas, pwd, payload, gasPrice)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		} else if err.Error() == config.ERROR_not_enough.Error() {
			res, err = model.Errcode(BalanceNotEnough)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

// 检查是否能成为见证者节点
func CheckCanWitness(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	data := "true"
	// 正式网络如果不是公网ip  就不能成为见证者节点
	//if config.Area.NodeManager.NetType == gconfig.NetType_release {
	//	if !utils.IsOnlyIp(config.Area.NodeManager.NodeSelf.Addr) {
	//		data = "false"
	//		res, err = model.Tojson(data)
	//		return res, err
	//	}
	//
	//	// 最小cpu数量
	//	if runtime.NumCPU() < config.WitnessMinCpuNum {
	//		data = "false"
	//		res, err = model.Tojson(data)
	//		return res, err
	//	}
	//
	//	// 最小内存
	//	m, errV := mem.VirtualMemory()
	//	if errV != nil {
	//		res, err = model.Errcode(SystemError, err.Error())
	//		return res, err
	//	}
	//	if m.Total < uint64(config.WitnessMinMem) {
	//		data = "false"
	//		res, err = model.Tojson(data)
	//		return res, err
	//	}
	//
	//}
	res, err = model.Tojson(data)
	return
}

func VoteOutNew(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	addrStr := ""
	var addr crypto.AddressCoin
	addrItr, ok := rj.Get("address")
	if ok {
		addrStr = addrItr.(string)
	}
	if addrStr != "" {
		addr = crypto.AddressFromB58String(addrStr)
		if !crypto.ValidAddr(config.AddrPre, addr) {
			res, err = model.Errcode(ContentIncorrectFormat, "address")
			return
		}
	}
	amount := toUint64(0)

	switch voteType {
	case mining.VOTE_TYPE_community:
	case mining.VOTE_TYPE_vote:
		amountItr, ok := rj.Get("amount")
		if !ok {
			res, err = model.Errcode(5002, "amount")
			return
		}
		amount = toUint64(amountItr.(float64))
		if amount <= 0 {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
	case mining.VOTE_TYPE_light:
	default:
		res, err = model.Errcode(ParamError, "votetype")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))
	if gas < config.Wallet_tx_gas_min {
		res, err = model.Errcode(5002, "gas is less tx_gas_min")
		return
	}
	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	payload := ""
	payloadItr, ok := rj.Get("payload")
	if ok {
		payload = payloadItr.(string)
	}

	err = mining.VoteOut(voteType, addr, amount, gas, pwd, payload, gasPrice)
	if err != nil {
		// engine.Log.Info("--------------- 取消投票错误" + err.Error())

		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		//余额不足
		if err.Error() == config.ERROR_not_enough.Error() {
			res, err = model.Errcode(BalanceNotEnough)
			return
		}
		//投票已经存在
		if err.Error() == config.ERROR_vote_exist.Error() {
			res, err = model.Errcode(VoteExist)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

/*
获得自己给哪些见证人投过票的列表
@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
*/
func GetVoteListNew(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	var items []*mining.DepositInfo
	switch voteType {
	case mining.VOTE_TYPE_community:
		items = mining.GetDepositCommunityListNew()
	case mining.VOTE_TYPE_light:
		items = mining.GetDepositLightListNew()
	case mining.VOTE_TYPE_vote:
		items = mining.GetDepositVoteListNew()
	}
	vinfos := Vinfos{
		infos: make([]VoteInfoVO, 0, len(items)),
	}
	for _, item := range items {
		var name string
		if voteType == mining.VOTE_TYPE_community {
			name = mining.FindWitnessName(item.WitnessAddr)
		} else {
			name = item.Name
		}

		viVO := VoteInfoVO{
			// Txid:        hex.EncodeToString(ti.Txid), //
			WitnessAddr: item.WitnessAddr.B58String(), //见证人地址
			Value:       item.Value,                   //投票数量
			// Height:      item.Height,           //区块高度
			AddrSelf: item.SelfAddr.B58String(), //自己投票的地址
			Payload:  name,                      //
		}
		vinfos.infos = append(vinfos.infos, viVO)
	}
	sort.Stable(&vinfos)
	res, err = model.Tojson(vinfos.infos)
	return
}

func WithDrawReward(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	return CommunityDistribute(rj, w, r) // 使用新接口逻辑

	/*
		var src crypto.AddressCoin
		addrItr, ok := rj.Get("srcaddress")
		if ok {
			srcaddr := addrItr.(string)
			if srcaddr != "" {
				src = crypto.AddressFromB58String(srcaddr)
				//判断地址前缀是否正确
				if !crypto.ValidAddr(config.AddrPre, src) {
					res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
					return
				}
				_, ok := config.Area.Keystore.FindAddress(src)
				if !ok {
					res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
					return
				}
			}
		}
		gasItr, ok := rj.Get("gas")
		if !ok {
			res, err = model.Errcode(model.NoField, "gas")
			return
		}
		gas := toUint64(gasItr.(float64))
		//var gasLimit = config.Wallet_tx_gas_min
		//if config.EVM_GAS_MAX > config.Wallet_tx_gas_min {
		//	gasLimit = config.EVM_GAS_MAX
		//}
		//if gas > gasLimit {
		//	res, err = model.Errcode(GasTooBig, "gas is too big")
		//	return
		//}
		gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
		gasPriceItr, ok := rj.Get("gas_price")
		if ok {
			gasPrice = toUint64(gasPriceItr.(float64))
			if gasPrice < config.DEFAULT_GAS_PRICE {
				res, err = model.Errcode(model.Nomarl, "gas_price is too low")
				return
			}
		}
		frozenHeight := toUint64(0)
		frozenHeightItr, ok := rj.Get("frozen_height")
		if ok {
			frozenHeight = toUint64(frozenHeightItr.(float64))
		}

		pwdItr, ok := rj.Get("pwd")
		if !ok {
			res, err = model.Errcode(model.NoField, "pwd")
			return
		}
		pwd := pwdItr.(string)
		amount := toUint64(0)
		amountItr, ok := rj.Get("amount")
		if !ok {
			res, err = model.Errcode(5002, "amount")
			return
		}
		amount = toUint64(amountItr.(float64))
		if amount <= 0 {
			res, err = model.Errcode(AmountIsZero, "amount")
			return
		}
		drawType := toUint64(1)
		drawTypeItr, ok := rj.Get("draw_type")
		if !ok {
			res, err = model.Errcode(5002, "draw_type")
			return
		}
		drawType = toUint64(drawTypeItr.(float64))
		if drawType != 1 && drawType != 2 && drawType != 3 {
			res, err = model.Errcode(model.Nomarl, "draw_type is illegal")
			return
		}
		comment := common.Bytes2Hex(precompiled.BuildWithDrawRewardInput(big.NewInt(int64(amount)), drawType))

		total, _ := mining.GetLongChain().GetBalance().BuildPayVinNew(&src, gas)
		if total < gas {
			//资金不够
			res, err = model.Errcode(BalanceNotEnough)
			return
		}

		txpay, err := mining.ContractTx(&src, &precompiled.RewardContract, 0, gas, frozenHeight, pwd, comment, "", 0, gasPrice, "")
		// engine.Log.Info("转账耗时 %s", config.TimeNow().Sub(startTime))
		if err != nil {
			// engine.Log.Info("创建转账交易错误 11111111")
			if err.Error() == config.ERROR_password_fail.Error() {
				// engine.Log.Info("创建转账交易错误 222222222222")
				res, err = model.Errcode(model.FailPwd)
				return
			}
			// engine.Log.Info("创建转账交易错误 333333333333")
			if err.Error() == config.ERROR_amount_zero.Error() {
				res, err = model.Errcode(AmountIsZero, "amount")
				return
			}
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}

		result, err := utils.ChangeMap(txpay)
		if err != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(result)

		return res, err
	*/
}

// 社区分发奖励
// 社区向自己的所有的轻节点分发奖励，并清零社区奖励池
// srcaddress 是社区节点/轻节点地址
func CommunityDistribute(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			//判断地址前缀是否正确
			if !crypto.ValidAddr(config.AddrPre, src) {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
			_, ok := config.Area.Keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := toUint64(gasItr.(float64))

	gasPrice := toUint64(config.DEFAULT_GAS_PRICE)
	gasPriceItr, ok := rj.Get("gas_price")
	if ok {
		gasPrice = toUint64(gasPriceItr.(float64))
		if gasPrice < config.DEFAULT_GAS_PRICE {
			res, err = model.Errcode(model.Nomarl, "gas_price is too low")
			return
		}
	}
	//frozenHeight := toUint64(0)
	//frozenHeightItr, ok := rj.Get("frozen_height")
	//if ok {
	//	frozenHeight = toUint64(frozenHeightItr.(float64))
	//}

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	role := mining.GetAddrState(src)
	if !(role == 2 || role == 3) {
		res, err = model.Errcode(model.Nomarl, "not community or light address")
		return
	}

	//comment := common.Bytes2Hex(precompiled.BuildCommunityDistributeInput())
	//txpay, err := mining.ContractTx(&src, &precompiled.RewardContract, 0, gas, frozenHeight, pwd, comment, "", config.Wallet_tx_type_contract, gasPrice, "")
	txpay, err := mining.CreateTxVoteRewardNew(&src, gas, pwd)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	//result, err := utils.ChangeMap(txpay)
	//if err != nil {
	//	res, err = model.Errcode(model.Nomarl, err.Error())
	//	return
	//}
	result := make(map[string]interface{})
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)

	return res, err
}
