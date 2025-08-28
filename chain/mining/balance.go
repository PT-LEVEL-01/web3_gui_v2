package mining

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/event"
	"web3_gui/chain/evm"
	"web3_gui/chain/evm/abi"
	"web3_gui/chain/evm/common"
	"web3_gui/chain/evm/common/evmutils"
	"web3_gui/chain/evm/precompiled"
	"web3_gui/chain/evm/vm"
	"web3_gui/chain/evm/vm/environment"
	"web3_gui/chain/evm/vm/opcodes"
	"web3_gui/chain/mining/name"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/chain/sqlite3_db"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
地址余额管理器
*/
type BalanceManager struct {
	countTotal    uint64            //
	chain         *Chain            //链引用
	syncBlockHead chan *BlockHeadVO //正在同步的余额，准备导入到余额中
	// notspentBalance   *TxItemManager      //未花费的余额索引
	// notspentBalanceDB *TxItemManagerDB    //未花费的余额索引，保存到数据库，降低内存消耗
	nodeWitness      *TxItem             //保存本节点见证人
	depositWitness   *sync.Map           //见证人节点押金 key:string=见证人节点地址;value:DepositInfo;
	depositCommunity *sync.Map           //社区节点押金 key:string=社区节点地址;value:DepositInfo;
	depositLight     *sync.Map           //轻节点押金 key:string=轻节点地址;value:DepositInfo;
	witnessVote      *sync.Map           //保存见证人节点获得总的投票数 key：string=见证人地址；value：投票数量 uint64
	depositVote      *sync.Map           //保存轻节点投票，key:string=轻节点地址;value:DepositInfo;
	communityVote    *sync.Map           //保存社区节点获得总的投票数 key:string=社区节点地址;value:uint64=投票数量;
	witnessBackup    *WitnessBackup      //
	otherDeposit     *sync.Map           //其他押金，key:uint64=交易类型;value:*sync.Map=押金列表;
	txManager        *TransactionManager //
	addBlockNum      *sync.Map           // 节点出块数  key:string 节点地址; value:uint64
	addBlockReward   *sync.Map           // 节点出块奖励  key:string 节点地址; value:uint64
	blockIndex       *sync.Map           // 出块记录索引
	// cacheTxLockLock *sync.RWMutex
	//cacheTxlock     []*TxItr
	cacheTxlock          *sync.Map //key:string=转出地址;value:*TxList=;
	eventBus             event.EventBus
	witnessRatio         *sync.Map //见证人比例,包含社区的比例,因为合约存的比例也是放在一起的 key：string=见证人地址；value：uint16
	lastDistributeHeight uint64    //记录上一次发奖励高度
	importBlockHeight    uint64    //当前正在导入的区块高度,临时变量,不需要加入快照
	witnessRewardPool    *sync.Map //见证人奖励池
	communityRewardPool  *sync.Map //社区奖励池,包含该社区下的轻节点奖励
	nameNetRewardPool    *sync.Map //链上域名奖励池(基金会,投资人等)
	witnessMapCommunitys *sync.Map //见证人与社区的关系映射
	communityMapLights   *sync.Map //社区与轻节点的关系映射
	txAveGas             *TxAveGas //全网平均Gas
	lastVoteOp           *sync.Map //上次投票操作高度
	addrReward           *sync.Map //该地址的累计奖励
	freeGasAddrSet       *sync.Map //key:address;value=*DepositFreeGas 免gas地址集合
	tmpCacheContract     *sync.Map //临时记录缓存合约地址,不要快照
}

// 地址奖励
type AddrReward struct {
	Address crypto.AddressCoin //地址
	Value   *big.Int           //金额
	Type    int32              //1=见证人,2=社区,3=轻节点,4=其他(包含基金会,投资人等)
	isPool  bool               //是否奖励池事件
}

// 质押免gas费地址
type DepositFreeGasItem struct {
	ContractAddresses crypto.AddressCoin
	Owner             crypto.AddressCoin
	Deposit           uint64
	LimitHeight       uint64
	LimitCount        uint64
}

func NewBalanceManager(wb *WitnessBackup, tm *TransactionManager, chain *Chain) *BalanceManager {
	bm := &BalanceManager{
		chain:         chain,
		syncBlockHead: make(chan *BlockHeadVO, 1), //正在同步的余额，准备导入到余额中
		// notspentBalance:   NewTxItemManager(),         //
		// notspentBalanceDB: NewTxItemManagerDB(),       //
		witnessBackup:    wb, //
		txManager:        tm, //
		depositWitness:   new(sync.Map),
		depositCommunity: new(sync.Map), //
		depositLight:     new(sync.Map), //
		depositVote:      new(sync.Map), //
		otherDeposit:     new(sync.Map), //
		communityVote:    new(sync.Map),
		addBlockNum:      new(sync.Map),
		addBlockReward:   new(sync.Map),
		blockIndex:       new(sync.Map),
		// cacheTxLockLock:  new(sync.RWMutex), //
		cacheTxlock:  new(sync.Map),
		eventBus:     event.NewEventBus(),
		witnessRatio: new(sync.Map),
		witnessVote:  new(sync.Map),
		txAveGas: &TxAveGas{
			mux:    &sync.RWMutex{},
			AllGas: [100]uint64{},
			Index:  0,
		},
		lastDistributeHeight: 0,
		witnessRewardPool:    new(sync.Map),
		communityRewardPool:  new(sync.Map),
		nameNetRewardPool:    new(sync.Map),
		witnessMapCommunitys: new(sync.Map),
		communityMapLights:   new(sync.Map),
		lastVoteOp:           new(sync.Map),
		addrReward:           new(sync.Map),
		freeGasAddrSet:       new(sync.Map),
		tmpCacheContract:     new(sync.Map),
	}
	utils.Go(bm.run, nil)
	// go bm.run()
	return bm
}
func (this *BalanceManager) GetEventBus() event.EventBus {
	return this.eventBus
}

// 获取全网平均Gas
func (this *BalanceManager) GetTxAveGas() TxAveGas {
	return *this.txAveGas
}

// 获取见证人奖励池
func (this *BalanceManager) GetWitnessRewardPool() sync.Map {
	return *this.witnessRewardPool
}

// 获取社区奖励池
func (this *BalanceManager) GetCommunityRewardPool() sync.Map {
	return *this.communityRewardPool
}

// 获取社区质押关系
func (this *BalanceManager) GetDepositCommunityMap() sync.Map {
	return *this.depositCommunity
}

// 获取见证人与社区映射关系
func (this *BalanceManager) GetWitnessMapCommunitys() sync.Map {
	return *this.witnessMapCommunitys
}

// 获取社区与轻节点映射关系
func (this *BalanceManager) GetCommunityMapLights() sync.Map {
	return *this.communityMapLights
}

// 获取出块索引
func (this *BalanceManager) GetBlockIndex(addr string) uint64 {
	index, ok := this.blockIndex.Load(strings.ToLower(addr))
	if !ok {
		return 0
	}
	return index.(uint64)
}

// 设置合约缓存
func (this *BalanceManager) SetCacheContract(setAddrs []crypto.AddressCoin, delAddrs []crypto.AddressCoin) []crypto.AddressCoin {
	for _, addr := range setAddrs {
		this.tmpCacheContract.Store(utils.Bytes2string(addr), true)
	}
	for _, addr := range delAddrs {
		this.tmpCacheContract.Delete(utils.Bytes2string(addr))
	}

	addrs := []crypto.AddressCoin{}
	this.tmpCacheContract.Range(func(key, value any) bool {
		addrs = append(addrs, crypto.AddressCoin(key.(string)))
		return true
	})

	return addrs
}

// 获取节点出块数量和奖励
func (this *BalanceManager) GetAddBlockNum(coin string) (uint64, uint64) {
	count := uint64(0)
	reward := uint64(0)
	addBlockNum, ok := this.addBlockNum.Load(coin)
	if ok {
		count = addBlockNum.(uint64)
	}

	addBlockReward, ok := this.addBlockReward.Load(coin)
	if ok {
		reward = addBlockReward.(uint64)
	}
	return count, reward
}

/*
	获取一个地址的余额列表
*/
// func (this *BalanceManager) FindBalanceOne(addr *crypto.AddressCoin) (*Balance, *Balance) {
// 	bas, bafrozens := this.FindBalance(addr)
// 	var ba, bafrozen *Balance
// 	if bas != nil || len(bas) >= 1 {
// 		// fmt.Println("这里错误222")
// 		ba = bas[0]
// 	}
// 	if bafrozens != nil || len(bafrozens) >= 1 {
// 		bafrozen = bafrozens[0]
// 	}
// 	return ba, bafrozen
// }

/*
获取本节点见证人押金
*/
func (this *BalanceManager) GetDepositIn() *TxItem {
	return this.nodeWitness
}

/*
获取见证人押金
*/
func (this *BalanceManager) GetDepositWitness(addr *crypto.AddressCoin) (item *DepositInfo) {
	v, ok := this.depositWitness.Load(utils.Bytes2string(*addr))
	if !ok {
		return
	}
	b := v.(*DepositInfo)
	// item = b.item
	return b
}

/*
获取社区节点押金
*/
func (this *BalanceManager) GetDepositCommunity(addr *crypto.AddressCoin) (item *DepositInfo) {
	v, ok := this.depositCommunity.Load(utils.Bytes2string(*addr))
	if !ok {
		return
	}
	b := v.(*DepositInfo)
	// item = b.item
	return b
}

/*
获取轻节点押金
*/
func (this *BalanceManager) GetDepositLight(addr *crypto.AddressCoin) (item *DepositInfo) {
	v, ok := this.depositLight.Load(utils.Bytes2string(*addr))
	if !ok {
		return
	}
	item = v.(*DepositInfo)
	return
}

/*
获取所有质押的轻节点
*/
func (this *BalanceManager) GetAllLights() []*DepositInfo {
	items := []*DepositInfo{}
	this.depositLight.Range(func(key, value any) bool {
		item := value.(*DepositInfo)
		items = append(items, item)
		return true
	})

	return items
}

/*
获取投票押金
*/
func (this *BalanceManager) GetDepositVote(addr *crypto.AddressCoin) *DepositInfo {
	v, ok := this.depositVote.Load(utils.Bytes2string(*addr))
	if !ok {
		return nil
	}
	b := v.(*DepositInfo)
	return b
}

/*
	获取一个地址的押金列表
*/
// func (this *BalanceManager) GetVoteInByTxid(voteAddr *crypto.AddressCoin) *TxItem {
// 	var tx *TxItem

// 	this.votein.Range(func(k, v interface{}) bool {
// 		b := v.(*Balance)
// 		b.Txs.Range(func(txidItr, v interface{}) bool {
// 			// dstTxid := txidItr.(string)
// 			txItem := v.(*TxItem)
// 			//0600000000000000b027d84883693a16de4df892c4d856cbf103ed0e28a2d5d98277199ea2d79345_0
// 			if bytes.Equal(txid, txItem.Txid) {
// 				tx = txItem
// 				return false
// 			}

// 			// if utils.Bytes2string(txid) == strings.SplitN(dstTxid, "_", 2)[0] {
// 			// 	tx = v.(*TxItem)
// 			// 	return false
// 			// }
// 			return true
// 		})
// 		if tx != nil {
// 			return false
// 		}
// 		return true
// 	})
// 	return tx
// }

/*
	统计多个地址的余额
*/
// func (this *BalanceManager) FindBalance(addrs crypto.AddressCoin) (uint64, uint64) {

// 	var count, countf uint64
// 	bas := this.notspentBalance.FindBalanceNotSpent(addrs)
// 	//统计冻结的余额
// 	bafrozens := this.notspentBalance.FindBalanceFrozen(addrs)
// 	// txitems, bfs := chain.balance.FindBalance(addr)
// 	for _, item := range bas {
// 		count = count + item.Value
// 	}

// 	for _, item := range bafrozens {
// 		// engine.Log.Info("统计冻结 111 %d", item.Value)
// 		countf = countf + item.Value
// 	}
// 	// engine.Log.Info("统计冻结 222 %d", countf)
// 	return count, countf

// }

/*
	查询各种状态的余额
*/
// func (this *BalanceManager) FindBalanceValue() (n, f, l uint64) {
// 	// if config.Wallet_txitem_save_db {
// 	// 	dbn, dbf, dbl := this.notspentBalanceDB.FindBalanceValue(this.chain.GetCurrentBlock())

// 	// 	// memn, _, _ := this.notspentBalance.FindBalanceValue()
// 	// 	// if dbn != memn {
// 	// 	// 	engine.Log.Info("Unequal mem:%d db:%d memTotal:%d dbTotal:%d", memn, dbn, memUTXOTotal, dbUTXOTotal)
// 	// 	// 	config.DBUG_import_height_max = GetLongChain().GetCurrentBlock()

// 	// 	// 	txItems := this.notspentBalance.FindBalanceFrozenAll()
// 	// 	// 	txItems = append(txItems, this.notspentBalance.FindBalanceLockupAll()...)
// 	// 	// 	txItems = append(txItems, this.notspentBalance.FindBalanceNotSpentAll()...)

// 	// 	// 	engine.Log.Info("txitem total:%d", len(txItems))

// 	// 	// }
// 	// 	return dbn, dbf, dbl
// 	// } else {
// 	return this.notspentBalance.FindBalanceValue()
// 	// }
// 	// return 0, 0, 0
// }

/*
	从最后一个块开始统计多个地址的余额
*/
// func (this *BalanceManager) FindBalanceAll() (map[string]uint64, map[string]uint64, map[string]uint64) {
// 	if config.Wallet_txitem_save_db {
// 		return this.notspentBalance.FindBalanceAllAddrs()
// 	} else {
// 		return this.notspentBalance.FindBalanceAllAddrs()
// 	}

// 	// bas := this.notspentBalance.FindBalanceNotSpentAll()
// 	// //统计冻结的余额
// 	// bafrozens := this.notspentBalance.FindBalanceFrozenAll()
// 	// //统计锁仓的余额
// 	// baLockup := this.notspentBalance.FindBalanceLockupAll()
// 	// return bas, bafrozens, baLockup
// }

/*
	从最后一个块开始统计多个地址的余额
*/
// func (this *BalanceManager) FindBalanceAll() ([]*TxItem, []*TxItem, []*TxItem) {
// 	bas := this.notspentBalance.FindBalanceNotSpentAll()
// 	//统计冻结的余额
// 	bafrozens := this.notspentBalance.FindBalanceFrozenAll()
// 	//统计锁仓的余额
// 	baLockup := this.notspentBalance.FindBalanceLockupAll()
// 	return bas, bafrozens, baLockup
// }

/*
	构建付款输入
	当扣款账户地址参数不为空，则从指定的扣款账户扣款，
	即使金额不够，也不会从其他账户扣款

	@srcAddress    *crypto.AddressCoin    扣款账户地址
	@amount        uint64                 需要的金额
*/
// func (this *BalanceManager) BuildPayVin(srcAddress crypto.AddressCoin, amount uint64) (uint64, []*TxItem) {

// 	var tis TxItemSort = make([]*TxItem, 0)

// 	//当指定扣款账户地址，则从指定地址扣款
// 	if srcAddress != nil {
// 		if config.Wallet_txitem_save_db {
// 			tis = this.notspentBalanceDB.FindBalanceByAddr(srcAddress, this.chain.GetCurrentBlock(), amount)
// 		} else {
// 			tis = this.notspentBalance.FindBalanceNotSpent(srcAddress)
// 		}
// 	} else {
// 		if config.Wallet_txitem_save_db {
// 			tis = this.notspentBalanceDB.FindBalanceAll(this.chain.GetCurrentBlock(), amount)
// 		} else {
// 			tis = this.notspentBalance.FindBalanceNotSpentAll()
// 		}
// 	}
// 	// fmt.Printf("查询到的有余额的交易 %d %+v\n", tis)

// 	if len(tis) <= 0 {
// 		return 0, nil
// 	}

// 	sort.Sort(&tis)

// 	total := uint64(0)
// 	items := make([]*TxItem, 0, len(tis))
// 	if tis[0].Value >= amount {
// 		item := tis[0]
// 		for i, one := range tis {
// 			if amount >= one.Value {
// 				item = tis[i]
// 				if i == len(tis)-1 {
// 					items = append(items, item)
// 					total = item.Value
// 				}
// 			} else {
// 				items = append(items, item)
// 				total = item.Value
// 				break
// 			}
// 		}
// 	} else {
// 		for i, one := range tis {
// 			items = append(items, tis[i])
// 			total = total + one.Value
// 			if total >= amount {
// 				break
// 			}
// 		}
// 	}
// 	return total, items
// }

func (this *BalanceManager) BuildPayVinNew(srcAddress *crypto.AddressCoin, amount uint64) (uint64, *TxItem) {
	if srcAddress != nil && len(*srcAddress) > 0 {
		//当指定扣款账户地址，则从指定地址扣款
		notspend, _, _ := GetBalanceForAddrSelf(*srcAddress)
		item := TxItem{
			Addr:  srcAddress,
			Value: notspend,
		}
		// tis, value := GetNotSpendBalance(srcAddress)
		return notspend, &item
	} else {
		// GetBalanceForAddr()
		var addr *crypto.AddressCoin
		var notspend uint64
		for _, one := range Area.Keystore.GetAddrAll() {
			addr = &one.Addr
			notspend, _, _ = GetBalanceForAddrSelf(one.Addr)
			if notspend < amount {
				continue
			}
			lockValue, _ := this.FindLockTotalByAddr(&one.Addr)
			notspend = notspend - lockValue
			if notspend < amount {
				continue
			}
			break
		}
		item := TxItem{
			Addr:  addr,
			Value: notspend,
		}

		// tis, value := FindNotSpendBalance(amount)
		return notspend, &item
	}
}

/*
引入最新的块
将交易计入余额
使用过的UTXO余额删除
*/
func (this *BalanceManager) CountBalanceForBlock(bhvo *BlockHeadVO) {
	this.countBlock(bhvo)

	//给已经确认的区块建立高度索引
	// engine.Log.Info("保存索引 %s", config.BlockHeight+strconv.Itoa(int(bhvo.BH.Height)))
	//db.LevelDB.Save([]byte(config.BlockHeight+strconv.Itoa(int(bhvo.BH.Height))), &bhvo.BH.Hash)
	db.LevelDB.Save(append(config.BlockHeight, []byte(strconv.Itoa(int(bhvo.BH.Height)))...), &bhvo.BH.Hash)
}

func (this *BalanceManager) run() {
	for bhvo := range this.syncBlockHead {
		this.countBlock(bhvo)
	}
}

/*
开始统计余额
*/
func (this *BalanceManager) countBlock(bhvo *BlockHeadVO) {
	this.countTotal++
	if this.countTotal == bhvo.BH.Height-config.Mining_block_start_height+1 {
		engine.Log.Info("count block group:%d height:%d total:%d %s",
			bhvo.BH.GroupHeight, bhvo.BH.Height, this.countTotal, hex.EncodeToString(bhvo.BH.Hash))
	} else {
		engine.Log.Error("count block group:%d height:%d total:%d Unequal", bhvo.BH.GroupHeight, bhvo.BH.Height, this.countTotal)
	}

	// start := config.TimeNow()

	// witness, _ := GetLongChain().GetLastBlock()
	// if witness != nil {
	// 	engine.Log.Info("打印最新区块出块时间 %s", utils.FormatTimeToSecond(time.Unix(witness.CreateBlockTime, 0)))
	// }

	// for _, one := range bhvo.Txs {
	// 	engine.Log.Info("统计到的交易hash %s", one.GetHashStr())
	// }

	SaveTxToBlockHead(bhvo)

	//统计社区奖励
	//this.countCommunityReward(bhvo)
	// engine.Log.Info("统计社区奖励耗时 %d %s", bhvo.BH.Height, config.TimeNow().Sub(start))

	//及时将已花费的交易作废，新的余额添加索引
	// this.MarkTransactionAsUsed(bhvo)
	// engine.Log.Info("余额添加索引耗时 %d %s", bhvo.BH.Height, config.TimeNow().Sub(start))

	CountBalanceOther(this.otherDeposit, bhvo)
	// engine.Log.Info("统计其他类型的交易耗时 %d %s", bhvo.BH.Height, config.TimeNow().Sub(start))

	// 见证人,社区,轻节点发奖励
	this.distributeReward(bhvo)

	//及时统计押金及投票
	this.countDepositAndVote(bhvo)

	//统计社区节点的奖励
	countCommunityVoteReward(bhvo)
	// engine.Log.Info("及时统计押金及投票耗时 %d %s", bhvo.BH.Height, config.TimeNow().Sub(start))

	//异步统计，包括可用余额累计和历史记录
	// this.SyncCountOthre(bhvo)
	// engine.Log.Info("统计交易及其他耗时 %d %s", bhvo.BH.Height, config.TimeNow().Sub(start))

	// start := config.TimeNow()
	//统计交易余额
	// this.countBalances(bhvo)
	// engine.Log.Info("区块高度 %d 统计交易余额耗时 %s", bhvo.BH.Height, config.TimeNow().Sub(start))

	//清理本节点保存的交易缓存
	this.CleanCacheTx(bhvo)

	start := config.TimeNow()
	//新版统计余额，统计全网地址余额
	this.CountBalancesNew(bhvo)
	// engine.Log.Info("countBalancesNew time:%s", config.TimeNow().Sub(start))
	if config.TimeNow().Sub(start).Seconds() > 1 {
		engine.Log.Info("countBalancesNew time too long")
	}

	//统计合约
	this.CountContract(bhvo)
	//统计NFT
	this.countNFT(bhvo)
	//统计多签
	this.CountMultTx(bhvo)
	//统计质押免gas费
	this.countDepositFreeGas(bhvo)

	//统计交易历史记录
	this.countTxHistory(bhvo)
	// engine.Log.Info("区块高度 %d 统计交易历史记录耗时 %s", bhvo.BH.Height, config.TimeNow().Sub(start))
	//统计合约事件
	//this.countContractEvents(bhvo)

	//冻结的交易未上链，回滚
	// if !this.chain.SyncBlockFinish {
	// 	//未同步完成，不回滚交易
	// 	return
	// }
	this.Unfrozen(bhvo.BH.Height-1, bhvo.BH.Time)
	// engine.Log.Info("区块高度 %d 交易回滚耗时 %s", bhvo.BH.Height, config.TimeNow().Sub(start))
	config.UpdateSoreName(bhvo.BH.Height)

	//排除已撮合订单
	oBook.handlePool()
	//撮合订单
	oBook.Matching(this.chain)
}

