package utilsleveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"sync"
	"web3_gui/utils"
)

type LevelDB struct {
	path                string
	db                  *leveldb.DB
	tr                  *leveldb.Transaction
	once                sync.Once
	leveldbIndexMapLock *sync.RWMutex         //
	leveldbIndexMap     map[string]*IndexLock //保存不同表的index
	cache               *Cache                //缓存
}

func CreateLevelDB(path string) (*LevelDB, error) {
	lldb := LevelDB{
		path:                path,
		once:                sync.Once{},
		leveldbIndexMapLock: new(sync.RWMutex),
		leveldbIndexMap:     make(map[string]*IndexLock),
		cache:               NewCache(),
	}
	err := lldb.InitDB()
	if err != nil {
		return nil, err
	}
	return &lldb, nil
}

// 链接leveldb
func (this *LevelDB) InitDB() (err error) {
	this.once.Do(func() {
		//没有db目录会自动创建
		this.db, err = leveldb.OpenFile(this.path, nil)
		if err != nil {
			return
		}
		return
	})
	return
}

/*
保存
*/
func (this *LevelDB) Save(key LeveldbKey, bs *[]byte) utils.ERROR {
	if ERR := checkValueSize(*bs); !ERR.CheckSuccess() {
		return ERR
	}
	err := this.db.Delete(key.key, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		err = this.db.Put(key.key, nil, nil)
	} else {
		err = this.db.Put(key.key, *bs, nil)
	}
	return utils.NewErrorSysSelf(err)
}

/*
保存
*/
func (this *LevelDB) SaveBatch(batch *leveldb.Batch) error {
	return this.db.Write(batch, nil)
}

/*
保存 事务处理
*/
func (this *LevelDB) Save_Transaction(key LeveldbKey, bs *[]byte) utils.ERROR {
	if ERR := checkValueSize(*bs); !ERR.CheckSuccess() {
		return ERR
	}
	err := this.tr.Delete(key.key, nil)
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	if bs == nil {
		err = this.tr.Put(key.key, nil, nil)
	} else {
		err = this.tr.Put(key.key, *bs, nil)
	}
	return utils.NewErrorSysSelf(err)
}

/*
保存多条数据
*/
func (this *LevelDB) SaveMore(kvps ...KVPair) error {
	err := this.OpenTransaction()
	if err != nil {
		return err
	}
	err = this.SaveMore_TransacTion(kvps...)
	if err != nil {
		this.Discard()
		return err
	}
	err = this.Commit()
	if err != nil {
		this.Discard()
		return err
	}
	return nil
}

