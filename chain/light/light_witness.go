package light

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"sort"
	"strconv"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/mining"
	"web3_gui/chain/rpc"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/message_center"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
)

func RegisterWitnessMsg() {
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_VOTEINNEW, VoteInNew)                     //投票见证人押金
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_VOTEOUTNEW, VoteOutNew)                   //退还投票见证人押金
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETVOTELISTNEW, GetVoteListNew)           //退还投票见证人押金
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETWITNESSINFO, GetWitnessInfo)           //获取见证人状态
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCANDIDATELIST, GetCandidateList)       //获得候选见证人列表
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETWITNESSESLIST, GetWitnessesList)       //获得见证人列表
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCOMMUNITYLISTNEW, GetCommunityListNew) //获取社区节点列表
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD, GetCommunityReward)   //获取一个社区累计奖励
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD, SendCommunityReward) //分发社区奖励
	Area.Register_neighbor(config.MSGID_LIGHT_NODE_GETVOTEADDR, GetVoteAddr)                 //分发社区奖励
}

type Vote struct {
	VoteType uint16
	Rate     uint16
	VoteTo   crypto.AddressCoin
	Voter    crypto.AddressCoin
	Amount   uint64
	Gas      uint64
	GasPrice uint64
	Pwd      string
	Name     string
}

// 返回投票需要的payload
func VoteInNew(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	v := new(Vote)
	err := json.Unmarshal(*message.Body.Content, &v)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	var payload []byte
	isWitness := mining.SearchWitnessIsExist(v.Voter)
	switch v.VoteType {
	case mining.VOTE_TYPE_community: //1=给见证人投票
		//是否是轻节点验证通过合约内部验证
		if isWitness {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "The voting address is already another role"))
			return
		}
		if !mining.SearchWitnessIsExist(v.VoteTo) {
			//被投票地址是见证人角色
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "the voted address must be is witness"))
			return
		}
		//检查押金
		if v.Amount != config.Mining_vote {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "Community node deposit is "+strconv.Itoa(int(config.Mining_vote/1e8))))
			return
		}
		//构造合约payload
		payload = precompiled.BuildAddCommunityInput(v.VoteTo, v.Rate, v.Name)
	case mining.VOTE_TYPE_vote: //2=给社区节点投票
		if isWitness {
			//投票地址已经是其他角色了
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "The voting address is already another role"))
			return
		}
		if mining.SearchWitnessIsExist(v.VoteTo) {
			//被投票地址是见证人角色
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "the voted address is witness"))
			return
		}
		//构建合约payload
		payload = precompiled.BuildAddVoteInput(v.VoteTo)
	case mining.VOTE_TYPE_light: //3=轻节点押金
		if isWitness {
			//投票地址已经是其他角色了
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "The voting address is already another role"))
			return
		}

		if v.Amount != config.Mining_light_min {
			//轻节点押金是
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "Light node deposit is "+strconv.Itoa(int(config.Mining_light_min/1e8))))
			return
		}
		v.VoteTo = nil
		//构造合约payload
		payload = precompiled.BuildAddLightInput(v.Name)
	default:
		//不能识别的投票类型
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(rpc.SystemError, "Unrecognized voting type"))
		return
	}
	engine.Log.Info(hex.EncodeToString(payload))
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEINNEW_REV, pkg(model.Success, hex.EncodeToString(payload)))
}

type VoteO struct {
	VoteType uint16
	Addr     crypto.AddressCoin
	Amount   uint64
	Gas      uint64
	GasPrice uint64
	Pwd      string
	Payload  string
}

func VoteOutNew(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	vo := new(VoteO)
	err := json.Unmarshal(*message.Body.Content, &vo)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEOUTNEW_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	payload := []byte{}
	switch vo.VoteType {
	case mining.VOTE_TYPE_community:
		//构造社区节点退出输入
		payload = precompiled.BuildDelCommunity()
	case mining.VOTE_TYPE_vote:
		payload = precompiled.BuildDelVote(big.NewInt(int64(vo.Amount)))
	case mining.VOTE_TYPE_light:
		payload = precompiled.BuildDelLight()
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_VOTEOUTNEW_REV, pkg(model.Success, hex.EncodeToString(payload)))
}

/*
获得自己给哪些见证人投过票的列表
@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
*/
type VoteList struct {
	VoteType uint16
	Addr     []*keystore.AddressInfo
}
type Vinfos struct {
	infos []rpc.VoteInfoVO
}

func (this *Vinfos) Len() int {
	return len(this.infos)
}

func (this *Vinfos) Less(i, j int) bool {
	if this.infos[i].Height < this.infos[j].Height {
		return false
	} else {
		return true
	}
}

