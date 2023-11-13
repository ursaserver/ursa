package ursa

import (
	"fmt"
	"net/http"
)

type (
	Duration                 int
	IsValidHeaderValue       func(string) bool
	SignatureFromHeaderValue func(string) string
)

type Rate struct {
	Capacity            int
	RefillDurationInSec Duration
}

type ErrReqSignature struct {
	Message    string
	LogMessage string
	Code       int
}

type RateBy struct {
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

type RouteRates = map[*RateBy]Rate

const (
	second Duration = 1 // Note second is intentionally unexported
	Minute          = second * 60
	Hour            = Minute * 60
	Day             = Hour * 24
)

const (
	NoRateDefinedOnRouteHTTPCode = http.StatusInternalServerError
	NoRateDefinedByUserOnRequest = http.StatusUnauthorized
)

var RateByIP = NewRateBy(
	"",
	func(_ string) bool { return true }, // Validation
	func(s string) string { return s },  // Header to signature map. We use identity here
	400,
	"")

func NewRateBy(
	name string,
	valid IsValidHeaderValue,
	signature SignatureFromHeaderValue,
	failCode int,
	failMsg string,
) *RateBy {
	return &RateBy{name, valid, signature, failCode, failMsg}
}

func NewRate(amount int, time Duration) Rate {
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
		if val := r.Header.Get(by.header); val != "" {
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
		if !limitRateBy.valid(key) {
			err = &ErrReqSignature{Code: limitRateBy.failCode, Message: limitRateBy.failMsg}
		}
		keySignature = limitRateBy.signature(key)
		keyReqSig = createReqSignature(limitRateBy, keySignature)
	} else {
		if rateBysCount == 0 {
			err = &ErrReqSignature{
				Code:       NoRateDefinedOnRouteHTTPCode,
				LogMessage: fmt.Sprintf("No rate bys defined on route pattern %s", route.Pattern),
			}
		} else {
			err = &ErrReqSignature{Code: NoRateDefinedByUserOnRequest}
		}
	}
	// If err exists return zero values for  rateBy and request signature
	if err != nil {
		return nil, "", err
	}
	return limitRateBy, keyReqSig, err
}

func createReqSignature(by *RateBy, val string) reqSignature {
	return reqSignature(fmt.Sprintf("%v-%v", by.header, val))
}
