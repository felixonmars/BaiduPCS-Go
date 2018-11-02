package pcscache

import (
	"sync"
)

type (
	IntFunc func(i int) int
)

func Memorize(f IntFunc) IntFunc {
	cache := sync.Map{}
	return func(i int) int {
		value, ok := cache.Load(i)
		if !ok {
			i2 := f(i)
			cache.Store(i, i2)
			return i2
		}
		return value.(int)
	}
}
