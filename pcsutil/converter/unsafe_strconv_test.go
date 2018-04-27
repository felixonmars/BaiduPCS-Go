// go test -test.bench=".*"
package converter

import (
	"testing"
)

var str = "asddsadfaalkdjsksajdfkashjkdfhashfliuhsadifhasifhaishdfiashdihaisdfhiuassfasdff"

func BenchmarkToBytes(b *testing.B) {
	for i := 0; i <= b.N; i++ {
		_ = ToBytes(str)
	}
}

func BenchmarkBytes(b *testing.B) {
	for i := 0; i <= b.N; i++ {
		_ = []byte(str)
	}
}
