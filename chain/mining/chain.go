package mining

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/evm"
	"web3_gui/chain/mining/snapshot"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
	// "github.com/hyahm/golog"
)

/*
查找一个地址的余额
*/
func FindSurplus(addr utils.Multihash) uint64 {
	return 0
}

/*
查找最后一组旷工地址
*/
func FindLastGroupMiner() []utils.Multihash {
	return []utils.Multihash{}
}

type Chain struct {
	No               uint64 //分叉编号
	StartingBlock    uint64 //区块开始高度
	StartBlockTime   uint64 //起始区块时间
	CurrentBlock     uint64 //内存中已经同步到的区块高度
	CurrentBlockHash []byte //已经统计的最新高度区块hash
	PulledStates     uint64 //正在同步的区块高度
	HighestBlock     uint64 //网络节点广播的区块最高高度
	//GroupHeight        uint64              //最新导入区块的组高度
	SyncBlockLock          *sync.RWMutex           //同步区块锁
	SyncBlockFinish        bool                    //同步区块是否完成，关系能否接收新的区块
	WitnessBackup          *WitnessBackup          //备用见证人
	WitnessChain           *WitnessChain           //见证人组链
	Balance                *BalanceManager         //
	TransactionManager     *TransactionManager     //交易管理器
	transactionSwapManager *TransactionSwapManager //交易管理器
	// history            *BalanceHistory     //
	StopSyncBlock       uint32        //当区块分叉或出现重大错误时候，终止同步功能。初始值为0,1=暂停
	Temp                *BlockHeadVO  //收到的区块高度最大的区块
	signLock            *sync.RWMutex //签名锁
	BlockHighStopGrowth uint32        //块高度停止增长 0区块正常增长，1区块增长异常，2停止步同步区块
}

/*
获取区块开始高度
*/
func (this *Chain) GetStartingBlock() uint64 {
	if this == nil {
		return 0
		engine.Log.Error("!!!! chain is nil !!!!")
	}
	return atomic.LoadUint64(&this.StartingBlock)
}

/*
获取起始区块出块时间
*/
func (this *Chain) GetStartBlockTime() uint64 {
	return atomic.LoadUint64(&this.StartBlockTime)
}

/*
设置区块开始高度
*/
func (this *Chain) SetStartingBlock(n, startBlockTime uint64) {
	atomic.StoreUint64(&this.StartingBlock, n)
	atomic.StoreUint64(&this.StartBlockTime, startBlockTime)
}

/*
获取已经同步到的区块高度
*/
func (this *Chain) GetCurrentBlock() uint64 {
	return atomic.LoadUint64(&this.CurrentBlock)
}

/*
获取已经同步到的区块组高度
*/
// func (this *Chain) GetCurrentGroupHeight() uint64 {
// 	return atomic.LoadUint64(&this.CurrentGroup)
// }

/*
设置已经统计了的区块高度和组高度
*/
func (this *Chain) SetCurrentBlock(blockHeight uint64, bhash []byte) {
	atomic.StoreUint64(&this.CurrentBlock, blockHeight)
	this.CurrentBlockHash = bhash
}

/*
获取正在同步的区块高度
*/
func (this *Chain) GetPulledStates() uint64 {
	return atomic.LoadUint64(&this.PulledStates)
}

func (this *Chain) SetPulledStates(n uint64) {
	// engine.Log.Warn("设置正在同步的区块高度 %d", n)
	atomic.StoreUint64(&this.PulledStates, n)
}

/*
	获取区块组高度
*/
// func (this *Chain) GetGroupHeights() uint64 {
// 	// if forks.GetLongChain() == nil {
// 	// 	return 0
// 	// }
// 	return this.GetLastBlock().Group.Height
// }

func NewChain() *Chain {
	chain := &Chain{}
	chain.StopSyncBlock = 0
	chain.BlockHighStopGrowth = 0
	chain.SyncBlockLock = new(sync.RWMutex)
	chain.SyncBlockFinish = false
	// chain.No = no
	wb := NewWitnessBackup(chain)
	wc := NewWitnessChain(wb, chain)
	tm := NewTransactionManager(wb)
	tsm := NewTransactionSwapManager()
	b := NewBalanceManager(wb, tm, chain)

	// chain.lastBlock = block
	chain.WitnessBackup = wb
	chain.WitnessChain = wc
	chain.Balance = b
	chain.TransactionManager = tm
	chain.transactionSwapManager = tsm
	chain.signLock = new(sync.RWMutex)
	// chain.history = NewBalanceHistory()
	// go chain.GoSyncBlock()
	utils.Go(chain.GoSyncBlock, nil)
	utils.Go(chain.GoForkCheck, nil)
	//重启模块
	utils.Go(chain.GoChainRestart, nil)
	return chain
}

/*
	克隆一个链
*/
//func (this *Chain) Clone() *Chain {
//Chain{
//	witnessBackup : this.witnessBackup ,     *WitnessBackup      //备用见证人
//	witnessChain       *WitnessChain       //见证人组链
//	lastBlock          *Block              //最新块
//	balance            *BalanceManager     //
//	transactionManager *TransactionManager //交易管理器
//}
//}

type Group struct {
	PreGroup  *Group   //前置组
	NextGroup *Group   //下一个组，有分叉，下标为0的是最长链
	Height    uint64   //组高度
	Blocks    []*Block //组中的区块
}

type Block struct {
	Id         []byte   //区块id
	PreBlockID []byte   //前置区块id
	PreBlock   *Block   //前置区块高度
	NextBlock  *Block   //下一个区块高度
	Group      *Group   //所属组
	Height     uint64   //区块高度
	Witness    *Witness //是哪个见证人出的块
	isCount    bool     //是否被统计过了
	// IdStr      string   //
	// LocalTime time.Time //
}

// func (this *Block) GetIdStr() string {
// 	if this.IdStr == "" {
// 		this.IdStr = hex.EncodeToString(this.Id)
// 	}
// 	return this.IdStr
// }

func (this *Block) Load() (*BlockHead, error) {

	//先判断缓存中是否存在
	bhashkey := config.BuildBlockHead(this.Id)
	blockHead, ok := TxCache.FindBlockHeadCache(bhashkey)
	if !ok {
		// // engine.Log.Error("未命中缓存 111 FindBlockHeadCache")
		// bh, err := db.Find(this.Id)
		// if err != nil {
		// 	//		if err == leveldb.ErrNotFound {
		// 	//			return
		// 	//		} else {
		// 	//		}
		// 	return nil, err
		// }
		// blockHead, err = ParseBlockHead(bh)
		var err error
		blockHead, err = LoadBlockHeadByHash(&this.Id)
		if err != nil {
			return nil, err
		}
		TxCache.AddBlockHeadCache(bhashkey, blockHead)
	}

	return blockHead, nil
}

