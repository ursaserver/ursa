package main

import (
	"fmt"
	"net/http"

	"github.com/ursaserver/ursa"
)

func main() {
	// Run ursa rate limiter in the foreground
	ratelimiter(ratelimiterPort)
}

// Setup a hello world server
func upstream() {
	http.HandleFunc("/", hello)
	http.ListenAndServe(upstreamAt, nil)
}

func hello(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, HelloMSG)
}

func ratelimiter(rateLimiterPort int) {
	// Start upstream
	go upstream()
	// Run a rate limiter proxy at given port
	configuration, err := conf()
	if err != nil {
		fmt.Println("invalid configuration, cannot start server")
	}
	handler := ursa.New(configuration)
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", rateLimiterPort), handler)
}
