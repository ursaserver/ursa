package ursa

import (
	"slices"
	"testing"
)

func TestZeroValueLinkedList(t *testing.T) {
	// Checks if the zero value of linkedList is useable
	l := new(linkedList[int])
	if l.head != nil {
		t.Errorf("expected nil, found non nil")
	}
	// Add a node
	n := &node[int]{value: 1}
	l.addNode(n)
	if l.head.value != 1 {
		t.Errorf("expected %v, found %v", n.value, l.head.value)
	}
	l.removeNode(n)
	if l.head != nil {
		t.Errorf("expected linked list's head to be nil after node deletion, but node is %v", l.head)
	}
}

func TestLinkedListToFromSlice(t *testing.T) {
	type test struct {
		items []int
	}
	tests := []test{
		{items: []int{}},
		{items: []int{}},
		{items: []int{1, 2}},
		{items: []int{1, 2, 3}},
	}
	for _, test := range tests {
		l := createLinkedListFromSlice[int](test.items)
		foundItems := make([]int, 0)
		current := l.head
		for current != nil {
			foundItems = append(foundItems, current.value)
			current = current.next
		}
		slices.Reverse(foundItems)
		compareSlices(test.items, foundItems, t)
		s := createSliceFromLinkedList(l)
		slices.Reverse(s)
		compareSlices(test.items, s, t)
	}
}

// Helper function for testing
func compareSlices[T comparable](exp, got []T, t *testing.T) {
	expectedLen, gotLen := len(exp), len(got)
	if expectedLen != gotLen {
		t.Errorf("expected %v items found %v", expectedLen, gotLen)
	} else {
		mismatchPosition := -1
		for i := 0; i < expectedLen && mismatchPosition == -1; i++ {
			// Note that items in linked list are in LIFO order
			if exp[i] != got[i] {
				mismatchPosition = i
			}
		}
		if mismatchPosition != -1 {
			t.Errorf("expected %v found %v at position %v",
				exp[mismatchPosition],
				got[mismatchPosition],
				mismatchPosition)
		}
	}
}

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
		l := createLinkedListFromSlice[int](test.itemsToAdd)
		items := createSliceFromLinkedList[int](l)
		// Note that items in linked list are in LIFO order
		slices.Reverse(items) // Note that
		compareSlices[int](test.itemsToAdd, items, t)
	}
}

func TesLinkedListNodeDeletion(t *testing.T) {
	type test struct {
		itemsToAdd []int
	}
	tests := []test{
		{itemsToAdd: []int{}},
		{itemsToAdd: []int{88}},
		{itemsToAdd: []int{1, 5, 10, 15, 20, 25, 30}},
	}
	for _, test := range tests {
		l := createLinkedListFromSlice[int](test.itemsToAdd)
		items := make([]int, 0)
		l.traverse(func(n *node[int]) {
			items = append(items, n.value)
			l.removeNode(n)
		})
		slices.Reverse(items)
		compareSlices[int](test.itemsToAdd, items, t)
		// We also expect the linked list to now point to nil since all the nodes are deleted
		if l.head != nil {
			t.Error("expected linked list to point to nil but got", l.head)
		}
	}
}
