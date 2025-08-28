package mining

import (
	"bytes"
	"math/big"
	"sort"
	"sync"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
候选见证人
*/
type WitnessBackup struct {
	chain             *Chain           //
	lock              sync.RWMutex     //
	witnesses         []*BackupWitness //
	witnessesMap      sync.Map         //key:string=候选见证人地址;value:*BackupWitness=备用见证人;
	VoteCommunity     sync.Map         //保存所有社区节点。key:string=候选见证人地址;value:*[]*VoteScore=投票人和押金;
	VoteCommunityList sync.Map         //保存所有社区节点。key:string=投票者地址;value:*VoteScore=投票人和押金;
	Vote              sync.Map         //投票押金 key:string=社区节点地址;value:*[]*VoteScore=投票人和押金;注意：投票押金要和见证人分开，因为区块回滚的时候，恢复见证人就不方便恢复投票押金。
	VoteList          sync.Map         //投票押金 key:string=投票人地址;value:*VoteScore=投票人和押金;
	LightNode         sync.Map         //轻节点押金 key:string=投票人地址;value:*VoteScore=投票人和押金;
	Blacklist         sync.Map         //黑名单。记录见证人连续未出块次数，超过3次提出候选见证人排序。解除黑名单需要退还押金，重新缴押金。
}

func (this WitnessBackup) Len() int {
	return len(this.witnesses)
}

func (this WitnessBackup) Less(i, j int) bool {
	return this.witnesses[i].VoteNum > this.witnesses[j].VoteNum
}

func (this WitnessBackup) Swap(i, j int) {
	this.witnesses[i], this.witnesses[j] = this.witnesses[j], this.witnesses[i]
}

/*
获取候选见证人数量
*/
func (this *WitnessBackup) GetBackupWitnessTotal() uint64 {
	this.lock.RLock()
	total := len(this.witnesses)
	this.lock.RUnlock()
	return uint64(total)
}

/*
统计备用见证人和见证人投票
*/
func (this *WitnessBackup) CountWitness(txs *[]TxItr) {
	for _, one := range *txs {
		switch one.Class() {
		case config.Wallet_tx_type_deposit_in: // 见证人质押
			//这里决定了交易输出地址才是见证人地址。
			vout := (*one.GetVout())[0]
			score := vout.Value
			depositIn := one.(*Tx_deposit_in)
			this.addWitness(depositIn.Puk, score)
			if depositIn.Payload != nil && len(depositIn.Payload) > 0 {
				//保存见证人名称
				witnessAddr := crypto.BuildAddr(config.AddrPre, depositIn.Puk)
				SaveWitnessName(witnessAddr, string(one.GetPayload()))
			}
			//设置比例
			this.chain.Balance.witnessRatio.Store(utils.Bytes2string(vout.Address), depositIn.Rate)
			//调用合约设置比例
			//虚拟机合约模式使能
			//if config.EVM_Reward_Enable {
			//	precompiled.SetRate(vout.Address, depositIn.Rate)
			//}
		case config.Wallet_tx_type_deposit_out: // 见证人取消质押
			vinOne := (*one.GetVin())[0]
			addr := vinOne.GetPukToAddr()
			this.DelWitness(addr)
			this.DelBlackList(*addr)

			//删除见证人名称
			DelWitnessName(*addr)
			//witnessName := FindWitnessName(*addr)
			//if witnessName != "" {
			//	DelWitnessName(witnessName)
			//}

		}
	}
}

/*
从虚拟机中一次性获取
所有见证人投票数
所有社区节点地址
*/
func (this *WitnessBackup) getWitnessVoteAndCommunityAddr(txs *[]TxItr) (map[string]*big.Int, map[string]precompiled.Community) {
	// voteWg := sync.WaitGroup{}
	witnessAddrs := []common.Address{}
	lightAddrs := []crypto.AddressCoin{}
	for _, one := range *txs {
		switch one.Class() {
		case config.Wallet_tx_type_deposit_in:
		case config.Wallet_tx_type_deposit_out:
		case config.Wallet_tx_type_contract:
			vout := (*one.GetVout())[0]
			if !bytes.Equal(vout.Address, precompiled.RewardContract) {
				continue
			}

			voteType, wit, _ := precompiled.UnpackPayloadV1(one.GetPayload())
			switch voteType {
			case 1:
				if wit == nil {
					continue
				}

				witnessAddrs = append(witnessAddrs, common.Address(evmutils.AddressCoinToAddress(wit)))
			case 2:
				if wit == nil {
					continue
				}

				lightAddrs = append(lightAddrs, wit)
			}
		}

	}

	// 见证人地址:投票数
	witnessVotes := make(map[string]*big.Int)
	_, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(precompiled.RewardContract, witnessAddrs)
	if err != nil {
		return nil, nil
	}
	if len(witnessAddrs) == len(votes) {
		for i, addr := range witnessAddrs {
			witnessVotes[utils.Bytes2string(evmutils.AddressToAddressCoin(addr.Bytes()))] = votes[i]
		}
	}

	// 轻节点地址:社区地址
	ligthCommunityAddrs := make(map[string]precompiled.Community)
	if len(lightAddrs) > 0 {
		communityAddrs := precompiled.GetCommunityList(lightAddrs)
		if len(lightAddrs) == len(communityAddrs) {
			for i, addr := range lightAddrs {
				ligthCommunityAddrs[utils.Bytes2string(addr)] = communityAddrs[i]
			}
		}
	}

	return witnessVotes, ligthCommunityAddrs
}

/*
从虚拟机中一次性获取所有见证人投票数
*/
func (this *WitnessBackup) getWitnessVote(txs *[]TxItr) map[string]*big.Int {
	witnessAddrs := []common.Address{}
	for _, one := range *txs {
		switch one.Class() {
		case config.Wallet_tx_type_deposit_in:
		case config.Wallet_tx_type_deposit_out:
		case config.Wallet_tx_type_contract:
			vout := (*one.GetVout())[0]
			if !bytes.Equal(vout.Address, precompiled.RewardContract) {
				continue
			}

			voteType, wit, _ := precompiled.UnpackPayloadV1(one.GetPayload())
			switch voteType {
			case 1:
				if wit == nil {
					continue
				}

				witnessAddrs = append(witnessAddrs, common.Address(evmutils.AddressCoinToAddress(wit)))
			}
		}

	}

	// 见证人地址:投票数
	witnessVotes := make(map[string]*big.Int)
	_, votes, err := precompiled.GetRewardRatioAndVoteByAddrs(precompiled.RewardContract, witnessAddrs)
	if err != nil {
		return nil
	}

	if len(witnessAddrs) == len(votes) {
		for i, addr := range witnessAddrs {
			witnessVotes[utils.Bytes2string(evmutils.AddressToAddressCoin(addr.Bytes()))] = votes[i]
		}
	}

	return witnessVotes
}

/*
查询一个见证人是否是候选见证人
*/
func (this *WitnessBackup) FindWitness(witnessAddr crypto.AddressCoin) bool {
	_, ok := this.witnessesMap.Load(utils.Bytes2string(witnessAddr))
	return ok
}

/*
添加一个见证人到投票列表
*/
func (this *WitnessBackup) addWitness(puk []byte, score uint64) {
	witnessAddr := crypto.BuildAddr(config.AddrPre, puk)

	// engine.Log.Info("添加一个见证人 %s", witnessAddr.B58String())
	// _, ok := this.witnessesMap.Load(witnessAddr.B58String())
	_, ok := this.witnessesMap.Load(utils.Bytes2string(witnessAddr))
	if ok {
		// fmt.Println("见证人已经存在")
		return
	}

	witness := &BackupWitness{
		Addr:  &witnessAddr, //见证人地址
		Puk:   puk,          //
		Score: score,        //押金
		// Vote:  new(sync.Map), //投票押金
	}

	// engine.Log.Info("统计这个见证人投票 %s", witnessAddr.B58String())
	//如果已经有投票，统计投票数量
	// _, lvss := this.FindScoreNew(&witnessAddr)
	// for _, one := range lvss {
	// 	// engine.Log.Info("这个见证人已经有投票了 +%d", one.Vote)
	// 	witness.VoteNum = witness.VoteNum + one.Vote
	// }
	// engine.Log.Info("这个见证人的投票总额 %d", witness.VoteNum)

	//if config.EVM_Reward_Enable {
	//	witness.VoteNum = precompiled.GetWitnessVote(witnessAddr)
	//} else {
	balanceMgr := this.chain.GetBalance()
	witness.VoteNum = balanceMgr.GetWitnessVote(&witnessAddr)
	//}

	this.lock.Lock()
	this.witnesses = append(this.witnesses, witness)
	this.lock.Unlock()
	this.witnessesMap.Store(utils.Bytes2string(witnessAddr), witness)
}

/*
删除一个见证人
*/
func (this *WitnessBackup) DelWitness(witnessAddr *crypto.AddressCoin) {
	this.lock.Lock()
	//	fmt.Println("++++++删除备用见证人前", len(this.witnesses), witnessAddr.B58String())
	//TODO 有待提高速度
	for i, one := range this.witnesses {
		if !bytes.Equal(*witnessAddr, *one.Addr) {
			continue
		}
		temp := this.witnesses[:i]
		this.witnesses = append(temp, this.witnesses[i+1:]...)
		break
	}
	//	fmt.Println("++++++删除备用见证人后", len(this.witnesses))
	this.lock.Unlock()
	this.witnessesMap.Delete(utils.Bytes2string(*witnessAddr))
}

/*
添加一个投票
@voteType       uint16                 投票类型
@witnessAddr    *crypto.AddressCoin    被投者地址
@voteAddr       *crypto.AddressCoin    投票者地址
@score          uint64                 投票数量
*/
func (this *WitnessBackup) addVote(voteType uint16, witnessAddr, voteAddr *crypto.AddressCoin, score uint64, rate uint16) {
	//不能自己给自己投票
	if bytes.Equal(*witnessAddr, *voteAddr) {
		return
	}

	isWitness := this.haveWitness(voteAddr)
	_, isCommunity := this.haveCommunityList(voteAddr)
	_, isLight := this.haveLight(voteAddr)

	newVote := new(VoteScore)
	newVote.Witness = witnessAddr
	newVote.Addr = voteAddr
	newVote.Scores = score

	switch voteType {
	case 1: //1=给见证人投票
		if isLight || isWitness {
			//投票地址已经是其他角色了
			return
		}
		if score != config.Mining_vote {
			//投票金额不对
			return
		}

		vs, ok := this.haveCommunityList(voteAddr)
		if ok {
			if bytes.Equal(*vs.Witness, *witnessAddr) {
				vs.Scores = vs.Scores + score
				//重复投票
				return
			}
			//不能给多个见证人投票
			return
		}

		// engine.Log.Info("添加社区节点开始高度的区块:%s", voteAddr.B58String())

		//保存到投票者索引列表
		this.VoteCommunityList.Store(utils.Bytes2string(*voteAddr), newVote)
		//保存到见证人索引列表
		v, ok := this.VoteCommunity.Load(utils.Bytes2string(*witnessAddr))
		if ok {
			vss := v.(*[]*VoteScore)
			*vss = append(*vss, newVote)
		} else {
			vss := make([]*VoteScore, 0)
			vss = append(vss, newVote)
			this.VoteCommunity.Store(utils.Bytes2string(*witnessAddr), &vss)
		}

		//新添加的社区，统计这个社区投票数量
		v, ok = this.Vote.Load(utils.Bytes2string(*voteAddr))
		if ok {
			vss := v.(*[]*VoteScore)
			voteNum := uint64(0)
			for _, one := range *vss {
				voteNum = voteNum + one.Scores
			}
			newVote.Vote = voteNum
			//见证人已存在，则把投票数量加给见证人
			v, ok := this.witnessesMap.Load(utils.Bytes2string(*witnessAddr))
			if ok {
				bw := v.(*BackupWitness)
				bw.VoteNum = bw.VoteNum + voteNum
			}
		}
	case 2: //2=给社区节点投票
		// engine.Log.Info("给社区节点投票 1111111111111")
		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return
		}
		vs, ok := this.haveVoteList(voteAddr)
		if ok {
			//追加投票
			if !bytes.Equal(*vs.Witness, *witnessAddr) {
				//不能给多个社区节点投票
				return
			}

			// if bytes.Equal(*voteAddr, config.SpecialAddrs) {
			// 	engine.Log.Debug("1节点 %s 给社区节点 %s 原有:%d 增加投票:%d 剩余:%d", voteAddr.B58String(), witnessAddr.B58String(), vs.Scores, score, vs.Scores+score)
			// }
			//追加投票
			vs.Scores = vs.Scores + score
			// engine.Log.Info("给社区节点投票 222222222 d%", vs.Scores)
		} else {
			// engine.Log.Debug("1节点 %s 给社区节点 %s 增加投票 %d", voteAddr.B58String(), witnessAddr.B58String(), score)
			//			engine.Log.Info("给社区节点投票 %s %+v", voteAddr.B58String(), newVote)
			this.VoteList.Store(utils.Bytes2string(*voteAddr), newVote)

			v, ok := this.Vote.Load(utils.Bytes2string(*witnessAddr))
			if ok {
				vss := v.(*[]*VoteScore)
				*vss = append(*vss, newVote)
				// if bytes.Equal(*voteAddr, config.SpecialAddrs) {
				// 	engine.Log.Debug("1节点 %s 给社区节点 %s 原有:%d 增加投票:%d 剩余:%d", voteAddr.B58String(), witnessAddr.B58String(), 0, score, 0)
				// }
			} else {
				vss := make([]*VoteScore, 0)
				vss = append(vss, newVote)
				this.Vote.Store(utils.Bytes2string(*witnessAddr), &vss)
				// if bytes.Equal(*voteAddr, config.SpecialAddrs) {
				// 	engine.Log.Debug("1节点 %s 给社区节点 %s 原有:%d 增加投票:%d 剩余:%d", voteAddr.B58String(), witnessAddr.B58String(), 0, score, 0)
				// }
			}

		}

		//如果社区节点已经存在，则给社区节点添加投票数量
		v, ok := this.VoteCommunityList.Load(utils.Bytes2string(*witnessAddr))
		if ok {
			vs := v.(*VoteScore)
			vs.Vote = vs.Vote + score
			// engine.Log.Info("给社区节点投票 44444444444 d%", vs.Vote)
			if bytes.Equal(*voteAddr, config.SpecialAddrs) {
				// engine.Log.Debug("1节点 %s 给社区节点 %s 增加投票:%d 剩余:%d", voteAddr.B58String(), witnessAddr.B58String(), score, vs.Vote)
			}
			//如果见证人已经存在，则给见证人添加投票数量
			v, ok := this.witnessesMap.Load(utils.Bytes2string(*vs.Witness))
			if ok {
				bw := v.(*BackupWitness)
				bw.VoteNum = bw.VoteNum + score
				// engine.Log.Info("给社区节点投票 555555555555 d%", bw.VoteNum)
			}
		}
	case 3: //3=轻节点押金
		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return
		}

		if score != config.Mining_light_min {
			//投票金额不对
			return
		}

		v, ok := this.LightNode.Load(utils.Bytes2string(*voteAddr))
		if ok {
			vs := v.(*VoteScore)
			vs.Scores = vs.Scores + score
			return
		}
		this.LightNode.Store(utils.Bytes2string(*voteAddr), newVote)

	default:
		return
	}

}

