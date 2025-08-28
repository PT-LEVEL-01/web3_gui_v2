package mining

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"
	"web3_gui/chain/event"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/utils"
)

type SnapshotChain go_protos.SnapshotChain

// 转换为快照见证人
func (this *SnapshotChain) toSnapshotWitness(in *Witness) *go_protos.SnapshotWitness {
	if in == nil {
		return nil
		return &go_protos.SnapshotWitness{}
	}

	smallGroupHeight := uint64(0)
	if in.Group != nil {
		smallGroupHeight = in.Group.Height
	}

	return &go_protos.SnapshotWitness{
		Addr:           *in.Addr,
		Puk:            in.Puk,
		Block:          this.toSnapshotBlock(in.Block),
		Score:          in.Score,
		CommunityVotes: this.toSnapshotVoteScores(in.CommunityVotes),
		Votes:          this.toSnapshotVoteScores(in.Votes),
		VoteNum:        in.VoteNum,
		//StopMining      chan bool                  `json:"-",copier:"-"` //停止出块命令
		BlockHeight:     in.BlockHeight,
		CreateBlockTime: in.CreateBlockTime,
		// 保存大组见证人的部分信息,方便恢复
		//WitnessBigGroup: this.toSnapshotWitnessBigGroupV2(in),
		CheckIsMining: in.CheckIsMining,
		//SyncBlockOnce   *sync.Once                 `json:"-",copier:"-"` //定时同步区块，只执行一次
		SignExist:        in.SignExist,
		CreateTime:       in.createTime.Unix(),
		SleepTime:        in.sleepTime,
		SmallGroupHeight: smallGroupHeight,
		PtrAddr:          fmt.Sprintf("%p", in),
	}
}

// 转换为投票押金类型
func (this *SnapshotChain) toSnapshotVoteScores(in []*VoteScore) (voteScores []*go_protos.VoteScore) {

	if in == nil || len(in) == 0 {
		return
	}

	for _, v := range in {
		voteScores = append(voteScores, this.toSnapshotVoteScore(v))
	}
	return
}

// 转换为投票押金类型
func (this *SnapshotChain) toSnapshotVoteScore(in *VoteScore) *go_protos.VoteScore {
	if in == nil {
		return nil
	}

	return &go_protos.VoteScore{
		Witness: *in.Witness,
		Addr:    *in.Addr,
		Scores:  in.Scores,
		Vote:    in.Vote,
	}
}

// 转换为投票押金类型
func (this *SnapshotChain) toVoteScores(in []*go_protos.VoteScore) (voteScores []*VoteScore) {

	if in == nil || len(in) == 0 {
		return
	}

	for _, v := range in {
		voteScores = append(voteScores, this.toVoteScore(v))
	}
	return
}

// 转换为投票押金类型
func (this *SnapshotChain) toVoteScore(in *go_protos.VoteScore) *VoteScore {
	if in == nil {
		return nil
	}
	witness := crypto.AddressCoin(in.Witness)
	addr := crypto.AddressCoin(in.Addr)
	return &VoteScore{
		Witness: &witness,
		Addr:    &addr,
		Scores:  in.Scores,
		Vote:    in.Vote,
	}
}

// 转换为见证人
func (this *SnapshotChain) toWitness(in *go_protos.SnapshotWitness) *Witness {
	if in == nil {
		return nil
		return &Witness{}
	}

	addr := crypto.AddressCoin(in.Addr)
	return &Witness{
		Group: &WitnessSmallGroup{
			Height: in.SmallGroupHeight,
		},
		PreWitness:      nil,
		NextWitness:     nil,
		Addr:            &addr,
		Puk:             in.Puk,
		Block:           this.toBlock(in.Block),
		Score:           in.Score,
		CommunityVotes:  this.toVoteScores(in.CommunityVotes),
		Votes:           this.toVoteScores(in.Votes),
		VoteNum:         in.VoteNum,
		StopMining:      make(chan bool, 1),
		BlockHeight:     in.BlockHeight,
		CreateBlockTime: in.CreateBlockTime,
		//WitnessBigGroup: nil,
		CheckIsMining: in.CheckIsMining,
		//syncBlockOnce: &sync.Once{},
		SignExist:  in.SignExist,
		createTime: time.Unix(in.CreateTime, 0),
		sleepTime:  in.SleepTime,
	}
}

// 转换为见证人
// 为未分配见证人使用
func (this *SnapshotChain) toWitness4NotGroup(in *go_protos.SnapshotWitness) *Witness {
	if in == nil {
		return nil
		return &Witness{}
	}

	addr := crypto.AddressCoin(in.Addr)
	return &Witness{
		Group:           nil,
		PreWitness:      nil,
		NextWitness:     nil,
		Addr:            &addr,
		Puk:             in.Puk,
		Block:           nil,
		Score:           in.Score,
		CommunityVotes:  this.toVoteScores(in.CommunityVotes),
		Votes:           this.toVoteScores(in.Votes),
		VoteNum:         in.VoteNum,
		WitnessBigGroup: nil,
		StopMining:      make(chan bool, 1),
		BlockHeight:     in.BlockHeight,
		CreateBlockTime: in.CreateBlockTime,
		CheckIsMining:   in.CheckIsMining,
		SignExist:       in.SignExist,
		createTime:      time.Unix(in.CreateTime, 0),
		sleepTime:       in.SleepTime,
		syncBlockOnce:   nil,
	}
}

