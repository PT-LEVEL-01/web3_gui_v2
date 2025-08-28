package test

import (
	"path/filepath"
	"time"
	"web3_gui/chain/config"
	gconfig "web3_gui/config"
	"web3_gui/keystore/v1/derivation"
	"web3_gui/keystore/v2"
	kv2config "web3_gui/keystore/v2/config"
)

const (
	Path_config_old   = "conf/config.json"
	Path_keystore_old = "conf/keystore.key"
)

/*
Start 区块链服务启动
名    字：THREE
出块逻辑：5秒出一次块，每天17280块，每块0.7
增    产：无
减    产：每3110400块减产10%，减产5次后不在减产
分配逻辑：每次奖励的10%分配给基金会，30%分配给星级节点，60%分配给用户节点，（星级节点共计4种类型（一星节点、二星节点、三星节点、超级节点），

	成为星级节点需要自身质押一定数量代币+自身算力，自身算力来源于自身质押+推广）

手续费分配：分配给超级节点
普通用户质押：需要先质押10代币，成为轻节点
*/
func InitConfig() {
	config.AddrPre = "TEST"                              //
	derivation.CoinType = kv2config.ZeroQuote + 98055361 //
	keystore.SetCoinTypeCustom(kv2config.ZeroQuote+98055361, config.AddrPre)
	keystore.SetCoinTypeCustomDH(kv2config.ZeroQuote + 98055361 + 1)
	//kv2config.CoinType_coin = 98055361                  //
	//kv2config.CoinType_dh = kv2config.CoinType_coin + 1 //

	gconfig.AreaName_str = "test_net"                                               //
	config.Path_config = "config.json"                                              //配置文件名称
	config.KeystoreFileAbsPath = filepath.Join(config.Path_configDir, "wallet.bin") //密钥文件存放地址
	gconfig.KeystoreFileAbsPath = config.KeystoreFileAbsPath                        //
	config.DB_path = filepath.Join("BlockChainData")                                //密钥文件存放地址
	config.EnableStartInputPassword = true                                          //是否启用命令行界面输入密码
	config.Init_LocalPort = 25331                                                   //

	config.Mining_block_time = 1 * time.Second       //出块时间，单位：秒
	config.Mining_coin_total = 3690 * 10000 * 1e8    //货币发行总量3690万
	config.Mining_coin_premining = 100 * 10000 * 1e8 //预挖量 100万
	config.Mining_vote = uint64(10000 * 1e8)         //社区节点押金
	config.Mining_light_min = uint64(10 * 1e8)       //轻节点押金
	config.Witness_backup_max = 31                   //最多见证人数量
	config.Witness_backup_reward_max = 99            //有奖励的最大见证人数量
	config.Mining_Reward_Interval = 1                //主链发奖励间隔
	config.CancelVote_Interval = 10                  //社区取消质押/轻节点取消投票间隔,默认100000块高度，1天多
	config.Wallet_tx_gas_min = uint64(1e7)           //交易手续费最低
	config.Max_Community_Count = 10000               //全网最大社区数量,默认100个
	config.Max_Light_Count = 10000                   //每个社区下的最大轻节点数量,默认200个
	config.Mining_witness_ratio = 90                 //见证人默认分奖比例，对初始块有效
	//config.Mining_first_witness_name = []byte("Hive Network Foundation") //首个见证人名称

	config.RoleRewardList = RoleRewardList                        //
	config.ClacRewardForBlockHeightFun = ClacRewardForBlockHeight //

	config.EnableFreeGas = false      //启用免gas费
	config.DisableCommunityTx = false //禁用社区交易
	config.DisableLightTx = false     //禁用轻节点交易
	config.Enable180DayLock = false   //是否启动180天冻结，线性释放
}

/*
每次奖励的10%分配给基金会，30%分配给星级节点，60%分配给用户节点，（星级节点共计4种类型（一星节点、二星节点、三星节点、超级节点），
成为星级节点需要自身质押一定数量代币+自身算力，自身算力来源于自身质押+推广）
Name_Foundation = "foundation" //链上域名：基金会10%
Name_investor   = "investor"   //链上域名：投资人8%
*/
var RoleRewardList = []config.RoleReward{
	config.RoleReward{
		Ratio:    1,
		RoleName: "root",
	},
}

/*
通过区块高度计算区块奖励
第一个区块奖励0.7枚，
每3110400块减产10%，减产5次后不在减产
周期1挖出总量:2177280.00000000
周期2挖出总量:1959552.00000000
周期3挖出总量:1763596.80000000
周期4挖出总量:1587237.12000000
周期5挖出总量:1428513.40800000
周期6挖出总量:1285662.06720000
总共：10201841.39520000
*/
func ClacRewardForBlockHeight(height uint64) uint64 {
	if height > 83253208 {
		return 0
	}
	fristReward := uint64(7e7)
	rand := height / 3110400
	if rand > 5 {
		rand = 5
	}
	for range rand {
		fristReward = uint64(float64(fristReward) * 0.9)
	}
	return fristReward
}
