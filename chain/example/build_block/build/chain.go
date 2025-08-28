package build

import (
	"bytes"
	"encoding/hex"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"math"
	"sync"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/mining"
	"web3_gui/chain/rpc"
	"web3_gui/keystore/adapter"
	"web3_gui/libp2parea/adapter"
	"web3_gui/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func BuildChain(bhvo *mining.BlockHeadVO) *mining.Chain {
	chain := &mining.Chain{}
	chain.StopSyncBlock = 0
	chain.BlockHighStopGrowth = 0
	chain.SyncBlockLock = new(sync.RWMutex)
	chain.SyncBlockFinish = false
	// chain.No = no
	wb := mining.NewWitnessBackup(chain)
	wc := mining.NewWitnessChain(wb, chain)
	tm := mining.NewTransactionManager(wb)
	b := mining.NewBalanceManager(wb, tm, chain)
	chain.WitnessBackup = wb
	chain.WitnessChain = wc
	chain.Balance = b
	chain.TransactionManager = tm
	//chain.signLock = new(sync.RWMutex)

	//计算余额
	chain.Balance.CountBalanceForBlock(bhvo)

	//统计交易中的备用见证人以及见证人投票
	chain.WitnessBackup.CountWitness(&bhvo.Txs)

	//
	chain.WitnessChain.BuildWitnessGroup(true, true)

	//把见证人设置为已出块
	chain.WitnessChain.SetWitnessBlock(bhvo)

	chain.WitnessChain.BuildBlockGroup(bhvo, nil)

	chain.CurrentBlockHash = bhvo.BH.Hash

	return chain
}

func LoopLoadBlock(chain *mining.Chain, bhvo *mining.BlockHeadVO) {
	beforBlockHash := bhvo.BH.Nextblockhash
	//blockHeight := uint64(0)
	for beforBlockHash != nil {
		bhvo, err := mining.LoadBlockHeadVOByHash(&beforBlockHash)
		if err != nil {
			panic(err)
		}
		//utils.LogZero.Info().Uint64("加载高度", bhvo.BH.Height).Send()
		//start := time.Now()
		//导入区块
		err = addBlock(chain, bhvo)
		if err != nil {
			panic(err)
		}
		//utils.LogZero.Info().Uint64("加载高度", bhvo.BH.Height).Str("花费时间", time.Now().Sub(start).String()).Send()
		//blockHeight = bhvo.BH.Height
		beforBlockHash = bhvo.BH.Nextblockhash
	}
}

func LoopBuildBlock(chain *mining.Chain, areas []*libp2parea.Area, events []*Event) {
	blockHeight := uint64(0)
	smallGroup := chain.WitnessChain.GetCurrGroup()
	keyMap := make(map[string]*libp2parea.Area)
	for i, _ := range areas {
		one := areas[i]
		keyMap[utils.Bytes2string(one.Keystore.GetCoinbase().Addr)] = areas[i]
	}
	for {
		//找到还未出块的见证人组
		for _, witnessOne := range smallGroup.Witness {
			//超过了现在的时间，就退出
			if time.Now().Unix() < witnessOne.CreateBlockTime {
				utils.LogZero.Info().Str("超过时间了", "").Send()
				return
			}
			var key keystore.KeystoreInterface
			//找到对应的key
			area, ok := keyMap[utils.Bytes2string(*witnessOne.Addr)]
			if ok {
				key = area.Keystore
			}

			//执行指定高度事件
			txs := ProcessEventHandlers(blockHeight, events)
			for _, one := range txs {
				utils.LogZero.Info().Hex("事件交易", *one.GetHash()).Send()
			}
			chain.GetTransactionManager().AddTxs(txs...)

			//构建区块
			bhvo := BuildBlockOnly(witnessOne, chain, key, witnessOne.CreateBlockTime)
			if bhvo == nil {
				utils.LogZero.Info().Str("构建区块失败", "").Send()
				return
			}

			//收集其他见证人签名
			for _, one := range witnessOne.WitnessBigGroup.Witnesses {
				if bytes.Equal(*one.Addr, *witnessOne.Addr) {
					continue
				}
				areaOne, ok := keyMap[utils.Bytes2string(*one.Addr)]
				if !ok {
					panic("未找到密钥")
				}
				tmp := *bhvo.BH
				blockHeadCopy := &tmp
				blockHeadCopy.InitExtSign()
				blockHeadCopy.BuildSign(areaOne.Keystore.GetCoinbase().Addr, areaOne.Keystore)
				//bhvo.BH.ExtSign = append(bhvo.BH.ExtSign, blockHeadCopy.Sign)
				bhvo.BH.SetExtSign(blockHeadCopy.Sign)
			}
			//自己重新签名
			bhvo.BH.BuildSign(key.GetCoinbase().Addr, key)
			bhvo.BH.BuildBlockHash()

			//
			//PrintBHVO(bhvo)

			//导入区块
			err := addBlock(chain, bhvo)
			if err != nil {
				if err.Error() != mining.ERROR_repeat_import_block.Error() {
					panic(err)
				}
			}
			blockHeight = bhvo.BH.Height
		}
		//return
		smallGroup = smallGroup.NextGroup
	}
}

