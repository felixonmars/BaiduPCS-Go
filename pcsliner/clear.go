package pcsliner

import (
	"github.com/peterh/liner"
	_ "unsafe" // allow go:linkname
)

//go:linkname eraseScreen github.com/iikira/BaiduPCS-Go/vendor/github.com/peterh/liner.(*State).eraseScreen
func eraseScreen(s *liner.State)

// ClearScreen 清空屏幕
func ClearScreen() {
	eraseScreen(nil)
}
