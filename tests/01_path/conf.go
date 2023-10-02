package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/ursaserver/ursa"
)

var (
	upstreamURLStr    = "http://localhost:8012"
	ratelimiterPort   = 3012
	pageDetailRate, _ = ursa.Rate(2, ursa.Minute)
	baseRate, _       = ursa.Rate(5, ursa.Minute)
)

func conf() (ursa.Conf, error) {
	upstream, err := url.Parse(upstreamURLStr)
	if err != nil {
		fmt.Println(err)
		return *new(ursa.Conf), err
	}

	// Note that r is just an alias to the function to save some typing
	r := regexp.MustCompile

	c := ursa.Conf{
		Upstream: upstream,
		BaseRate: baseRate,
		// regexp.MustCompile The pattern `/page/[^\/]+`  (note that
		// backticks just denote string literal) tells to match the
		// string that starts with "/page/" then has one or more of
		// anything except forward slash. Thus, it matches, "/page/1"
		// and "page/hello-123-word" but not "page/1/world" or
		// "page/1/". Note that the trailing slash matters.
		Routes: []ursa.Route{
			{Pattern: r(`/page/[^\/]+`), Rates: ursa.RouteRates{ursa.RateByIP: pageDetailRate}},
		},
	}
	if hasError := ursa.ValidateConf(c, true); hasError {
		return c, errors.New("error in configuration")
	}
	return c, nil
}
