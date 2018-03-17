// Package cachepool []byte缓存池
package cachepool

import (
	"sync"
	"sync/atomic"
)

var (
	// CachePool []byte 缓存池
	CachePool = cachePool{
		cachepool: sync.Map{},
	}
)

type cachePool struct {
	lastID    int32
	cachepool sync.Map
}

func (cp *cachePool) Apply(size int) (id int32) {
	for {
		_, ok := cp.cachepool.Load(cp.lastID)
		atomic.AddInt32(&cp.lastID, 1)
		if ok {
			continue
		}
		break
	}

	cp.Set(cp.lastID, size)
	return cp.lastID
}

func (cp *cachePool) Existed(id int32) (existed bool) {
	_, existed = cp.cachepool.Load(id)
	return
}

func (cp *cachePool) Get(id int32) []byte {
	cache, _ := cp.cachepool.Load(id)
	return cache.([]byte)
}

func (cp *cachePool) Set(id int32, size int) []byte {
	cache := make([]byte, size)
	cp.cachepool.Store(id, cache)
	return cp.Get(id)
}

func (cp *cachePool) SetIfNotExist(id int32, size int) []byte {
	ok := cp.Existed(id)
	if !ok {
		cp.Set(id, size)
	}

	return cp.Get(id)
}

func (cp *cachePool) Delete(id int32) {
	cp.cachepool.Store(id, nil)
	cp.cachepool.Delete(id)
}

func (cp *cachePool) DeleteAll() {
	cp.cachepool.Range(func(k interface{}, _ interface{}) bool {
		cp.Delete(k.(int32))
		return true
	})
}

// Apply 申请缓存, 返回缓存id
func Apply(size int) (id int32) {
	return CachePool.Apply(size)
}

// Existed 通过缓存id检测是否存在缓存
func Existed(id int32) bool {
	return CachePool.Existed(id)
}

// Get 通过缓存id获取缓存[]byte
func Get(id int32) []byte {
	return CachePool.Get(id)
}

// Set 设置缓存, 通过给定的缓存id
func Set(id int32, size int) []byte {
	return CachePool.Set(id, size)
}

// SetIfNotExist 如果缓存不存在, 则设置缓存池
func SetIfNotExist(id int32, size int) []byte {
	return CachePool.SetIfNotExist(id, size)
}

// Delete 通过缓存id删除缓存
func Delete(id int32) {
	CachePool.Delete(id)
}

// DeleteAll 清空缓存池
func DeleteAll() {
	CachePool.DeleteAll()
}
