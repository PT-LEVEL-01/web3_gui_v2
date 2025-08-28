package db

import (
	"errors"
	"sync"
	"web3_gui/chain/config"
	db "web3_gui/chain/db/leveldb"
	"web3_gui/libp2parea/adapter/engine"
)

const (
	LevelDbType = iota
	LevelTempDbType
)

var Once_ConnLevelDB sync.Once

var LevelDB *db.LevelDB
var LevelTempDB *db.LevelDB

// 链接leveldb
func InitDB(name, tempDBName string) (err error) {
	Once_ConnLevelDB.Do(func() {
		LevelDB, err = db.CreateLevelDB(name)
		if err != nil {
			return
		}

		LevelTempDB = LevelDB

		// 禁用快照功能需要删除tempdb
		if config.DisableSnapshot || !IsSnapshotExists() {
			if err = RemoveTempDb(); err != nil {
				engine.Log.Error("tempdb清空失败！%s", err.Error())
			}
		}

		return
	})
	return
}

// 删除tempdb
// 只保留 0,1,0 前缀的kv值,0,1 是库标识
func RemoveTempDb() error {
	if count, err := LevelTempDB.LevelDBDelInRange([]byte{0, db.KVType, 1}); err != nil {
		return err
	} else {
		engine.Log.Info("清除db，key %v 数量：%s", []byte{0, db.KVType, 1}, count.String())
	}
	if count, err := LevelTempDB.LevelDBDelInRange([]byte{0, db.HashType}); err != nil {
		return err
	} else {
		engine.Log.Info("清除db，key %v 数量：%s", []byte{0, db.HashType}, count.String())
	}
	if count, err := LevelTempDB.LevelDBDelInRange([]byte{0, db.HSizeType}); err != nil {
		return err
	} else {
		engine.Log.Info("清除db，key %v 数量：%s", []byte{0, db.HSizeType}, count.String())
	}
	return nil
}

// 删除tempdb
func RemoveTempDb_Old() error {
	it := LevelTempDB.NewIterator()
	b := LevelTempDB.GetKvBatch()
	b.Lock()
	defer b.Unlock()

	count := 0
	var err error

	for it.Next() {
		var ty byte
		var key []byte

		ty, _ = LevelTempDB.GetKeyType(it.Key())
		//if err != nil {
		//	return err
		//}

		switch ty {
		case db.KVType:
			key, err = LevelTempDB.DecodeKVKey(it.Key())
		case db.HashType:
			key, _, err = LevelTempDB.HDecodeHashKey(it.Key())
		default:
			err = errors.New("invalid key")
		}
		if err != nil {
			//此处为无效key（未通过封装kv或hash结构进行存储的key），可直接删除
			b.Delete(it.Key())
			continue
		}

		if len(key) < 1 {
			return errors.New("key is too small")
		}

		if key[0] == LevelDbType {
			continue
		}

		b.Delete(it.Key())
		count++
	}

	it.Release()
	err = it.Error()
	if err != nil {
		return err
	}

	err = b.Commit()
	if err != nil {
		return err
	}

	engine.Log.Info("清除tempdb，keys数量：%d", count)
	return nil
}

// 是否存在快照
func IsSnapshotExists() bool {
	ok, _ := LevelTempDB.Exists(config.DBKEY_snapshot_height)
	return ok == 1
}

//var Once_ConnLevelDB sync.Once
//
//var LevelDB *utils.LedisDB
//var LevelTempDB *utils.LedisDB
//
////链接leveldb
//func InitDB(name, tempDBName string) (err error) {
//	Once_ConnLevelDB.Do(func() {
//		LevelDB, err = utils.CreateLedisDB(name)
//		if err != nil {
//			return
//		}
//		os.RemoveAll(tempDBName)
//		LevelTempDB, err = utils.CreateLedisDB(tempDBName)
//		if err != nil {
//			return
//		}
//		return
//	})
//	return
//}

// var Once_ConnLevelDB sync.Once

// var LevelDB *utils.LevelDB
// var LevelTempDB *utils.LevelDB

// //链接leveldb
// func InitDB(name, tempDBName string) (err error) {
// 	Once_ConnLevelDB.Do(func() {
// 		LevelDB, err = utils.CreateLevelDB(name)
// 		if err != nil {
// 			return
// 		}
// 		os.RemoveAll(tempDBName)
// 		LevelTempDB, err = utils.CreateLevelDB(tempDBName)
// 		if err != nil {
// 			return
// 		}
// 		return
// 	})
// 	return
// }
