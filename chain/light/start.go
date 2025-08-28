package light

import (
	"web3_gui/chain/config"
	lightsnapshot "web3_gui/chain/light/snapshot"
	"web3_gui/chain/mining"
	"web3_gui/libp2parea/adapter/engine"
)

/*
轻节点方式启动
*/
func StartModelLight() {
	engine.Log.Info("============= 轻节点模式 =============")
	// 初始化本地SDK rpc handler映射
	SetupHanlders()

	//轻节点遇到分叉，自动修改分叉后，自动重启。
	for {
		bhvo := mining.LoadStartBlock()
		if bhvo == nil {
			engine.Log.Info("neighbor initiation block build chain")
			// chaninfo := mining.FindStartBlockForNeighbor()

			//从邻居节点同步区块
			err := mining.GetFirstBlock()
			if err != nil {
				engine.Log.Error("get first block error: %s", err.Error())
				panic(err.Error())
			}

			// 初始化轻节点快照
			mining.InitLightChainSnap()

			// engine.Log.Info("用邻居节点区块构建链2")
			mining.FindBlockHeight()
		} else {
			engine.Log.Info("load db initiation block build chain")
			if lightsnapshot.Height() > 0 {
				if err := mining.StartLightChainSnap(); err != nil {
					panic(err)
				}
				CountBalanceHeight = lightsnapshot.Height()
			} else {
				config.StartBlockHash = bhvo.BH.Hash
				//从本地数据库创始区块构建链
				mining.BuildFirstChain(bhvo)
				// 初始化轻节点快照
				mining.InitLightChainSnap()
				mining.FindBlockHeight()
			}
		}

		SyncBlock()
	}
}