func (this *Vinfos) Swap(i, j int) {
	this.infos[i], this.infos[j] = this.infos[j], this.infos[i]
}
func GetVoteListNew(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	vl := new(VoteList)
	err := json.Unmarshal(*message.Body.Content, &vl)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETVOTELISTNEW_REV, pkg(rpc.SystemError, err.Error()))
		return
	}

	items := make([]*mining.DepositInfo, 0)
	list := []crypto.AddressCoin{}
	for _, one := range vl.Addr {
		list = append(list, one.Addr)
	}

	switch vl.VoteType {
	case mining.VOTE_TYPE_community:
		clist := precompiled.GetCommunityList(list)
		for _, v := range clist {
			coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
			witCoin := evmutils.AddressToAddressCoin(v.Wit.Bytes())
			items = append(items, &mining.DepositInfo{
				WitnessAddr: witCoin,
				SelfAddr:    coin,
				Value:       v.Score.Uint64(),
			})
		}
	case mining.VOTE_TYPE_light:
		clist := precompiled.GetLightList(list)
		for _, v := range clist {
			coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
			witCoin := evmutils.AddressToAddressCoin(v.C.Bytes())
			items = append(items, &mining.DepositInfo{
				WitnessAddr: witCoin,
				SelfAddr:    coin,
				Value:       v.Score.Uint64(),
				Name:        v.Cname,
			})
		}
	case mining.VOTE_TYPE_vote:
		clist := precompiled.GetLightList(list)
		for _, v := range clist {
			coin := evmutils.AddressToAddressCoin(v.Addr.Bytes())
			witCoin := evmutils.AddressToAddressCoin(v.C.Bytes())
			items = append(items, &mining.DepositInfo{
				WitnessAddr: witCoin,
				SelfAddr:    coin,
				Value:       v.Vote.Uint64(),
				Name:        v.Cname,
			})
		}
	}
	vinfos := Vinfos{
		infos: make([]rpc.VoteInfoVO, 0, len(items)),
	}
	for _, item := range items {
		var name string
		if vl.VoteType == mining.VOTE_TYPE_community {
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
		vinfos.infos = append(vinfos.infos, viVO)
	}
	sort.Stable(&vinfos)
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETVOTELISTNEW_REV, pkg(model.Success, vinfos.infos))
}

/*
获取候选见证人列表
*/
func GetWitnessInfo(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	var addr *keystore.AddressInfo
	err := json.Unmarshal(*message.Body.Content, &addr)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSINFO_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	winfo := rpc.WitnessInfo{}
	chain := mining.GetLongChain()
	if chain == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSINFO_REV, pkg(model.Success, winfo))
		return
	}

	var witnessAddr crypto.AddressCoin
	winfo.IsCandidate, winfo.IsBackup, winfo.IsKickOut, witnessAddr, winfo.Value = mining.GetWitnessStatusByAddr(addr)
	winfo.Addr = witnessAddr.B58String()

	winfo.Payload = mining.FindWitnessName(addr.Addr)
	winfo.Value = 0
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSINFO_REV, pkg(model.Success, winfo))
}

/*
获取候选见证人列表
*/
func GetCandidateList(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	currWitness := mining.GetLongChain().WitnessChain.GetCurrGroupLastWitness()
	wbg := currWitness.WitnessBigGroup

	allWiness := append(wbg.Witnesses, wbg.WitnessBackup...)
	addrs := make([]common.Address, len(allWiness))
	for k, one := range allWiness {
		addrs[k] = common.Address(evmutils.AddressCoinToAddress(*one.Addr))
	}

	from := config.Area.Keystore.GetCoinbase().Addr
	_, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(from, addrs)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCANDIDATELIST_REV, pkg(rpc.SystemError, "address"))
		return
	}

	wvos := make([]rpc.WitnessVO, 0)
	for k, one := range allWiness {
		wvo := rpc.WitnessVO{
			Addr:            one.Addr.B58String(),              //见证人地址
			Payload:         mining.FindWitnessName(*one.Addr), //
			Score:           one.Score,                         //押金
			Vote:            votes[k].Uint64(),                 //      voteValue,            //投票票数
			CreateBlockTime: one.CreateBlockTime,               //预计出块时间
		}
		wvos = append(wvos, wvo)

	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCANDIDATELIST_REV, pkg(model.Success, wvos))
}

/*
获取见证人列表
*/
func GetWitnessesList(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	wbg := mining.GetWitnessListSort()

	wvos := make([]rpc.WitnessVO, 0)
	for _, one := range wbg.Witnesses {
		name := mining.FindWitnessName(*one.Addr)
		vote := precompiled.GetWitnessVote(*one.Addr)
		wvo := rpc.WitnessVO{
			Addr:            one.Addr.B58String(), //见证人地址
			Payload:         name,                 //
			Score:           one.Score,            //押金
			Vote:            vote,                 //      voteValue,            //投票票数
			CreateBlockTime: one.CreateBlockTime,  //预计出块时间
		}
		wvos = append(wvos, wvo)
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETWITNESSESLIST_REV, pkg(model.Success, wvos))
}