/*
保存多条数据 事务处理
*/
func (this *LevelDB) SaveMore_TransacTion(kvps ...KVPair) error {
	var err error
	for _, one := range kvps {
		if one.IsAddOrDel {
			//levedb保存相同的key，原来的key保存的数据不会删除，因此保存之前先删除原来的数据
			err = this.tr.Delete(one.Key, nil)
			if err != nil {
				return err
			}
			err = this.tr.Put(one.Key, one.Value, nil)
		} else {
			err = this.tr.Delete(one.Key, nil)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

/*
删除
*/
func (this *LevelDB) Remove(key LeveldbKey) error {
	return this.db.Delete(key.key, nil)
}

/*
删除 事务处理
*/
func (this *LevelDB) Remove_Transaction(key LeveldbKey) error {
	return this.tr.Delete(key.key, nil)
}

/*
删除多个操作
*/
func (this *LevelDB) RemoveMore(kvps ...KVPair) error {
	return this.SaveMore(kvps...)
}

/*
删除多个操作 事务处理
*/
func (this *LevelDB) RemoveMore_Transaction(kvps ...KVPair) error {
	return this.SaveMore_TransacTion(kvps...)
}

/*
查找
*/
func (this *LevelDB) Find(key LeveldbKey) (*DBItem, error) {
	items, err := this.FindMore(key)
	if err != nil {
		return nil, err
	}
	if items != nil && len(items) > 0 {
		return &items[0], nil
	}
	return nil, nil

}

/*
查找多条记录
*/
func (this *LevelDB) FindMore(keys ...LeveldbKey) ([]DBItem, error) {
	items := make([]DBItem, 0, len(keys))
	sn, err := this.db.GetSnapshot()
	if err != nil {
		return nil, err
	}
	for _, one := range keys {
		value, err := sn.Get(one.key, nil)
		if err != nil {
			if err == leveldb.ErrNotFound {
				return nil, nil
			}
			sn.Release()
			return nil, err
		}
		item := DBItem{
			Key:   one,
			Value: value,
		}
		items = append(items, item)
	}
	sn.Release()
	return items, nil
}

//var once = new(sync.Once)

/*
打印所有key
*/
func (this *LevelDB) PrintAll() ([][]byte, error) {
	//once.Do(func() {
	//	golog.InitLogger("logs", 0, true)
	//})
	iter := this.db.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		if len(value) > 100 {
			utils.Log.Info().Hex("key", iter.Key()).Int("valueLength", len(iter.Value())).
				Hex("value", iter.Value()[:100]).Send()
			//golog.Info("key:%+v valueLength:%d %+v", iter.Key(), len(value), value[:100])
		} else {
			utils.Log.Info().Hex("key", iter.Key()).Int("valueLength", len(iter.Value())).
				Hex("value", iter.Value()).Send()
			//golog.Info("key:%+v valueLength:%d %+v", iter.Key(), len(value), value)
		}
		// fmt.Println("key", hex.EncodeToString(iter.Key()), "value", hex.EncodeToString(iter.Value()))
		//fmt.Println("key", iter.Key(), "value", iter.Value())
	}
	utils.Log.Info().Str("end", "-----------------------").Send()
	//golog.Info("end---------------------")
	iter.Release()
	err := iter.Error()
	return nil, err
}

/*
检查key是否存在
@return    bool    true:存在;false:不存在;
*/
func (this *LevelDB) Has(key LeveldbKey) (bool, error) {
	sn, err := this.db.GetSnapshot()
	if err != nil {
		return false, err
	}
	has, err := sn.Has(key.key, nil)
	sn.Release()
	return has, err
}

/*
获取数据库连接
*/
func (this *LevelDB) GetDB() *leveldb.DB {
	return this.db
}

/*
关闭leveldb连接
*/
func (this *LevelDB) Close() {
	this.db.Close()
}

/*
开启事务
*/
func (this *LevelDB) OpenTransaction() error {
	var err error
	this.tr, err = this.db.OpenTransaction()
	return err
}

/*
提交事务
*/
func (this *LevelDB) Commit() error {
	err := this.tr.Commit()
	if err != nil {
		this.tr.Discard()
		return err
	}
	return nil
}

/*
回滚事务
*/
func (this *LevelDB) Discard() {
	this.tr.Discard()
}

/*
获取数据库所在磁盘路径
*/
func (this *LevelDB) GetPath() string {
	return this.path
}

/*
分页查询
*/
func (this *LevelDB) FindRange(dbkey LeveldbKey, startIndex []byte, limit uint64, order bool) ([]DBItem, utils.ERROR) {
	total := uint64(0)
	var err error
	lists := make([]DBItem, 0)
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
	//dbkeyLen := len(tempKey)

	//fmt.Printf("find key:%+v\n", tempKey)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	//fmt.Println("查询的起始字节", len(startIndex), startIndex)
	have := false
	if startIndex != nil && len(startIndex) > 0 {
		startIndexBs, ERR := LeveldbBuildKey(startIndex)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		//tempKey = append(dbkey.key, DataType_Data_List_bs...)
		tempKey = dbkey.key
		tempKey = append(tempKey, startIndexBs...)
		have = iter.Seek(tempKey)
		if !have {
			return nil, utils.NewErrorSuccess()
		}
		//这里决定了不包含startIndex
		if order {
			have = iter.Next()
		} else {
			have = iter.Prev()
		}
	} else {
		if order {
			have = iter.First()
		} else {
			have = iter.Last()
		}
	}
	if !have {
		return nil, utils.NewErrorSuccess()
	}
	for {
		//fmt.Println("key:", iter.Key(), "value:", iter.Value())
		total++
		keyBs := make([]byte, len(iter.Key()))
		copy(keyBs, iter.Key())
		//keys, ERR := LeveldbParseKeyMore(keyBs)
		//if ERR.CheckFail() {
		//	return nil, ERR
		//}
		//index, ERR := keys[len(keys)-1].BaseKey()
		//if ERR.CheckFail() {
		//	return nil, ERR
		//}

		//indexBs := make([]byte, len(iter.Key())-dbkeyLen)
		//copy(indexBs, iter.Key()[dbkeyLen:])
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		itemOne := DBItem{
			//Index: indexBs[2:],
			//Index: index,
			Key:   LeveldbKey{keyBs},
			Value: value,
		}
		//utils.Log.Info().Msgf("保存一个记录:%+v", itemOne)
		lists = append(lists, itemOne)
		if limit == 0 {
			//查询所有
		} else if total >= limit {
			break
		}
		if order {
			have = iter.Next()
		} else {
			have = iter.Prev()
		}
		if !have {
			break
		}
	}
	iter.Release()
	err = iter.Error()
	return lists, utils.NewErrorSysSelf(err)
}

/*
分页查询，包含startIndex
*/
func (this *LevelDB) FindRangeExclude(dbkey, startKey *LeveldbKey, limit uint64, order bool,
	exclude map[string]*struct{}) ([]DBItem, utils.ERROR) {
	total := uint64(0)
	var err error
	lists := make([]DBItem, 0)
	iter := this.db.NewIterator(util.BytesPrefix(dbkey.Byte()), nil)
	have := false
	if startKey != nil {
		//startIndexBs, ERR := LeveldbBuildKey(startIndex)
		//if !ERR.CheckSuccess() {
		//	return nil, ERR
		//}
		tempKey := dbkey.JoinKey(startKey).Byte()
		have = iter.Seek(tempKey)
		if !have {
			return nil, utils.NewErrorSuccess()
		}
		//这里决定了不包含startIndex
		//if order {
		//	have = iter.Next()
		//} else {
		//	have = iter.Prev()
		//}
	} else {
		if order {
			have = iter.First()
		} else {
			have = iter.Last()
		}
	}
	if !have {
		return nil, utils.NewErrorSuccess()
	}
	for {
		//排除缓存中删除的项目
		if _, ok := exclude[utils.Bytes2string(iter.Key())]; !ok {
			total++
			keyBs := make([]byte, len(iter.Key()))
			copy(keyBs, iter.Key())
			value := make([]byte, len(iter.Value()))
			copy(value, iter.Value())
			itemOne := DBItem{
				Key:   LeveldbKey{keyBs},
				Value: value,
			}
			//utils.Log.Info().Msgf("保存一个记录:%+v", itemOne)
			lists = append(lists, itemOne)
		}
		if limit == 0 {
			//查询所有
		} else if total >= limit {
			break
		}
		if order {
			have = iter.Next()
		} else {
			have = iter.Prev()
		}
		if !have {
			break
		}
	}
	iter.Release()
	err = iter.Error()
	return lists, utils.NewErrorSysSelf(err)
}

/*
获取index
*/
func (this *LevelDB) GetIndex(dbkey *LeveldbKey) ([]byte, utils.ERROR) {
	//缓存中查index
	indexLock := this.findIndexLock(utils.Bytes2string(dbkey.key))
	indexLock.lock.Lock()
	defer indexLock.lock.Unlock()
	//首次保存，需要查询数据库index最大值
	if indexLock.Index.Uint64() == 0 {
		//没有就查数据库
		//tempDataKey := append(dbkey.key, DataType_Data_List_bs...)
		tempDataKey := dbkey.key
		indexBs, ERR := this.loadIndex(tempDataKey)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		if indexBs != nil && len(indexBs) > 0 {
			indexLock.Index = new(big.Int).SetBytes(indexBs)
		}
	}
	indexBig := new(big.Int).Add(indexLock.Index, big.NewInt(1))
	indexLock.Index = indexBig
	return indexBig.Bytes(), utils.NewErrorSuccess()
}
