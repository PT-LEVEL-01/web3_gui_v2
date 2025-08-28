package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/mining/restart"

	"github.com/shirou/gopsutil/v3/mem"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/libp2parea/adapter/nodeStore"
	"web3_gui/utils"
)

//同步保存区块头队列
// var syncSaveBlockHead = make(chan *BlockHeadVO, 1)

// 触发同步区块信号
var syncSignalChain = make(chan bool, 1)

// 开始同步区块信号
var syncForNeighborChain = make(chan bool, 1)

//var syncHeightBlock = new(sync.Map)
//var heightBlockGroup = new(sync.WaitGroup)

/*
通知加载区块到内存的信号
*/
func (this *Chain) NoticeLoadBlockForDB() {
	// if atomic.LoadUint32(&this.BlockHighStopGrowth) == 2 {
	// 	return
	// }
	// if atomic.LoadUint32(&this.BlockHighStopGrowth) == 1 {
	// 	atomic.StoreUint32(&this.BlockHighStopGrowth, 2)
	// }
	select {
	case syncSignalChain <- false:
		engine.Log.Info("Put in sync queue")
	default:
	}
}

/*
循环检测区块高度是否增长
*/
func (this *Chain) GoForkCheck() {
	//5分钟检查一次高度是否增长
	//增长直接进入下一个5分钟
	//不增长this.ForkCheck(&bh.Hash)，下一次直接一小时后检查
	interval := config.Mining_block_time * 5
	if interval < time.Second*20 {
		interval = time.Second * 20
	}

	moderateTimer := time.NewTimer(time.Hour) //一小时后减速同步
	timer := time.NewTicker(interval)         //不少于20秒要检查高度是否增长

	count := 0
	countLevel_1 := 100
	countMax := 200
	// interval := time.Minute * 5
	beforBlockHeight := uint64(0)
	isSync := false
	for {
		// if this == nil || this.Temp == nil {
		// 	//engine.Log.Warn("GetLongChain为空11111111111111111111111111111111111111111111111111111111111111111111")
		// 	time.Sleep(time.Minute * 5)
		// 	continue
		// }
		isSync = false
		beforBlockHeight = this.GetCurrentBlock()
		if count < countMax {
			select {
			case <-syncSignalChain:
				engine.Log.Info("接收到同步信号")
				if count > countLevel_1 {
					engine.Log.Info("接收到同步信号")
					continue
				}
				isSync = true
			case <-timer.C:
				//高度在增长，不需要同步
				if this.GetCurrentBlock() > beforBlockHeight {
					continue
				}
				isSync = true
			}
		} else {
			<-moderateTimer.C
			isSync = true
		}

		if !isSync {
			engine.Log.Info("接收到同步信号")
			continue
		}
		engine.Log.Info("接收到同步信号")
		//执行一次同步
		this.SyncBlockOnce()

		//检查本次同步是否成功
		if this.GetCurrentBlock() > beforBlockHeight {
			//成功
			count = 0
		} else {
			count++
		}

		moderateTimer.Reset(time.Hour)
		timer.Reset(interval)
		continue

		select {
		case syncForNeighborChain <- false:
		default:
		}

		// time.Sleep(interval)
		// currentHeight := this.GetCurrentBlock()
		// if currentHeight <= beforBlockHeight {
		// 	//engine.Log.Warn("currentHeight <= beforBlockHeight1小时后再来22222222222222222222222222222222222222222222222222222222222222")
		// 	atomic.StoreUint32(&this.BlockHighStopGrowth, 1)
		// 	this.ForkCheck(nil)
		// 	interval = time.Hour
		// } else {
		// 	//engine.Log.Warn("区块正常，5分钟后检查！33333333333333333333333333333333333333333333333333333333333333")
		// 	atomic.StoreUint32(&this.BlockHighStopGrowth, 0)
		// 	interval = time.Minute * 5
		// }
	}
}

/*
链重启模块
*/
func (this *Chain) GoChainRestart() {
	//是否启用自重启功能
	if !config.EnableRestart {
		return
	}

	//记录检测高度
	checkHeight := this.CurrentBlock
	//检测自己节点高度定时器,默认1min
	checkHeightDur := config.Mining_block_time * 60
	checkHeightTicker := time.NewTicker(checkHeightDur)
	//重启定时器,默认30min
	restartDur := config.Mining_block_time * 1800
	restartTimer := time.NewTimer(restartDur)

	engine.Log.Info("Chain Enable Restart Feature Interval %s", restartDur)

	for {
		select {
		case <-checkHeightTicker.C:
			if int64(this.CurrentBlock)-int64(checkHeight) > config.Mining_group_max || !this.SyncBlockFinish {
				//检测自己区块高度是否增长,增长则重置重启定时器
				restartTimer.Stop()
				restartTimer.Reset(restartDur)
				checkHeight = this.CurrentBlock
			}
		case <-restartTimer.C:
			peerInfos, maxRemoteHeight := FindRemoteCurrentHeight()
			if int64(maxRemoteHeight)-int64(checkHeight) < config.Mining_group_max {
				return
			}

			//启动节点参数
			args := ""
			if peerInfos.Len() == 0 || maxRemoteHeight == 0 {
				//未取得邻节点的信息则不重启, 避免遇到所有节点均断网的情况
				return
			}

			engine.Log.Info("Got Chain Restart Signal SelfHeight %d RemoteHeight %d", checkHeight, maxRemoteHeight)

			//暂停同步区块
			atomic.StoreUint32(&this.StopSyncBlock, 1)
			//重启前自动修复区块哈希顺序修复
			this.AutoFixBlockNextHash()

			//通知重启
			if err := restart.NotifyRestart(args); err != nil {
				engine.Log.Error("NotifyRestart Fail: %v", err)
			}
		}
	}
}