/*
取消给一个见证人投票
@voteType       uint16                 投票类型
@witnessAddr    *crypto.AddressCoin    被投票者地址
@voteAddr       *crypto.AddressCoin    投票者地址
@score          uint64                 投票数量
*/
func (this *WitnessBackup) DelVote(voteType uint16, witnessAddr, voteAddr *crypto.AddressCoin, score uint64) {

	switch voteType {
	case 1: //1=给见证人投票

		// engine.Log.Info("删除社区节点开始高度的区块:%s", voteAddr.B58String())

		v, ok := this.VoteCommunityList.Load(utils.Bytes2string(*voteAddr))
		if !ok {
			// engine.Log.Info("删除社区节点开始高度的区块 不存在:%s", voteAddr.B58String())
			return
		}
		vs := v.(*VoteScore)
		vs.Scores = vs.Scores - score
		// engine.Log.Info("删除社区节点开始高度的区块:%s %d", voteAddr.B58String(), vs.Scores)
		//投票数量为0则删除这个节点的记录
		if vs.Scores > 0 {
			// engine.Log.Info("此社区节点票数大于0，则不删除:%s", voteAddr.B58String())
			return
		}

		//统计这个社区的所有投票，见证人减少这些投票
		v, ok = this.Vote.Load(utils.Bytes2string(*voteAddr))
		if ok {
			vss := v.(*[]*VoteScore)
			voteNum := uint64(0)
			for _, one := range *vss {
				voteNum = voteNum + one.Scores
			}
			//见证人已存在，则把投票数量减少
			v, ok := this.witnessesMap.Load(utils.Bytes2string(*witnessAddr))
			if ok {
				bw := v.(*BackupWitness)
				bw.VoteNum = bw.VoteNum - voteNum
			}
		}

		// engine.Log.Info("删除社区节点开始高度的区块:%s", voteAddr.B58String())
		//删除社区节点开始高度的区块
		db.LevelTempDB.Remove(config.BuildCommunityAddrStartHeight(*voteAddr))

		//先从投票者记录中删除
		this.VoteCommunityList.Delete(utils.Bytes2string(*voteAddr))
		//
		v, ok = this.VoteCommunity.Load(utils.Bytes2string(*witnessAddr))
		if !ok {
			return
		}
		vss := v.(*[]*VoteScore)
		for i, one := range *vss {
			if bytes.Equal(*one.Addr, *voteAddr) {
				temp := (*vss)[:i]
				temp = append(temp, (*vss)[i+1:]...)
				this.VoteCommunity.Store(utils.Bytes2string(*witnessAddr), &temp)
				break
			}
		}
	case 2: //2=给社区节点投票
		// engine.Log.Debug("1节点 %s 给社区节点 %s 减少投票 %d", voteAddr.B58String(), witnessAddr.B58String(), score)
		v, ok := this.VoteList.Load(utils.Bytes2string(*voteAddr))
		if !ok {
			// engine.Log.Debug("2节点 %s 给社区节点 %s 减少投票 %d", voteAddr.B58String(), witnessAddr.B58String(), score)
			return
		}
		vs := v.(*VoteScore)
		// if bytes.Equal(*voteAddr, config.SpecialAddrs) {
		// 	engine.Log.Debug("3节点 %s 给社区节点 %s 原有:%d 减少投票:%d 剩余:%d", voteAddr.B58String(), witnessAddr.B58String(), vs.Scores, score, vs.Scores-score)
		// }
		vs.Scores = vs.Scores - score
		//更新社区节点的投票
		//如果社区节点已经存在，则给社区节点减少投票数量
		v, ok = this.VoteCommunityList.Load(utils.Bytes2string(*witnessAddr))
		if ok {
			vs := v.(*VoteScore)
			vs.Vote = vs.Vote - score
			// engine.Log.Debug("4给社区节点减少投票 %d", vs.Vote)
			//更新见证人的投票
			//如果见证人已经存在，则给见证人减少投票数量
			v, ok := this.witnessesMap.Load(utils.Bytes2string(*vs.Witness))
			if ok {
				bw := v.(*BackupWitness)
				bw.VoteNum = bw.VoteNum - score
			}
		}

		//投票数量为0则删除这个节点的记录
		if vs.Scores > 0 {
			// engine.Log.Debug("5给社区节点减少投票 %d", vs.Scores)
			return
		}
		// engine.Log.Debug("6给社区节点减少投票 %d", vs.Scores)
		//先从投票者记录中删除
		this.VoteList.Delete(utils.Bytes2string(*voteAddr))

		v, ok = this.VoteList.Load(utils.Bytes2string(*voteAddr))
		if !ok {
			// engine.Log.Debug("7给社区节点减少投票 %s", voteAddr.B58String())
		} else {
			vs = v.(*VoteScore)
			// engine.Log.Debug("8给社区节点减少投票 %d", vs.Scores)
		}
		//
		v, ok = this.Vote.Load(utils.Bytes2string(*witnessAddr))
		if !ok {
			return
		}
		vss := v.(*[]*VoteScore)
		for i, one := range *vss {
			if bytes.Equal(*one.Addr, *voteAddr) {
				temp := (*vss)[:i]
				temp = append(temp, (*vss)[i+1:]...)
				this.Vote.Store(utils.Bytes2string(*witnessAddr), &temp)
				break
			}
		}
	case 3: //3=轻节点押金
		v, ok := this.LightNode.Load(utils.Bytes2string(*voteAddr))
		if !ok {
			return
		}
		vs := v.(*VoteScore)
		vs.Scores = vs.Scores - score
		//投票数量为0则删除这个节点的记录
		if vs.Scores > 0 {
			return
		}
		//从记录中删除
		this.LightNode.Delete(utils.Bytes2string(*voteAddr))
	}
}

