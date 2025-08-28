/*
见证人网络地址白名单管理
*/
package mining

import (
	"bytes"
	"context"
	"sync"
	"time"

	"web3_gui/chain/config"
	"web3_gui/chain/protos/go_protos"

	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/utils"
)

func init() {
	// go loopAddWitness()
	//utils.Go(loopAddWitness)
	utils.Go(loopAddWitnessTicker, nil)

}

// 延迟删除的白名单地址,旧的白名单地址将保留60组  addr:GroupHeight
var delayRemoveWhitelistNetMap = new(sync.Map)

var newWitnessAddrNetMap = new(sync.Map)

// var newWitnessAddrNets = make(chan *[]*crypto.AddressCoin, 100)
var addrnetWhilelist = new(sync.Map) //保存见证人对应网络地址.key:string=见证人地址;value:*nodeStore.AddressNet=见证人网络地址;

func loopAddWitnessTicker() {
	//定时器
	backoffTimer := utils.NewBackoffTimerChan(5*time.Second, 5*time.Second, 5*time.Second, 5*time.Second, 5*time.Second, 5*time.Second, 5*time.Second, 5*time.Second, 5*time.Second, 5*time.Second, time.Second*30)
	for {
		backoffTimer.Wait(context.Background())
		chain := GetLongChain()
		//当区块同步完成后再去查找见证人网络地址，否则将定时器重置间隔时间从头开始
		if chain != nil && chain.SyncBlockFinish {
			loopAddWitness()
		} else {
			backoffTimer.Reset()
		}

	}

}

func loopAddWitness() {
	lookupAddrs := new(go_protos.RepeatedBytes)
	newWitnessAddrNetMap.Range(func(k, v interface{}) bool {
		addrCoin := v.(*crypto.AddressCoin)
		//对比见证人地址，新增的添加
		if findWitnessAddrNet(addrCoin) == nil {
			//engine.Log.Warn("要寻找的见证人addrCoin地址:%s", addrCoin.B58String())
			lookupAddrs.Bss = append(lookupAddrs.Bss, *addrCoin)
		}
		// engine.Log.Info("1111111111111")

		return true
	})
	LookupWitnessAddrNet(lookupAddrs)
	return
}

/*func loopAddWitness() {
	for addrCoins := range newWitnessAddrNets {
		// engine.Log.Info("1111111111111")
		chain := GetLongChain()
		if !chain.SyncBlockFinish {
			continue
		}
		// engine.Log.Info("1111111111111")
		//对比见证人地址，多余的删除
		removeNet := make([]nodeStore.AddressNet, 0)
		addrnetWhilelist.Range(func(k, v interface{}) bool {
			key := k.(string)
			have := false
			for _, one := range *addrCoins {
				if key == utils.Bytes2string(*one) {
					have = true
					break
				}
			}
			if !have {
				addrNet := v.(*nodeStore.AddressNet)
				removeNet = append(removeNet, *addrNet)
			}
			return true
		})
		//for _, one := range removeNet {
		//	Area.RemoveWhiteList(one)
		//}
		// engine.Log.Info("1111111111111")
		//对比见证人地址，新增的添加
		lookupAddrs := new(go_protos.RepeatedBytes)
		for _, one := range *addrCoins {
			if findWitnessAddrNet(one) == nil {
				if bytes.Equal(*one, Area.Keystore.GetCoinbase().Addr) {
					continue
				}
				engine.Log.Error("要寻找的见证人addrCoin地址:%s", one.B58String())
				lookupAddrs.Bss = append(lookupAddrs.Bss, *one)
			}
		}
		if len(lookupAddrs.Bss) <= 0 {
			return
		}
		// engine.Log.Info("1111111111111")
		engine.Log.Error("本节点addrCoin地址%s net地址:%s", Area.Keystore.GetCoinbase().Addr.B58String(), Area.GetNetId().B58String())
		LookupWitnessAddrNet(lookupAddrs)
	}
}*/

func addWitnessAddrNet(addrCoin *crypto.AddressCoin, addrNet *nodeStore.AddressNet) {
	addrnetWhilelist.Store(utils.Bytes2string(*addrCoin), addrNet)
	//addrnetWhilelist.Range(func(k, v interface{}) bool {
	//	val := v.(*nodeStore.AddressNet)
	//	engine.Log.Error("-----------------------------保存的的见证人net地址：%s", val.B58String())
	//	return true
	//})
}