/*
见证人方式出块
出块并广播
@gh    uint64    出块的组高度
@id    []byte    押金id
*/
func BuildBlockOnly(this *mining.Witness, chain *mining.Chain, key keystore.KeystoreInterface, createBlockTime int64) *mining.BlockHeadVO {
	// var this *Witness
	coinbase := key.GetCoinbase()

	//自己是见证人才能出块，否则自己出块了，其他节点也不会承认
	if !bytes.Equal(*this.Addr, coinbase.Addr) {
		return nil
	}

	//workModeLockStatic.GetLock()

	//start := config.TimeNow()

	//查找出未确认的块
	preBlock, blocks := this.FindUnconfirmedBlock()
	// start2 := config.TimeNow().Sub(start)
	// utils.LogZero.Info().Msgf("上一个块高度 %d %d %d", preBlock.Height, blocks[0].Height, blocks[0].witness.Group.Height)
	// utils.LogZero.Info().Msgf("获取上一个块高度消耗时间 %s", config.TimeNow().Sub(start))
	utils.LogZero.Info().Msgf("=== build block === group:%d height:%d", this.Group.Height, preBlock.Height+1)

	//存放交易
	tx := make([]mining.TxItr, 0)
	txids := make([][]byte, 0)

	var reward *mining.Tx_reward
	//检查本组是否给上一组见证人奖励
	if this.WitnessBigGroup != preBlock.Witness.WitnessBigGroup {
		// utils.LogZero.Info().Msgf("开始构建上一组见证人奖励 %s %d", fmt.Sprintf("%p", preBlock.witness.WitnessBigGroup),
		// preBlock.witness.Group.Height)
		reward, _ = preBlock.Witness.WitnessBigGroup.CountRewardToWitnessGroup(chain, preBlock.Height+1, blocks, preBlock)
		//reward = preBlock.witness.WitnessBigGroup.CountRewardToWitnessGroupByContract(preBlock.Height+1, blocks, preBlock)
		tx = append(tx, reward)
		// utils.LogZero.Info().Msgf("reward:%+v", reward)
		txids = append(txids, reward.Hash)
	}

	//云存储代理奖励
	//TODO 需求变更,待修改
	//if 1 == 0 && (preBlock.Height+1)%config.CloudStorage_Reward_Interval == 0 {
	//	cloudStorageProxyPayload := precompiled.BuildCloudStorageProxyRewardInput()
	//	cloudStorageProxyTx := CreateTxCloudStorageProxy(precompiled.CloudStorageProxyContract, preBlock.Height+1, cloudStorageProxyPayload)
	//	tx = append(tx, cloudStorageProxyTx)
	//	txids = append(txids, cloudStorageProxyTx.Hash)
	//}

	// start3 := config.TimeNow().Sub(start)
	// utils.LogZero.Info().Msgf("394构建区块奖励消耗时间 %s,%s", start2, start3)

	//打包所有交易
	//chain := forks.GetLongChain()
	txs, ids := chain.TransactionManager.Package(reward, preBlock.Height+1, blocks, this.CreateBlockTime)
	tx = append(tx, txs...)
	txids = append(txids, ids...)

	// utils.LogZero.Info().Msgf("打包消耗时间 %s", config.TimeNow().Sub(start))

	//准备块中的交易
	// fmt.Println("准备块中的交易")
	//coinbase := key.GetCoinbase()

	var bh *mining.BlockHead
	//now := config.TimeNow().Unix()
	//now := config.TimeNow().UnixNano()
	//pl time
	//for i := int64(0); i < (config.Mining_block_time*2)-1; i++ {
	for i := int64(0); i < int64(math.Ceil(config.Mining_block_time.Seconds()))*2-1; i++ {
		//开始生成块
		bh = &mining.BlockHead{
			Height:            preBlock.Height + 1, //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       this.Group.Height,   // preGroup.Height + 1,               //矿工组高度
			Previousblockhash: preBlock.Id,         //上一个区块头hash
			NTx:               uint64(len(tx)),     //交易数量
			Tx:                txids,               //本区块包含的交易id
			Time:              createBlockTime + i, //unix时间戳
			Witness:           coinbase.Addr,       //此块矿工地址
			ExtSign:           make([][]byte, 0),
		}

		// if !testBug && bh.Height > config.Mining_block_start_height+12 && this.Group.FirstWitness() {
		// 	utils.LogZero.Error().Msgf("把区块高度减1")
		// 	testBug = true
		// 	bh.Height = bh.Height - 1
		// }

		bh.BuildMerkleRoot()
		bh.BuildSign(coinbase.Addr, key)
		//bh.BuildBlockHash()
		if ok, _ := bh.CheckHashExist(); ok {
			bh = nil
			continue
		} else {
			break
		}
	}
	if bh == nil {
		//workModeLockStatic.BackLock()
		utils.LogZero.Info().Msgf("Block out failed, all hash have collisions")
		//出块失败，所有hash都有碰撞
		return nil
	}

	bhvo := mining.CreateBlockHeadVO(config.StartBlockHash, bh, tx)

	//先保存到数据库再广播，否则其他节点查询不到
	// SaveBlockHead(bhvo)

	bhvo.FromBroadcast = true
	// utils.LogZero.Info().Msgf("打包消耗时间 %s", config.TimeNow().Sub(start))
	return bhvo
}