/*
加载本区块的所有交易
*/
func (this *Block) LoadTxs() (*BlockHead, *[]TxItr, error) {
	bh, err := this.Load()
	if err != nil {
		// fmt.Println("加载区块错误", err)
		return nil, nil, err
	}
	txs := make([]TxItr, 0)
	for i, one := range bh.Tx {

		// key := ""
		var txItr TxItr
		ok := false
		//是否启用缓存
		txhashkey := config.BuildBlockTx(bh.Tx[i])
		if config.EnableCache {
			// key = hex.EncodeToString(one)
			// key = utils.Bytes2string(one)
			//先判断缓存中是否存在
			txItr, ok = TxCache.FindTxInCache(txhashkey)
		}
		if !ok {
			// engine.Log.Error("未命中缓存 222 FindTxInCache")
			// bs, err := db.Find(one)
			// if err != nil {
			// 	return nil, nil, err
			// }

			// txItr, err = ParseTxBase(ParseTxClass(one), bs)
			var err error
			txItr, err = LoadTxBase(one)
			if err != nil {
				return nil, nil, err
			}
			//如果缓存已经启用，则把交易放入缓存
			if config.EnableCache {
				TxCache.AddTxInCache(txhashkey, txItr)
			}
		}

		txs = append(txs, txItr)
	}
	return bh, &txs, nil
}

/*
其它见证人的区块导入
*/
func (this *Chain) AddBlockOther(bhvo *BlockHeadVO) error {
	workModeLockStatic.GetLock()
	err := this.addBlock(bhvo)
	workModeLockStatic.BackLock()
	return err
}

/*
自己的区块导入
*/
func (this *Chain) AddBlockSelf(bhvo *BlockHeadVO) error {
	err := this.addBlock(bhvo)
	workModeLockStatic.BackLock()
	return err
}

/*
添加新的区块到分叉中
返回(true,chain)  区块添加到分叉链上
返回(false,chain) 区块添加到主链上的
返回(true,nil)    区块分叉超过了区块确认数量
@return    bool      是否有分叉
@return    *Chain    区块添加到的链
*/
// var AddBlockLock = new(sync.RWMutex)

// var stopBlockHeight = config.Mining_block_start_height + utils.GetRandNum(10) + 3
// var tempTotal = 0

var importHeight uint64 = 100
var importHeightCount uint64 = 0
var importHeightCountMax uint64 = 200

var notImportGroupStart uint64 = 120                    //150
var notImportGroupEnd uint64 = notImportGroupStart + 70 //70
var frstAdd = false

