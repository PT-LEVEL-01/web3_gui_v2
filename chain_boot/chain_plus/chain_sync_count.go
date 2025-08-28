package chain_plus

import (
	"sync/atomic"
	"time"
	"web3_gui/chain/mining"
	"web3_gui/chain_boot/config"
	"web3_gui/utils"
	"web3_gui/utils/utilsleveldb"
)

type CountBlockFun func(bhvo *mining.BlockHeadVO)

/*
区块同步程序
1.建议将不同系统的统计分开，避免统计性能不同，相互影响。
2.假如半途中要加一种从0高度开始新的规则统计程序，分开则能实现，统一则不行。
*/
type ChainSyncCount struct {
	leveldb         *utilsleveldb.LevelDB    //
	dbKey           *utilsleveldb.LeveldbKey //
	PullHeight      uint64                   //爬取区块高度
	ChainPullFinish *atomic.Bool             //链端是否同步完成
	countFun        CountBlockFun            //
}

func NewChainSyncCount(leveldb *utilsleveldb.LevelDB, dbkey *utilsleveldb.LeveldbKey, countFun CountBlockFun) *ChainSyncCount {
	csc := ChainSyncCount{
		leveldb:         leveldb,
		dbKey:           dbkey,
		PullHeight:      0,
		ChainPullFinish: new(atomic.Bool),
		countFun:        countFun,
	}
	csc.ChainPullFinish.Store(false)
	go csc.runCountBlock()
	return &csc
}

/*
获取链端节点同步到的高度
*/
func (this *ChainSyncCount) GetChainCurrentHeight() uint64 {
	remoteHeight := mining.GetCurrentHeight()
	return remoteHeight
}

/*
获取链端节点同步到的高度
*/
func (this *ChainSyncCount) getBlockHeadByHeight(blockHeight uint64) (*mining.BlockHeadVO, utils.ERROR) {
	blockHeadOne := mining.LoadBlockHeadByHeight(blockHeight)
	bhvo, err := mining.LoadBlockHeadVOByHash(&blockHeadOne.Hash)
	if err != nil {
		utils.Log.Error().Err(err).Send()
		return nil, utils.NewErrorSysSelf(err)
	}
	return bhvo, utils.NewErrorSuccess()
}

/*
加载数据库中的步到高度
*/
func (this *ChainSyncCount) loadSyncHeight() (uint64, utils.ERROR) {
	if this.dbKey == nil || this.leveldb == nil {
		return 0, utils.NewErrorBus(config.ERROR_CODE_CHAIN_leveldb_null, "")
	}
	item, err := this.leveldb.Find(*this.dbKey)
	if err != nil {
		return 0, utils.NewErrorSysSelf(err)
	}
	height := uint64(0)
	if item != nil {
		height = utils.BytesToUint64ByBigEndian(item.Value)
	}
	return height, utils.NewErrorSuccess()
}

/*
保存步到高度
*/
func (this *ChainSyncCount) saveSyncHeight(height uint64) utils.ERROR {
	if this.dbKey == nil || this.leveldb == nil {
		return utils.NewErrorBus(config.ERROR_CODE_CHAIN_leveldb_null, "")
	}
	bs := utils.Uint64ToBytesByBigEndian(height)
	ERR := this.leveldb.Save(*this.dbKey, &bs)
	return ERR
}

/*
获取链端节点同步到的高度
*/
func (this *ChainSyncCount) runCountBlock() {
	//加载数据库中的同步高度
	for {
		pullHeight, ERR := this.loadSyncHeight()
		if ERR.CheckFail() {
			time.Sleep(time.Second)
			continue
		}
		this.PullHeight = pullHeight
		break
	}
	//先追赶统计区块高度
	for {
		remoteHeight := this.GetChainCurrentHeight()
		if remoteHeight == 0 {
			time.Sleep(time.Second)
			continue
		}
		localHeight := atomic.LoadUint64(&this.PullHeight)
		if remoteHeight < localHeight {
			time.Sleep(time.Second)
			continue
		}
		if remoteHeight == localHeight {
			break
		}
		this.countBlockRange(localHeight, remoteHeight)
	}
	//标记为同步完成
	this.ChainPullFinish.Store(true)
	for range time.NewTicker(time.Second).C {
		remoteHeight := this.GetChainCurrentHeight()
		localHeight := atomic.LoadUint64(&this.PullHeight)
		this.countBlockRange(localHeight, remoteHeight)
	}
}

/*
统计区块
*/
func (this *ChainSyncCount) countBlockRange(startHeight, endHeight uint64) {
	//utils.Log.Info().Uint64("查询区块高度", startHeight).Uint64("endHeight", endHeight).Send()
	for i := range endHeight - startHeight {
		//utils.Log.Info().Uint64("查询区块高度", startHeight).Uint64("add", i).Send()
		pullHeight := startHeight + i + 1
		this.countBlockOne(pullHeight)
		atomic.StoreUint64(&this.PullHeight, pullHeight)
		ERR := this.saveSyncHeight(pullHeight)
		//统计程序正确执行，是对统计正确性的保护
		if ERR.CheckFail() {
			panic(ERR.String())
		}
	}
}

/*
统计区块
*/
func (this *ChainSyncCount) countBlockOne(blockHeight uint64) {
	bhvo, ERR := this.getBlockHeadByHeight(blockHeight)
	if ERR.CheckFail() {
		utils.Log.Error().Str("ERR", ERR.String()).Send()
		return
	}
	//
	this.countFun(bhvo)
}

/*
统计区块
*/
func (this *ChainSyncCount) SetDB(leveldb *utilsleveldb.LevelDB) {
	this.leveldb = leveldb
}

/*
统计区块
*/
func (this *ChainSyncCount) GetChainPullFinish() bool {
	return this.ChainPullFinish.Load()
}

/*
获取爬取高度
*/
func (this *ChainSyncCount) GetPullHeight() uint64 {
	return atomic.LoadUint64(&this.PullHeight)
}
