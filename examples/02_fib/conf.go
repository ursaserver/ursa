package main

import (
	"net/url"
	"regexp"

	"github.com/ursaserver/ursa"
)

func ursaconfig(upstream *url.URL) ursa.Conf {
	Rate1 := ursa.NewRate(5, ursa.Minute)

	// Define the upstream server
	var conf ursa.Conf
	conf.Upstream = upstream

	// Define the rates for various routes
	IP := ursa.RateByIP
	r := regexp.MustCompile
	conf.Routes = []ursa.Route{
		{
			Methods: []string{"GET"},
			Pattern: r(".*"),
			Rates: ursa.RouteRates{
				IP: Rate1,
			},
		},
	}
	return conf
}