func (this *Chain) addBlock(bhvo *BlockHeadVO) error {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)

	start := config.TimeNow()
	TxCheckCond.SetImportTag()
	defer TxCheckCond.ResetImportTag()

	//测试，一段组导入不进去
	// if bhvo.BH.GroupHeight >= notImportGroupStart && bhvo.BH.GroupHeight < notImportGroupEnd {
	// 	if bhvo.BH.GroupHeight == notImportGroupStart && frstAdd == false {
	// 		frstAdd = true
	// 	} else {
	// 		return nil
	// 	}
	// }

	// if bhvo.BH.Height == importHeight && importHeightCount < importHeightCountMax {
	// 	importHeightCount++
	// 	return nil
	// }
	//测试，一段组导入不进去
	// if bhvo.BH.GroupHeight > notImportGroupStart && bhvo.BH.GroupHeight < notImportGroupEnd {
	// 	return nil
	// }

	// if bhvo.BH.Height > importHeight {
	// 	time.Sleep(time.Second + time.Second/2)
	// }

	//DBUG模式，只加载到此高度
	// if config.DBUG_import_height_max != 0 && bhvo.BH.Height > config.DBUG_import_height_max {
	// 	return nil
	// }

	engine.Log.Info("====== Import block group:%d block:%d prehash:%s hash:%s witness:%s", bhvo.BH.GroupHeight,
		bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash), hex.EncodeToString(bhvo.BH.Hash), bhvo.BH.Witness.B58String())
	// if bhvo.BH.Height <= this.GetCurrentBlock()+1 {
	// 	engine.Log.Info("Block height too low")
	// 	return nil
	// }
	bhvo.BH.BuildBlockHash()
	//先保存区块
	ok, err := SaveBlockHead(bhvo)
	if err != nil {
		engine.Log.Warn("save block error %s", err.Error())
		return err
	}
	if !ok {
		engine.Log.Error("收到的区块不连续，则从邻居节点同步")
		//收到的区块不连续，则从邻居节点同步
		this.NoticeLoadBlockForDB()
		return nil
	}

	//不能导入比统计高度低的区块
	if bhvo.BH.Height <= this.GetCurrentBlock() {
		engine.Log.Info("不能导入比统计高度低的区块")
		//检查区块是否已经导入过了
		if this.WitnessChain.CheckRepeatImportBlock(bhvo) {
			engine.Log.Info("重复导入区块")
			return ERROR_repeat_import_block
		}
		return ERROR_fork_import_block
	}

	//如果前置区块是统计过的区块，则前置区块hash必须对应，不能分叉。
	if this.CurrentBlockHash != nil && len(this.CurrentBlockHash) != 0 {
		if bhvo.BH.Height == this.GetCurrentBlock()+1 && !bytes.Equal(this.CurrentBlockHash, bhvo.BH.Previousblockhash) {
			engine.Log.Info("比已经统计的区块高度+1的区块，前置区块必须对应")
			return ERROR_fork_import_block
		}
	}

	//检查出块时间和本机时间相比，出块只能滞后，不能提前
	now := config.TimeNow().Unix()
	if bhvo.BH.Time > now+int64(config.Mining_block_time.Seconds()) {
		//if time.Unix(bhvo.BH.Time, 0).UnixNano() > now+config.Mining_block_time.Nanoseconds() {
		engine.Log.Warn("Build block It's too late %d %d", bhvo.BH.Time, now)
		//出块时间提前了
		return config.ERROR_advanced_by_build_block
	}
	//更新网络广播块高度
	if bhvo.BH.Height > forks.GetHighestBlock() {
		forks.SetHighestBlock(bhvo.BH.Height)
		this.Temp = bhvo
	}
	//排除上链但不合法的交易
	for _, one := range config.Exclude_Tx {
		if bhvo.BH.Height != one.Height {
			continue
		}
		for j, two := range bhvo.Txs {
			if !bytes.Equal(one.TxByte, *two.GetHash()) {
				// engine.Log.Info("交易hash不相同 %d %s %d %s", len(one.TxByte),
				// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
				continue
			}
			notExcludeTx := bhvo.Txs[:j]
			bhvo.Txs = append(notExcludeTx, bhvo.Txs[j+1:]...)
			break
		}
	}

	//验证撮合交易
	if err := oBook.CheckSwap(bhvo.Txs); err != nil {
		engine.Log.Warn("CheckSwap failed %s", err)
		return ERROR_swap_tx_check_fail
	}

	//检查区块是否已经导入过了
	if this.WitnessChain.CheckRepeatImportBlock(bhvo) {
		engine.Log.Info("重复导入区块")
		return ERROR_repeat_import_block
	}

	//在导入区块时,检查交易合法性
	if err := this.checkTxOfImportBlock(bhvo); err != nil {
		engine.Log.Error("在导入区块时,存在交易不合法: %s", err.Error())
		return ERROR_import_block_illegal_tx
	}

	//查找前置区块是否存在，并且合法
	preWitness := this.WitnessChain.FindPreWitnessForBlock(bhvo.BH.Previousblockhash)

	//if bhvo.BH.Height == importHeight && importHeightCount < importHeightCountMax {
	//	time.Sleep(1 * time.Second)
	//	importHeightCount++
	//	preWitness = nil
	//}

	if preWitness == nil {
		//有见证人节点分叉了，还在继续出块，但是分叉的长度没有自己的链长
		if this.GetCurrentBlock() > bhvo.BH.Height {

			engine.Log.Error("有见证人节点分叉了，还在继续出块，但是分叉的长度没有自己的链长")
			return nil
		}
		// engine.Log.Warn("找不到前置区块，新区块不连续 新高度：%d", bhvo.BH.Height)
		engine.Log.Warn("The front block cannot be found, and the new block is discontinuous with a new height:%d preblockhash:%s",
			bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash))
		//这里有可能分叉，有可能有断掉的区块未同步到位，现在都默认为区块未同步到位
		//TODO 有可能产生分叉，处理分叉的情况

		//从邻居节点同步区块
		this.NoticeLoadBlockForDB()
		return ERROR_fork_import_block
	} else {
		//判断区块高度是否连续
		// engine.Log.Info("pre witness group:%d height:%d bhvo height:%d",
		// 	preWitness.Group.Height, preWitness.Block.Height, bhvo.BH.Height)
		if preWitness.Block.Height+1 != bhvo.BH.Height {
			engine.Log.Error("new block height fail,pre height:%d new height:%d", preWitness.Block.Height, bhvo.BH.Height)
			return ERROR_import_block_height_not_continuity
		}
		// engine.Log.Info("查找前一个组的见证人")
		preGroupWitness := preWitness //查找前一个组的见证人
		for preGroupWitness.Group.Height == bhvo.BH.GroupHeight {
			//继续向前查找,直到找到不相等的组
			preGroupWitness = this.WitnessChain.FindPreWitnessForBlock(preGroupWitness.Block.PreBlockID)
			if preGroupWitness == nil {
				engine.Log.Info("ERROR_fork_import_block")
				return ERROR_fork_import_block
			}
		}
		// engine.Log.Info("找到的前置组:%+v", preGroupWitness.Group.NextGroup)
		//检查新添加的区块和前置区块之间是否有已经确认的组
		group := preGroupWitness.Group.NextGroup
		for group != nil && group.Height < bhvo.BH.GroupHeight {
			if group.BlockGroup != nil {
				engine.Log.Info("ERROR_fork_import_block")
				return ERROR_fork_import_block
			}
			group = group.NextGroup
		}
	}
	var currentWitness *Witness
	var isOverWitnessGroupChain bool
	//是首个区块，这里不构建前面的组
	if bhvo.BH.GroupHeight != config.Mining_group_start_height {
		//检查新区块是否在备用见证人组中
		currentWitness, isOverWitnessGroupChain = this.WitnessChain.FindWitnessForBlockOnly(bhvo)
		// engine.Log.Info("FindWitnessForBlockOnly:%t", isOverWitnessGroupChain)
		if isOverWitnessGroupChain {
			engine.Log.Info("The new height of the witness for this block cannot be found:%d", bhvo.BH.Height)
			//找到最后一个见证人
			var lastGroup = this.WitnessChain.WitnessGroup
			for ; lastGroup != nil; lastGroup = lastGroup.NextGroup {
				if lastGroup.NextGroup == nil {
					break
				}
			}
			//查找出未确认的块
			preBlock, _ := lastGroup.Witness[len(lastGroup.Witness)-1].FindUnconfirmedBlock()
			// engine.Log.Info("未确认的最后块:%d", preBlock.Height)
			//先统计之前的区块
			preBlock.Witness.Group.BuildGroup(&preBlock.Id)
			this.CountBlock(preBlock.Witness.Group)

			// _, lastBlock := this.GetLastBlock()
			// engine.Log.Info("最新高度:%d", lastBlock.Height)

			//先统计之前的区块
			// this.WitnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
			// engine.Log.Info("再构建后面的备用见证人组")
			//再构建后面的备用见证人组
			this.WitnessChain.CompensateWitnessGroupByGroupHeight(bhvo.BH.GroupHeight)
			// engine.Log.Info("再统计之前的区块 start")
			// this.WitnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
			// engine.Log.Info("再统计之前的区块 end")
		}
	}
	currentWitness, _ = this.WitnessChain.FindWitnessForBlockOnly(bhvo)
	if currentWitness == nil {
		engine.Log.Error("not font witness")
		//找不到这个见证人
		return config.ERROR_not_font_witness
	}

	//严格的导入区块顺序,如 组内W1,W2,W3三个见证人,但是出块顺序却是W2,W1,W3
	if err := this.strictAddBlock(bhvo); err != nil {
		engine.Log.Error("Not strict import block")
		return err
	}

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
	// 			engine.Log.Info("有相同的区块高度 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
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
	// 			engine.Log.Info("有相同的区块高度 %d %d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// 			return config.ERROR_block_height_equality
	// 		}
	// 	}
	// }

	// engine.Log.Info("导入区块签名总个数:%d 现有:%d 最少:%d", len(currentWitness.WitnessBigGroup.Witnesses),
	// 	len(bhvo.BH.ExtSign)+1, config.BftMajorityPrinciple(len(currentWitness.WitnessBigGroup.Witnesses)))
	//验证区块见证人签名
	if !bhvo.BH.CheckExtSign(currentWitness.WitnessBigGroup.Witnesses) {
		// 	//区块验证不通过，区块不合法
		engine.Log.Warn("AddBlock verification failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		return config.ERROR_sign_fail
	}
	//验证区块见证人签名
	// if !bhvo.BH.CheckBlockHead(currentWitness.Puk) {
	// 	//区块验证不通过，区块不合法
	// 	engine.Log.Warn("Block verification failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// 	// this.chain.NoticeLoadBlockForDB(false)
	// 	return
	// }

	//engine.Log.Info("add block spend time 666 %s", config.TimeNow().Sub(start))
	// engine.Log.Info("add block group:%d,height:%d 444444444444444444444444444", bhvo.BH.GroupHeight, bhvo.BH.Height)
	//把见证人设置为已出块 ===合约检查奖励验证===
	ok = this.WitnessChain.SetWitnessBlock(bhvo)
	if !ok {
		//从邻居节点同步
		// engine.Log.Info("------------------------------------------------------------ 22222222222222")
		engine.Log.Debug("Setting witness block failed")
		// this.NoticeLoadBlockForDB(false)
		return errors.New("Setting witness block failed")
	}
	//this.GroupHeight = bhvo.BH.GroupHeight
	// engine.Log.Info("add block group:%d,height:%d 5555555555555555555555555555555", bhvo.BH.GroupHeight, bhvo.BH.Height)

	//导入区块后，清理分叉
	this.cleanFork()

	//是首个区块，这里不构建前面的组
	if bhvo.BH.GroupHeight == config.Mining_group_start_height {
		this.WitnessChain.WitnessGroup.BuildGroup(nil)
		this.WitnessChain.BuildWitnessGroup(false, false)
		this.WitnessChain.WitnessGroup = this.WitnessChain.WitnessGroup.NextGroup
		this.WitnessChain.BuildWitnessGroup(false, true)
	} else {
		//本组有2个以上的见证人出块，那么本组才有可能被确认
		// engine.Log.Info("111111111 %d", currentWitness.Group.Height)
		//判断这个组是否多人出块
		// ok, group := currentWitness.Group.CheckBlockGroup(nil)
		ok, group := currentWitness.Group.CheckBlockGroup(&currentWitness.Block.Id)
		//第三版共识添加的内容。两个组ABC ABC，当第二组第一个块A导入后，立即统计前组最后一个块C。
		if !ok && group != nil {
			// engine.Log.Info("BlockID 111:%d", len(group.Blocks[0].PreBlockID))
			//找到上一个未确认的组最后一个见证人
			witness := this.WitnessChain.FindPreWitnessForBlock(group.Blocks[0].PreBlockID)
			// engine.Log.Info("witness is nil:%v", witness)
			//第三版共识，新组一旦有区块导入，立即确认和统计上一组所有区块
			//wg := witness.Group.BuildGroup(&witness.Block.Id)
			//if wg != nil && witness.Group.IsBuildGroup {
			//	// engine.Log.Info("44444444444")
			//	for _, one := range wg.Witness {
			//		if !one.CheckIsMining {
			//			if one.Block == nil {
			//				this.WitnessBackup.AddBlackList(*one.Addr)
			//			} else {
			//				this.WitnessBackup.SubBlackList(*one.Addr)
			//			}
			//			one.CheckIsMining = true
			//		}
			//	}
			//}
			//统计上一组 合约统计
			// str := ""
			// for _, one := range witness.Group.BlockGroup.Blocks {
			// 	str = str + " " + strconv.Itoa(int(one.Height))
			// }
			// engine.Log.Info("开始统计:%s", str)
			this.CountBlock(witness.Group)
		}
		//第三版共识，本组一旦有多人出块，当有新区块连接它，则统计之前的区块
		//例如：本组有3个块，ABC，当导入C块时，立即统计AB块
		if ok {
			//统计本组  合约统计
			this.CountBlockCustom(currentWitness.Group, group)
		}

		if ok {
			// engine.Log.Info("222222222 %s", hex.EncodeToString(group.Blocks[0].PreBlockID))
			// engine.Log.Info("BlockID 111:%d", len(group.Blocks[0].PreBlockID))
			//找到上一个未确认的组最后一个见证人
			witness := this.WitnessChain.FindPreWitnessForBlock(group.Blocks[0].PreBlockID)
			// engine.Log.Info("witness is nil:%v", witness)
			// engine.Log.Info("3333333333 %s", hex.EncodeToString(witness.Block.Id))
			witness.Group.BuildGroup(&witness.Block.Id)
			// wg := preWitness.Group.BuildGroup(&bhvo.BH.Previousblockhash)
			//找到上一组到本组的见证人组，开始查找没有出块的见证人，未出块的见证人要加入黑名单
			// engine.Log.Info("5555555555")
			//合约统计
			this.CountBlock(witness.Group)
			// this.witnessChain.BuildBlockGroup(bhvo, preWitness)

			this.WitnessChain.WitnessGroup = currentWitness.Group

			this.WitnessChain.BuildWitnessGroup(false, true)

		}

		//白名单计数，区块统计后才统计白名单计数。不合法的组不统计，只统计合法组中未出块的见证人
		this.CountBlackListNumber()

		// 触发保存快照
		if config.DisableSnapshot {
			db.LevelDB.StoreSnap(nil)
			db.LevelDB.ResetSnap()
		} else {
			if (ok || (!ok && group != nil)) && this.SnapshotExtraCondition() {
				if err := snapshot.Save(this.CurrentBlock); err != nil {
					engine.Log.Warn(">>>>>> Snapshot[保存] 未保存快照 %v,高度 %d", err, this.CurrentBlock)
				}
			}
		}
	}
	// engine.Log.Info("add block group:%d,height:%d 6666666666666666666666666", bhvo.BH.GroupHeight, bhvo.BH.Height)
	//TODO 检查从此时开始，未来还未出块的见证人组有多少，太少则创建新的组，避免跳过太多组后出块暂停

	// engine.Log.Info("添加区块 666 耗时 %s", config.TimeNow().Sub(start))
	// engine.Log.Info("保存区块 99999999999999 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// SaveTxToBlockHead(bhvo)
	//engine.Log.Info("Save block Time spent %s", config.TimeNow().Sub(start))

	//是首个区块，这里不构建前面的组
	// if bhvo.BH.GroupHeight != config.Mining_group_start_height {
	// 	this.witnessChain.BuildBlockGroup(bhvo, preWitness)
	// }
	// engine.Log.Info("添加区块 777 耗时 %s", config.TimeNow().Sub(start))

	this.WitnessChain.BuildMiningTime()

	//回收内存，将前n个见证人之前的见证人链删除。
	this.WitnessChain.GCWitness()

	//if bhvo.BH.Height == config.Witness_backup_group_overheight {
	//	config.Witness_backup_group = config.Witness_backup_group_new
	//}

	// engine.Log.Info("添加区块 888 耗时 %s", config.TimeNow().Sub(start))

	//删除之前区块的交易缓存
	// TxCache.RemoveHeightTxInCache(bhvo.BH.Height - (config.Mining_group_max * 2))

	//判断自己是否是见证人，自己是见证人，则添加其他见证人白名单连接

	// if this.witnessChain.FindWitness(keystore.GetCoinbase().Addr) {

	// }

	//pl
	endT := config.TimeNow().Sub(start)
	// engine.Log.Info("Import block Time spent %s", endT)
	if endT.Seconds() > 1 {
		engine.Log.Info("Import block Time spent too long %s", endT)
	}
	if SaveBlockTime == nil {
		SaveBlockTime = make(map[uint64]string)
	} else {
		SaveBlockTime[bhvo.BH.Height] = endT.String()
	}

	return nil
}

