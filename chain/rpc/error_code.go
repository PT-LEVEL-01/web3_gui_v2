package rpc

import "web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"

const (
	NotEnough                = 5008 //余额不足
	ContentIncorrectFormat   = 5009 //参数格式不正确
	AmountIsZero             = 5010 //转账不能为0
	RuleField                = 5011 //地址角色不正确
	BalanceNotEnough         = 5012 //余额不足
	VoteExist                = 5013 //投票已经存在
	VoteNotOpen              = 5014 //投票功能还未开放
	RewardCountSync          = 5015 //轻节点奖励异步执行中
	CommentOverLengthMax     = 5016 //备注信息字符串超过最大长度
	GasTooLittle             = 5017 //交易手续费太少
	NotDepositOutLight       = 5018 //有投票，不能取消轻节点奖励
	SystemError              = 5019 //系统错误
	NotLightNode             = 5020 //不是轻节点
	NotWitnessNode           = 5021 //不是见证者节点
	NotCommunityNode         = 5022 //不是社区节点
	BalanceTooLittle         = 5023 //金额太小
	PubKeyNotExists          = 5024 //公钥不存在
	ChainNotSync             = 5025 // 链没有初始化
	RewardNotLinked          = 5026 // 奖励没有上链
	RepeatReward             = 5027 // 重复奖励
	DistributeRewardTooEarly = 5028 // 分配奖励太早
	DistributeRatioTooBig    = 5029 //分配比例不能大于100
	ParamError               = 5030 // 参数错误
	NftNotOwner              = 5031 // nft不属于自己
	GasTooBig                = 5032 // gas费用过高
	TestCoinLimit            = 5033 //测试币领取限制
	NotTestChain             = 5034 //不是测试链
	WitnessDepositExist      = 5035 //见证人押金已经存在
	WitnessDepositLess       = 5036 //见证人押金数量不对
	NameExist                = 5037 //域名已经存在
	BalanceTooBig            = 5038 //金额太大
	SysTemError              = 9999 //不是测试链
	JSON_MARSHAL_ERROR       = 24101
	JSON_UNMARSHAL_ERROR     = 24102
	Tx_not_exist             = 24103
)

func RegisterErrorCode() {
	model.RegisterErrcode(NotEnough, "not enough")
	model.RegisterErrcode(ContentIncorrectFormat, "")
	model.RegisterErrcode(AmountIsZero, "amount must gt zero")
	model.RegisterErrcode(RuleField, "")
	model.RegisterErrcode(BalanceNotEnough, "BalanceNotEnough")
	model.RegisterErrcode(VoteExist, "VoteExist")
	model.RegisterErrcode(VoteNotOpen, "VoteNotOpen")
	model.RegisterErrcode(RewardCountSync, "reward sync execution")
	model.RegisterErrcode(CommentOverLengthMax, "comment over length max")
	model.RegisterErrcode(GasTooLittle, "gas too little")
	model.RegisterErrcode(NotDepositOutLight, "Not Deposit Out Light")
	model.RegisterErrcode(SystemError, "system error")
	model.RegisterErrcode(NotLightNode, "not light node")
	model.RegisterErrcode(NotWitnessNode, "not witness node")
	model.RegisterErrcode(NotCommunityNode, "not community node")
	model.RegisterErrcode(BalanceTooLittle, "balance too little")
	model.RegisterErrcode(BalanceTooBig, "balance too big")
	model.RegisterErrcode(PubKeyNotExists, "public key not exists")
	model.RegisterErrcode(ChainNotSync, "chain not sync")
	model.RegisterErrcode(RewardNotLinked, "reward not linked")
	model.RegisterErrcode(RepeatReward, "repeat reward")
	model.RegisterErrcode(DistributeRewardTooEarly, "DistributeRewardTooEarly")
	model.RegisterErrcode(DistributeRatioTooBig, "DistributeRatioTooBig")
	model.RegisterErrcode(NftNotOwner, "nft not owner")
	model.RegisterErrcode(GasTooBig, "gas too big")
	model.RegisterErrcode(ParamError, "param error")
	model.RegisterErrcode(TestCoinLimit, "lock time has not expired. Please try again later")
	model.RegisterErrcode(NotTestChain, "the rpc interface only call in test chain")
	model.RegisterErrcode(SysTemError, "system error")
	model.RegisterErrcode(JSON_MARSHAL_ERROR, "json marshal failed")
	model.RegisterErrcode(JSON_UNMARSHAL_ERROR, "json unmarshal failed")
	model.RegisterErrcode(Tx_not_exist, "tx not exist")

}
