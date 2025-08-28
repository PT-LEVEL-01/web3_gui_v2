package leveldb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const defaultFilterBits int = 10

type LevelDB struct {
	path string      //存储路径
	db   *leveldb.DB //db库

	indexVarBuf []byte //db库，默认0
	kvBatch     *Batch //kv结构批处理
	hashBatch   *Batch //hash结构批处理

	wLock sync.RWMutex

	snap *LeveldbSnap //db增量快照
}

// 创建leveldb
func CreateLevelDB(path string) (*LevelDB, error) {
	opts := newOptions()
	db, err := leveldb.OpenFile(path, opts)
	if err != nil {
		return nil, err
	}

	lldb := &LevelDB{
		path: path,
		db:   db,
	}
	lldb.kvBatch = lldb.NewBatch()
	lldb.hashBatch = lldb.NewBatch()
	lldb.snap = lldb.SetSnap()
	buf := make([]byte, 10)
	n := binary.PutUvarint(buf, uint64(0))
	lldb.indexVarBuf = buf[0:n]
	return lldb, nil
}

// db配置选项
func newOptions() *opt.Options {
	opts := new(opt.Options)
	opts.ErrorIfMissing = false

	opts.Filter = filter.NewBloomFilter(defaultFilterBits)

	//opts.BlockSize = 1 * MB
	opts.BlockSize = 16 * KB

	opts.BlockCacheCapacity = 64 * MB
	opts.WriteBuffer = 64 * MB
	opts.OpenFilesCacheCapacity = 1024

	opts.CompactionTableSize = 32 * 1024 * 1024
	opts.WriteL0SlowdownTrigger = 16
	opts.WriteL0PauseTrigger = 64
	return opts
}

// 获取leveldb
func (l *LevelDB) GetDB() *LevelDB {
	return l
}

// 获取leveldb的kv批处理
func (l *LevelDB) GetKvBatch() *Batch {
	return l.kvBatch
}

// 新建一个查询迭代器
func (l *LevelDB) NewIterator() *Iterator {
	return NewIterator(l)
}

// 新建一个批处理
func (l *LevelDB) NewBatch() *Batch {
	b := new(Batch)
	b.leveldb = l.db
	b.batch = new(leveldb.Batch)
	b.Locker = &dbBatchLocker{l: &sync.Mutex{}, wrLock: &l.wLock}
	return b
}

func (l *LevelDB) Close() {
	l.db.Close()
}

// 新建一个范围查询迭代器
func (l *LevelDB) RangeLimitIterator(min, max []byte, rangeType uint8, offset, count int) *RangeLimitIterator {
	return NewRangeLimitIterator(l.NewIterator(), &Range{min, max, rangeType}, &Limit{offset, count})
}

// 保存数据
func (l *LevelDB) Save(key []byte, bs *[]byte) error {
	var err error

	if bs == nil {
		err = l.Set(key, nil)
	} else {
		err = l.Set(key, *bs)
	}
	return err
}

// 查询一个key
func (l *LevelDB) Find(key []byte) (*[]byte, error) {
	value, err := l.Get(key)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// 检查一个hash是否存在
func (l *LevelDB) CheckHashExist(hash []byte) (bool, error) {
	n, err := l.Exists(hash)
	if err != nil {
		return false, err
	}
	if n <= 0 {
		return false, nil
	}
	return true, nil
}

// 删除一个key
func (l *LevelDB) Remove(id []byte) error {
	_, err := l.Del(id)
	return err
}

// 检查一个key的索引库(类似redis的select库)
func (l *LevelDB) checkKeyIndex(buf []byte) (int, error) {
	if len(buf) < len(l.indexVarBuf) {
		return 0, fmt.Errorf("key is too small")
	} else if !bytes.Equal(l.indexVarBuf, buf[0:len(l.indexVarBuf)]) {
		return 0, fmt.Errorf("invalid db index")
	}

	return len(l.indexVarBuf), nil
}

// 新建范围迭代器
func (l *LevelDB) NewIteratorWithOption(slice *util.Range, ro *opt.ReadOptions) *Iterator {
	s := &util.Range{Start: l.encodeKVKey(slice.Start), Limit: l.encodeKVKey(slice.Limit)}
	return NewIteratorWithOption(l, s, ro)
}

func (l *LevelDB) GetEncodeKVKey(key []byte) []byte {
	return l.encodeKVKey(key)
}

// 获取key的数据结构类型
func (l *LevelDB) GetKeyType(key []byte) (byte, error) {
	return l.getKeyType(key)
}

// 获取key的数据结构类型
func (l *LevelDB) getKeyType(buf []byte) (byte, error) {
	if len(buf) < 2 {
		return 0, fmt.Errorf("key is too small")
	}

	return buf[1], nil
}

// 获取leveldb snap
func (l *LevelDB) GetSnap() *LeveldbSnap {
	return l.snap
}

// 包装LevelDB前缀查询
func (l *LevelDB) WrapLevelDBPrekeyRange(preKey []byte) map[string][]byte {
	endKey := make([]byte, len(preKey))
	copy(endKey, preKey)
	for i := len(endKey) - 1; i >= 0; i-- {
		endKey[i] += 1
		if endKey[i] != 0 {
			break
		}
	}

	s := l.GetEncodeKVKey(preKey)
	e := l.GetEncodeKVKey(endKey)

	kvs := l.snap.GetInRange(s, e, RangeROpen)

	return kvs
}

// LevelDB原生前缀删除,不走增量处理逻辑
func (l *LevelDB) LevelDBDelInRange(prekey []byte) (*big.Int, error) {
	count := new(big.Int)
	n := 20000 //批处理key个数
	it := l.db.NewIterator(util.BytesPrefix(prekey), nil)
	batch := &leveldb.Batch{}
	for it.Next() {
		batch.Delete(it.Key())
		if batch.Len() >= n {
			if err := l.db.Write(batch, nil); err != nil {
				return nil, err
			}
			count.Add(count, big.NewInt(int64(batch.Len())))
			batch.Reset()
		}
	}
	if batch.Len() > 0 {
		if err := l.db.Write(batch, nil); err != nil {
			return nil, err
		}
		count.Add(count, big.NewInt(int64(batch.Len())))
		batch.Reset()
	}
	it.Release()
	if err := it.Error(); err != nil {
		return nil, err
	}

	return count, nil
}