// 转换备用见证人
func (this *SnapshotChain) toSnapshotWitnessBackup(in *WitnessBackup) *go_protos.SnapshotWitnessBackup {
	if in == nil {
		return nil
		return &go_protos.SnapshotWitnessBackup{}
	}

	return &go_protos.SnapshotWitnessBackup{
		Witnesses:         this.toSnapshotBackupWitness(in.witnesses),
		WitnessesMap:      this.toSnapshotBackupWitnessMap(in.witnessesMap),
		VoteCommunity:     this.toSnapshotVoteScoreMapArr(in.VoteCommunity),
		VoteCommunityList: this.toSnapshotVoteScoreMap(in.VoteCommunityList),
		Vote:              this.toSnapshotVoteScoreMapArr(in.Vote),
		VoteList:          this.toSnapshotVoteScoreMap(in.VoteList),
		LightNode:         this.toSnapshotVoteScoreMap(in.LightNode),
		Blacklist:         this.syncmap2mapStringUint(in.Blacklist),
	}
}

// 全网交易平均Gas
func (this *SnapshotChain) toSnapshotTxAveGas(in *TxAveGas) *go_protos.TxAveGas {
	if in == nil {
		return nil
		return &go_protos.TxAveGas{}
	}

	return &go_protos.TxAveGas{
		AllGas: in.AllGas[:],
		Index:  in.Index,
	}
}

// 转换map类型
func (this *SnapshotChain) toSnapshotBackupWitnessMap(in sync.Map) map[string]*go_protos.BackupWitness {
	snapshotBackupWitnessMap := make(map[string]*go_protos.BackupWitness)
	in.Range(func(k, v interface{}) bool {
		backupWitness := v.(*BackupWitness)
		key := k.(string)
		snapshotBackupWitnessMap[key] = &go_protos.BackupWitness{
			Addr:    *backupWitness.Addr,
			Puk:     backupWitness.Puk,
			Score:   backupWitness.Score,
			VoteNum: backupWitness.VoteNum,
		}
		return true
	})
	return snapshotBackupWitnessMap
}

// 转换map类型
func (this *SnapshotChain) toSnapshotVoteScoreMap(in sync.Map) map[string]*go_protos.VoteScore {
	snapshotVoteScoreMap := make(map[string]*go_protos.VoteScore)
	in.Range(func(k, v interface{}) bool {
		voteScore := v.(*VoteScore)
		key := k.(string)
		snapshotVoteScoreMap[key] = &go_protos.VoteScore{
			Witness: *voteScore.Witness,
			Addr:    *voteScore.Addr,
			Scores:  voteScore.Scores,
			Vote:    voteScore.Vote,
		}
		return true
	})
	return snapshotVoteScoreMap
}

// 转换map类型
func (this *SnapshotChain) toSnapshotVoteScoreMapArr(in sync.Map) map[string]*go_protos.VoteScoreArr {
	snapshotVoteScoreMap := make(map[string]*go_protos.VoteScoreArr)
	in.Range(func(k, v interface{}) bool {
		voteScoreArr := v.([]*VoteScore)
		key := k.(string)

		var snapshotVoteScores []*go_protos.VoteScore
		for _, vv := range voteScoreArr {
			snapshotVoteScores = append(snapshotVoteScores, &go_protos.VoteScore{
				Witness: *vv.Witness,
				Addr:    *vv.Addr,
				Scores:  vv.Scores,
				Vote:    vv.Vote,
			})
		}

		snapshotVoteScoreMap[key] = &go_protos.VoteScoreArr{
			VoteScoreArr: snapshotVoteScores,
		}
		return true
	})
	return snapshotVoteScoreMap
}

// 转换备用见证人类型
func (this *SnapshotChain) toSnapshotBackupWitness(in []*BackupWitness) (backupWitness []*go_protos.BackupWitness) {
	if in == nil || len(in) == 0 {
		return
	}

	for _, v := range in {
		backupWitness = append(backupWitness, &go_protos.BackupWitness{
			Addr:    *v.Addr,
			Puk:     v.Puk,
			Score:   v.Score,
			VoteNum: v.VoteNum,
		})
	}
	return
}

