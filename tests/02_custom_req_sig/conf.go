package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/ursaserver/ursa"
)

const HelloMSG = "Hello World\n"

var (
	upstreamAt               = ":8012"
	ratelimiterPort          = 3012
	homeAuthenticatedRate, _ = ursa.Rate(10, ursa.Minute)
	baseRate, _              = ursa.Rate(2, ursa.Minute)
)

func conf() (ursa.Conf, error) {
	RateByAuth := ursa.RateBy("Authorization") // Header field to limit rate by
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
		BaseRate: baseRate,
		Routes: []ursa.Route{
			{Pattern: r("/"), Rates: ursa.RouteRates{
				ursa.RateByIP: baseRate,
				RateByAuth:    homeAuthenticatedRate,
			}},
		},
	}
	if hasError := ursa.ValidateConf(c, true); hasError {
		return c, errors.New("error in configuration")
	}
	return c, nil
}
