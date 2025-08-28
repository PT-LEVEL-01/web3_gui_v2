package cache

import (
	"sync"
	"sync/atomic"
)

type Cache struct {
	sync.Locker
	tm *sync.Map
	it int64
}

func CreateCache() *Cache {
	return &Cache{new(sync.Mutex), new(sync.Map), 0}
}

func (t *Cache) Exists(k any) bool {
	_, ok := t.tm.Load(k)
	return ok
}

func (t *Cache) Set(k, v any) bool {
	t.Lock()
	defer t.Unlock()

	_, ok := t.tm.LoadOrStore(k, v)
	if !ok {
		atomic.AddInt64(&t.it, 1)
	}

	return !ok
}

func (t *Cache) Get(k any) (any, bool) {
	return t.tm.Load(k)
}

func (t *Cache) GetM() *sync.Map {
	return t.tm
}

func (t *Cache) GetAndDel(k any) (any, bool) {
	t.Lock()
	defer t.Unlock()

	value, ok := t.tm.LoadAndDelete(k)
	if ok {
		atomic.AddInt64(&t.it, -1)
	}

	return value, ok
}

func (t *Cache) GetLen() int64 {
	return t.it
}
