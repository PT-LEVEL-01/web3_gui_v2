package light

import (
	"bytes"
	"encoding/hex"
	"errors"
	"runtime"
	"strconv"
	"time"
	lightsnapshot "web3_gui/chain/light/snapshot"

	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/mining"

	"github.com/shirou/gopsutil/v3/mem"
	"web3_gui/libp2parea/adapter/engine"
)

var CountBalanceHeight = uint64(0)

/*
开始同步区块
*/
func SyncBlock() error {
	CountBalanceHeight = 0
	bhvo, err := LoadBlockChain()
	if err != nil {
		return err
	}
	// engine.Log.Info("SyncBlock hash:%s", hex.EncodeToString(bhvo.BH.Hash))
	chain := mining.GetLongChain()
	bhvo, err = SycnBlockChain(bhvo)
	if err != nil {
		for chain.Temp == nil {
			time.Sleep(time.Second * 6)
		}
		chain.ForkCheck(nil)
		return nil
	}
	chain.SyncBlockFinish = true

	LoopSycnBlockChain(bhvo)
	for chain.Temp == nil {
		time.Sleep(time.Second * 6)
	}
	chain.ForkCheck(nil)
	// ForkCheck(bhsh)
	return nil
}

/*
从本地数据库加载区块
*/
func LoadBlockChain() (*mining.BlockHeadVO, error) {
	engine.Log.Info("Start loading blocks in database")
	mining.SetHighestBlock(db.GetHighstBlock())

	// maxHeight := db.GetHighstBlock()
	// //处理分叉的情况，抛弃最后几个区块
	// maxHeight = maxHeight - config.Mining_group_max

	// mining.loadBlockForDB()
	chain := mining.GetLongChain()

	//_, lastBlock := chain.GetLastBlock()

	headid := mining.LoadBlockHashByHeight(chain.CurrentBlock)
	if headid == nil {
		headid = &config.StartBlockHash
	}
	//headid := lastBlock.Id
	// engine.Log.Info("block hash:%s", hex.EncodeToString(headid))
	bhvo, err := mining.LoadBlockHeadVOByHash(headid)
	if err != nil {
		return bhvo, err
	}
	// engine.Log.Info("block hash:%s", hex.EncodeToString(bhvo.BH.Hash))
	if bhvo.BH.Nextblockhash == nil {
		return bhvo, nil
	}
	blockhash := bhvo.BH.Nextblockhash
	var bhvoTemp *mining.BlockHeadVO
	for blockhash != nil && len(blockhash) > 0 {
		// engine.Log.Info("1111111")
		//这里查看内存，控制速度
		memInfo, _ := mem.VirtualMemory()
		if memInfo.UsedPercent > config.Wallet_Memory_percentage_max {
			runtime.GC()
			time.Sleep(time.Second)
		}
		// engine.Log.Info("1111111")
		bhvoTemp, err = mining.LoadBlockHeadVOByHash(&blockhash)
		if err != nil || bhvoTemp == nil {
			// engine.Log.Info("1111111")
			break
		}
		// engine.Log.Info("1111111")
		// if bhvoTemp.BH.Height >= maxHeight {
		// 	break
		// }

		bhvo = bhvoTemp

		// 轻节点统计
		lightCountBlock(chain, bhvoTemp)

		//临时改变nexthash
		// if nextHash, ok := config.BlockNextHash.Load(utils.Bytes2string(bhvo.BH.Hash)); ok {
		// 	nextHashBs := nextHash.(*[]byte)
		// 	blockhash = *nextHashBs
		// 	continue
		// }
		// if bhvo.BH.Height == 817445 {
		// 	bhvo.BH.Nextblockhash = config.SpecialBlockHash
		// }

		if bhvoTemp.BH.Nextblockhash == nil || len(bhvoTemp.BH.Nextblockhash) <= 0 {
			break
		}
		blockhash = bhvoTemp.BH.Nextblockhash
	}
	engine.Log.Info("end loading blocks in database")
	// engine.Log.Info("block hash:%s", hex.EncodeToString(bhvo.BH.Hash))
	return bhvo, nil
}