// 转换备用见证人
func (this *SnapshotChain) toWitnessBackup(in *go_protos.SnapshotWitnessBackup) *WitnessBackup {
	if in == nil {
		return nil
		return &WitnessBackup{}
	}

	return &WitnessBackup{
		chain:             nil,
		lock:              sync.RWMutex{},
		witnesses:         this.toBackupWitness(in.Witnesses),
		witnessesMap:      this.toBackupWitnessSyncMap(in.WitnessesMap),
		VoteCommunity:     this.toVoteScoreSyncMapArr(in.VoteCommunity),
		VoteCommunityList: this.toVoteScoreSyncMap(in.VoteCommunityList),
		Vote:              this.toVoteScoreSyncMapArr(in.Vote),
		VoteList:          this.toVoteScoreSyncMap(in.VoteList),
		LightNode:         this.toVoteScoreSyncMap(in.LightNode),
		Blacklist:         this.mapStringUint2syncmap(in.Blacklist, 64),
	}
}

// 转换全网平均Gas
func (this *SnapshotChain) toTxPayAveGas(in *go_protos.TxAveGas) *TxAveGas {
	if in == nil {
		return &TxAveGas{
			mux:    &sync.RWMutex{},
			AllGas: [100]uint64{},
			Index:  0,
		}
		return nil
	}

	allGas := [100]uint64{}
	for i, _ := range in.AllGas {
		if i >= 100 {
			break
		}
		allGas[i] = in.AllGas[i]
	}
	return &TxAveGas{
		mux:    &sync.RWMutex{},
		AllGas: allGas,
		Index:  in.Index,
	}
}

// 转换map类型
func (this *SnapshotChain) toBackupWitnessSyncMap(in map[string]*go_protos.BackupWitness) sync.Map {
	m := sync.Map{}
	if in == nil || len(in) == 0 {
		return m
	}
	for k, v := range in {
		addr := crypto.AddressCoin(v.Addr)

		m.Store(k, &BackupWitness{
			Addr:    &addr,
			Puk:     v.Puk,
			Score:   v.Score,
			VoteNum: v.VoteNum,
		})

	}
	return m
}

// 转换备用见证人类型
func (this *SnapshotChain) toBackupWitness(in []*go_protos.BackupWitness) (backupWitness []*BackupWitness) {
	if in == nil || len(in) == 0 {
		return
	}
	for _, v := range in {
		addr := crypto.AddressCoin(v.Addr)
		backupWitness = append(backupWitness, &BackupWitness{
			Addr:    &addr,
			Puk:     v.Puk,
			Score:   v.Score,
			VoteNum: v.VoteNum,
		})
	}
	return
}

// 转换map类型
func (this *SnapshotChain) toVoteScoreSyncMap(in map[string]*go_protos.VoteScore) sync.Map {
	m := sync.Map{}
	if in == nil || len(in) == 0 {
		return m
	}
	for k, v := range in {
		witness := crypto.AddressCoin(v.Witness)
		addr := crypto.AddressCoin(v.Addr)
		m.Store(k, &VoteScore{
			Witness: &addr,
			Addr:    &witness,
			Scores:  v.Scores,
			Vote:    v.Vote,
		})
	}

	return m
}

// 转换map类型
func (this *SnapshotChain) toVoteScoreSyncMapArr(in map[string]*go_protos.VoteScoreArr) sync.Map {
	m := sync.Map{}
	if in == nil || len(in) == 0 {
		return m
	}
	for k, v := range in {
		var voteScores []*VoteScore
		for _, vv := range v.VoteScoreArr {
			witness := crypto.AddressCoin(vv.Witness)
			addr := crypto.AddressCoin(vv.Addr)
			voteScores = append(voteScores, &VoteScore{
				Witness: &addr,
				Addr:    &witness,
				Scores:  vv.Scores,
				Vote:    vv.Vote,
			})

		}
		m.Store(k, voteScores)
	}

	return m
}

// 转换为快照块
func (this *SnapshotChain) toSnapshotBlock(in *Block) *go_protos.SnapshotBlock {
	if in == nil {
		return nil
		return &go_protos.SnapshotBlock{}
	}

	return &go_protos.SnapshotBlock{
		Id:         in.Id,         //区块id
		PreBlockID: in.PreBlockID, //前置区块id
		//PreBlock   *SnapshotBlock   //前置区块高度
		//NextBlock  *SnapshotBlock   //下一个区块高度
		//Group   *SnapshotGroup   //所属组
		Height: in.Height, //区块高度
		//Witness *SnapshotWitness //是哪个见证人出的块
		IsCount: in.isCount, //是否被统计过了
		// IdStr      string   //
		// LocalTime time.Time //
	}
}

// 转换为块
func (this *SnapshotChain) toBlock(in *go_protos.SnapshotBlock) *Block {
	if in == nil {
		return nil
		return &Block{}
	}

	return &Block{
		Id:         in.Id,      //区块id
		PreBlockID: nil,        //前置区块id
		PreBlock:   nil,        //前置区块高度
		NextBlock:  nil,        //下一个区块高度
		Group:      nil,        //所属组
		Height:     in.Height,  //区块高度
		Witness:    nil,        //是哪个见证人出的块
		isCount:    in.IsCount, //是否被统计过了
		// IdStr      string   //
		// LocalTime time.Time //
	}
}

