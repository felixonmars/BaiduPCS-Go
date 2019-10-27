package downloader_test

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
	"testing"
)

func TestRangeListGen1(t *testing.T) {
	gen := downloader.NewRangeListGen1(1024, 10)
	_, genF := gen.GenFunc()

	i := 0
	for r := genF(); r != nil; r = genF() {
		fmt.Printf("%d: %s\n", i, r)
		i++
	}
}
