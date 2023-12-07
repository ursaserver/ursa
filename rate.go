package ursa

import (
	"fmt"
	"net/http"
)

type (
	duration                 int // used in the Rate struct
	IsValidHeaderValue       func(string) bool
	SignatureFromHeaderValue func(string) string
)

// This struct is made public for library authors, if you're writing a rate
// limiter server using this package, you should use the function [ursa.NewRate]
// This struct may be made private in future versions.
type Rate struct {
	Capacity            int
	RefillDurationInSec duration
}

// This is the error objec that is returned if the there is an error creating
// request signature from a request. A request signature for an unathenticated
// user may mean their IP address.
type ErrReqSignature struct {
	Message    string
	LogMessage string
	Code       int
}

// This struct is made public for library authors, if you're writing a rate
// limiter server using this package, you should probably use the function
// NewRateBy
//
// This struct may be made private in future versions. If you want to
// perform rate limiting for unauthenticated users, [ursa.RateByIP] is
// already provided for you
type RateBy struct {
	Header string // Header field to limit the rate by
	Valid  func(string) bool
	// Signature is a function that converts the header value into
	// something. Here Signature means the identity of the user/downstream
	// client that this header value represents. For example if the header
	// value is JWT token, the Signature function is the one that takes an
	// access token and returns user id (or sth like that) that serves as
	// the name of the bucket. For details see
	// https://github.com/ursaserver/ursa/issues/12
	Signature func(string) string
	FailCode  int    // Status code when the validation fails
	FailMsg   string // Message to respond with if the validation fails
}

// RouteRates is a map from RateBys for the route. This is a one of the things
// to describe when defining a [ursa.Route] wherein you describe the different
// rate lmiting methods that are supported for the route and their
// corresponding Rates.
type RouteRates = map[*RateBy]Rate

// ursa.duration can be created by using one of ursa.Minute, ursa.Hour or ursa.Day
const (
	second duration = 1 // Note second is intentionally unexported
	Minute          = second * 60
	Hour            = Minute * 60
	Day             = Hour * 24
)

const (
	// Error of this status code is returned if a request is made to a route
	// where no rate is defined
	NoRateDefinedOnRouteHTTPCode = http.StatusInternalServerError
	// Error of this status code is returned the desired header value
	// is not found in the request when creating the request signature
	HeaderValueNotFoundInRequestForRateLimiting = http.StatusUnauthorized
)

// RateBy for rate limiting by IP. Note that that users using the same public
// gateway might be rate limted for each other's request. Currently there is no
// workaround for that.
var RateByIP = NewRateBy(
	"",
	func(_ string) bool { return true }, // Validation
	func(s string) string { return s },  // Header to signature map. We use identity here
	400,
	"")

// Create a new RateBy based on an arbitary header
//
// Params:
// - header: HTTP Header to preform rate limting by
// - valid: function that checks if the value for that header is valid
// - signature: function that transforms header value to a user identifier
// - failCode: status code to respond if the validation of header value fails
// - failMsg: message if the validation of header value fails
func NewRateBy(
	header string,
	valid IsValidHeaderValue,
	signature SignatureFromHeaderValue,
	failCode int,
	failMsg string, // Message to respond if the validation of header value fails
) *RateBy {
	return &RateBy{header, valid, signature, failCode, failMsg}
}

// Create a Rate object
//
// Params:
// - amount: How many requests are allowed
// - time: the duration of time for the amount of requests
//
// You'll have to use either of the three durations:
// [ursa.Day], [ursa.Hour], [ursa.Minute],
//
// If you want to set a rate limit of 20 requests per minute they you say
//
//	rate := ursa.NewRate(20, ursa.Minute)
func NewRate(amount int, time duration) Rate {
	return Rate{amount, time}
}

func isMethodInMethods(candidate string, methods []string) bool {
	for _, current := range methods {
		if current == candidate {
			return true
		}
	}
	return false
}

// Returns the route on configuration that should be used for the a given
// reqPath. If no matching rate is found, nil, is returned.
func routeForPath(p reqPathAndMethod, conf *Conf) *Route {
	path := p.path
	method := p.method
	// Search linearly through the routes in the configuration to find a
	// pattern that matches reqPath. Note that speed won't be an issue here
	// since this function is supposed to be memoized when using.
	// Memoization should be possible since the configuration is not changed once loaded.
	for _, r := range conf.Routes {
		if r.Pattern.MatchString(string(path)) && isMethodInMethods(method, r.Methods) {
			return &r
		}
	}
	return nil
}

// Returns *rateBy, reqSignature, *ErrReqSignature for a *Route based on
// *http.Request If the route contains no rates to apply for the request, send
// appropriate error.
func getReqSignature(r *http.Request, route *Route) (*RateBy, reqSignature, *ErrReqSignature) {
	var limitRateBy *RateBy
	keySignature := ""
	key := ""
	var err *ErrReqSignature = nil
	var keyReqSig reqSignature = ""
	rateBysCount := 0

	for by := range route.Rates {
		rateBysCount++
		if by == RateByIP {
			limitRateBy = RateByIP
			continue
		}
		if val := r.Header.Get(by.Header); val != "" {
			limitRateBy = by
			key = val
			break
		}
	}

	if limitRateBy == RateByIP {
		k, e := clientIpAddr(r)
		key = k
		if e != nil {
			err = &ErrReqSignature{Code: http.StatusBadRequest, Message: e.Error()}
		}
	}
	if limitRateBy != nil {
		if !limitRateBy.Valid(key) {
			err = &ErrReqSignature{Code: limitRateBy.FailCode, Message: limitRateBy.FailMsg}
		}
		keySignature = limitRateBy.Signature(key)
		keyReqSig = createReqSignature(limitRateBy, keySignature)
	} else {
		if rateBysCount == 0 {
			err = &ErrReqSignature{
				Code:       NoRateDefinedOnRouteHTTPCode,
				LogMessage: fmt.Sprintf("No rate bys defined on route pattern %s", route.Pattern),
			}
		} else {
			err = &ErrReqSignature{Code: HeaderValueNotFoundInRequestForRateLimiting}
		}
	}
	// If err exists return zero values for  rateBy and request signature
	if err != nil {
		return nil, "", err
	}
	return limitRateBy, keyReqSig, err
}

func createReqSignature(by *RateBy, val string) reqSignature {
	return reqSignature(fmt.Sprintf("%v-%v", by.Header, val))
}