/*
严格的导入区块顺序,如 组内W1,W2,W3三个见证人,但是导入区块顺序却是W2,W1,W3
如以下场景:
wg:157715 iComATBPTBm9H3CtXTo9VR9NYfiC84ZqfNNyP5 2024-03-09 19:34:37 565000000000
wg:157715 iCom7fhMLft2gfVaAKPi5GXhsRoQVpXnRRwTe5 2024-03-09 19:34:38 564300000000
wg:157715 iCom6xWdjPjXcMAXBNS6k1nDcjV5LJZoF7heQ5 2024-03-09 19:34:39 519100000000

2024/03/09 19:34:37.903 [I] [chain.go:346]  ====== Import block group:157715 block:442319 prehash:f63b1e86f1758f7178e998ae810f93339b9c5c3469085b9b7603a6999177b473 hash:d5a1515d560d961efee122c8faaae4852068a24751bc6145fcf0847001003752 witness:iCom7fhMLft2gfVaAKPi5GXhsRoQVpXnRRwTe5
2024/03/09 19:34:38.383 [I] [chain.go:346]  ====== Import block group:157715 block:442320 prehash:d5a1515d560d961efee122c8faaae4852068a24751bc6145fcf0847001003752 hash:7d51c5eaa1cbe441df89f5fc3a25f36f9e3c2a02fb0d41da13900cfd662b52bf witness:iComATBPTBm9H3CtXTo9VR9NYfiC84ZqfNNyP5
2024/03/09 19:34:38.945 [I] [chain.go:346]  ====== Import block group:157715 block:442321 prehash:7d51c5eaa1cbe441df89f5fc3a25f36f9e3c2a02fb0d41da13900cfd662b52bf hash:1eb1634af6be8b17ea3aacfc27e8552c85638cd7b2d032db6dde66f8c9551fd3 witness:iCom6xWdjPjXcMAXBNS6k1nDcjV5LJZoF7heQ5

函数功能:
1.相邻区块见证人组高度大于前一个见证人组高度,即W2组高度大于W1组高度
2.相邻区块见证人组相等,见证人区块不能比前一个见证人先导入,需严格按照见证人顺序导入,即W1组高度=W2组高度,严格按照W1,W2的区块顺序导入
*/
func (this *Chain) strictAddBlock(bhvo *BlockHeadVO) error {
	if bhvo.BH.Height == config.Mining_block_start_height {
		return nil
	}

	currentWitness, _ := this.WitnessChain.FindWitnessForBlockOnly(bhvo)
	if currentWitness == nil {
		return config.ERROR_not_font_witness
	}

	//preWitness, _ := this.WitnessChain.FindWitnessForBlockOnly(bhvo.BH.Previousblockhash)
	//preWitness := currentWitness.PreWitness
	//if preWitness == nil {
	//	return config.ERROR_not_font_witness
	//}

	//if currentWitness.Group.Height > preWitness.Group.Height {
	//	return nil
	//}

	//if currentWitness.Group.Height < preWitness.Group.Height {
	//	return ERROR_not_strict_import_block
	//}

	currentBlockId := bhvo.BH.Previousblockhash
	gotCurrentWitness := false
	for _, wit := range currentWitness.Group.Witness {
		if wit == currentWitness {
			gotCurrentWitness = true
		}
		if wit.Block == nil {
			continue
		}

		if bytes.Equal(currentBlockId, wit.Block.Id) && gotCurrentWitness {
			return ERROR_not_strict_import_block
		}
	}

	return nil
}

