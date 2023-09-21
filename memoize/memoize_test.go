package memoize

import (
	"testing"
	"time"
)

func TestUnary(t *testing.T) {
	callCounts := 0
	adder := func(numbers [2]int) int {
		callCounts++
		return numbers[0] + numbers[1]
	}
	cachedAdder := Unary(adder)
	tests := []struct {
		fn                 func() int
		expectedResult     int
		expectedCallCounts int
	}{
		{fn: func() int { return cachedAdder([2]int{1, 2}) }, expectedResult: 3, expectedCallCounts: 1},
		{fn: func() int { return cachedAdder([2]int{1, 2}) }, expectedResult: 3, expectedCallCounts: 1},

		{fn: func() int { return cachedAdder([2]int{4, 5}) }, expectedResult: 9, expectedCallCounts: 2},
		{fn: func() int { return cachedAdder([2]int{4, 5}) }, expectedResult: 9, expectedCallCounts: 2},

		{fn: func() int { return cachedAdder([2]int{6, 7}) }, expectedResult: 13, expectedCallCounts: 3},
		{fn: func() int { return cachedAdder([2]int{6, 7}) }, expectedResult: 13, expectedCallCounts: 3},
	}
	for _, test := range tests {
		got := test.fn()
		if got != test.expectedResult {
			t.Errorf("expected %v, got %v", got, test.expectedResult)
		}
		if callCounts != test.expectedCallCounts {
			t.Errorf("expected call counts: %v, is %v", test.expectedCallCounts, callCounts)
		}
	}
}

func TestUnaryFib(t *testing.T) {
	// A naive fib function would usually take a long time to complete.
	// Memoization should fix the slog.
	var memoizedFib func(int) int
	memoizedFib = Unary(func(n int) int {
		if n < 2 {
			return 1
		}
		return memoizedFib(n-1) + memoizedFib(n-2)
	})

	timeToRunFib := 1 * time.Second
	abort := time.After(timeToRunFib)

	n := 0
	finish := false

	for !finish {
		select {
		case <-abort:
			finish = true
		default:
			memoizedFib(n)
			n++
		}
	}
	if n < 1000 {
		t.Errorf("produced %v values in fib sequence after %v seconds. expected more than 1000", n, timeToRunFib)
	}
}