/*
查找备用见证人列表中是否有查询的见证人
*/
func (this *WitnessBackup) haveWitness(witnessAddr *crypto.AddressCoin) (have bool) {
	this.lock.RLock()
	for _, one := range this.witnesses {
		have = bytes.Equal(*witnessAddr, *one.Addr)
		if have {
			break
		}
	}
	this.lock.RUnlock()
	return
}

/*
通过见证人查找是否有社区投票
*/
func (this *WitnessBackup) haveCommunity(witnessAddr *crypto.AddressCoin) (*[]*VoteScore, bool) {
	v, ok := this.VoteCommunity.Load(utils.Bytes2string(*witnessAddr))
	if ok {
		value := v.(*[]*VoteScore)
		return value, ok
	}
	return nil, ok
}

/*
通过投票者查找是否有社区投票
*/
func (this *WitnessBackup) haveCommunityList(addr *crypto.AddressCoin) (*VoteScore, bool) {
	v, ok := this.VoteCommunityList.Load(utils.Bytes2string(*addr))
	if ok {
		value := v.(*VoteScore)
		return value, ok
	}
	return nil, ok
}

/*
通过社区节点地址查询轻节点
*/
func (this *WitnessBackup) haveVote(witnessAddr *crypto.AddressCoin) (*[]*VoteScore, bool) {
	v, ok := this.Vote.Load(utils.Bytes2string(*witnessAddr))
	if ok {
		value := v.(*[]*VoteScore)
		return value, ok
	}
	return nil, ok
}