/*
节点启动后，首次同步区块
*/
func (this *Chain) FirstDownloadBlock() error {
	// engine.Log.Info("节点启动后，首次同步区块 %v", config.LoadNode)
	engine.Log.Info("After the node is started, the block is synchronized for the first time %v", config.LoadNode)
	// if config.InitNode {
	// 	return nil
	// }
	//如果是拉起首节点，需要确认未确认的区块
	// if config.LoadNode {
	// 	finishFirstLoadBlockChain()
	// 	return nil
	// }
	greenChannel := false //本次同步不需要添加错误次数
	var err error
	count := 0
	this.SyncBlockLock.Lock()
	this.LoadBlockChain()
	oldCurrentBlockHeight := uint64(0)
	total := 0
	for {
		count++
		// engine.Log.Info("第几轮同步区块 %d", count)
		engine.Log.Info("Which round of synchronization block %d", count)
		currentBlock := GetLongChain().GetCurrentBlock()
		if greenChannel {
			greenChannel = false
		} else {
			if currentBlock > oldCurrentBlockHeight {
				total = 0
				oldCurrentBlockHeight = currentBlock
			} else {

				total++
				if total > 3 {
					////避免造成区块都导入不进去，还一直同步，占用带宽
					//engine.Log.Info("FirstDownloadBlock fail")
					////获取邻居节点最高区块
					//peerBlockinfo, _ := FindRemoteCurrentHeight()
					//bh, err := FindLastBlockForNeighbor(peerBlockinfo)
					//if err != nil || bh == nil {
					//	break
					//}
					////对方区块高度还没有自己高的时候，不检查分叉
					//if bh.Height <= currentBlock {
					//	break
					//}
					//engine.Log.Error("ForkCheck")
					//this.ForkCheck(&bh.Hash)
					break
				}
			}
		}
		peerBlockinfo, remoteCuurentHeightMax := FindRemoteCurrentHeight()
		//engine.Log.Info("当前已经同步到的区块高度currentBlock：%d 邻居节点已经导入的区块高度remoteCuurentHeightMax： %d", currentBlock, remoteCuurentHeightMax)
		if currentBlock >= remoteCuurentHeightMax {
			engine.Log.Info("FirstDownloadBlock finish")
			break
		}
		err = this.SyncBlockHead(peerBlockinfo)
		//TODO 导入区块失败，或者重复导入区块的情况，应该退出循环，因为区块同步卡住了
		if err == nil {
			continue
		} else {
			engine.Log.Info("FirstDownloadBlock error:%s", err.Error())
			if err.Error() == config.ERROR_chain_sync_block_timeout.Error() || err.Error() == config.ERROR_wait_msg_timeout.Error() {
				time.Sleep(time.Second * 60)
				greenChannel = true
				continue
			}
		}
		time.Sleep(time.Second * 11)
	}
	this.SyncBlockLock.Unlock()
	//engine.Log.Info("共用了几轮同步区块完成 %d", count)
	// engine.Log.Info("Shared several rounds of synchronization blocks to complete %d", i)
	FinishFirstLoadBlockChain()
	// config.EnableCache = true
	return err
}
func (this *Chain) GoSyncBlock() {
	// fmt.Println("--开始循环接收同步链端消息")

	utils.Go(func() {
		//创始节点启动需要一次加载
		// tag := false
		//beforBlockHeight := uint64(0)
		//stopBlockHeightTotal := 0
		//greenChannel := false //本次同步不需要添加错误次数
		// l := new(sync.RWMutex)
		for range syncForNeighborChain {
			this.SyncBlockOnce()
		}
	}, nil)
}

/*
同步一次区块
*/
func (this *Chain) SyncBlockOnce() {
	engine.Log.Info("111111111111111")
	currentHeight := GetLongChain().GetCurrentBlock()

	_, _, isKickOut, _, _ := GetWitnessStatus()
	// engine.Log.Info("是否按时出块 %v", isKickOut)
	if isKickOut {
		engine.Log.Info("isKickOut")
		return
	}
	engine.Log.Info("111111111111111")
	// l.Lock()
	this.SyncBlockLock.Lock()
	// fmt.Println("------------- start")
	this.LoadBlockChain()
	peerBlockinfo, _ := FindRemoteCurrentHeight()
	engine.Log.Info("节点数量:%d", len(peerBlockinfo.Peers))
	peers := peerBlockinfo.Sort()
	if len(peers) <= 0 || peers[0].CurrentHeight <= currentHeight {
		engine.Log.Info("等待一分钟 start")
		// time.Sleep(time.Second * 60)
		//greenChannel = true
		this.SyncBlockLock.Unlock()
		engine.Log.Info("等待一分钟 end")
		return
	}
	engine.Log.Info("111111111111111")
	err := this.SyncBlockHead(peerBlockinfo)
	if err != nil && (err.Error() == config.ERROR_chain_sync_block_timeout.Error() ||
		err.Error() == config.ERROR_wait_msg_timeout.Error()) {
		// time.Sleep(time.Second * 60)
		//greenChannel = true
		this.SyncBlockLock.Unlock()
		return
	}
	engine.Log.Info("222222222222222222222222")
	//同步出块时间
	this.WitnessChain.StopAllMining()
	this.WitnessChain.BuildMiningTime()
	// fmt.Println("------------- end")
	engine.Log.Info("111111111111111")
	// tag = true
	// config.AlreadyMining = true

	// l.Unlock()
	this.SyncBlockLock.Unlock()
	engine.Log.Info("111111111111111")
}

