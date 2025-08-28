package config

import (
	"github.com/shopspring/decimal"
	"sync"
)

const (
	Name_Foundation = "foundation" //链上域名：基金会6%
	Name_investor   = "investor"   //链上域名：投资人8%
	Name_team       = "team"       //链上域名：团队14%
)

var (
	Name_store = "store" //链上域名：资源节点72%(资源奖励39%，投票奖励33%)
)

const (
	//团队奖励
	Name_reward_v3_team       = "team"       //链上域名：团队10%
	Name_reward_v3_ecology    = "ecology"    //链上域名：生态8%
	Name_reward_v3_operate    = "operate"    //链上域名：运营-稳定基金4%
	Name_reward_v3_divvy      = "divvy"      //链上域名：运营-分红4%
	Name_reward_v3_foundation = "foundation" //链上域名：基金会12%
)

const (
	Name_reward_v3_witness = "witness" //用于区分见证人与团队角色
)

type RoleReward struct {
	Ratio    float64 //出块奖励总比例
	RoleName string  //角色名称
}

var RoleRewardList = []RoleReward{
	RoleReward{
		Ratio:    0.10,
		RoleName: Name_reward_v3_team,
	},
	RoleReward{
		Ratio:    0.08,
		RoleName: Name_reward_v3_ecology,
	},
	RoleReward{
		Ratio:    0.04,
		RoleName: Name_reward_v3_operate,
	},
	RoleReward{
		Ratio:    0.04,
		RoleName: Name_reward_v3_divvy,
	},
	RoleReward{
		Ratio:    0.12,
		RoleName: Name_reward_v3_foundation,
	},
}

var tmpRoleRewardSet = sync.Map{}

func init() {
	accuracy := 8                       //精度为8
	ratioSum := decimal.NewFromFloat(0) //比例总和
	for _, roleReward := range RoleRewardList {
		ratio := decimal.NewFromFloat(roleReward.Ratio)
		n := int(ratio.Exponent() * -1)
		if n > accuracy {
			panic("role reward ratio accuracy")
		}
		ratioSum = ratioSum.Add(ratio)
		tmpRoleRewardSet.Store(roleReward.RoleName, struct{}{})
	}

	if ratioSum.Cmp(decimal.NewFromFloat(1)) == 1 {
		panic("role reward ratio sum")
	}
}

func IsRoleReward(name string) bool {
	_, ok := tmpRoleRewardSet.Load(name)
	return ok
}
