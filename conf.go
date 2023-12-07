package ursa

import (
	"io"
	"net/url"
	"regexp"
)

// Configuration to provide when creating server using [ursa.New]
//
// Upstream is the url of the backend that the requests should be proxied for.
// Note that you can only specify one upstream
//
// Routes is a slice of [ursa.Route]. Note that the order of Routes matter
// in that the rate limiting is done based on the rules of the the first matching route
//
// Logfile is an io.Writer where the logs should be written
type Conf struct {
	Upstream *url.URL
	Routes   []Route
	Logfile  io.Writer
}

// A Route describes the rules of rate limiting for urls matched by the regex Pattern
// of the route.
//
// Methods is a slice of strings representing method names to match. Method
// names can be arbitrary. A nil Methods matches all methods. Method names are
// case insensetive.
//
// Rates is a map ([ursa.RouteRates]) that maps describes the different rates for different
// RateBys for the route. This is useful for example if on the api/product/ route you want
// different rate liting rules for for authenticated or non authenticated users. In which
// case you'll probably use the [ursa.RateBy IP] as as RateBy to describe rate for non
// authenticated users and another RateBy created using [ursa.NewRateBy] for authenticated
// users
type Route struct {
	Methods []string
	Pattern *regexp.Regexp // regex describing HTTP path to match
	Rates   RouteRates
}
