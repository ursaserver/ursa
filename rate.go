package ursa

import (
	"errors"
)

// Header field to limit the rate by
type RateBy string

type duration int

type rate struct {
	Capacity            int
	RefillDurationInSec duration
}

const (
	second duration = 1 // Note second is intentionally unexported
	Minute          = second * 60
	Hour            = Minute * 60
	Day             = Hour * 24
)

const (
	RateByIP      = RateBy("IP")
	MaxRatePerSec = 1 // per second
)

var (
	ErrMaxRateExceed = errors.New("rate exceed maximum capacity")
	errRouteNotFound = errors.New("route not found")
)

// Create a rate of some amount per given time for example, to create a rate of
// 500 request per hour, say Rate(500, ursa.Hour)
//
// The error returned is non-nil when the rate exceeds the maximum supported
// rate. The rate value when error is not nil must be discarded.
func Rate(amount int, time duration) (rate, error) {
	if amount/int(time) > MaxRatePerSec {
		return rate{}, ErrMaxRateExceed
	}
	return rate{amount, time}, nil
}

type RouteRates = map[RateBy]rate

// Returns the route on configuration that should be used for the a given
// reqPath. If no matching rate is found, nil, is returned.
func routeForPath(p reqPath, conf *Conf) *Route {
	// Search linearly through the routes in the configuration to find a
	// pattern that matches reqPath. Note that speed won't be an issue here
	// since this function is supposed to be memoized when using.
	// Memoization should be possible since the configuration is not changed once loaded.
	for _, r := range conf.Routes {
		if r.Pattern.MatchString(string(p)) {
			return &r
		}
	}
	return nil
}

// Returns the rate to be used for the the given route based on given
// configuration and and rateBy params. Expects conf and route to be non nil.
// TODO, still needs to be reasonsed what are the consequences of returning
// *rate vs rate
func rateForRoute(conf *Conf, r *Route, rateBy RateBy) *rate {
	var toReturn *rate
	if v, ok := r.Rates[rateBy]; !ok {
		toReturn = &conf.BaseRate
	} else {
		toReturn = &v
	}
	return toReturn
}
