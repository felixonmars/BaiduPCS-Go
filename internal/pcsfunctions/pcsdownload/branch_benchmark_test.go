package pcsdownload_test

import (
	"github.com/iikira/BaiduPCS-Go/internal/pcsfunctions/pcsdownload"
	"testing"
)

func BenchmarkIsSkipMd5Checksum(b *testing.B) {
	md5Str := "8091310d5f4719769995a74d8e6d0530"
	for i := 0; i < b.N; i++ {
		pcsdownload.IsSkipMd5Checksum(120, md5Str)
	}
}
