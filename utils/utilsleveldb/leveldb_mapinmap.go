package utilsleveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
	"web3_gui/utils"
)

/*
保存到 Map中有Map 集合中
@dbkey    LeveldbKey    数据库ID
@keyOut   LeveldbKey    外层map索引
@keyIn    LeveldbKey    内层map索引
@value    []byte        内层map索引对应的值
*/
func (this *LevelDB) SaveMapInMap(dbkey, keyOut, keyIn LeveldbKey, value []byte, batch *leveldb.Batch) utils.ERROR {
	if ERR := checkValueSize(value); !ERR.CheckSuccess() {
		return ERR
	}
	tempKey := append(dbkey.key, keyOut.key...)
	tempKey = append(tempKey, keyIn.key...)
	if batch == nil {
		err := this.db.Put(tempKey, value, nil)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
	} else {
		batch.Put(tempKey, value)
	}
	return utils.NewErrorSuccess()
}

/*
保存到 Map中有Map 集合中 事务处理
@dbkey    LeveldbKey    数据库ID
@keyOut   LeveldbKey    外层map索引
@keyIn    LeveldbKey    内层map索引
@value    []byte        内层map索引对应的值
*/
func (this *LevelDB) SaveMapInMap_Transaction(dbkey, keyOut, keyIn LeveldbKey, value []byte) utils.ERROR {
	if ERR := checkValueSize(value); !ERR.CheckSuccess() {
		return ERR
	}
	tempKey := append(dbkey.key, keyOut.key...)
	tempKey = append(tempKey, keyIn.key...)
	err := this.tr.Put(tempKey, value, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
查询 Map中有Map 集合中内层key对应的值
@dbkey    LeveldbKey    数据库ID
@keyOut   LeveldbKey    外层map索引
@keyIn    LeveldbKey    内层map索引
@return   []byte        内层map索引对应的值
*/
func (this *LevelDB) FindMapInMapByKeyIn(dbkey, keyOut, keyIn LeveldbKey) (*[]byte, error) {
	tempKey := append(dbkey.key, keyOut.key...)
	tempKey = append(tempKey, keyIn.key...)
	value, err := this.db.Get(tempKey, nil)
	if err != nil && err == leveldb.ErrNotFound {
		return nil, nil
	}
	return &value, err
}

/*
查询 Map中有Map 集合中内层key对应的值
@dbkey    LeveldbKey    数据库ID
@keyOut   LeveldbKey    外层map索引
@return   []DBItem      内层map索引对应的值
*/
func (this *LevelDB) FindMapInMapByKeyOut(dbkey, keyOut LeveldbKey) ([]DBItem, utils.ERROR) {
	items := make([]DBItem, 0)
	tempKey := append(dbkey.key, keyOut.key...)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		keys, ERR := LeveldbParseKeyMore(key)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		one := DBItem{
			Key:   keys[len(keys)-1],
			Value: value,
		}
		items = append(items, one)
	}
	return items, utils.NewErrorSuccess()
}

/*
查询 Map中有Map 集合中内层key对应的值
@dbkey    LeveldbKey    数据库ID
@keyOut   LeveldbKey    外层map索引
@keyIn    LeveldbKey    内层map索引
@return   []byte        内层map索引对应的值
*/
func (this *LevelDB) RemoveMapInMapByKeyIn(dbkey, keyOut, keyIn LeveldbKey, batch *leveldb.Batch) error {
	tempKey := append(dbkey.key, keyOut.key...)
	tempKey = append(tempKey, keyIn.key...)
	if batch == nil {
		err := this.db.Delete(tempKey, nil)
		return err
	} else {
		batch.Delete(tempKey)
	}
	return nil
}

/*
删除 Map中有Map 集合中外层key对应的值
@dbkey       LeveldbKey       数据库ID
@keyOut      LeveldbKey       外层map索引
*/
func (this *LevelDB) RemoveMapInMapByKeyOut(dbkey, keyOut LeveldbKey, batch *leveldb.Batch) error {
	var err error
	tempKey := append(dbkey.key, keyOut.key...)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		if batch == nil {
			err = this.db.Delete(iter.Key(), nil)
			if err != nil {
				return err
			}
		} else {
			batch.Delete(iter.Key())
		}
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		return err
	}
	return nil
}

/*
查询 Map中有Map 集合中内层key对应的值 事务处理
@dbkey    LeveldbKey    数据库ID
@keyOut   LeveldbKey    外层map索引
@keyIn    LeveldbKey    内层map索引
@return   []byte        内层map索引对应的值
*/
func (this *LevelDB) RemoveMapInMapByKeyIn_Transaction(dbkey, keyOut, keyIn LeveldbKey) error {
	tempKey := append(dbkey.key, keyOut.key...)
	tempKey = append(tempKey, keyIn.key...)
	err := this.tr.Delete(tempKey, nil)
	return err
}

/*
间隔删除 Map中有Map 集合中外层key对应的值
当删除大量数据时，会花很长时间，长期占用数据库，让其他业务无法使用数据库。
可以分批次删除，并且设置每批次间隔时间
@dbkey       LeveldbKey       数据库ID
@keyOut      LeveldbKey       外层map索引
@num         uint64           一次删除条数
@interval    time.Duration    删除间隔时间
*/
func (this *LevelDB) RemoveMapInMapByKeyOutInterval(dbkey, keyOut LeveldbKey, num uint64, interval time.Duration) error {
	var err error
	if num == 0 {
		num = DBRemoveNum
	}
	if interval == 0 {
		interval = DBRemoveInterval
	}
	tempKey := append(dbkey.key, keyOut.key...)
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