/*
是否满足外部额外的快照条件
*/
func (this *Chain) SnapshotExtraCondition() bool {
	//1.同步过程中,增大快照间隔
	if !this.SyncBlockFinish {
		if this.CurrentBlock-snapshot.Height() < config.SnapshotMinInterval*10 {
			return false
		}
	}

	//2.判断上一个成立的组是否满组,即组的见证人数量=出块数量
	group := this.WitnessChain.WitnessGroup
	if group == nil {
		return false
	}

	if group.PreGroup == nil {
		return false
	}

	if group.PreGroup.BlockGroup != nil {
		//上一个成立的组见证人/块数量
		groupWitnessNum := len(group.PreGroup.Witness)
		groupBlocksNum := len(group.PreGroup.BlockGroup.Blocks)

		//上一个成立的组是满组出块
		if groupWitnessNum == groupBlocksNum {
			return true
		}
	}

	return false
}

/*
清理分叉
当出现多个合法组时，则保留组高度最高的合法组
*/
func (this *Chain) cleanFork() {
	if this.CurrentBlockHash == nil || len(this.CurrentBlockHash) == 0 {
		return
	}
	//导入区块后，清理分叉
	//例如：45007:133975 45007:133976 45008:133975 45008:133976 45009:133977 45016:133977
	//当出现多个合法组时，则保留组高度最高的合法组
	countBlockWitness := this.WitnessChain.FindPreWitnessForBlock(this.CurrentBlockHash)

	groups := make([]*WitnessSmallGroup, 0)
	//engine.Log.Info("countBlockWitness:%+v %v", countBlockWitness, this.CurrentBlockHash)
	for group := countBlockWitness.Group.NextGroup; group != nil; group = group.NextGroup {
		ok, g := group.CheckBlockGroup(nil)
		if !ok {
			continue
		}
		//检查是否能链到之前的组
		have := false
		for _, one := range groups {
			//engine.Log.Info("合法组%d前置区块:%s", g.Height, hex.EncodeToString(g.Blocks[0].Id))
			//engine.Log.Info("查询前置组:%d", one.Height)
			if ok, _ := one.CheckBlockGroup(&g.Blocks[0].PreBlockID); ok {
				//engine.Log.Info("找到前置区块组:%d", one.Height)
				one = group
				have = true
				break
			}
		}
		if have {
			continue
		}
		//engine.Log.Info("添加合法组:%d", group.Height)
		groups = append(groups, group)

	}
	if len(groups) > 1 {
		for i := 0; i < len(groups)-1; i++ {
			for _, one := range groups[i].Witness {
				//engine.Log.Info("清理Block:%d", one.Block.Height)
				one.Block = nil
			}
		}
	}

}

