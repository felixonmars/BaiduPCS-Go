package pcsutil

import (
	"fmt"
	"testing"
	"time"
)

func TestWg(t *testing.T) {
	wg := NewWaitGroup(2)
	for i := 0; i < 60; i++ {
		wg.AddDelta()
		go func() {
			fmt.Println(i, wg.Parallel())
			time.Sleep(1e9)
			wg.Done()
		}()
	}
	wg.Wait()
}