/*
获取社区节点列表通过合约获取
*/
func GetCommunityListNew(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	vss := mining.GetCommunityListSortNew()
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYLISTNEW_REV, pkg(model.Success, vss))
}

/*
获取一个社区累计奖励
*/
func GetCommunityReward(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(model.NoField, "address"))
		return
	}
	addrStr := addrItr.(string)
	if addrStr != "" {
		addrMul := crypto.AddressFromB58String(addrStr)
		addr = &addrMul
	}

	if addrStr != "" {
		dst := crypto.AddressFromB58String(addrStr)
		if !crypto.ValidAddr(config.AddrPre, dst) {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(rpc.ContentIncorrectFormat, "address"))
			return
		}
	}
	if mining.GetAddrState(*addr) != 2 {
		//本地址不是社区节点
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(rpc.RuleField, "address"))
		return
	}

	chain := mining.GetLongChain()
	if chain == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(rpc.ChainNotSync, "The chain end is not synchronized"))
		return
	}
	currentHeight := chain.GetCurrentBlock()

	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(addr)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//查询代上链的奖励是否已经上链
		*notSend, _ = rpc.CheckTxUpchain(*notSend, currentHeight)
		//有未分配完的奖励
		community := ns.Reward / 10
		light := ns.Reward - community
		lightNum := ns.LightNum
		if lightNum > 0 {
			lightNum--
		}
		rewardTotal := mining.RewardTotal{
			CommunityReward: community,                           //社区节点奖励
			LightReward:     light,                               //轻节点奖励
			StartHeight:     ns.StartHeight,                      //
			Height:          ns.EndHeight,                        //最新区块高度
			IsGrant:         false,                               //是否可以分发奖励，24小时候才可以分发奖励
			AllLight:        lightNum,                            //所有轻节点数量
			RewardLight:     ns.LightNum - uint64(len(*notSend)), //已经奖励的轻节点数量
			IsNew:           false,
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(model.Success, rewardTotal))
		return
	}
	if ns == nil {
		//需要加载以前的奖励快照
		rpc.FindCommunityStartHeightByAddr(*addr)
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(model.Nomarl, "load reward history"))
		return
	}
	startHeight := ns.EndHeight + 1
	//奖励都分配完了，查询新的奖励
	rt, _, err := mining.GetRewardCount(addr, startHeight, 0)
	if err != nil {
		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(rpc.RewardCountSync, err.Error()))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	rt.IsNew = true

	lightNum := rt.AllLight
	if lightNum > 0 {
		lightNum--
	}
	rewardTotal := mining.RewardTotal{
		CommunityReward: rt.CommunityReward, //社区节点奖励
		LightReward:     rt.LightReward,     //轻节点奖励
		StartHeight:     rt.StartHeight,     //
		Height:          rt.Height,          //最新区块高度
		IsGrant:         rt.IsGrant,         //是否可以分发奖励，24小时候才可以分发奖励
		AllLight:        lightNum,           //所有轻节点数量
		RewardLight:     rt.RewardLight,     //已经奖励的轻节点数量
		IsNew:           rt.IsNew,
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETCOMMUNITYREWARD_REV, pkg(model.Success, rewardTotal))
	return
}

