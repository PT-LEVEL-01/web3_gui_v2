package leveldb

import (
	"bytes"
	"github.com/syndtr/goleveldb/leveldb/util"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"web3_gui/utils"
)

// db增量快照内存层
type LeveldbSnap struct {
	seq  uint64              //快照序号
	db   *leveldb.DB         //快照库
	keys map[string][]byte   //快照增量kv
	rem  map[string]struct{} //快照标记删除
	mu   sync.RWMutex
}

// 给db设置一个增量快照
func (l *LevelDB) SetSnap() *LeveldbSnap {
	return &LeveldbSnap{
		0,
		l.db,
		make(map[string][]byte),
		make(map[string]struct{}),
		sync.RWMutex{},
	}
}

// 清理存储缓存
func (l *LevelDB) ResetSnap() {
	l.snap.keys = make(map[string][]byte)
	l.snap.rem = make(map[string]struct{})
	l.snap.seq++
}

// 提交指定增量数据
func (l *LevelDB) commit(prefix []byte, keys ...[]byte) error {
	l.snap.mu.Lock()
	defer l.snap.mu.Unlock()

	b := l.kvBatch

	b.Lock()
	defer b.Unlock()

	rmkeys := make(map[string]struct{})
	snapkeys := make(map[string]struct{})
	for _, key := range keys {
		fullkey := append(prefix, key...)
		keyStr := l.snap.encodeKey(fullkey)
		//处理删除
		if _, ok := l.snap.rem[keyStr]; ok {
			b.Delete(fullkey)
			rmkeys[keyStr] = struct{}{}
		}

		//处理更新
		if v, ok := l.snap.keys[keyStr]; ok {
			b.Put(fullkey, v)
			snapkeys[keyStr] = struct{}{}
		}
	}

	if err := b.Commit(); err != nil {
		return err
	}

	// 清除增量
	for k := range rmkeys {
		delete(l.snap.rem, k)
	}

	for k := range snapkeys {
		delete(l.snap.keys, k)
	}

	return nil
}

// 提交指定增量数据
func (l *LevelDB) Commit(keys ...[]byte) error {
	return l.commit([]byte{0, 1}, keys...)
}

func (l *LevelDB) CommitPrefix(prefix []byte, keys ...[]byte) error {
	return l.commit(prefix, keys...)
}

// 提交包含前缀的增量数据
func (l *LevelDB) CommitPrekSnap(prekeys ...[]byte) error {
	l.snap.mu.Lock()
	defer l.snap.mu.Unlock()

	b := l.kvBatch

	b.Lock()
	defer b.Unlock()

	rmkeys := make(map[string]struct{})
	snapkeys := make(map[string]struct{})
	//处理删除
	for k := range l.snap.rem {
		key := l.snap.decodeKey(k)
		for _, prekey := range prekeys {
			prekey = append([]byte{0, 1}, prekey...)
			if bytes.HasPrefix(key, prekey) {
				b.Delete(key)
				rmkeys[k] = struct{}{}
			}
		}
	}

	//处理更新
	for k, v := range l.snap.keys {
		key := l.snap.decodeKey(k)
		for _, prekey := range prekeys {
			prekey = append([]byte{0, 1}, prekey...)
			if bytes.HasPrefix(key, prekey) {
				b.Put(key, v)
				snapkeys[k] = struct{}{}
			}
		}
	}

	if err := b.Commit(); err != nil {
		return err
	}

	// 清除增量
	for k := range rmkeys {
		delete(l.snap.rem, k)
	}

	for k := range snapkeys {
		delete(l.snap.keys, k)
	}

	return nil
}

func (l *LevelDB) StoreSnap(pairs []KVPair) error {
	l.snap.mu.Lock()
	defer l.snap.mu.Unlock()

	b := l.kvBatch

	b.Lock()
	defer b.Unlock()

	//处理删除
	for k := range l.snap.rem {
		b.Delete(l.snap.decodeKey(k))
	}

	//处理更新
	for k, v := range l.snap.keys {
		b.Put(l.snap.decodeKey(k), v)
	}

	//处理传入的待存储的kv
	for _, v := range pairs {
		tk := l.encodeKVKey(v.Key)
		b.Put(tk, v.Value)
	}

	if err := b.Commit(); err != nil {
		return err
	}

	return nil
}

// 获取快照序号
func (l *LevelDB) GetSeq() uint64 {
	return l.snap.seq
}

// 给快照新建一个迭代器
func (l *LeveldbSnap) NewIterator() *Iterator {
	return &Iterator{l.db.NewIterator(nil, nil)}
}

// 存储
func (l *LeveldbSnap) Put(k, v []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	//先查询是否标记删除，删除了则取消删除标记
	if l.isTagRem(k) {
		l.cancelTagRem(k)
	}

	l.put(k, v)
	return nil
}

// 查询
func (l *LeveldbSnap) Get(k []byte) (value []byte, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	//先查询是否标记删除，删除了则直接返回nil
	if l.isTagRem(k) {
		return nil, nil
	}

	//查询缓存
	if v, ok := l.get(k); ok {
		return v, nil
	}

	//查询db
	return l.db.Get(k, nil)
}

// 批量查询key
func (l *LeveldbSnap) MGet(ks ...[]byte) ([][]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	it := l.NewIterator()
	values := make([][]byte, len(ks))
	for i := range ks {
		//先查询是否标记删除，删除了则直接返回nil
		if l.isTagRem(ks[i]) {
			values[i] = nil
			continue
		}

		//查询缓存
		if v, ok := l.get(ks[i]); ok {
			values[i] = v
			continue
		}

		//查询db
		value := it.Find(ks[i])
		nv := make([]byte, len(value))
		copy(nv, value)
		values[i] = nv
	}

	return values, nil
}

