package ursa

import (
	"testing"
	"time"
)

func TestTickOnceEvery(t *testing.T) {
	type test struct {
		r rate
		t time.Duration
	}
	tests := []test{
		{r: rate{60, Minute}, t: time.Second * 1},
		{r: rate{30, Minute}, t: time.Second * 2},
		{r: rate{60, Hour}, t: time.Minute},
		{r: rate{30, Hour}, t: time.Duration(float64(time.Minute) * 2)},
		{r: rate{30, Day}, t: time.Duration(float64(time.Hour) * 24 / 30)},
	}
	for _, test := range tests {
		expected := test.t
		got := tickOnceEvery(test.r)
		if expected != got {
			t.Errorf("expected tick interval %v got %v for rate %vreqs/%vsecs", expected, got, test.r.capacity, test.r.sec)
		}
	}
}
