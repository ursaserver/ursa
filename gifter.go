package ursa

import (
	"sync"
	"time"
)

// A gifter daemon that gifts tokens to buckets every some interval
type gifter struct {
	buckets   *linkedList[*bucket] // Buckets is a linked list of nodes. Each node holds *bucket.
	rate      rate
	isRunning bool
	ticker    time.Ticker
	server    *server
	sync.RWMutex
}

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
}

// Returns the duration at which it it needs to tick. This ticking duration is
// used mostly by the gifter to determine when to gift a token.
func tickOnceEvery(r rate) time.Duration {
	noOfTickingsPerSecond := float64(r.capacity) / float64(r.sec)
	ticksOnceEveryXSeconds := 1 / noOfTickingsPerSecond
	return time.Duration(ticksOnceEveryXSeconds * float64(time.Second))
}

func (g *gifter) start() {
	g.Lock()
	g.isRunning = true
	g.Unlock()
	ticker := time.NewTicker(tickOnceEvery(g.rate))
	// tickingCh := time.Tick(tickOnceEvery(g.rate))
	go func() {
		for g.isRunning {
			<-ticker.C  // Block until a tick is received
			go g.gift() // We run gift in a goroutine so that we get to next iteration of gift on time
		}
	}()
}

func (g *gifter) gift() {
	// Goes through each node in the buckets linked list and gifts a token to
	// each non-stale bucket that isn't full. It also deletes the node containing
	// buckets that are stale
	g.RLock()
	defer g.RUnlock()

	// It should be safe to read to server's fields that are read only
	staleDuration := g.server.bucketsStaleAfter
	traverseLinkedList[*bucket](*g.buckets, func(n *node[*bucket]) {
		bucket := n.value
		bucket.Lock()
		if bucket.tokens < bucket.rate.capacity {
			bucket.tokens++
		} else {
			// If the bucket is full remove the node containing bucket from
			// gifters linked list chain if the stale time has exceeded
			if time.Now().After(bucket.lastAccessed.Add(staleDuration)) {
				g.buckets.removeNode(n)
			}
		}
		bucket.Unlock()
	})

}

func (g *gifter) stop() {
	g.Lock()
	g.isRunning = false
	g.Unlock()
}

func (g *gifter) resume() {
	g.Lock()
	g.isRunning = true
	g.Unlock()
}

// Add a bucket to the linked list chain of gifters' buckets
func (g *gifter) addBucket(b *bucket) {
	g.RLock()
	if g.buckets == nil {
		g.RUnlock()
		l := &linkedList[*bucket]{}
		g.Lock()
		g.buckets = l
		g.Unlock()
	}
	g.buckets.addNode(&node[*bucket]{value: b})
}

// Adds a node to the beginning of the linked list chain in O(1) time
func (l *linkedList[any]) addNode(n *node[any]) {
	if l.head == nil {
		l.head = n
	}
	n.next = l.head
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
		return n.next
	}
	// l has more than 1 nodes
	// If n is the last node
	if n.next == nil {
		n.prev = nil
		return n.prev
	}
	// l has more than 2 nodes
	// Since n isn't first or the last, n has both prev and next nodes
	n.prev.next = n.next
	return n.next
}

// Traverses linked list by calling the traverse function on each node
// It is save for traverse function deletes the node by calling removeNode
// method on linked list, the traverse should proceed as usual.
func traverseLinkedList[T any](l linkedList[T], traverse func(*node[T])) {
	current := l.head
	for current != nil {
		// Note that we don't pass current.value which would be of type T
		// and instead pass *node[T] to allow users to call something like
		// removeNode method that takes a *node[T] as argument
		traverse(current)
		current = current.next
	}
}
