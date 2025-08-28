package mining

import (
	"encoding/hex"
	"runtime"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/protos/go_protos"

	"web3_gui/libp2parea/adapter"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/utils"
)

// 深度同步多块
func (this *Chain) deepCycleSyncBlockMore(bhvo *BlockHeadVO, c <-chan time.Time, height uint64, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
	//这里查看内存，控制速度
	//memInfo, _ := mem.VirtualMemory()
	//if memInfo.UsedPercent > config.Wallet_Memory_percentage_max {
	//	runtime.GC()
	//	time.Sleep(time.Second)
	//}
	//改良控制速度
	var m runtime.MemStats

	bhash := &bhvo.BH.Nextblockhash
	// fmt.Println("本次同步hash", hex.EncodeToString(*bhash))
	if bhash == nil || len(*bhash) <= 0 {
		// engine.Log.Warn("查询的下一个块hash为空")
		engine.Log.Warn("The next block hash of the query is empty")
		return bhvo, nil
	}

	for {
		req := &go_protos.GetBlockHeadVOMoreRequest{
			Height: bhvo.BH.Height + 1,
			Length: config.Deep_Cycle_Sync_Block_More_Length_client,
			BHash:  bhvo.BH.Nextblockhash,
		}
		//engine.Log.Info("要请求同步的起始高度: %d ,区块hash: %v", req.Height, req.BHash)
		reqMsg, err := req.Marshal()
		if err != nil {
			engine.Log.Warn("GetBlockHeadVOMoreRequest Marshal Err")
			return bhvo, err
		}

		bhvos, err := this.syncBlockForDBAndNeighborMore(&reqMsg, peerBlockInfo)
		if err != nil {
			engine.Log.Info("Error synchronizing block: %s", err.Error())
			return bhvo, err
		}

		runtime.ReadMemStats(&m)
		//拉起过程中,内存消耗过高,限制在最小配置内存
		if m.Sys >= uint64(config.MinFreeMemory)*1024*1024 && config.MinFreeMemory > 0 {
			runtime.GC()
			time.Sleep(time.Millisecond * 10)
		}

		for _, one := range bhvos {
			//检查hash是否修改顺序
			if blockHashs, ok := config.BlockHashsMap[hex.EncodeToString(one.BH.Hash)]; ok {
				peerBlockinfo, _ := FindRemoteCurrentHeight()
				if currentbhvo := this.testLoadBlock(blockHashs, peerBlockinfo); currentbhvo != nil {
					bhvo = currentbhvo
				}
				break
			}

			bhvo = &BlockHeadVO{FromBroadcast: false, BH: one.BH, Txs: one.Txs}
			if err = this.AddBlockOther(bhvo); err != nil {
				if err.Error() == ERROR_repeat_import_block.Error() {
					//可以重复导入区块
				} else {
					engine.Log.Info("deepCycleSyncBlock error: %s", err.Error())
					return bhvo, err
				}
			}
		}
		//没有下一个区块时退出
		if &bhvo.BH.Nextblockhash == nil || len(bhvo.BH.Nextblockhash) <= 0 {
			//engine.Log.Info("-------------------------最后一个区块 %d", bhvo.BH.Height)
			break
		}
	}
	//同步最新高度
	if GetHighestBlock() < bhvo.BH.Height {
		SetHighestBlock(bhvo.BH.Height)
	}

	//定时同步区块最新高度
	select {
	case <-c:
		// go FindBlockHeight()
		utils.Go(FindBlockHeight, nil)
	default:
	}

	return bhvo, nil
}

