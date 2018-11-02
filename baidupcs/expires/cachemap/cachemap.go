package cachemap

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires"
	"sync"
)

type (
	CacheMap struct {
		cachePool sync.Map
	}
)

func (cm *CacheMap) LazyInitCachePoolOp(op string) *sync.Map {
	cm.ClearInvalidate()
	cacheItf, ok := cm.cachePool.Load(op)
	if !ok {
		cache := &sync.Map{}
		cm.cachePool.Store(op, cache)
		return cache
	}
	return cacheItf.(*sync.Map)
}

func (cm *CacheMap) ClearInvalidate() {
	cm.cachePool.Range(func(_, cacheItf interface{}) bool {
		cache := cacheItf.(*sync.Map)
		cache.Range(func(key, validateItf interface{}) bool {
			expire := validateItf.(expires.Expires)
			if expire.IsExpires() {
				cache.Delete(key)
			}
			return true
		})
		return true
	})
}
