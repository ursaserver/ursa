package ursa

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
)

func TestGetReqSignature(t *testing.T) {
	// Rate by auth
	authHeader := "Authorization"
	var isAuthValid IsValidHeaderValue = func(x string) bool { return len(x) > 5 }
	identitySigFn := func(x string) string { return x }
	authFailMsg := "Auth Failed"
	RateByAuth := RateByHeader(
		authHeader,
		isAuthValid,
		identitySigFn,
		401,
		authFailMsg,
	)

	// Rate by application API
	applicationAPIHeader := "Application"
	validAppAPIs := []string{"1", "101", "101011", "10000"}
	isApplicationAPIValid := func(x string) bool {
		for _, v := range validAppAPIs {
			if v == x {
				return true
			}
		}
		return false
	}
	invalidAPIMsg := "Invalid API"
	RateByApplicattionAPI := RateByHeader(
		applicationAPIHeader,
		isApplicationAPIValid,
		identitySigFn,
		401,
		invalidAPIMsg,
	)

	type test struct {
		req       *http.Request
		route     *Route
		expRateBy *rateBy
		expReqSig reqSignature
		expErr    *ErrReqSignature
	}

	requestForPath := func(path string) *http.Request {
		urlStr := fmt.Sprintf("https://example.com%v", path)
		r, _ := http.NewRequest("GET", urlStr, nil)
		return r
	}
	addHeaderToRequest := func(r *http.Request, key string, value string) *http.Request {
		r.Header.Add(key, value)
		return r
	}

	tests := []test{
		{
			req: requestForPath("/about"),
			route: &Route{
				Pattern: regexp.MustCompile("/about"),
				Rates:   RouteRates{},
			},
			expRateBy: nil,
			expReqSig: "",
			expErr:    &ErrReqSignature{Code: NoRateDefinedOnRouteHTTPCode, LogMessage: "No rate bys defined on route pattern /about"},
		},
		{
			req: requestForPath("/about"),
			route: &Route{
				Pattern: regexp.MustCompile("/about"),
				Rates:   RouteRates{RateByAuth: Rate(100, Hour)},
			},
			expRateBy: nil,
			expReqSig: "",
			expErr:    &ErrReqSignature{Code: NoRateDefinedByUserOnRequest},
		},
		{
			req: requestForPath("/about"),
			route: &Route{
				Pattern: regexp.MustCompile("/about"),
				Rates:   RouteRates{RateByApplicattionAPI: Rate(100, Minute)},
			},
			expRateBy: nil,
			expReqSig: "",
			expErr:    &ErrReqSignature{Code: NoRateDefinedByUserOnRequest},
		},
		{
			req: requestForPath("/about"),
			route: &Route{
				Pattern: regexp.MustCompile("/about"),
				Rates: RouteRates{
					RateByApplicattionAPI: Rate(100, Minute),
					RateByAuth:            Rate(100, Hour),
				},
			},
			expRateBy: nil,
			expReqSig: "",
			expErr:    &ErrReqSignature{Code: NoRateDefinedByUserOnRequest},
		},
		{
			req: requestForPath("/about"),
			route: &Route{
				Pattern: regexp.MustCompile("/about"),
				Rates: RouteRates{
					RateByApplicattionAPI: Rate(100, Minute),
					RateByAuth:            Rate(100, Hour),
					RateByIP:              Rate(60, Hour),
				},
			},
			expRateBy: nil,
			expReqSig: "", // client ipv6 address
			// Note that we get error in this request because the the IP address is invalid
			// as the request isn't fully sent, only the http request
			expErr: &ErrReqSignature{Code: http.StatusBadRequest, Message: "invalid IP"},
		},
	}

	auths := []string{"a", "ab", "abc", "abcd", "abcde", "abcdef"}
	for _, auth := range auths {
		testCase := test{
			req: addHeaderToRequest(requestForPath("/about"), authHeader, auth),
			route: &Route{
				Pattern: regexp.MustCompile("/about"),
				Rates: RouteRates{
					RateByApplicattionAPI: Rate(100, Minute),
					RateByAuth:            Rate(100, Hour),
					RateByIP:              Rate(60, Hour),
				},
			},
		}
		// Check if it's valid
		if RateByAuth.valid(auth) {
			testCase.expRateBy = RateByAuth
			testCase.expReqSig = createReqSignature(RateByAuth, RateByAuth.signature(auth))
		} else {
			testCase.expErr = &ErrReqSignature{Code: RateByAuth.failCode, Message: RateByAuth.failMsg}
		}
		tests = append(tests, testCase)
	}

	for _, test := range tests {
		gotRateBy, gotReqSig, gotErr := getReqSignature(test.req, test.route)

		if (test.expErr == nil && gotErr != nil) || (test.expErr != nil && gotErr == nil) {
			t.Errorf("got error %v expected error %v\n", gotErr, test.expErr)
		} else if test.expErr != nil && gotErr != nil {
			if test.expErr.Code != gotErr.Code {
				t.Errorf("got error code %v expected error code %v\n",
					gotErr.Code, test.expErr.Code)
			}
			if test.expErr.Message != gotErr.Message {
				t.Errorf("got error message %q expected error message %q\n",
					gotErr.Message, test.expErr.Message)
			}
			if test.expErr.LogMessage != gotErr.LogMessage {
				t.Errorf("got error log message %q expected error log message %q\n",
					gotErr.LogMessage, test.expErr.Message)
			}
		}

		if test.expRateBy != gotRateBy {
			t.Errorf("got rate by %v expected  rate by %v\n", gotRateBy, test.expRateBy)
		}

		if gotReqSig != test.expReqSig {
			t.Errorf("expected reqSig %v got %v\n", test.expReqSig, gotReqSig)
		}

	}
}