// 转换为快照组
func (this *SnapshotChain) toSnapshotGroup(in *Group) *go_protos.SnapshotGroup {
	if in == nil {
		return nil
		return &go_protos.SnapshotGroup{}
	}

	snapshotBlocks := []*go_protos.SnapshotBlock{}
	for _, v := range in.Blocks {
		snapshotBlocks = append(snapshotBlocks, &go_protos.SnapshotBlock{
			Id:         v.Id,         //区块id
			PreBlockID: v.PreBlockID, //前置区块id
			//PreBlock   *SnapshotBlock   //前置区块高度
			//NextBlock  *SnapshotBlock   //下一个区块高度
			//Group   *SnapshotGroup   //所属组
			Height: v.Height, //区块高度
			//Witness *SnapshotWitness //是哪个见证人出的块
			IsCount: v.isCount, //是否被统计过了
			// IdStr      string   //
			// LocalTime time.Time //
		})
	}

	out := &go_protos.SnapshotGroup{
		Height: in.Height,      //组高度
		Blocks: snapshotBlocks, //组中的区块
	}

	return out
}

// 转换为组
func (this *SnapshotChain) toGroup(in *go_protos.SnapshotGroup) *Group {
	if in == nil {
		return nil
		return &Group{}
	}

	blocks := []*Block{}
	for _, b := range in.Blocks {
		blocks = append(blocks, this.toBlock(b))
	}

	out := &Group{
		PreGroup:  nil,
		NextGroup: nil,
		Height:    in.Height, //组高度
		Blocks:    blocks,    //组中的区块
	}

	return out
}

// 转换为快照小组
func (this *SnapshotChain) toSnapshotWitnessSmallGroup(in *WitnessSmallGroup) *go_protos.SnapshotWitnessSmallGroup {
	if in == nil {
		return nil
		return &go_protos.SnapshotWitnessSmallGroup{}
	}

	snapshotWitnesses := []*go_protos.SnapshotWitness{}
	for _, v := range in.Witness {
		snapshotWitnesses = append(snapshotWitnesses, this.toSnapshotWitness(v))
	}

	out := &go_protos.SnapshotWitnessSmallGroup{
		Task:         in.Task,
		Height:       in.Height,
		Witness:      snapshotWitnesses,
		IsBuildGroup: in.IsBuildGroup,
		BlockGroup:   this.toSnapshotGroup(in.BlockGroup),
		Tag:          in.tag,
		IsCount:      in.IsCount,
	}

	return out
}

// 转换为小组
func (this *SnapshotChain) toWitnessSmallGroup(in *go_protos.SnapshotWitnessSmallGroup) *WitnessSmallGroup {
	if in == nil {
		return nil
		return &WitnessSmallGroup{}
	}

	witnesses := []*Witness{}
	for _, wit := range in.Witness {
		witnesses = append(witnesses, this.toWitness(wit))
	}

	out := &WitnessSmallGroup{
		//Task:          in.Task,
		Task:          false, // 全部重新构建时间
		PreGroup:      nil,
		NextGroup:     nil,
		Height:        in.Height,
		Witness:       witnesses,
		BlockGroup:    this.toGroup(in.BlockGroup),
		IsBuildGroup:  in.IsBuildGroup,
		tag:           in.Tag,
		IsCount:       in.IsCount,
		CollationLock: &sync.Mutex{},
	}

	return out
}

// 转换为快照大组
func (this *SnapshotChain) toSnapshotWitnessBigGroup(in *WitnessBigGroup) *go_protos.SnapshotWitnessBigGroup {
	if in == nil {
		return nil
		return &go_protos.SnapshotWitnessBigGroup{}
	}

	snapshotWitnesses := []*go_protos.SnapshotWitness{}
	for _, v := range in.Witnesses {
		snapshotWitnesses = append(snapshotWitnesses, this.toSnapshotWitness(v))
	}

	snapshotWitnessBackupes := []*go_protos.SnapshotWitness{}
	for _, v := range in.WitnessBackup {
		snapshotWitnessBackupes = append(snapshotWitnessBackupes, this.toSnapshotWitness(v))
	}

	out := &go_protos.SnapshotWitnessBigGroup{
		Keys:          []string{},
		Witnesses:     snapshotWitnesses,
		WitnessBackup: snapshotWitnessBackupes,
	}

	return out
}