/*
清理关于本节点的交易缓存，余额解锁
*/
func (this *BalanceManager) CleanCacheTx(bhvo *BlockHeadVO) {
	if bhvo.BH.Height == config.Mining_block_start_height {
		return
	}
	for _, one := range bhvo.Txs {
		if one.Class() == config.Wallet_tx_type_mining {
			continue
		}
		this.DelLockTx(one)
	}
}

/*
	异步统计其他的
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
// func (this *BalanceManager) SyncCountOthre(bhvo *BlockHeadVO) {
// 	// start := config.TimeNow()
// 	//统计交易余额
// 	// this.countBalances(bhvo)
// 	// engine.Log.Info("区块高度 %d 统计交易余额耗时 %s", bhvo.BH.Height, config.TimeNow().Sub(start))

// 	start := config.TimeNow()
// 	//新版统计余额，统计全网地址余额
// 	this.countBalancesNew(bhvo)
// 	engine.Log.Info("countBalancesNew time:%s", config.TimeNow().Sub(start))
// 	if config.TimeNow().Sub(start).Seconds() > 1 {
// 		engine.Log.Info("countBalancesNew time too long")
// 	}

// 	//统计交易历史记录
// 	this.countTxHistory(bhvo)
// 	// engine.Log.Info("区块高度 %d 统计交易历史记录耗时 %s", bhvo.BH.Height, config.TimeNow().Sub(start))

// 	//冻结的交易未上链，回滚
// 	if !this.chain.SyncBlockFinish {
// 		//未同步完成，不回滚交易
// 		return
// 	}
// 	this.Unfrozen(bhvo.BH.Height-1, bhvo.BH.Time)
// 	// engine.Log.Info("区块高度 %d 交易回滚耗时 %s", bhvo.BH.Height, config.TimeNow().Sub(start))

// }

/*
	及时将已花费的交易作废，新的余额添加索引
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
// func (this *BalanceManager) MarkTransactionAsUsed(bhvo *BlockHeadVO) {

// 	//查询排除的交易
// 	// excludeTx := make([]config.ExcludeTx, 0)
// 	// for i, one := range config.Exclude_Tx {
// 	// 	if bhvo.BH.Height == one.Height {
// 	// 		excludeTx = append(excludeTx, config.Exclude_Tx[i])
// 	// 	}
// 	// }

// 	for _, txItr := range bhvo.Txs {
// 		// engine.Log.Info("开始标记交易 %s", hex.EncodeToString(*txItr.GetHash()))
// 		//排除的交易不验证
// 		// for _, two := range excludeTx {
// 		// 	if bytes.Equal(two.TxByte, *txItr.GetHash()) {
// 		// 		continue
// 		// 	}
// 		// }

// 		//删除未导入区块的交易无效的标记。
// 		// txidBs := txItr.GetHash()
// 		db.LevelDB.Remove(config.BuildTxNotImport(*txItr.GetHash()))

// 		// txHashStr := txItr.GetHashStr()
// 		// engine.Log.Info("开始标记交易余额 %s", txHashStr)
// 		for _, vin := range *txItr.GetVin() {
// 			//是区块奖励
// 			if txItr.Class() == config.Wallet_tx_type_mining {
// 				continue
// 			}
// 			// preTxHashStr := vin.GetTxidStr()
// 			//删除一个已经使用了的交易输出
// 			db.LevelTempDB.Remove(BuildKeyForUnspentTransaction(vin.Txid, vin.Vout))
// 			// engine.Log.Info("删除一个已经使用过的交易输出 %s", BuildKeyForUnspentTransaction(preTxHashStr, vin.Vout))
// 		}

// 		//生成新的UTXO收益，保存到列表中
// 		for voutIndex, vout := range *txItr.GetVout() {
// 			//保存未使用的vout索引
// 			bs := utils.Uint64ToBytes(vout.FrozenHeight)
// 			// db.Save(BuildKeyForUnspentTransaction(txHashStr, uint64(voutIndex)), &bs)
// 			db.LevelTempDB.Save(BuildKeyForUnspentTransaction(*txItr.GetHash(), uint64(voutIndex)), &bs)
// 			// engine.Log.Info("添加一个交易输出 %s", BuildKeyForUnspentTransaction(txHashStr, uint64(voutIndex)))
// 		}
// 	}
// }

/*
统计押金和投票
@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countDepositAndVote(bhvo *BlockHeadVO) {
	// engine.Log.Info("开始统计交易中的投票 000")
	for _, txItr := range bhvo.Txs {

		txItr.BuildHash()
		// txHashStr := txItr.GetHashStr()

		txCtrl := GetTransactionCtrl(txItr.Class())
		if txCtrl != nil {
			// txCtrl.CountBalance(this.notspentBalance, this.otherDeposit, bhvo, uint64(txIndex))
			return
		}

		switch txItr.Class() {
		case config.Wallet_tx_type_mining:
		case config.Wallet_tx_type_deposit_in:
			voutOne := (*txItr.GetVout())[0]
			addr := voutOne.Address //质押/投票目标地址
			value := voutOne.Value  //质押/投票金额
			SetDepositWitnessAddr(&voutOne.Address, voutOne.Value)

			item := &DepositInfo{
				WitnessAddr: nil,
				SelfAddr:    addr,
				Value:       value,
				Name:        string(txItr.GetPayload()),
				Height:      bhvo.BH.Height,
			}
			this.depositWitness.Store(utils.Bytes2string(addr), item)

			//和自己无关的地址
			if !voutOne.CheckIsSelf() {
				continue
			}
			this.nodeWitness = &TxItem{
				Addr:  &voutOne.Address,
				Value: voutOne.Value,
				Name:  string(txItr.GetPayload()),
			}
			//创始节点直接打开挖矿
			if config.InitNode {
				config.AlreadyMining = true
			}
			//自己提交见证人押金后，再打开出块的开关
			if config.SubmitDepositin {
				config.AlreadyMining = true
			}
		case config.Wallet_tx_type_deposit_out:
			voutOne := (*txItr.GetVout())[0]
			addr := voutOne.Address //质押/投票目标地址
			RemoveDepositWitnessAddr(&addr)

			this.depositWitness.Delete(utils.Bytes2string(addr))

			// engine.Log.Info("取消见证人押金:%s %s %d", vin.GetPukToAddr().B58String(), (*txItr.GetVout())[0].Address, (*txItr.GetVout())[0].Value)
			if !voutOne.CheckIsSelf() {
				continue
			}
			this.nodeWitness = nil
		case config.Wallet_tx_type_pay:
		case config.Wallet_tx_type_vote_in:
			voutOne := (*txItr.GetVout())[0]
			voteIn := txItr.(*Tx_vote_in)

			addr := voutOne.Address //质押/投票目标地址
			dstAddr := voteIn.Vote  //质押/投票目标地址
			value := voutOne.Value  //质押/投票金额

			//记录操作高度
			this.lastVoteOp.Store(utils.Bytes2string(addr), bhvo.BH.Height)

			switch voteIn.VoteType {
			case VOTE_TYPE_community: //社区节点押金
				//记录押金
				SetDepositCommunityAddr(&addr, value)
				name := ""
				if len(txItr.GetPayload()) > 0 {
					name = string(txItr.GetPayload())
				}

				txItem := DepositInfo{
					WitnessAddr: dstAddr, //见证人/社区节点地址
					SelfAddr:    addr,    //轻节点/社区节点地址
					Value:       value,   //押金或投票金额
					Height:      bhvo.BH.Height,
					Name:        name,
				}

				//记录社区质押比例
				this.witnessRatio.Store(utils.Bytes2string(addr), voteIn.Rate)
				//记录投票及关系
				this.depositCommunity.Store(utils.Bytes2string(addr), &txItem)
				this.setWitnessMapCommunitys(dstAddr, addr)
				//保存社区节点开始高度的区块
				blockhash, _ := db.GetTxToBlockHash(txItr.GetHash())
				db.LevelTempDB.Save(config.BuildCommunityAddrStartHeight(addr), blockhash)
				// 更新见证人票数
				if commVote := this.GetCommunityVote(&addr); commVote > 0 {
					this.updateWitnessVote(dstAddr, commVote, true)
				}
			case VOTE_TYPE_vote: //轻节点投票
				lightKey := utils.Bytes2string(addr)
				//添加投票
				if itemItr, ok := this.depositVote.Load(lightKey); ok {
					item := itemItr.(*DepositInfo)
					item.Value += value
					// 更新见证人和社区的票数
					this.updateCommunityVote(dstAddr, value, true)
					if comminfo := this.GetDepositCommunity(&dstAddr); comminfo != nil {
						this.updateWitnessVote(comminfo.WitnessAddr, value, true)
					}
					//更新轻节点票数
					SetDepositLightVoteValue(&item.SelfAddr, &item.WitnessAddr, item.Value)
					continue
				}

				lightName := ""
				if lightinfo := this.GetDepositLight(&addr); lightinfo != nil {
					lightName = lightinfo.Name
				}
				txItem := DepositInfo{
					WitnessAddr: dstAddr, //见证人/社区节点地址
					SelfAddr:    addr,    //轻节点/社区节点地址
					Value:       value,   //押金或投票金额
					Height:      bhvo.BH.Height,
					Name:        lightName,
				}
				this.depositVote.Store(lightKey, &txItem)
				//SetVoteAddr(vin.GetPukToAddr(), &addr)
				//engine.Log.Error("加票更新票数222：%s", addr.B58String())
				this.setCommunityMapLights(dstAddr, addr)
				// 更新见证人和社区的票数
				this.updateCommunityVote(dstAddr, value, true)
				if comminfo := this.GetDepositCommunity(&dstAddr); comminfo != nil {
					this.updateWitnessVote(comminfo.WitnessAddr, value, true)
				}
				//更新轻节点票数
				SetDepositLightVoteValue(&txItem.SelfAddr, &txItem.WitnessAddr, txItem.Value)
			case VOTE_TYPE_light: //轻节点押金
				//记录押金
				SetDepositLightAddr(&addr, value)

				name := ""
				if len(txItr.GetPayload()) > 0 {
					name = string(txItr.GetPayload())
				}
				txItem := DepositInfo{
					WitnessAddr: nil,   //见证人/社区节点地址
					SelfAddr:    addr,  //轻节点/社区节点地址
					Value:       value, //押金或投票金额
					Height:      bhvo.BH.Height,
					Name:        name,
				}
				this.depositLight.Store(utils.Bytes2string(addr), &txItem)
			}
		case config.Wallet_tx_type_vote_out:
			voutOne := (*txItr.GetVout())[0]
			voteOut := txItr.(*Tx_vote_out)

			addr := voutOne.Address //质押/投票地址
			value := voutOne.Value  //质押/投票金额

			//记录操作高度
			this.lastVoteOp.Store(utils.Bytes2string(addr), bhvo.BH.Height)

			switch voteOut.VoteType {
			case VOTE_TYPE_community: //社区节点押金
				addrkey := utils.Bytes2string(addr)
				if itemItr, ok := this.depositCommunity.Load(addrkey); ok {
					item := itemItr.(*DepositInfo)
					this.removeWitnessMapCommunitys(item.WitnessAddr, item.SelfAddr)

					// 减去见证人票数
					if voteItr, ok := this.communityVote.Load(addrkey); ok {
						this.updateWitnessVote(item.WitnessAddr, voteItr.(uint64), false)
					}
				}

				//清除押金记录
				RemoveDepositCommunityAddr(&addr)
				this.depositCommunity.Delete(addrkey)
				//删除记录社区质押比例
				this.witnessRatio.Delete(addrkey)
			case VOTE_TYPE_vote: //轻节点投票
				lightAddr := utils.Bytes2string(addr)
				if itemItr, ok := this.depositVote.Load(lightAddr); ok {
					lightinfo := itemItr.(*DepositInfo)
					lightinfo.Value -= value
					if lightinfo.Value == 0 {
						this.depositVote.Delete(lightAddr)
						this.removeCommunityMapLights(lightinfo.WitnessAddr, lightinfo.SelfAddr)
					}
					// 更新见证人和社区的票数
					this.updateCommunityVote(lightinfo.WitnessAddr, value, false)
					if comminfo := this.GetDepositCommunity(&lightinfo.WitnessAddr); comminfo != nil {
						this.updateWitnessVote(comminfo.WitnessAddr, value, false)
					}

					//更新轻节点票数
					SetDepositLightVoteValue(&lightinfo.SelfAddr, &lightinfo.WitnessAddr, lightinfo.Value)
				}
			case VOTE_TYPE_light: //轻节点押金
				//清除押金记录
				RemoveDepositLightAddr(&addr)
				this.depositLight.Delete(utils.Bytes2string(addr))
			}
		case config.Wallet_tx_type_contract:
			//统计社区和轻节点质押
			//this.countCommunityOrLightDeposit(txItr, bhvo)
		}
	}
}

// 见证人,社区,轻节点发奖励
func (this *BalanceManager) distributeReward(bhvo *BlockHeadVO) {
	startAt := config.TimeNow()
	//跳过创世块
	if bhvo.BH.Height == config.Mining_block_start_height {
		return
	}
	this.importBlockHeight = bhvo.BH.Height //记录当前导入区块高度
	//见证人发奖励直接到账,社区到社区奖励池
	this.distributeWitnessReward(bhvo)
	//社区,轻节点发奖励直接到账
	this.distributeCommunityAndLightReward(bhvo)
	engine.Log.Info("REWARD-DEBUG reward time spent:%s", config.TimeNow().Sub(startAt))
}

// 出块奖励分配到见证人奖励池
// 间隔见证人发奖励到账，同时分发到社区奖励池
func (this *BalanceManager) distributeWitnessReward(bhvo *BlockHeadVO) {
	//是否包含发奖励交易
	//startAt := config.TimeNow()
	hasRewardTx := false
	var rewardTx []byte
	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_mining {
			continue
		}

		hasRewardTx = true
		rewardTx = *txItr.GetHash()

		//原方法
		//vouts := *txItr.GetVout()
		//for _, vout := range vouts {
		//	key := utils.Bytes2string(vout.Address)
		//	isOther := true
		//	//a1.出块奖励分配到见证人奖励池
		//	if _, ok := this.depositWitness.Load(key); ok {
		//		this.saveCurrentWitnessRewardPool(key, new(big.Int).SetUint64(vout.Value))
		//		isOther = false
		//	}

		//	//a2.出块奖励分配域名奖励池
		//	if name.ExistedNamesToNet(config.Name_Foundation, config.Name_investor, config.Name_team, config.Name_store) {
		//		this.saveCurrentNameNetRewardPool(key, new(big.Int).SetUint64(vout.Value))
		//		isOther = false
		//	}

		//	//a3.其它未知类型的奖励直接发放
		//	if isOther {
		//		addrReward := AddrReward{Address: vout.Address, Value: new(big.Int).SetUint64(vout.Value), Type: 4}
		//		// 其它未知奖励发放
		//		this.SaveRewardToAddr(addrReward)
		//	}
		//}

		//NOTE: 解决见证人与团队地址相同时,无法区分奖励问题
		hashKey := utils.Bytes2string(bhvo.BH.Hash)
		if item, ok := witAndRoleRewardSet.Load(hashKey); ok {
			tmpVouts := item.([]*Vout)
			for _, vout := range tmpVouts {
				key := utils.Bytes2string(vout.Address)
				if config.IsRoleReward(vout.AddrStr) { //出块奖励分配域名奖励池
					this.saveCurrentNameNetRewardPool(key, new(big.Int).SetUint64(vout.Value))
				} else if vout.AddrStr == config.Name_reward_v3_witness { //出块奖励分配到见证人奖励池
					this.saveCurrentWitnessRewardPool(key, new(big.Int).SetUint64(vout.Value))
				} else {
					// 其它未知奖励发放
					addrReward := AddrReward{Address: vout.Address, Value: new(big.Int).SetUint64(vout.Value), Type: 4}
					this.SaveRewardToAddr(addrReward)
				}
			}
			witAndRoleRewardSet.Delete(hashKey)
		}
	}

	//清理已经统计过的区块奖励缓存
	witAndRoleRewardSet.Range(func(key, value any) bool {
		hashKey := []byte(key.(string))
		if blockHead, err := LoadBlockHeadByHash(&hashKey); err == nil {
			if GetLongChain().CurrentBlock > blockHead.Height {
				witAndRoleRewardSet.Delete(key)
			}
		}
		return true
	})

	//engine.Log.Info("REWARD-DEBUG 见证人奖励池:%s", config.TimeNow().Sub(startAt))
	//b1.间隔见证人发奖励到账，同时分配社区奖励池
	if bhvo.BH.Height-this.lastDistributeHeight >= config.Mining_Reward_Interval && hasRewardTx {
		ars := []AddrReward{}
		this.witnessRewardPool.Range(func(key, value any) bool {
			witAddr := crypto.AddressCoin(key.(string))
			// 通过见证人地址计算见证人奖励和社区奖励池
			witReward, communityRewardPools := this.CalculateWitnessRewardAndCommunityRewardPools(witAddr, value.(*big.Int))
			//engine.Log.Info("REWARD-DEBUG 见证人总奖励池 地址:%s 总金额:%d 放入:%d", witAddr.B58String(), value.(*big.Int), witReward.Uint64())

			//记录见证人奖励
			if witReward.Cmp(big.NewInt(0)) == 1 {
				ars = append(ars, AddrReward{Address: witAddr, Value: witReward, Type: 1})
			}
			//清空见证人奖励池
			this.witnessRewardPool.Delete(key)
			//暂存社区奖励池
			for communitykey, communityReward := range communityRewardPools {
				communityAddr := crypto.AddressCoin(communitykey)
				//engine.Log.Info("REWARD-DEBUG 放入社区奖励池 地址:%s 金额:%d", communityAddr.B58String(), communityReward.Uint64())
				this.saveCurrentCommunityRewardPool(communitykey, communityReward)
				if r, ok := this.communityRewardPool.Load(communitykey); ok {
					ars = append(ars, AddrReward{Address: communityAddr, Value: r.(*big.Int), Type: 2, isPool: true})
				}
			}

			return true
		})

		this.nameNetRewardPool.Range(func(key, value any) bool {
			nameNetAddr := crypto.AddressCoin(key.(string))
			nameNetReward := value.(*big.Int)
			//记录域名奖励
			if nameNetReward.Cmp(big.NewInt(0)) == 1 {
				ars = append(ars, AddrReward{Address: nameNetAddr, Value: nameNetReward, Type: 4})
			}
			//清空域名奖励池
			this.nameNetRewardPool.Delete(key)
			return true
		})

		//记录发奖高度
		this.lastDistributeHeight = bhvo.BH.Height

		// 见证人和域名奖励发放
		this.SaveRewardToAddr(ars...)
		//engine.Log.Info("REWARD-DEBUG 见证人到账,社区奖励池:%s", config.TimeNow().Sub(startAt))

		//添加见证人分奖励事件
		this.AddCustomRewardTxEvent(rewardTx, ars...)
	}
}

// 通过见证人组计算见证人分奖励交易输入
func (this *BalanceManager) CalculateWitnessRewards(allReward uint64, allVote uint64, witnesses []*crypto.AddressCoin, witVotes map[string]uint64) []*Vout {
	vouts := []*Vout{}
	totalReward := new(big.Int).SetUint64(allReward)
	witCount := len(witVotes) //有70%奖励的见证人数量

	witnessesNum := len(witnesses)

	//移植合约方法distribute
	//30%平分，70%按票数权重分配
	witnessAveRatio := big.NewInt(int64(config.Witness_Ave_Ratio))
	reward30 := new(big.Int).Div(new(big.Int).Mul(totalReward, witnessAveRatio), big.NewInt(100))
	reward70 := new(big.Int).Sub(totalReward, reward30)

	//30%平分奖励
	reward30Avg := new(big.Int).Div(reward30, big.NewInt(int64(witnessesNum)))

	//计算前index位总票数
	totalVote := new(big.Int).SetUint64(allVote)
	//for i, wit := range witnesses {
	//	if i >= witCount {
	//		break
	//	}

	//	witkey := utils.Bytes2string(*wit)
	//	if item, ok := this.witnessVote.Load(witkey); ok {
	//		witVote := int64(item.(uint64))
	//		totalVote.Add(totalVote, big.NewInt(witVote))
	//	}
	//}

	//70%的平均奖励
	reward70Avg := big.NewInt(0)
	if totalVote.Cmp(big.NewInt(0)) == 0 {
		reward70Avg.Div(reward70, big.NewInt(int64(witCount)))
	}

	//见证人的奖励计算
	reward70Used := big.NewInt(0)
	firstWit := crypto.AddressCoin{}
	for i, wit := range witnesses {
		witkey := utils.Bytes2string(*wit)
		if i == 0 {
			firstWit = *witnesses[0]
			//if vote, ok := witVotes[witkey]; ok {
			//	engine.Log.Info("DebugH 见证人股权比例:%p=%s %d/%d", witnesses[i], witnesses[i].B58String(), vote, totalVote.Uint64())
			//}
			//跳过处理第一个见证人
			continue
		}

		reward := new(big.Int).Set(reward30Avg)
		if i < witCount {
			if totalVote.Cmp(big.NewInt(0)) == 1 {
				reward70Avg = big.NewInt(0)
				selfVote := big.NewInt(0)
				if vote, ok := witVotes[witkey]; ok {
					selfVote.SetUint64(vote)
				}
				//if item, ok := this.witnessVote.Load(witkey); ok {
				//	selfVote.SetUint64(item.(uint64))
				//}
				//动态票数比例
				voteRate := calcWitnessVoteRatio(selfVote, totalVote, big.NewInt(int64(witCount)))
				//engine.Log.Info("DebugH 见证人股权比例:%p=%s %d/%d", witnesses[i], witnesses[i].B58String(), selfVote.Uint64(), totalVote.Uint64())
				reward70Avg = new(big.Int).Mul(reward70, voteRate)
				reward70Avg.Div(reward70Avg, big.NewInt(1e8))
				reward70Used.Add(reward70Used, reward70Avg)
			}
			reward.Add(reward, reward70Avg)
		}

		vouts = append(vouts, &Vout{
			Value:   reward.Uint64(),
			Address: *wit,
		})
	}

	//处理零钱,发给第一个见证人
	reward := new(big.Int).Add(reward30Avg, reward30.Mod(reward30, big.NewInt(int64(witnessesNum))))
	if totalVote.Cmp(big.NewInt(0)) == 1 {
		reward.Add(reward, reward70)
		reward.Sub(reward, reward70Used)
	} else {
		reward.Add(reward, reward70Avg)
		reward.Add(reward, new(big.Int).Mod(reward70, big.NewInt(int64(witCount))))
	}
	vouts = append(vouts, &Vout{
		Value:   reward.Uint64(),
		Address: firstWit,
	})

	vouts = MergeVouts(&vouts)

	return vouts
}

// 通过见证人地址计算见证人奖励和社区奖励池
func (this *BalanceManager) CalculateWitnessRewardAndCommunityRewardPools(witAddr crypto.AddressCoin, witRewardPool *big.Int) (*big.Int, map[string]*big.Int) {
	witReward := big.NewInt(0)
	communityRewardPools := make(map[string]*big.Int)
	//间隔见证人发奖励到账，同时分配社区奖励池
	//见证人质押比例,uint16
	witkey := utils.Bytes2string(witAddr)
	witRateItr, ok := this.witnessRatio.Load(witkey)
	if !ok {
		return witReward, communityRewardPools
	}
	witRate := new(big.Int).SetInt64(int64(witRateItr.(uint16)))

	witReward.Set(witRewardPool)

	//见证人票数,uint64
	witVote := big.NewInt(0)
	witVoteItr, ok := this.witnessVote.Load(witkey)
	if ok {
		witVote = new(big.Int).SetUint64(witVoteItr.(uint64))
	}

	//计算社区的总奖励
	disReward := big.NewInt(0)
	if witVote.Cmp(big.NewInt(0)) == 1 {
		disReward = disReward.Mul(witRewardPool, witRate)
		disReward.Div(disReward, big.NewInt(100))
	}

	//零钱
	disRewardSurplus := big.NewInt(0)

	//分配给社区
	if disReward.Cmp(big.NewInt(0)) == 1 {
		communityAddrIter, ok := this.witnessMapCommunitys.Load(witkey)
		if !ok {
			return witReward, communityRewardPools
		}
		disRewardSurplus.Set(disReward)

		communityAddrs := communityAddrIter.([]crypto.AddressCoin)
		for _, communityAddr := range communityAddrs {
			communitykey := utils.Bytes2string(communityAddr)
			communityVote, ok := this.communityVote.Load(communitykey)
			if !ok || communityVote.(uint64) == 0 {
				continue
			}

			//社区票数比例
			voteRate := big.NewInt(0)
			if witVote.Cmp(big.NewInt(0)) == 1 {
				voteRate.SetUint64(communityVote.(uint64))
				voteRate.Mul(voteRate, big.NewInt(1e8)) //放大系数1e8
				voteRate.Div(voteRate, witVote)
			}

			communityReward := big.NewInt(0)
			if voteRate.Cmp(big.NewInt(0)) == 1 {
				communityReward.Mul(disReward, voteRate)
				communityReward.Div(communityReward, big.NewInt(1e8))
			}

			if communityReward.Cmp(big.NewInt(0)) == 0 {
				continue
			}

			disRewardSurplus.Sub(disRewardSurplus, communityReward)

			//暂存社区奖励池
			communityRewardPools[communitykey] = communityReward
		}
	}

	//发奖励给见证人,同时处理零钱,发给见证人
	witReward.Sub(witRewardPool, disReward)
	witReward.Add(witReward, disRewardSurplus)

	return witReward, communityRewardPools
}

/*
// 出块奖励分配到见证人奖励池
// 间隔见证人发奖励到账，同时分发到社区奖励池
func (this *BalanceManager) distributeWitnessReward(bhvo *BlockHeadVO) {
	abiObj, _ := abi.JSON(strings.NewReader(precompiled.REWARD_RATE_ABI))
	//是否包含发奖励交易
	hasRewardTx := false
	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_mining || len(txItr.GetPayload()) < 8 {
			continue
		}
		hasRewardTx = true

		//a.出块奖励分配到见证人奖励池
		_, contractVout := (*txItr.GetVin())[0], (*txItr.GetVout())[0]
		tatolReward := big.NewInt(0).SetUint64(contractVout.Value)
		out, _ := abiObj.Methods["distribute"].Inputs.Unpack(txItr.GetPayload()[4:])
		witVmAddrs := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
		index := *abi.ConvertType(out[1], new(big.Int)).(*big.Int)

		witnesses := make([]crypto.AddressCoin, 0)
		for _, addr := range witVmAddrs {
			witnesses = append(witnesses, evmutils.AddressToAddressCoin(addr.Bytes()))
		}

		witnessesNum := len(witnesses)

		//移植合约方法distribute
		//30%平分，70%按票数权重分配
		reward30 := new(big.Int).Div(new(big.Int).Mul(tatolReward, big.NewInt(30)), big.NewInt(100))
		reward70 := new(big.Int).Sub(tatolReward, reward30)

		//30%平分奖励
		reward30Avg := new(big.Int).Div(reward30, big.NewInt(int64(witnessesNum)))

		//计算前index位总票数
		totalVote := big.NewInt(0)
		for i, wit := range witnesses {
			if i >= int(index.Int64()) {
				break
			}

			witkey := utils.Bytes2string(wit)
			if item, ok := this.witnessVote.Load(witkey); ok {
				witVote := int64(item.(uint64))
				totalVote.Add(totalVote, big.NewInt(witVote))
			}
		}

		//70%的平均奖励
		reward70Avg := big.NewInt(0)
		if totalVote.Cmp(big.NewInt(0)) == 0 {
			reward70Avg.Div(reward70, &index)
		}

		//分配
		reward70Used := big.NewInt(0)
		firstWit := crypto.AddressCoin{}
		for i, wit := range witnesses {
			witkey := utils.Bytes2string(wit)
			if i == 0 {
				firstWit = witnesses[0]
				//跳过处理第一个见证人
				continue
			}

			reward := new(big.Int).Set(reward30Avg)
			if i < int(index.Int64()) {
				if totalVote.Cmp(big.NewInt(0)) == 1 {
					reward70Avg = big.NewInt(0)
					selfVote := big.NewInt(0)
					if item, ok := this.witnessVote.Load(witkey); ok {
						selfVote.SetUint64(item.(uint64))
					}
					//动态票数比例
					voteRate := calcWitnessVoteRatio(selfVote, totalVote, &index)
					reward70Avg = new(big.Int).Mul(reward70, voteRate)
					reward70Avg.Div(reward70Avg, big.NewInt(1e8))
					reward70Used.Add(reward70Used, reward70Avg)
				}
				reward.Add(reward, reward70Avg)
			}

			//暂存见证人奖励池
			this.saveCurrentWitnessRewardPool(witkey, reward)
		}

		//处理零钱,发给第一个见证人
		reward := new(big.Int).Add(reward30Avg, reward30.Mod(reward30, big.NewInt(int64(witnessesNum))))
		if totalVote.Cmp(big.NewInt(0)) == 1 {
			reward.Add(reward, reward70)
			reward.Sub(reward, reward70Used)
		} else {
			reward.Add(reward, reward70Avg)
			reward.Add(reward, new(big.Int).Mod(reward70, &index))
		}

		key := utils.Bytes2string(firstWit)
		//暂存第一个见证人奖励池
		this.saveCurrentWitnessRewardPool(key, reward)
	}

	//b.间隔见证人发奖励到账，同时分配社区奖励池
	if bhvo.BH.Height-this.lastDistributeHeight >= config.Mining_Reward_Interval && hasRewardTx {
		this.witnessRewardPool.Range(func(witkey, value any) bool {
			witRewardPool := new(big.Int).SetUint64(value.(uint64))
			//见证人质押比例,uint16
			witRateItr, ok := this.witnessRatio.Load(witkey)
			if !ok {
				return true
			}
			witRate := new(big.Int).SetInt64(int64(witRateItr.(uint16)))

			//见证人票数,uint64
			witVote := big.NewInt(0)
			witVoteItr, ok := this.witnessVote.Load(witkey)
			if ok {
				witVote = new(big.Int).SetUint64(witVoteItr.(uint64))
			}

			//计算社区的总奖励
			disReward := big.NewInt(0)
			if witVote.Cmp(big.NewInt(0)) == 1 {
				disReward = disReward.Mul(witRewardPool, witRate)
				disReward.Div(disReward, big.NewInt(100))
			}

			//零钱
			disRewardSurplus := big.NewInt(0)

			//分配给社区
			if disReward.Cmp(big.NewInt(0)) == 1 {
				communityAddrIter, ok := this.witnessMapCommunitys.Load(witkey)
				if !ok {
					return true
				}
				disRewardSurplus.Set(disReward)

				communityAddrs := communityAddrIter.([]crypto.AddressCoin)
				for _, communityAddr := range communityAddrs {
					communitykey := utils.Bytes2string(communityAddr)
					communityVote, ok := this.communityVote.Load(communitykey)
					if !ok || communityVote.(uint64) == 0 {
						return true
					}

					//社区票数比例
					voteRate := big.NewInt(0)
					if witVote.Cmp(big.NewInt(0)) == 1 {
						voteRate.Div(big.NewInt(0).SetUint64(communityVote.(uint64)), witVote)
					}

					witReward := big.NewInt(0)
					if voteRate.Cmp(big.NewInt(0)) == 1 {
						voteRate.Mul(voteRate, big.NewInt(1e8))
						witReward.Mul(disReward, voteRate)
						witReward.Div(witReward, big.NewInt(1e8))
					}

					if witReward.Cmp(big.NewInt(0)) == 0 {
						return true
					}

					disRewardSurplus.Sub(disRewardSurplus, witReward)

					//暂存社区奖励池
					this.saveCurrentCommunityRewardPool(communitykey, witReward)
				}
			}

			//发奖励给见证人,同时处理零钱,发给见证人
			witReward := new(big.Int).Sub(witRewardPool, disReward)
			witReward.Add(witReward, disRewardSurplus)
			if witReward.Cmp(big.NewInt(0)) == 1 {
				witAddr := crypto.AddressCoin([]byte(witkey.(string)))
				this.SaveRewardToAddr(witAddr, witReward)
				//清空见证人奖励池
				this.witnessRewardPool.Delete(witkey)
			}

			return true
		})

		//记录发奖高度
		this.lastDistributeHeight = bhvo.BH.Height
	}
}
*/

