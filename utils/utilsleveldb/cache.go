package utilsleveldb

import (
	"sync"
	"web3_gui/utils"
)

type Cache struct {
	subLock *sync.RWMutex        //
	sub     map[string]*struct{} //需要删除的数据
	addLock *sync.RWMutex        //
	add     map[string]*[]byte   //需要保存和修改的数据
}

func NewCache() *Cache {
	cache := Cache{
		subLock: new(sync.RWMutex),
		sub:     make(map[string]*struct{}),
		addLock: new(sync.RWMutex),
		add:     make(map[string]*[]byte),
	}
	return &cache
}

func (this *Cache) Save(key, value *[]byte) {
	//newKey := make([]byte, len(*key))
	//copy(newKey, *key)
	this.addLock.Lock()
	this.add[utils.Bytes2string(*key)] = value
	this.addLock.Unlock()
}

func (this *Cache) SaveMore(kvMore ...KVMore) {
	this.addLock.Lock()
	for _, one := range kvMore {
		//newKey := make([]byte, len(one.Key))
		//copy(newKey, one.Key)
		//this.add[utils.Bytes2string(newKey)] = &one.Value
		this.add[utils.Bytes2string(one.Key.Byte())] = &one.Value
	}
	//utils.Log.Info().Msgf("保存的记录:%+v", this.add)
	this.addLock.Unlock()
}

/*
查找一条数据
@return    *[]byte    查询到的结果
@return    bool       是否在删除列表中
@return    error      错误
*/
func (this *Cache) Find(key *[]byte) (*[]byte, bool) {
	//newKey := make([]byte, len(*key))
	//copy(newKey, *key)
	this.subLock.RLock()
	_, ok := this.sub[utils.Bytes2string(*key)]
	this.subLock.RUnlock()
	if ok {
		return nil, true
	}
	//在缓存中查询结果
	this.addLock.RLock()
	value, ok := this.add[utils.Bytes2string(*key)]
	this.addLock.RUnlock()
	if ok {
		return value, false
	}
	return value, false
}

func (this *Cache) Remove(key *[]byte) {
	this.subLock.Lock()
	this.sub[utils.Bytes2string(*key)] = nil
	this.subLock.Unlock()
}

/*
获取要删除的键值对
*/
func (this *Cache) getRemoves() map[string]*struct{} {
	result := make(map[string]*struct{}, 0)
	this.subLock.RLock()
	for k, v := range this.sub {
		result[k] = v
	}
	this.subLock.RUnlock()
	return result
}

/*
获取要保存的键值对
*/
func (this *Cache) getSaves() map[string]*[]byte {
	result := make(map[string]*[]byte, 0)
	this.addLock.RLock()
	for k, v := range this.add {
		result[k] = v
	}
	this.addLock.RUnlock()
	return result
}

func (this *Cache) Base_Save(key *LeveldbKey, value *[]byte) {
	this.addLock.Lock()
	this.add[utils.Bytes2string(key.Byte())] = value
	this.addLock.Unlock()
}

func (this *Cache) Base_SaveMore(kvps ...KVPair) {
	this.addLock.Lock()
	for _, one := range kvps {
		this.add[utils.Bytes2string(one.Key)] = &one.Value
	}
	this.addLock.Unlock()
}

/*
查找一条数据
@return    *[]byte    查询到的结果
@return    bool       是否在删除列表中
@return    error      错误
*/
func (this *Cache) Base_Find(key *LeveldbKey) (*[]byte, bool) {
	this.subLock.RLock()
	_, ok := this.sub[utils.Bytes2string(key.Byte())]
	this.subLock.RUnlock()
	if ok {
		return nil, true
	}
	//在缓存中查询结果
	this.addLock.RLock()
	value, ok := this.add[utils.Bytes2string(key.Byte())]
	this.addLock.RUnlock()
	if ok {
		return value, false
	}
	return value, false
}

/*
查找多条数据
@keys    [][]byte    要查询的key
@must    bool        必须返回的；如果为true，则未查询到则报错
*/
func (this *Cache) Base_FindMore(keys [][]byte, must bool) (map[string][]byte, utils.ERROR) {
	keysRemain, ERR := this.base_FindMore_del(keys, must)
	if ERR.CheckFail() {
		return nil, ERR
	}
	//在缓存中查询结果
	values := make(map[string][]byte, 0)
	this.addLock.RLock()
	for _, one := range keysRemain {
		value, ok := this.add[utils.Bytes2string(one)]
		if ok {
			values[utils.Bytes2string(one)] = *value
		}
	}
	this.addLock.RUnlock()
	return values, utils.NewErrorSuccess()
}

/*
在删除的缓存中查找多条数据，返回剩余没有被删除的keys
@keys    [][]byte    要查询的key
@must    bool        必须返回的；如果为true，则未查询到则报错
@return    [][]byte    bool    剩余没有被删除的keys
*/
func (this *Cache) base_FindMore_del(keys [][]byte, must bool) ([][]byte, utils.ERROR) {
	keysRemain := make([][]byte, 0, len(keys))
	this.subLock.RLock()
	defer this.subLock.RUnlock()
	if must {
		for _, one := range keys {
			_, ok := this.sub[utils.Bytes2string(one)]
			if ok {
				return nil, utils.NewErrorBus(ERROR_CODE_not_found, "")
			}
		}
	} else {
		for _, one := range keys {
			_, ok := this.sub[utils.Bytes2string(one)]
			if !ok {
				keysRemain = append(keysRemain, one)
			}
		}
	}
	return keysRemain, utils.NewErrorSuccess()
}

func (this *Cache) Base_Remove(key *LeveldbKey) {
	this.subLock.Lock()
	this.sub[utils.Bytes2string(key.Byte())] = nil
	this.subLock.Unlock()
}