/*
从数据库查询区块，如果数据库没有，从网络邻居节点查询区块
查询到区块后，修改他们的指向hash值和UTXO输出的指向
*/
func (this *Chain) syncBlockForDBAndNeighborMore(reqMsg *[]byte, peerBlockInfo *PeerBlockInfoDESC) ([]*BlockHeadVO, error) {
	//此注释代码会导致区块同步中断。本地查到区块，next区块值为null，导致同步中断。

	//再查找邻居节点
	bhvos, err := FindBlockForNeighborMore(reqMsg, peerBlockInfo)
	if err != nil {
		engine.Log.Error("find next block error:%s", err.Error())
		return nil, err
	}
	if bhvos == nil || len(bhvos) <= 0 {
		//同步失败，未找到区块
		engine.Log.Error("find next block fail")
		return nil, config.ERROR_chain_sysn_block_fail
	}

	for _, bhvo := range bhvos {
		//保存区块中的交易
		for i, _ := range bhvo.Txs {
			bhvo.Txs[i].BuildHash()
			// bhvo.Txs[i].SetBlockHash(*bhash)
			// bs, err := bhvo.Txs[i].Json()
			bs, err := bhvo.Txs[i].Proto()
			if err != nil {
				//TODO 严谨的错误处理
				// fmt.Println("严重错误1", err)
				engine.Log.Error("load tx error:%s", err.Error())
				return nil, err
			}
			//			fmt.Println("保存交易", hex.EncodeToString(*bhvo.Txs[i].GetHash()))
			txhashkey := config.BuildBlockTx(*bhvo.Txs[i].GetHash())
			db.LevelDB.Save(txhashkey, bs)

		}

		//先将前一个区块修改next
		if this.GetStartingBlock() > config.Mining_block_start_height {
			// bs, err := db.Find(bhvo.BH.Previousblockhash)
			// if err != nil {
			// 	//TODO 区块未同步完整可以查找不到之前的区块
			// 	return nil, nil, err
			// }
			// bh, err := ParseBlockHead(bs)
			// if err != nil {
			// 	// fmt.Println("严重错误5", err)
			// 	return nil, nil, err
			// }

			bh, err := LoadBlockHeadByHash(&bhvo.BH.Previousblockhash)
			if err != nil {
				engine.Log.Error("load blockhead error:%s", err.Error())
				return nil, err
			}

			bh.Nextblockhash = bhvo.BH.Hash

			// if bh.Nextblockhash == nil {
			// 	engine.Log.Error("save block nextblockhash nil %s", string(*bs))
			// }

			// bs, err = bh.Json()
			bs, err := bh.Proto()
			if err != nil {
				// fmt.Println("严重错误6", err)
				engine.Log.Error("parse blockhead error:%s", err.Error())
				return nil, err
			}

			bhashkey := config.BuildBlockHead(bh.Hash)
			db.LevelDB.Save(bhashkey, bs)
		}

		//保存区块
		// bs, err := bhvo.BH.Json()
		bs, err := bhvo.BH.Proto()
		if err != nil {
			//TODO 严谨的错误处理
			// fmt.Println("严重错误7", err)
			engine.Log.Error("parse blockhead error:%s", err.Error())
			return nil, err
		}
		// if bhvo.BH.Nextblockhash == nil {
		// 	engine.Log.Error("save block nextblockhash nil %s", string(*bs))
		// }

		bhashkey := config.BuildBlockHead(bhvo.BH.Hash)
		db.LevelDB.Save(bhashkey, bs)

		// engine.Log.Info("get block info %s", string(*bs))

	}

	return bhvos, nil
}

