package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/ursaserver/ursa"
)

const HelloMSG = "Hello World\n"

var (
	upstreamAt      = ":8012"
	ratelimiterPort = 3012
	// Note that we keep these numbers small for the ease of testing.
	// In particular, one the issues that may arise if this number is large is that
	// when we expect the request to be rate limited, it might succeed because
	// the gifter has had time to gift the token.
	// This has not been a problem in running tests locally, but we've seen this
	// in tests run in github actions.
	homeAuthenticatedRate = ursa.NewRate(10, ursa.Hour)
	baseRate              = ursa.NewRate(1, ursa.Hour)
)

func conf() (ursa.Conf, error) {
	RateByAuth := ursa.NewRateBy(
		"Authorization",
		func(s string) bool { return len(s) > 1 },
		func(s string) string { return s },
		http.StatusUnauthorized,
		"Unauthorized")
	upstreamURLStr := fmt.Sprintf("http://localhost%s", upstreamAt)
	upstream, err := url.Parse(upstreamURLStr)
	if err != nil {
		fmt.Println(err)
		return *new(ursa.Conf), err
	}
	// Note that r is just an alias to the function to save some typing
	r := regexp.MustCompile

	c := ursa.Conf{
		Upstream: upstream,
		Routes: []ursa.Route{
			{
				Methods: []string{"GET"},
				Pattern: r("/"),
				Rates: ursa.RouteRates{
					ursa.RateByIP: baseRate,
					RateByAuth:    homeAuthenticatedRate,
				},
			},
		},
	}
	if hasError := ursa.ValidateConf(c, true); hasError {
		return c, errors.New("error in configuration")
	}
	return c, nil
}