func addBlock(chain *mining.Chain, bhvo *mining.BlockHeadVO) error {

	start := config.TimeNow()

	utils.LogZero.Info().Msgf("====== Import block group:%d block:%d prehash:%s hash:%s witness:%s", bhvo.BH.GroupHeight,
		bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash), hex.EncodeToString(bhvo.BH.Hash), bhvo.BH.Witness.B58String())
	// if bhvo.BH.Height <= this.GetCurrentBlock()+1 {
	// 	utils.LogZero.Info().Msgf("Block height too low")
	// 	return nil
	// }
	bhvo.BH.BuildBlockHash()
	//先保存区块
	ok, err := mining.SaveBlockHead(bhvo)
	if err != nil {
		utils.LogZero.Warn().Msgf("save block error %s", err.Error())
		return err
	}
	if !ok {
		utils.LogZero.Error().Msgf("收到的区块不连续，则从邻居节点同步")
		//收到的区块不连续，则从邻居节点同步
		//this.NoticeLoadBlockForDB()
		return nil
	}

	//不能导入比统计高度低的区块
	if bhvo.BH.Height <= chain.GetCurrentBlock() {
		utils.LogZero.Info().Msgf("不能导入比统计高度低的区块")
		//检查区块是否已经导入过了
		if chain.WitnessChain.CheckRepeatImportBlock(bhvo) {
			utils.LogZero.Info().Msgf("重复导入区块")
			return mining.ERROR_repeat_import_block
		}
		return mining.ERROR_fork_import_block
	}

	//如果前置区块是统计过的区块，则前置区块hash必须对应，不能分叉。
	if chain.CurrentBlockHash != nil && len(chain.CurrentBlockHash) != 0 {
		if bhvo.BH.Height == chain.GetCurrentBlock()+1 && !bytes.Equal(chain.CurrentBlockHash, bhvo.BH.Previousblockhash) {
			utils.LogZero.Info().Msgf("比已经统计的区块高度+1的区块，前置区块必须对应")
			return mining.ERROR_fork_import_block
		}
	}

	//检查出块时间和本机时间相比，出块只能滞后，不能提前
	now := config.TimeNow().Unix()
	if bhvo.BH.Time > now+int64(config.Mining_block_time.Seconds()) {
		//if time.Unix(bhvo.BH.Time, 0).UnixNano() > now+config.Mining_block_time.Nanoseconds() {
		utils.LogZero.Warn().Msgf("Build block It's too late %d %d", bhvo.BH.Time, now)
		//出块时间提前了
		return config.ERROR_advanced_by_build_block
	}
	//排除上链但不合法的交易
	for _, one := range config.Exclude_Tx {
		if bhvo.BH.Height != one.Height {
			continue
		}
		for j, two := range bhvo.Txs {
			if !bytes.Equal(one.TxByte, *two.GetHash()) {
				// utils.LogZero.Info().Msgf("交易hash不相同 %d %s %d %s", len(one.TxByte),
				// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
				continue
			}
			notExcludeTx := bhvo.Txs[:j]
			bhvo.Txs = append(notExcludeTx, bhvo.Txs[j+1:]...)
			break
		}
	}
	//检查区块是否已经导入过了
	if chain.WitnessChain.CheckRepeatImportBlock(bhvo) {
		utils.LogZero.Info().Msgf("重复导入区块")
		return mining.ERROR_repeat_import_block
	}
	//查找前置区块是否存在，并且合法
	preWitness := chain.WitnessChain.FindPreWitnessForBlock(bhvo.BH.Previousblockhash)

	//if bhvo.BH.Height == importHeight && importHeightCount < importHeightCountMax {
	//	time.Sleep(1 * time.Second)
	//	importHeightCount++
	//	preWitness = nil
	//}

	if preWitness == nil {
		//有见证人节点分叉了，还在继续出块，但是分叉的长度没有自己的链长
		if chain.GetCurrentBlock() > bhvo.BH.Height {

			utils.LogZero.Error().Msgf("有见证人节点分叉了，还在继续出块，但是分叉的长度没有自己的链长")
			return nil
		}
		// utils.LogZero.Warn().Msgf("找不到前置区块，新区块不连续 新高度：%d", bhvo.BH.Height)
		utils.LogZero.Warn().Msgf("The front block cannot be found, and the new block is discontinuous with a new height:%d preblockhash:%s",
			bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash))
		//这里有可能分叉，有可能有断掉的区块未同步到位，现在都默认为区块未同步到位
		//TODO 有可能产生分叉，处理分叉的情况

		//从邻居节点同步区块
		//chain.NoticeLoadBlockForDB()
		return mining.ERROR_fork_import_block
	} else {
		//判断区块高度是否连续
		// utils.LogZero.Info().Msgf("pre witness group:%d height:%d bhvo height:%d",
		// 	preWitness.Group.Height, preWitness.Block.Height, bhvo.BH.Height)
		if preWitness.Block.Height+1 != bhvo.BH.Height {
			utils.LogZero.Error().Msgf("new block height fail,pre height:%d new height:%d", preWitness.Block.Height, bhvo.BH.Height)
			return mining.ERROR_import_block_height_not_continuity
		}
		// utils.LogZero.Info().Msgf("查找前一个组的见证人")
		preGroupWitness := preWitness //查找前一个组的见证人
		for preGroupWitness.Group.Height == bhvo.BH.GroupHeight {
			//继续向前查找,直到找到不相等的组
			preGroupWitness = chain.WitnessChain.FindPreWitnessForBlock(preGroupWitness.Block.PreBlockID)
			if preGroupWitness == nil {
				utils.LogZero.Info().Msgf("ERROR_fork_import_block")
				return mining.ERROR_fork_import_block
			}
		}
		// utils.LogZero.Info().Msgf("找到的前置组:%+v", preGroupWitness.Group.NextGroup)
		//检查新添加的区块和前置区块之间是否有已经确认的组
		group := preGroupWitness.Group.NextGroup
		for group != nil && group.Height < bhvo.BH.GroupHeight {
			if group.BlockGroup != nil {
				utils.LogZero.Info().Msgf("ERROR_fork_import_block")
				return mining.ERROR_fork_import_block
			}
			group = group.NextGroup
		}
	}
	var currentWitness *mining.Witness
	var isOverWitnessGroupChain bool
	//是首个区块，这里不构建前面的组
	if bhvo.BH.GroupHeight != config.Mining_group_start_height {
		//检查新区块是否在备用见证人组中
		currentWitness, isOverWitnessGroupChain = chain.WitnessChain.FindWitnessForBlockOnly(bhvo)
		// utils.LogZero.Info().Msgf("FindWitnessForBlockOnly:%t", isOverWitnessGroupChain)
		if isOverWitnessGroupChain {
			utils.LogZero.Info().Msgf("The new height of the witness for this block cannot be found:%d", bhvo.BH.Height)
			//找到最后一个见证人
			var lastGroup = chain.WitnessChain.GetCurrGroup()
			for ; lastGroup != nil; lastGroup = lastGroup.NextGroup {
				if lastGroup.NextGroup == nil {
					break
				}
			}
			//查找出未确认的块
			preBlock, _ := lastGroup.Witness[len(lastGroup.Witness)-1].FindUnconfirmedBlock()
			// utils.LogZero.Info().Msgf("未确认的最后块:%d", preBlock.Height)
			//先统计之前的区块
			preBlock.Witness.Group.BuildGroup(&preBlock.Id)
			chain.CountBlock(preBlock.Witness.Group)

			// _, lastBlock := this.GetLastBlock()
			// utils.LogZero.Info().Msgf("最新高度:%d", lastBlock.Height)

			//先统计之前的区块
			// this.WitnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
			// utils.LogZero.Info().Msgf("再构建后面的备用见证人组")
			//再构建后面的备用见证人组
			chain.WitnessChain.CompensateWitnessGroupByGroupHeight(bhvo.BH.GroupHeight)
			// utils.LogZero.Info().Msgf("再统计之前的区块 start")
			// this.WitnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
			// utils.LogZero.Info().Msgf("再统计之前的区块 end")
		}
	}
	currentWitness, _ = chain.WitnessChain.FindWitnessForBlockOnly(bhvo)
	if currentWitness == nil {
		utils.LogZero.Error().Msgf("not font witness")
		//找不到这个见证人
		return config.ERROR_not_font_witness
	}

	//严格的导入区块顺序,如 组内W1,W2,W3三个见证人,但是出块顺序却是W2,W1,W3
	//if err := chain.strictAddBlock(bhvo); err != nil {
	//	utils.LogZero.Error().Msgf("Not strict import block")
	//	return err
	//}

	//把多余的区块清理掉
	//if bhvo.BH.GroupHeight > this.GroupHeight {
	//	witness := this.WitnessChain.FindPreWitnessForBlock(bhvo.BH.Previousblockhash)
	//	if witness != nil {
	//		witness = witness.NextWitness
	//		for ; witness != nil; witness = witness.NextWitness {
	//			witness.Block = nil
	//		}
	//	}
	//}
	//}

	//判断是否相同组，相同组内不能有相同高度的区块
	// if currentWitness == preWitness {
	// 	for _, one := range currentWitness.Group.Witness {
	// 		if one.Block == nil {
	// 			continue
	// 		}
	// 		if one.Block.Height == bhvo.BH.Height {
	// 			utils.LogZero.Info().Msgf("有相同的区块高度 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// 			return config.ERROR_block_height_equality
	// 		}
	// 	}
	// } else {
	// 	//不同组也不能有相同高度的区块
	// 	//group:1195 block:2499 prehash:66d738aa4ccaef4f2eba63f05fc2c32518689b05c71a8889b306c1fcbfaeeab9 hash:77834994f87a964c8c0728968c04ec084339b7013e5f08870dc9df523f0aba3f
	// 	//group:1195 block:2500 prehash:77834994f87a964c8c0728968c04ec084339b7013e5f08870dc9df523f0aba3f hash:ce1c98913b7d219af20413c4f80168b837fb540d81f5bddbb3ad2d92babe218f
	// 	//group:1195 block:2501 prehash:ce1c98913b7d219af20413c4f80168b837fb540d81f5bddbb3ad2d92babe218f hash:26f78fdd73056df6596db487174cd9203d979bca600ca7ef05b9a5f1dc6fb114
	// 	//group:1196 block:2501 prehash:ce1c98913b7d219af20413c4f80168b837fb540d81f5bddbb3ad2d92babe218f hash:b6179c2fcd8b18e1eb6d0e3a1ea2201290ebe8f30776258802bb8e53ecee4500
	// 	for _, one := range preWitness.Group.Witness {
	// 		if one.Block == nil {
	// 			continue
	// 		}
	// 		if one.Block.Height == bhvo.BH.Height {
	// 			utils.LogZero.Info().Msgf("有相同的区块高度 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// 			return config.ERROR_block_height_equality
	// 		}
	// 	}
	// }

	// utils.LogZero.Info().Msgf("导入区块签名总个数:%d 现有:%d 最少:%d", len(currentWitness.WitnessBigGroup.Witnesses),
	// 	len(bhvo.BH.ExtSign)+1, config.BftMajorityPrinciple(len(currentWitness.WitnessBigGroup.Witnesses)))
	//验证区块见证人签名
	if !bhvo.BH.CheckExtSign(currentWitness.WitnessBigGroup.Witnesses) {
		// 	//区块验证不通过，区块不合法
		utils.LogZero.Warn().Msgf("AddBlock verification failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return config.ERROR_sign_fail
	}
	//验证区块见证人签名
	// if !bhvo.BH.CheckBlockHead(currentWitness.Puk) {
	// 	//区块验证不通过，区块不合法
	// 	utils.LogZero.Warn().Msgf("Block verification failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// 	// this.chain.NoticeLoadBlockForDB(false)
	// 	return
	// }

	//utils.LogZero.Info().Msgf("add block spend time 666 %s", config.TimeNow().Sub(start))
	// utils.LogZero.Info().Msgf("add block group:%d,height:%d 444444444444444444444444444", bhvo.BH.GroupHeight, bhvo.BH.Height)
	//把见证人设置为已出块 ===合约检查奖励验证===
	ok = chain.WitnessChain.SetWitnessBlock(bhvo)
	if !ok {
		//从邻居节点同步
		// utils.LogZero.Info().Msgf("------------------------------------------------------------ 22222222222222")
		utils.LogZero.Debug().Msgf("Setting witness block failed")
		// this.NoticeLoadBlockForDB(false)
		return errors.New("Setting witness block failed")
	}
	//this.GroupHeight = bhvo.BH.GroupHeight
	// utils.LogZero.Info().Msgf("add block group:%d,height:%d 5555555555555555555555555555555", bhvo.BH.GroupHeight, bhvo.BH.Height)

	//导入区块后，清理分叉
	//chain.cleanFork()

	//是首个区块，这里不构建前面的组
	if bhvo.BH.GroupHeight == config.Mining_group_start_height {
		chain.WitnessChain.GetCurrGroup().BuildGroup(nil)
		chain.WitnessChain.BuildWitnessGroup(false, false)
		chain.WitnessChain.WitnessGroup = chain.WitnessChain.WitnessGroup.NextGroup
		chain.WitnessChain.BuildWitnessGroup(false, true)
	} else {
		//本组有2个以上的见证人出块，那么本组才有可能被确认
		// utils.LogZero.Info().Msgf("111111111 %d", currentWitness.Group.Height)
		//判断这个组是否多人出块
		// ok, group := currentWitness.Group.CheckBlockGroup(nil)
		ok, group := currentWitness.Group.CheckBlockGroup(&currentWitness.Block.Id)
		//第三版共识添加的内容。两个组ABC ABC，当第二组第一个块A导入后，立即统计前组最后一个块C。
		if !ok && group != nil {
			//找到上一个未确认的组最后一个见证人
			witness := chain.WitnessChain.FindPreWitnessForBlock(group.Blocks[0].PreBlockID)
			chain.CountBlock(witness.Group)
		}
		//第三版共识，本组一旦有多人出块，当有新区块连接它，则统计之前的区块
		//例如：本组有3个块，ABC，当导入C块时，立即统计AB块
		if ok {
			//统计本组  合约统计
			chain.CountBlockCustom(currentWitness.Group, group)
		}
		if ok {
			//找到上一个未确认的组最后一个见证人
			witness := chain.WitnessChain.FindPreWitnessForBlock(group.Blocks[0].PreBlockID)
			witness.Group.BuildGroup(&witness.Block.Id)
			//合约统计
			chain.CountBlock(witness.Group)
			chain.WitnessChain.WitnessGroup = currentWitness.Group
			chain.WitnessChain.BuildWitnessGroup(false, true)
		}
		//白名单计数，区块统计后才统计白名单计数。不合法的组不统计，只统计合法组中未出块的见证人
		chain.CountBlackListNumber()

		// 触发保存快照
		//if (ok || (!ok && group != nil)) && this.IsFullBlockGroup() {
		//	if err := snapshot.Save(this.CurrentBlock); err != nil {
		//		utils.LogZero.Warn().Msgf(">>>>>> Snapshot[保存] 未保存快照 %v,高度 %d", err, this.CurrentBlock)
		//	}
		//}

		//提交所有增量数据
		if err := db.LevelDB.StoreSnap(nil); err == nil {
			db.LevelDB.ResetSnap()
		}
	}
	// utils.LogZero.Info().Msgf("add block group:%d,height:%d 6666666666666666666666666", bhvo.BH.GroupHeight, bhvo.BH.Height)
	//TODO 检查从此时开始，未来还未出块的见证人组有多少，太少则创建新的组，避免跳过太多组后出块暂停

	// utils.LogZero.Info().Msgf("添加区块 666 耗时 %s", config.TimeNow().Sub(start))
	// utils.LogZero.Info().Msgf("保存区块 99999999999999 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// SaveTxToBlockHead(bhvo)
	//utils.LogZero.Info().Msgf("Save block Time spent %s", config.TimeNow().Sub(start))

	//是首个区块，这里不构建前面的组
	// if bhvo.BH.GroupHeight != config.Mining_group_start_height {
	// 	this.witnessChain.BuildBlockGroup(bhvo, preWitness)
	// }
	// utils.LogZero.Info().Msgf("添加区块 777 耗时 %s", config.TimeNow().Sub(start))

	//chain.WitnessChain.BuildMiningTime()

	//回收内存，将前n个见证人之前的见证人链删除。
	chain.WitnessChain.GCWitness()

	//if bhvo.BH.Height == config.Witness_backup_group_overheight {
	//	config.Witness_backup_group = config.Witness_backup_group_new
	//}

	// utils.LogZero.Info().Msgf("添加区块 888 耗时 %s", config.TimeNow().Sub(start))

	//删除之前区块的交易缓存
	// TxCache.RemoveHeightTxInCache(bhvo.BH.Height - (config.Mining_group_max * 2))

	//判断自己是否是见证人，自己是见证人，则添加其他见证人白名单连接

	// if this.witnessChain.FindWitness(keystore.GetCoinbase().Addr) {

	// }

	//pl
	endT := config.TimeNow().Sub(start)
	// utils.LogZero.Info().Msgf("Import block Time spent %s", endT)
	if endT.Seconds() > 1 {
		utils.LogZero.Info().Msgf("Import block Time spent too long %s", endT)
	}
	//if SaveBlockTime == nil {
	//	SaveBlockTime = make(map[uint64]string)
	//} else {
	//	SaveBlockTime[bhvo.BH.Height] = endT.String()
	//}

	return nil
}

/*
执行指定高度的事件
@gh    uint64    出块的组高度
@id    []byte    押金id
*/
func ProcessEventHandlers(blockHeight uint64, events []*Event) []mining.TxItr {
	txs := make([]mining.TxItr, 0)
	for _, one := range events {
		//utils.LogZero.Info().Interface("打印事件信息", one).Send()
		if blockHeight < one.StartHeight || blockHeight > one.EndHeight {
			continue
		}
		tx := one.EventHandler()
		txs = append(txs, tx)
	}
	return txs
}

func PrintBHVO(bhvo *mining.BlockHeadVO) {
	bh := bhvo.BH
	txs := make([]string, 0)
	for _, one := range bhvo.BH.Tx {
		txs = append(txs, hex.EncodeToString(one))
	}
	bhVO := rpc.BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		GroupHeightGrowth: bh.GroupHeightGrowth,                     //
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		ExtSign:           make([]string, 0, len(bh.ExtSign)),       //
	}
	for _, one := range bh.ExtSign {
		bhVO.ExtSign = append(bhVO.ExtSign, hex.EncodeToString(one))
	}

	bs, _ := json.Marshal(bhVO)
	utils.LogZero.Info().Msgf(string(bs))
}