/*
通过投票者查找是否有投票者
*/
func (this *WitnessBackup) haveVoteList(addr *crypto.AddressCoin) (*VoteScore, bool) {
	v, ok := this.VoteList.Load(utils.Bytes2string(*addr))
	if ok {
		value := v.(*VoteScore)
		return value, ok
	}
	return nil, ok
}

/*
通过投票者查找轻节点押金
*/
func (this *WitnessBackup) haveLight(addr *crypto.AddressCoin) (*VoteScore, bool) {
	v, ok := this.LightNode.Load(utils.Bytes2string(*addr))
	if ok {
		value := v.(*VoteScore)
		return value, ok
	}
	return nil, ok
}

/*
参加选举的备用见证人
*/
type BackupWitness struct {
	Addr    *crypto.AddressCoin //见证人地址
	Puk     []byte              //见证人公钥
	Score   uint64              //评分
	VoteNum uint64              //投票押金总和
	// Vote  *sync.Map        //投票押金 key:string=投票人地址;value:*VoteScore=投票人和押金;
}

var createGroupTotal = uint64(0)

/*
根据这一时刻见证人投票排序，生成见证人链
@return    *Witness    备用见证人链中的一个见证人指针
*/
func (this *WitnessBackup) CreateWitnessGroup() []*Witness {
	if len(this.witnesses) < config.Witness_backup_min {
		// engine.Log.Info("见证人列表为空")
		return nil
	}
	wbg := this.GetWitnessListSortOld()
	// engine.Log.Info("创建见证人组次数:%d", atomic.AddUint64(&createGroupTotal, 1))

	//待出块的见证人，加入白名单连接
	IsBackup := this.chain.WitnessChain.FindWitness(Area.Keystore.GetCoinbase().Addr)
	if IsBackup {
		witnessAddrCoins := make([]*crypto.AddressCoin, 0)
		for _, one := range wbg.Witnesses {
			witnessAddrCoins = append(witnessAddrCoins, one.Addr)
		}
		//异步添加白名单连接
		go AddWitnessAddrNets(this.chain, witnessAddrCoins)
		// utils.Go(AddWitnessAddrNets(witnessAddrCoins))
	}

	//for _, one := range wbg.Witnesses {
	//	total := uint64(0)
	//	for _, two := range one.CommunityVotes {
	//		 engine.Log.Info("vote one :%s %d", two.Addr.B58String(), two.Vote)
	//		total += two.Vote
	//	}
	//	 engine.Log.Info("total :%d", total)
	//}

	random := this.chain.HashRandom()

	// engine.Log.Info("前n个块hash %s", hex.EncodeToString(*random))
	//bs := make([]byte, 0)
	//random := utils.Hash_SHA3_256(bs)
	wbg.Witnesses = OrderWitness(wbg.Witnesses, random)
	for i, _ := range wbg.Witnesses {
		// engine.Log.Info("排序后的顺序%d-%d:%s", len(wbg.Witnesses), i+1, wbg.Witnesses[i].Addr.B58String())
		wbg.Witnesses[i].WitnessBigGroup = wbg
	}

	// change := false
	// if randomHashTemp != nil && bytes.Equal(*random, randomHashTemp) {
	// 	for i, one := range wbg.WitnessBackup {
	// 		// engine.Log.Info("本次排序:%s", one.Addr.B58String())
	// 		if !bytes.Equal(*randomHashTempVlue[i].Addr, *one.Addr) {
	// 			change = true
	// 		}
	// 	}
	// }
	// if change {
	// 	for i, one := range randomHashTempVlue {
	// 		engine.Log.Info("原来的排序:%d %s", i, one.Addr.B58String())
	// 	}
	// 	for i, one := range wbg.Witnesses {
	// 		engine.Log.Info("新的排序:%d %s", i, one.Addr.B58String())
	// 	}
	// }
	// randomHashTemp = *random
	// randomHashTempVlue = wbg.Witnesses
	return wbg.Witnesses
}

