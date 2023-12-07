THIS IS A WORK IN PROGRESS

# Ursa.

`ursa` is a Go package that provides a HTTP rate limiting proxy that can be
used as a throttler in front of your server. Rate limiting is done based on
configuration object provided. It supports rate limiting based on arbitrary
header fields for arbitrary HTTP methods.

## Installation

Install the package using `go get`
```
go get github.com/ursaserver/ursa/
```

## Example

Create a simple rate limiter server. Note that this assumes you have a backend
(that you want to setup a rate limiting for) running at port 8000 in your
localhost. If you don't, you can test the code against `python http.server` by invoking
`python -m http.server 8000`

Here's the code for your rate limiter. As you read the following code note that ursa allows you to:

1. Set up different rate limiting rules based on arbitrary header.
1. If you want to do rate limiting by IP, that's supported and setup for you.

```go
package main

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/ursaserver/ursa"
)

func main() {
    // Create a ursa rate lmiter that is a proxy to your backend
    // Configuration is specified in a separate function for clarity
	ursaLimiter := ursa.New(conf())
	http.ListenAndServe(":3000", ursaLimiter)
}

func conf() ursa.Conf {
    // Define the kind of rates you'd like to use
	BaseRate := ursa.NewRate(5, ursa.Minute)
	UserRate := ursa.NewRate(60, ursa.Minute)

    // The IP address of your backend
	upstream, _ := url.Parse("http://localhost:8000")

    // Here, we are setting up a mechanism for by creating a
    // RateBy struct that lets us perform rate liming based on any 
    // header field (Authentication) in this case.
	RateByUser := ursa.NewRateBy(
		"Authentication", // Heder name
		// Update this function body to put in the logic for checking if the provided value is valid auth token
		// This function either has to go to database, or (better) if you're using JWT token you can just check if 
		// the signature is valid by running without doing anything that expensive
		func(s string) bool { return true },
		// Generate user id from the auth token. Otherwise users will keep generating new ID everytime current token expires.
		func(s string) string { return s },
		400,
		"Invalid authentication",
	)

	// Define the upstream server
	var conf ursa.Conf
	conf.Upstream = upstream

	conf.Routes = []ursa.Route{
		{
			Methods: []string{"GET"},
			Pattern: regexp.MustCompile(".*"),
			Rates: ursa.RouteRates{
				ursa.RateByIP:         BaseRate,
				RateByUser: UserRate,
			},
		},
	}
	return conf
}
```

## Beware
1. Rate limiting by IP will deduct the tokens for users sharing the IP. This is
   a problem for organizational clients sitting under a common gateway. There's
   no workaround currently for public APIs. Use authenticated rate limting rules 
   whenever possible (like `RateByUser` in the above example.) 

## TODOS
Benchmarking 
