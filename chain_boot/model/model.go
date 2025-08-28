package model

import "web3_gui/chain/mining"

// 详情
type Getinfo struct {
	TotalAmount     uint64                `json:"TotalAmount"`      //发行总量
	Balance         uint64                `json:"balance"`          //可用余额
	BalanceFrozen   uint64                `json:"BalanceFrozen"`    //冻结的余额
	BalanceLockup   uint64                `json:"BalanceLockup"`    //锁仓的余额
	BalanceVote     uint64                `json:"BalanceVote"`      //当前奖励总金额
	Testnet         bool                  `json:"testnet"`          //是否是测试网络
	Blocks          uint64                `json:"blocks"`           //已经同步到的区块高度
	Group           uint64                `json:"group"`            //区块组高度
	StartingBlock   uint64                `json:"StartingBlock"`    //区块开始高度
	StartBlockTime  uint64                `json:"StartBlockTime"`   //创始区块出块时间
	HighestBlock    uint64                `json:"HighestBlock"`     //所链接的节点的最高高度
	CurrentBlock    uint64                `json:"CurrentBlock"`     //已经同步到的区块高度
	PulledStates    uint64                `json:"PulledStates"`     //正在同步的区块高度
	SnapshotHeight  uint64                `json:"SnapshotHeight"`   //快照高度
	BlockTime       uint64                `json:"BlockTime"`        //出块时间
	LightNode       uint64                `json:"LightNode"`        //轻节点押金数量
	CommunityNode   uint64                `json:"CommunityNode"`    //社区节点押金数量
	WitnessNode     uint64                `json:"WitnessNode"`      //见证人押金数量
	NameDepositMin  uint64                `json:"NameDepositMin"`   //域名押金最少金额
	AddrPre         string                `json:"AddrPre"`          //地址前缀
	TokenInfos      []*mining.TokenInfoV0 `json:"TokenInfos"`       //
	SyncBlockFinish bool                  `json:"SyncBlockFinish"`  //同步区块是否完成
	ContractAddress string                `json:"contract_address"` //奖励合约地址
}

type AccountVO struct {
	Index               int      //排序
	Name                string   //名称
	AddrCoin            string   //收款地址
	MainAddrCoin        string   //主地址收款地址
	SubAddrCoins        []string //从地址收款地址
	Value               uint64   //可用余额
	ValueFrozen         uint64   //冻结余额
	ValueLockup         uint64   //
	BalanceVote         uint64   `json:"BalanceVote"` //当前奖励总金额
	AddressFrozenStatus bool     //地址绑定冻结状态
	Type                int      //1=见证人;2=社区节点;3=轻节点;4=什么也不是
}

type NameinfoVO struct {
	Name           string   //域名
	Owner          string   //拥有者
	NetIds         []string //节点地址
	AddrCoins      []string //钱包收款地址
	Height         uint64   //注册区块高度，通过现有高度计算出有效时间
	NameOfValidity uint64   //有效块数量
	Deposit        uint64   //冻结金额
	IsMultName     bool     //是否多签域名
}

/*
见证人信息
*/
type WitnessInfo struct {
	IsCandidate bool   //是否是候选见证人
	IsBackup    bool   //是否是备用见证人
	IsKickOut   bool   //没有按时出块，已经被踢出局，只有退还押金，重新缴纳押金成为候选见证人
	Addr        string //见证人地址
	Payload     string //
	Value       uint64 //见证人押金
}

/*
见证人
*/
type WitnessInfoVO struct {
	Addr            string //见证人地址
	Payload         string //
	Score           uint64 //押金
	Vote            uint64 //投票者押金
	CreateBlockTime int64  //预计出块时间
}

type WitnessInfoV0 struct {
	Vote           uint64          `json:"vote"`             //总票数
	Deposit        uint64          `json:"deposit"`          // 总的质押数
	AddBlockCount  uint64          `json:"add_block_count"`  // 出块数
	AddBlockReward uint64          `json:"add_block_reward"` // 出块奖励
	RewardRatio    float64         `json:"reward_ratio"`     // 奖励比例
	TotalReward    uint64          `json:"total_reward"`     // 累计获得奖励
	FrozenReward   uint64          `json:"frozen_reward"`    //未提现奖励
	CommunityCount uint64          `json:"community_count"`  // 社区数量
	Name           string          `json:"name"`             // 见证者节点名称
	CommunityNode  []CommunityNode `json:"community_node"`   //
	DestroyNum     uint64          `json:"destroy_num"`      //销毁数量
}