/*
处理分叉和同步卡住的情况
*/
func (this *Chain) ForkCheck(bhash *[]byte) {
	if atomic.LoadUint32(&this.StopSyncBlock) == 1 {
		return
	}
	atomic.StoreUint32(&this.StopSyncBlock, 1)

	engine.Log.Error("ForkCheck")
	//从最高高度，往前查找，重新建立前后区块关系
	// temp := this.Temp
	if bhash == nil && this.Temp == nil {
		return
	}
	// this.StopSyncBlock = true
	// atomic.StoreUint32(&this.StopSyncBlock, 1)
	var preBlockHash []byte
	if bhash != nil {
		preBlockHash = *bhash
	} else {
		engine.Log.Info("%v", this.Temp.BH)
		preBlockHash = this.Temp.BH.Previousblockhash
	}
	var nextBlockHash []byte
	for {
		engine.Log.Info("load block:%s", hex.EncodeToString(preBlockHash))
		//检查数据库是否存在这个区块
		bh, err := LoadBlockHeadByHash(&preBlockHash)
		if err != nil || bh == nil {
			//本地数据库没有，就从邻居节点下载
			peerBlockinfo, _ := FindRemoteCurrentHeight()
			bhvo, err := SyncBlockFlashDB(&preBlockHash, peerBlockinfo)
			if err != nil {
				engine.Log.Error("SyncBlockHead error:%s", err.Error())
				return
			}
			bh = bhvo.BH
		}

		// db.Find()

		engine.Log.Info("check fork height:%d", bh.Height)
		//本地已经存在这个区块，判断这个区块是否是链上收录的块。是收录的块则找到这个分叉点了
		//bhash, err := db.LevelDB.Find([]byte(config.BlockHeight + strconv.Itoa(int(bh.Height))))
		bhash, err := db.LevelDB.Find(append(config.BlockHeight, []byte(strconv.Itoa(int(bh.Height)))...))
		if err == nil && bytes.Equal(*bhash, bh.Hash) {
			engine.Log.Info("fork block:%s", hex.EncodeToString(*bhash))
			//把此区块next刷新一次
			//nextBlockHash存在，则刷新本地数据库
			if nextBlockHash != nil && len(nextBlockHash) > 0 {
				bh.Nextblockhash = nextBlockHash
				bs, err := bh.Proto()

				bhashkey := config.BuildBlockHead(bh.Hash)
				err = db.LevelDB.Save(bhashkey, bs)
				if err != nil {
					engine.Log.Error("save block error:%s", err.Error())
					return
				}
			}

			//尝试修复快照分叉情况
			this.FixChainForkForSnapshot(bh.GroupHeight)

			break
		}

		preBlockHash = bh.Previousblockhash
		nextBlockHash = bh.Hash

		// if this.witnessChain.FindBlockInCurrent(bh) {
		// 	break
		// }
	}

	engine.Log.Error("ForkCheck finish!")
}

/*
自动修复区块哈希顺序修复
*/
func (this *Chain) AutoFixBlockNextHash() {
	//邻节点最高区块信息
	peerBlockinfo, _ := FindRemoteCurrentHeight()
	remoteBlockInfo, err := FindLastBlockForNeighbor(peerBlockinfo)
	if err != nil {
		engine.Log.Error("AutoFixBlockNextHash Error: %v", err)
		return
	}

	//从邻节点最高区块开始修复节点nexthash
	var preBlockHash = remoteBlockInfo.Hash
	var nextBlockHash = []byte{}
	for {
		bhvo, err := SyncBlockFlashDB(&preBlockHash, peerBlockinfo)
		if err != nil {
			engine.Log.Error("AutoFixBlockNextHash SyncBlockFlashDB Error: %v", err)
			return
		}

		bhash, err := db.LevelDB.Find(append(config.BlockHeight, []byte(strconv.Itoa(int(bhvo.BH.Height)))...))
		if err == nil && bytes.Equal(*bhash, bhvo.BH.Hash) {
			//把此区块next刷新一次
			//nextBlockHash存在，则刷新本地数据库
			if nextBlockHash != nil && len(nextBlockHash) > 0 {
				bhvo.BH.Nextblockhash = nextBlockHash
				if bs, err := bhvo.BH.Proto(); err == nil {
					bhashkey := config.BuildBlockHead(bhvo.BH.Hash)
					db.LevelDB.Save(bhashkey, bs)
				}
			}

			this.FixChainForkForSnapshot(bhvo.BH.GroupHeight)
			break
		}
		if bhvo.BH.Height%100 == 0 {
			//打印少量日志
			engine.Log.Info("AutoFixBlockNextHash Fixed Height: %d", bhvo.BH.Height)
		}
		preBlockHash = bhvo.BH.Previousblockhash
		nextBlockHash = bhvo.BH.Hash
	}
}