// 社区和轻节点发奖励到账
// 1.取消社区
// 2.轻节点+-票
// 3.手动分账
func (this *BalanceManager) distributeCommunityAndLightReward(bhvo *BlockHeadVO) {
	//startAt := config.TimeNow()
	var communityAddr *crypto.AddressCoin
	for _, txItr := range bhvo.Txs {
		switch txItr.Class() {
		case config.Wallet_tx_type_vote_in:
			//轻节点投票
			voteIn := txItr.(*Tx_vote_in)
			if voteIn.VoteType == VOTE_TYPE_vote {
				communityAddr = &voteIn.Vote
			}
		case config.Wallet_tx_type_vote_out:
			voteOut := txItr.(*Tx_vote_out)
			switch voteOut.VoteType {
			case VOTE_TYPE_community: //社区取消质押
				vinOne := (*txItr.GetVin())[0]
				communityAddr = vinOne.GetPukToAddr()
			case VOTE_TYPE_vote: //轻节点取消投票
				communityAddr = &voteOut.Vote
			}
		case config.Wallet_tx_type_voting_reward:
			//手动提取奖励
			voutOne := (*txItr.GetVout())[0]
			if commInfo := this.GetDepositCommunity(&voutOne.Address); commInfo != nil {
				communityAddr = &commInfo.SelfAddr
			} else if lightInfo := this.GetDepositVote(&voutOne.Address); lightInfo != nil {
				communityAddr = &lightInfo.WitnessAddr
			}
		}

		//社区地址不为空,则触发分奖
		if communityAddr == nil {
			continue
		}

		// 通过社区地址,计算社区和轻节点的奖励
		communityReward, lightRewards := this.CalculateCommunityRewardAndLightReward(*communityAddr)
		ars := []AddrReward{}
		//社区奖励
		if communityReward.Cmp(big.NewInt(0)) == 1 {
			ars = append(ars, AddrReward{Address: *communityAddr, Value: communityReward, Type: 2})
		}
		//轻节点奖励
		for lightkey, lightReward := range lightRewards {
			lightAddr := crypto.AddressCoin(lightkey)
			ars = append(ars, AddrReward{Address: lightAddr, Value: lightReward, Type: 3})
		}

		//社区奖励池清空
		this.communityRewardPool.Delete(utils.Bytes2string(*communityAddr))
		//添加社区分奖励池事件
		ars = append(ars, AddrReward{Address: *communityAddr, Value: big.NewInt(0), Type: 2, isPool: true})

		//engine.Log.Info("REWARD-DEBUG 准备社区奖励:%s", config.TimeNow().Sub(startAt))
		//社区和轻节点奖励发放
		this.SaveRewardToAddr(ars...)
		//engine.Log.Info("REWARD-DEBUG 社区奖励到账:%s", config.TimeNow().Sub(startAt))

		//添加自定义社区和轻节点奖励事件
		this.AddCustomRewardTxEvent(*txItr.GetHash(), ars...)
		//engine.Log.Info("REWARD-DEBUG 社区奖励事件:%s", config.TimeNow().Sub(startAt))
	}
}

// 通过社区地址,计算社区和轻节点的奖励
func (this *BalanceManager) CalculateCommunityRewardAndLightReward(communityAddr crypto.AddressCoin) (*big.Int, map[string]*big.Int) {
	communityReward := big.NewInt(0)
	lightRewards := make(map[string]*big.Int)
	communitykey := utils.Bytes2string(communityAddr)
	//社区奖励池
	communityRewardPoolItr, ok := this.communityRewardPool.Load(communitykey)
	if !ok {
		return communityReward, lightRewards
	}

	communityRewardPool := communityRewardPoolItr.(*big.Int)
	if communityRewardPool.Cmp(big.NewInt(0)) == 0 {
		//社区奖励池=0
		return communityReward, lightRewards
	}

	//engine.Log.Info("REWARD-DEBUG 社区总奖励 地址:%s 金额:%d", communityAddr.B58String(), communityRewardPool.Uint64())

	//社区质押比例
	communityRatioItr, ok := this.witnessRatio.Load(communitykey)
	if !ok {
		return communityReward, lightRewards
	}
	communityRatio := big.NewInt(int64(communityRatioItr.(uint16)))

	//社区总票数
	communityVoteItr, ok := this.communityVote.Load(communitykey)
	if !ok {
		return communityReward, lightRewards
	}
	communityVote := big.NewInt(0).SetUint64(communityVoteItr.(uint64))

	//轻节点
	lightAddrsItr, ok := this.communityMapLights.Load(communitykey)
	if !ok {
		return communityReward, lightRewards
	}
	lightAddrs := lightAddrsItr.([]crypto.AddressCoin)

	//轻节点奖励池
	lightRewardPool := new(big.Int).Mul(communityRewardPool, communityRatio)
	lightRewardPool.Div(lightRewardPool, big.NewInt(100))

	//分配轻节点奖励
	disRewardSurplus := big.NewInt(0)
	if lightRewardPool.Cmp(big.NewInt(0)) == 1 {
		disRewardSurplus.Set(lightRewardPool)
		for _, lightAddr := range lightAddrs {
			lightkey := utils.Bytes2string(lightAddr)
			lightInfoItr, ok := this.depositVote.Load(lightkey)
			if !ok {
				continue
			}
			lightInfo := lightInfoItr.(*DepositInfo)

			//轻节点票数比例
			voteRate := big.NewInt(0)
			if communityVote.Cmp(big.NewInt(0)) == 1 {
				voteRate.SetUint64(lightInfo.Value)
				voteRate.Mul(voteRate, big.NewInt(1e8)) //放大系数1e8
				voteRate.Div(voteRate, communityVote)
			}

			lightReward := big.NewInt(0)
			if voteRate.Cmp(big.NewInt(0)) == 1 {
				lightReward.Mul(lightRewardPool, voteRate)
				lightReward.Div(lightReward, big.NewInt(1e8))
			}
			disRewardSurplus.Sub(disRewardSurplus, lightReward)

			//轻节点奖励
			if lightReward.Cmp(big.NewInt(0)) == 1 {
				lightRewards[lightkey] = lightReward
			}
		}
	}

	//发奖励给社区,同时处理零钱,发给社区
	communityReward.Sub(communityRewardPool, lightRewardPool)
	communityReward.Add(communityReward, disRewardSurplus)

	return communityReward, lightRewards
}

// 保存本轮的见证人奖励到奖励池
func (this *BalanceManager) saveCurrentWitnessRewardPool(witkey any, reward *big.Int) {
	if item, ok := this.witnessRewardPool.LoadOrStore(witkey, reward); ok {
		oldReward := item.(*big.Int)
		oldReward.Add(oldReward, reward)
	}
}

// 保存本轮的链上域名(基金会,投资人等)奖励到奖励池
func (this *BalanceManager) saveCurrentNameNetRewardPool(otherkey any, reward *big.Int) {
	if item, ok := this.nameNetRewardPool.LoadOrStore(otherkey, reward); ok {
		oldReward := item.(*big.Int)
		oldReward.Add(oldReward, reward)
	}
}

