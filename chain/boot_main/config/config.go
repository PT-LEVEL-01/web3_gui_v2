package config

import (
	"time"
	"web3_gui/chain/config"
	"web3_gui/keystore/adapter/derivation"
)

const (
	//团队奖励
	Name_reward_v3_team          = "team"          //链上域名：团队10%
	Name_reward_v3_ecology       = "ecology"       //链上域名：生态8%
	Name_reward_v3_operate       = "operate"       //链上域名：运营8%
	Name_reward_v3_operate_divvy = "operate_divvy" //链上域名：运营分红8%
	Name_reward_v3_foundation    = "foundation"    //链上域名：基金会12%
)

func InitConfig() {

	derivation.CoinType = 98055361

	config.MinCPUCores = uint32(0)         //uint32(3)        //最低cpu核数
	config.MinFreeMemory = uint32(0)       //uint32(8 * 1024) //最低可用内存(MB)
	config.MinFreeDisk = uint32(0)         //uint32(6 * 1024) //最低可用磁盘空间(MB)
	config.MinNetworkBandwidth = uint32(0) //uint32(10)       //最低带宽(MB/s)

	config.Mining_coin_total = 30 * 10000 * 10000 * 1e8 //货币发行总量13亿
	config.Mining_coin_premining = 10000 * 10000 * 1e8  //1 * 10000 * 10000 * 1e8
	config.Mining_block_time = 1 * time.Second          //出块时间，单位：秒
	//主链发奖励间隔,默认600块高度(等于1及时到账),真实交易中是含有每个见证人的奖励的,但奖励会首先放入各自见证人的奖励池中.待时间间隔到了累计的奖励全部到账
	config.Mining_Reward_Interval = uint64(600)
	config.Mining_deposit = uint64(100 * 10000 * 1e8)             //见证人押金最少金额
	config.Mining_vote = uint64(10 * 10000 * 1e8)                 //社区节点投票押金最少金额
	config.Mining_light_min = uint64(1000 * 1e8)                  //轻节点押金最少金额
	config.Max_Community_Count = 1000                             //全网最大社区数量,默认1000个
	config.Max_Light_Count = 1000                                 //每个社区下的最大轻节点数量,默认1000个
	config.Mining_witness_ratio = 90                              //见证人默认分奖比例，对初始块有效
	config.Mining_first_witness_name = []byte("FirstWitness")     //首个见证人名称
	config.RoleRewardList = RoleRewardList                        //
	config.ClacRewardForBlockHeightFun = ClacRewardForBlockHeight //

	config.EnableFreeGas = true       //启用免gas费
	config.DisableCommunityTx = false //禁用社区交易
	config.DisableLightTx = false     //禁用轻节点交易
	config.Enable180DayLock = false   //是否启动180天冻结，线性释放
}

/*
角色奖励
*/
var RoleRewardList = []config.RoleReward{
	config.RoleReward{
		Ratio:    0.10,
		RoleName: Name_reward_v3_team,
	},
	config.RoleReward{
		Ratio:    0.08,
		RoleName: Name_reward_v3_ecology,
	},
	config.RoleReward{
		Ratio:    0.04,
		RoleName: Name_reward_v3_operate,
	},
	config.RoleReward{
		Ratio:    0.04,
		RoleName: Name_reward_v3_operate_divvy,
	},
	config.RoleReward{
		Ratio:    0.12,
		RoleName: Name_reward_v3_foundation,
	},
}

/*
通过区块高度计算区块奖励
第一个区块奖励42枚，
*/
func ClacRewardForBlockHeight(height uint64) uint64 {
	//前1-3个月
	if height < 3*30*24*60*60 {
		return 100 * 1e8
	} else if height < 6*30*24*60*60 {
		//4-6月
		return 30 * 1e8
	} else if height < 9*30*24*60*60 {
		//7-9月
		return 9 * 1e8
	} else if height < 12*30*24*60*60 {
		//10-12月
		return 27000 * 10000
	} else if height < 3*12*30*24*60*60 {
		//第2-3年
		return 18900 * 10000
	} else if height < 5*12*30*24*60*60 {
		//第4-5年
		return 13230 * 10000
	} else if height < 7*12*30*24*60*60 {
		//第6-7年
		return 9260 * 10000
	} else if height < 9*12*30*24*60*60 {
		//第8-9年
		return 6480 * 10000
	} else if height < 11*12*30*24*60*60 {
		//第10-11年
		return 3200 * 10000
	} else if height < 13*12*30*24*60*60 {
		//第12-13年
		return 3200 * 10000
	} else if height < 15*12*30*24*60*60 {
		//第14-15年
		return 3200 * 10000
	} else if height < 17*12*30*24*60*60 {
		//第16-17年
		return 3200 * 10000
	} else if height < 31*12*30*24*60*60 {
		//第18-31年
		return 3200 * 10000
	} else {
		//31年之后
		return 0
	}
}
