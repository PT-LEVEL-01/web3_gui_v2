package main

import (
	"math/big"
	"web3_gui/libp2parea/v2/config"
	nodeStore "web3_gui/libp2parea/v2/node_store"
	"web3_gui/utils"
)

/*
搜索逻辑节点id算法
*/
func main() {
	example()
}

func example() {
	utils.Log.Info().Msgf("start")

	kl := nodeStore.NewBucket(0)

	//要查找的节点ID
	findNetid := nodeStore.AddressFromB58String("J4VhVRKxqsuTrqTDreHz8qemMWAT3S5kg9mBdgusuv5p")

	// netids := make([]*big.Int, 0)
	netid1 := nodeStore.AddressFromB58String("65U2vEhNAq2PNzMT9SCK2d6sZCULSFvD8vvLwRZpJkgx")
	kl.Add(new(big.Int).SetBytes(netid1.Data()))
	// netids = append(netids, new(big.Int).SetBytes(netid1))
	netid2 := nodeStore.AddressFromB58String("AfwFmdvpqjXmVpR7MP1vQonLucaf1rQg6v7BJ9W6AirL")
	kl.Add(new(big.Int).SetBytes(netid2.Data()))
	// netids = append(netids, new(big.Int).SetBytes(netid2))
	netid3 := nodeStore.AddressFromB58String("6vk8xDbbZaT74eMfvbfby1VCifYF31qKmTpcwFk41aQH")
	kl.Add(new(big.Int).SetBytes(netid3.Data()))
	// netids = append(netids, new(big.Int).SetBytes(netid3))
	netid4 := nodeStore.AddressFromB58String("5bzPVbPzp9WMbXkH6w1qaYSD93FGVdbyuHYUNbioZCrU")
	kl.Add(new(big.Int).SetBytes(netid4.Data()))
	// netids = append(netids, new(big.Int).SetBytes(netid4))
	netid5 := nodeStore.AddressFromB58String("YYBNvYGxyAxPeLNwjkBrjwuPnaCawwD3nbEPH7vyuZ5")
	kl.Add(new(big.Int).SetBytes(netid5.Data()))
	// netids = append(netids, new(big.Int).SetBytes(netid5))
	netid6 := nodeStore.AddressFromB58String("9cV2PRMi9g7g92EQjQG4jU3pA3Rdu1ffxGu3i4KgQ9yX")
	kl.Add(new(big.Int).SetBytes(netid6.Data()))
	// netids = append(netids, new(big.Int).SetBytes(netid6))

	// asc := nodeStore.NewIdASC(new(big.Int).SetBytes(findNetid), netids)
	// result := asc.Sort()

	// for _, idBig := range result {
	// 	netidBs := idBig.Bytes()
	// 	utils.Log.Info().Msgf("最近的节点:%s", nodeStore.AddressNet(netidBs).B58String())
	// }

	targetIds := kl.Get(new(big.Int).SetBytes(findNetid.Data()))

	targetId := targetIds[0]

	targetIdBs := targetId.Bytes()

	addrData := utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length)

	mh, ERR := nodeStore.BuildAddrByData(findNetid.GetPre(), *addrData) // nodeStore.AddressNet(*utils.FullHighPositionZero(&targetIdBs, config.Addr_byte_length))
	if ERR.CheckFail() {
		utils.Log.Error().Str("错误", ERR.String()).Send()
		return
	}
	utils.Log.Info().Msgf("检查id结果:%s", mh.B58String())

}