// 保存本轮的社区奖励到奖励池
func (this *BalanceManager) saveCurrentCommunityRewardPool(communitykey any, reward *big.Int) {
	if item, ok := this.communityRewardPool.LoadOrStore(communitykey, reward); ok {
		oldReward := item.(*big.Int)
		oldReward.Add(oldReward, reward)
	}
}

// 奖励发放方法
// NOTE: 任何地址的奖励到账只能调用该方法
func (this *BalanceManager) SaveRewardToAddr(addrRewards ...AddrReward) {
	//startAt := config.TimeNow()
	if len(addrRewards) == 0 {
		return
	}

	//批量设置余额集合,合并地址奖励
	setAddrRewards := sync.Map{}
	eg := errgroup.Group{}
	eg.SetLimit(config.CPUNUM)
	for _, ar := range addrRewards {
		//不处理奖励池的地址
		if ar.isPool {
			continue
		}

		ar := ar
		eg.Go(func() error {
			key := utils.Bytes2string(ar.Address)
			//计算该地址的累计奖励
			if v, ok := this.addrReward.LoadOrStore(key, new(big.Int).Set(ar.Value)); ok {
				val := v.(*big.Int)
				val.Add(val, ar.Value)
			}

			//合并地址的奖励
			if item, ok := setAddrRewards.LoadOrStore(key, &AddrReward{
				Address: ar.Address,
				Value:   new(big.Int).Set(ar.Value),
				Type:    ar.Type,
			}); ok {
				val := item.(*AddrReward)
				val.Value.Add(val.Value, ar.Value)
			}

			return nil
		})
	}
	eg.Wait()

	//虚拟机模式,不执行到账
	//if config.EVM_Reward_Enable {
	//	return
	//}

	//批量设置冻结金额
	frozenHeightAddrs := make(map[string][]FrozenValue, 0)
	setBalances := make(map[string]uint64)
	setAddrRewards.Range(func(key, value any) bool {
		keyStr := key.(string)
		addr := crypto.AddressCoin(keyStr)
		items := []FrozenValue{}
		vouts := LinearRelease180Day(addr, value.(*AddrReward).Value.Uint64(), this.importBlockHeight)
		//engine.Log.Info("REWARD-DEBUG 到账金额 地址[Type=%d]:%s 金额:%d", value.(*AddrReward).Type, value.(*AddrReward).Address.B58String(), value.(*AddrReward).Value.Uint64())
		for _, vout := range vouts {
			//小于等于导入区块高度,则直接到账
			if vout.FrozenHeight <= this.importBlockHeight {
				_, oldValue := db.GetNotSpendBalance(&addr)
				setBalances[keyStr] = oldValue + vout.Value
				continue
			}
			items = append(items, FrozenValue{
				Addr:         keyStr,
				Value:        int64(vout.Value),
				FrozenHeight: vout.FrozenHeight,
			})
		}
		frozenHeightAddrs[keyStr] = items
		return true
	})

	//调试打印
	//for _, fhas := range frozenHeightAddrs {
	//	for _, fha := range fhas {
	//		tmpaddr := crypto.AddressCoin(fha.Addr)
	//		engine.Log.Info("REWARD-DEBUG 冻结金额 地址:%s 高度:%d 金额:%d", tmpaddr.B58String(), fha.FrozenHeight, fha.Value)
	//	}
	//}

	//engine.Log.Info("REWARD-DEBUG 准备奖励到账:%s", config.TimeNow().Sub(startAt))
	//批量直接到账
	db.SetNotSpendBalances(setBalances)
	//engine.Log.Info("REWARD-DEBUG 奖励直接到账:%s", config.TimeNow().Sub(startAt))
	//批量处理冻结高度余额
	handleFrozenHeightBalance(frozenHeightAddrs)
	//engine.Log.Info("REWARD-DEBUG 处理完奖励到账:%s", config.TimeNow().Sub(startAt))
}

// 自定义见证人/社区/轻节点奖励事件
func (this *BalanceManager) AddCustomRewardTxEvent(txid []byte, addrRewards ...AddrReward) {
	//批量设置自定义事件,关联交易
	customTxEvents := []*go_protos.CustomTxEvent{}
	customTxPools := []*go_protos.CustomTxEvent{}
	for _, ar := range addrRewards {
		if ar.isPool {
			//是奖励池事件
			customTxPools = append(customTxPools, &go_protos.CustomTxEvent{
				Addr:  ar.Address.B58String(),
				Value: ar.Value.Uint64(),
				Type:  ar.Type,
			})
		} else {
			//是到账事件
			customTxEvents = append(customTxEvents, &go_protos.CustomTxEvent{
				Addr:  ar.Address.B58String(),
				Value: ar.Value.Uint64(),
				Type:  ar.Type,
			})
		}
	}

	//设置自定义交易事件
	if len(customTxEvents) > 0 || len(customTxPools) > 0 {
		e := go_protos.CustomTxEvents{
			TxId:           txid,
			CustomTxEvents: customTxEvents,
			RewardPools:    customTxPools,
		}

		//保存
		if bs, err := e.Marshal(); err == nil {
			db.LevelTempDB.Save(config.BuildCustomTxEvent(txid), &bs)
		}
	}
}

// 本轮的见证人动态分奖比例
func calcWitnessVoteRatio(vote, totalVote, index *big.Int) *big.Int {
	tmpDecimal := new(big.Int).SetInt64(1e8)
	tmpDeposit := new(big.Int).SetUint64(config.Mining_deposit)
	tmpSelfVote := new(big.Int).Add(vote, tmpDeposit)
	tmpSelfVote.Mul(tmpSelfVote, tmpDecimal)
	tmpTotalVote := new(big.Int).Mul(index, tmpDeposit)
	tmpTotalVote.Add(tmpTotalVote, totalVote)
	return tmpSelfVote.Div(tmpSelfVote, tmpTotalVote)
}

/*
统计社区节点地址和社区节点的锁定奖励
@bhvo    *BlockHeadVO    区块头和所有交易
*/
func countCommunityVoteReward(bhvo *BlockHeadVO) {
	// engine.Log.Info("开始统计交易中的投票 000")
	for _, txItr := range bhvo.Txs {

		switch txItr.Class() {
		case config.Wallet_tx_type_mining:
			for _, vout := range *txItr.GetVout() {
				//找出需要统计余额的地址
				if !ExistCommunityAddr(&vout.Address) {
					// engine.Log.Info("不是社区节点:%s", vout.Address.B58String())
					continue
				}
				value := GetCommunityVoteRewardFrozen(&vout.Address)
				SetCommunityVoteRewardFrozen(&vout.Address, value+vout.Value)
				// engine.Log.Info("添加社区锁定余额:%s old:%d now:%d new:%d", vout.Address.B58String(), value, vout.Value, value+vout.Value)
			}
		case config.Wallet_tx_type_vote_in:
			voteIn := txItr.(*Tx_vote_in)
			if voteIn.VoteType == VOTE_TYPE_community {
				//给见证人投票，成为社区节点
				addr := voteIn.Vout[0].Address
				// engine.Log.Info("保存社区节点开始高度的区块:%s", addr.B58String())
				//保存社区节点开始高度的区块
				db.LevelTempDB.Save(config.BuildCommunityAddrStartHeight(addr), &bhvo.BH.Hash)

				// engine.Log.Info("设置社区节点:%s", addr.B58String())
				SetCommunityAddr(&addr)

				//取消后再次质押的情况，不能删除以前锁定的记录
				if !ExistCommunityVoteRewardFrozen(&addr) {
					SetCommunityVoteRewardFrozen(&addr, 0)
				}
			}

		case config.Wallet_tx_type_vote_out:
			vinOne := (*txItr.GetVin())[0]
			voteOut := txItr.(*Tx_vote_out)
			// preTxItr, err := LoadTxBase(vinOne.Txid)
			// if err != nil {
			// 	//TODO 不能解析上一个交易，程序出错退出
			// 	continue
			// }
			if voteOut.VoteType == VOTE_TYPE_community {
				addr := vinOne.GetPukToAddr()
				RemoveCommunityAddr(addr)
			}
			// if vinOne.Vout == 0 && preTxItr.Class() == config.Wallet_tx_type_vote_in {
			// 	votein := preTxItr.(*Tx_vote_in)
			// 	addr := votein.Vout[0].Address
			// 	RemoveCommunityAddr(&addr)
			// 	break
			// }

			//case config.Wallet_tx_type_voting_reward:
			//	vinOne := (*txItr.GetVin())[0]
			//	communitAddr := vinOne.GetPukToAddr()
			//	totalValue := txItr.GetGas() // uint64(0)
			//	for _, voutOne := range *txItr.GetVout() {
			//		//社区分给自己的10%不计入
			//		// if bytes.Equal(voutOne.Address, *communitAddr) {
			//		// 	// engine.Log.Info("社区自己获得10%奖励:%d", voutOne.Value)
			//		// 	continue
			//		// }
			//		SetLightNodeVoteReward(&voutOne.Address, vinOne.GetPukToAddr(), voutOne.Value)
			//		totalValue += voutOne.Value
			//	}
			//	value := GetCommunityVoteRewardFrozen(communitAddr)
			//	// engine.Log.Info("社区此次分配奖励:%s old:%d now:%d new:%d", communitAddr.B58String(), value, totalValue, value-totalValue)
			//	SetCommunityVoteRewardFrozen(communitAddr, value-totalValue)

			//	//如果奖励分配完了，也取消了社区节点，那么删除记录
			//	if value-totalValue == 0 {
			//		//

			//	}

		}

	}
}

/*
	统计交易余额
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
// func (this *BalanceManager) countBalances(bhvo *BlockHeadVO) {

// 	//将txitem集中起来，一次性添加
// 	itemCount := TxItemCount{
// 		Additems: make([]*TxItem, 0),
// 		SubItems: make([]*TxSubItems, 0),
// 		// deleteKey: make([]string, 0),
// 	}
// 	itemsChan := make(chan *TxItemCount, len(bhvo.Txs))
// 	wg := new(sync.WaitGroup)
// 	wg.Add(len(bhvo.Txs))
// 	utils.Go(
// 		func() {
// 			goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
// 			_, file, line, _ := runtime.Caller(0)
// 			engine.AddRuntime(file, line, goroutineId)
// 			defer engine.DelRuntime(file, line, goroutineId)
// 			for i := 0; i < len(bhvo.Txs); i++ {
// 				one := <-itemsChan
// 				if one != nil {
// 					itemCount.Additems = append(itemCount.Additems, one.Additems...)
// 					itemCount.SubItems = append(itemCount.SubItems, one.SubItems...)
// 					// itemCount.deleteKey = append(itemCount.deleteKey, one.deleteKey...)
// 				}
// 				wg.Done()
// 			}
// 		})

// 	NumCPUTokenChan := make(chan bool, runtime.NumCPU()*6)
// 	for _, txItr := range bhvo.Txs {
// 		go this.countBalancesTxOne(txItr, bhvo.BH.Height, NumCPUTokenChan, itemsChan)
// 	}

// 	wg.Wait()
// 	// start := config.TimeNow()
// 	if config.Wallet_txitem_save_db {
// 		this.notspentBalanceDB.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)
// 		// this.notspentBalance.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)
// 	} else {
// 		this.notspentBalance.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)
// 	}
// 	// engine.Log.Info("统计交易 耗时 %s", config.TimeNow().Sub(start))
// }

// /*
// 	统计单个交易余额，方便异步统计
// */
// func (this *BalanceManager) countBalancesTxOne(txItr TxItr, height uint64, tokenCPU chan bool, itemChan chan *TxItemCount) {
// 	// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
// 	// _, file, line, _ := runtime.Caller(0)
// 	// engine.AddRuntime(file, line, goroutineId)
// 	// defer engine.DelRuntime(file, line, goroutineId)
// 	tokenCPU <- false
// 	txItr.BuildHash()

// 	itemCount := txItr.CountTxItems(height)
// 	itemChan <- itemCount
// 	// engine.Log.Info("统单易6耗时 %s %s", txItr.GetHashStr(), config.TimeNow().Sub(start))
// 	<-tokenCPU
// }

/*
统计交易余额
@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) CountBalancesNew(bhvo *BlockHeadVO) {
	//将txitem集中起来，一次性添加
	itemCount := TxItemCountMap{
		AddItems: make(map[string]*map[uint64]int64, 0),
		Nonce:    make(map[string]big.Int),
	}
	itemsChan := make(chan *TxItemCountMap, len(bhvo.Txs))

	// start := config.TimeNow()

	wg := new(sync.WaitGroup)
	wg.Add(len(bhvo.Txs))
	utils.Go(
		func() {
			// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
			// _, file, line, _ := runtime.Caller(0)
			// engine.AddRuntime(file, line, goroutineId)
			// defer engine.DelRuntime(file, line, goroutineId)
			for i := 0; i < len(bhvo.Txs); i++ {
				one := <-itemsChan
				if one != nil {
					//合并余额
					for addrStr, itemMap := range one.AddItems {
						oldItemMap, ok := itemCount.AddItems[addrStr]
						if ok {
							for frozenHeight, value := range *itemMap {
								oldValue, ok := (*oldItemMap)[frozenHeight]
								if ok {
									oldValue += value
									(*oldItemMap)[frozenHeight] = oldValue
								} else {
									(*oldItemMap)[frozenHeight] = value
								}
							}
						} else {
							itemCount.AddItems[addrStr] = itemMap
						}
					}
					//取出最大的nonce
					for addrStr, nonce := range one.Nonce {
						oldNonce, ok := itemCount.Nonce[addrStr]
						if ok {
							if oldNonce.Cmp(&nonce) < 0 {
								itemCount.Nonce[addrStr] = nonce
							}
						} else {
							itemCount.Nonce[addrStr] = nonce
						}
					}
				}
				wg.Done()
			}
		}, nil)

	NumCPUTokenChan := make(chan bool, runtime.NumCPU()*6)
	for _, txItr := range bhvo.Txs {
		go this.countBalancesNewOne(txItr, bhvo.BH.Height, NumCPUTokenChan, itemsChan)
	}

	wg.Wait()
	// engine.Log.Info("count block spend time:%s", config.TimeNow().Sub(start))

	// start = config.TimeNow()

	//记录超出高度的
	frozenHeightAddrs := make(map[string][]FrozenValue)

	//提取交易所有地址
	addrs := make([]crypto.AddressCoin, 0, len(itemCount.AddItems))
	for addrStr := range itemCount.AddItems {
		addr := crypto.AddressCoin(addrStr)
		addrs = append(addrs, addr)
	}

	//批量查询地址余额
	balances := db.GetNotSpendBalances(addrs)

	//待设置的余额集合
	setBalances := make(map[string]uint64)

	//循环查询出来的所有余额
	for k, v := range balances {
		addrStr := utils.Bytes2string(*v.Addr)
		//查询遍历出地址的高度与金额
		for frozenHeight, value := range *itemCount.AddItems[addrStr] {
			//高度<=区块高度，更新旧值，存入setBalances中
			if frozenHeight <= bhvo.BH.Height {
				if value > 0 {
					balances[k].Value += uint64(value)
					setBalances[addrStr] = balances[k].Value
				} else if value < 0 {
					balances[k].Value -= uint64(1 - value - 1)
					setBalances[addrStr] = balances[k].Value
				}
			} else {
				//记录超过的锁定高度信息
				frozenHeightAddrs[addrStr] = append(frozenHeightAddrs[addrStr], FrozenValue{
					Value:        value,
					FrozenHeight: frozenHeight,
				})

			}
		}
	}

	//处理正常高度的余额
	db.SetNotSpendBalances(setBalances)

	//处理锁定高度余额
	handleFrozenHeightBalance(frozenHeightAddrs)

	//保存最新的nonce
	var err error
	for addrStr, nonce := range itemCount.Nonce {
		addr := crypto.AddressCoin([]byte(addrStr))
		err = SetAddrNonce(&addr, &nonce)
		if err != nil {
			engine.Log.Error("SetAddrNonce error:%s", err.Error())
		}
	}

	// engine.Log.Info("count block spend time:%s", config.TimeNow().Sub(start))

	//snap 余额变动更新snapshot
	//AccountSnap.Update(AccountSnap_Balance, setBalances)
}

/*
统计多签交易
@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) CountMultTx(bhvo *BlockHeadVO) {
	for _, txItr := range bhvo.Txs {
		switch txItr.Class() {
		case config.Wallet_tx_type_multsign_addr:
			tx := txItr.(*Tx_Multsign_Addr)
			//存储多签地址
			pukSet := &go_protos.MultsignSet{
				MultAddress: tx.MultAddress,
				Puks:        make([][]byte, len(tx.MultVins)),
				RandNum:     tx.RandNum.Bytes(),
				Name:        tx.Comment,
			}
			for i, vin := range tx.MultVins {
				pukSet.Puks[i] = vin.Puk
			}
			SaveMultsignAddrSet(pukSet)

		case config.Wallet_tx_type_multsign_pay:
			//清除多签请求交易
			key := config.BuildMultsignRequestTx(*txItr.GetHash())
			db.LevelDB.Remove(key)
		case config.Wallet_tx_type_multsign_name:
			var depositIn *sync.Map
			v, ok := this.otherDeposit.Load(config.Wallet_tx_type_account)
			if ok {
				depositIn = v.(*sync.Map)
			} else {
				depositIn = new(sync.Map)
				this.otherDeposit.Store(config.Wallet_tx_type_account, depositIn)
			}
			tx := txItr.(*Tx_Multsign_Name)

			//约定vout为owner
			ownerAddr := tx.Vout[0].Address

			switch tx.NameActionType {
			case NameInActionReg: //域名注册
				nameInfo := &name.Nameinfo{
					Name:       string(tx.Account), //域名
					Owner:      ownerAddr,          //
					Txid:       *txItr.GetHash(),   //交易id
					NetIds:     tx.NetIds,          //节点地址列表
					AddrCoins:  tx.AddrCoins,       //钱包收款地址
					Height:     bhvo.BH.Height,     //注册区块高度
					Deposit:    tx.Vout[0].Value,   //
					IsMultName: true,
				}

				// 判断自己是否多签地址成员
				if IsMultAddrParter(ownerAddr) {
					//保存自己相关的域名到内存
					name.AddName(*nameInfo)
					txItem := &TxItem{
						Addr:  &ownerAddr,
						Value: nameInfo.Deposit, //押金
					}
					//保存押金到内存
					depositIn.Store(string(tx.Account), txItem)
				} else {
					//清除
					name.DelName(tx.Account)
					depositIn.Delete(string(tx.Account))
				}
				//保存域名
				nameinfoBS, _ := nameInfo.Proto()
				db.LevelTempDB.Save(append([]byte(config.Name), tx.Account...), &nameinfoBS)

			case NameInActionTransfer: //域名转让
				if nameInfo := name.FindNameToNet(string(tx.Account)); nameInfo != nil {
					nameInfo.Owner = ownerAddr

					//先判断是否是多签地址
					if _, ok := ExistMultsignAddrSet(ownerAddr); ok {
						nameInfo.IsMultName = true
						// 判断自己是否多签地址成员
						if IsMultAddrParter(ownerAddr) {
							//保存自己相关的域名到内存
							name.AddName(*nameInfo)
							txItem := &TxItem{
								Addr:  &ownerAddr,
								Value: nameInfo.Deposit, //押金
							}
							//保存押金到内存
							depositIn.Store(string(tx.Account), txItem)
						} else {
							//清除
							name.DelName(tx.Account)
							depositIn.Delete(string(tx.Account))
						}
					} else {
						// 非多签地址
						nameInfo.IsMultName = false
						// 判断自己的地址
						if _, ok := config.Area.Keystore.FindAddress(ownerAddr); ok {
							//保存自己相关的域名到内存
							name.AddName(*nameInfo)
							txItem := &TxItem{
								Addr:  &ownerAddr,
								Value: nameInfo.Deposit, //押金
							}
							//保存押金到内存
							depositIn.Store(string(tx.Account), txItem)
						} else {
							//清除
							name.DelName(tx.Account)
							depositIn.Delete(string(tx.Account))
						}
					}
					//保存域名
					nameinfoBS, _ := nameInfo.Proto()
					db.LevelTempDB.Save(append([]byte(config.Name), tx.Account...), &nameinfoBS)
				}

			case NameInActionRenew: //域名续费
				if nameInfo := name.FindNameToNet(string(tx.Account)); nameInfo != nil {
					nameInfo.Height = bhvo.BH.Height
					nameInfo.IsMultName = true

					// 判断自己是否多签地址成员
					if IsMultAddrParter(nameInfo.Owner) {
						//保存自己相关的域名到内存
						name.AddName(*nameInfo)
						txItem := &TxItem{
							Addr:  &nameInfo.Owner,
							Value: nameInfo.Deposit, //押金
						}
						//保存押金到内存
						depositIn.Store(string(tx.Account), txItem)
					} else {
						//清除
						name.DelName(tx.Account)
						depositIn.Delete(string(tx.Account))
					}

					//保存域名
					nameinfoBS, _ := nameInfo.Proto()
					db.LevelTempDB.Save(append([]byte(config.Name), tx.Account...), &nameinfoBS)
				}
			case NameInActionUpdate: //域名更新
				if nameInfo := name.FindNameToNet(string(tx.Account)); nameInfo != nil {
					nameInfo.NetIds = tx.NetIds
					nameInfo.AddrCoins = tx.AddrCoins
					nameInfo.IsMultName = true

					// 判断自己是否多签地址成员
					if IsMultAddrParter(nameInfo.Owner) {
						//保存自己相关的域名到内存
						name.AddName(*nameInfo)
						txItem := &TxItem{
							Addr:  &nameInfo.Owner,
							Value: nameInfo.Deposit, //押金
						}
						//保存押金到内存
						depositIn.Store(string(tx.Account), txItem)
					} else {
						//清除
						name.DelName(tx.Account)
						depositIn.Delete(string(tx.Account))
					}

					//保存域名
					nameinfoBS, _ := nameInfo.Proto()
					db.LevelTempDB.Save(append([]byte(config.Name), tx.Account...), &nameinfoBS)
				}

			case NameOutAction: //域名注销
				//退还押金
				if nameInfo := name.FindNameToNet(string(tx.Account)); nameInfo != nil {
					_, oldValue := db.GetNotSpendBalance(&nameInfo.Owner)
					db.SetNotSpendBalance(&nameInfo.Owner, oldValue+nameInfo.Deposit)
				}
				//清除记录
				name.DelName(tx.Account)
				depositIn.Delete(string(tx.Account))
				db.LevelTempDB.Del(append([]byte(config.Name), tx.Account...))
			}

			//清除多签请求交易
			key := config.BuildMultsignRequestTx(*txItr.GetHash())
			db.LevelDB.Remove(key)
		}
	}

}

//region 优化交易统计

/*
*
批量设置余额
*/
func SetNotSpendBalancesOld(balances map[string]uint64) {
	if len(balances) == 0 {
		return
	}

	//待设置余额的地址集合
	setBalanceAddrs := make([]db2.KVPair, 0)

	//大余额的地址集合
	setBigAddrs := make([]db2.KVPair, 0)

	for k, v := range balances {
		addr := crypto.AddressCoin(k)
		if v > config.Mining_coin_total {
			engine.Log.Info("大余额:%d", v)
			setBigAddrs = append(setBigAddrs, db2.KVPair{
				Key:   config.BuildAddrValueBig(addr),
				Value: nil,
			})
		}
		setBalanceAddrs = append(setBalanceAddrs, db2.KVPair{
			Key:   config.BuildAddrValue(addr),
			Value: utils.Uint64ToBytes(v),
		})
	}

	LedisMultiSaves(setBigAddrs)

	LedisMultiSaves(setBalanceAddrs)
}

