package ursa

import (
	"net/url"
	"regexp"
)

type Conf struct {
	routes       []Route
	defaultRates map[RateBy]rate
	nonRateLimit []NonRateLimitRoute
}

type Route struct {
	pattern   *regexp.Regexp // regex describing HTTP path to match
	rate      map[RateBy]rate
	forwardTo *url.URL // the address of the server to forward requests to
}

type NonRateLimitRoute struct {
	pattern   *regexp.Regexp // regex describing HTTP path to match
	forwardTo string         // the address of the server to forward requests to
}