/*
统计白名单计数
区块统计后才统计白名单计数。不合法的组不统计，只统计合法组中未出块的见证人
*/
func (this *Chain) CountBlackListNumber() {
	if this.CurrentBlockHash == nil || len(this.CurrentBlockHash) == 0 {
		return
	}
	countBlockWitness := this.WitnessChain.FindPreWitnessForBlock(this.CurrentBlockHash)
	if countBlockWitness == nil {
		return
	}
	for group := countBlockWitness.Group; group != nil && !group.Witness[0].CheckIsMining; group = group.PreGroup {
		if group.BlockGroup == nil || !group.IsCount {
			continue
		}
		for _, one := range group.Witness {
			one.CheckIsMining = true
			if one.Block == nil {
				this.WitnessBackup.AddBlackList(*one.Addr)
			} else {
				have := false
				for _, oneBlock := range group.BlockGroup.Blocks {
					//engine.Log.Info("对比区块，判断是否出块:%d %d %d %d %t", one.Block.Group.Height, one.Block.Height,
					//	oneBlock.Group.Height, oneBlock.Height, bytes.Equal(oneBlock.Id, one.Block.Id))
					if bytes.Equal(oneBlock.Id, one.Block.Id) {
						have = true
						break
					}
				}
				if have {
					this.WitnessBackup.SubBlackList(*one.Addr)
				} else {
					this.WitnessBackup.AddBlackList(*one.Addr)
				}
			}
		}
	}
}

// 在导入区块时,检查交易合法性
func (this *Chain) checkTxOfImportBlock(bhvo *BlockHeadVO) error {
	var vmRun *evm.VmRun
	for _, txItr := range bhvo.Txs {
		if txItr.Class() == config.Wallet_tx_type_pay {
			if _, ok := this.TransactionManager.checkedTxCache.Load(utils.Bytes2string(*txItr.GetHash())); ok {
				return nil
			}

			//验证大余额
			for _, one := range *txItr.GetVin() {
				addr := one.GetPukToAddr()
				if GetAddrValueBig(addr) {
					return config.ERROR_addr_value_big
				}
			}

			//验证余额
			fromAddr := (*txItr.GetVin())[0].GetPukToAddr()
			spend, _ := this.TransactionManager.unpackedTransaction.FindAddrSpend(fromAddr)
			notSpend, _, _ := GetNotspendByAddrOther(this, *fromAddr)
			if notSpend < spend+txItr.GetSpend() {
				engine.Log.Info("余额不足:%s %s %d<%d+%d", hex.EncodeToString(*txItr.GetHash()), fromAddr.B58String(), notSpend, spend, txItr.GetSpend())
				return config.ERROR_not_enough
			}
			return nil
		}

		//合约预执行
		if txItr.Class() == config.Wallet_tx_type_contract {
			if vmRun == nil {
				//预执行交易evm
				vmRun = evm.NewCountVmRun(nil)
				//初始化storage
				vmRun.SetStorage(nil)
			}
			from := (*txItr.GetVin())[0].GetPukToAddr()
			to := (*txItr.GetVout())[0].Address
			vmRun.SetTxContext(*from, to, *txItr.GetHash(), bhvo.BH.Time, bhvo.BH.Height, nil, nil)
			return txItr.(*Tx_Contract).PreExecV1(vmRun)
		}
	}
	return nil
}

var SaveBlockTime map[uint64]string

/*
强制统计之前的组
*/
// func (this *Block) BuildGroupBeforeForce(smallGroup *WitnessSmallGroup) {
// 	smallGroup.

// }

// /*
// 	添加本区块的下一个区块
// */
// func (this *Block) AddNextBlock(bhash []byte) error {
// this.NextBlock
// 	return nil
// }

// /*
// 	修改本区块的下一个区块中最长区块下标为0
// */
// func (this *Block) UpdateNextIndex(bhash [][]byte) error {
// 	this.NextBlock

// 	return nil
// }

/*
修改next区块顺序
*/
func (this *Block) FlashNextblockhash() error {
	// engine.Log.Info("000 update block nextblockhash,this %d blockid:%s nextid:%s", this.Height, hex.EncodeToString(this.Id), hex.EncodeToString(this.NextBlock.Id))
	bh, err := this.Load()
	if err != nil {
		return err
	}

	bh.Nextblockhash = this.NextBlock.Id

	// bs, err := bh.Json()
	bs, err := bh.Proto()
	if err != nil {
		return err
	}

	bhashkey := config.BuildBlockHead(this.Id)
	TxCache.FlashBlockHeadCache(bhashkey, bh)

	if bh.Nextblockhash == nil {
		engine.Log.Error("save block nextblockhash nil %s", string(*bs))
	}

	err = db.LevelDB.Save(bhashkey, bs)
	if err != nil {
		return err
	}

	db.LevelDB.Commit(bhashkey)

	// bs, _ = db.Find(this.Id)
	// engine.Log.Info("打印刚保存的区块 \n %s", string(*bs))

	return nil
	// return &bs, nil
}

var countBlockLock sync.RWMutex

/*
统计当前正在出块的组中的区块
*/
func (this *Chain) CountBlockCustom(witnessGroup *WitnessSmallGroup, group *Group) {
	// engine.Log.Info("开始统计当前出块组 %+v", group.Height)
	countBlockLock.Lock()
	defer countBlockLock.Unlock()

	// start := config.TimeNow()

	have := len(group.Blocks)
	need := config.ConfirmMinNumber(len(witnessGroup.Witness))
	if need >= have {
		return
	}
	for i := 0; i < have-1; i++ {
		one := group.Blocks[i]
		if one.isCount {
			continue
		}
		// engine.Log.Info("开始统计 22222222222 %d %d", one.Group.Height, one.Height)
		// engine.Log.Info("统计交易 111 耗时 %s", config.TimeNow().Sub(start))
		//创始块不需要统计
		if one.Height == config.Mining_block_start_height {
			continue
		}
		// engine.Log.Info("统计已经确认的组中的区块")
		bh, txs, err := one.LoadTxs()
		if err != nil {
			//TODO 是个严重的错误
			continue
		}
		bhvo := &BlockHeadVO{BH: bh, Txs: *txs}

		//排除上链但不合法的交易
		for _, one := range config.Exclude_Tx {
			if bhvo.BH.Height != one.Height {
				continue
			}
			for j, two := range bhvo.Txs {
				if !bytes.Equal(one.TxByte, *two.GetHash()) {
					// engine.Log.Info("交易hash不相同 %d %s %d %s", len(one.TxByte),
					// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
					continue
				}

				// engine.Log.Info("排除交易前 %d 排除第 %d 个交易", len(bhvo.Txs), j)

				notExcludeTx := bhvo.Txs[:j]
				bhvo.Txs = append(notExcludeTx, bhvo.Txs[j+1:]...)

				// engine.Log.Info("排除交易后 %d", len(bhvo.Txs))
				break
			}
		}
		// engine.Log.Info("开始统计 33333 %d %d", one.Group.Height, one.Height)
		//把上一个块的 Nextblockhash 修改为最长链的区块
		one.PreBlock.FlashNextblockhash()

		//把上一个组的区块 Nextblockhash 刷新一次，应对上一组分叉导致nextHash保存错误的问题
		// if one.PreBlock.Group.Height != one.Group.Height {
		// 	for _, oneBlock := range one.PreBlock.Group.Blocks {
		// 		oneBlock.PreBlock.FlashNextblockhash()
		// 	}
		// }

		// engine.Log.Info("统计交易 222 耗时 %s", config.TimeNow().Sub(start))

		//计算余额
		this.Balance.CountBalanceForBlock(bhvo)

		// engine.Log.Info("统计交易 333 耗时 %s", config.TimeNow().Sub(start))
		//统计交易中的备用见证人以及见证人投票
		this.WitnessBackup.CountWitness(&bhvo.Txs)
		// engine.Log.Info("统计交易 444 耗时 %s", config.TimeNow().Sub(start))
		//删除已经打包了的交易
		this.TransactionManager.DelTx(bhvo.Txs)
		// engine.Log.Info("统计交易 555 耗时 %s", config.TimeNow().Sub(start))
		//设置最新高度
		this.SetCurrentBlock(one.Height, one.Id)
		//清除掉内存中已经过期的交易
		this.TransactionManager.CleanIxOvertime(one.Height)
		one.isCount = true
		// engine.Log.Info("统计交易 666 耗时 %s", config.TimeNow().Sub(start))
	}
}

