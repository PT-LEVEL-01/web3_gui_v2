package utils

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
)

/*
保存到MapInList集合中
*/
func (this *LevelDB) SaveMapInList(dbkey, key, value []byte) (uint64, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return 0, err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return 0, err
	}
	if err = checkValueSize(value); err != nil {
		return 0, err
	}
	tr, err := this.db.OpenTransaction()
	if err != nil {
		return 0, err
	}
	//查询Map中是否有这个key存在
	tempKey := append(dbkey, DataType_Data_Map...)
	tempKey = append(tempKey, key...)
	_, err = tr.Get(tempKey, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			//不存在则保存
			err = tr.Put(tempKey, value, nil)
			if err != nil {
				tr.Discard()
				return 0, err
			}
		} else {
			tr.Discard()
			return 0, err
		}
	}
	if value == nil || len(value) == 0 {
		err = tr.Commit()
		if err != nil {
			tr.Discard()
			return 0, err
		}
		return 0, nil
	}
	//查询这个dbkey的最大index是多少
	indexDBKEY := append(dbkey, DataType_Index...)
	indexDBKEY = append(indexDBKEY, key...)
	indexBs, err := tr.Get(indexDBKEY, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			indexBs = []byte{0, 0, 0, 0, 0, 0, 0, 0}
		} else {
			tr.Discard()
			return 0, err
		}
	}
	index := BytesToUint64ByBigEndian(indexBs)
	index++
	indexBs = Uint64ToBytesByBigEndian(index)
	//保存最新的index
	err = tr.Put(indexDBKEY, indexBs, nil)
	if err != nil {
		tr.Discard()
		return 0, err
	}
	//保存记录
	tempKey = append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	tempKey = append(tempKey, indexBs...)
	err = tr.Put(tempKey, value, nil)
	if err != nil {
		tr.Discard()
		return 0, err
	}
	err = tr.Commit()
	if err != nil {
		tr.Discard()
		return 0, err
	}
	return index, nil
}

/*
修改MapInList集合中的一个key的value值
*/
func (this *LevelDB) UpdateMapInListByIndex(dbkey, key []byte, index uint64, value []byte) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return err
	}
	if err = checkValueSize(value); err != nil {
		return err
	}
	tr, err := this.db.OpenTransaction()
	if err != nil {
		return err
	}
	//查询Map中是否有这个key存在
	tempKey := append(dbkey, DataType_Data_Map...)
	tempKey = append(tempKey, key...)
	_, err = tr.Get(tempKey, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			//不存在
			tr.Discard()
			return err
		} else {
			tr.Discard()
			return err
		}
	}
	indexBs := Uint64ToBytesByBigEndian(index)
	//修改记录
	tempKey = append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	tempKey = append(tempKey, indexBs...)
	_, err = tr.Get(tempKey, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			//不存在
			tr.Discard()
			return err
		} else {
			tr.Discard()
			return err
		}
	}
	err = tr.Put(tempKey, value, nil)
	if err != nil {
		tr.Discard()
		return err
	}
	err = tr.Commit()
	if err != nil {
		tr.Discard()
		return err
	}
	return nil
}

/*
查找MapInList集合中的一个key的value值
*/
func (this *LevelDB) FindMapInListByIndex(dbkey, key []byte, index uint64) ([]byte, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return nil, err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return nil, err
	}

	//查询Map中是否有这个key存在
	indexBs := Uint64ToBytesByBigEndian(index)
	tempKey := append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	tempKey = append(tempKey, indexBs...)
	value, err := this.db.Get(tempKey, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			//不存在
			return nil, err
		} else {
			return nil, err
		}
	}
	return value, nil
}

/*
删除MapInList集合中的一个key
*/
func (this *LevelDB) RemoveMapInListByKey(dbkey, key []byte) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return err
	}
	tr, err := this.db.OpenTransaction()
	if err != nil {
		return err
	}
	//查询Map中是否有这个key存在
	tempKey := append(dbkey, DataType_Data_Map...)
	tempKey = append(tempKey, key...)
	err = tr.Delete(tempKey, nil)
	if err != nil {
		tr.Discard()
		return err

	}
	//查询这个dbkey的最大index是多少
	indexDBKEY := append(dbkey, DataType_Index...)
	indexDBKEY = append(indexDBKEY, key...)
	err = tr.Delete(indexDBKEY, nil)
	if err != nil {
		tr.Discard()
		return err

	}
	//删除记录
	tempKey = append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	// tempKey = append(tempKey, indexBs...)
	iter := tr.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		err = tr.Delete(iter.Key(), nil)
		if err != nil {
			tr.Discard()
			return err
		}
	}
	err = tr.Commit()
	if err != nil {
		tr.Discard()
		return err
	}
	return nil
}

