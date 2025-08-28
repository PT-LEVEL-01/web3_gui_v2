package utilsleveldb

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"sort"
	"web3_gui/utils"
)

/*
保存
*/
func (this *LevelDB) Cache_Save(key, value *[]byte) {
	this.cache.Save(key, value)
}

/*
保存多条
*/
func (this *LevelDB) Cache_SaveMore(kvMore ...KVMore) {
	this.cache.SaveMore(kvMore...)
}

/*
查询
*/
func (this *LevelDB) Cache_Find(key *[]byte) (*[]byte, error) {
	//如果在删除列表中，则查询结果为空
	value, ok := this.cache.Find(key)
	if ok {
		return nil, nil
	}
	if value != nil {
		return value, nil
	}
	//缓存中没有，则查询数据库
	value2, err := this.db.Get(*key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &value2, nil
}

/*
删除
*/
func (this *LevelDB) Cache_Remove(key *[]byte) {
	this.cache.Remove(key)
}

/*
提交数据
*/
func (this *LevelDB) Cache_Commit() error {
	this.cache.addLock.Lock()
	defer this.cache.addLock.Unlock()
	this.cache.subLock.Lock()
	defer this.cache.subLock.Unlock()
	batch := new(leveldb.Batch)
	for k, v := range this.cache.add {
		//fmt.Println("key", []byte(k), "value", v)
		//utils.Log.Info().Hex("保存的key", []byte(k)).Hex("value", *v).Send()
		key := make([]byte, len([]byte(k)))
		copy(key, k)
		batch.Put(key, *v)
	}
	for k, _ := range this.cache.sub {
		//utils.Log.Info().Hex("删除的key", []byte(k)).Send()
		batch.Delete([]byte(k))
	}
	err := this.db.Write(batch, nil)
	if err != nil {
		return err
	}
	this.cache.add = make(map[string]*[]byte, 0)
	this.cache.sub = make(map[string]*struct{}, 0)
	return err
}

/*
提交指定的cache数据
*/
func (this *LevelDB) Cache_CommitCache(cache *Cache) error {
	cache.addLock.Lock()
	cache.subLock.Lock()
	batch := new(leveldb.Batch)
	for k, v := range cache.add {
		if v == nil {
			batch.Put([]byte(k), nil)
			continue
		}
		batch.Put([]byte(k), *v)
	}
	for k, _ := range cache.sub {
		batch.Delete([]byte(k))
	}
	err := this.db.Write(batch, nil)
	cache.subLock.Unlock()
	cache.addLock.Unlock()
	return err
}

/*
删除Map中的一个key 事务处理
*/
func (this *LevelDB) Cache_Set_Save(dbkey, key *LeveldbKey, value *[]byte) {
	this.cache.Set_Save(dbkey, key, value)
}

/*
删除Map中的一个key 事务处理
*/
func (this *LevelDB) Cache_Set_Find(dbkey, key *LeveldbKey) (*[]byte, utils.ERROR) {
	newKey := dbkey.JoinKey(key)
	tempKeyBs := newKey.Byte()
	bs, ok := this.cache.Find(&tempKeyBs)
	if ok {
		//在删除列表中
		return nil, utils.NewErrorBus(ERROR_CODE_not_found, "")
	}
	if bs != nil {
		//缓存中找到了
		return bs, utils.NewErrorSuccess()
	}
	item, err := this.Find(*newKey)
	if err != nil {
		return nil, utils.NewErrorSysSelf(err)
	}
	return &item.Value, utils.NewErrorSuccess()
}

/*
删除Map中的一个key 事务处理
*/
func (this *LevelDB) Cache_Set_Remove(dbkey, key *LeveldbKey) {
	this.cache.Set_Remove(dbkey, key)
}

/*
在Map中查找多个key，有一个不存在都报错
*/
func (this *LevelDB) Cache_Set_FindMore(dbkey *LeveldbKey, keys []*LeveldbKey, must bool) ([]DBItem, utils.ERROR) {
	dbkeys := make([][]byte, 0, len(keys))
	for _, one := range keys {
		dbkeys = append(dbkeys, dbkey.JoinKey(one).Byte())
	}
	cacheMap, ERR := this.cache.Base_FindMore(dbkeys, must)
	if ERR.CheckFail() {
		utils.Log.Info().Str("查询set集合 错误", ERR.String()).Send()
		return nil, ERR
	}
	var err error
	lists := make([]DBItem, 0, len(keys))
	iter := this.db.NewIterator(nil, nil)
	for i, one := range dbkeys {
		//先查询缓存
		valueOne, ok := cacheMap[utils.Bytes2string(one)]
		if ok {
			item := DBItem{
				Key:   *keys[i],
				Value: valueOne,
			}
			lists = append(lists, item)
			continue
		}
		//缓存中没有，则查询数据库；验证key是否相等
		if !iter.Seek(one) || !bytes.Equal(iter.Key(), one) {
			if must {
				utils.Log.Info().Str("查询set集合 错误", "").Hex("one", one).Hex("key", iter.Key()).Send()
				return nil, utils.NewErrorBus(ERROR_CODE_not_found, "")
			}
			//未找到
			continue
		}
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		item := DBItem{
			Key:   *keys[i],
			Value: value,
		}
		lists = append(lists, item)
	}
	iter.Release()
	err = iter.Error()
	return lists, utils.NewErrorSysSelf(err)
}

/*
@order    bool    排序；true=升序;false=倒序;
*/
func (this *LevelDB) Cache_Set_FindRange(dbkey, startKey *LeveldbKey, limit uint64, order bool) ([]DBItem, utils.ERROR) {
	//utils.Log.Info().Uint64("查询set集合范围", limit).Hex("dbkey", dbkey.Byte()).Hex("startKey", startKey.Byte()).Send()
	//获取要删除的数据
	exclude := this.cache.getRemoves()
	//先从数据库查询，排除要删除的缓存
	items, ERR := this.FindRangeExclude(dbkey, startKey, limit, order, exclude)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//for i, one := range items {
	//	utils.Log.Info().Int("查看一遍数据", i).Hex("key", one.Key.Byte()).Send()
	//}
	//utils.Log.Info().Uint64("查询set集合范围", limit).Hex("dbkey", dbkey.Byte()).Hex("startKey", startKey.Byte()).Send()

	//获取要保存的数据
	saves := this.cache.getSaves()
	ss := SliceSort{order, make([][]byte, 0, len(items)+len(saves))}
	//放入缓存中的key
	for k, _ := range saves {
		//前缀过滤
		if !bytes.HasPrefix([]byte(k), dbkey.Byte()) {
			continue
		}
		ss.list = append(ss.list, []byte(k))
	}
	//放入数据库中的key
	for _, one := range items {
		_, ok := saves[utils.Bytes2string(one.Key.Byte())]
		if ok {
			continue
		}
		ss.list = append(ss.list, one.Key.Byte())
	}
	//把内存中的key结构按照leveldb中存放顺序排序。
	sort.Sort(ss)
	if startKey != nil {
		for i, one := range ss.list {
			if bytes.Equal(one, startKey.Byte()) {
				ss.list = ss.list[i+1:]
				break
			}
		}
	}
	//截取要查询的数据范围
	if limit != 0 {
		if len(ss.list) < int(limit) {
			limit = uint64(len(ss.list))
		}
		ss.list = ss.list[:limit]
	} else {
		//0是查全部
	}
	//utils.Log.Info().Int("最后list数量", len(ss.list)).Send()
	//还原value
	newItems := make([]DBItem, 0, len(ss.list))
	for _, one := range ss.list {
		//utils.Log.Info().Int("查询数据", i).Hex("key", one).Send()
		//使用缓存中最新的数据
		value, ok := saves[utils.Bytes2string(one)]
		if ok {
			//utils.Log.Info().Int("查询数据", i).Send()
			newItems = append(newItems, DBItem{
				//Index: nil,
				Key:   LeveldbKey{one},
				Value: *value,
			})
			continue
		}
		//utils.Log.Info().Int("查询数据", i).Send()
		//缓存中没有，使用数据库中的查询结果
		for i, itemOne := range items {
			//utils.Log.Info().Hex("对比key1", itemOne.Key.Byte()).Hex("对比key2", one).Send()
			if bytes.Equal(itemOne.Key.Byte(), one) {
				//utils.Log.Info().Hex("命中key", one).Send()
				newItems = append(newItems, itemOne)
				items = items[:i]
				break
			}
		}
	}

	//去掉返回key中的dbkey
	for i, one := range newItems {
		newKey := make([]byte, len(one.Key.Byte())-len(dbkey.Byte()))
		copy(newKey, one.Key.Byte()[len(dbkey.Byte()):])
		newItems[i].Key = LeveldbKey{newKey}
	}

	//去掉startKey记录
	if len(newItems) > 0 && startKey != nil {
		//utils.Log.Info().Hex("最后一条记录 key", lastItem.Key.Byte()).Hex("对比 key ", startKey.Byte()).Send()
		//匹配去掉最后一条记录
		if bytes.Equal(newItems[len(newItems)-1].Key.Byte(), startKey.Byte()) {
			newItems = newItems[:len(newItems)-1]
		}
		//匹配去掉第一条记录
		if bytes.Equal(newItems[0].Key.Byte(), startKey.Byte()) {
			newItems = newItems[1:]
		}
	}
	return newItems, ERR
}

type SliceSort struct {
	direction bool //排序方向
	list      [][]byte
}

func (this SliceSort) Len() int {
	return len(this.list)
}

func (this SliceSort) Less(i, j int) bool {
	a := this.list[i]
	b := this.list[j]
	if len(a) > len(b) {
		return true
	}
	if len(a) == len(b) {
		for index, one := range a {
			if this.direction {
				if one > b[index] {
					return true
				}
			} else {
				if one < b[index] {
					return true
				}
			}
		}
	}
	return false
}

func (this SliceSort) Swap(i, j int) {
	this.list[i], this.list[j] = this.list[j], this.list[i]
}
