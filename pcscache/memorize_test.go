package pcscache_test

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcscache"
	"testing"
)

func add(i int) int {
	return i + 1
}

func TestMemorize(t *testing.T) {
	cacheAdd := pcscache.Memorize(add)
	fmt.Println(cacheAdd(1))
	fmt.Println(cacheAdd(1))
	fmt.Println(cacheAdd(2))
}
