package hive

import (
	"math/big"
	"path/filepath"
	"time"
	"web3_gui/chain/config"
	gconfig "web3_gui/config"
	"web3_gui/keystore/v1/derivation"
	"web3_gui/keystore/v2"
	kconfig "web3_gui/keystore/v2/config"
)

func InitConfig() {
	//derivation.CoinType = 0x800055aa                     //
	config.AddrPre = "HIVE"          //
	derivation.CoinType = 0x800055aa //
	keystore.SetCoinTypeCustom(0x800055aa, config.AddrPre)
	keystore.SetCoinTypeCustomDH(0x800055aa + 1)
	kconfig.ShortAddress = false //使用长地址

	gconfig.AreaName_str = "hive_net"                                               //
	config.Path_config = "config.json"                                              //配置文件名称
	config.KeystoreFileAbsPath = filepath.Join(config.Path_configDir, "wallet.bin") //密钥文件存放地址
	gconfig.KeystoreFileAbsPath = config.KeystoreFileAbsPath                        //
	config.DB_path = filepath.Join("BlockChainData")                                //密钥文件存放地址
	config.EnableStartInputPassword = true                                          //是否启用命令行界面输入密码
	config.Init_LocalPort = 25331                                                   //

	config.Mining_block_time = 10 * time.Second                          //出块时间，单位：秒
	config.Mining_coin_total = 1077600000 * 1e8                          //货币发行总量10亿7760万
	config.Mining_coin_premining = 1200 * 10000 * 1e8                    //预挖量 1200万
	config.Mining_deposit = 10 * 10000 * 1e8                             //见证人质押金额
	config.Mining_vote = uint64(10000 * 1e8)                             //社区节点投票押金金额
	config.Mining_light_min = uint64(10 * 1e8)                           //轻节点押金金额
	config.Witness_backup_max = 31                                       //最多见证人数量
	config.Witness_backup_reward_max = 99                                //有奖励的最大见证人数量
	config.Mining_Reward_Interval = 1                                    //主链发奖励间隔
	config.CancelVote_Interval = 100000                                  //社区取消质押/轻节点取消投票间隔,默认100000块高度，1天多
	config.Wallet_tx_gas_min = uint64(1e7)                               //交易手续费最低
	config.Max_Community_Count = 10000                                   //全网最大社区数量,默认100个
	config.Max_Light_Count = 10000                                       //每个社区下的最大轻节点数量,默认200个
	config.Mining_witness_ratio = 90                                     //见证人默认分奖比例，对初始块有效
	config.Mining_first_witness_name = []byte("Hive Network Foundation") //首个见证人名称

	config.RoleRewardList = RoleRewardList                        //
	config.ClacRewardForBlockHeightFun = ClacRewardForBlockHeight //

	config.EnableFreeGas = false      //启用免gas费
	config.DisableCommunityTx = false //禁用社区交易
	config.DisableLightTx = false     //禁用轻节点交易
	config.Enable180DayLock = true    //是否启动180天冻结，线性释放

}

/*
Name_Foundation = "foundation" //链上域名：基金会6%
Name_investor   = "investor"   //链上域名：投资人8%
Name_team       = "team"       //链上域名：团队14%
*/
var RoleRewardList = []config.RoleReward{
	config.RoleReward{
		Ratio:    0.06,
		RoleName: "foundation",
	},
	config.RoleReward{
		Ratio:    0.08,
		RoleName: "investor",
	},
	config.RoleReward{
		Ratio:    0.14,
		RoleName: "team",
	},
}

/*
通过区块高度计算区块奖励
第一个区块奖励286枚，
1到30000高度，每增加一个块加0.0193枚/块；
30001到60000高度，每增加一个块加0.0238枚/块；
60001到90000高度，每增加一个块加0.0268枚/块；
90001到120000高度，每增加一个块加0.0319枚/块；
120001到184999高度，每增加一个块加0.0349枚/块；
185000减产10%；
185000开始每增加10472高度减产10%。
*/
func ClacRewardForBlockHeight(height uint64) uint64 {
	// height += BlockRewardHeightOffset
	// def calcr_new3(height):
	//    res = 286.0
	//    if height < 490000*31+1:
	//        res = (res+(int(height/31))*0.00347)/31
	//    else:
	//        base = res + 490000 * 0.00347
	//        res = 0.95*base
	//        res = res*math.pow( 0.95, int((height - 490000*31)/(20000*31)) )/31
	//        res = max(res, 0)

	//    return res

	// sum = 0
	// for i in range(2110001*31):
	//    sum = sum + calcr_new3(i)

	// #sum = 1299664262
	// print("sum=", sum)

	//-----------------
	// ClacRewardForBlockHeightBase(height)

	// engine.Log.Info("====================================")

	heightBig := big.NewInt(int64(height))
	num31 := big.NewInt(31)
	num347 := big.NewInt(347000)
	res := big.NewInt(28600000000)
	if height < 490000*31+1 {
		temp := new(big.Int).Div(heightBig, num31)
		temp = new(big.Int).Mul(temp, num347)
		temp = new(big.Int).Add(res, temp)
		temp = new(big.Int).Div(temp, num31)
		return temp.Uint64()
	} else {
		num49 := big.NewInt(490000)
		num95 := big.NewInt(95)
		num100 := big.NewInt(100)
		num2 := big.NewInt(20000)
		base := new(big.Int).Mul(num49, num347)
		base = new(big.Int).Add(res, base)
		res = new(big.Int).Mul(num95, base)

		// engine.Log.Info("2222222222 base = %v res = %v", base.Uint64(), res.Uint64())
		//res*math.pow( 0.95, int((height - 490000*31)/(20000*31)) )/31
		temp1 := new(big.Int).Mul(num49, num31)
		temp1 = new(big.Int).Sub(heightBig, temp1)
		temp2 := new(big.Int).Mul(num2, num31)
		temp := new(big.Int).Div(temp1, temp2)
		// engine.Log.Info("222222222 temp = %v", temp)

		div1 := new(big.Int).Exp(num95, temp, nil)
		div2 := new(big.Int).Exp(num100, temp, nil)

		// engine.Log.Info("div1 = %v div2 = %v", div1, div2)

		res = new(big.Int).Mul(res, div1)
		res = new(big.Int).Div(res, num100)
		temp = new(big.Int).Div(res, div2)

		// engine.Log.Info("div1 = %v div2 = %v", div1, div2)

		// engine.Log.Info("temp = %v", temp.Uint64())
		temp = new(big.Int).Div(temp, num31)

		// engine.Log.Info("22222222222 temp =

		// engine.Log.Info("22222222222 temp = %v", temp)

		if temp.Cmp(big.NewInt(0)) == -1 {
			return 0
		}

		return temp.Uint64()
	}

}