// 转换为快照大组,只包含见证人少量信息，避免递归
func (this *SnapshotChain) toSnapshotWitnessBigGroupV2(in *Witness) *go_protos.SnapshotWitnessBigGroup {
	if in == nil {
		return nil
	}

	biggroup := in.WitnessBigGroup

	snapshotWitnesses := []*go_protos.SnapshotWitness{}
	for i, _ := range biggroup.Witnesses {
		//if biggroup.Witnesses[i].CreateBlockTime == 0 {
		//	continue
		//}
		in := biggroup.Witnesses[i]
		snapshotWitnesses = append(snapshotWitnesses, &go_protos.SnapshotWitness{
			Addr:           *in.Addr,
			Puk:            in.Puk,
			Block:          nil,
			Score:          in.Score,
			CommunityVotes: this.toSnapshotVoteScores(in.CommunityVotes),
			Votes:          this.toSnapshotVoteScores(in.Votes),
			VoteNum:        in.VoteNum,
			//StopMining      chan bool                  `json:"-",copier:"-"` //停止出块命令
			BlockHeight:     in.BlockHeight,
			CreateBlockTime: in.CreateBlockTime,
			// 保存大组见证人的部分信息,方便恢复
			//WitnessBigGroup: this.toSnapshotWitnessBigGroupV2(in),
			CheckIsMining: in.CheckIsMining,
			//SyncBlockOnce   *sync.Once                 `json:"-",copier:"-"` //定时同步区块，只执行一次
			SignExist:  in.SignExist,
			CreateTime: in.createTime.Unix(),
			PtrAddr:    fmt.Sprintf("%p", in),
		})
	}

	snapshotWitnessBackupes := []*go_protos.SnapshotWitness{}
	for i, _ := range biggroup.WitnessBackup {
		//if biggroup.WitnessBackup[i].CreateBlockTime == 0 {
		//	continue
		//}
		snapshotWitnessBackupes = append(snapshotWitnessBackupes, &go_protos.SnapshotWitness{
			Addr:            *biggroup.WitnessBackup[i].Addr,
			CreateBlockTime: biggroup.WitnessBackup[i].CreateBlockTime,
			SleepTime:       1, //仅用于调试标记该见证人未被重复赋值
		})
	}

	out := &go_protos.SnapshotWitnessBigGroup{
		Keys:          []string{this.witnessKey(in)},
		Witnesses:     snapshotWitnesses,
		WitnessBackup: snapshotWitnessBackupes,
	}

	return out
}

// 转换为大组
func (this *SnapshotChain) toWitnessBigGroup(in *go_protos.SnapshotWitnessBigGroup) *WitnessBigGroup {
	if in == nil {
		return nil
		return &WitnessBigGroup{}
	}

	witnesses := []*Witness{}
	for _, v := range in.Witnesses {
		witnesses = append(witnesses, this.toWitness(v))
	}

	witnessBackupes := []*Witness{}
	for _, v := range in.WitnessBackup {
		witnessBackupes = append(witnessBackupes, this.toWitness(v))
	}

	out := &WitnessBigGroup{
		Witnesses:     witnesses,
		WitnessBackup: witnessBackupes,
	}

	return out
}

// 余额管理
func (this *SnapshotChain) toSnapshotBalanceManager(in *BalanceManager) *go_protos.SnapshotBalanceManager {
	if in == nil {
		return nil
		return &go_protos.SnapshotBalanceManager{}
	}

	out := &go_protos.SnapshotBalanceManager{
		CountTotal:  in.countTotal,
		NodeWitness: this.toSnapshotTxItem(in.nodeWitness),
		//DepositWitness:   this.toSnapshotDepositMap(*in.depositWitness),
		//DepositCommunity: this.toSnapshotDepositMap(*in.depositCommunity),
		//DepositLight:  this.toSnapshotDepositMap(*in.depositLight),
		//DepositVote:   this.toSnapshotDepositMap(*in.depositVote),
		//CommunityVote: this.syncmap2mapStringUint(*in.communityVote),
		WitnessBackup: this.toSnapshotWitnessBackup(in.witnessBackup),
		//txManager     :nil,
		OtherDeposit:   this.toSnapshotOtherDeposit(in.otherDeposit),
		AddBlockNum:    this.syncmap2mapStringUint(*in.addBlockNum),
		AddBlockReward: this.syncmap2mapStringUint(*in.addBlockReward),
		BlockIndex:     this.syncmap2mapStringUint(*in.blockIndex),
		// cacheTxLockLo:nil,
		//cacheTxlock  :nil,
		//CacheTxlock: this.syncmap2map(in.cacheTxlock),
		//eventBus     e:nil,
		LastDistributeHeight: in.lastDistributeHeight,
		//WitnessRatio:         this.syncmap2mapStringUint(*in.witnessRatio),
		//WitnessVote:          this.syncmap2mapStringUint(*in.witnessVote),
		//WitnessRewardPool:    this.syncmapBigInt2mapStringBytes(*in.witnessRewardPool),
		//CommunityRewardPool:  this.syncmapBigInt2mapStringBytes(*in.communityRewardPool),
		//NameNetRewardPool:    this.syncmapBigInt2mapStringBytes(*in.nameNetRewardPool),
		//WitnessMapCommunitys: this.syncmap2mapStringAddress(*in.witnessMapCommunitys),
		//CommunityMapLights:   this.syncmap2mapStringAddress(*in.communityMapLights),
		TxAveGas: this.toSnapshotTxAveGas(in.txAveGas),
		//LastVoteOp:           this.syncmap2mapStringUint(*in.lastVoteOp),
		//AddrReward:           this.syncmapBigInt2mapStringBytes(*in.addrReward),
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		out.DepositWitness = this.toSnapshotDepositMap(*in.depositWitness)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.DepositCommunity = this.toSnapshotDepositMap(*in.depositCommunity)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.DepositLight = this.toSnapshotDepositMap(*in.depositLight)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.DepositVote = this.toSnapshotDepositMap(*in.depositVote)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.CommunityVote = this.syncmap2mapStringUint(*in.communityVote)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.LastVoteOp = this.syncmap2mapStringUint(*in.lastVoteOp)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.AddrReward = this.syncmapBigInt2mapStringBytes(*in.addrReward)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.DepositFreeGas = this.syncmap2mapStringDepositFreeGas(*in.freeGasAddrSet)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.WitnessRatio = this.syncmap2mapStringUint(*in.witnessRatio)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.WitnessVote = this.syncmap2mapStringUint(*in.witnessVote)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.WitnessRewardPool = this.syncmapBigInt2mapStringBytes(*in.witnessRewardPool)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.CommunityRewardPool = this.syncmapBigInt2mapStringBytes(*in.communityRewardPool)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.NameNetRewardPool = this.syncmapBigInt2mapStringBytes(*in.nameNetRewardPool)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.WitnessMapCommunitys = this.syncmap2mapStringAddress(*in.witnessMapCommunitys)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		out.CommunityMapLights = this.syncmap2mapStringAddress(*in.communityMapLights)
		wg.Done()
	}()

	wg.Wait()

	return out
}

