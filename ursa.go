// Ursa rate limiter is a http.Handler
package ursa

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
	"time"

	"github.com/ursaserver/ursa/memoize"
)

type (
	reqSignature string
	reqPath      string
)

type server struct {
	id                string
	conf              *Conf
	rateBys           []*rateBy
	bucketsStaleAfter time.Duration
	boxes             map[reqSignature]*box
	gifters           map[gifterId]*gifter
	routeForPath      func(reqPath) *Route
	proxy             *httputil.ReverseProxy
	mu                sync.RWMutex
}

func (s *server) String() string {
	return fmt.Sprintf("server %v", s.id)
}

type bucketId string

type box struct {
	server  *server
	id      reqSignature // request signature
	rateBy  *rateBy
	buckets map[bucketId]*bucket
	sync.RWMutex
}

func (b *box) String() string {
	return fmt.Sprintf("box %v: %s", b.id, b.server)
}

type bucket struct {
	id           bucketId
	tokens       int
	lastAccessed time.Time
	lastGifted   time.Time
	rate         *rate
	box          *box
	sync.Mutex
}

func (b *bucket) String() string {
	return fmt.Sprintf("bucket %v: %s", b.id, b.box)
}

// Create a server based on provided configuration.
func New(conf Conf) *server {
	// Validates configuration. The validation func takes care of exist in case of error.
	ValidateConf(conf, true)
	serverId := fmt.Sprintf("%v", rand.Float64())
	s := &server{conf: &conf, id: serverId}
	s.boxes = make(map[reqSignature]*box)
	s.gifters = make(map[gifterId]*gifter)
	s.bucketsStaleAfter = time.Duration(0)
	s.proxy = httputil.NewSingleHostReverseProxy(conf.Upstream)
	s.routeForPath = memoize.Unary(func(r reqPath) *Route {
		// Note that memoization is possible since the configuration is not
		// changed once loaded.
		return routeForPath(r, &conf)
	})
	allRateBys := make(map[*rateBy]bool)
	for _, route := range conf.Routes {
		rates := route.Rates
		for rateBy, r := range rates {
			allRateBys[rateBy] = true
			gifterId := generateGifterId(r)
			// Check if the gifter with the id already exists
			s.mu.RLock()
			_, ok := s.gifters[gifterId]
			s.mu.RUnlock()
			if !ok {
				// Create a gifter
				g := &gifter{
					rate:    r,
					server:  s,
					id:      gifterId,
					buckets: new(linkedList[*bucket]),
				}
				s.mu.Lock()
				s.gifters[gifterId] = g
				s.mu.Unlock()
			}
		}
	}
	s.rateBys = make([]*rateBy, 0)
	for k := range allRateBys {
		s.rateBys = append(s.rateBys, k)
	}
	// Start gifters
	for _, g := range s.gifters {
		g.start()
	}
	// init reverse proxy
	return s
}