/*
保存一个区块，返回这个区块是否连续，判断依据是能够找到并修改前置区块
@return    bool    是否找到前置区块
*/
func SaveBlockHead(bhvo *BlockHeadVO) (bool, error) {
	bhvo.BH.BuildBlockHash()
	ok := true

	//保存区块中的交易
	SaveBhvoTxs(bhvo)

	//查询前一个区块
	bhashkey := config.BuildBlockHead(bhvo.BH.Previousblockhash)
	_, err := db.LevelDB.Find(bhashkey)
	// if err == nil {
	// 	bh, err := ParseBlockHead(bs)
	if err != nil {
		// fmt.Println("查询前置区块错误", err)
		ok = false
	}

	//保存区块
	// bs, err := bhvo.BH.Json()
	bs, err := bhvo.BH.Proto()
	if err != nil {
		//TODO 严谨的错误处理
		// fmt.Println("格式化新区块错误", err)
		// ok = false
		return false, err
	}

	// TxCache.AddBlockHeadCache(hex.EncodeToString(bhvo.BH.Hash), bhvo.BH)
	bhashkey = config.BuildBlockHead(bhvo.BH.Hash)
	TxCache.AddBlockHeadCache(bhashkey, bhvo.BH)

	//先查询数据库是否存在，不存在则保存
	if ok, _ := db.LevelDB.CheckHashExist(bhashkey); !ok {
		// if bhvo.BH.Nextblockhash == nil {
		// 	engine.Log.Error("save block nextblockhash nil %s", string(*bs))
		// }

		err = db.LevelDB.Save(bhashkey, bs)
		if err != nil {
			//TODO 严谨的错误处理
			// fmt.Println("保存新区块错误", err)
			// ok = false
			return false, err
		}
	}

	// 提交区块增量数据
	err = db.LevelDB.Commit(bhashkey)
	if err != nil {
		return false, err
	}

	return ok, nil
}

