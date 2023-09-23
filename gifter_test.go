package ursa

import (
	"testing"
	"time"
)

func TestLinkedListNodeInsertion(t *testing.T) {
	type test struct {
		itemsToAdd []int
	}
	tests := []test{
		{itemsToAdd: []int{}},
		{itemsToAdd: []int{88}},
		{itemsToAdd: []int{55, 12, 38, 34, 30}},
	}
	for _, test := range tests {
		l := linkedList[int]{}
		for _, item := range test.itemsToAdd {
			n := &node[int]{value: item}
			l.addNode(n)
		}
		items := make([]int, 0)
		traverseLinkedList[int](l, func(n *node[int]) {
			items = append(items, n.value)
		})
		expectedLen, gotLen := len(test.itemsToAdd), len(items)
		if expectedLen != gotLen {
			t.Errorf("expected %v items found %v", expectedLen, gotLen)
		} else {
			mismatchPosition := -1
			for i := 0; i < expectedLen && mismatchPosition == -1; i++ {
				if items[i] != test.itemsToAdd[i] {
					mismatchPosition = i
				}
			}
			if mismatchPosition != -1 {
				t.Errorf("expected %v found %v at position %v",
					test.itemsToAdd[mismatchPosition],
					items[mismatchPosition],
					mismatchPosition)
			}
		}
	}
}

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