/*
从远程节点同步区块
*/
func SycnBlockChain(bhvo *mining.BlockHeadVO) (*mining.BlockHeadVO, error) {
	engine.Log.Info("Start synchronizing blocks from neighbor nodes")
	bhvoLocal := bhvo
	// engine.Log.Info("Start Sync hash:%s", hex.EncodeToString(bhvoLocal.BH.Hash))

	chain := mining.GetLongChain()

	peerBlockinfo, rch := mining.FindRemoteCurrentHeight()

	// var bhvoRemote *mining.BlockHeadVO
	// var err error
	//刷新最后一个区块
	bhvoRemote, err := mining.LightWrapperSyncBlockFlashDB(&bhvoLocal.BH.Hash, peerBlockinfo)
	if err != nil {
		engine.Log.Info("SyncBlockFlashDB error:%s", err.Error())
		// SyncFailedTotal++
		return nil, err
	}
	bhvoLocal = bhvoRemote

	//记录邻居节点有最新高度，却同步不到的次数，这种情况可能是分叉了。
	SyncFailedTotal := 0
	for SyncFailedTotal < 60 {

		if rch <= CountBalanceHeight {
			peerBlockinfo, rch = mining.FindRemoteCurrentHeight()
			if rch <= CountBalanceHeight {
				// SyncFailedTotal = 0
				if !chain.SyncBlockFinish {
					chain.SyncBlockFinish = true
				}
				//同步完成，并且同步成功
				return bhvoLocal, nil
			}
		}

		mining.SetHighestBlock(rch)

		bhvoRemote, err = mining.LightWrapperSyncBlockFlashDB(&bhvoLocal.BH.Nextblockhash, peerBlockinfo)
		if err != nil {
			if err.Error() == config.ERROR_chain_sync_block_timeout.Error() || err.Error() == config.ERROR_wait_msg_timeout.Error() {
				time.Sleep(time.Second * 10)
				peerBlockinfo, rch = mining.FindRemoteCurrentHeight()
			}
			engine.Log.Info("SyncBlockFlashDB error:%s", err.Error())
			SyncFailedTotal++
			continue
		}
		engine.Log.Info("height:%d CountBalanceHeight:%d", bhvoRemote.BH.Height, CountBalanceHeight)

		if bhvoRemote.BH.Height > CountBalanceHeight {
			SyncFailedTotal = 0
			// 轻节点统计
			lightCountBlock(chain, bhvoRemote)
			// select {
			// case nowChan <- false:
			// default:
			// }
		}
		//对方能查到另一个分支的区块，但是，是断头的区块
		if bhvoRemote.BH.Nextblockhash == nil || len(bhvoRemote.BH.Nextblockhash) <= 0 {
			engine.Log.Info("block not nextblockhash: %s", hex.EncodeToString(bhvoRemote.BH.Hash))
			SyncFailedTotal++
			continue
		}

		// 保存轻节点快照
		if err := lightsnapshot.Save(chain.CurrentBlock); err != nil {
			engine.Log.Warn(">>>>>> Light Snapshot[保存] 未保存快照 %v,高度 %d", err, chain.CurrentBlock)
		}

		engine.Log.Info("sync next block hash")
		bhvoLocal = bhvoRemote
	}
	// chain.SyncBlockFinish = true
	engine.Log.Info("End synchronizing blocks from neighbor nodes")
	return bhvoLocal, errors.New("")
}

/*
循环从远程节点同步区块
*/
func LoopSycnBlockChain(bhvo *mining.BlockHeadVO) *[]byte {
	engine.Log.Info("Start loop synchronizing blocks from neighbor nodes")
	bhvoLocal := bhvo
	// engine.Log.Info("Start Sync hash:%s", hex.EncodeToString(bhvoLocal.BH.Hash))

	chain := mining.GetLongChain()

	var bhRemote *mining.BlockHead
	var bhvoRemote *mining.BlockHeadVO
	var err error

	//pl time
	//ticker := time.NewTicker(time.Second * config.Mining_block_time).C
	ticker := time.NewTicker(config.Mining_block_time).C
	nowChan := make(chan bool, 1)

	//第一次不等待，直接放行
	// nowChan <- false

	//记录邻居节点有最新高度，却同步不到的次数，这种情况可能是分叉了。
	SyncFailedTotal := 0

	var peerBlockinfo *mining.PeerBlockInfoDESC
	var rch uint64 // := mining.FindRemoteCurrentHeight()

	for SyncFailedTotal < 60 {

		select {
		case <-ticker: //同步到最新高度了，等待10秒钟拉取一次最新高度
			peerBlockinfo, rch = mining.FindRemoteCurrentHeight()
			if rch <= CountBalanceHeight {
				SyncFailedTotal = 0
				continue
			}
		case <-nowChan: //未同步到最新高度，不等待，直接继续同步。
		}
		engine.Log.Info("remote height:%d local height:%d", rch, CountBalanceHeight)
		if rch <= CountBalanceHeight {
			continue
		}
		//查询远程节点区块头
		bhRemote, err = mining.FindBlockHeadNeighbor(&bhvoLocal.BH.Hash, peerBlockinfo)
		if err != nil {
			engine.Log.Info("FindBlockHeadNeighbor error:%s", err.Error())
			SyncFailedTotal++
			continue
		}
		//下一个区块不一样，则修改本地
		if !bytes.Equal(bhRemote.Nextblockhash, bhvoLocal.BH.Nextblockhash) {
			bhvoLocal.BH.Nextblockhash = bhRemote.Nextblockhash
			bs, err := bhvoLocal.BH.Proto()
			if err != nil {
				return nil
			}
			bhashkey := config.BuildBlockHead(bhvoLocal.BH.Hash)
			db.LevelDB.Save(bhashkey, bs)
		}

		mining.SetHighestBlock(rch)

		//先查询本地是否有这个区块
		bhvoRemote, err = mining.LoadBlockHeadVOByHash(&bhvoLocal.BH.Nextblockhash)
		if err != nil || len(bhvoRemote.BH.Nextblockhash) == 0 {
			//本地没有，则从远程节点同步
			bhvoRemote, err = mining.LightWrapperSyncBlockFlashDB(&bhvoLocal.BH.Nextblockhash, peerBlockinfo)
			if err != nil {
				engine.Log.Info("SyncBlockFlashDB error:%s", err.Error())
				SyncFailedTotal++
				continue
			}
		}

		if bhvoRemote.BH.Height > CountBalanceHeight {
			SyncFailedTotal = 0
			// 轻节点统计
			lightCountBlock(chain, bhvoRemote)
			select {
			case nowChan <- false:
			default:
			}
		}
		//对方能查到另一个分支的区块，但是，是断头的区块
		if bhvoRemote.BH.Nextblockhash == nil || len(bhvoRemote.BH.Nextblockhash) <= 0 {
			engine.Log.Info("block not nextblockhash: %s", hex.EncodeToString(bhvoRemote.BH.Hash))
			SyncFailedTotal++
			continue
		}

		// 保存轻节点快照
		if err := lightsnapshot.Save(chain.CurrentBlock); err != nil {
			engine.Log.Warn(">>>>>> Light Snapshot[保存] 未保存快照 %v,高度 %d", err, chain.CurrentBlock)
		}

		engine.Log.Info("sync next block hash")
		bhvoLocal = bhvoRemote
	}
	chain.SyncBlockFinish = true
	return &bhvoLocal.BH.Hash
}