// 转换交易类型
func (this *SnapshotChain) toSnapshotTxItem(in *TxItem) *go_protos.TxItem {
	if in == nil {
		return nil
		return &go_protos.TxItem{}
	}

	return &go_protos.TxItem{
		Addr:         *in.Addr,
		Value:        in.Value,
		VoteType:     uint64(in.VoteType),
		Name:         in.Name,
		LockupHeight: in.LockupHeight,
	}
}

// 转换质押投票类型
func (this *SnapshotChain) toSnapshotDeposit(in *DepositInfo) *go_protos.DepositInfo {
	if in == nil {
		return nil
		return &go_protos.DepositInfo{}
	}

	return &go_protos.DepositInfo{
		WitnessAddr: in.WitnessAddr,
		SelfAddr:    in.SelfAddr,
		Value:       in.Value,
		Name:        in.Name,
		Height:      in.Height,
	}
}

// 转换质押投票类型
func (this *SnapshotChain) toDeposit(in *go_protos.DepositInfo) *DepositInfo {
	if in == nil {
		return nil
		return &DepositInfo{}
	}

	return &DepositInfo{
		WitnessAddr: in.WitnessAddr,
		SelfAddr:    in.SelfAddr,
		Value:       in.Value,
		Name:        in.Name,
		Height:      in.Height,
	}
}

// 转换质押投票类型
func (this *SnapshotChain) toSnapshotDepositMap(in sync.Map) map[string]*go_protos.DepositInfo {
	snapshotDepositMap := make(map[string]*go_protos.DepositInfo)
	in.Range(func(k, v interface{}) bool {
		key := k.(string)
		if obj, ok := v.(*DepositInfo); ok {
			snapshotDepositMap[key] = this.toSnapshotDeposit(obj)
			return true
		}
		return true
	})
	return snapshotDepositMap
}

// 转换map类型
func (this *SnapshotChain) toSnapshotTxItemMap(in sync.Map) map[string]*go_protos.TxItem {
	snapshotTxItemMap := make(map[string]*go_protos.TxItem)
	in.Range(func(k, v interface{}) bool {
		key := k.(string)
		if obj, ok := v.(*TxItem); ok {
			snapshotTxItemMap[key] = this.toSnapshotTxItem(obj)
			return true
		}
		return true
	})
	return snapshotTxItemMap
}

