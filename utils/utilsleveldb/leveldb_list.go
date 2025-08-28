package utilsleveldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"time"
	"web3_gui/utils"
)

/*
获取带读写锁的索引
*/
func (this *LevelDB) findIndexLock(key string) *IndexLock {
	//缓存中查index
	this.leveldbIndexMapLock.Lock()
	indexLock, ok := this.leveldbIndexMap[key]
	if ok {
		this.leveldbIndexMapLock.Unlock()
	} else {
		indexLock = NewIndexLock(big.NewInt(0))
		this.leveldbIndexMap[key] = indexLock
		this.leveldbIndexMapLock.Unlock()
	}
	return indexLock
}

/*
获取带读写锁的索引
*/
func (this *LevelDB) loadIndex(key []byte) ([]byte, utils.ERROR) {
	var indexBs []byte
	var ERR utils.ERROR
	//查数据库
	dbkeyLen := len(key)
	iter := this.db.NewIterator(util.BytesPrefix(key), nil)
	//从最后一条记录中找到最大的index
	ok := iter.Last()
	if ok {
		indexBs = make([]byte, len(iter.Key())-dbkeyLen)
		copy(indexBs, iter.Key()[dbkeyLen:])
		_, indexBs, ERR = LeveldbParseKey(indexBs)
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		//fmt.Println("查询到的index", indexLock.Index.Bytes())
	} else {
		//没有记录，则index从0开始
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return indexBs, utils.NewErrorSuccess()
}

/*
保存1条记录到List集合中
@return           []byte    返回本记录索引。big.Int类型的byte字节序
*/
func (this *LevelDB) SaveList(dbkey LeveldbKey, value []byte, batch *leveldb.Batch) ([]byte, utils.ERROR) {
	var ERR utils.ERROR
	if ERR = checkValueSize(value); !ERR.CheckSuccess() {
		return nil, ERR
	}
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
	//保存记录
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
	indexKey, ERR := LeveldbBuildKey(indexBig.Bytes())
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	tempKey = append(tempKey, indexKey...)
	var err error
	if batch == nil {
		err = this.db.Put(tempKey, value, nil)
	} else {
		batch.Put(tempKey, value)
	}
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	//fmt.Println("保存记录的key", len(tempKey), tempKey)
	indexLock.Index = indexBig
	return indexBig.Bytes(), utils.NewErrorSuccess()
}

/*
保存多条记录到List集合中
@isTransaction    bool      是否是事务操作
@return           []byte    返回本记录索引。big.Int类型的byte字节序
*/
func (this *LevelDB) SaveListMore(dbkey LeveldbKey, value [][]byte, batch *leveldb.Batch) ([][]byte, utils.ERROR) {
	var ERR utils.ERROR
	var err error
	if len(value) == 0 {
		return nil, utils.NewErrorSuccess()
	}
	for _, one := range value {
		if ERR = checkValueSize(one); !ERR.CheckSuccess() {
			return nil, ERR
		}
	}
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

	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
	indexBs := make([][]byte, 0, len(value))
	for _, one := range value {
		indexLock.Index = new(big.Int).Add(indexLock.Index, big.NewInt(1))
		//保存记录
		indexKey, ERR := LeveldbBuildKey(indexLock.Index.Bytes())
		if !ERR.CheckSuccess() {
			return nil, ERR
		}
		tempKey = append(tempKey, indexKey...)
		if batch == nil {
			err = this.db.Put(tempKey, one, nil)
			if err != nil {
				return nil, utils.NewErrorSysSelf(err)
			}
		} else {
			batch.Put(tempKey, one)
		}
		indexBs = append(indexBs, indexLock.Index.Bytes())
	}
	//fmt.Println("保存记录的key", len(tempKey), tempKey)
	//indexLock.Index = indexBig
	return indexBs, utils.NewErrorSuccess()
}

/*
修改List集合中的记录
@isTransaction    bool      是否是事务操作
*/
func (this *LevelDB) SaveOrUpdateListByIndex(dbkey LeveldbKey, index, value []byte, batch *leveldb.Batch) utils.ERROR {
	var ERR utils.ERROR
	if ERR = checkValueSize(value); !ERR.CheckSuccess() {
		return ERR
	}
	//保存记录
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
	indexKey, ERR := LeveldbBuildKey(index)
	if !ERR.CheckSuccess() {
		return ERR
	}
	tempKey = append(tempKey, indexKey...)
	//fmt.Printf("save key:%+v\n", dbkey)
	var err error
	if batch == nil {
		err = this.db.Put(tempKey, value, nil)
	} else {
		batch.Put(tempKey, value)
	}
	if err != nil {
		return utils.NewErrorSysSelf(err)
	}
	return utils.NewErrorSuccess()
}

/*
删除List集合中1条数据
@isTransaction    bool      是否是事务操作
*/
func (this *LevelDB) RemoveListByIndex(dbkey LeveldbKey, index []byte, batch *leveldb.Batch) utils.ERROR {
	if len(index) == 0 {
		return utils.NewErrorSuccess()
	}
	indexOne, ERR := LeveldbBuildKey(index)
	if !ERR.CheckSuccess() {
		return ERR
	}
	//删除
	//tempKeyPre := append(dbkey.key, DataType_Data_List_bs...)
	tempKeyPre := dbkey.key
	tempKey := append(tempKeyPre, indexOne...)
	if batch == nil {
		err := this.db.Delete(tempKey, nil)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
	} else {
		batch.Delete(tempKey)
	}
	return utils.NewErrorSuccess()
}

/*
删除List集合中多条数据
@isTransaction    bool      是否是事务操作
*/
func (this *LevelDB) RemoveListMore(isTransaction bool, dbkey LeveldbKey, index ...[]byte) utils.ERROR {
	//fmt.Println("开始删除")
	var err error
	if len(index) == 0 {
		return utils.NewErrorSuccess()
	}
	indexBs := make([][]byte, 0, len(index))
	for _, one := range index {
		indexOne, ERR := LeveldbBuildKey(one)
		if !ERR.CheckSuccess() {
			return ERR
		}
		indexBs = append(indexBs, indexOne)
	}
	//多记录操作，没有使用事务的，加上事务
	var tr *leveldb.Transaction
	if isTransaction {
		tr = this.tr
	} else {
		tr, err = this.db.OpenTransaction()
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
		defer func() {
			if err == nil {
				//事务提交
				err = tr.Commit()
				if err != nil {
					tr.Discard()
					//utils.Log.Error().Msgf("事务提交失败:%s", err.Error())
					return
				}
				return
			}
			//事务回滚
			tr.Discard()
			//fmt.Println("事务回滚")
		}()
	}
	//循环删除
	//tempKeyPre := append(dbkey.key, DataType_Data_List_bs...)
	tempKeyPre := dbkey.key
	for _, one := range indexBs {
		tempKey := append(tempKeyPre, one...)
		//fmt.Println("删除的key", tempKey)
		err = tr.Delete(tempKey, nil)
		if err != nil {
			return utils.NewErrorSysSelf(err)
		}
	}
	return utils.NewErrorSuccess()
}

/*
间隔删除List集合中的所有记录
当删除大量数据时，会花很长时间，长期占用数据库，让其他业务无法使用数据库。
可以分批次删除，并且设置每批次间隔时间
@dbkey       []byte           数据库ID
@num         uint64           一次删除条数
@interval    time.Duration    删除间隔时间
*/
func (this *LevelDB) RemoveListInterval(dbkey LeveldbKey, num uint64, interval time.Duration) error {
	var err error
	if num == 0 {
		num = DBRemoveNum
	}
	if interval == 0 {
		interval = DBRemoveInterval
	}
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
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
删除List集合中的所有记录
@dbkey       []byte           数据库ID
*/
func (this *LevelDB) RemoveListAll(dbkey LeveldbKey, batch *leveldb.Batch) error {
	var err error
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
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
修改List集合中的记录
*/
func (this *LevelDB) FindListByIndex(dbkey LeveldbKey, index []byte) (*DBItem, utils.ERROR) {
	var err error
	//保存记录
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
	indexKey, ERR := LeveldbBuildKey(index)
	if !ERR.CheckSuccess() {
		return nil, ERR
	}
	tempKey = append(tempKey, indexKey...)
	//fmt.Println("保存记录的key", len(tempKey), tempKey)
	value, err := this.db.Get(tempKey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, utils.NewErrorSuccess()
		}
		return nil, utils.NewErrorSysSelf(err)
	}
	item := DBItem{
		Index: index,
		Value: value,
	}
	return &item, utils.NewErrorSuccess()
}

/*
查询List集合中的所有项目
*/
func (this *LevelDB) FindListAll(dbkey LeveldbKey) ([]DBItem, error) {
	var err error
	lists := make([]DBItem, 0)
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
	dbkeyLen := len(tempKey)
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	for iter.Next() {
		//fmt.Println("长度：", len(iter.Key()), dbkeyLen, len(DataType_Data_List_bs))
		indexBs := make([]byte, len(iter.Key())-dbkeyLen)
		copy(indexBs, iter.Key()[dbkeyLen:])
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		lists = append(lists, DBItem{
			Index: indexBs[2:],
			Value: value,
		})
	}
	iter.Release()
	err = iter.Error()
	return lists, err
}

/*
查询List集合中记录总条数
@return    uint64    记录总条数
@return    []byte    开始index
@return    []byte    结束index
@return    error     错误
*/
func (this *LevelDB) FindListTotal(dbkey LeveldbKey) (uint64, []byte, []byte, utils.ERROR) {
	//fmt.Println("查找总量")
	total := uint64(0)
	var ERR utils.ERROR
	//var startKey, endKey *[]byte
	//tempKey := append(dbkey.key, DataType_Data_List_bs...)
	tempKey := dbkey.key
	iter := this.db.NewIterator(util.BytesPrefix(tempKey), nil)
	//fmt.Println("查找总量")
	//先解析第一条记录的index
	ok := iter.First()
	if !ok {
		//fmt.Println("查找总量，未找到首记录")
		return total, nil, nil, utils.NewErrorSuccess()
	}
	//fmt.Println("查找总量，找到首记录")
	total++
	dbkeyLen := len(tempKey)
	startIndexBs := make([]byte, len(iter.Key())-dbkeyLen)
	copy(startIndexBs, iter.Key()[dbkeyLen:])
	_, startIndexBs, ERR = LeveldbParseKey(startIndexBs)
	if !ERR.CheckSuccess() {
		return total, nil, nil, ERR
	}
	//计算记录总量
	for iter.Next() {
		//fmt.Println("查找总量，找到下一条记录")
		total++
	}
	//解析最后一条记录index
	ok = iter.Last()
	if !ok {
		//fmt.Println("查找总量，找到最后一条记录")
		return total, nil, nil, utils.NewErrorSuccess()
	}
	endIndexBs := make([]byte, len(iter.Key())-dbkeyLen)
	copy(endIndexBs, iter.Key()[dbkeyLen:])
	_, endIndexBs, ERR = LeveldbParseKey(endIndexBs)
	if !ERR.CheckSuccess() {
		return total, nil, nil, ERR
	}
	//释放这个游标
	iter.Release()
	err := iter.Error()
	if err != nil {
		//fmt.Println("查找总量，查找有错误")
		return total, nil, nil, utils.NewErrorSysSelf(err)
	}
	//fmt.Println("查找总量，查找end")
	return total, startIndexBs, endIndexBs, utils.NewErrorSysSelf(err)
}

/*
查询List集合中一个范围的记录
不包含startIndex
@order    bool    查询顺序。true=从前向后查询;false=从后向前查询;
*/
func (this *LevelDB) FindListRange(dbkey LeveldbKey, startIndex []byte, limit uint64, order bool) ([]DBItem, utils.ERROR) {
	return this.FindRange(dbkey, startIndex, limit, order)
}
