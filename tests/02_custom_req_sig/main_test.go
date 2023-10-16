package main

import (
	"fmt"
	"net/http"
	"testing"
)

type test struct {
	expectCode int
	url        string
	headers    map[string]string
}

func TestPre(t *testing.T) {
	// Start rate limiter
	// Other tests won't have to start the rate limiter
	go ratelimiter(ratelimiterPort)
}

func helperChecker(tests []test, t *testing.T) {
	for _, test := range tests {
		client := new(http.Client)
		req, err := http.NewRequest("GET", test.url, nil)
		if err != nil {
			// Calling fatal here because it's most likely our
			// mistake rather than the library we're testing
			t.Fatal("error creating request", err)
		}
		for key, val := range test.headers {
			req.Header.Add(key, val)
		}
		rsp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}
		if gotCode := rsp.StatusCode; gotCode != test.expectCode {
			t.Error("expected code", test.expectCode, "got code", gotCode)
		}
	}
}

// func TestUnAuthenticated(t *testing.T) {
// 	tests := make([]test, 0)
// 	// Send unauthenticated request
// 	for i := 0; i < baseRate.Capacity; i++ {
// 		tests = append(tests, test{expectCode: 200, url: upstreamUrl()})
// 	}
// 	// Send one more request, expect to be rate limited
// 	tests = append(tests, test{expectCode: 429, url: upstreamUrl()})
// 	helperChecker(tests, t)
// }

// func TestAuthenticated(t *testing.T) {
// 	// Start rate limiter
// 	// go ratelimiter(ratelimiterPort)
//
// 	tests := make([]test, 0)
// 	// Auth
// 	authHeader := make(map[string]string)
// 	authHeader["Authorization"] = "123"
//
// 	// Send valid authenticated requests, expect to succeed
// 	for i := 0; i < homeAuthenticatedRate.Capacity; i++ {
// 		tests = append(tests,
// 			test{expectCode: http.StatusOK, url: upstreamUrl(), headers: authHeader})
// 	}
// 	helperChecker(tests, t)
// }

func TestInvalidAuth(t *testing.T) {
	// Start rate limiter
	// go ratelimiter(ratelimiterPort)

	tests := make([]test, 0)
	// Auth
	authHeaderInvalid := make(map[string]string)
	authHeaderInvalid["Authorization"] = "1"
	// Send invalid authentication request, expect to fail
	tests = append(tests,
		test{expectCode: http.StatusUnauthorized, url: upstreamUrl(), headers: authHeaderInvalid})
	helperChecker(tests, t)
}

func upstreamUrl() string {
	return fmt.Sprintf("http://localhost:%d", ratelimiterPort)
}
