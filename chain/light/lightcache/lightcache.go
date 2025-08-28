/*
light 轻节点缓存
轻节点提供的SDK方法中,有很多方法都需要运行虚拟机获取数据,因此在db里面缓存一份相同的数据,SDK方法直接查询db缓存数据
缓存的颗粒度取决于调用虚拟机的方法.
首选查询db缓存数据是否过期,过期则执行虚拟机获取最新数据,同时更新缓存数据和写入db.
重启加载db数据至缓存
*/
package lightcache

import (
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"math/big"
	"strings"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/evm/abi"
	"web3_gui/utils"
)

// 监听数据过期事件
var cacheGroupEvent = sync.Map{}

// 轻节点缓存
// 采用lru算法,缓存256条数据，1h过期
var lightCache = lru.NewLRU[string, any](256, nil, time.Hour)

// 轻节点缓存初始化
func Setup(rewardABI string) {
	abiObj, _ := abi.JSON(strings.NewReader(rewardABI))
	// 缓存方法
	getCommunityList := abiObj.Methods["getCommunityList"]
	//allCReward := abiObj.Methods["allCReward"]
	//getRate := abiObj.Methods["getRate"]
	getLightListByC := abiObj.Methods["getLightListByC"]
	//communityRewardPool := abiObj.Methods["communityRewardPool"]
	//queryAllReward := abiObj.Methods["queryAllReward"]
	getLightList := abiObj.Methods["getLightList"]
	//allLReward := abiObj.Methods["allLReward"]
	getWitnessInfo := abiObj.Methods["getWitnessInfo"]

	// 触发缓存过期方法
	addCommunity := abiObj.Methods["addCommunity"]
	delCommunity := abiObj.Methods["delCommunity"]
	addVote := abiObj.Methods["addVote"]
	delVote := abiObj.Methods["delVote"]
	addLight := abiObj.Methods["addLight"]
	delLight := abiObj.Methods["delLight"]
	communityDistribute := abiObj.Methods["communityDistribute"]
	setRate := abiObj.Methods["setRate"]
	distribute := abiObj.Methods["distribute"]
	withDrawW := abiObj.Methods["withDrawW"]

	//cacheMethodGroup := []string{
	//	utils.Bytes2string(getCommunityList.ID),
	//	utils.Bytes2string(allCReward.ID),
	//	utils.Bytes2string(getRate.ID),
	//	utils.Bytes2string(getLightListByC.ID),
	//	utils.Bytes2string(communityRewardPool.ID),
	//	utils.Bytes2string(queryAllReward.ID),
	//	utils.Bytes2string(getLightList.ID),
	//	utils.Bytes2string(allLReward.ID),
	//	utils.Bytes2string(getWitnessInfo.ID),
	//}

	onMethodNames := []string{
		utils.Bytes2string(addCommunity.ID),
		utils.Bytes2string(delCommunity.ID),
		utils.Bytes2string(addVote.ID),
		utils.Bytes2string(delVote.ID),
		utils.Bytes2string(addLight.ID),
		utils.Bytes2string(delLight.ID),
		utils.Bytes2string(communityDistribute.ID),
		utils.Bytes2string(setRate.ID),
		utils.Bytes2string(distribute.ID),
		utils.Bytes2string(withDrawW.ID),
	}

	onExpiredCache(utils.Bytes2string(getCommunityList.ID), onMethodNames...)
	//onExpiredCache(utils.Bytes2string(allCReward.ID), onMethodNames...)
	//onExpiredCache(utils.Bytes2string(getRate.ID), onMethodNames...)
	onExpiredCache(utils.Bytes2string(getLightListByC.ID), onMethodNames...)
	//onExpiredCache(utils.Bytes2string(communityRewardPool.ID), onMethodNames...)
	//onExpiredCache(utils.Bytes2string(queryAllReward.ID), onMethodNames...)
	onExpiredCache(utils.Bytes2string(getLightList.ID), onMethodNames...)
	//onExpiredCache(utils.Bytes2string(allLReward.ID), onMethodNames...)
	onExpiredCache(utils.Bytes2string(getWitnessInfo.ID), onMethodNames...)
}

// 获取轻节点缓存
func GetCache(input []byte) (interface{}, bool) {
	if !isLight() || len(input) < 4 {
		return nil, false
	}

	return lightCache.Get(utils.Bytes2string(input))
}

// 写入轻节点缓存
func SetCache(input []byte, value interface{}) {
	if !isLight() || len(input) < 4 {
		return
	}

	lightCache.Add(utils.Bytes2string(input), value)
}

// 监听方法签名=>清空对应的缓存
func WatchCache(input []byte) {
	if !isLight() || len(input) < 4 {
		return
	}

	methodID := make([]byte, 4)
	copy(methodID, input[:4])

	val, ok := cacheGroupEvent.Load(utils.Bytes2string(methodID))
	if !ok {
		return
	}

	// 清空缓存
	//for _, cacheMethodID := range val.([]string) {
	//	for _, cacheKey := range lightCache.Keys() {
	//		if strings.HasPrefix(cacheKey, cacheMethodID) {
	//			lightCache.Remove(cacheKey)
	//		}
	//	}
	//}

	_ = val
	lightCache.Purge()
}

// 补充一些多个返回值情况的结构
type CacheGetRoleTotal struct {
	CommunityCount uint64
	LightCount     uint64
}

type CacheGetRewardRatioAndVoteByAddrs struct {
	Rates []uint8
	Votes []*big.Int
}

// 是否轻节点
func isLight() bool {
	if config.Model == config.Model_light {
		return true
	}
	return false
}

// 监听数据过期事件
// cacheMethodName 缓存变化的方法签名
// onMethodNames 事件：方法签名
func onExpiredCache(cacheMethodName string, onMethodNames ...string) {
	for _, name := range onMethodNames {
		if item, ok := cacheGroupEvent.Load(name); ok {
			cacheNames := item.([]string)
			cacheNames = append(cacheNames, cacheMethodName)
			cacheGroupEvent.Store(name, cacheNames)
		} else {
			cacheGroupEvent.Store(name, []string{cacheMethodName})
		}
	}
}