/*
保存一个区块中的交易所属的区块hash
*/
func SaveTxToBlockHead(bhvo *BlockHeadVO) error {
	bhvo.BH.BuildBlockHash()
	var err error
	//保存区块中的交易
	for i, _ := range bhvo.Txs {
		one := bhvo.Txs[i]
		one.BuildHash()
		err = db.SaveTxToBlockHash(one.GetHash(), &bhvo.BH.Hash)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
保存未确认的交易，其中有个标记
*/
func SaveTempTx(txItr TxItr) error {
	// engine.Log.Info("保存交易 %s %s", hex.EncodeToString(*txItr.GetHash()), hex.EncodeToString(*txItr.GetBlockHash()))
	// bs, err := txItr.Json()
	bs, err := txItr.Proto()
	if err != nil {
		//TODO 严谨的错误处理
		// fmt.Println("Save block error, transaction JSON format error", err)
		// ok = false
		return err
	}
	// engine.Log.Info("保存交易 %s", string(*bs))
	// AddTxInCache(hex.EncodeToString(*bhvo.Txs[i].GetHash()), bhvo.Txs[i])

	txhashkey := config.BuildBlockTx(*txItr.GetHash())
	err = db.LevelDB.Save(txhashkey, bs)
	if err != nil {
		// fmt.Println("Save transaction error", bhvo.Txs[i].GetHash())
		return err
	}

	db.LevelDB.Save(config.BuildTxNotImport(*txItr.GetHash()), nil)
	return nil
}

/*
查询邻居节点区块高度
从邻居节点中查找最高区块高度
*/
func FindBlockHeight() {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	syncHeightBlock := new(sync.Map)

	//	heightBlockGroup = new(sync.WaitGroup)
	//	count := 0
	for _, key := range Area.NodeManager.GetLogicNodes() {
		sessionName := ""
		session, ok := Area.SessionEngine.GetSession(config.AreaName[:], utils.Bytes2string(key))
		if ok {
			sessionName = session.GetName()
		}

		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_heightBlock, &key, nil, config.Mining_block_time*time.Second)

		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_heightBlock, &key, &config.StartBlockHash, config.Mining_sync_timeout)
		if err != nil {
			engine.Log.Info("get block height:%s", err.Error())
			continue
		}
		if bs == nil {
			continue
		}

		chain := forks.GetLongChain()
		//				startHeight := binary.LittleEndian.Uint64((*bs)[:8])
		heightBlock := binary.LittleEndian.Uint64((*bs)[8:])
		//收到的区块高度比自己低，则不保存
		if chain != nil && chain.GetCurrentBlock() > heightBlock {
			continue
		}

		if GetHighestBlock() < heightBlock {
			SetHighestBlock(heightBlock)
		}

		syncHeightBlock.Store(sessionName, heightBlock)

		// message, err := mc.SendNeighborMsg(config.MSGID_heightBlock, &key, nil)
		// if err == nil {
		// 	// bs := flood.WaitRequest(mc.CLASS_findHeightBlock, hex.EncodeToString(message.Body.Hash), 0)
		// 	bs, _ := flood.WaitRequest(mc.CLASS_findHeightBlock, utils.Bytes2string(message.Body.Hash), 0)
		// 	//		fmt.Println("有消息返回了啊")
		// 	if bs == nil {
		// 		// fmt.Println("11111 发送共享文件消息失败，可能超时")
		// 		continue
		// 	}
		// 	chain := forks.GetLongChain()
		// 	//				startHeight := binary.LittleEndian.Uint64((*bs)[:8])
		// 	heightBlock := binary.LittleEndian.Uint64((*bs)[8:])
		// 	//收到的区块高度比自己低，则不保存
		// 	if chain.GetCurrentBlock() > heightBlock {
		// 		continue
		// 	}

		// 	if GetHighestBlock() < heightBlock {
		// 		SetHighestBlock(heightBlock)
		// 	}

		// 	syncHeightBlock.Store(sessionName, heightBlock)
		// }

	}
	//以下情况都返回
	//1.没有邻居节点。
	//2.查询邻居节点全部失败。
	//3.邻居节点都未同步完成。
	count := 0
	syncHeightBlock.Range(func(key, value interface{}) bool {
		count++
		return false //只要有数据就够了
	})
	if count <= 0 {
		return
	}

	//统计区块高度结果，给结果投票
	heightBlockVote := new(sync.Map)
	syncHeightBlock.Range(func(key, value interface{}) bool {
		//		fmt.Println("统计投票", key, value)
		height := value.(uint64)
		v, ok := heightBlockVote.Load(height)
		if ok {
			total := v.(uint64)
			heightBlockVote.Store(height, uint64(total+1))
		} else {
			heightBlockVote.Store(height, uint64(1))
		}
		return true
	})

	//查看区块高度投票结果，采用票数高的，票数都一样，采用区块高度最高的。
	heightBlockMax := uint64(0) //区块最高高度
	heightBlock := uint64(0)    //票数最高的区块高度
	heightTotal := uint64(0)    //最高票数
	// isEqual := false            //最高票数是否有不同的区块高度
	heightBlockVote.Range(func(k, v interface{}) bool {
		//		fmt.Println("投票结果", k, v)
		height := k.(uint64)
		if height == 0 {
			return true
		}
		if height > heightBlockMax {
			heightBlockMax = height
		}
		total := v.(uint64)
		if total > heightTotal {
			heightTotal = total
			heightBlock = height
		} else if total == heightTotal {
			// isEqual = true
		}
		return true
	})
	//TODO 考虑相同票数该选哪个
	//直接使用票数最多的区块高度
	//	atomic.StoreUint64(&chain.StartingBlock, 1)
	SetHighestBlock(heightBlock)
	// atomic.StoreUint64(&forks.GetLongChain().HighestBlock, heightBlock)
	// fmt.Println("收到的区块高度", heightBlock, "自己的高度", atomic.LoadUint64(&forks.CurrentBlock))

}

/*
查询邻居节点已经导入的区块高度
*/
func FindRemoteCurrentHeight() (*PeerBlockInfoDESC, uint64) {
	//先用1秒钟超时测试
	peerDesc, height := FindRemoteCurrentHeightWithTimeout(time.Second)
	if height != 0 {
		return peerDesc, height
	}
	//再用20秒超时
	return FindRemoteCurrentHeightWithTimeout(config.Mining_sync_timeout)
}

