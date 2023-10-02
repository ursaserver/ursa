package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
)

func TestPathSignature(t *testing.T) {
	go ratelimiter(ratelimiterPort)

	// Send just the number of allowed requests to one of page/1, ...
	// page/9 pages. All of these requests should go through and the next
	// one should give return status code 429. As per issue
	// https://github.com/ursaserver/ursa/issues/6
	// we want to ensure that requests made to page/1, .. page/9
	// all deduct the token from the same bucket.

	maxPageId := 9
	getPageId := func() int {
		return (rand.Int() % maxPageId) + 1
	}

	getPageUrl := func(pageId int) string {
		pagePath := fmt.Sprintf("/page/%d", pageId)
		return fmt.Sprintf("http://localhost:%d%s", ratelimiterPort, pagePath)
	}

	for i := 0; i < pageDetailRate.Capacity; i++ {
		// Send request to rate limiter
		// Ensure all requests go through
		pageId := getPageId()
		url := getPageUrl(pageId)
		rsp, err := http.Get(url)
		if err != nil {
			t.Error(err)
		}
		// Ensure status OK
		if code := rsp.StatusCode; code != 200 {
			t.Error("got status code", code)
		}
		// Ensure correct response body
		bf := new(bytes.Buffer)
		io.Copy(bf, rsp.Body)
		got := bf.String()
		want := innerPageMsg(pageId)
		if got != want {
			t.Errorf("expected msg body %s\n got %s\n", want, got)
		}
		defer rsp.Body.Close()
	}

	// Send a request to rate limiter
	// Ensure request is rate limited
	url := getPageUrl(getPageId())
	rsp, err := http.Get(url)
	if err != nil {
		t.Error(err)
	}
	// Ensure rate limited
	if code := rsp.StatusCode; code != 429 {
		t.Errorf("expected status code %d got %d\n", 429, code)
	}
}
