package mining

import (
	"bytes"
	"encoding/hex"
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/libp2parea/adapter/engine"
)

/*
从数据库中加载区块
先找到内存中最高区块，从区块由低到高开始加载
*/
func (this *Chain) LoadBlockChain() error {
	engine.Log.Info("Start loading blocks in database")
	forks.SetHighestBlock(db.GetHighstBlock())

	_, lastBlock := this.GetLastBlock()
	headid := lastBlock.Id
	bh, _, err := loadBlockForDB(&headid)
	if err != nil {
		return err
	}

	engine.Log.Info("Starting loading blocks from database, %d,%d,%s", bh.Height, bh.GroupHeight, hex.EncodeToString(bh.Hash))

	if bh.Nextblockhash == nil {
		return nil
	}

	blockhash := bh.Nextblockhash
	var bhvo *BlockHeadVO
	for blockhash != nil && len(blockhash) > 0 {
		//这里查看内存，控制速度
		memInfo, _ := mem.VirtualMemory()
		if memInfo.UsedPercent > config.Wallet_Memory_percentage_max {
			runtime.GC()
			time.Sleep(time.Second)
		}

		if bhvo != nil && bhvo.BH.Height+1 == config.CutBlockHeight && bhvo.BH.Nextblockhash != nil && bytes.Equal(bhvo.BH.Nextblockhash, config.CutBlockHash) {
			//回滚区块高度
			break
		}

		bhvo = this.deepCycleLoadBlock(&blockhash)
		if bhvo == nil {
			break
		}
		//临时改变nexthash
		// if nextHash, ok := config.BlockNextHash.Load(utils.Bytes2string(bhvo.BH.Hash)); ok {
		// 	nextHashBs := nextHash.(*[]byte)
		// 	blockhash = *nextHashBs
		// 	continue
		// }
		// if bhvo.BH.Height == 817445 {
		// 	bhvo.BH.Nextblockhash = config.SpecialBlockHash
		// }

		if bhvo.BH.Nextblockhash == nil || len(bhvo.BH.Nextblockhash) <= 0 {
			break
		}
		blockhash = bhvo.BH.Nextblockhash
	}
	engine.Log.Info("end loading blocks in database")
	return nil
}

/*
深度循环加载区块，包括分叉的链的加载
加载到出错或者加载完成为止
*/
func (this *Chain) deepCycleLoadBlock(bhash *[]byte) *BlockHeadVO {
	//if len(config.BlockHashs) > 0 && bytes.Equal(*bhash, config.BlockHashs[0]) {
	//	bhash = this.testLoadBlock()
	//	// return
	//}
	//检查hash是否修改顺序
	if blockHashs, ok := config.BlockHashsMap[hex.EncodeToString(*bhash)]; ok {
		peerBlockinfo, _ := FindRemoteCurrentHeight()
		if currentbhvo := this.testLoadBlock(blockHashs, peerBlockinfo); currentbhvo != nil {
			bhash = &currentbhvo.BH.Nextblockhash
		}
	}

	bh, txItrs, err := loadBlockForDB(bhash)
	if err != nil {
		engine.Log.Info("load block for db error:%s", err.Error())
		return nil
	}
	// engine.Log.Info("--------深度循环加载区块 %d %s", bh.Height, hex.EncodeToString(bh.Hash))
	bhvo := &BlockHeadVO{FromBroadcast: false, BH: bh, Txs: txItrs}
	err = this.AddBlockOther(bhvo)
	if err != nil {
		if err.Error() == ERROR_repeat_import_block.Error() {
			//可以重复导入区块
		} else {
			engine.Log.Info("add block error:%s", err.Error())
			return nil
		}
	}
	//	chain.AddBlock(bh, &txItrs)
	if bh.Nextblockhash == nil {
		engine.Log.Info("load block next blockhash nil")
		return nil
	}
	// for i, _ := range bh.Nextblockhash {
	// 	// engine.Log.Info("--深度循环加载区块的下一个区块 %d %s", bh.Height, hex.EncodeToString(bh.Nextblockhash[i]))
	// 	this.deepCycleLoadBlock(&bh.Nextblockhash[i])
	// }
	// this.deepCycleLoadBlock(&bh.Nextblockhash)
	return bhvo
}

/*
深度循环加载区块，包括分叉的链的加载
加载到出错或者加载完成为止
*/
func (this *Chain) testLoadBlock(blockHashs [][]byte, peerBlockinfo *PeerBlockInfoDESC) *BlockHeadVO {
	engine.Log.Info("start testLoadBlock---------------------------")
	var currentbhvo *BlockHeadVO
	for _, one := range blockHashs {
		bhvo := &BlockHeadVO{}
		bh, txItrs, err := loadBlockForDB(&one)
		if err != nil {
			//continue
			engine.Log.Warn("find next block from db error:%s", err.Error())
			//去邻居节点同步
			if peerBlockinfo == nil {
				continue
			}
			bhvo, err = FindBlockForNeighbor(&one, peerBlockinfo)
			if err != nil {
				engine.Log.Error("find next block from neighbor error:%s", err.Error())
				continue
			}
		} else {
			bhvo.BH = bh
			bhvo.Txs = txItrs
		}

		// engine.Log.Info("--------深度循环加载区块 %d %s", bh.Height, hex.EncodeToString(bh.Hash))
		//bhvo := &BlockHeadVO{FromBroadcast: false, BH: bh, Txs: txItrs}
		err = this.AddBlockOther(bhvo)
		if err != nil {
			continue
		}
		//	chain.AddBlock(bh, &txItrs)
		if bhvo.BH == nil {
			continue
		}
		currentbhvo = bhvo
	}
	engine.Log.Info("end testLoadBlock---------------------------")
	return currentbhvo
}

/*
	从数据库中加载一个区块
*/
// func loadBlockForDB(bhash *[]byte) (*BlockHead, []TxItr, error) {
// 	head, err := db.Find(*bhash)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	hB, err := ParseBlockHead(head)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	txItrs := make([]TxItr, 0)
// 	for _, one := range hB.Tx {
// 		// txItr, err := FindTxBase(one, hex.EncodeToString(one))
// 		txItr, err := FindTxBase(one)

// 		// txBs, err := db.Find(one)
// 		if err != nil {
// 			// fmt.Println("3333", err)
// 			return nil, nil, err
// 		}
// 		// txItr, err := ParseTxBase(ParseTxClass(one), txBs)
// 		txItrs = append(txItrs, txItr)
// 	}

// 	return hB, txItrs, nil
// }

/*
加载数据库中的初始块
*/
func LoadStartBlock() *BlockHeadVO {
	exist, err := db.LevelDB.CheckHashExist(config.Key_block_start)
	if err != nil {
		engine.Log.Info("load start block hash error:%s", err.Error())
		return nil
	}
	if !exist {
		engine.Log.Info("This is an empty database")
		return nil
	}

	headid, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		//认为这是一个空数据库
		engine.Log.Info("This is an empty database")
		return nil
	}
	bh, txItrs, err := loadBlockForDB(headid)
	if err != nil {
		return nil
	}
	bhvo := BlockHeadVO{
		BH:  bh,     //区块
		Txs: txItrs, //交易明细
	}
	return &bhvo
}
