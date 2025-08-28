package cache

import (
	"sync"
)

var TransCache *Cache

var Once_TransCache sync.Once

func init() {
	Once_TransCache.Do(func() {
		TransCache = CreateCache()
	})
}