/*
统计区块的余额
*/
func CountBalance(chain *mining.Chain, bhvo *mining.BlockHeadVO) {
	engine.Log.Info("====== Count block group:%d block:%d prehash:%s hash:%s witness:%s", bhvo.BH.GroupHeight,
		bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash), hex.EncodeToString(bhvo.BH.Hash), bhvo.BH.Witness.B58String())
	mining.GetLongChain().Balance.CountBalanceForBlock(bhvo)
	chain.WitnessBackup.CountWitness(&bhvo.Txs)
}

/*
轻节点统计区块
*/
func lightCountBlock(chain *mining.Chain, bhvo *mining.BlockHeadVO) {
	engine.Log.Info("====== Count block group:%d block:%d prehash:%s hash:%s witness:%s", bhvo.BH.GroupHeight,
		bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash), hex.EncodeToString(bhvo.BH.Hash), bhvo.BH.Witness.B58String())
	// 统计余额
	mining.GetLongChain().Balance.CountBalanceForBlock(bhvo)
	// 统计备用见证人和见证人投票
	chain.WitnessBackup.CountWitness(&bhvo.Txs)
	chain.SetPulledStates(bhvo.BH.Height)
	chain.SetCurrentBlock(bhvo.BH.Height, bhvo.BH.Hash)
	CountBalanceHeight = bhvo.BH.Height
}

/*
处理分叉和同步卡住的情况
*/
func ForkCheck(bhash *[]byte) {
	engine.Log.Error("ForkCheck hash:%s", hex.EncodeToString(*bhash))
	//从最高高度，往前查找，重新建立前后区块关系
	// temp := this.Temp
	if bhash == nil {
		return
	}
	preBlockHash := bhash

	// finish := false

	// var nextBlockHash []byte
	for {
		engine.Log.Info("LoadBlockHeadByHash:%s", hex.EncodeToString(*preBlockHash))
		//检查数据库是否存在这个区块
		bh, err := mining.LoadBlockHeadByHash(preBlockHash)
		if err != nil || bh == nil {
			engine.Log.Error("ForkCheck faile!")
			return
		}

		peerBlockinfo, _ := mining.FindRemoteCurrentHeight()
		bhvo, err := mining.SyncBlockFlashDB(preBlockHash, peerBlockinfo)
		if err == nil && bhvo != nil {
			//判断分叉交叉点，查询远端区块是否是本地确认的区块
			//本地已经存在这个区块，判断这个区块是否是链上收录的块。是收录的块则找到这个分叉点了
			//bhash, err := db.LevelDB.Find([]byte(config.BlockHeight + strconv.Itoa(int(bhvo.BH.Height))))
			bhash, err := db.LevelDB.Find(append(config.BlockHeight, []byte(strconv.Itoa(int(bhvo.BH.Height)))...))
			if err == nil && bytes.Equal(*bhash, bhvo.BH.Hash) {
				engine.Log.Info("fork block:%s", hex.EncodeToString(*bhash))
				break
			}
		}

		preBlockHash = &bh.Previousblockhash
		// bh = bhvo.BH

		// //nextBlockHash存在，则刷新本地数据库
		// if nextBlockHash != nil && len(nextBlockHash) > 0 {
		// 	bh.Nextblockhash = nextBlockHash
		// 	bs, err := bh.Proto()
		// 	err = db.LevelDB.Save(bh.Hash, bs)
		// 	if err != nil {
		// 		return
		// 	}
		// }
		// preBlockHash = bh.Previousblockhash
		// nextBlockHash = bh.Hash
		// // db.Find()

		// if finish {
		// 	break
		// }
	}
	engine.Log.Error("ForkCheck finish!")
}