/*
间隔删除MapInList集合中的一个key
当删除大量数据时，会花很长时间，长期占用数据库，让其他业务无法使用数据库。
可以分批次删除，并且设置每批次间隔时间
@num         uint64           一次删除条数
@interval    time.Duration    删除间隔时间
*/
func (this *LevelDB) RemoveMapInListByKeyInterval(dbkey, key []byte, num uint64, interval time.Duration) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return err
	}
	if num == 0 {
		num = DBRemoveNum
	}
	if interval == 0 {
		interval = DBRemoveInterval
	}

	//查询这个dbkey的最大index是多少
	indexDBKEY := append(dbkey, DataType_Index...)
	indexDBKEY = append(indexDBKEY, key...)
	err = this.db.Delete(indexDBKEY, nil)
	if err != nil {
		return err
	}

	tempKey := append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	// tempKey := append(dbkey, DataType_Data_Map...)
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
删除MapInList集合中的一个item
*/
func (this *LevelDB) RemoveMapInListByIndex(dbkey, key []byte, index uint64) error {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return err
	}
	tr, err := this.db.OpenTransaction()
	if err != nil {
		return err
	}
	//查询Map中是否有这个key存在
	tempKey := append(dbkey, DataType_Data_Map...)
	tempKey = append(tempKey, key...)
	_, err = tr.Get(tempKey, nil)
	if err != nil {
		if err.Error() == leveldb.ErrNotFound.Error() {
			//不存在
			tr.Discard()
			return err
		} else {
			tr.Discard()
			return err
		}
	}
	indexBs := Uint64ToBytesByBigEndian(index)
	//修改记录
	tempKey = append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	tempKey = append(tempKey, indexBs...)
	err = tr.Delete(tempKey, nil)
	if err != nil {
		tr.Discard()
		return err
	}
	err = tr.Commit()
	if err != nil {
		tr.Discard()
		return err
	}
	return nil
}

/*
查询MapInList集合中所有key
*/
func (this *LevelDB) FindMapInListAllKeyToList(dbkey []byte) ([]DBItem, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return nil, err
	}
	lists := make([]DBItem, 0)
	tempKey := append(dbkey, DataType_Data_Map...)
	dbkeyLen := len(tempKey)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		key := make([]byte, len(iter.Key())-dbkeyLen)
		copy(key, iter.Key()[dbkeyLen:])
		_, key, err = LeveldbParseKey(key)
		if err != nil {
			break
		}
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		item := DBItem{
			Key:   key,
			Value: value,
		}
		lists = append(lists, item)
	}
	iter.Release()
	if err != nil {
		return nil, err
	}
	err = iter.Error()
	return lists, err
}

/*
查询MapInList集合中一个key下的列表
*/
func (this *LevelDB) FindMapInListKeyList(dbkey, key []byte) ([]DBItem, error) {
	oldKey := key
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return nil, err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return nil, err
	}
	lists := make([]DBItem, 0)
	tempKey := append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	keyLen := len(tempKey)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		indexBs := make([]byte, len(iter.Key())-keyLen)
		copy(indexBs, iter.Key()[keyLen:])
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		lists = append(lists, DBItem{
			Index: indexBs,
			Key:   oldKey,
			Value: value,
		})
	}
	iter.Release()
	err = iter.Error()
	return lists, err
}

/*
查询MapInList集合中一个key下的列表记录总数
*/
func (this *LevelDB) FindMapInListCount(dbkey, key []byte) (uint64, error) {
	var err error
	if dbkey, err = LeveldbBuildKey(dbkey); err != nil {
		return 0, err
	}
	if key, err = LeveldbBuildKey(key); err != nil {
		return 0, err
	}
	tempKey := append(dbkey, DataType_Data_List...)
	tempKey = append(tempKey, key...)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	count := uint64(0)
	for iter.Next() {
		count++
	}
	iter.Release()
	err = iter.Error()
	return count, err
}
