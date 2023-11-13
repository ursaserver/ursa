package main

import (
	"fmt"
	"log"
	"net/http"
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
			Pattern: r(".*"), Rates: ursa.RouteRates{
				IP: Rate1,
			},
		},
	}
	return conf
}

func main() {
	// Note that you should have a some sort of server running at localhost 8000
	// You can easily initialize such server by running something like:
	// python -m http.server -b localhost 8000
	upstream, err := url.Parse("http://localhost:8000")
	if err != nil {
		log.Fatalf("error parsing url")
	}
	conf := ursaconfig(upstream)
	handler := ursa.New(conf)

	// Start th rate limiter
	addr := ":3000"
	fmt.Println("listening at", addr)
	http.ListenAndServe(addr, handler)
}
