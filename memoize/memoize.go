package memoize

import "sync"

// Memoize a unary function that returns a single value.
// The argument to the function should satisfy the comparable
// constraints. The return type  can be anything.
//
// When working with a recursive function, for example,
//
//	func fib(n){
//		if (n < 2) {return 0} else {return fib(n-1) + fib(n-2)}
//	}
//
//	 It is NOT enough to create a memoized function by merely saying
//
// memoizedFib := memoize.Unary(fib)
//
// This is because the recursive calls inside fib are still referring
// to fib not the memoized function.
// Workaround in such cases, is to write new definition of fib as follows:
//
//	var memoizedFib func (int) int
//	memoizedFib = Unary(func(n int) int {
//		if n < 2 {
//			return 1
//		}
//		return memoizedFib(n-1) + memoizedFib(n-2)
//	})
//
// This function returned by Unary is safe for concurrent use.
func Unary[K comparable, V any](fn func(K) V) func(K) V {
	cache := make(map[K]V)
	var mu sync.RWMutex
	return func(arg K) V {
		mu.RLock()
		if _, ok := cache[arg]; !ok {
			mu.RUnlock()
			result := fn(arg)
			mu.Lock()
			cache[arg] = result
			mu.Unlock()
			mu.RLock()
		}
		res := cache[arg]
		mu.RUnlock()
		return res
	}
}