// 批量删除kv
func (l *LeveldbSnap) Del(ks ...[]byte) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.del(ks...)
}

// 批量删除kv
func (l *LeveldbSnap) del(ks ...[]byte) {
	for _, k := range ks {
		kStr := l.encodeKey(k)
		//标记k删除
		l.rem[kStr] = struct{}{}

		//查询是否存在k的记录
		if _, ok := l.keys[kStr]; ok {
			delete(l.keys, kStr)
		}
	}
}

// key是否标记删除
func (l *LeveldbSnap) isTagRem(k []byte) bool {
	_, ok := l.rem[l.encodeKey(k)]
	return ok
}

// key取消删除标记
func (l *LeveldbSnap) cancelTagRem(k []byte) {
	delete(l.rem, l.encodeKey(k))
}

// 查询缓存
func (l *LeveldbSnap) get(k []byte) ([]byte, bool) {
	v, ok := l.keys[l.encodeKey(k)]
	return v, ok
}

// 存储到缓存
func (l *LeveldbSnap) put(k, v []byte) {
	l.keys[l.encodeKey(k)] = v
}

// key转string
func (l *LeveldbSnap) encodeKey(k []byte) string {
	return utils.Bytes2string(k)
}

// key转[]byte
func (l *LeveldbSnap) decodeKey(k string) []byte {
	return []byte(k)
}

// 给snap创建范围迭代器
func (l *LeveldbSnap) RangeLimitIterator(min, max []byte, rangeType uint8, offset, count int) *RangeLimitIterator {
	return NewRangeLimitIterator(l.NewIterator(), &Range{min, max, rangeType}, &Limit{offset, count})
}

// 删除start:stop范围的所有kv
func (l *LeveldbSnap) DelInRange(start, stop []byte) {
	//在db中迭代start:stop范围中的所有key，并保存到delKeys集合中
	delKeys := make(map[string]struct{})
	//it := l.RangeLimitIterator(start, stop, RangeROpen, 0, -1)
	//for ; it.Valid(); it.Next() {
	//	nk := make([]byte, len(it.it.Key()))
	//	copy(nk, it.it.Key())
	//	delKeys[l.encodeKey(nk)] = struct{}{}
	//	count++
	//}

	it2 := l.db.NewIterator(&util.Range{Start: start, Limit: stop}, nil)
	for it2.Next() {
		nk := make([]byte, len(it2.Key()))
		copy(nk, it2.Key())
		delKeys[l.encodeKey(nk)] = struct{}{}
	}
	it2.Release()
	it2.Error()

	l.mu.Lock()
	defer l.mu.Unlock()

	//查询增量中的所有key，因为增量中的key可能还没有存入db中
	for k, _ := range l.keys {
		if l.checkKeyInRange(l.decodeKey(k), start, stop) {
			//增量存在删除范围内，记录到delKeys集合中
			delKeys[k] = struct{}{}
		}
	}

	delKeys1 := make([][]byte, len(delKeys))
	i := 0
	for k := range delKeys {
		delKeys1[i] = l.decodeKey(k)
		i++
	}

	//在增量中处理待删除的keys数组
	l.del(delKeys1...)
}

// 删除start:stop范围的所有kv
func (l *LeveldbSnap) DelInKey(key []byte) {
}

// 查询start:stop范围的所有kv
func (l *LeveldbSnap) GetInRange(start, stop []byte, rangeType uint8) map[string][]byte {
	//在db中迭代start:stop范围中的所有key，并保存到kvs集合中
	kvs := make(map[string][]byte)
	//it := l.RangeLimitIterator(start, stop, rangeType, 0, -1)
	//for ; it.Valid(); it.Next() {
	//	nk := make([]byte, len(it.it.Key()))
	//	copy(nk, it.it.Key())

	//	nv := make([]byte, len(it.it.Value()))
	//	copy(nv, it.it.Value())

	//	kvs[l.encodeKey(nk)] = nv
	//}

	it2 := l.db.NewIterator(&util.Range{Start: start, Limit: stop}, nil)
	for it2.Next() {
		nk := make([]byte, len(it2.Key()))
		copy(nk, it2.Key())

		nv := make([]byte, len(it2.Value()))
		copy(nv, it2.Value())

		kvs[l.encodeKey(nk)] = nv
	}

	l.mu.RLock()
	defer l.mu.RUnlock()
	//遍历增量中的删除标记，找到范围内的，执行删除对应db key操作
	//因为db里的数据可能已经在增量中被标记删除了
	for k, _ := range l.rem {
		if l.checkKeyInRange(l.decodeKey(k), start, stop) {
			delete(kvs, k)
		}
	}

	//查询增量中的所有key，因为增量中的key可能还没有存入db中
	for k, v := range l.keys {
		if l.checkKeyInRange(l.decodeKey(k), start, stop) {
			//增量存在查询范围内，记录到keys集合中
			kvs[k] = v
		}
	}

	return kvs
}

// 检查key是否存在某个范围中
func (l *LeveldbSnap) checkKeyInRange(key, start, stop []byte) bool {
	return bytes.Compare(key, start) >= 0 && bytes.Compare(key, stop) <= 0
}

func (l *LeveldbSnap) GetKeys() map[string][]byte {
	return l.keys
}

func (l *LeveldbSnap) GetRem() map[string]struct{} {
	return l.rem
}

func (l *LeveldbSnap) DecodeKey(k string) []byte {
	return l.decodeKey(k)
}