/*
从邻居节点查询区块头和区块中的交易
*/
func FindBlockForNeighborMore(reqMsg *[]byte, peerBlockInfo *PeerBlockInfoDESC) ([]*BlockHeadVO, error) {
	var bhvo []*BlockHeadVO
	var bs *[]byte
	var err error
	var newBhvo []*BlockHeadVO

	//根据超时时间排序
	// logicNodes := nodeStore.GetLogicNodes()
	// logicNodes = append(logicNodes, nodeStore.GetNodesClient()...)
	// logicNodesInfo := libp2parea.SortNetAddrForSpeed(logicNodes)
	//工作模式，false=询问未超时节点;true=询问超时节点;
	//如果不切换工作模式，会导致节点同步永远落后几个区块高度

	// mode := false
	// for i := 0; i < 2; i++ {
	// engine.Log.Info("节点数量:%d", len(peerBlockInfo.Peers))
	// TAG:
	peers := peerBlockInfo.Sort()
	// engine.Log.Info("节点数量:%d", len(peers))
	addrs := make([]nodeStore.AddressNet, 0)
	for _, one := range peers {
		addrs = append(addrs, *one.Addr)
	}
	logicNodesInfo := libp2parea.SortNetAddrForSpeed(addrs)
	// engine.Log.Info("节点数量:%d", len(logicNodesInfo))
	for i, _ := range logicNodesInfo {
		//询问未超时节点工作模式下：遇到超时的节点，则退出
		// if !mode && one.Speed >= int64(time.Second*config.Wallet_sync_block_timeout) {
		// 	continue
		// }
		//询问超时节点工作模式下：遇到未超时的节点，则退出
		// if mode && one.Speed < int64(time.Second*config.Wallet_sync_block_timeout) {
		// 	continue
		// }
		key := &logicNodesInfo[i].AddrNet
		// engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
		engine.Log.Info("Send query message to node %s %s", key.B58String(), hex.EncodeToString(*reqMsg))
		bs, err = getBlockHeadVOMore(*key, reqMsg)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
			continue
		}
		if bs == nil {
			engine.Log.Info("Send query message to node from:%s bs is nil", key.B58String())
			continue
		}
		// engine.Log.Info("bs长度:%d", len(*bs))
		newBhvo, err = ParseBlockHeadVOProtoMore(bs)
		// newBhvo, err = ParseBlockHeadVO(bs)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
			continue
		}
		// engine.Log.Info("获取到的区块:%+v", newBhvo)
		//没有拿到
		if len(newBhvo) < 1 {
			continue
		}
		// engine.Log.Info("sync block info:%+v", newBhvo)
		bhvo = newBhvo
		//检查本区块是否有nextHash
		if newBhvo[len(newBhvo)-1].BH.Nextblockhash != nil && len(newBhvo[len(newBhvo)-1].BH.Nextblockhash) > 0 {
			// engine.Log.Info("this block next block hash not nil")
			return newBhvo, err
		}
		engine.Log.Info("this block next block hash is nil")
		//为空也返回
		// return newBhvo, nil
	}
	//如果从，未超时的节点同步到区块，则不继续同步了
	if bhvo != nil {
		return bhvo, nil
	}
	//TODO 超时的节点会永久打入冷宫，给超时的节点一个翻身的机会
	//如果从，未超时的节点，都没同步到区块，则尝试已超时的节点。
	// if !mode {
	// 	engine.Log.Info("switch sync timeout nodes mode")
	// 	mode = true
	// 	// goto TAG
	// 	continue
	// }
	// }
	// if bs != nil {
	// 	engine.Log.Info("this block nextblock nil %s", string(*bs))
	// }
	return bhvo, err
}

/*
查询邻居节点数据库，key：value查询
*/
func getBlockHeadVOMore(key nodeStore.AddressNet, reqMsg *[]byte) (*[]byte, error) {

	start := config.TimeNow()

	//pl time
	//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getBlockHeadVO, &key, bhash, config.Mining_block_time*time.Second)
	bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getBlockHeadVOMore, &key, reqMsg, config.Mining_sync_timeout)
	if err != nil {
		engine.Log.Info("getBlockHeadVO error:%s", err.Error())
		return nil, err
	}

	// message, _ := message_center.SendNeighborMsg(config.MSGID_getBlockHeadVO, &key, bhash)
	// // engine.Log.Info("44444444444 %s", key.B58String())
	// // bs := flood.WaitRequest(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), config.Mining_block_time)
	// bs, _ := flood.WaitRequest(message_center.CLASS_getTransaction, utils.Bytes2string(message.Body.Hash), config.Wallet_sync_block_timeout)
	if bs == nil {
		endTime := config.TimeNow()
		// engine.Log.Info("5555555555555555 %s", key.B58String())
		//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
		engine.Log.Error("Receive %s message timeout %s", key.B58String(), config.TimeNow().Sub(start))
		//有可能是对方没有查询到区块，返回空，则判定它超时
		if (endTime.Unix() - start.Unix()) < config.Wallet_sync_block_timeout {
			libp2parea.AddNodeAddrSpeed(key, time.Second*(config.Wallet_sync_block_timeout+1))
		} else {
			//TODO 应该取平均数
			//保存上一次同步超时时间
			// config.NetSpeedMap.Store(utils.Bytes2string(key), config.TimeNow().Sub(start))
			libp2parea.AddNodeAddrSpeed(key, config.TimeNow().Sub(start))
		}
		// err = errors.New("Failed to send shared file message, may timeout")

		return nil, config.ERROR_chain_sync_block_timeout
	}
	libp2parea.AddNodeAddrSpeed(key, config.TimeNow().Sub(start))
	// engine.Log.Info("Receive message %s", key.B58String())
	return bs, nil
}