// var randomHashTemp []byte         //
// var randomHashTempVlue []*Witness //

/*
打印备用见证人列表
*/
func (this *WitnessBackup) PrintWitnessBackup() {
	// fmt.Println("打印备用见证人")

	this.lock.Lock()
	// sort.Sort(this)
	sort.Stable(this)
	this.lock.Unlock()

	for i, one := range this.witnesses {
		if i >= config.Witness_backup_max {
			//只获取排名靠前的n个备用见证人
			break
		} else {
			engine.Log.Info("backup witness %d %s", i, one.Addr.B58String())
		}
	}
}

/*
加入黑名单
*/
func (this *WitnessBackup) AddBlackList(addr crypto.AddressCoin) {
	addrStr := utils.Bytes2string(addr)
	// engine.Log.Info("黑名单中 +1 %s", addrStr)
	//在备用见证人里面才加入黑名单
	// _, ok := this.witnessesMap.Load(addrStr)
	// if !ok {
	// 	engine.Log.Info("不在备用见证人列表中")
	// 	return
	// }
	v, ok := this.Blacklist.Load(addrStr)
	if ok {
		total := v.(uint64)
		//最多连续不出块3个
		// if total < 3 {
		total++
		// }
		this.Blacklist.Store(addrStr, total)
		// value, ok := this.Blacklist.Load(addrStr)
		// engine.Log.Info("黑名单中的值 %s %d %v", addrStr, value, ok)
		return
	}
	this.Blacklist.Store(addrStr, uint64(1))
	// value, ok := this.Blacklist.Load(addrStr)
	// engine.Log.Info("黑名单中的值 %s %d %v", addrStr, value, ok)
}