/*
统计已经确认的组中的区块
*/
func (this *Chain) CountBlock(witnessGroup *WitnessSmallGroup) {
	// engine.Log.Info("开始统计上个确认的组 %+v", witnessGroup)
	countBlockLock.Lock()
	defer countBlockLock.Unlock()

	// start := config.TimeNow()

	//如果本组没有评选出最多出块组，则不统计本组
	if witnessGroup.BlockGroup == nil {
		return
	}

	if witnessGroup.IsCount {
		return
	}

	//获取本组中的出块
	for _, one := range witnessGroup.BlockGroup.Blocks {
		if one.isCount {
			continue
		}
		// engine.Log.Info("开始统计 22222222222 %d %d", one.Group.Height, one.Height)
		// engine.Log.Info("统计交易 111 耗时 %s", config.TimeNow().Sub(start))
		//创始块不需要统计
		if one.Height == config.Mining_block_start_height {
			continue
		}
		// engine.Log.Info("统计已经确认的组中的区块")
		bh, txs, err := one.LoadTxs()
		if err != nil {
			//TODO 是个严重的错误
			continue
		}
		bhvo := &BlockHeadVO{BH: bh, Txs: *txs}

		//排除上链但不合法的交易
		for _, one := range config.Exclude_Tx {
			if bhvo.BH.Height != one.Height {
				continue
			}
			for j, two := range bhvo.Txs {
				if !bytes.Equal(one.TxByte, *two.GetHash()) {
					// engine.Log.Info("交易hash不相同 %d %s %d %s", len(one.TxByte),
					// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
					continue
				}

				// engine.Log.Info("排除交易前 %d 排除第 %d 个交易", len(bhvo.Txs), j)

				notExcludeTx := bhvo.Txs[:j]
				bhvo.Txs = append(notExcludeTx, bhvo.Txs[j+1:]...)

				// engine.Log.Info("排除交易后 %d", len(bhvo.Txs))
				break
			}
		}

		//把上一个块的 Nextblockhash 修改为最长链的区块
		one.PreBlock.FlashNextblockhash()

		//把上一个组的区块 Nextblockhash 刷新一次，应对上一组分叉导致nextHash保存错误的问题
		// if one.PreBlock.Group.Height != one.Group.Height {
		// 	for _, oneBlock := range one.PreBlock.Group.Blocks {
		// 		oneBlock.PreBlock.FlashNextblockhash()
		// 	}
		// }

		// engine.Log.Info("统计交易 222 耗时 %s", config.TimeNow().Sub(start))

		//计算余额
		this.Balance.CountBalanceForBlock(bhvo)

		// engine.Log.Info("统计交易 333 耗时 %s", config.TimeNow().Sub(start))
		//统计交易中的备用见证人以及见证人投票
		this.WitnessBackup.CountWitness(&bhvo.Txs)
		// engine.Log.Info("统计交易 444 耗时 %s", config.TimeNow().Sub(start))
		//删除已经打包了的交易
		this.TransactionManager.DelTx(bhvo.Txs)
		// engine.Log.Info("统计交易 555 耗时 %s", config.TimeNow().Sub(start))
		//设置最新高度
		this.SetCurrentBlock(one.Height, one.Id)
		//清除掉内存中已经过期的交易
		this.TransactionManager.CleanIxOvertime(one.Height)
		one.isCount = true
		// engine.Log.Info("统计交易 666 耗时 %s", config.TimeNow().Sub(start))
	}
	witnessGroup.IsCount = true

}

/*
获取本链已经确认的最高区块
*/
func (this *Chain) GetLastBlock() (witness *Witness, block *Block) {
	witnessGroup := this.WitnessChain.WitnessGroup
	if witnessGroup == nil {
		return
	}
	//先找到最后一个见证人组
	for ; witnessGroup.NextGroup != nil; witnessGroup = witnessGroup.NextGroup {
	}
	//找到最后一个已确认的组
	for ; witnessGroup.BlockGroup == nil; witnessGroup = witnessGroup.PreGroup {
		if witnessGroup.PreGroup == nil {
			break
		}
	}
	block = witnessGroup.BlockGroup.Blocks[len(witnessGroup.BlockGroup.Blocks)-1]
	witness = block.Witness
	return

	// witnessGroup := this.WitnessChain.witnessGroup
	// if witnessGroup == nil {
	// 	return
	// }
	// if witnessGroup.Height != config.Mining_group_start_height {
	// 	for {
	// 		witnessGroup = witnessGroup.PreGroup
	// 		if witnessGroup == nil {
	// 			break
	// 		}
	// 		//找到合法的区块组
	// 		if witnessGroup.BlockGroup != nil {
	// 			break
	// 		}
	// 	}
	// }
	// block = witnessGroup.BlockGroup.Blocks[len(witnessGroup.BlockGroup.Blocks)-1]
	// witness = block.witness
	// return
}

/*
获取本链的收益管理器
*/
func (this *Chain) GetBalance() *BalanceManager {
	return this.Balance
}

/*
获取本链的交易管理器
*/
func (this *Chain) GetTransactionManager() *TransactionManager {
	return this.TransactionManager
}

/*
打印块列表
*/
func (this *Chain) PrintBlockList(bhvo *BlockHeadVO) {
	//lastWit, lastBlock := this.GetLastBlock()
	//engine.Log.Info(">>>>>> Snapshot 正在导入的块 %d,%v", bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Hash))
	//engine.Log.Info(">>>>>> Snapshot 最新的见证人出块 %d,%v", lastWit.Block.Height, hex.EncodeToString(lastWit.Block.Id))
	//engine.Log.Info(">>>>>> Snapshot 最新2块 %d,%v", lastBlock.PreBlock.Height, hex.EncodeToString(lastBlock.PreBlock.Id))
	//engine.Log.Info(">>>>>> Snapshot 最新1块 %d,%v", lastBlock.Height, hex.EncodeToString(lastBlock.Id))
}

