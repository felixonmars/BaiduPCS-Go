// +build !windows

package pcsliner

import (
	"fmt"
)

// ClearScreen 清空屏幕
func (pl *PCSLiner) ClearScreen() {
	fmt.Print("\x1b[H\x1b[2J")
}