/*
*
批量存储
*/
func LedisMultiSaves(vals []db2.KVPair) error {
	if len(vals) == 0 {
		return nil
	}

	//先删除，再保存
	//_, err := db.LevelTempDB.GetDB().Del(func() [][]byte {
	//	r := make([][]byte, len(vals))
	//	for k, v := range vals {
	//		r[k] = v.Key
	//	}
	//	return r
	//}()...)

	//if err != nil {
	//	return err
	//}

	return db.LevelTempDB.MSet(vals...)
}

/*
*
批量查询账户余额
*/
func GetNotSpendBalancesOld(addrs []crypto.AddressCoin) []TxItem {
	if len(addrs) == 0 {
		return nil
	}

	keys := make([][]byte, len(addrs))
	for k, v := range addrs {
		keys[k] = config.BuildAddrValue(v)
	}

	values, err := db.LevelTempDB.MGet(keys...)

	if err != nil {
		return nil
	}

	r := make([]TxItem, len(addrs))
	for k, v := range values {
		addr := addrs[k]
		r[k].Addr = &addr
		r[k].Value = utils.BytesToUint64(v)
	}

	return r
}

type FrozenValue struct {
	Addr         string
	Value        int64
	FrozenHeight uint64
}

type FrozenHeightValue struct {
	Addr  string
	Value uint64
}

/*
*
处理锁定高度余额
*/
func handleFrozenHeightBalance(frozenHeightAddrs map[string][]FrozenValue) {
	//startAt := config.TimeNow()
	if frozenHeightAddrs == nil || len(frozenHeightAddrs) == 0 {
		return
	}

	//组装批量查询的key
	keys := make([][]byte, 0, len(frozenHeightAddrs))

	//记录超过Wallet_frozen_time_min的交易,key为height
	frozenValues := map[uint64][]FrozenHeightValue{}
	//frozenValues := make(map[uint64]map[string]uint64)

	for k, v := range frozenHeightAddrs {
		addr := crypto.AddressCoin(k)
		keys = append(keys, config.BuildAddrFrozen(addr))

		for _, v1 := range v {
			if v1.FrozenHeight < config.Wallet_frozen_time_min {
				frozenValues[v1.FrozenHeight] = append(frozenValues[v1.FrozenHeight],
					FrozenHeightValue{Addr: k, Value: uint64(v1.Value)})

				//_, ok := frozenValues[v1.FrozenHeight]
				//if ok {
				//	frozenValues[v1.FrozenHeight][k] = uint64(v1.Value)
				//} else {
				//	frozenValues[v1.FrozenHeight] = map[string]uint64{k: uint64(v1.Value)}
				//}
			} else {
				//此处无法优化，采用原来的处理方式
				//AddZSetFrozenHeightForTime(&addr, int64(v1.FrozenHeight), uint64(v1.Value))
			}
		}
	}

	//批量查询出所有的历史余额
	oldValues := getAddrFrozenValues(keys)

	if oldValues != nil && len(oldValues) > 0 {
		kvPairs := make([]db2.KVPair, 0, len(keys))
		for k, v := range keys {
			value := uint64(0)

			for _, v1 := range frozenHeightAddrs[string(v[len(config.DBKEY_addr_frozen_value):])] {
				value += uint64(v1.Value)
			}

			kvPairs = append(kvPairs, db2.KVPair{
				Key:   v,
				Value: utils.Uint64ToBytes(oldValues[k] + value),
			})
		}
		LedisMultiSaves(kvPairs)

	}

	//engine.Log.Info("REWARD-DEBUG 批量保存冻结余额:%s", config.TimeNow().Sub(startAt))

	//更新处理超过Wallet_frozen_time_min的交易
	if len(frozenValues) > 0 {
		AddFrozenHeightForAddrValues(frozenValues)
	}

	//engine.Log.Info("REWARD-DEBUG 批量添加冻结高度:%s", config.TimeNow().Sub(startAt))
}

func HandleTokenFrozenHeightBalance(frozenHeightAddrs map[string][]FrozenValue) {
	//startAt := config.TimeNow()
	if frozenHeightAddrs == nil || len(frozenHeightAddrs) == 0 {
		return
	}

	//组装批量查询的key
	keys := make([][]byte, 0, len(frozenHeightAddrs))

	//记录超过Wallet_frozen_time_min的交易,key为height
	frozenValues := map[uint64][]FrozenHeightValue{}
	//frozenValues := make(map[uint64]map[string]uint64)

	for k, v := range frozenHeightAddrs {
		addr := crypto.AddressCoin(k)
		keys = append(keys, config.BuildAddrFrozen(addr))

		for _, v1 := range v {
			if v1.FrozenHeight < config.Wallet_frozen_time_min {
				frozenValues[v1.FrozenHeight] = append(frozenValues[v1.FrozenHeight],
					FrozenHeightValue{Addr: k, Value: uint64(v1.Value)})

				//_, ok := frozenValues[v1.FrozenHeight]
				//if ok {
				//	frozenValues[v1.FrozenHeight][k] = uint64(v1.Value)
				//} else {
				//	frozenValues[v1.FrozenHeight] = map[string]uint64{k: uint64(v1.Value)}
				//}
			} else {
				//此处无法优化，采用原来的处理方式
				//AddZSetFrozenHeightForTime(&addr, int64(v1.FrozenHeight), uint64(v1.Value))
			}
		}
	}

	//批量查询出所有的历史余额
	oldValues := getAddrFrozenValues(keys)

	if oldValues != nil && len(oldValues) > 0 {
		kvPairs := make([]db2.KVPair, 0, len(keys))
		for k, v := range keys {
			value := uint64(0)

			for _, v1 := range frozenHeightAddrs[string(v[len(config.DBKEY_addr_frozen_value):])] {
				value += uint64(v1.Value)
			}

			kvPairs = append(kvPairs, db2.KVPair{
				Key:   v,
				Value: utils.Uint64ToBytes(oldValues[k] + value),
			})
		}
		LedisMultiSaves(kvPairs)

	}

	//engine.Log.Info("REWARD-DEBUG 批量保存冻结余额:%s", config.TimeNow().Sub(startAt))

	//更新处理超过Wallet_frozen_time_min的交易
	if len(frozenValues) > 0 {
		AddFrozenHeightForAddrValues(frozenValues)
	}

	//engine.Log.Info("REWARD-DEBUG 批量添加冻结高度:%s", config.TimeNow().Sub(startAt))
}

/*
*
hash 批量添加到冻结高度
*/
func AddFrozenHeightForAddrValues(frozenValues map[uint64][]FrozenHeightValue) error {
	//startAt := config.TimeNow()
	frozenKfs := make([]db2.KFPair, 0, len(frozenValues))

	frozenKfvs := make([]db2.KFVPair, 0, len(frozenValues))
	heights := make([]uint64, 0, len(frozenValues))

	for k, v := range frozenValues {
		key := config.BuildFrozenHeight(k)
		Addrs := make([][]byte, 0, len(v))
		heights = append(heights, k)

		for _, v1 := range v {
			Addrs = append(Addrs, crypto.AddressCoin(v1.Addr))
		}
		frozenKfs = append(frozenKfs, db2.KFPair{key, Addrs})
		frozenKfvs = append(frozenKfvs, db2.KFVPair{key, nil})
	}

	//engine.Log.Info("REWARD-DEBUG 准备获取批量冻结高度:%s", config.TimeNow().Sub(startAt))

	values, err := db.LevelTempDB.HKMget(frozenKfs)
	if err != nil {
		return err
	}

	//engine.Log.Info("REWARD-DEBUG 获取批量冻结高度:%s", config.TimeNow().Sub(startAt))

	for k := range frozenKfvs {
		fvPairs := make([]db2.FVPair, len(frozenValues[heights[k]]))
		for k1, v1 := range frozenValues[heights[k]] {
			oldValue := uint64(0)
			if values[k][k1] != nil && len(values[k][k1]) > 0 {
				oldValue = utils.BytesToUint64(values[k][k1])
			}

			v1.Value += oldValue

			fvPairs[k1] = db2.FVPair{
				Field: crypto.AddressCoin(v1.Addr),
				Value: utils.Uint64ToBytes(v1.Value),
			}
		}
		frozenKfvs[k].Values = fvPairs
	}

	return db.LevelTempDB.HKMset(frozenKfvs)
}

/*
*
json 批量添加到冻结高度
*/
func AddFrozenHeightForAddrValues1(frozenValues map[uint64]map[string]uint64) error {
	//记录所有key
	keys := make([][]byte, 0, len(frozenValues))
	heights := make([]uint64, 0, len(frozenValues))
	for k := range frozenValues {
		key := config.BuildFrozenHeight(k)
		keys = append(keys, key)
		heights = append(heights, k)
	}
	st := config.TimeNow()
	values := getAddrFrozenHeights(keys)
	engine.Log.Info("锁定高度4_1:%s，长度：%d", config.TimeNow().Sub(st), len(frozenValues))

	st = config.TimeNow()
	kvPairs := make([]db2.KVPair, 0, len(keys))

	for k, v := range keys {
		for k1 := range frozenValues[heights[k]] {
			addr := crypto.AddressCoin(k1)
			addrK := addr.B58String()
			_, ok := values[k][addrK]
			if ok {
				values[k][addrK] += frozenValues[heights[k]][k1]
			} else {
				if values[k] == nil {
					values[k] = make(map[string]uint64)
				}
				values[k][addrK] = frozenValues[heights[k]][k1]
			}
		}

		val, _ := json.Marshal(values[k])
		kvPairs = append(kvPairs, db2.KVPair{
			Key:   v,
			Value: val,
		})
	}
	engine.Log.Info("锁定高度4_2:%s，长度：%d", config.TimeNow().Sub(st), len(frozenValues))

	st = config.TimeNow()
	err := LedisMultiSaves(kvPairs)
	engine.Log.Info("锁定高度4_3:%s，长度：%d", config.TimeNow().Sub(st), len(frozenValues))

	return err
}

/*
*
循环hash批量添加到冻结高度
*/
func AddFrozenHeightForAddrValues2(frozenValues map[uint64][]FrozenHeightValue) error {
	for k, v := range frozenValues {
		key := config.BuildFrozenHeight(k)
		//组装批量查询field的数据
		addrs := make([][]byte, 0, len(v))

		for _, v1 := range v {
			addrs = append(addrs, []byte(v1.Addr))
		}

		//查询出该height所有地址的value
		values, err := db.LevelTempDB.HMget(key, addrs...)
		if err != nil {
			return err
		}

		//组装mset的数据
		fvPairs := make([]db2.FVPair, 0, len(v))
		for k2, v2 := range v {
			oldValue := uint64(0)
			if values[k2] != nil && len(values[k2]) > 0 {
				oldValue = utils.BytesToUint64(values[k2])
			}
			v2.Value += oldValue

			fvPairs = append(fvPairs, db2.FVPair{
				Field: crypto.AddressCoin(v2.Addr),
				Value: utils.Uint64ToBytes(v2.Value),
			})
		}

		db.LevelTempDB.HMset(key, fvPairs...)
	}
	return nil
}

func getAddrFrozenValues(keys [][]byte) []uint64 {
	values, err := db.LevelTempDB.MGet(keys...)
	if err != nil {
		return nil
	}
	r := make([]uint64, 0, len(values))
	for _, v := range values {
		r = append(r, utils.BytesToUint64(v))
	}
	return r
}

func getAddrFrozenHeights(keys [][]byte) []map[string]uint64 {
	values, err := db.LevelTempDB.MGet(keys...)
	if err != nil {
		return nil
	}

	r := make([]map[string]uint64, 0, len(keys))
	for _, v := range values {
		var res map[string]uint64
		json.Unmarshal(v, &res)
		r = append(r, res)
	}
	return r
}

func BytesToInt(bys []byte) uint64 {
	bytebuff := bytes.NewBuffer(bys)
	var data uint64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

func IntToBytes(n uint64) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, n)
	return bytebuf.Bytes()
}

//endregion
/*
	统计交易余额
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countBalancesNewOne(txItr TxItr, height uint64, tokenCPU chan bool, itemChan chan *TxItemCountMap) {
	tokenCPU <- false
	txItr.BuildHash()
	var itemCount *TxItemCountMap
	if txItr.Class() == config.Wallet_tx_type_pay {
		this.countTxAveGas(txItr.GetGas())
	}
	itemCount = txItr.CountTxItemsNew(height)
	itemChan <- itemCount
	<-tokenCPU
}

/*
统计所有块交易平均Gas
*/
func (this *BalanceManager) countTxAveGas(newGas uint64) {
	this.txAveGas.mux.Lock()
	defer this.txAveGas.mux.Unlock()
	if this.txAveGas.Index >= uint64(len(this.txAveGas.AllGas)) {
		this.txAveGas.Index = 0
	}
	this.txAveGas.AllGas[this.txAveGas.Index] = newGas
	this.txAveGas.Index++
}

/*
统计社区奖励
@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countCommunityReward(bhvo *BlockHeadVO) {

	var err error
	var addr crypto.AddressCoin
	var ok bool
	var cs *CommunitySign
	var sn *sqlite3_db.SnapshotReward
	var rt *RewardTotal
	var r *[]sqlite3_db.RewardLight
	for _, txItr := range bhvo.Txs {
		//判断交易类型
		if txItr.Class() != config.Wallet_tx_type_pay {
			continue
		}
		//检查签名
		addr, ok, cs = CheckPayload(txItr)
		if !ok {
			//签名不正确
			continue
		}
		//判断地址是否属于自己
		_, ok = Area.Keystore.FindAddress(addr)
		if !ok {
			//签名者地址不属于自己
			continue
		}

		//判断有没有这个快照
		sn, _, err = FindNotSendReward(&addr)
		if err != nil && err.Error() != xorm.ErrNotExist.Error() {
			engine.Log.Error("querying database Error %s", err.Error())
			return
		}
		//同步快照
		if sn == nil || sn.EndHeight < cs.EndHeight {
			//创建快照
			rt, r, err = GetRewardCount(&addr, cs.StartHeight, cs.EndHeight)
			if err != nil {
				return
			}
			err = CreateRewardCount(addr, rt, *r)
			if err != nil {
				return
			}
		}
	}

}

/*
统计交易历史记录
*/
func (this *BalanceManager) countTxHistory(bhvo *BlockHeadVO) {
	// start := config.TimeNow()
	gas := uint64(0)
	for _, txItr := range bhvo.Txs {

		// start := config.TimeNow()
		txItr.BuildHash()
		// txHashStr := hex.EncodeToString(*txItr.GetHash())
		// engine.Log.Info("统计余额 000")
		//将之前的UTXO标记为已经使用，余额中减去。

		txItr.CountTxHistory(bhvo.BH.Height)
		gas += txItr.GetGas()
	}

	// 添加见证人出块相关数据
	reward := config.ClacRewardForBlockHeightFun(bhvo.BH.Height) + gas
	//reward := config.BLOCK_TOTAL_REWARD + gas
	this.addMinerData(bhvo.BH.Witness, reward, bhvo.BH.Hash)

	//添加创世块奖励统计
	if bhvo.BH.Height == config.Mining_block_start_height {
		this.saveCurrentWitnessRewardPool(utils.Bytes2string(bhvo.BH.Witness), new(big.Int).SetUint64(reward))
	}

	// engine.Log.Info("统计余额及奖励 666 耗时 %s", config.TimeNow().Sub(start))
}

func (this *BalanceManager) addMinerData(witness crypto.AddressCoin, reward uint64, hash []byte) {
	// 添加所有见证者节点出块数量
	addBlockNum, ok := this.addBlockNum.Load(witness.B58String())
	if ok {
		this.addBlockNum.Store(witness.B58String(), addBlockNum.(uint64)+1)
	} else {
		this.addBlockNum.Store(witness.B58String(), uint64(1))
	}

	// 添加见证者节点出块奖励
	addBlockReward, ok := this.addBlockReward.Load(witness.B58String())
	if ok {
		this.addBlockReward.Store(witness.B58String(), addBlockReward.(uint64)+reward)
	} else {
		this.addBlockReward.Store(witness.B58String(), reward)
	}

	// 添加出块记录
	this.addMinerTxToDb(&witness, hash)
}

/*
添加出块记录
*/
func (this *BalanceManager) addMinerTxToDb(inAddr *crypto.AddressCoin, blockHash []byte) error {
	// 记录地址最大索引
	index := uint64(1)
	addBlockIndex, ok := this.blockIndex.Load(strings.ToLower(inAddr.B58String()))
	if ok {
		index = addBlockIndex.(uint64) + 1
		this.blockIndex.Store(strings.ToLower(inAddr.B58String()), index)
	} else {
		this.blockIndex.Store(strings.ToLower(inAddr.B58String()), uint64(1))
	}

	//记录地址交易hash和索引
	//addrkey := []byte(config.Miner_history_tx + "_" + strings.ToLower(inAddr.B58String()))
	addrkey := append(config.Miner_history_tx, []byte("_"+strings.ToLower(inAddr.B58String()))...)
	indexBs := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBs, index)
	_, err := db.LevelTempDB.HSet(addrkey, indexBs, blockHash)
	if err != nil {
		return err
	}

	return err
}

/*
	统计交易历史记录
*/
// func (this *BalanceManager) countTxHistory(bhvo *BlockHeadVO) {
// 	for _, txItr := range bhvo.Txs {

// 		// start := config.TimeNow()
// 		txItr.BuildHash()
// 		// txHashStr := hex.EncodeToString(*txItr.GetHash())
// 		// engine.Log.Info("统计余额 000")
// 		//将之前的UTXO标记为已经使用，余额中减去。

// 		// engine.Log.Info("统计余额及奖励 666 耗时 %s", config.TimeNow().Sub(start))