/*
查询邻居节点已经导入的区块高度,带超时参数
*/
func FindRemoteCurrentHeightWithTimeout(timeout time.Duration) (*PeerBlockInfoDESC, uint64) {
	remoteCuurentHeightMax := uint64(0) //邻居节点最高高度
	// syncHeightBlock := new(sync.Map)
	peers := make([]*PeerBlockInfo, 0)

	logicNodes := Area.NodeManager.GetLogicNodes()
	logicNodes = append(logicNodes, Area.NodeManager.GetNodesClient()...)
	// logicNodesInfo := core.SortNetAddrForSpeed(logicNodes)
	for i, _ := range logicNodes {
		key := logicNodes[i]
		engine.Log.Info("FindRemoteCurrentHeight to:%s", key.B58String())

		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_heightBlock, &key, nil, config.Mining_block_time*time.Second)

		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_heightBlock, &key, &config.StartBlockHash, timeout)
		if err != nil {
			engine.Log.Info("FindRemoteCurrentHeight error:%s", err.Error())
			continue
		}
		if bs == nil {
			engine.Log.Info("FindRemoteCurrentHeight bs is nil")
			continue
		}
		heightBlock := binary.LittleEndian.Uint64((*bs)[8:])
		//收到的区块高度比自己低，则不保存
		if remoteCuurentHeightMax > heightBlock {
			engine.Log.Info("FindRemoteCurrentHeight height low")
			continue
		}
		remoteCuurentHeightMax = heightBlock

		peers = append(peers, &PeerBlockInfo{
			Addr:          &key,
			CurrentHeight: heightBlock,
		})

		engine.Log.Info("FindRemoteCurrentHeight height:%d", heightBlock)

		// message, err := mc.SendNeighborMsg(config.MSGID_heightBlock, &key, nil)
		// if err != nil {
		// 	// >>>>>>> dev
		// 	// engine.Log.Info("FindRemoteCurrentHeight send error:%s", err.Error())
		// 	continue
		// }
		// // bs := flood.WaitRequest(mc.CLASS_findHeightBlock, hex.EncodeToString(message.Body.Hash), 0)
		// bs, _ := flood.WaitRequest(mc.CLASS_findHeightBlock, utils.Bytes2string(message.Body.Hash), 5)
		// //		fmt.Println("有消息返回了啊")
		// if bs == nil {
		// 	// fmt.Println("11111 发送共享文件消息失败，可能超时")
		// 	// engine.Log.Info("FindRemoteCurrentHeight timeout to:%s", key.B58String())
		// 	continue
		// }
		// // chain := forks.GetLongChain()
		// //				startHeight := binary.LittleEndian.Uint64((*bs)[:8])
		// heightBlock := binary.LittleEndian.Uint64((*bs)[8:])
		// //收到的区块高度比自己低，则不保存
		// if remoteCuurentHeightMax > heightBlock {
		// 	// engine.Log.Info("FindRemoteCurrentHeight height low")
		// 	continue
		// }
		// remoteCuurentHeightMax = heightBlock

		// peers = append(peers, &PeerBlockInfo{
		// 	Addr:          &key,
		// 	CurrentHeight: heightBlock,
		// })
		// // engine.Log.Info("FindRemoteCurrentHeight %s %d", key.B58String(), heightBlock)

		// // if GetHighestBlock() < heightBlock {
		// // 	SetHighestBlock(heightBlock)
		// // }

		// // syncHeightBlock.Store(sessionName, heightBlock)

	}
	// engine.Log.Info("节点数量:%d", len(peers))
	peersDESC := NewPeerBlockInfoDESC(peers)
	return peersDESC, remoteCuurentHeightMax

}

/*
查询邻居节点已经导入的区块高度,带超时参数
*/
func FindRemoteCurrentHeightWithPeersAndTimeout(logicNodes []nodeStore.AddressNet, timeout time.Duration) (*PeerBlockInfoDESC, uint64) {
	remoteCuurentHeightMax := uint64(0) //邻居节点最高高度
	// syncHeightBlock := new(sync.Map)
	peers := make([]*PeerBlockInfo, 0)

	// logicNodesInfo := core.SortNetAddrForSpeed(logicNodes)
	for i, _ := range logicNodes {
		key := logicNodes[i]
		engine.Log.Info("FindRemoteCurrentHeight to:%s", key.B58String())

		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_heightBlock, &key, nil, config.Mining_block_time*time.Second)

		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_heightBlock, &key, &config.StartBlockHash, timeout)
		if err != nil {
			engine.Log.Info("FindRemoteCurrentHeight error:%s", err.Error())
			continue
		}
		if bs == nil {
			engine.Log.Info("FindRemoteCurrentHeight bs is nil")
			continue
		}
		heightBlock := binary.LittleEndian.Uint64((*bs)[8:])
		//收到的区块高度比自己低，则不保存
		if remoteCuurentHeightMax > heightBlock {
			engine.Log.Info("FindRemoteCurrentHeight height low")
			continue
		}
		remoteCuurentHeightMax = heightBlock

		peers = append(peers, &PeerBlockInfo{
			Addr:          &key,
			CurrentHeight: heightBlock,
		})

		engine.Log.Info("FindRemoteCurrentHeight height:%d", heightBlock)
	}
	// engine.Log.Info("节点数量:%d", len(peers))
	peersDESC := NewPeerBlockInfoDESC(peers)
	return peersDESC, remoteCuurentHeightMax

}

// var SyncBlockHead_Lock = new(sync.RWMutex)

/*
从邻居节点同步区块
*/
func (this *Chain) SyncBlockHead(peerBlockInfo *PeerBlockInfoDESC) error {
	engine.Log.Info("Start synchronizing blocks from neighbor nodes")
	//获得本节点的最新块hash
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	bhash := block.Id
	// engine.Log.Info("节点数量:%d", len(peerBlockInfo.Peers))
	//最新块一定要去邻居节点拉取一次，更新next
	bhvo, err := SyncBlockFlashDB(&bhash, peerBlockInfo)
	if err != nil {
		engine.Log.Error("SyncBlockHead error:%s", err.Error())
		return err
	}
	// engine.Log.Info("节点数量:%d", len(peerBlockInfo.Peers))
	if bhvo == nil {
		engine.Log.Error("SyncBlockHead finish")
		return nil
	}

	bhvo.BH.BuildBlockHash()
	engine.Log.Info("Print blocks synchronized to %d %s", bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Hash))

	tiker := time.NewTicker(time.Minute)

	//sh := bhvo.BH.Height + 1
	//st := config.TimeNow().UnixMilli()
	//bhvo, err = this.deepCycleSyncBlock(bhvo, tiker.C, bhvo.BH.Height+1, peerBlockInfo)
	bhvo, err = this.deepCycleSyncBlockMore(bhvo, tiker.C, bhvo.BH.Height+1, peerBlockInfo)
	//ut := config.TimeNow().UnixMilli() - st
	//uh := bhvo.BH.Height - sh
	//engine.Log.Info("222222222222222222222共同步 %d 个区块 用时 %d ms 速率 %v /ms", uh, ut, float64(uh)/float64(ut))

	tiker.Stop()
	if err != nil {
		engine.Log.Error("SyncBlockHead error:%s", err.Error())
		return err
	}
	//比较耗费时间，在这里做了
	this.Balance.Unfrozen(bhvo.BH.Height-1, bhvo.BH.Time)

	_, block = this.GetLastBlock()
	bhvos, err := GetUnconfirmedBlockForNeighbor(block.Height, peerBlockInfo)
	if err != nil {
		engine.Log.Error("GetUnconfirmedBlockForNeighbor error:%s", err.Error())
		return err
	}
	for _, one := range bhvos {
		engine.Log.Info("Import GetUnconfirmedBlockForNeighbor height:%d", one.BH.Height)
		one.FromBroadcast = false
		this.AddBlockOther(one)
	}

	this.SyncBlockFinish = true
	engine.Log.Info("Sync block complete")
	return nil
}

