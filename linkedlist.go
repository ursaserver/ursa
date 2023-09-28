package ursa

// A doubly pointed node
type node[T any] struct {
	value T // A node holds value of type T
	next  *node[T]
	prev  *node[T]
}

// Linked list is a pointer to a doubly pointed node
// thus making a doubly linked list.
// It is safe to use zero value of this type.
type linkedList[T any] struct {
	head *node[T]
	tail *node[T]
}

// Adds a node to the beginning of the linked list chain in O(1) time
func (l *linkedList[any]) addNode(n *node[any]) {
	if l.head == nil {
		l.head = n
		l.tail = n
		return
	}
	n.next = l.head
	l.head.prev = n
	l.head = n
}

// Assumes that the linked list the given node n
// Removes the node from the linked list chain in O(1) time
// and returns the pointer to the node next to the node just returned
//
// It is safe to call this method to remove a node inside the traverse
// function if the traversal is made using traverseLinkedList function
func (l *linkedList[any]) removeNode(n *node[any]) *node[any] {
	// Given the precondition that n is a node in linkedlist l,
	// we assume that l is not empty

	// l has more than 0 nodes
	// If n is the first node
	if l.head == n {
		l.head = n.next
		// Since n could also be the last node, we need
		// do the following conditionally
		if n.next != nil {
			n.next.prev = nil // Note not a circular linked list, so first node's prev is nil
		}
		return n.next
	}
	// l has more than 1 nodes
	// If n is the last node and not the first node
	if n.next == nil {
		n.prev.next = nil // Note not a circular linked list, so last node's next is nil
		return n.prev
	}
	// l has more than 2 nodes
	// Since n isn't first or the last, n has both prev and next nodes
	// Set the next reference of the previous node
	n.prev.next = n.next
	// Set the previous reference of the next node
	n.next.prev = n.prev
	return n.next
}

// Traverses linked list by calling the traverse function on each node
// It is save for traverse function deletes the node by calling removeNode
// method on linked list, the traverse should proceed as usual.
func (l *linkedList[any]) traverse(f func(*node[any])) {
	current := l.head
	for current != nil {
		// Note that we don't pass current.value which would be of type T
		// and instead pass *node[T] to allow users to call something like
		// removeNode method that takes a *node[T] as argument
		f(current)
		current = current.next
	}
}

// Creates a linked list in the order opposite of the slice
func createLinkedListFromSlice[T any](items []T) linkedList[T] {
	l := linkedList[T]{}
	for _, item := range items {
		n := &node[T]{value: item}
		l.addNode(n)
	}
	return l
}

// Creates a slice in the same order as the order of nodes in linked list
func createSliceFromLinkedList[T any](l linkedList[T]) []T {
	items := make([]T, 0)
	l.traverse(func(n *node[T]) {
		items = append(items, n.value)
	})
	return items
}