/*
依次获取前n个区块的hash，连接起来做一次hash
*/
func (this *Chain) HashRandom() *[]byte {
	_, lastBlock := this.GetLastBlock()

	// if lastBlock == nil || lastBlock.Height > config.RandomHashHeightMin {
	// 	// bs := make([]byte, 0)
	// 	// bs = utils.Hash_SHA3_256(bs)
	// 	// return &bs
	// 	return &config.RandomHashFixed
	// }

	if lastBlock != nil {
		engine.Log.Info("lastblock:%d", lastBlock.Height)
		//curr := this.WitnessChain.witnessGroup
		//engine.Log.Error("当前出块组：%d，lastblock:%d %d", curr.Height, lastBlock.Group.Height, lastBlock.Height)
		// for preWitness := lastBlock.witness.PreWitness; preWitness != nil; preWitness = preWitness.PreWitness {
		// 	if preWitness.WitnessBackupGroup == lastBlock.witness.WitnessBackupGroup {
		// 		continue
		// 	}
		// 	if preWitness.Group.BlockGroup == nil {
		// 		continue
		// 	}
		// 	lastBlock = preWitness.Block
		// 	break
		// }
		// engine.Log.Info("lastBlock :%s", hex.EncodeToString(lastBlock.Id))
		if random, ok := config.RandomMap.Load(utils.Bytes2string(lastBlock.Id)); ok {
			bs := random.(*[]byte)
			// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(*bs))
			return bs
		}
	}

	startHeight, endHeight := uint64(0), uint64(0)
	var preHash *[]byte
	bs := make([]byte, 0)
	// witness := this.witnessChain.witness
	//链端初始化时候，this.witnessChain.witness为nil
	for i := 0; lastBlock != nil && i < config.Mining_block_hash_count; i++ {
		if lastBlock.Height < config.NextHashHeightMax {
			if preHash == nil {
				if random, ok := config.NextHash.Load(utils.Bytes2string(lastBlock.Id)); ok {
					bsOne := random.(*[]byte)
					bs = append(bs, *bsOne...)
					preHash = bsOne
					// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(*bs))
					// return bs
					continue
				}
			} else {
				if random, ok := config.NextHash.Load(utils.Bytes2string(*preHash)); ok {
					bsOne := random.(*[]byte)
					bs = append(bs, *bsOne...)
					preHash = bsOne
					// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(*bs))
					// return bs
					continue
				}
			}
		}
		if i == 0 {
			startHeight = lastBlock.Height
		}
		bs = append(bs, lastBlock.Id...)
		if lastBlock.PreBlock == nil {
			break
		}
		endHeight = lastBlock.Height
		lastBlock = lastBlock.PreBlock
	}
	// engine.Log.Info("HashRandom :%s", hex.EncodeToString(bs))
	bs = utils.Hash_SHA3_256(bs)
	// if lastBlock != nil {
	// 	// golog.Infof("lastBlock height:%d HashRandom hash:%s", lastBlock.Height, hex.EncodeToString(bs))
	// 	engine.Log.Info("lastBlock height:%d-%d Random hash:%s", lastBlock.Height,
	// 		lastBlock.Height+config.Mining_block_hash_count, hex.EncodeToString(bs))
	// }
	// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(bs))

	engine.Log.Info("DebugH HashRandom: %d-%d %s", startHeight, endHeight, hex.EncodeToString(bs))
	return &bs
}

// func SetCurrentBlock(n uint64) {
// 	atomic.StoreUint64(&forks.CurrentBlock, n)
// }

// func GetCurrentBlock(n uint64) {
// 	atomic.StoreUint64(&forks.CurrentBlock, n)
// }

/*
查询历史交易记录
*/
func (this *Chain) GetHistoryBalance(start *big.Int, total int) []HistoryItem {
	// return this.history.Get(start, total)
	return balanceHistoryManager.Get(start, total)
}

/*
倒叙查询历史交易记录
*/
func (this *Chain) GetHistoryBalanceDesc(start uint64, total int) []HistoryItem {
	// return this.history.Get(start, total)
	return balanceHistoryManager.GetDesc(start, total)
}

/*
获取历史交易数量
*/
func (this *Chain) GetHistoryBalanceCount() uint64 {
	return balanceHistoryManager.GenerateMaxId.Uint64()
}

/*
获取账号历史交易数量
*/
func (this *Chain) GetAccountHistoryBalanceCount(addr string) uint64 {
	index, ok := balanceHistoryManager.GenerateMaxIdForAccount.Load(strings.ToLower(addr))
	if ok {
		return index.(uint64)
	}
	return 0
}

/*
调用这个函数说明分叉链哈希顺序已经修复完成了

a.移除快照标记:
1.该标记位会在下一次重启生效,等同主链重启临时禁用快照
2.该标记位会在下一次快照复位(若不重启，主链恢复正常，该标记也会复位)

b.强制要求重启
1.快照高度大于分叉高度，此时快照是分叉的快照了，只能清空快照和统计数据，从头开始重新拉区块.
2.某些情况存在快照高度小于分叉高度，此时快照是个有效快照,重启拉起会从快照高度拉起.
*/
func (this *Chain) FixChainForkForSnapshot(forkGroupHeight uint64) {
	// 提交部分增量数据
	//Note:目前的逻辑已经保证的快照的有效性(满组才会快照)
	db.LevelDB.CommitPrekSnap(config.DBKEY_snapshot_height, config.DBKEY_BlockHead, config.DBKEY_Block_tx)
	////快照是否无效
	//invalidSnapshot := true
	////查询快照高度区块信息
	//if bh := LoadBlockHeadByHeight(snapshot.Height()); bh != nil {
	//	// 快照组高度小于分叉组高度是有效的
	//	if bh.GroupHeight < forkGroupHeight {
	//		invalidSnapshot = false
	//	}
	//}
	//if invalidSnapshot {
	//	db.LevelTempDB.Del(config.DBKEY_snapshot_height)
	//}
	//// 提交部分增量数据
	//db.LevelDB.CommitPrekSnap(config.DBKEY_snapshot_height, config.DBKEY_BlockHead, config.DBKEY_Block_tx)

	////暂停主链修复区块行为
	//atomic.StoreUint32(&this.StopSyncBlock, 1)
	//engine.Log.Warn(">>>>>> 主链需要重新启动 <<<<<<")
}

/*
获取本链的Swap交易管理器
*/
func (this *Chain) GetTransactionSwapManager() *TransactionSwapManager {
	return this.transactionSwapManager
}

var onChainInitialized = make(chan bool, 1)

// 链初始化完成信号
func ChainInitialized() {
	select {
	case got := <-onChainInitialized:
		if len(onChainInitialized) == 0 {
			onChainInitialized <- got
		}
	}
}
