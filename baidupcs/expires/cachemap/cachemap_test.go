package cachemap_test

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires"
	"github.com/iikira/BaiduPCS-Go/baidupcs/expires/cachemap"
	"testing"
	"time"
)

func TestCacheMapDataExpires(t *testing.T) {
	cm := cachemap.CacheOpMap{}
	cache := cm.LazyInitCachePoolOp("op")
	cache.Store("key_1", expires.NewDataExpires("value_1", 1*time.Second))

	time.Sleep(1 * time.Second)
	data, ok := cache.Load("key_1")
	if !ok {
		t.FailNow()
	}
	fmt.Printf("data: %s\n", data.Data())
}

func TestCacheOperation(t *testing.T) {
	cm := cachemap.CacheOpMap{}
	data := cm.CacheOperation("op", "key_1", func() expires.DataExpires {
		return expires.NewDataExpires("value_1", 1*time.Second)
	})
	fmt.Printf("data: %s\n", data.Data())

	newData := cm.CacheOperation("op", "key_1", func() expires.DataExpires {
		return expires.NewDataExpires("value_3", 1*time.Second)
	})
	if data != newData {
		t.FailNow()
	}
	fmt.Printf("data: %s\n", data.Data())
}