// 		txCtrl := GetTransactionCtrl(txItr.Class())
// 		if txCtrl != nil {
// 			// txCtrl.CountBalance(this.notspentBalance, this.otherDeposit, bhvo, uint64(txIndex))
// 			continue
// 		}

// 		// engine.Log.Info("统计余额及奖励 777 耗时 %s", config.TimeNow().Sub(start))

// 		//其他类型交易，自己节点不支持，直接当做普通交易处理

// 		// voutAddrs := make([]*crypto.AddressCoin, 0)
// 		// for i, one := range *txItr.GetVout() {
// 		// 	//如果地址是自己的，就可以不用显示
// 		// 	if keystore.FindAddress(one.Address) {
// 		// 		continue
// 		// 	}
// 		// 	voutAddrs = append(voutAddrs, &(*txItr.GetVout())[i].Address)
// 		// }

// 		// engine.Log.Info("统计余额及奖励 888 耗时 %s", config.TimeNow().Sub(start))

// 		vinAddrs := make([]*crypto.AddressCoin, 0)

// 		hasPayOut := false //是否有支付类型转出记录
// 		//将之前的UTXO标记为已经使用，余额中减去。
// 		for _, vin := range *txItr.GetVin() {

// 			isSelf := vin.CheckIsSelf()

// 			//检查地址是否是自己的
// 			// addrInfo, isSelf := keystore.FindPuk(vin.Puk)

// 			// addr, isSelf := vin.ValidateAddr()
// 			// fmt.Println(addr)
// 			vinAddrs = append(vinAddrs, vin.GetPukToAddr())
// 			if !isSelf {
// 				continue
// 			}

// 			//是区块奖励
// 			if txItr.Class() == config.Wallet_tx_type_mining {
// 				continue
// 			}

// 			switch txItr.Class() {
// 			case config.Wallet_tx_type_mining:
// 			case config.Wallet_tx_type_deposit_in:
// 			case config.Wallet_tx_type_deposit_out:

// 				// if preTxItr.Class() == config.Wallet_tx_type_deposit_in {
// 				// 	if this.depositin != nil {
// 				// 		if bytes.Equal(addrInfo.Addr, *this.depositin.Addr) {
// 				// 			this.depositin = nil
// 				// 		}
// 				// 	}
// 				// }
// 			case config.Wallet_tx_type_pay:
// 				if !hasPayOut {
// 					//和自己相关的输入地址
// 					vinAddrsSelf := make([]*crypto.AddressCoin, 0)
// 					for _, vin := range *txItr.GetVin() {
// 						//检查地址是否是自己的
// 						addrInfo, isSelf := keystore.FindPuk(vin.Puk)
// 						// addr, isSelf := vin.ValidateAddr()
// 						// fmt.Println(addr)
// 						if isSelf {
// 							vinAddrsSelf = append(vinAddrsSelf, &addrInfo.Addr)
// 						}
// 					}

// 					temp := make([]*crypto.AddressCoin, 0) //转出地址
// 					amount := uint64(0)                    //转出金额
// 					for i, one := range *txItr.GetVout() {
// 						//如果地址是自己的，就可以不用显示
// 						// if keystore.FindAddress(one.Address) {
// 						// 	continue
// 						// }
// 						_, ok := keystore.FindAddress(one.Address)
// 						if ok {
// 							continue
// 						}
// 						temp = append(temp, &(*txItr.GetVout())[i].Address)
// 						amount = amount + one.Value
// 					}
// 					if amount > 0 {
// 						//将转出保存历史记录
// 						hi := HistoryItem{
// 							IsIn:    false,         //资金转入转出方向，true=转入;false=转出;
// 							Type:    txItr.Class(), //交易类型
// 							InAddr:  temp,          //输入地址
// 							OutAddr: vinAddrsSelf,  //输出地址
// 							// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
// 							Value:  amount,           //交易金额
// 							Txid:   *txItr.GetHash(), //交易id
// 							Height: bhvo.BH.Height,   //
// 							// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
// 						}
// 						this.chain.history.Add(hi)
// 						// engine.Log.Info("转出记录", bhvo.BH.Height, hi, (*preTxItr.GetVout())[vin.Vout].Value)
// 					}
// 					hasPayOut = true
// 				}
// 			case config.Wallet_tx_type_vote_in:
// 			case config.Wallet_tx_type_vote_out:

// 				// if preTxItr.Class() == config.Wallet_tx_type_vote_in {
// 				// 	votein := preTxItr.(*Tx_vote_in)
// 				// 	b, ok := this.votein.Load(votein.Vote.B58String())
// 				// 	if ok {
// 				// 		ba := b.(*Balance)
// 				// 		ba.Txs.Delete(hex.EncodeToString(*preTxItr.GetHash()) + "_" + strconv.Itoa(int(vin.Vout)))
// 				// 		this.votein.Store(votein.Vote.B58String(), ba)
// 				// 	}
// 				// }
// 			}
// 		}

// 		// engine.Log.Info("统计余额及奖励 999 耗时 %s", config.TimeNow().Sub(start))
// 		//生成新的UTXO收益，保存到列表中
// 		for _, vout := range *txItr.GetVout() {

// 			//和自己无关的地址
// 			// if !keystore.FindAddress(vout.Address) {
// 			// 	continue
// 			// }
// 			ok := vout.CheckIsSelf()
// 			// _, ok := keystore.FindAddress(vout.Address)
// 			if !ok {
// 				continue
// 			}

// 			// fmt.Println("放入内存的txid为", base64.StdEncoding.EncodeToString(*txItr.GetHash()))

// 			switch txItr.Class() {
// 			case config.Wallet_tx_type_mining:
// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  []*crypto.AddressCoin{&vout.Address}, //输入地址
// 					OutAddr: nil,                                  //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				this.chain.history.Add(hi)
// 			case config.Wallet_tx_type_deposit_in:

// 			case config.Wallet_tx_type_deposit_out:
// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  vinAddrs,                             //输入地址
// 					OutAddr: []*crypto.AddressCoin{&vout.Address}, //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				this.chain.history.Add(hi)
// 			case config.Wallet_tx_type_pay:
// 				//判断是否是找零地址，判断依据是输入地址是否有自己钱包的地址
// 				have := false
// 				for _, one := range vinAddrs {
// 					// if keystore.FindAddress(*one) {
// 					// 	have = true
// 					// 	break
// 					// }
// 					_, ok := keystore.FindAddress(*one)
// 					if ok {
// 						have = true
// 						break
// 					}
// 				}

// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  []*crypto.AddressCoin{&vout.Address}, //
// 					OutAddr: vinAddrs,                             //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				if !have {
// 					hi.InAddr = []*crypto.AddressCoin{&vout.Address}
// 					this.chain.history.Add(hi)
// 				}

// 			case config.Wallet_tx_type_vote_in:

// 			case config.Wallet_tx_type_vote_out:
// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  []*crypto.AddressCoin{&vout.Address}, //输入地址
// 					OutAddr: vinAddrs,                             //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				this.chain.history.Add(hi)
// 			}

// 		}

// 	}
// }

/*
缴纳押金，并广播
*/
func (this *BalanceManager) DepositIn(amount, gas uint64, pwd, payload string, rate uint16) error {
	// addrInfo := keystore.GetCoinbase()

	//不能重复提交押金
	if this.nodeWitness != nil {
		return errors.New("Deposit cannot be paid repeatedly")
	}
	// if this.txManager.FindDeposit(addrInfo.Puk) {
	// 	return errors.New("Deposit cannot be paid repeatedly")
	// }
	if amount != config.Mining_deposit {
		return errors.New("Deposit not less than" + strconv.Itoa(int(uint64(config.Mining_deposit)/Unit)))
	}

	deposiIn, err := CreateTxDepositIn(amount, gas, pwd, payload, rate)
	if err != nil {
		return err
	}
	if deposiIn == nil {
		return errors.New("Failure to pay deposit")
	}
	deposiIn.BuildHash()
	MulticastTx(deposiIn)

	this.txManager.AddTx(deposiIn)
	return nil
}

/*
退还押金，并广播
*/
func (this *BalanceManager) DepositOut(addr string, amount, gas uint64, pwd string) error {

	if this.nodeWitness == nil {
		return errors.New("I didn't pay the deposit")
	}

	deposiOut, err := CreateTxDepositOut(addr, amount, gas, pwd)
	if err != nil {
		return err
	}
	if deposiOut == nil {
		return errors.New("Failure to pay deposit")
	}
	deposiOut.BuildHash()

	MulticastTx(deposiOut)

	this.txManager.AddTx(deposiOut)
	return nil
}

