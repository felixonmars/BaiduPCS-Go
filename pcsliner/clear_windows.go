package pcsliner

import (
	"github.com/peterh/liner"
	_ "unsafe" // for go:linkname
)

var (
	defaultLinerState *liner.State
)

//go:linkname eraseScreen github.com/iikira/BaiduPCS-Go/vendor/github.com/peterh/liner.(*State).eraseScreen
func eraseScreen(s *liner.State)

// ClearScreen 清空屏幕
func (pl *PCSLiner) ClearScreen() {
	eraseScreen(pl.State)
}

// ClearScreen 清空屏幕
func ClearScreen() {
	if defaultLinerState == nil {
		defaultLinerState = liner.NewLiner()
	}
	defaultLinerState.ClearScreen()
}
