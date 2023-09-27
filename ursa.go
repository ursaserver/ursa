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
	rateBys           []RateBy
	bucketsStaleAfter time.Duration
	boxes             map[reqSignature]*box
	gifters           map[gifterId]*gifter
	pathRate          func(reqPath) *Route
	proxy             *httputil.ReverseProxy
	sync.RWMutex
}

func (s *server) String() string {
	return fmt.Sprintf("server %v", s.id)
}

type bucketId string

type box struct {
	server  *server
	id      reqSignature // request signature
	rateBy  RateBy
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
	ValidateConf(conf)
	serverId := fmt.Sprintf("%v", rand.Float64())
	s := &server{conf: &conf, id: serverId}
	s.boxes = make(map[reqSignature]*box)
	s.gifters = make(map[gifterId]*gifter)
	s.bucketsStaleAfter = time.Duration(0)
	s.proxy = httputil.NewSingleHostReverseProxy(conf.Upstream)
	s.pathRate = memoize.Unary(func(r reqPath) *Route {
		// Note that memoization is possible since the configuration is not
		// changed once loaded.
		return routeForPath(r, &conf)
	})
	allRateBys := make(map[RateBy]bool)
	for _, route := range conf.Routes {
		rates := route.Rates
		for rateBy, r := range rates {
			allRateBys[rateBy] = true
			gifterId := generateGifterId(r)
			// Check if the gifter with the id already exists
			s.RLock()
			_, ok := s.gifters[gifterId]
			s.RUnlock()
			if !ok {
				// Create a gifter
				g := &gifter{
					rate:    r,
					server:  s,
					id:      gifterId,
					buckets: new(linkedList[*bucket]),
				}
				s.Lock()
				s.gifters[gifterId] = g
				s.Unlock()
			}
		}
	}
	s.rateBys = make([]RateBy, 0)
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

// Validate configuration
// exists all the error messages if the config is invalid
func ValidateConf(conf Conf) {
	hasError := false
	err := func() { hasError = true }
	if conf.Upstream == nil {
		fmt.Println("upstream url can't be nil")
		err()
	}
	if hasError {
		os.Exit(1)
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO if the request is made to non rate limited path, forward to reverse
	// proxy immediately
	rateBy, sig := findReqSignature(r, s.rateBys)
	// Find a box for given signature
	s.RLock()
	_, ok := s.boxes[sig]
	s.RUnlock()
	if !ok {
		// create box with given signature and rateBy fields
		s.Lock()
		b := box{id: sig, server: s, rateBy: rateBy, buckets: map[bucketId]*bucket{}}
		s.boxes[sig] = &b
		s.Unlock()
	}
	s.RLock()
	b := s.boxes[sig]
	path := findPath(r)
	b.RLock()
	_, ok = b.buckets[bucketId(path)]
	b.RUnlock()
	if !ok {
		s.createBucket(path, b)
	}

	// At this position, we can safely assume that the gifter isn't deleting
	// this bucket as it would require gifter to acquire a Write Lock to the box
	// which can't be granted while there's still a reader.
	buck := b.buckets[bucketId(path)]
	b.RUnlock()

	buck.Lock()
	defer buck.Unlock()
	// We check if the no. of tokens is >= 1
	// Note that by allowing the tokens to go below negative value, we're enforcing
	// a punishment mechanism for when request is made when you're already rate limited.
	buck.tokens--
	if buck.tokens < 0 {
		// TODO enhance rejection message. Probably allow it to make customizable
		// Note that by allowing the tokens to go below negative value, we're enforcing
		// a punishment mechanism for when request is made when you're already rate limited.
		tryAgainInSeconds := buck.tokens * int(tickOnceEvery(*buck.rate).Seconds())
		fmt.Fprintf(w, "Rate limited. Try again in %v seconds", tryAgainInSeconds)
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	// Just before leaving, we set the last accessed time on the bucket
	buck.lastAccessed = time.Now()
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
	matchingRoute := s.pathRate(id)
	if matchingRoute == nil {
		rate = &b.server.conf.BaseRate
	} else {
		rate = rateForRoute(b.server.conf, matchingRoute, b.rateBy)
	}
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
	b.buckets[bucketId(id)] = newBucket
	b.Unlock()
	b.server.RLock()
	gifter, ok := b.server.gifters[generateGifterId(*rate)]
	if !ok {
		log.Fatalf("cannot find gifter for rate %v", *rate)
	}
	// add the bucket to appropriate gifter
	gifter.addBucket(newBucket)
}

// Find what the request signature should be for a request and also finds what
// is the thing to rate limit by (RateBy) the given request.
func findReqSignature(req *http.Request, rateBys []RateBy) (RateBy, reqSignature) {
	// Find if any of the header fields in RateBy are present.
	rateby := RateByIP // default
	key := ""
	for _, r := range rateBys {
		if req.Header.Get(string(r)); r != "" {
			rateby = r
			key = string(r)
			break
		}
	}
	if rateby == RateByIP {
		key = clientIpAddr(req)
	}
	return rateby, reqSignature(fmt.Sprintf("%v-%v", rateby, key))
}

// Gets path of the request. This is made a separte function in case there is
// somethign to do with trailing slashes or such.
func findPath(r *http.Request) reqPath {
	return reqPath(r.URL.Path)
}
