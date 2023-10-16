package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestAuthenticationSignature(t *testing.T) {
	// Start rate limiter
	go ratelimiter(ratelimiterPort)

	type test struct {
		expectCode int
		url        string
		headers    map[string]string
	}
	tests := make([]test, 0)
	// Send unauthenticated request
	for i := 0; i < baseRate.Capacity; i++ {
		tests = append(tests, test{expectCode: 200, url: upstreamUrl()})
	}
	// Send one more request, expect to be rate limited
	tests = append(tests, test{expectCode: 429, url: upstreamUrl()})

	// Auth header
	authHeader := make(map[string]string)
	authHeaderInvalid := make(map[string]string)
	authHeader["Authorization"] = "123"
	authHeaderInvalid["Authorization"] = "1"

	// Send valid authenticated requests, expect to succeed
	for i := 0; i < homeAuthenticatedRate.Capacity; i++ {
		tests = append(tests,
			test{expectCode: http.StatusOK, url: upstreamUrl(), headers: authHeader})
	}
	// Send invalid authentication request, expect to fail
	tests = append(tests,
		test{expectCode: http.StatusUnauthorized, url: upstreamUrl(), headers: authHeaderInvalid})

	// Send one more expect to be rate limited
	tests = append(tests,
		test{expectCode: http.StatusTooManyRequests, url: upstreamUrl(), headers: authHeader})
	tests = append(tests,
		test{expectCode: http.StatusTooManyRequests, url: upstreamUrl(), headers: authHeader})

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

func upstreamUrl() string {
	return fmt.Sprintf("http://localhost:%d", ratelimiterPort)
}
