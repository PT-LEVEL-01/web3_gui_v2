package mining

import (
	"bytes"
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	"web3_gui/chain/mining/cache"
	"web3_gui/chain/protos/go_protos"
	"web3_gui/chain/rpc/limiter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

const (
	CheckSwap_Surplus = 1 << iota
	CheckSwap_Lockheight
)

const CleanSwapCacheTime = 10 * time.Second

type SwapTransaction struct {
	Tx      *TxSwap   //兑换交易
	Linked  []*TxSwap //成交链路
	Surplus *big.Int
}

type SwapTransaction_V0 struct {
	Tx      TxSwap_VO   `json:"tx"`     //兑换交易
	Linked  []TxSwap_VO `json:"linked"` //成交链路
	Surplus string      `json:"surplus"`
}

type TransactionSwapManager struct {
	swapTxPool *sync.Map
	count      int64

	locker       *sync.RWMutex
	promoterPool map[string]*TxSwap
}

var addswaptxlock sync.Mutex

func (tm *TransactionSwapManager) AddTx(tx *TxSwap, step Swap_Step) error {
	if err := tm.checkTx(tx, step); err != nil {
		engine.Log.Debug("TransactionSwapManager 交易验证失败 %s %s", hex.EncodeToString(*tx.GetHash()), err.Error())
		//forks.GetLongChain().Balance.DelLockTx(tx)
		return err
	}

	//冻结挂单
	GetLongChain().TransactionManager.unpackedTransaction.addTx4TokenPay(tx)

	tm.AddTxSignal(tx)

	return nil
}

func (tm *TransactionSwapManager) checkTx(tx *TxSwap, step Swap_Step) error {
	var err error
	if CheckOutOfMemory() {
		engine.Log.Error("TransactionSwapManager AddTx:%s, CheckOutOfMemory: %d", hex.EncodeToString(*tx.GetHash()), config.Wallet_Memory_percentage_max)
		return errors.New("out of memory")
	}

	//交易限速
	config.GetTxRate(tx.Class(), true)

	gasUsed, err := GetTxGasUsed(tx)
	if err != nil {
		engine.Log.Error("TransactionSwapManager AddTx:%s, check GasUsed Error: %s", hex.EncodeToString(*tx.GetHash()), err.Error())
		return errors.Wrap(err, "check gas used")
	}
	tx.SetGasUsed(gasUsed)

	//Gas限流
	if !AddTxLimit(tx) {
		//engine.Log.Error("当前限流器 使用：%d 剩余：%d 交易gasUsed:%d", limiter.RpcReqLimiter.Len(limiter_handletx), limiter.RpcReqLimiter.Surplus(limiter_handletx), GetTxGasUsed(txItr))
		//engine.Log.Error("当前限流器 使用：%d 剩余：%d 交易gasUsed:%d", limiter.RpcReqLimiter.Len(limiter_handletx), limiter.RpcReqLimiter.Surplus(limiter_handletx), gasUsed)
		engine.Log.Error("TransactionSwapManager AddTx:%s, 当前限流器 使用：%d 剩余：%d 交易gasUsed:%d", hex.EncodeToString(*tx.GetHash()), limiter.RpcReqLimiter.Len(limiter_handletx), limiter.RpcReqLimiter.Surplus(limiter_handletx), gasUsed)
		return errors.New("tx limited")
	}

	//验证签名
	err = tx.CheckSignStep(step)
	if err != nil {
		engine.Log.Error("TransactionSwapManager AddTx:%s, Failed to verify transaction signature %s", hex.EncodeToString(*tx.GetHash()), err.Error())
		DelTxLimit(tx)
		return errors.Wrap(err, "sign")
	}

	addswaptxlock.Lock()
	defer addswaptxlock.Unlock()

	txhashkey := config.BuildBlockTx(*tx.GetHash())
	hk := utils.Bytes2string(txhashkey)

	if cache.TransCache.Exists(hk) {
		DelTxLimit(tx)
		engine.Log.Error("TransactionSwapManager AddTx:%s, TxCheckCond.WaitCheck,check cache is exist 65", hex.EncodeToString(*tx.GetHash()))
		return errors.New("cache tx existed")
	}

	bs, err := tx.Proto()
	if err != nil {
		engine.Log.Error("TransactionSwapManager AddTx:%s, tx proto fail:%s", hex.EncodeToString(*tx.GetHash()), err.Error())
		DelTxLimit(tx)
		return errors.Wrap(err, "proto")
	}

	//存入缓存
	ok := cache.TransCache.Set(hk, *bs)
	if ok {
		engine.Log.Warn("TransactionManager AddTx:%s, cache set existed", hex.EncodeToString(*tx.GetHash()))
	}

	//验证冻结操作，避免双花
	txCtrl := GetTransactionCtrl(tx.Class())
	if txCtrl != nil {
		err = txCtrl.CheckMultiplePayments(tx)
		if err != nil {
			// engine.Log.Warn("CheckMultiplePayments %s", hex.EncodeToString(*txItr.GetHash()))
			return err
		}
	}

	//持久化挂单
	if err = tm.storeProm(tx); err != nil {
		engine.Log.Error("TransactionSwapManager AddTx storeProm:%s", err.Error())
		return err
	}
	return nil
}

func (tm *TransactionSwapManager) AddTxSignal(tx *TxSwap) bool {
	t := &SwapTransaction{
		Tx:      tx,
		Linked:  make([]*TxSwap, 0),
		Surplus: new(big.Int).Set(tx.AmountOut),
	}

	tm.locker.Lock()
	tm.swapTxPool.Store(utils.Bytes2string(*tx.GetHash()), t)
	tm.promoterPool[utils.Bytes2string(*tx.GetHash())] = t.Tx
	atomic.AddInt64(&tm.count, 1)
	tm.locker.Unlock()
	return true
}

func NewTransactionSwapManager() *TransactionSwapManager {
	tm := TransactionSwapManager{
		swapTxPool: new(sync.Map), //未打包的交易,key:string=交易hash id；value:TransactionRatio=;
		locker:     new(sync.RWMutex),

		promoterPool: make(map[string]*TxSwap),
	}

	//启动清理swap池任务
	utils.Go(tm.loopCleanSwapTx, nil)
	return &tm
}

// 初始化swap池
func (tm *TransactionSwapManager) Initialize() {
	tm.initializeDB()
	tm.initializeFrozen()
}

// 初始化冻结
func (tm *TransactionSwapManager) initializeFrozen() {
	tm.swapTxPool.Range(func(k, v any) bool {
		tx := v.(*SwapTransaction)
		addr := crypto.BuildAddr(config.AddrPre, tx.Tx.PukPromoter)
		key := utils.Bytes2string(*config.BuildDBKeyTokenAddrFrozenValue(tx.Tx.TokenTxidOut, addr))
		value := new(big.Int).Set(tx.Surplus)
		if addrToken, ok := GetLongChain().TransactionManager.unpackedTransaction.addrsToken.LoadOrStore(key, &AddrToken{
			TokenId: tx.Tx.TokenTxidOut,
			Addr:    addr,
			Value:   value,
		}); ok {
			addrToken.(*AddrToken).Value.Add(addrToken.(*AddrToken).Value, value)
		}

		return true
	})
}

// 初始化db
func (tm *TransactionSwapManager) initializeDB() {
	list := db.LoadSwapTxPromPool()
	if len(list) == 0 {
		return
	}

	for i := range list {
		v := list[i]
		//过滤锁定高度
		if v.TxBase.LockHeight < GetLongChain().GetCurrentBlock() {
			continue
		}
		tm.AddTxSignal(&TxSwap{
			TxBase: TxBase{
				Type:       config.Wallet_tx_type_swap,
				Hash:       v.TxBase.Hash,
				LockHeight: v.TxBase.LockHeight,
			},
			TokenTxidOut:      v.TokenTxidOut,
			TokenTxidIn:       v.TokenTxidIn,
			AmountOut:         new(big.Int).SetBytes(v.AmountOut),
			AmountIn:          new(big.Int).SetBytes(v.AmountIn),
			LockhightPromoter: v.LockhightPromoter,
			PukPromoter:       v.PukPromoter,
			SignPromoter:      v.SignPromoter,
		})

		linkedIds, _ := LoadSwapTxRecord(v.TxBase.Hash)
		if len(linkedIds) > 0 {
			linkedTxs := make([]*TxSwap, 0, len(linkedIds))
			for _, linedId := range linkedIds {
				t, _ := LoadTxBase(linedId)
				linkedTxs = append(linkedTxs, t.(*TxSwap))
			}
			tm.addLinked(v.TxBase.Hash, linkedTxs...)
		}
		//余额不足就删掉
		if txItr, ok := tm.swapTxPool.Load(utils.Bytes2string(v.TxBase.Hash)); ok {
			tx := txItr.(*SwapTransaction)
			if tx.Surplus.Cmp(new(big.Int).SetInt64(0)) <= 0 {
				tm.swapTxPool.Delete(utils.Bytes2string(v.TxBase.Hash))
			}
		}
	}
}

// 定时清理swap缓存池
func (tm *TransactionSwapManager) loopCleanSwapTx() {
	ticker := time.NewTicker(CleanSwapCacheTime)
	for range ticker.C {
		engine.Log.Info("清理swap缓存池 count:%d", tm.count)
		if tm.count > 0 {
			tm.cleanSwapTx()
		}
	}
}

func (tm *TransactionSwapManager) cleanSwapTx() {
	tm.locker.Lock()
	defer tm.locker.Unlock()

	tm.swapTxPool.Range(func(key, value any) bool {
		tm.CheckAndDeltx([]byte(key.(string)), CheckSwap_Surplus|CheckSwap_Lockheight)
		return true
	})
}

func (tm *TransactionSwapManager) GetSwapPool(step Swap_Step) []*SwapTransaction {
	var txs []*SwapTransaction
	tm.locker.RLock()
	defer tm.locker.RUnlock()

	tm.swapTxPool.Range(func(key, value any) bool {
		txs = append(txs, value.(*SwapTransaction))
		return true
	})

	return txs
}

func (tm *TransactionSwapManager) GetSwapTx(txid []byte, step Swap_Step) *TxSwap {
	tm.locker.RLock()
	defer tm.locker.RUnlock()
	switch step {
	case Swap_Step_Promoter:
		return tm.promoterPool[utils.Bytes2string(txid)]
	}
	return nil
}

func (tm *TransactionSwapManager) GetSwapSwapTransaction(txid []byte) *SwapTransaction {
	tm.locker.RLock()
	defer tm.locker.RUnlock()

	if v, ok := tm.swapTxPool.Load(utils.Bytes2string(txid)); ok {
		return v.(*SwapTransaction)
	}
	return nil
}

func StoreSwapTxRecord(txid, value []byte) error {
	key := config.BuildSwapTxKey(txid)
	bs, err := LoadSwapTxRecord(txid)
	if err != nil {
		return err
	}
	bs = append(bs, value)
	v := bytes.Join(bs, nil)
	return db.LevelTempDB.Save(key, &v)
}

func LoadSwapTxRecord(txid []byte) ([][]byte, error) {
	bs, err := db.LevelTempDB.Find(config.BuildSwapTxKey(txid))
	if err != nil {
		return nil, err
	}

	r := make([][]byte, 0)
	for i := 0; i < len(*bs); i += 40 {
		end := i + 40
		if end > len(*bs) {
			end = len(*bs)
		}
		r = append(r, (*bs)[i:end])
	}
	return r, nil
}

// 更新swap池状态
func (tm *TransactionSwapManager) UpdateSwapPoolStatus(hashProm []byte, swapRecv *TxSwap) {
	tm.InitSwap(hashProm, swapRecv)

	//记录swap交易链
	if err := StoreSwapTxRecord(hashProm, swapRecv.Hash); err != nil {
		engine.Log.Error("记录swap交易链失败：%s", err.Error())
	}

	tm.addLinked(hashProm, swapRecv)

	//检查是否需要清除swap池缓存
	tm.locker.Lock()
	tm.CheckAndDeltx(hashProm, CheckSwap_Surplus)
	tm.locker.Unlock()
}

// 查询是否需要删除swap缓存
func (tm *TransactionSwapManager) CheckAndDeltx(hashProm []byte, cond int) {
	txI, ok := tm.swapTxPool.Load(utils.Bytes2string(hashProm))
	if !ok {
		return
	}

	tx := txI.(*SwapTransaction)
	if cond&CheckSwap_Surplus != 0 {
		if !tx.isSurplus() {
			tm.delTx(hashProm)
		}
	}
	if cond&CheckSwap_Lockheight != 0 {
		if tx.Tx.CheckLockHeight(GetLongChain().GetCurrentBlock()) != nil {
			tm.delTx(hashProm)
			if tx.isSurplus() {
				//有余额，需要退款
				GetLongChain().TransactionManager.unpackedTransaction.delTx4TokenPay(tx.Tx)
			}
		}
	}
}

// 删除swaptx
func (tm *TransactionSwapManager) delTx(hashProm []byte) {
	key := utils.Bytes2string(hashProm)
	if _, ok := tm.swapTxPool.Load(key); !ok {
		delete(tm.promoterPool, key)
		goto RemoveSwapTxDb
	}

	tm.swapTxPool.Delete(key)
	delete(tm.promoterPool, key)

RemoveSwapTxDb:
	if err := db.RemoveSwapTxProm(hashProm); err != nil {
		engine.Log.Error("TransactionSwapManager delTx error:%s %s", hex.EncodeToString(hashProm), err.Error())
	}
}

// 添加swap linked
func (tm *TransactionSwapManager) addLinked(hashProm []byte, swapRecv ...*TxSwap) {
	tm.locker.Lock()
	defer tm.locker.Unlock()

	txItr, ok := tm.swapTxPool.Load(utils.Bytes2string(hashProm))
	if !ok {
		return
	}
	tx := txItr.(*SwapTransaction)
	tx.Linked = append(tx.Linked, swapRecv...)

	//更新余额
	for _, v := range swapRecv {
		dstAmount := new(big.Int).Mul(v.AmountOut, v.AmountDeal)
		dstAmount = dstAmount.Div(dstAmount, v.AmountIn)
		tx.Surplus = tx.Surplus.Sub(tx.Surplus, dstAmount)
	}
}

// 存储发单
func (tm *TransactionSwapManager) storeProm(tx *TxSwap) error {
	t := &go_protos.TxSwap{
		TxBase: &go_protos.TxBase{
			Type:       config.Wallet_tx_type_swap,
			Hash:       tx.TxBase.Hash,
			LockHeight: tx.LockHeight,
		},
		TokenTxidOut:      tx.TokenTxidOut,
		TokenTxidIn:       tx.TokenTxidIn,
		AmountOut:         tx.AmountOut.Bytes(),
		AmountIn:          tx.AmountIn.Bytes(),
		PukPromoter:       tx.PukPromoter,
		LockhightPromoter: tx.LockhightPromoter,
		SignPromoter:      tx.SignPromoter,
	}

	return db.SaveSwapTxProm(t)
}

// 初始化池中swap交易
func (tm *TransactionSwapManager) InitSwap(hashProm []byte, swapRecv *TxSwap) {
	if _, ok := tm.swapTxPool.Load(utils.Bytes2string(hashProm)); ok {
		return
	}

	//恢复一笔prom交易
	txProm := TxSwap{
		TxBase: TxBase{
			Type:       config.Wallet_tx_type_swap,
			Hash:       hashProm,
			LockHeight: swapRecv.LockHeight,
		},
		TokenTxidOut:      swapRecv.TokenTxidOut,
		TokenTxidIn:       swapRecv.TokenTxidIn,
		AmountOut:         swapRecv.AmountOut,
		AmountIn:          swapRecv.AmountIn,
		LockhightPromoter: swapRecv.LockhightPromoter,
		PukPromoter:       swapRecv.PukPromoter,
		SignPromoter:      swapRecv.SignPromoter,
	}
	//加进池子
	tm.AddTxSignal(&txProm)

	linkedIds, _ := LoadSwapTxRecord(hashProm)
	if len(linkedIds) > 0 {
		linkedTxs := make([]*TxSwap, 0, len(linkedIds))
		for _, v := range linkedIds {
			t, _ := LoadTxBase(v)
			linkedTxs = append(linkedTxs, t.(*TxSwap))
		}
		tm.addLinked(hashProm, linkedTxs...)
	}
}

func (st *SwapTransaction) GetVOJSON() interface{} {
	r := SwapTransaction_V0{
		Tx:      st.Tx.GetVOJSON().(TxSwap_VO),
		Linked:  make([]TxSwap_VO, 0, len(st.Linked)),
		Surplus: st.Surplus.String(),
	}
	for _, v := range st.Linked {
		r.Linked = append(r.Linked, v.GetVOJSON().(TxSwap_VO))
	}

	return r
}

// 计算是否还有剩余额度
func (st *SwapTransaction) isSurplus() bool {
	return st.Surplus.Cmp(new(big.Int).SetInt64(0)) > 0
}