/*
投票押金，并广播
不能自己给自己投票，统计票的时候会造成循环引用
给见证人投票的都是社区节点，一个社区节点只能给一个见证人投票。
给社区节点投票的都是轻节点，轻节点投票前先缴押金。
轻节点可以给轻节点投票，相当于一个轻节点尾随另一个轻节点投票。
引用关系不能出现死循环
@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
*/
func (this *BalanceManager) VoteInOld(voteType, rate uint16, witnessAddr crypto.AddressCoin, addr crypto.AddressCoin, amount, gas uint64, pwd, payload string) error {

	//不能自己给自己投票
	if bytes.Equal(witnessAddr, addr) {
		return errors.New("You can't vote for yourself")
	}
	dstAddr := addr

	isWitness := this.witnessBackup.haveWitness(&dstAddr)
	_, isCommunity := this.witnessBackup.haveCommunityList(&dstAddr)
	_, isLight := this.witnessBackup.haveLight(&dstAddr)
	// fmt.Println("查看自己的角色", addr, isWitness, isCommunity, isLight)
	switch voteType {
	case 1: //1=给见证人投票
		if isLight || isWitness {
			return errors.New("The voting address is already another role")
		}
		vs, ok := this.witnessBackup.haveCommunityList(&dstAddr)
		if ok {
			if bytes.Equal(*vs.Witness, witnessAddr) {
				return errors.New("Can't vote again")
			}
			return errors.New("Cannot vote for multiple witnesses")
		}
		//检查押金
		if amount != config.Mining_vote {
			return errors.New("Community node deposit is " + strconv.Itoa(int(config.Mining_vote/1e8)))
		}
		//检查该社区节点是否已经有投票
		hasVote := uint64(0)
		voteList, ok := this.witnessBackup.haveVote(&addr)
		if ok {
			for _, one := range *voteList {
				hasVote += one.Scores
			}
		}
	case 2: //2=给社区节点投票

		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		//检查是否成为轻节点
		if !isLight {
			//先成为轻节点
			return errors.New("Become a light node first")
		}

		vs, ok := this.witnessBackup.haveVoteList(&dstAddr)
		if ok {
			if !bytes.Equal(*vs.Witness, witnessAddr) {
				//不能给多个社区节点投票
				return errors.New("Cannot vote for multiple community nodes")
			}
		}
	case 3: //3=轻节点押金
		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		if isLight {
			//已经是轻节点了
			return errors.New("It's already a light node")
		}
		// engine.Log.Info("轻节点押金是 %d %d", amount, config.Mining_light_min)

		if amount != config.Mining_light_min {
			//轻节点押金是
			return errors.New("Light node deposit is " + strconv.Itoa(int(config.Mining_light_min/1e8)))
		}
		witnessAddr = nil
	default:
		//不能识别的投票类型
		return errors.New("Unrecognized voting type")

	}

	voetIn, err := CreateTxVoteIn(voteType, rate, witnessAddr, addr, amount, gas, pwd, payload)
	if err != nil {
		return err
	}
	if voetIn == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	voetIn.BuildHash()
	// bs, err := voetIn.Json()
	// if err != nil {
	// 	//		fmt.Println("33333333333333 33333")
	// 	return err
	// }
	//	fmt.Println("4444444444444444")
	//	fmt.Println("5555555555555555")
	// txbase, err := ParseTxBase(bs)
	// if err != nil {
	// 	return err
	// }
	// voetIn.BuildHash()
	//	fmt.Println("66666666666666")
	//验证交易
	// if err := voetIn.CheckLockHeight(GetHighestBlock()); err != nil {
	// 	return err
	// }
	// if err := voetIn.Check(); err != nil {
	// 	//交易不合法，则不发送出去
	// 	// fmt.Println("交易不合法，则不发送出去")
	// 	return err
	// }
	MulticastTx(voetIn)
	this.txManager.AddTx(voetIn)
	// fmt.Println("添加投票押金是否成功", ok)
	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	//	fmt.Println("7777777777777777")
	return nil
}
func (this *BalanceManager) VoteIn(voteType, rate uint16, witnessAddr, addr crypto.AddressCoin, amount, gas uint64, pwd, comment string) error {
	//不能自己给自己投票
	if bytes.Equal(witnessAddr, addr) {
		return errors.New("You can't vote for yourself")
	}

	//比例必须小于100
	if rate >= 100 {
		return errors.New("The voting rate setting error")
	}

	addrType := GetAddrState(addr)
	voteToType := GetAddrState(witnessAddr)
	switch voteType {
	case VOTE_TYPE_community: //1=给见证人投票
		if addrType != 4 {
			return errors.New("The voting address is already another role")
		}

		if voteToType != 1 {
			//被投票地址是见证人角色
			return errors.New("the voted address must be is witness")
		}

		//检查押金
		if amount != config.Mining_vote {
			return errors.New("Community node deposit is " + strconv.Itoa(int(config.Mining_vote/1e8)))
		}

		//检查社区节点数量
		var currCommunityCount = 0
		this.depositCommunity.Range(func(key, value any) bool {
			currCommunityCount++
			return true
		})
		if currCommunityCount >= config.Max_Community_Count {
			return errors.New("community node limit overflow")
		}

	case VOTE_TYPE_vote: //2=给社区节点投票
		if addrType == 1 || addrType == 2 {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		//检查是否成为轻节点
		if addrType != 3 {
			//先成为轻节点
			return errors.New("Become a light node first")
		}

		if voteToType != 2 {
			//目标必须是社区节点
			return errors.New("the voted address must be community nodes")
		}

		depositInfo := this.GetDepositVote(&addr)
		if depositInfo != nil && !bytes.Equal(depositInfo.WitnessAddr, witnessAddr) {
			//不能给多个社区节点投票
			return errors.New("Cannot vote for multiple community nodes")
		}

		if depositInfo == nil {
			//检查轻节点数量
			lights, ok := this.communityMapLights.Load(utils.Bytes2string(witnessAddr))
			if ok {
				if len(lights.([]crypto.AddressCoin)) >= config.Max_Light_Count {
					return errors.New("light node limit overflow")
				}
			}
		}

	case VOTE_TYPE_light: //3=轻节点押金
		if addrType == 1 || addrType == 2 {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}

		if addrType == 3 {
			//已经是轻节点了
			return errors.New("It's already a light node")
		}

		if amount != config.Mining_light_min {
			//轻节点押金是
			return errors.New("Light node deposit is " + strconv.Itoa(int(config.Mining_light_min/1e8)))
		}
		witnessAddr = nil

	default:
		//不能识别的投票类型
		return errors.New("Unrecognized voting type")
	}

	voetIn, err := CreateTxVoteIn(voteType, rate, witnessAddr, addr, amount, gas, pwd, comment)
	if err != nil {
		return err
	}
	if voetIn == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	voetIn.BuildHash()
	MulticastTx(voetIn)
	this.txManager.AddTx(voetIn)
	return nil
}

// /*
// 	退还一笔投票押金，并广播
// */
// func (this *BalanceManager) VoteOutOne(txid, addr string, amount, gas uint64, pwd string) error {
// 	tx := this.GetVoteInByTxid(txid)
// 	if tx == nil {
// 		return errors.New("没有找到这个交易")
// 	}
// 	deposiOut, err := CreateTxVoteOutOne(tx, addr, amount, gas, pwd)
// 	if err != nil {
// 		return err
// 	}
// 	if deposiOut == nil {
// 		//		fmt.Println("33333333333333 22222")
// 		return errors.New("退押金失败")
// 	}
// 	deposiOut.BuildHash()
// 	bs, err := deposiOut.Json()
// 	if err != nil {
// 		//		fmt.Println("33333333333333 33333")
// 		return err
// 	}
// 	//	fmt.Println("4444444444444444")
// 	MulticastTx(bs)
// 	//	fmt.Println("5555555555555555")
// 	txbase, err := ParseTxBase(bs)
// 	if err != nil {
// 		return err
// 	}
// 	txbase.BuildHash()
// 	//	fmt.Println("66666666666666")
// 	//验证交易
// 	if !txbase.Check() {
// 		//交易不合法，则不发送出去
// 		// fmt.Println("交易不合法，则不发送出去")
// 		return errors.New("交易不合法，则不发送出去")
// 	}
// 	this.txManager.AddTx(txbase)
// 	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
// 	//	fmt.Println("7777777777777777")
// 	return nil
// 	return nil

// }

/*
退还投票押金，并广播
@voteType        uint16               取消类型
@witnessAddr     crypto.AddressCoin   见证人/社区节点地址
@addr            crypto.AddressCoin   取消的节点地址
@amount          uint64               取消投票额度
@gas             uint64               交易手续费
@pwd             string               支付密码
@payload         string               备注
*/
func (this *BalanceManager) VoteOut(voteType uint16, addr crypto.AddressCoin, amount, gas uint64, pwd, payload string) error {

	// waddr := witnessAddr

	// if witnessAddr != nil && witnessAddr.B58String() != "" {
	// 	waddr = witnessAddr
	// }
	// engine.Log.Info("---------------------------查询这个见证人" + waddr + "end")
	// balance := this.GetVoteIn(*waddr)
	// if balance == nil {
	// 	//没有对这个见证人投票
	// 	return errors.New("No vote for this witness")
	// }

	deposiOut, err := CreateTxVoteOut(voteType, addr, amount, gas, pwd, payload)
	if err != nil {
		return err
	}
	if deposiOut == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	deposiOut.BuildHash()
	MulticastTx(deposiOut)
	this.txManager.AddTx(deposiOut)
	return nil
}

/*
构建一个其他交易，并广播
*/
func (this *BalanceManager) BuildOtherTx(class uint64, srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {

	ctrl := GetTransactionCtrl(class)
	txItr, err := ctrl.BuildTx(this.otherDeposit, srcAddr, addr, amount, gas, frozenHeight, pwd, comment, params...)
	if err != nil {
		return nil, err
	}
	txItr.BuildHash()
	MulticastTx(txItr)

	if err := this.txManager.AddTx(txItr); err != nil {
		GetLongChain().Balance.DelLockTx(txItr)
		return nil, errors.Wrap(err, "add other tx fail!")
	}
	return txItr, nil
}

/*
	获得自己轻节点押金列表
*/
// func (this *BalanceManager) GetVoteList() []*Balance {
// 	balances := make([]*Balance, 0)
// 	this.votein.Range(func(k, v interface{}) bool {
// 		b := v.(*Balance)
// 		balances = append(balances, b)
// 		return true
// 	})
// 	return balances
// }

/*
统计NFT
@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countNFT(bhvo *BlockHeadVO) {
	// start := config.TimeNow()
	wg := new(sync.WaitGroup)

	NumCPUTokenChan := make(chan bool, runtime.NumCPU()*6)
	for _, txItr := range bhvo.Txs {
		if txItr.Class() != config.Wallet_tx_type_nft {
			continue
		}
		wg.Add(1)
		NumCPUTokenChan <- false
		go this.countNFTOne(txItr, NumCPUTokenChan, wg)
	}
	wg.Wait()
	// engine.Log.Info("count block nft time:%s", config.TimeNow().Sub(start))
}

/*
统计免gas质押金
*/
func (this *BalanceManager) countDepositFreeGas(bhvo *BlockHeadVO) {
	for _, txItr := range bhvo.Txs {
		class := txItr.Class()
		if class == config.Wallet_tx_type_deposit_free_gas {
			tx := txItr.(*Tx_DepositFreeGas)
			item := &DepositFreeGasItem{
				ContractAddresses: precompiled.RewardContract,
				Owner:             *tx.Vin[0].GetPukToAddr(),
				Deposit:           tx.Deposit,
				LimitHeight:       bhvo.BH.Height + config.DepositFreeGasLimitHeight,
				LimitCount:        config.DepositFreeGasLimitCount,
			}
			this.freeGasAddrSet.Store(utils.Bytes2string(tx.DepositAddress), item)
		}
		if class == config.Wallet_tx_type_pay || class == config.Wallet_tx_type_multsign_pay || class == config.Wallet_tx_type_address_transfer {
			vins := *txItr.GetVin()
			src := vins[0].GetPukToAddr()
			if value, ok := GetLongChain().Balance.freeGasAddrSet.Load(utils.Bytes2string(*src)); ok {
				item := value.(*DepositFreeGasItem)
				if item.LimitCount > 0 {
					item.LimitCount--
				}
			}
		}
	}
}

/*
列表质押免gas地址项
*/
func (this *BalanceManager) ListFreeGasAddrSet() []*DepositFreeGasItem {
	items := []*DepositFreeGasItem{}
	this.freeGasAddrSet.Range(func(key, value any) bool {
		item := value.(*DepositFreeGasItem)
		items = append(items, &DepositFreeGasItem{
			ContractAddresses: item.ContractAddresses,
			Owner:             item.Owner,
			Deposit:           item.Deposit,
			LimitHeight:       item.LimitHeight,
			LimitCount:        item.LimitCount,
		})
		return true
	})
	return items
}

/*
统计NFT交易
@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countNFTOne(txItr TxItr, tokenCPU chan bool, wg *sync.WaitGroup) {
	// engine.Log.Info("统计NFT交易")
	txItr.BuildHash()
	nftTx := txItr.(*Tx_nft)
	if nftTx.NFT_ID == nil || len(nftTx.NFT_ID) == 0 {
		//铸造
		err := AddAddrNFT(&nftTx.NFT_Owner, *nftTx.GetHash())
		if err != nil {
			engine.Log.Error("AddAddrNFT error:%s", err.Error())
		}
		err = AddNFTOwner(*nftTx.GetHash(), &nftTx.NFT_Owner)
		if err != nil {
			engine.Log.Error("AddNFTOwner error:%s", err.Error())
		}
	} else {
		//转让
		//查询之前的拥有者
		oldOwner, err := FindNFTOwner(nftTx.NFT_ID)
		if err != nil {
			engine.Log.Error("FindNFTOwner error:%s", err.Error())
		}
		//给地址删除一个NFT
		err = RemoveAddrNFT(oldOwner, nftTx.NFT_ID)
		if err != nil {
			engine.Log.Error("RemoveAddrNFT error:%s", err.Error())
		}
		//给地址添加一个NFT
		err = AddAddrNFT(&nftTx.NFT_Owner, nftTx.NFT_ID)
		if err != nil {
			engine.Log.Error("AddAddrNFT error:%s", err.Error())
		}
		//添加新的拥有者
		err = AddNFTOwner(nftTx.NFT_ID, &nftTx.NFT_Owner)
		if err != nil {
			engine.Log.Error("AddNFTOwner error:%s", err.Error())
		}
	}
	// itemCount := txItr.CountTxItemsNew(height)
	// itemChan <- itemCount
	<-tokenCPU
	wg.Done()
}

// 统计合约
func (this *BalanceManager) CountContract(bhvo *BlockHeadVO) {
	//start := config.TimeNow()
	var contractCount uint64 = 0
	//todo 待优化，目前串行执行，优化成并行执行

	block := environment.Block{
		Coinbase:  bhvo.BH.Witness,
		Timestamp: evmutils.New(bhvo.BH.Time),
		Number:    evmutils.New(int64(bhvo.BH.Height)),
		Hash:      evmutils.BytesDataToEVMIntHash(bhvo.BH.Hash),
	}

	vmRun := evm.NewCountVmRun(&block)
	tmpCacheContractAddrs := []crypto.AddressCoin{
		precompiled.RewardContract,
	}
	this.tmpCacheContract.Range(func(key, value any) bool {
		addr := crypto.AddressCoin(key.(string))
		tmpCacheContractAddrs = append(tmpCacheContractAddrs, addr)
		return true
	})
	vmRun.SetStorage(tmpCacheContractAddrs...)

	//记录区块所有交易的所有事件记录
	bhEvents := make([]*go_protos.ContractEventInfo, 0)
	for _, txItr := range bhvo.Txs {
		//普通交易转换为合约交易,跳过创世块
		//if bhvo.BH.Height > 1 {
		//	txItr = this.normalTxConvertContractTx(txItr)
		//}

		//if txItr.Class() == config.Wallet_tx_type_mining && len(txItr.GetPayload()) > 0 {
		//	vmRun.SetReward(true)
		//	contractVin, contractVout := (*txItr.GetVin())[0], (*txItr.GetVout())[0]
		//	vmRun.SetTxContext(*contractVin.GetPukToAddr(), contractVout.Address, *txItr.GetHash(), nil, nil)

		//	//更新缓存余额
		//	vmRun.UpdateCacheObjBalance(precompiled.RewardContract, contractVout.Value)
		//	//统计奖励合约
		//	events := this.countContractRewardV1(vmRun, txItr, bhvo)
		//	if len(events) > 0 {
		//		bhEvents = append(bhEvents, events...)
		//	}
		//}
		if txItr.Class() != config.Wallet_tx_type_contract {
			continue
		}
		vmRun.SetReward(false)
		contractVin, contractVout := (*txItr.GetVin())[0], (*txItr.GetVout())[0]
		vmRun.SetTxContext(*contractVin.GetPukToAddr(), contractVout.Address, *txItr.GetHash(), bhvo.BH.Time, bhvo.BH.Height, evmutils.New(0), evmutils.New(100000))
		events := this.countContractOneV1(vmRun, txItr, bhvo, &contractCount)
		if len(events) > 0 {
			bhEvents = append(bhEvents, events...)
		}
		this.countRewardContract(txItr)
	}
	err := db.IncContractCount(big.NewInt(int64(contractCount)))
	if err != nil {
		engine.Log.Error("更新合约数量失败:%s", err.Error())
	}
	//engine.Log.Info("count block 合约 time:%s", config.TimeNow().Sub(start))

	// 合约执行成功,缓存全局合约对象
	vmRun.CacheStorage(precompiled.RewardContract)

	//计算区块bloom
	bloom := CreateBloom(bhEvents)
	if err := SaveBlockHeadBloom(bhvo.BH.Height, bloom.Bytes()); err != nil {
		engine.Log.Error("保存区块bloom失败：%s", err.Error())
	}
}

func (this *BalanceManager) countContractOneV1(vmRun *evm.VmRun, txItr TxItr, bhvo *BlockHeadVO, contractCount *uint64) []*go_protos.ContractEventInfo {

	// engine.Log.Info("统计合约交易")
	txItr.BuildHash()
	payTx := txItr.(*Tx_Contract)
	contractVin, contractVout := payTx.GetContractInfo()

	var (
		evmErr         error
		result         vm.ExecuteResult
		eventsInfo     []*go_protos.ContractEventInfo
		contractEvents []*go_protos.ContractEvent
	)
	if payTx.Action == "create" {
		//创建合约，提前构建相应结构
		result, contractEvents, evmErr = evm.Run(txItr.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, true, vmRun)

		if evmErr != nil {
			engine.Log.Error("合约创建失败,错误信息：%s,退出码%s", evmErr.Error(), result.ExitOpCode.String())
		} else {
			engine.Log.Info("合约创建成功，退出码：%s", result.ExitOpCode.String())
			atomic.AddUint64(contractCount, 1)
			if len(payTx.GzipSource) > 0 {
				err := SetContractSource(payTx.GzipSource, contractVout.Address)
				if err != nil {
					engine.Log.Info("合约源代码保存失败:%s", err.Error())
				}
			}
			if len(payTx.GzipAbi) > 0 {
				err := SetContractAbi(payTx.GzipAbi, contractVout.Address)
				if err != nil {
					engine.Log.Info("合约源abi保存失败:%s", err.Error())
				}
			}
			if len(payTx.TxBase.Payload) > 0 {
				err := SetContractBin(payTx.TxBase.Payload, contractVout.Address)
				if err != nil {
					engine.Log.Info("合约源bin保存失败:%s", err.Error())
				}
			}
			//记录特定类型合约部署
			if payTx.ContractClass > 0 {
				err := SetSpecialContract(*contractVin.GetPukToAddr(), contractVout.Address, payTx.ContractClass)
				if err != nil {
					engine.Log.Info("特定类型合约记录失败:%s", err.Error())
				}
			}

			erc, isErc20 := precompiled.CheckErc20(vmRun, *contractVin.GetPukToAddr(), contractVout.Address)
			if isErc20 {
				//存储Erc20
				db.SaveErc20Info(erc)
			}

			erc721, isErc721 := precompiled.CheckErc721(vmRun, *contractVin.GetPukToAddr(), contractVout.Address)
			if isErc721 {
				//存储Erc721
				db.SaveErc721Info(erc721)
			}

			erc1155, isErc1155 := precompiled.CheckErc1155(vmRun, *contractVin.GetPukToAddr(), contractVout.Address)
			if isErc1155 {
				//存储Erc1155
				db.SaveErc1155Info(erc1155)
			}
		}
		//result, _ = evm.Run(from, to, *payTx.GetHash(), payTx.GetPayload(), payTx.GetGas(), contractVout.Value, &block, true)
		//todo 判断合约是20 还是721 还是1155
	} else {
		//engine.Log.Error("统计调用合约 hash:%s", hex.EncodeToString(*txItr.GetHash()))
		//调用合约
		//特殊处理云储存合约
		if payTx.ContractClass == uint64(config.CLOUD_STORAGE_PROXY_CONTRACT) {
			result, contractEvents, evmErr = evm.Run(txItr.GetPayload(), config.Mining_coin_total, payTx.GasPrice, 1000000*1e8, false, vmRun)
		} else {
			result, contractEvents, evmErr = evm.Run(txItr.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, false, vmRun)
		}

		//result, _ = evm.Run(from, to, *payTx.GetHash(), payTx.GetPayload(), payTx.GetGas(), contractVout.Value, &block, false)
		//if evmErr != nil {
		//	engine.Log.Error("合约调用失败,错误信息：%s,退出码%s", evmErr.Error(), result.ExitOpCode.String())
		//} else {
		//	//engine.Log.Info("合约调用成功，退出码：%s", result.ExitOpCode.String())
		//}
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				msg, err := abi.UnpackRevert(result.ResultData)
				if err != nil {
					engine.Log.Error("合约调用失败,错误信息：%s,退出码%s", err.Error(), result.ExitOpCode.String())
				}
				engine.Log.Error("合约调用失败,退出码:%s,解析出的失败原因:%s", result.ExitOpCode.String(), msg)
			}
			if evmErr != nil {
				engine.Log.Error("合约调用失败，退出码:%s,失败原因:%s", result.ExitOpCode.String(), evmErr.Error())
			}
			engine.Log.Error("执行合约调用失败，退出码:%s", result.ExitOpCode.String())
		}

		//todo 获取用户代币余额
	}
	//notice:以太坊是全部处理后，一次发送区块所有的日志
	//推送日志消息，即合约事件
	//logsMap := result.StorageCache.Logs
	//contractEvents, _ := evm.EmitContractEvent(logsMap, string(*payTx.GetHash()), to.B58String())
	for _, v := range contractEvents {
		eventInfo := &go_protos.ContractEventInfo{
			BlockHeight:     int64(bhvo.BH.Height),
			Topic:           v.Topic,
			TxId:            v.TxId,
			ContractAddress: v.ContractAddress,
			EventData:       v.EventData,
		}
		eventsInfo = append(eventsInfo, eventInfo)
	}
	//engine.Log.Info("发送的事件长度%d", len(eventsInfo))
	//保存单笔交易的bloom
	if len(eventsInfo) > 0 {
		bloom := CreateBloom(eventsInfo)
		//payTx.Bloom = bloom.Bytes()
		txItr.SetBloom(bloom.Bytes())
		this.eventBus.Publish(event.ContractEventTopic, &go_protos.ContractEventInfoList{ContractEvents: eventsInfo})
	}
	return eventsInfo
}
func (this *BalanceManager) countContractOne(txItr TxItr, bhvo *BlockHeadVO, contractCount *uint64) {
	// engine.Log.Info("统计合约交易")
	txItr.BuildHash()
	payTx := txItr.(*Tx_Contract)
	contractVin, contractVout := payTx.GetContractInfo()
	from := *contractVin.GetPukToAddr()
	//fmt.Println("交易发起方：", from.B58String())
	to := contractVout.Address
	//fmt.Println("调用方:", to.B58String())
	//这里生成block对象
	block := environment.Block{
		Coinbase:   bhvo.BH.Witness,
		Timestamp:  evmutils.New(bhvo.BH.Time),
		Number:     evmutils.New(int64(bhvo.BH.Height)),
		Difficulty: evmutils.New(0),
		GasLimit:   evmutils.New(100000),
	}

	var (
		evmErr         error
		result         vm.ExecuteResult
		eventsInfo     []*go_protos.ContractEventInfo
		contractEvents []*go_protos.ContractEvent
	)
	vmRun := evm.NewVmRun(from, to, *payTx.GetHash(), &block)
	if payTx.Action == "create" {
		//创建合约，提前构建相应结构
		result, contractEvents, evmErr = vmRun.Run(payTx.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, true)
		if evmErr != nil {
			engine.Log.Error("合约创建失败,错误信息：%s,退出码%s", evmErr.Error(), result.ExitOpCode.String())
		} else {
			engine.Log.Info("合约创建成功，退出码：%s", result.ExitOpCode.String())
			atomic.AddUint64(contractCount, 1)
			if len(payTx.GzipSource) > 0 {
				err := SetContractSource(payTx.GzipSource, to)
				if err != nil {
					engine.Log.Info("合约源代码保存失败:%s", err.Error())
				}
			}
			if len(payTx.GzipAbi) > 0 {
				err := SetContractAbi(payTx.GzipAbi, to)
				if err != nil {
					engine.Log.Info("合约源abi保存失败:%s", err.Error())
				}
			}
			if len(payTx.TxBase.Payload) > 0 {
				err := SetContractBin(payTx.TxBase.Payload, to)
				if err != nil {
					engine.Log.Info("合约源bin保存失败:%s", err.Error())
				}
			}
			//记录特定类型合约部署
			if payTx.ContractClass > 0 {
				err := SetSpecialContract(from, to, payTx.ContractClass)
				if err != nil {
					engine.Log.Info("特定类型合约记录失败:%s", err.Error())
				}
			}

			//utils.Go(func() {
			//	erc, isErc20 := precompiled.CheckErc20(from, to)
			//	if isErc20 {
			//		//存储Erc20
			//		db.SaveErc20Info(erc)
			//	}
			//})
		}
		//result, _ = evm.Run(from, to, *payTx.GetHash(), payTx.GetPayload(), payTx.GetGas(), contractVout.Value, &block, true)
		//todo 判断合约是20 还是721 还是1155
	} else {
		//调用合约
		result, contractEvents, evmErr = vmRun.Run(payTx.GetPayload(), payTx.GetGasLimit(), payTx.GasPrice, contractVout.Value, false)
		//result, _ = evm.Run(from, to, *payTx.GetHash(), payTx.GetPayload(), payTx.GetGas(), contractVout.Value, &block, false)
		//if evmErr != nil {
		//	engine.Log.Error("合约调用失败,错误信息：%s,退出码%s", evmErr.Error(), result.ExitOpCode.String())
		//} else {
		//	//engine.Log.Info("合约调用成功，退出码：%s", result.ExitOpCode.String())
		//}
		if evmErr != nil || result.ExitOpCode == opcodes.REVERT {
			if result.ExitOpCode == opcodes.REVERT {
				msg, err := abi.UnpackRevert(result.ResultData)
				if err != nil {
					engine.Log.Error("合约调用失败,错误信息：%s,退出码%s", err.Error(), result.ExitOpCode.String())
				}
				engine.Log.Error("合约调用失败,退出码:"+result.ExitOpCode.String()+",解析出的失败原因:"+msg, "")
			}
			if evmErr != nil {
				engine.Log.Error("合约调用失败，退出码"+result.ExitOpCode.String()+",失败原因:"+evmErr.Error(), "")
			}
			engine.Log.Error("执行合约调用失败，退出码"+result.ExitOpCode.String(), "")
		}

		//todo 获取用户代币余额
	}
	//notice:以太坊是全部处理后，一次发送区块所有的日志
	//推送日志消息，即合约事件
	//logsMap := result.StorageCache.Logs
	//contractEvents, _ := evm.EmitContractEvent(logsMap, string(*payTx.GetHash()), to.B58String())
	for _, v := range contractEvents {
		eventInfo := &go_protos.ContractEventInfo{
			BlockHeight:     int64(bhvo.BH.Height),
			Topic:           v.Topic,
			TxId:            v.TxId,
			ContractAddress: v.ContractAddress,
			EventData:       v.EventData,
		}
		eventsInfo = append(eventsInfo, eventInfo)
	}
	//engine.Log.Info("发送的事件长度%d", len(eventsInfo))
	if len(eventsInfo) > 0 {
		this.eventBus.Publish(event.ContractEventTopic, &go_protos.ContractEventInfoList{ContractEvents: eventsInfo})
	}
}

func (this *BalanceManager) countContractRewardV1(vmRun *evm.VmRun, txItr TxItr, bhvo *BlockHeadVO) []*go_protos.ContractEventInfo {
	start := config.TimeNow()
	//engine.Log.Info("统计合约交易")
	txItr.BuildHash()

	var (
		evmErr         error
		result         vm.ExecuteResult
		eventsInfo     []*go_protos.ContractEventInfo
		contractEvents []*go_protos.ContractEvent
	)

	//调用合约
	result, contractEvents, evmErr = evm.Run(txItr.GetPayload(), config.Mining_coin_total, config.DEFAULT_GAS_PRICE, 0, false, vmRun)
	if evmErr == nil { // 记录奖励
		// 额外记录合约交易事件
		// 见证人出块奖励
		//CustomRewardTxEvent(*txItr.GetHash())
	}

	engine.Log.Info("统计奖励合约交易:%s", config.TimeNow().Sub(start))
	if evmErr != nil {
		engine.Log.Error("奖励合约调用失败,错误信息：%s,退出码%s", evmErr.Error(), result.ExitOpCode.String())
	}

	//处理奖励合约事件
	for _, v := range contractEvents {
		eventInfo := &go_protos.ContractEventInfo{
			BlockHeight:     int64(bhvo.BH.Height),
			Topic:           v.Topic,
			TxId:            v.TxId,
			ContractAddress: v.ContractAddress,
			EventData:       v.EventData,
		}
		eventsInfo = append(eventsInfo, eventInfo)
	}
	//保存单笔交易的bloom
	if len(eventsInfo) > 0 {
		bloom := CreateBloom(eventsInfo)
		//txItr.(*Tx_reward).Bloom = bloom.Bytes()
		txItr.SetBloom(bloom.Bytes())
		this.eventBus.Publish(event.ContractEventTopic, &go_protos.ContractEventInfoList{ContractEvents: eventsInfo})
	}
	return eventsInfo
}

func (this *BalanceManager) countContractReward(txItr TxItr, bhvo *BlockHeadVO) {
	start := config.TimeNow()
	//engine.Log.Info("统计合约交易")
	txItr.BuildHash()
	payTx := txItr.(*Tx_reward)
	contractVin, contractVout := payTx.Vin[0], payTx.Vout[0]
	from := *contractVin.GetPukToAddr()
	//fmt.Println("交易发起方：", from.B58String())
	to := contractVout.Address
	//fmt.Println("调用方:", to.B58String())
	block := environment.Block{
		Coinbase:  bhvo.BH.Witness,
		Timestamp: evmutils.New(bhvo.BH.Time),
		Number:    evmutils.New(int64(bhvo.BH.Height)),
	}

	var (
		evmErr error
		result vm.ExecuteResult
		//contractEvents []*go_protos.ContractEvent
	)
	vmRun := evm.NewRewardVmRun(from, to, *payTx.GetHash(), &block)
	//调用合约
	result, _, evmErr = vmRun.Run(payTx.GetPayload(), config.Mining_coin_total, config.DEFAULT_GAS_PRICE, 0, false)
	//result, _ = evm.Run(from, to, *payTx.GetHash(), payTx.GetPayload(), payTx.GetGas(), contractVout.Value, &block, false)

	engine.Log.Info("统计奖励合约交易:%s", config.TimeNow().Sub(start))
	if evmErr != nil {
		engine.Log.Error("奖励合约调用失败,错误信息：%s,退出码%s", evmErr.Error(), result.ExitOpCode.String())
	}
}

func (this *BalanceManager) VoteInContract(voteType, rate uint16, voteTo crypto.AddressCoin, voter crypto.AddressCoin, amount, gas uint64, pwd, name string, gasPrice uint64) error {

	payload := []byte{}
	//不能自己给自己投票
	if bytes.Equal(voteTo, voter) {
		return errors.New("You can't vote for yourself")
	}
	isWitness := this.witnessBackup.haveWitness(&voter)
	switch voteType {
	case VOTE_TYPE_community: //1=给见证人投票
		//是否是轻节点验证通过合约内部验证
		if isWitness {
			return errors.New("The voting address is already another role")
		}
		if !this.witnessBackup.haveWitness(&voteTo) {
			//被投票地址是见证人角色
			return errors.New("the voted address must be is witness")
		}
		//检查押金
		if amount != config.Mining_vote {
			return errors.New("Community node deposit is " + strconv.Itoa(int(config.Mining_vote/1e8)))
		}
		//构造合约payload
		payload = precompiled.BuildAddCommunityInput(voteTo, rate, name)
	case VOTE_TYPE_vote: //2=给社区节点投票
		if isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		if this.witnessBackup.haveWitness(&voteTo) {
			//被投票地址是见证人角色
			return errors.New("the voted address is witness")
		}
		//构建合约payload
		payload = precompiled.BuildAddVoteInput(voteTo)
	case VOTE_TYPE_light: //3=轻节点押金
		if isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}

		if amount != config.Mining_light_min {
			//轻节点押金是
			return errors.New("Light node deposit is " + strconv.Itoa(int(config.Mining_light_min/1e8)))
		}
		voteTo = nil
		//构造合约payload
		payload = precompiled.BuildAddLightInput(name)
	default:
		//不能识别的投票类型
		return errors.New("Unrecognized voting type")

	}
	//创建合约交易
	unlock := createTxMutex.Lock(voter.B58String())
	tx, err := CreateTxContractNew(&voter, &precompiled.RewardContract, amount, gas, 0, pwd, common.Bytes2Hex(payload), "", 0, gasPrice, "")
	unlock()
	if err != nil {
		return err
	}
	tx.BuildHash()

	MulticastTx(tx)
	this.txManager.AddTx(tx)

	return nil
}

/*
退还投票押金，并广播
@voteType        uint16               取消类型
@witnessAddr     crypto.AddressCoin   见证人/社区节点地址
@addr            crypto.AddressCoin   取消的节点地址
@amount          uint64               取消投票额度
@gas             uint64               交易手续费
@pwd             string               支付密码
@payload         string               备注
*/
func (this *BalanceManager) VoteOutContract(voteType uint16, addr crypto.AddressCoin, amount, gas uint64, pwd string, gasPrice uint64) error {
	payload := []byte{}
	switch voteType {
	case VOTE_TYPE_community:
		//构造社区节点退出输入
		payload = precompiled.BuildDelCommunity()
	case VOTE_TYPE_vote:
		payload = precompiled.BuildDelVote(big.NewInt(int64(amount)))
	case VOTE_TYPE_light:
		payload = precompiled.BuildDelLight()
	}

	//创建合约交易
	unlock := createTxMutex.Lock(addr.B58String())
	tx, err := CreateTxContractNew(&addr, &precompiled.RewardContract, 0, gas, 0, pwd, common.Bytes2Hex(payload), "", 0, gasPrice, "")
	unlock()
	if err != nil {
		return err
	}
	tx.BuildHash()

	MulticastTx(tx)
	this.txManager.AddTx(tx)

	return nil
}

/*
*
统计社区节点和轻节点质押
*/
//func (this *BalanceManager) countCommunityOrLightDeposit(txItr TxItr, bhvo *BlockHeadVO) {
//	tx := txItr.(*Tx_Contract)
//	vin, vout := tx.GetContractInfo()
//	if !bytes.Equal(vout.Address, precompiled.RewardContract) {
//		return
//	}
//
//	//解析payload
//	voteType, dstAddr, isAdd := precompiled.UnpackPayloadV1(tx.GetPayload())
//	if voteType == 0 {
//		return
//	}
//
//	switch voteType {
//	case VOTE_TYPE_community: //社区节点
//		if !isAdd {
//			if itemItr, ok := this.depositCommunity.Load(utils.Bytes2string(*vin.GetPukToAddr())); ok {
//				item := itemItr.(*DepositInfo)
//				this.removeWitnessMapCommunitys(item.WitnessAddr, *vin.GetPukToAddr())
//
//				// 减去见证人票数
//				witKey := utils.Bytes2string(item.WitnessAddr)
//				if witVote, ok := this.witnessVote.Load(utils.Bytes2string(item.WitnessAddr)); ok {
//					if commVote, ok := this.communityVote.Load(utils.Bytes2string(item.SelfAddr)); ok {
//						this.witnessVote.Store(witKey, witVote.(uint64)-commVote.(uint64))
//					}
//				}
//			}
//
//			//清除押金记录
//			RemoveDepositCommunityAddr(vin.GetPukToAddr())
//			//清除投票及关系
//			this.depositCommunity.Delete(utils.Bytes2string(*vin.GetPukToAddr()))
//
//			return
//		}
//
//		//记录押金
//		SetDepositCommunityAddr(vin.GetPukToAddr(), vout.Value)
//		txItem := DepositInfo{
//			WitnessAddr: dstAddr,             //见证人/社区节点地址
//			SelfAddr:    *vin.GetPukToAddr(), //轻节点/社区节点地址
//			Value:       vout.Value,          //押金或投票金额
//		}
//
//		//记录社区质押比例
//		_, commRate := precompiled.UnpackPayloadCommunityRate(tx.GetPayload())
//		this.witnessRatio.Store(utils.Bytes2string(*vin.GetPukToAddr()), commRate)
//		//记录投票及关系
//		this.depositCommunity.Store(utils.Bytes2string(*vin.GetPukToAddr()), &txItem)
//		this.setWitnessMapCommunitys(dstAddr, *vin.GetPukToAddr())
//	case VOTE_TYPE_light: //轻节点
//		if !isAdd {
//			//清除押金记录
//			RemoveDepositLightAddr(vin.GetPukToAddr())
//			this.depositLight.Delete(utils.Bytes2string(*vin.GetPukToAddr()))
//			return
//		}
//
//		//记录押金
//		SetDepositLightAddr(vin.GetPukToAddr(), vout.Value)
//		txItem := DepositInfo{
//			WitnessAddr: dstAddr,             //见证人/社区节点地址
//			SelfAddr:    *vin.GetPukToAddr(), //轻节点/社区节点地址
//			Value:       vout.Value,          //押金或投票金额
//		}
//		this.depositLight.Store(utils.Bytes2string(*vin.GetPukToAddr()), &txItem)
//	case VOTE_TYPE_vote: //投票
//		lightAddr := utils.Bytes2string(*vin.GetPukToAddr())
//		itemItr, ok := this.depositVote.Load(lightAddr)
//		if !isAdd {
//			if ok {
//				//减少投票
//				item := itemItr.(*DepositInfo)
//				_, vote := precompiled.UnpackReward(tx.GetPayload())
//				item.Value -= vote.(uint64)
//				if item.Value == 0 {
//					this.depositVote.Delete(lightAddr)
//					this.removeCommunityMapLights(item.WitnessAddr, item.SelfAddr)
//					//RemoveVoteAddr(vin.GetPukToAddr())
//				}
//				// 更新见证人和社区的票数
//				//engine.Log.Error("减票更新票数：%s", item.WitnessAddr.B58String())
//				this.updateWitnessAndCommunityVote(item.WitnessAddr, vote.(uint64), false)
//			}
//			return
//		}
//
//		//添加投票
//		if ok {
//			item := itemItr.(*DepositInfo)
//			item.Value += vout.Value
//			// 更新见证人和社区的票数
//			//engine.Log.Error("加票更新票数111：%s", dstAddr.B58String())
//			this.updateWitnessAndCommunityVote(dstAddr, vout.Value, true)
//			return
//		}
//
//		txItem := DepositInfo{
//			WitnessAddr: dstAddr,             //见证人/社区节点地址
//			SelfAddr:    *vin.GetPukToAddr(), //轻节点/社区节点地址
//			Value:       vout.Value,          //押金或投票金额
//		}
//		this.depositVote.Store(lightAddr, &txItem)
//		//SetVoteAddr(vin.GetPukToAddr(), &dstAddr)
//		// 更新见证人和社区的票数
//		//engine.Log.Error("加票更新票数222：%s", dstAddr.B58String())
//		this.updateWitnessAndCommunityVote(dstAddr, vout.Value, true)
//		this.setCommunityMapLights(txItem.WitnessAddr, txItem.SelfAddr)
//	}
//}
//

// 记录见证人和社区的映射关系
func (this *BalanceManager) setWitnessMapCommunitys(witAddr, communityAddr crypto.AddressCoin) {
	if v, ok := this.witnessMapCommunitys.Load(utils.Bytes2string(witAddr)); ok {
		communityAddrs := v.([]crypto.AddressCoin)
		communityAddrs = append(communityAddrs, communityAddr)
		this.witnessMapCommunitys.Store(utils.Bytes2string(witAddr), communityAddrs)
	} else {
		communityAddrs := []crypto.AddressCoin{
			communityAddr,
		}
		this.witnessMapCommunitys.Store(utils.Bytes2string(witAddr), communityAddrs)
	}
}

// 移除见证人和社区的映射关系
func (this *BalanceManager) removeWitnessMapCommunitys(witAddr, communityAddr crypto.AddressCoin) {
	if v, ok := this.witnessMapCommunitys.Load(utils.Bytes2string(witAddr)); ok {
		communityAddrs := v.([]crypto.AddressCoin)
		newCommunityAddrs := []crypto.AddressCoin{}
		for i, addr := range communityAddrs {
			if bytes.Equal(addr, communityAddr) {
				continue
			}
			newCommunityAddrs = append(newCommunityAddrs, communityAddrs[i])
		}
		this.witnessMapCommunitys.Store(utils.Bytes2string(witAddr), newCommunityAddrs)
	}
}

// 记录社区和轻节点的映射关系
func (this *BalanceManager) setCommunityMapLights(communityAddr, lightAddr crypto.AddressCoin) {
	if v, ok := this.communityMapLights.Load(utils.Bytes2string(communityAddr)); ok {
		lightAddrs := v.([]crypto.AddressCoin)
		lightAddrs = append(lightAddrs, lightAddr)
		this.communityMapLights.Store(utils.Bytes2string(communityAddr), lightAddrs)
	} else {
		lightAddrs := []crypto.AddressCoin{
			lightAddr,
		}
		this.communityMapLights.Store(utils.Bytes2string(communityAddr), lightAddrs)
	}
}

// 移除社区和轻节点的映射关系
func (this *BalanceManager) removeCommunityMapLights(communityAddr, lightAddr crypto.AddressCoin) {
	if v, ok := this.communityMapLights.Load(utils.Bytes2string(communityAddr)); ok {
		lightAddrs := v.([]crypto.AddressCoin)
		newLightAddrs := []crypto.AddressCoin{}
		for i, addr := range lightAddrs {
			if bytes.Equal(addr, lightAddr) {
				continue
			}
			newLightAddrs = append(newLightAddrs, lightAddrs[i])
		}
		this.communityMapLights.Store(utils.Bytes2string(communityAddr), newLightAddrs)
	}
}

// 更新社区的票数
func (this *BalanceManager) updateCommunityVote(c crypto.AddressCoin, v uint64, isAdd bool) {
	// 记录社区票数
	voteNew := v
	commkey := utils.Bytes2string(c)
	if vote, ok := this.communityVote.Load(commkey); ok {
		if isAdd {
			voteNew = vote.(uint64) + v
		} else {
			voteNew = vote.(uint64) - v
		}
	}

	this.communityVote.Store(commkey, voteNew)
}

// 更新见证人票数
func (this *BalanceManager) updateWitnessVote(w crypto.AddressCoin, v uint64, isAdd bool) {
	// 记录见证人票数
	voteNew := v
	witkey := utils.Bytes2string(w)
	if vote, ok := this.witnessVote.Load(witkey); ok {
		if isAdd {
			voteNew = vote.(uint64) + v
		} else {
			voteNew = vote.(uint64) - v
		}
	}

	this.witnessVote.Store(witkey, voteNew)

	//更新备用见证人的票数
	if wbinfo, ok := this.witnessBackup.witnessesMap.Load(witkey); ok {
		wbinfo.(*BackupWitness).VoteNum = voteNew
	}
	for i, wit := range this.witnessBackup.witnesses {
		if bytes.Equal(w, *wit.Addr) {
			this.witnessBackup.witnesses[i].VoteNum = voteNew
			break
		}
	}
}

// 获取缓存中见证人/社区质押比例
func (this *BalanceManager) GetDepositRate(w *crypto.AddressCoin) uint16 {
	v, ok := this.witnessRatio.Load(utils.Bytes2string(*w))
	if !ok {
		return 0
	}

	return v.(uint16)

}

// 获取缓存中见证人投票
func (this *BalanceManager) GetWitnessVote(w *crypto.AddressCoin) uint64 {
	v, ok := this.witnessVote.Load(utils.Bytes2string(*w))
	if !ok {
		return 0
	}

	return v.(uint64)
}

// 获取缓存中社区投票
func (this *BalanceManager) GetCommunityVote(w *crypto.AddressCoin) uint64 {
	v, ok := this.communityVote.Load(utils.Bytes2string(*w))
	if !ok {
		return 0
	}

	return v.(uint64)
}

// 上次投票操作高度
func (this *BalanceManager) GetLastVoteOp(addr crypto.AddressCoin) uint64 {
	v, ok := this.lastVoteOp.Load(utils.Bytes2string(addr))
	if ok {
		return v.(uint64)
	}

	return 0
}

// 地址累计奖励
func (this *BalanceManager) GetAddrReward(addr crypto.AddressCoin) *big.Int {
	v, ok := this.addrReward.Load(utils.Bytes2string(addr))
	if ok {
		return v.(*big.Int)
	}

	return big.NewInt(0)
}

// 处理奖励合约相关调用
func (this *BalanceManager) countRewardContract(txItr TxItr) {
	tx := txItr.(*Tx_Contract)
	vin, vout := tx.GetContractInfo()
	if !bytes.Equal(vout.Address, precompiled.RewardContract) {
		return
	}

	tag, out := precompiled.UnpackReward(txItr.GetPayload())
	switch tag {
	case precompiled.SET_RATE: //设置比例
		this.witnessRatio.Store(utils.Bytes2string(*vin.GetPukToAddr()), uint16(out.(uint8)))
	}

}

/*
// NOTE: 普通交易转换为合约交易
func (this *BalanceManager) normalTxConvertContractTx(txItr TxItr) TxItr {
	if config.EVM_Reward_Enable {
		switch txItr.Class() {
		case config.Wallet_tx_type_mining:
			item := txItr.(*Tx_reward)
			vouts := *txItr.GetVout()
			index := item.Index
			rewardWitness := []*crypto.AddressCoin{}
			allReward := uint64(0)
			for i, _ := range vouts {
				allReward += vouts[i].Value
				rewardWitness = append(rewardWitness, &vouts[i].Address)
			}

			payload := precompiled.BuildDistributeInput(rewardWitness, index, allReward)
			txItr.SetPayload(payload)
			item.Vout[0].Address = precompiled.RewardContract
			return txItr
		case config.Wallet_tx_type_vote_in:
			item := txItr.(*Tx_vote_in)
			item.Vout[0].Address = precompiled.RewardContract
			newTxItr := &Tx_Contract{
				TxBase: item.TxBase,
				Action: "call",
			}
			switch item.VoteType {
			case VOTE_TYPE_community: //社区节点押金
				name := ""
				if len(txItr.GetPayload()) > 0 {
					name = string(txItr.GetPayload())
				}

				payload := precompiled.BuildAddCommunityInput(item.Vote, item.Rate, name)
				newTxItr.SetPayload(payload)
				newTxItr.SetClass(config.Wallet_tx_type_contract)
			case VOTE_TYPE_vote: //轻节点投票
				payload := precompiled.BuildAddVoteInput(item.Vote)
				newTxItr.SetPayload(payload)
				newTxItr.SetClass(config.Wallet_tx_type_contract)
			case VOTE_TYPE_light: //轻节点押金
				name := ""
				if len(txItr.GetPayload()) > 0 {
					name = string(txItr.GetPayload())
				}
				payload := precompiled.BuildAddLightInput(name)
				newTxItr.SetPayload(payload)
				newTxItr.SetClass(config.Wallet_tx_type_contract)
			}
			return newTxItr
		case config.Wallet_tx_type_vote_out:
			item := txItr.(*Tx_vote_out)
			item.Vout[0].Address = precompiled.RewardContract
			newTxItr := &Tx_Contract{
				TxBase: item.TxBase,
				Action: "call",
			}
			switch item.VoteType {
			case VOTE_TYPE_community: //社区节点押金
				payload := precompiled.BuildDelCommunity()
				newTxItr.SetPayload(payload)
				newTxItr.SetClass(config.Wallet_tx_type_contract)
			case VOTE_TYPE_vote: //轻节点投票
				voutOne := (*txItr.GetVout())[0]
				value := voutOne.Value //质押/投票金额
				payload := precompiled.BuildDelVote(big.NewInt(int64(value)))
				newTxItr.SetPayload(payload)
				newTxItr.SetClass(config.Wallet_tx_type_contract)
			case VOTE_TYPE_light: //轻节点押金
				payload := precompiled.BuildDelLight()
				newTxItr.SetPayload(payload)
				newTxItr.SetClass(config.Wallet_tx_type_contract)
			}
			return newTxItr
		}
	}
	return txItr
}
*/

// 更新备用见证人票数
//func (this *BalanceManager) updateWitnessBackupVote(addr crypto.AddressCoin) {
//	this.chain.Balance.witnessVote.Range(func(key, value any) bool {
//		if v, ok := this.witnessBackup.witnessesMap.Load(key); ok {
//			v.(*BackupWitness).VoteNum = value.(uint64)
//		}
//		return true
//	})
//
//	for i, wit := range this.witnessBackup.witnesses {
//		if v, ok := this.witnessBackup.witnessesMap.Load(utils.Bytes2string(*wit.Addr)); ok {
//			this.witnessBackup.witnesses[i].VoteNum = v.(*BackupWitness).VoteNum
//		}
//	}
//}

type LHBalance struct {
	Address         string
	ContractAddress string
	Value           string
}

type LHParams struct {
	AllMainCoin bool
	LHBalances  []*LHBalance
}

// 锁高度地址余额
func LockHeightAllAddrBalance(param *LHParams) (uint64, []*LHBalance) {
	type tokenAddrInfo struct {
		Address         crypto.AddressCoin
		ContractAddress crypto.AddressCoin
		Value           string
	}
	mainCoinAddrs := []crypto.AddressCoin{}
	tokenAddrs := []tokenAddrInfo{}
	for _, item := range param.LHBalances {
		if item.ContractAddress != "" {
			addr := crypto.AddressFromB58String(item.Address)
			caddr := crypto.AddressFromB58String(item.ContractAddress)
			tokenAddrs = append(tokenAddrs, tokenAddrInfo{
				Address:         addr,
				ContractAddress: caddr,
			})
		} else {
			addr := crypto.AddressFromB58String(item.Address)
			mainCoinAddrs = append(mainCoinAddrs, addr)
		}
	}

	workModeLockStatic.GetLock()
	lhb := []*LHBalance{}
	allMainCoinItems := make(map[string][]byte)
	mainCoinItems := []*db2.KVPair{}
	if param.AllMainCoin {
		allMainCoinItems = db.LevelDB.WrapLevelDBPrekeyRange(config.DBKEY_addr_value)
	} else if !param.AllMainCoin && len(mainCoinAddrs) > 0 {
		for _, addr := range mainCoinAddrs {
			dbkey := config.BuildAddrValue(addr)
			valueBs, err := db.LevelDB.Find(dbkey)
			if err != nil {
				continue
			}
			mainCoinItems = append(mainCoinItems, &db2.KVPair{
				Key:   addr,
				Value: *valueBs,
			})
		}
	}
	for i, tokenAddr := range tokenAddrs {
		balance := precompiled.GetBigBalanceWithCryptoAddress(tokenAddr.Address, tokenAddr.Address, tokenAddr.ContractAddress)
		tokenAddrs[i].Value = balance.String()
	}
	height := GetLongChain().CurrentBlock
	for k, v := range allMainCoinItems {
		addr := crypto.AddressCoin(k[4:])
		frozenValue := GetAddrFrozenValue(&addr)
		value := utils.BytesToUint64(v)
		lhb = append(lhb, &LHBalance{
			Address: addr.B58String(),
			Value:   fmt.Sprintf("%d", value+frozenValue),
		})
	}
	for _, item := range mainCoinItems {
		addr := crypto.AddressCoin(item.Key)
		frozenValue := GetAddrFrozenValue(&addr)
		value := utils.BytesToUint64(item.Value)
		lhb = append(lhb, &LHBalance{
			Address: addr.B58String(),
			Value:   fmt.Sprintf("%d", value+frozenValue),
		})
	}
	workModeLockStatic.BackLock()
	for _, tokenAddr := range tokenAddrs {
		lhb = append(lhb, &LHBalance{
			Address:         tokenAddr.Address.B58String(),
			Value:           tokenAddr.Value,
			ContractAddress: tokenAddr.ContractAddress.B58String(),
		})
	}
	return height, lhb
}