/*
继续出块，可以慢慢从列表中移除
*/
func (this *WitnessBackup) SubBlackList(addr crypto.AddressCoin) {
	addrStr := utils.Bytes2string(addr)
	// engine.Log.Info("黑名单中 -1 %s", addrStr)

	//在候选见证人列表里才-1
	_, ok := this.witnessesMap.Load(utils.Bytes2string(addr))
	if !ok {
		return
	}

	v, ok := this.Blacklist.Load(addrStr)
	if ok {
		total := v.(uint64)
		total--
		//最多连续不出块3个
		if total <= 0 {
			this.Blacklist.Delete(addrStr)
		} else {
			this.Blacklist.Store(addrStr, total)
		}
	}
	// value, ok := this.Blacklist.Load(addrStr)
	// engine.Log.Info("黑名单中的值 %s %d %v", addrStr, value, ok)
}

/*
从黑名单中移除
*/
func (this *WitnessBackup) DelBlackList(addr crypto.AddressCoin) {
	// engine.Log.Info("从黑名单中删除 %s", addr.B58String())
	this.Blacklist.Delete(utils.Bytes2string(addr))
}

/*
创建备用见证人列表
*/
func NewWitnessBackup(chain *Chain) *WitnessBackup {
	wb := WitnessBackup{
		chain:             chain, //
		lock:              *new(sync.RWMutex),
		witnesses:         make([]*BackupWitness, 0),
		witnessesMap:      *new(sync.Map),
		Vote:              *new(sync.Map),
		VoteList:          *new(sync.Map),
		VoteCommunity:     *new(sync.Map),
		VoteCommunityList: *new(sync.Map),
		LightNode:         *new(sync.Map),
	}
	return &wb
}

/*
投票押金，作为股权分红
*/
type VoteScore struct {
	Witness *crypto.AddressCoin //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
	Addr    *crypto.AddressCoin //投票人地址
	Scores  uint64              //押金
	Vote    uint64              //获得票数
	// Children *sync.Map           // []VoteScore         //给投票人的投票 key:string=;value:*VoteScore=;
}

/*
在黑名单中查找一个见证人地址，看是否存在
*/
func (this *WitnessBackup) FindWitnessInBlackList(addr crypto.AddressCoin) (have bool) {
	// addrStr := addr.B58String()
	// engine.Log.Info("查找一个见证人是否存在 %s", addrStr)
	total, ok := this.Blacklist.Load(utils.Bytes2string(addr))
	// engine.Log.Info("查找结果 %d %v", total, ok)
	if !ok {
		return false
	}
	t := total.(uint64)
	if config.CheckAddBlacklist(this.GetBackupWitnessTotal(), t) {
		return true
	}
	return false

	// if t := total.(uint64); t >= 3 {
	// 	return true
	// }
	// return false

	// have = false
	// this.Blacklist.Range(func(k, v interface{}) bool {
	// 	total := v.(uint64)
	// 	if total >= 3 {
	// 		addrOne := k.(string)
	// 		if addrStr == addrOne {
	// 			have = true
	// 			return false
	// 		}
	// 	}
	// 	return true
	// })
	// return
}

/*
根据这一时刻获取见证人投票排序
@return    *Witness    备用见证人链中的一个见证人指针
*/
func (this *WitnessBackup) GetWitnessListSortOld() *WitnessBigGroup {
	// currentHeight := this.chain.GetCurrentBlock()
	witCount := this.GetBackupWitnessTotal()
	//排除n次未出块的见证人
	this.Blacklist.Range(func(k, v interface{}) bool {
		total := v.(uint64)
		//ok := config.CheckAddBlacklist(this.GetBackupWitnessTotal(), total)
		ok := config.CheckAddBlacklist(witCount, total)
		// if currentHeight < config.CheckAddBlacklistChangeHeight {
		// 	ok = !ok
		// }
		if ok {
			// if total >= config.Wallet_not_build_block_max {
			addrStr := k.(string)
			// addr := crypto.AddressCoin([]byte(addrStr)) // crypto.AddressFromB58String(addrStr)
			addr := crypto.AddressCoin([]byte(addrStr))
			// engine.Log.Info("本次删除的备用见证人:%s", addr.B58String())
			this.DelWitness(&addr)
			// }
		}
		return true
	})

	//统计所有投票

	//按投票数量排序
	this.lock.Lock()
	sort.Sort(this)
	sort.Stable(this)
	this.lock.Unlock()
	wbg := WitnessBigGroup{
		Witnesses:     make([]*Witness, 0), //
		WitnessBackup: make([]*Witness, 0), //备用见证人
	}

	for i, _ := range this.witnesses {
		newWitness := new(Witness)
		newWitness.Addr = this.witnesses[i].Addr
		newWitness.Puk = this.witnesses[i].Puk
		newWitness.Score = this.witnesses[i].Score
		//newWitness.CommunityVotes, newWitness.Votes = this.FindScoreNew(newWitness.Addr)
		newWitness.VoteNum = this.witnesses[i].VoteNum
		newWitness.StopMining = make(chan bool, 1)
		// newWitness.GroupStart = false

		//取前n个见证人
		if i < config.Witness_backup_max {
			wbg.Witnesses = append(wbg.Witnesses, newWitness)
		} else {
			wbg.WitnessBackup = append(wbg.WitnessBackup, newWitness)
		}
	}

	// engine.Log.Info("打印备用见证人数量 %d %d", len(wbg.Witnesses), len(wbg.WitnessBackup))

	return &wbg
}