// Checks if the provided configuration is valid.
// If exitOnErr is true, prints all the error messages and exists the process
// by calling os.Exit(1).
// If exitOnErr is false then returns a boolean if the configuration is valid.
func ValidateConf(conf Conf, exitOnErr bool) bool {
	hasError := false
	err := func() { hasError = true }
	print := func(str string) {
		if exitOnErr {
			fmt.Println(str)
		}
	}
	if conf.Upstream == nil {
		print("upstream url can't be nil")
		err()
	}
	if hasError && exitOnErr {
		os.Exit(1)
	}
	return hasError
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := findPath(r)
	// s.pathRate can be read safely without locking because it's never
	// mutated once set during server initialization
	route := s.routeForPath(path)

	// If no route found, send request to upstream without rate limting
	if route == nil {
		s.proxy.ServeHTTP(w, r)
		return
	}

	rateBy, sig, err := getReqSignature(r, route)
	if err != nil {
		w.WriteHeader(err.Code)
		if err.Message != "" {
			fmt.Fprint(w, err.Message)
		}
		if err.LogMessage != "" {
			log.Println()
		}
		return
	}

	log.Println("got request at", r.URL.Path)

	// Find a box for given signature
	s.mu.RLock()
	_, ok := s.boxes[sig]
	s.mu.RUnlock()
	if !ok {
		// Create box with given signature and rateBy fields
		s.mu.Lock()
		log.Println("creating box with signature", sig)
		bx := box{id: sig, server: s, rateBy: rateBy, buckets: map[bucketId]*bucket{}}
		s.boxes[sig] = &bx
		s.mu.Unlock()
	}
	s.mu.RLock()
	bx := s.boxes[sig]
	s.mu.RUnlock()
	bx.RLock()

	// Find appropriate bucket
	buckId := bucketIdForRoute(route, path)
	_, ok = bx.buckets[buckId]
	bx.RUnlock()
	if !ok {
		log.Println("creating bucket for path", path)
		s.createBucket(path, bx)
	}

	// At this position, we can safely assume that the gifter isn't deleting
	// this bucket as it would require gifter to acquire a Write Lock to the **box**
	// which can't be granted while we still have a read lock to the box.
	bx.RLock()
	buck := bx.buckets[buckId]
	log.Println("bucket is", buck)
	bx.RUnlock()

	log.Println("before locking bucket to check for token count", buck)
	buck.Lock()
	log.Println("locking bucket to check for token count", buck)
	// We check if the no. of tokens is >= 1
	// Note that by allowing the tokens to go below negative value, we're enforcing
	// a punishment mechanism for when request is made when you're already rate limited.
	buck.tokens--
	if buck.tokens < 0 {
		// TODO enhance rejection message. Probably allow it to make customizable
		// Note that by allowing the tokens to go below negative value, we're enforcing
		// a punishment mechanism for when request is made when you're already rate limited.
		tryAgainInSeconds := secondsBeforeSuccess(
			time.Now(), buck.lastGifted, buck.rate, buck.tokens)
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, "Rate limited. Try again in %v seconds", tryAgainInSeconds)
		buck.Unlock()
		return
	}
	// Just before leaving, we set the last accessed time on the bucket
	buck.lastAccessed = time.Now()
	// Note that it's important to release this lock before calling ServeHTTP
	// because we would otherwise be unnecessarily holding the lock until we get
	// response from upstream and return that response. This is also the
	// reason why we can't use defer buck.Unlock() or defer s.Unlock()
	// unless we group our code into smaller functions that have no other code
	// besides the critical section.
	buck.Unlock()
	// Call HTTPServer of the underlying ReverseProxy
	s.proxy.ServeHTTP(w, r)
}

// Create a bucket with given id inside the given box.
// Initializes various properties of the bucket like capacity, state time, etc.
// and then registers the bucket to the gifter to collect gift tokens.
func (s *server) createBucket(id reqPath, b *box) {
	b.Lock()
	var rate *rate
	// s.pathRate can be read safely without locking because it's never
	// mutated once set during server initialization
	matchingRoute := s.routeForPath(id)
	if matchingRoute == nil {
		rate = &b.server.conf.BaseRate
	} else {
		rate = rateForRoute(b.server.conf, matchingRoute, b.rateBy)
	}
	acc := time.Now()
	tokens := rate.Capacity
	idForBucket := bucketIdForRoute(matchingRoute, id)
	newBucket := &bucket{
		id:           idForBucket,
		tokens:       tokens,
		rate:         rate,
		lastAccessed: acc,
		lastGifted:   acc,
		box:          b,
		Mutex:        sync.Mutex{},
	}
	b.buckets[idForBucket] = newBucket
	log.Println("created new bucket", newBucket)
	b.Unlock()
	b.server.mu.RLock()
	gifter, ok := b.server.gifters[generateGifterId(*rate)]
	b.server.mu.RUnlock()
	if !ok {
		log.Fatalf("cannot find gifter for rate %v", *rate)
	}
	log.Println("adding gifter to appropriate gifter", id)
	// add the bucket to appropriate gifter
	gifter.addBucket(newBucket)
	log.Println("gifter added", id)
}

// Gets path of the request. This is made a separte function in case there is
// somethign to do with trailing slashes or such.
func findPath(r *http.Request) reqPath {
	return reqPath(r.URL.Path)
}

// Create bucket id for route
func bucketIdForRoute(r *Route, _ reqPath) bucketId {
	return bucketId(r.Pattern.String())
}
