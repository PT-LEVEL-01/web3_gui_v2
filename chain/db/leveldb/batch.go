package leveldb

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

//批处理
type Batch struct {
	leveldb *leveldb.DB
	batch   *leveldb.Batch
	sync.Locker
}

//批处理设置kv
func (b *Batch) Put(key, value []byte) {
	b.batch.Put(key, value)
}

//批处理查询key
func (b *Batch) Load(key []byte) error {
	return b.batch.Load(key)
}

//批处理持久化提交
func (b *Batch) Commit() error {
	return b.leveldb.Write(b.batch, nil)
}

//批处理删除
func (b *Batch) Delete(key []byte) {
	b.batch.Delete(key)
}

func (b *Batch) Lock() {
	b.Locker.Lock()
}

func (b *Batch) Unlock() {
	b.batch.Reset()
	b.Locker.Unlock()
}

type dbBatchLocker struct {
	l      *sync.Mutex
	wrLock *sync.RWMutex
}

func (l *dbBatchLocker) Lock() {
	l.wrLock.RLock()
	l.l.Lock()
}

func (l *dbBatchLocker) Unlock() {
	l.l.Unlock()
	l.wrLock.RUnlock()
}
