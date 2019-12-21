package cachemap

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires"
)

func (cm *CacheOpMap) CacheOperation(op string, key interface{}, opFunc func() expires.DataExpires) (data expires.DataExpires) {
	var (
		cache = cm.LazyInitCachePoolOp(op)
		ok    bool
	)
	data, ok = cache.Load(key)
	if !ok {
		data = opFunc()
		if data != nil {
			cache.Store(key, data)
		}
		return
	}

	return
}

func (cm *CacheOpMap) CacheOperationWithError(op string, key interface{}, opFunc func() (expires.DataExpires, error)) (data expires.DataExpires, err error) {
	var (
		cache = cm.LazyInitCachePoolOp(op)
		ok    bool
	)
	data, ok = cache.Load(key)
	if !ok {
		data, err = opFunc()
		if err != nil {
			return
		}
		cache.Store(key, data)
	}

	return
}
