package ursa

import (
	"fmt"
	"sync"
	"time"
)

type gifterId string

// A gifter daemon that gifts tokens to buckets every some interval
type gifter struct {
	id        gifterId
	buckets   *linkedList[*bucket] // Buckets is a linked list of nodes. Each node holds *bucket.
	rate      rate
	isRunning bool
	ticker    time.Ticker
	server    *server
	sync.RWMutex
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
	g.buckets.traverse(func(n *node[*bucket]) {
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

// Generate gifter id based on rate
func generateGifterId(r rate) gifterId {
	return gifterId(fmt.Sprintf("%v-%v", r.capacity, r.sec))
}