type CommunityNode struct {
	Name        string  `json:"name"`         // 名字
	Addr        string  `json:"addr"`         // 地址
	Deposit     uint64  `json:"deposit"`      // 质押量
	Reward      uint64  `json:"reward"`       // 奖励
	LightNum    uint64  `json:"light_num"`    // 轻节点数量(社区人数)
	VoteNum     uint64  `json:"vote_num"`     // 轻节点投票总数
	RewardRatio float64 `json:"reward_ratio"` // 奖励比例
}

type CommunityNodeSort []CommunityNode

func (c CommunityNodeSort) Len() int {
	return len(c)
}
func (c CommunityNodeSort) Less(i, j int) bool {
	return c[i].VoteNum > c[j].VoteNum
}
func (c CommunityNodeSort) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

/*
见证人列表
*/
type WitnessInfoList struct {
	Addr           string  `json:"addr"`             //见证人地址
	Payload        string  `json:"payload"`          //
	Score          uint64  `json:"score"`            //押金
	Vote           uint64  `json:"vote"`             //投票者押金
	AddBlockCount  uint64  `json:"add_block_count"`  // 出块数量
	AddBlockReward uint64  `json:"add_block_reward"` // 出块奖励
	Ratio          float64 `json:"ratio"`            // 奖励比例
}

type WitnessListSort []WitnessInfoList

func (e WitnessListSort) Len() int {
	return len(e)
}

func (e WitnessListSort) Less(i, j int) bool {
	return e[i].Vote > e[j].Vote
}

func (e WitnessListSort) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

/*
社区节点列表
*/
type CommunityList struct {
	Addr           string  `json:"addr"`             //社区节点地址
	WitnessAddr    string  `json:"witness_addr"`     //见证人地址
	Payload        string  `json:"payload"`          //
	Score          uint64  `json:"score"`            //押金
	Vote           uint64  `json:"vote"`             //投票者押金
	AddBlockCount  uint64  `json:"add_block_count"`  // 出块数量
	AddBlockReward uint64  `json:"add_block_reward"` // 出块奖励
	RewardRatio    float64 `json:"reward_ratio"`     // 奖励比例
	DisRatio       float64 `json:"dis_ratio"`        // 分配比例
}

/*
投票详情
*/
type VoteInfoVO struct {
	Txid        string //交易id
	WitnessAddr string //见证人地址
	Value       uint64 //投票数量
	Height      uint64 //区块高度
	AddrSelf    string //自己投票的地址
	Payload     string //
}

type Vinfos struct {
	infos []VoteInfoVO
}

func (this *Vinfos) GetInfos() []VoteInfoVO {
	return this.infos
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

func NewVinfos(infos []VoteInfoVO) Vinfos {
	return Vinfos{
		infos: infos,
	}
}

type CommunityInfoV0 struct {
	Deposit      uint64      `json:"deposit"`       // 总的质押数
	Vote         uint64      `json:"vote"`          //总票数
	LightCount   uint64      `json:"light_count"`   // 投票成员
	RewardRatio  float64     `json:"reward_ratio"`  // 奖励比例
	Reward       uint64      `json:"reward"`        // 累计获得奖励
	WitnessName  string      `json:"witness_name"`  // 见证人名称
	WitnessAddr  string      `json:"witness_addr"`  // 见证人地址
	StartHeight  uint64      `json:"start_height"`  // 成为社区节点的高度
	FrozenReward uint64      `json:"frozen_reward"` // 未提现奖励
	Contract     string      `json:"contract"`      // 提现合约地址
	Name         string      `json:"name"`          // 社区节点名称
	LightNode    []LightNode `json:"light_node"`
}

type LightNode struct {
	Addr            string  `json:"addr"`              // 地址
	Reward          uint64  `json:"reward"`            // 奖励
	RewardRatio     float64 `json:"reward_ratio"`      // 奖励比例
	VoteNum         uint64  `json:"vote_num"`          // 轻节点投票总数
	Deposit         uint64  `json:"deposit"`           // 质押量
	Name            string  `json:"name"`              //轻节点名字
	LastVotedHeight uint64  `json:"last_voted_height"` //轻节点最后投票高度
}

type LightNodeSort []LightNode

func (c LightNodeSort) Len() int {
	return len(c)
}
func (c LightNodeSort) Less(i, j int) bool {
	return c[i].VoteNum > c[j].VoteNum
}
func (c LightNodeSort) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

/*
地址余额，角色，押金，投票给谁，及投票数量
*/
type AddrVoteInfo struct {
	Balance       uint64 `json:"balance"`   //可用余额
	BalanceFrozen uint64 `json:"balance_f"` //锁定余额
	Role          int    `json:"role"`      //角色
	DepositIn     uint64 `json:"depositin"` //押金
	VoteAddr      string `json:"voteaddr"`  //给哪个社区节点地址投票
	VoteIn        uint64 `json:"votein"`    //投票金额
	VoteNum       uint64 `json:"votenum"`   //获得的投票数量
}
