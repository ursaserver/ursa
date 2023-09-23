package ursa

import (
	"sync"
	"time"
)

type gifter struct {
	buckets   *node // Buckets is a linked list of nodes
	rate      rate
	isRunning bool
	ticker    time.Ticker
	sync.RWMutex
}

type node struct {
	value *bucket
	next  *node
	prev  *node
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
	n := g.buckets
	for n != nil {
		// TODO
		// Handle node
		// Go to next node
		n = n.next
	}
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
	currentFirst := g.buckets
	newNode := node{value: b, next: currentFirst}
	currentFirst.prev = &newNode
	g.buckets = &newNode
}
