package escaper_test

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs/escaper"
	"testing"
)

func TestEscape(t *testing.T) {
	fmt.Println(escaper.Escape(`asdfasdfasd[]a[\[][sdf\[d]`, []rune{'['}))
}