/*
分发社区奖励
*/
func SendCommunityReward(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addrItr, ok := rj.Get("address") //社区节点地址
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.NoField, "address"))
		return
	}
	addrStr := addrItr.(string)
	if addrStr == "" {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}
	addr := crypto.AddressFromB58String(addrStr)
	if !crypto.ValidAddr(config.AddrPre, addr) {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	if mining.GetAddrState(addr) != 2 {
		//本地址不是社区节点
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.RuleField, "address"))
		return
	}
	// engine.Log.Info("22222222222222")
	//查询社区节点公钥
	puk, ok := config.Area.Keystore.GetPukByAddr(addr)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.PubKeyNotExists, config.ERROR_public_key_not_exist.Error()))
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.NoField, "gas"))
		return
	}
	gas := uint64(gasItr.(float64))
	// engine.Log.Info("22222222222222")
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.NoField, "pwd"))
		return
	}
	pwd := pwdItr.(string)

	startHeightItr, ok := rj.Get("startheight")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.NoField, "startheight"))
		return
	}
	startHeight := uint64(startHeightItr.(float64))

	endheightItr, ok := rj.Get("endheight")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.NoField, "endheight"))
		return
	}
	endheight := uint64(endheightItr.(float64))
	chain := mining.GetLongChain()
	if chain == nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.ChainNotSync, "The chain end is not synchronized"))
		return
	}
	currentHeight := chain.GetCurrentBlock()
	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(&addr)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	// engine.Log.Info("22222222222222")
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//查询待上链的奖励是否已经上链
		*notSend, ok = rpc.CheckTxUpchain(*notSend, currentHeight)
		if ok {
			//有未上链的奖励
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.RewardNotLinked, "There are rewards that are not linked"))
			return
		}
		cs := mining.NewCommunitySign(puk, ns.StartHeight, ns.EndHeight)
		//有未分配完的奖励，继续分配
		err = mining.DistributionReward(&addr, notSend, gas, pwd, cs, ns.StartHeight, ns.EndHeight, currentHeight)
		if err != nil {
			if err.Error() == config.ERROR_password_fail.Error() {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.FailPwd, ""))
				return
			}
			if err.Error() == config.ERROR_not_enough.Error() {
				_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.NotEnough, ""))
				return
			}
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.Nomarl, err.Error()))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.Success, ""))
		return
	}
	//检查奖励开始高度，避免重复奖励
	if ns != nil && startHeight <= ns.EndHeight {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.RepeatReward, "Repeat reward"))
		return
	}
	// engine.Log.Info("22222222222222")
	//奖励都分配完了，查询新的奖励
	rt, notSend, err := mining.GetRewardCount(&addr, startHeight, endheight)
	if err != nil {
		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
			_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.RewardCountSync, err.Error()))
			return
		}
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	if !rt.IsGrant {
		//now := config.TimeNow().Unix()
		//pl time
		now := config.TimeNow().UnixNano()
		blockNum := uint64(config.Mining_community_reward_time/config.Mining_block_time) - (rt.Height - rt.StartHeight)
		wait := blockNum * uint64(config.Mining_block_time.Nanoseconds())
		futuer := time.Unix(0, now+int64(wait))
		// engine.Log.Info("22222222222222")
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(rpc.DistributeRewardTooEarly,
			"Please distribute the reward after "+futuer.Format("2006-01-02 15:04:05")))
		return
	}
	//创建快照
	err = mining.CreateRewardCount(addr, rt, *notSend)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.Nomarl, err.Error()))
		return
	}
	//创建快照成功后，删除缓存
	mining.CleanRewardCountProcessMap(&addr)
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_SENDCOMMUNITYREWARD_REV, pkg(model.Success, ""))
}

/*
获取一个地址投票给谁，投了多少票
*/
func GetVoteAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	rj := new(model.RpcJson)
	err := json.Unmarshal(*message.Body.Content, &rj)
	if err != nil {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETVOTEADDR_REV, pkg(rpc.SystemError, err.Error()))
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETVOTEADDR_REV, pkg(model.NoField, "address"))
		return
	}
	if !rj.VerifyType("address", "string") {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETVOTEADDR_REV, pkg(model.TypeWrong, "address"))
		return
	}

	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	if !ok {
		_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETVOTEADDR_REV, pkg(rpc.ContentIncorrectFormat, "address"))
		return
	}

	//查余额
	value, valueFrozen, _ := mining.GetNotspendByAddrOther(mining.GetLongChain(), addrCoin)
	//查角色
	role := 4
	depositIn := uint64(0)
	//查社区节点地址
	voteAddr := ""
	//查投票金额
	voteIn := uint64(0)

	//是轻节点
	light := precompiled.GetLightList([]crypto.AddressCoin{addrCoin})
	if len(light) > 0 && light[0].Score.Uint64() != 0 {
		role = 3
		depositIn = config.Mining_light_min
		communityAddr := evmutils.AddressToAddressCoin(light[0].C.Bytes())
		voteAddr = communityAddr.B58String()
		voteIn = light[0].Vote.Uint64()
	}

	//是社区节点
	community := precompiled.GetCommunityList([]crypto.AddressCoin{addrCoin})
	if len(community) > 0 && community[0].Wit != evmutils.ZeroAddress {
		role = 2
		depositIn = config.Mining_vote
		witnessAddr := evmutils.AddressToAddressCoin(community[0].Wit.Bytes())
		voteAddr = witnessAddr.B58String()
		voteIn = community[0].Vote.Uint64()
	}

	//是见证人
	if mining.GetDepositWitnessAddr(&addrCoin) > 0 {
		role = 1
		depositIn = config.Mining_deposit
	}

	getaccount := rpc.AddrInfo{
		Balance:       value,
		BalanceFrozen: valueFrozen,
		Role:          role,      //角色
		DepositIn:     depositIn, //押金
		VoteAddr:      voteAddr,  //给哪个社区节点地址投票
		VoteIn:        voteIn,    //投票金额
		VoteNum:       mining.GetLongChain().WitnessBackup.GetCommunityVote(&addrCoin),
	}
	_ = Area.SendP2pReplyMsg(message, config.MSGID_LIGHT_NODE_GETVOTEADDR_REV, pkg(model.Success, getaccount))
	return
}
