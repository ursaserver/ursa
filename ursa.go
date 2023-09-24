// Ursa rate limiter is a http.Handler
package ursa

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ursaserver/ursa/memoize"
)

type reqSignature string
type reqPath string

type server struct {
	conf              Conf
	rateBys           []RateBy
	bucketsStaleAfter time.Duration
	boxes             map[reqSignature]*box
	gifters           map[gifterId]*gifter
	pathRate          func(reqPath) *rate
	sync.RWMutex
}

type bucketId string

type box struct {
	server  *server
	id      reqSignature // request signature
	buckets map[reqPath]*bucket
	sync.RWMutex
}

type bucket struct {
	tokens       int
	id           bucketId
	lastAccessed time.Time
	rate         *rate
	box          *box
	sync.Mutex
}

// Create a server based on provided configuration.
// Initializes gifters
func New(conf Conf) *server {
	s := &server{conf: conf}
	s.boxes = make(map[reqSignature]*box)
	s.gifters = make(map[gifterId]*gifter)
	s.bucketsStaleAfter = time.Duration(0)
	s.pathRate = memoize.Unary(func(r reqPath) *rate {
		// Note that memoization is possible since the configuration is not
		// changed once loaded.
		return rateForPath(r, conf)
	})
	for _, route := range conf.routes {
		rates := route.rate
		for _, r := range rates {
			gifterId := generateGifterId(r)
			// Check if the gifter with the id already exists
			s.RLock()
			_, ok := s.gifters[gifterId]
			s.RUnlock()
			if !ok {
				// Create a gifter
				g := new(gifter)
				s.Lock()
				s.gifters[gifterId] = g
				s.Unlock()
			}
		}
	}
	// Start gifters
	for _, g := range s.gifters {
		g.start()
	}
	// init reverse proxy
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO if the request is made to non rate limited path, forward to reverse
	// proxy immediately
	sig := findReqSignature(r, s.rateBys)
	// Find a box for given signature
	s.RLock()
	if _, ok := s.boxes[sig]; !ok {
		s.RUnlock()
		// create box with given signature
		s.Lock()
		b := box{id: sig, server: s}
		s.boxes[sig] = &b
		s.Unlock()
		s.RLock()
	}
	b := s.boxes[sig]
	path := findPath(r)
	b.RLock()
	if _, ok := b.buckets[path]; !ok {
		s.RUnlock()
		// create bucket,
		s.createBucket(path, b)
		s.RLock()
	}
	// While we're here, we can safely assume that the gifter isn't deleting
	// thus this bucket as it would require gifter to acquire a Write Lock
	// which can't be granted while there's still a reader.
	buck := b.buckets[path]
	b.RUnlock()

	buck.Lock()
	defer buck.Unlock()
	// We check if the no. of tokens is >= 1
	// Just before leaving, we set the last accessed time on the bucket
	buck.tokens--
	if buck.tokens < 0 {
		// TODO
		// Reject downstream & return
	}
	// TODO
	// Call HTTPServer of the underlying ReverseProxy
	buck.lastAccessed = time.Now()
	buck.Unlock()
}

// Create a bucket with given id inside the given box.
// Initializes various properties of the bucket like capacity, state time, etc.
// and then registers the bucket to the gifter to collect gift tokens.
func (s *server) createBucket(id reqPath, b *box) {
	b.Lock()
	rate := s.pathRate(id)
	acc := time.Now()
	tokens := rate.capacity
	newBucket := &bucket{
		id:           bucketId(id),
		tokens:       tokens,
		rate:         rate,
		lastAccessed: acc,
		box:          b,
		Mutex:        sync.Mutex{},
	}
	b.buckets[id] = newBucket
	b.Unlock()
	b.server.RLock()
	gifter, ok := b.server.gifters[generateGifterId(*rate)]
	if !ok {
		log.Fatalf("cannot find gifter for rate %v", *rate)
	}
	// add the bucket to appropriate gifter
	gifter.addBucket(newBucket)
}

func findReqSignature(req *http.Request, rateBys []RateBy) reqSignature {
	// Find if any of the header fields in RateBy are present.
	rateby := rateByIP // default
	key := ""
	for _, r := range rateBys {
		if req.Header.Get(string(r)); r != "" {
			rateby = r
			key = string(r)
			break
		}
	}
	if rateby == rateByIP {
		key = clientIpAddr(req)
	}
	return reqSignature(fmt.Sprintf("%v-%v", rateby, key))
}

func findPath(r *http.Request) reqPath {
	// TODO
	// decide what how to handle trailing, leading forward slashes
	return reqPath(r.URL.Path)
}