// 余额管理
func (this *SnapshotChain) toBalanceManager(in *go_protos.SnapshotBalanceManager) *BalanceManager {
	if in == nil {
		return &BalanceManager{}
	}

	depositCommunity := this.toDespositSyncMap(in.DepositCommunity)
	depositLight := this.toDespositSyncMap(in.DepositLight)
	depositVote := this.toDespositSyncMap(in.DepositVote)
	communityVote := this.mapStringUint2syncmap(in.CommunityVote, 64)
	//otherDeposit := this.map2syncmap(in.OtherDeposit)
	addBlockNum := this.mapStringUint2syncmap(in.AddBlockNum, 64)
	addBlockReward := this.mapStringUint2syncmap(in.AddBlockReward, 64)
	blockIndex := this.mapStringUint2syncmap(in.BlockIndex, 64)
	//cacheTxlock := this.map2syncmap(in.CacheTxlock)
	witnessRatio := this.mapStringUint2syncmap(in.WitnessRatio, 16)
	witnessVote := this.mapStringUint2syncmap(in.WitnessVote, 64)
	witnessRewardPool := this.mapStringBytes2syncmapBigInt(in.WitnessRewardPool)
	communityRewardPool := this.mapStringBytes2syncmapBigInt(in.CommunityRewardPool)
	nameNetRewardPool := this.mapStringBytes2syncmapBigInt(in.NameNetRewardPool)
	witnessMapCommunitys := this.mapStringAddress2syncmap(in.WitnessMapCommunitys)
	communityMapLights := this.mapStringAddress2syncmap(in.CommunityMapLights)
	txAveGas := this.toTxPayAveGas(in.TxAveGas)
	lastVoteOp := this.mapStringUint2syncmap(in.LastVoteOp, 64)
	addrReward := this.mapStringBytes2syncmapBigInt(in.AddrReward)
	freeGasAddrSet := this.toDespositFreeGasSyncMap(in.DepositFreeGas)
	depositWitness := this.toDespositSyncMap(in.DepositWitness)
	otherDeposit := this.toOtherDepositSyncMap(in.OtherDeposit)

	//拷贝快照给其它节点用时,该信息需要重新更新为自己的信息
	nodeWitness := this.toTxItem(in.NodeWitness)
	coinAddr := Area.Keystore.GetCoinbase()
	if item, ok := depositWitness.Load(utils.Bytes2string(coinAddr.Addr)); ok {
		//转为自己的见证人快照
		info := item.(*DepositInfo)
		nodeWitness.Addr = &info.SelfAddr
		nodeWitness.Value = info.Value
		nodeWitness.Name = info.Name
	} else {
		//转为自己的普通快照
		nodeWitness = nil
	}

	out := &BalanceManager{
		countTotal:    in.CountTotal,
		chain:         nil,
		syncBlockHead: make(chan *BlockHeadVO, 1),
		//nodeWitness:      this.toTxItem(in.NodeWitness),
		nodeWitness:      nodeWitness,
		depositWitness:   &depositWitness,
		depositCommunity: &depositCommunity,
		depositLight:     &depositLight,
		depositVote:      &depositVote,
		communityVote:    &communityVote,
		witnessBackup:    this.toWitnessBackup(in.WitnessBackup),
		txManager:        nil,
		otherDeposit:     otherDeposit,
		addBlockNum:      &addBlockNum,
		addBlockReward:   &addBlockReward,
		blockIndex:       &blockIndex,
		// cacheTxLockLo:nil,
		cacheTxlock: &sync.Map{},
		//cacheTxlock: &cacheTxlock,
		eventBus:             event.NewEventBus(),
		witnessRatio:         &witnessRatio,
		witnessVote:          &witnessVote,
		lastDistributeHeight: in.LastDistributeHeight,
		witnessRewardPool:    &witnessRewardPool,
		communityRewardPool:  &communityRewardPool,
		nameNetRewardPool:    &nameNetRewardPool,
		witnessMapCommunitys: &witnessMapCommunitys,
		communityMapLights:   &communityMapLights,
		txAveGas:             txAveGas,
		lastVoteOp:           &lastVoteOp,
		addrReward:           &addrReward,
		freeGasAddrSet:       freeGasAddrSet,
		tmpCacheContract:     new(sync.Map),
	}

	return out
}

// 转换交易类型
func (this *SnapshotChain) toTxItem(in *go_protos.TxItem) *TxItem {
	if in == nil {
		return &TxItem{}
	}
	addr := crypto.AddressCoin(in.Addr)
	return &TxItem{
		Addr:         &addr,
		Value:        in.Value,
		VoteType:     uint16(in.VoteType),
		LockupHeight: in.LockupHeight,
	}
}

// 转换map类型
func (this *SnapshotChain) toTxItemSyncMap(in map[string]*go_protos.TxItem) sync.Map {
	m := sync.Map{}
	if in == nil || len(in) == 0 {
		return m
	}

	for k, v := range in {
		m.Store(k, this.toTxItem(v))
	}

	return m
}

// 转换map类型
func (this *SnapshotChain) toDespositSyncMap(in map[string]*go_protos.DepositInfo) sync.Map {
	m := sync.Map{}
	if in == nil || len(in) == 0 {
		return m
	}

	for k, v := range in {
		m.Store(k, this.toDeposit(v))
	}

	return m
}

// 转换map类型
func (this *SnapshotChain) toDespositFreeGasSyncMap(in map[string]*go_protos.DepositFreeGas) *sync.Map {
	m := new(sync.Map)
	if in == nil || len(in) == 0 {
		return m
	}

	for k, v := range in {
		m.Store(k, &DepositFreeGasItem{
			Owner:             v.Owner,
			ContractAddresses: v.ContractAddresses,
			Deposit:           v.Deposit,
			LimitHeight:       v.LimitHeight,
			LimitCount:        v.LimitCount,
		})
	}

	return m
}