func (this *WitnessBackup) GetWitnessListSort() *WitnessBigGroup {
	// currentHeight := this.chain.GetCurrentBlock()
	//排除n次未出块的见证人
	witCount := this.GetBackupWitnessTotal()
	this.Blacklist.Range(func(k, v interface{}) bool {
		total := v.(uint64)
		ok := config.CheckAddBlacklist(witCount, total)
		// if currentHeight < config.CheckAddBlacklistChangeHeight {
		// 	ok = !ok
		// }
		if ok {
			// if total >= config.Wallet_not_build_block_max {
			addrStr := k.(string)
			// addr := crypto.AddressCoin([]byte(addrStr)) // crypto.AddressFromB58String(addrStr)
			addr := crypto.AddressCoin([]byte(addrStr))
			// engine.Log.Info("本次删除的备用见证人:%s", addr.B58String())
			this.DelWitness(&addr)
			// }
		}
		return true
	})

	//统计所有投票

	//按投票数量排序
	this.lock.Lock()
	sort.Sort(this)
	sort.Stable(this)
	this.lock.Unlock()
	wbg := WitnessBigGroup{
		Witnesses:     make([]*Witness, 0), //
		WitnessBackup: make([]*Witness, 0), //备用见证人
	}
	//balanceMgr := GetLongChain().GetBalance()
	//wmc := balanceMgr.GetWitnessMapCommunitys()
	//cml := balanceMgr.GetCommunityMapLights()
	for i, _ := range this.witnesses {
		newWitness := new(Witness)
		newWitness.Addr = this.witnesses[i].Addr
		newWitness.Puk = this.witnesses[i].Puk
		newWitness.Score = this.witnesses[i].Score
		newWitness.CommunityVotes, newWitness.Votes = this.FindScoreNew(newWitness.Addr)
		//if witInfo := balanceMgr.GetDepositWitness(newWitness.Addr); witInfo != nil {
		//	if items, ok := wmc.Load(utils.Bytes2string(*newWitness.Addr)); ok {
		//		commAddrs := items.([]crypto.AddressCoin)
		//		newWitness.CommunityVotes = make([]*VoteScore, len(commAddrs))
		//		for j, commAddr := range commAddrs {
		//			if commInfo := balanceMgr.GetDepositCommunity(&commAddr); commInfo != nil {
		//				newWitness.CommunityVotes[j] = &VoteScore{
		//					Witness: &commInfo.WitnessAddr,
		//					Addr:    &commInfo.SelfAddr,
		//					Scores:  config.Mining_vote,
		//					Vote:    commInfo.Value,
		//				}
		//				if items, ok := cml.Load(utils.Bytes2string(commInfo.SelfAddr)); ok {
		//					lightAddrs := items.([]crypto.AddressCoin)
		//					newWitness.Votes = make([]*VoteScore, len(lightAddrs))
		//					for k, lightAddr := range lightAddrs {
		//						if lightInfo := balanceMgr.GetDepositVote(&lightAddr); lightInfo != nil {
		//							newWitness.Votes[k] = &VoteScore{
		//								Witness: &lightInfo.WitnessAddr,
		//								Addr:    &lightInfo.SelfAddr,
		//								Scores:  config.Mining_light_min,
		//								Vote:    lightInfo.Value,
		//							}
		//						}
		//					}
		//				}
		//			}
		//		}
		//	}
		//}
		newWitness.VoteNum = this.witnesses[i].VoteNum
		newWitness.StopMining = make(chan bool, 1)
		// newWitness.GroupStart = false

		//取前n个见证人
		if i < config.Witness_backup_max {
			wbg.Witnesses = append(wbg.Witnesses, newWitness)
		} else {
			wbg.WitnessBackup = append(wbg.WitnessBackup, newWitness)
		}
	}

	// engine.Log.Info("打印备用见证人数量 %d %d", len(wbg.Witnesses), len(wbg.WitnessBackup))

	return &wbg
}

