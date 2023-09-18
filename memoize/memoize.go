package memoize

import "sync"

// Memoize a unary function that returns a single value.
// Note that the argument to the function should satisfy the comparable
// contraints. The return type  can be anything.
func Unary[K comparable, V any](fn func(K) V) func(K) V {
	cache := make(map[K]V)
	var mu sync.RWMutex
	return func(arg K) V {
		mu.RLock()
		if _, ok := cache[arg]; !ok {
			mu.RUnlock()
			result := fn(arg)
			mu.WLock()
			cache[arg] = result
			mu.Unlock()
			mu.RLock()
		}
		res := cache[arg]
		mu.RUnlock()
		return res
	}
}
