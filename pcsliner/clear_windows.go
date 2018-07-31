package pcsliner

import (
	"github.com/peterh/liner"
	_ "unsafe" // for go:linkname
)

//go:linkname eraseScreen github.com/iikira/BaiduPCS-Go/vendor/github.com/peterh/liner.(*State).eraseScreen
func eraseScreen(s *liner.State)

// ClearScreen 清空屏幕
func (pl *PCSLiner) ClearScreen() {
	eraseScreen(pl.State)
}

// ClearScreen 清空屏幕
func ClearScreen() {
	liner := NewLiner()
	liner.ClearScreen()
	liner.Close()
}
