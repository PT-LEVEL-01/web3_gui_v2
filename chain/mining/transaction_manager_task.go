package mining

import (
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/chain/rpc/limiter"

	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"

	"sync"
	"time"

	"web3_gui/chain/mining/cache"
	"web3_gui/chain/mining/task"
)

const (
	task_tx_cache    = "task_tx_cache"
	task_tx_interval = time.Second * 1
)

var saveCacheLock sync.Mutex

func InitTxCacheTask() {
	newT := task.CreateTask(task_tx_cache, SaveTxCache, task_tx_interval)
	task.TaskM.AddAndDo(task_tx_cache, newT)

	//ptTask := task.CreateTask("task_limiter_print", printLimiter, task_tx_interval)
	//task.TaskM.AddAndDo("task_limiter_print", ptTask)
}

func StopTxCacheTask() {
	task.TaskM.Stop(task_tx_cache)
}

func SaveTxCache() {
	saveCacheLock.Lock()
	defer saveCacheLock.Unlock()

	cacheLen := cache.TransCache.GetLen()

	if cacheLen == 0 {
		return
	}

	//engine.Log.Info("缓存数量：%d", cacheLen)

	keys := make([][]byte, 0, cacheLen)
	cache.TransCache.GetM().Range(func(k, v any) bool {
		keys = append(keys, []byte(k.(string)))
		return true
	})

	values, err := db.LevelDB.MGet(keys...)
	if err != nil {
		engine.Log.Error("mget tx cache error：%s", err.Error())
		return
	}

	kvs := make([]db2.KVPair, 0, cacheLen)
	kvsNotImport := make([]db2.KVPair, 0, cacheLen)
	tmpKeys := make([]string, 0, cacheLen)
	for k, v := range values {
		tmpKey := utils.Bytes2string(keys[k])
		if v != nil && len(v) > 0 {
			//清除数据库已存在的交易缓存
			DelCacheTx(tmpKey)
			continue
		}

		kvsNotImport = append(kvsNotImport, db2.KVPair{config.BuildTxNotImport(keys[k]), nil})

		tmp, ok := cache.TransCache.Get(tmpKey)
		if !ok {
			continue
		}
		tmpKeys = append(tmpKeys, tmpKey)

		kvs = append(kvs, db2.KVPair{keys[k], tmp.([]byte)})
	}

	if len(kvs) == 0 {
		return
	}

	err = db.LevelDB.MSet(kvs...)
	if err != nil {
		engine.Log.Error("mset tx cache error:%s", err.Error())
		return
	}

	db.LevelDB.MSet(kvsNotImport...)

	//清空缓存
	for _, v := range tmpKeys {
		DelCacheTx(v)
	}

	// engine.Log.Info("save tx cache success 》》》》》》》》》》》》》")
}

func printLimiter() {
	engine.Log.Info("当前限流器》》》》》》》》》》》》》：使用：%d 剩余：%d", limiter.RpcReqLimiter.Len(limiter_handletx), limiter.RpcReqLimiter.Surplus(limiter_handletx))
}
