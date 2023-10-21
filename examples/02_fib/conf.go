package main

import (
	"net/url"
	"regexp"

	"github.com/ursaserver/ursa"
)

func ursaconfig(upstream *url.URL) ursa.Conf {
	BaseRate := ursa.Rate(5, ursa.Minute)
	Rate1 := ursa.Rate(5, ursa.Minute)

	// Define the upstream server
	var conf ursa.Conf
	conf.Upstream = upstream

	// Define base rate
	conf.BaseRate = BaseRate

	// Define the rates for various routes
	IP := ursa.RateByIP
	r := regexp.MustCompile
	conf.Routes = []ursa.Route{
		{Pattern: r(".*"), Rates: ursa.RouteRates{
			IP: Rate1,
		}},
	}
	return conf
}