/*
深度循环同步区块，包括分叉的链的同步
加载到出错或者加载完成为止
@return    *BlockHeadVO    返回最后一个同步到的区块
*/
func (this *Chain) deepCycleSyncBlock(bhvo *BlockHeadVO, c <-chan time.Time, height uint64, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {

	//这里查看内存，控制速度
	memInfo, _ := mem.VirtualMemory()
	if memInfo.UsedPercent > config.Wallet_Memory_percentage_max {
		runtime.GC()
		time.Sleep(time.Second)
	}

	//临时改变nexthash
	// nextHash, ok := config.BlockNextHash.Load(utils.Bytes2string(bhvo.BH.Hash))
	// if ok {
	// 	nextHashBs := nextHash.(*[]byte)
	// 	bhvo.BH.Nextblockhash = *nextHashBs
	// }

	bhash := &bhvo.BH.Nextblockhash
	// fmt.Println("本次同步hash", hex.EncodeToString(*bhash))
	if bhash == nil || len(*bhash) <= 0 {
		// engine.Log.Warn("查询的下一个块hash为空")
		engine.Log.Warn("The next block hash of the query is empty")
		return bhvo, nil
	}

	start := config.TimeNow()
	// engine.Log.Info("Synchronize block from neighbor node this time %d hash %s", height, hex.EncodeToString(*bhash))
	bh, txItrs, err := this.syncBlockForDBAndNeighbor(bhash, peerBlockInfo)
	if err != nil {
		engine.Log.Info("Error synchronizing block: %s", err.Error())
		return bhvo, err
	}
	engine.Log.Info("下载耗时:%s", config.TimeNow().Sub(start))
	start = config.TimeNow()
	// engine.Log.Info("获取到的区块详细信息 %+v", bh)

	bhvo = &BlockHeadVO{FromBroadcast: false, BH: bh, Txs: txItrs}
	if err = this.AddBlockOther(bhvo); err != nil {
		if err.Error() == ERROR_repeat_import_block.Error() {
			//可以重复导入区块
		} else {
			engine.Log.Info("deepCycleSyncBlock error: %s", err.Error())
			return bhvo, err
		}
	}
	engine.Log.Info("导入耗时:%s", config.TimeNow().Sub(start))

	//同步最新高度
	if GetHighestBlock() < bh.Height {
		SetHighestBlock(bh.Height)
	}

	//定时同步区块最新高度
	select {
	case <-c:
		// go FindBlockHeight()
		utils.Go(FindBlockHeight, nil)
	default:
	}

	// fmt.Println("区块的next个数", len(bh.Nextblockhash), "高度", bh.Height)
	// engine.Log.Info("区块的next个数 %d 高度 %d", len(bh.Nextblockhash), bh.Height)
	// for _, one := range bh.Nextblockhash {
	// 	this.deepCycleSyncBlock(&one, c)
	// }
	engine.Log.Info("Next block %s", hex.EncodeToString(bh.Nextblockhash))

	return this.deepCycleSyncBlock(bhvo, c, bh.Height+1, peerBlockInfo)
}

/*
从数据库查询区块，如果数据库没有，从网络邻居节点查询区块
查询到区块后，修改他们的指向hash值和UTXO输出的指向
*/
func (this *Chain) syncBlockForDBAndNeighbor(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHead, []TxItr, error) {
	//此注释代码会导致区块同步中断。本地查到区块，next区块值为null，导致同步中断。

	//再查找邻居节点
	bhvo, err := FindBlockForNeighbor(bhash, peerBlockInfo)
	if err != nil {
		engine.Log.Error("find next block error:%s", err.Error())
		return nil, nil, err
	}
	if bhvo == nil {
		//同步失败，未找到区块
		engine.Log.Error("find next block fail")
		return nil, nil, config.ERROR_chain_sysn_block_fail
	}
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
			return nil, nil, err
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
			return nil, nil, err
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
			return nil, nil, err
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
		return nil, nil, err
	}
	// if bhvo.BH.Nextblockhash == nil {
	// 	engine.Log.Error("save block nextblockhash nil %s", string(*bs))
	// }
	bhashkey := config.BuildBlockHead(bhvo.BH.Hash)
	db.LevelDB.Save(bhashkey, bs)

	// engine.Log.Info("get block info %s", string(*bs))

	return bhvo.BH, bhvo.Txs, nil
}

