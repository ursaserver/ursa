package ursa

import (
	"net/url"
	"regexp"
)

type Conf struct {
	Upstream *url.URL
	Routes   []Route
	BaseRate rate
	// Todo add default rates
	// DefaultRates map[RateBy]rate
	// TODO add non options to specify non-rate limiting routes
	// nonRateLimit []NonRateLimitRoute
}

type Route struct {
	Pattern *regexp.Regexp // regex describing HTTP path to match
	Rates   RouteRates
}

// type NonRateLimitRoute struct {
// 	Pattern *regexp.Regexp // regex describing HTTP path to match
// }
