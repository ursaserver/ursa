package main

import "testing"

func TestFib(t *testing.T) {
	type test struct {
		n      int
		expect int
	}
	tests := []test{
		{n: 0, expect: 1},
		{n: 1, expect: 1},
		{n: 2, expect: 2},
		{n: 3, expect: 3},
		{n: 4, expect: 5},
		{n: 5, expect: 8},
		{n: 7, expect: 21},
	}
	for _, test := range tests {
		got := fib(test.n)
		if want := test.expect; got != want {
			t.Errorf("expected %d got %d", got, want)
		}
	}
}
