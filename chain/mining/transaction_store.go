package mining

import (
	"web3_gui/chain/config"
	"web3_gui/chain/db"
	db2 "web3_gui/chain/db/leveldb"
	"web3_gui/libp2parea/adapter/engine"
	"web3_gui/utils"
)

/*
*
保存区块交易
*/
func SaveBhvoTxs(bhvo *BlockHeadVO) {
	txLen := len(bhvo.Txs)
	keys := make([][]byte, 0, txLen)
	for _, v := range bhvo.Txs {
		v.BuildHash()
		txhashkey := config.BuildBlockTx(*v.GetHash())
		bs, err := v.Proto()
		if err != nil {
			engine.Log.Error("tx proto error：%s", err.Error())
			return
		}
		keys = append(keys, txhashkey)
		db.LevelDB.Save(txhashkey, bs)
		notImportKey := config.BuildTxNotImport(txhashkey)
		keys = append(keys, notImportKey)
		db.LevelDB.Save(notImportKey, nil)
	}
	// 提交交易增量数据
	db.LevelDB.Commit(keys...)
}

/*
*
保存区块交易
*/
func SaveBhvoTxs_Old(bhvo *BlockHeadVO) {
	txLen := len(bhvo.Txs)
	keys := make([][]byte, 0, txLen)
	for _, v := range bhvo.Txs {
		v.BuildHash()
		//刷新缓存
		txhashkey := config.BuildBlockTx(*v.GetHash())
		TxCache.FlashTxInCache(txhashkey, v)
		keys = append(keys, txhashkey)
	}

	values, err := db.LevelDB.MGet(keys...)
	if err != nil {
		engine.Log.Error("mget tx error：%s", err.Error())
		return
	}

	kvs := make([]db2.KVPair, 0, txLen)
	kvsNotImport := make([]db2.KVPair, 0, txLen)
	tmpKeys := make([]string, 0, txLen)
	for k, v := range values {
		tmpKey := utils.Bytes2string(keys[k])
		if v != nil && len(v) > 0 {
			//清除数据库已存在的交易缓存
			DelCacheTx(tmpKey)
			continue
		}

		tmp := bhvo.Txs[k]
		tmpKeys = append(tmpKeys, tmpKey)

		bs, err := tmp.Proto()
		if err != nil {
			return
		}
		kvs = append(kvs, db2.KVPair{keys[k], *bs})
		kvsNotImport = append(kvsNotImport, db2.KVPair{config.BuildTxNotImport(keys[k]), nil})
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

	// 提交交易增量数据
	db.LevelDB.Commit(keys...)
}