/*
查询一个见证人地址的投票数量
@return    []*VoteScore    社区投票列表
@return    []*VoteScore    轻节点投票列表
*/
func (this *WitnessBackup) FindScore(addr *crypto.AddressCoin) ([]*VoteScore, []*VoteScore) {
	// if bytes.Equal(*addr, config.SpecialAddrs) {
	// 	engine.Log.Info("查询这个见证人地址的所有投票 %s", addr.B58String())
	// }
	vssAll := make([]*VoteScore, 0)
	communityScore := make([]*VoteScore, 0)

	//先查找社区
	v, ok := this.VoteCommunity.Load(utils.Bytes2string(*addr))
	if !ok {
		// if bytes.Equal(*addr, config.SpecialAddrs) {
		// 	// engine.Log.Info("查询这个见证人地址的所有投票 %s", addr.B58String())
		// 	engine.Log.Info("查询这个见证人地址的所有投票 11111111111111111111")
		// }
		return communityScore, vssAll
	}
	vss := v.(*[]*VoteScore)

	for _, one := range *vss {
		vs := VoteScore{
			Witness: one.Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
			Addr:    one.Addr,    //投票人地址
			Scores:  one.Scores,  //押金
			Vote:    one.Vote,    //获得票数
		}
		communityScore = append(communityScore, &vs)

		// if bytes.Equal(*one.Addr, config.SpecialAddrs) {
		// 	engine.Log.Info("查询这个见证人地址的所有投票 %s %d", one.Addr.B58String(), one.Vote)
		// }

		// engine.Log.Info("查询这个见证人地址的所有投票 2222222222 %d %d", one.Scores, one.Vote)
	}
	// engine.Log.Info("查询这个见证人地址的所有投票 3333333333333")
	for _, one := range *vss {
		v, ok := this.Vote.Load(utils.Bytes2string(*one.Addr))
		if !ok {
			continue
		}
		vss := v.(*[]*VoteScore)
		voteNum := uint64(0)
		for _, one := range *vss {
			voteNum = voteNum + one.Scores
			vs := VoteScore{
				Witness: one.Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
				Addr:    one.Addr,    //投票人地址
				Vote:    one.Scores,  //获得票数
			}
			vssAll = append(vssAll, &vs)
			// engine.Log.Info("查询这个见证人地址的所有投票 444444444444 %s %s %d", one.Addr.B58String(), one.Witness.B58String(), one.Scores)
		}
		one.Vote = voteNum
		// engine.Log.Info("查询这个见证人地址的所有投票 555555555555 %d %d", voteNum, one.Vote)
	}
	return communityScore, vssAll
}

// /*
// 	判断是否循环投票
// */
// func (this *WitnessBackup) CheckLoopVote(ws *[]crypto.AddressCoin, witnessAddr *crypto.AddressCoin) (ok bool, wb *BackupWitness) {

// 	v, ok := this.Vote.Load(witnessAddr.B58String())
// 	if !ok {
// 		return false, nil
// 	}
// 	vss := v.(*[]*VoteScore)
// 	//判断是否有重复
// 	for _, one := range *vss {
// 		if bytes.Equal(one, *vs.Witness) {
// 			return true, nil
// 		}
// 	}

// 	//查找这个见证人是否存在
// 	value, ok := this.witnessesMap.Load(vs.Witness.B58String())
// 	if ok {
// 		wb = value.(*BackupWitness)
// 		return false, wb
// 	}

// 	return this.CheckLoopVote(ws, vs.Witness)
// }

type VoteScoreVO struct {
	Witness string //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
	Addr    string //投票人地址
	Payload string //
	Score   uint64 //押金
	Vote    uint64 //获得的投票
}

type VoteScoreV1 struct {
	Witness string //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
	Addr    string //投票人地址
	Payload string //
	Score   uint64 //押金
	Vote    uint64 //获得的投票
	Reward  uint64
	Rate    uint8
}

/*
获取这一时刻的所有社区节点
*/
func (this *WitnessBackup) GetCommunityListSort() []*VoteScoreVO {
	vssVO := make([]*VoteScoreVO, 0)
	this.VoteCommunityList.Range(func(k, v interface{}) bool {
		vsOne := v.(*VoteScore)

		//查询投票总额
		voteNum := uint64(0)
		v, ok := this.Vote.Load(utils.Bytes2string(*vsOne.Addr))
		if ok {
			vss := v.(*[]*VoteScore)
			for _, one := range *vss {
				voteNum = voteNum + one.Scores
			}
		}

		name := FindWitnessName(*vsOne.Addr)

		vsVOOne := VoteScoreVO{
			Witness: vsOne.Witness.B58String(), //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
			Addr:    vsOne.Addr.B58String(),    //投票人地址
			Payload: name,                      //
			Score:   vsOne.Scores,              //押金
			Vote:    voteNum,                   //
		}

		vssVO = append(vssVO, &vsVOOne)
		return true
	})
	return vssVO
}

/*
查询社区节点获得的投票
*/
func (this *WitnessBackup) GetCommunityVote(addr *crypto.AddressCoin) uint64 {
	//查询投票总额
	voteNum := uint64(0)
	v, ok := this.Vote.Load(utils.Bytes2string(*addr))
	if ok {
		vss := v.(*[]*VoteScore)
		for _, one := range *vss {
			voteNum = voteNum + one.Scores
		}
	}
	return voteNum
}

/*
查询一个见证人地址的投票数量 通过合约查询
@return    []*VoteScore    社区投票列表
@return    []*VoteScore    轻节点投票列表
*/
func (this *WitnessBackup) FindScoreNew(addr *crypto.AddressCoin) ([]*VoteScore, []*VoteScore) {
	vssAll := make([]*VoteScore, 0)
	communityScore := make([]*VoteScore, 0)
	//查找社区质押
	communityList := precompiled.GetCommunityListByWit(*addr)
	for _, v := range communityList {
		wit := evmutils.AddressToAddressCoin(v.Wit.Bytes())
		com := evmutils.AddressToAddressCoin(v.Addr.Bytes())
		vs := VoteScore{
			Witness: &wit,
			Addr:    &com,
			Scores:  v.Score.Uint64(),
			Vote:    v.Vote.Uint64(),
		}
		communityScore = append(communityScore, &vs)
		//查找轻节点
		lightList := precompiled.GetLightListByC(com)
		for _, light := range lightList {
			lightAddr := evmutils.AddressToAddressCoin(light.Addr.Bytes())
			vss := VoteScore{
				Witness: &com,                //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
				Addr:    &lightAddr,          //投票人地址
				Vote:    light.Vote.Uint64(), //获得票数
			}
			vssAll = append(vssAll, &vss)
		}
	}
	return communityScore, vssAll
}

// 获取所有见证人
func (this *WitnessBackup) GetAllWitness() []*BackupWitness {
	return this.witnesses
}