// 转换map类型
func (this *SnapshotChain) syncmap2mapStringUint(in sync.Map) map[string]uint64 {
	mapStringUint64 := make(map[string]uint64)
	in.Range(func(k, v interface{}) bool {
		num, ok := v.(uint64)
		if !ok {
			num = uint64(v.(uint16))
		}

		mapStringUint64[k.(string)] = num
		return true
	})
	return mapStringUint64
}

// 转换map类型
func (this *SnapshotChain) syncmap2mapStringDepositFreeGas(in sync.Map) map[string]*go_protos.DepositFreeGas {
	item := make(map[string]*go_protos.DepositFreeGas)
	in.Range(func(k, v interface{}) bool {
		if val, ok := v.(*DepositFreeGasItem); ok {
			item[k.(string)] = &go_protos.DepositFreeGas{
				Owner:             val.Owner,
				ContractAddresses: val.ContractAddresses,
				Deposit:           val.Deposit,
				LimitHeight:       val.LimitHeight,
				LimitCount:        val.LimitCount,
			}
		}
		return true
	})
	return item
}

// map转sync.map
func (this *SnapshotChain) mapStringUint2syncmap(src map[string]uint64, t int) sync.Map {
	out := sync.Map{}
	for k, v := range src {
		if t == 16 {
			out.Store(k, uint16(v))
		} else {
			out.Store(k, v)
		}

	}
	return out
}

// 转换map类型
func (this *SnapshotChain) syncmapBigInt2mapStringBytes(in sync.Map) map[string][]byte {
	mapStringBigInt := make(map[string][]byte)
	in.Range(func(k, v interface{}) bool {
		mapStringBigInt[k.(string)] = v.(*big.Int).Bytes()
		return true
	})
	return mapStringBigInt
}

// map转sync.map
func (this *SnapshotChain) mapStringBytes2syncmapBigInt(src map[string][]byte) sync.Map {
	out := sync.Map{}
	for k, v := range src {
		out.Store(k, new(big.Int).SetBytes(v))
	}
	return out
}

// 转换map类型,key:string,value:[]crypto.AddressCoin
func (this *SnapshotChain) syncmap2mapStringAddress(in sync.Map) map[string]*go_protos.Addresses {
	mapStringAddress := make(map[string]*go_protos.Addresses)
	in.Range(func(k, v interface{}) bool {
		vals := v.([]crypto.AddressCoin)
		addrs := make([][]byte, len(vals))
		for i, _ := range vals {
			addrs[i] = vals[i]
		}
		mapStringAddress[k.(string)] = &go_protos.Addresses{
			Addresses: addrs,
		}
		return true
	})
	return mapStringAddress
}

// 转换map类型,key:string,value:[]crypto.AddressCoin
func (this *SnapshotChain) mapStringAddress2syncmap(in map[string]*go_protos.Addresses) sync.Map {
	out := sync.Map{}
	for k, v := range in {
		addrs := []crypto.AddressCoin{}
		for i, _ := range v.Addresses {
			addrs = append(addrs, v.Addresses[i])
		}
		out.Store(k, addrs)
	}
	return out
}

// 转换其它质押投票类型 sync.Map 嵌套 sync.Map
func (this *SnapshotChain) toOtherDepositSyncMap(in map[int64][]byte) *sync.Map {
	out := &sync.Map{}
	for k, v := range in {
		itemmap := make(map[string]*TxItem)
		if err := json.Unmarshal(v, &itemmap); err != nil {
			continue
		}

		item := &sync.Map{}
		for k2, v2 := range itemmap {
			item.Store(k2, v2)
		}
		out.Store(int(k), item)
	}
	return out
}

// 转换快照类型 sync.Map 嵌套 sync.Map
func (this *SnapshotChain) toSnapshotOtherDeposit(in *sync.Map) map[int64][]byte {
	if in == nil {
		return nil
	}

	mapStringSyncMap := make(map[int64][]byte)
	in.Range(func(k, v interface{}) bool {
		if item, ok := v.(*sync.Map); ok {
			itemmap := make(map[string]*TxItem)
			item.Range(func(key, value any) bool {
				if keyStr, ok := key.(string); ok {
					itemmap[keyStr] = value.(*TxItem)
				}
				return true
			})
			if b, err := json.Marshal(itemmap); err == nil {
				if valInt, ok := k.(int); ok {
					mapStringSyncMap[int64(valInt)] = b
				}
			}
		}
		return true
	})
	return mapStringSyncMap
}

// 生成一个见证人的唯一键
func (this *SnapshotChain) witnessKey(in *Witness) string {
	return fmt.Sprintf("%s_%d", in.Addr.B58String(), in.CreateBlockTime)
}
