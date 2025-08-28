package utilsleveldb

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
	"web3_gui/utils"
)

/*
保存到Map
*/
func (this *LevelDB) SaveMap(dbkey, key LeveldbKey, value []byte, batch *leveldb.Batch) utils.ERROR {
	var ERR utils.ERROR
	if ERR = checkValueSize(value); !ERR.CheckSuccess() {
		return ERR
	}
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	tempKey = append(tempKey, key.key...)
	var err error
	if batch == nil {
		err = this.db.Put(tempKey, value, nil)
	} else {
		batch.Put(tempKey, value)
	}
	return utils.NewErrorSysSelf(err)
}

/*
保存到Map 事务处理
*/
func (this *LevelDB) SaveMap_Transaction(dbkey, key LeveldbKey, value []byte) utils.ERROR {
	var ERR utils.ERROR
	if ERR = checkValueSize(value); !ERR.CheckSuccess() {
		return ERR
	}
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	tempKey = append(tempKey, key.key...)
	err := this.tr.Put(tempKey, value, nil)
	return utils.NewErrorSysSelf(err)
}

/*
保存到Map，可以一次保存多个
*/
//func (this *LevelDB) SaveMapMore(dbkey LeveldbKey, kvs ...KVPair) error {
//	tempKey := append(dbkey.key, DataType_Data_Map_bs...)
//	batch := new(leveldb.Batch)
//	for _, one := range kvs {
//		if one.IsAddOrDel {
//			batch.Put(append(tempKey, one.Key...), one.Value)
//		} else {
//			batch.Delete(append(tempKey, one.Key...))
//		}
//	}
//	return this.db.Write(batch, nil)
//}

/*
保存到Map，可以一次保存多个 事务处理
*/
//func (this *LevelDB) SaveMapMore_Transaction(dbkey LeveldbKey, kvs ...KVPair) error {
//	tempKey := append(dbkey.key, DataType_Data_Map_bs...)
//	var err error
//	for _, one := range kvs {
//		if one.IsAddOrDel {
//			err = this.tr.Put(append(tempKey, one.Key...), one.Value, nil)
//		} else {
//			err = this.tr.Delete(append(tempKey, one.Key...), nil)
//		}
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

/*
查询Map
*/
func (this *LevelDB) FindMap(dbkey, key LeveldbKey) ([]byte, error) {
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	value, err := this.db.Get(append(tempKey, key.key...), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

/*
查询Map中所有的key和value值
*/
func (this *LevelDB) FindMapByKeys(dbkey LeveldbKey, keys ...LeveldbKey) ([]DBItem, error) {
	//golog.InitLogger("logs", 0, true)
	//
	//this.PrintAll()
	var err error
	lists := make([]DBItem, 0)
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	//golog.Info("查询的key:%+v", tempKey)
	dbkeyLen := len(tempKey)
	iter := this.db.NewIterator(nil, nil)
	for _, one := range keys {
		//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
		tempKey := dbkey.key
		tempKey = append(tempKey, one.key...)
		if !iter.Seek(tempKey) {
			//未找到
			continue
		}
		//验证key是否相等
		if !bytes.Equal(iter.Key(), tempKey) {
			//未找到
			continue
		}
		key := make([]byte, len(iter.Key())-(dbkeyLen))
		copy(key, iter.Key()[dbkeyLen:])
		ldbKey := LeveldbKey{key}
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		item := DBItem{
			Key:   ldbKey,
			Value: value,
		}
		lists = append(lists, item)
	}
	iter.Release()
	err = iter.Error()
	return lists, err
}

/*
查询Map中所有的key和value值
*/
func (this *LevelDB) FindMapAllToList(dbkey LeveldbKey) ([]DBItem, error) {
	var err error
	lists := make([]DBItem, 0)
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	dbkeyLen := len(tempKey)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		key := make([]byte, len(iter.Key())-(dbkeyLen))
		copy(key, iter.Key()[dbkeyLen:])
		ldbKey := LeveldbKey{key}
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		item := DBItem{
			Key:   ldbKey,
			Value: value,
		}
		lists = append(lists, item)
	}
	iter.Release()
	err = iter.Error()
	return lists, err
}

/*
删除Map中的一个key
*/
func (this *LevelDB) RemoveMapByKey(dbkey, key LeveldbKey, batch *leveldb.Batch) error {
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	tempKey = append(tempKey, key.key...)
	if batch == nil {
		err := this.db.Delete(tempKey, nil)
		return err
	} else {
		batch.Delete(tempKey)
		return nil
	}
}

/*
删除Map中的一个key 事务处理
*/
func (this *LevelDB) RemoveMapByKey_Transaction(dbkey, key LeveldbKey) error {
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	tempKey = append(tempKey, key.key...)
	return this.tr.Delete(tempKey, nil)
}

/*
删除Map
当删除大量数据时，会花很长时间，长期占用数据库，让其他业务无法使用数据库。
可以分批次删除，并且设置每批次间隔时间
@num         uint64           一次删除条数
@interval    time.Duration    删除间隔时间
*/
func (this *LevelDB) RemoveMapByDbKey(dbkey LeveldbKey, num uint64, interval time.Duration) error {
	var err error
	if num == 0 {
		num = DBRemoveNum
	}
	if interval == 0 {
		interval = DBRemoveInterval
	}
	//tempKey := append(dbkey.key, DataType_Data_Map_bs...)
	tempKey := dbkey.key
	ticker := time.NewTicker(time.Nanosecond)
	defer ticker.Stop()
	total := uint64(0)
	for range ticker.C {
		total = 0
		iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
		for iter.Next() {
			err = this.db.Delete(iter.Key(), nil)
			if err != nil {
				return err
			}
			total++
			if total >= num {
				break
			}
		}
		iter.Release()
		err = iter.Error()
		if err != nil {
			return err
		}
		if total == 0 {
			break
		}
		ticker.Reset(interval)
	}
	return nil
}

/*
查询Map集合中一个范围的记录
不包含startIndex
@order    bool    查询顺序。true=从前向后查询;false=从后向前查询;
*/
func (this *LevelDB) FindMapAllToListRange(dbkey LeveldbKey, startIndex []byte, limit uint64, order bool) ([]DBItem, utils.ERROR) {
	return this.FindRange(dbkey, startIndex, limit, order)
}
