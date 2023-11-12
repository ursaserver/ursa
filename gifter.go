package ursa

import (
	"fmt"
	"math"
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

func (g *gifter) String() string {
	return fmt.Sprintf("gifter %v: %v", g.id, g.server)
}

// Returns the duration at which it it needs to tick. This ticking duration is
// used mostly by the gifter to determine when to gift a token.
func tickOnceEvery(r rate) time.Duration {
	// noOfTickingsPerSecond := float64(r.Capacity) / float64(r.RefillDurationInSec)
	// ticksOnceEveryXSeconds := 1 / noOfTickingsPerSecond
	// return time.Duration(ticksOnceEveryXSeconds * float64(time.Second))
	return time.Duration(r.RefillDurationInSec * duration(time.Second))
}

func (g *gifter) start() {
	g.Lock()
	g.isRunning = true
	g.Unlock()
	ticker := time.NewTicker(tickOnceEvery(g.rate))
	g.ticker = *ticker
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
		if bucket.tokens < bucket.rate.Capacity {
			// Gifter giftes rate.Capacity tokens at max
			bucket.tokens = min(bucket.tokens+bucket.rate.Capacity, bucket.rate.Capacity)
			bucket.lastGifted = time.Now()
			g.server.logger.Info("gifting tokens", "bucket", bucket.id, "tokens", bucket.tokens)
		} else {
			// If the bucket is full remove the node containing bucket from
			// gifters linked list chain if the stale time has exceeded
			if time.Now().After(bucket.lastAccessed.Add(staleDuration)) {
				g.server.logger.Info("removing stale bucket", "bucket", bucket.id)
				// delete the bucket from the box
				g.buckets.removeNode(n)
				g.server.logger.Info("removed bucket from gifters chain", "bucket", bucket.id)
				n.value.box.Lock()
				delete(n.value.box.buckets, bucket.id)
				g.server.logger.Info("removed bucket from the boxes buckets map", "bucket", bucket.id)
				n.value.box.Unlock()
			}
		}
		bucket.Unlock()
	})
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
	return gifterId(fmt.Sprintf("%v-%v", r.Capacity, r.RefillDurationInSec))
}

// Find in seconds the seconds to wait before you'll have > 0 tokens
func secondsBeforeSuccess(currentTime time.Time, lastGiftedTime time.Time, r *rate, tokens int) int {
	// If at least one token present currently, no need to wait
	if tokens > 0 {
		return 0
	}
	tokensNeeded := float64(-1*tokens + 1)
	noOfGiftingsNeeded := math.Ceil(tokensNeeded / float64(r.Capacity))
	successAt := lastGiftedTime.Add(time.Duration(noOfGiftingsNeeded * float64(tickOnceEvery(*r))))
	return int(successAt.Sub(currentTime).Seconds())
}