func findWitnessAddrNet(addrCoin *crypto.AddressCoin) *nodeStore.AddressNet {
	// engine.Log.Info("查询一个白名单:%s", addrCoin.B58String())
	value, ok := addrnetWhilelist.Load(utils.Bytes2string(*addrCoin))
	if ok {
		// engine.Log.Info("查询一个白名单:%s", addrCoin.B58String())
		if value != nil {
			// engine.Log.Info("查询一个白名单:%s", addrCoin.B58String())
			addrNet := value.(*nodeStore.AddressNet)
			return addrNet
		}
	}
	return nil
}

func AddWitnessAddrNets(chain *Chain, addrs []*crypto.AddressCoin) {

	//chain := GetLongChain()
	if !chain.SyncBlockFinish {
		return
	}

	//对比见证人地址，多余的删除
	addrnetWhilelist.Range(func(k, v interface{}) bool {
		key := k.(string)
		have := false
		for _, one := range addrs {
			if key == utils.Bytes2string(*one) {
				have = true
				break
			}
		}
		if !have {
			//addrNet := v.(*nodeStore.AddressNet)
			//Area.RemoveWhiteList(*addrNet)
			//addrnetWhilelist.Delete(key)
			delayRemoveWhitelist(key)
		}
		return true
	})

	first := true
	newWitnessAddrNetMap.Range(func(k, v interface{}) bool {
		first = false
		key := k.(string)
		have := false
		for _, one := range addrs {
			if key == utils.Bytes2string(*one) {
				have = true
				break
			}
		}
		if !have {
			//newWitnessAddrNetMap.Delete(key)
			delayRemoveWhitelist(key)
		}
		//addr := v.(*crypto.AddressCoin)
		//engine.Log.Warn("-----------------------------见证人Coin地址：%s", addr.B58String())
		return true
	})

	addLen := 0
	for _, one := range addrs {
		if !bytes.Equal(*one, Area.Keystore.GetCoinbase().Addr) {
			newWitnessAddrNetMap.Store(utils.Bytes2string(*one), one)
			addLen++
		}
	}

	if first && addLen > 0 {
		//engine.Log.Warn("同步完成后第一次马上要去找---------------")
		loopAddWitness()
	}

	//select {
	//case newWitnessAddrNets <- &addrs:
	//default:
	//}
}

// 延迟删除白名单功能,记录地址的已构建的最高组高度,超过这个高度才删除地址
func delayRemoveWhitelist(key string) {
	group := GetLongChain().WitnessChain.WitnessGroup
	currentGroupHeight := group.Height
	if _, ok := delayRemoveWhitelistNetMap.Load(key); !ok {
		//找到预构建组的最高组
		maxGroupHeight := group.Height
		for group != nil && group.NextGroup != nil {
			maxGroupHeight = group.Height
			group = group.NextGroup
		}
		//记录地址:最大组高度
		delayRemoveWhitelistNetMap.Store(key, maxGroupHeight)
	}

	delayRemoveWhitelistNetMap.Range(func(key, value any) bool {
		delGroupHeight := value.(uint64)
		//超过组高度才删除
		if currentGroupHeight > delGroupHeight {
			if addrNet, ok := addrnetWhilelist.Load(key); ok {
				Area.RemoveWhiteList(*addrNet.(*nodeStore.AddressNet))
				addrnetWhilelist.Delete(key)
			}

			newWitnessAddrNetMap.Delete(key)
		}

		return true
	})
}

/*
寻找见证人地址
*/
func LookupWitnessAddrNet(lookupAddrs *go_protos.RepeatedBytes) {
	// engine.Log.Info("1111111111111")
	if lookupAddrs == nil || len(lookupAddrs.Bss) <= 0 {
		return
	}
	// engine.Log.Info("1111111111111")
	bs, err := lookupAddrs.Marshal()
	if err != nil {
		engine.Log.Error("LookupWitnessAddrNet error:%s", err.Error())
		return
	}
	//engine.Log.Error("开始广播要寻找的见证人地址,本节点addrCoin地址%s net地址:%s", Area.Keystore.GetCoinbase().Addr.B58String(), Area.GetNetId().B58String())
	//开始广播要寻找的见证人地址
	err = Area.SendMulticastMsg(config.MSGID_multicast_find_witness, &bs)
	if err != nil {
		engine.Log.Error("multicase find witness whilt address net error:%s", err.Error())
		return
	}
}
