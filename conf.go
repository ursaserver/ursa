package ursa

import (
	"io"
	"net/url"
	"regexp"
)

type Conf struct {
	Upstream *url.URL
	Routes   []Route
	Logfile  io.Writer
}

type Route struct {
	Methods []string
	Pattern *regexp.Regexp // regex describing HTTP path to match
	Rates   RouteRates
}
