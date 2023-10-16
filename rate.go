package ursa

import (
	"errors"
)

type duration int

type rate struct {
	Capacity            int
	RefillDurationInSec duration
}

type rateByHeader struct {
	header string // Header field to limit the rate by
	valid  func(string) bool
	// signature is a function that converts the header value into
	// something. Here signature means the identity of the user/downstream
	// client that this header value represents. For example if the header
	// value is JWT token, the signature function is the one that takes an
	// access token and returns user id (or sth like that) that serves as
	// the name of the bucket. For details see
	// https://github.com/ursaserver/ursa/issues/12
	signature func(string) string
	failCode  int    // Status code when the validation fails
	failMsg   string // Message to respond with if the validation fails
}

type RouteRates = map[*rateByHeader]rate

const (
	second duration = 1 // Note second is intentionally unexported
	Minute          = second * 60
	Hour            = Minute * 60
	Day             = Hour * 24
)

const (
	MaxRequestsPerSec = 1 // per second
)

var (
	RateByIP = RateByHeader(
		"IP",
		func(_ string) bool { return true }, // Validation
		func(s string) string { return s },  // Header to signature map. We use identity here
		400,
		"")
	ErrMaxRateExceed = errors.New("rate exceed maximum capacity")
	errRouteNotFound = errors.New("route not found")
)

func RateByHeader(name string, valid func(string) bool, signature func(string) string, failCode int, failMsg string) *rateByHeader {
	return &rateByHeader{name, valid, signature, failCode, failMsg}
}

// Create a rate of some amount per given time for example, to create a rate of
// 500 request per hour, say Rate(500, ursa.Hour)
//
// The error returned is non-nil when the rate exceeds the maximum supported
// rate. The rate value when error is not nil must be discarded.
func Rate(amount int, time duration) (rate, error) {
	if amount/int(time) > MaxRequestsPerSec {
		return rate{}, ErrMaxRateExceed
	}
	return rate{amount, time}, nil
}

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
func rateForRoute(conf *Conf, r *Route, rateBy *rateByHeader) *rate {
	var toReturn *rate
	if v, ok := r.Rates[rateBy]; !ok {
		toReturn = &conf.BaseRate
	} else {
		toReturn = &v
	}
	return toReturn
}