/*
同步区块并刷新本地数据库
*/
func SyncBlockFlashDB(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
	// engine.Log.Info("节点数量:%d", len(peerBlockInfo.Peers))
	bhvo, err := FindBlockForNeighbor(bhash, peerBlockInfo)
	if err != nil {
		return nil, err
	}
	if bhvo == nil {
		return nil, config.ERROR_chain_sysn_block_fail
	}
	bhvo.BH.BuildBlockHash()
	// bs, err := bhvo.BH.Json()
	bs, err := bhvo.BH.Proto()
	if err != nil {
		return nil, err
	}
	// if bhvo.BH.Nextblockhash == nil {
	// 	engine.Log.Error("save block nextblockhash nil %s", string(*bs))
	// }
	bhashkey := config.BuildBlockHead(*bhash)
	db.LevelDB.Save(bhashkey, bs)
	for _, one := range bhvo.Txs {
		// for j, _ := range *one.GetVout() {
		// 	(*bhvo.Txs[i].GetVout())[j].Txid = nil
		// }
		// one.SetBlockHash(*bhash)
		// bs, err := one.Json()
		bs, err := one.Proto()
		if err != nil {
			return nil, err
		}
		txhashkey := config.BuildBlockTx(*one.GetHash())
		db.LevelDB.Save(txhashkey, bs)
	}
	return bhvo, nil
}

func GetRemoteTxAndSave(txid []byte) TxItr {
	txhashkey := config.BuildBlockHead(txid)
	bs := GetRemoteKeyFlashLocalKey(txhashkey)
	// txItr, err := ParseTxBase(ParseTxClass(txid), bs)
	txItr, err := ParseTxBaseProto(ParseTxClass(txid), bs)
	if err != nil {
		return nil
	}
	return txItr
}

/*
从邻居节点数据库查询一个key返回value，并保存或刷新到本地数据库
*/
func GetRemoteKeyFlashLocalKey(txid []byte) *[]byte {

	// engine.Log.Info("从隔壁节点获取一个交易")

	var bs *[]byte

	logicNodes := Area.NodeManager.GetLogicNodes()
	logicNodes = OrderNodeAddr(logicNodes)
	for _, key := range logicNodes {

		//pl time
		//bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getDBKey_one, &key, &txid, config.Mining_block_time*time.Second)
		bs, err := Area.SendNeighborMsgWaitRequest(config.MSGID_getDBKey_one, &key, &txid, config.Mining_sync_timeout)
		if err != nil {
			engine.Log.Error("Receive message error:%s %s", key.B58String(), err.Error())
			continue
		}

		// message, _ := message_center.SendNeighborMsg(config.MSGID_getDBKey_one, &key, &txid)
		// // engine.Log.Info("44444444444 %s", key.B58String())
		// bs, _ = flood.WaitRequest(mc.CLASS_getTransaction, utils.Bytes2string(message.Body.Hash), config.Mining_block_time)
		// // bs := flood.WaitRequest(mc.CLASS_getTransaction_one, utils.Bytes2string(message.Body.Hash), config.Mining_block_time)
		if bs == nil {
			// engine.Log.Info("5555555555555555 %s", key.B58String())
			//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
			engine.Log.Error("Receive message timeout %s", key.B58String())
			// err = errors.New("Failed to send shared file message, may timeout")

			continue
		}
		// engine.Log.Info("Receive message %s", key.B58String())
		// engine.Log.Info("获取的交易 %s", bs)
		txhashkey := config.BuildBlockTx(txid)
		db.LevelDB.Save(txhashkey, bs)
		break
	}
	return bs
}

type PeerBlockInfo struct {
	Addr          *nodeStore.AddressNet
	CurrentHeight uint64
}

/*
从大到小排序
*/
type PeerBlockInfoDESC struct {
	Peers []*PeerBlockInfo
}

func (this PeerBlockInfoDESC) Len() int {
	return len(this.Peers)
}

func (this PeerBlockInfoDESC) Less(i, j int) bool {
	if this.Peers[i].CurrentHeight < this.Peers[j].CurrentHeight {
		return false
	} else {
		return true
	}
}

func (this PeerBlockInfoDESC) Swap(i, j int) {
	this.Peers[i], this.Peers[j] = this.Peers[j], this.Peers[i]
}

func (this PeerBlockInfoDESC) Sort() []*PeerBlockInfo {
	if len(this.Peers) <= 0 {
		return nil
	}
	sort.Sort(this)
	//取前几个高度最高且相同的
	height := uint64(0)
	for i, one := range this.Peers {
		if i == 0 {
			height = one.CurrentHeight
			continue
		}
		// engine.Log.Info("判断高度:%d %d", one.CurrentHeight, height)
		if one.CurrentHeight < height {
			// engine.Log.Info("判断高度:%d", i)
			return this.Peers[:i]
		}
	}
	return this.Peers
}

func NewPeerBlockInfoDESC(peers []*PeerBlockInfo) *PeerBlockInfoDESC {
	return &PeerBlockInfoDESC{
		Peers: peers,
	}
}
