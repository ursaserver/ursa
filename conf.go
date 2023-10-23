package ursa

import (
	"net/url"
	"regexp"
)

type Conf struct {
	Upstream *url.URL
	Routes   []Route
}

type Route struct {
	Methods []string
	Pattern *regexp.Regexp // regex describing HTTP path to match
	Rates   RouteRates
}
