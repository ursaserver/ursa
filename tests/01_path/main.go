// Helper functions for performing integrated testing
package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ursaserver/ursa"
)

// Run basic server at given port
func upstream() {
	upstreamPort := strings.Split(upstreamURLStr, ":")[2]
	http.HandleFunc("/", home)
	http.HandleFunc("/about", about)
	http.HandleFunc("/pages", pages)
	for i := 0; i < 10; i++ {
		path := fmt.Sprintf("/page/%d", i)
		http.HandleFunc(path, innerPage(i))
	}
	http.ListenAndServe(fmt.Sprintf(":%s", upstreamPort), nil)
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
	http.ListenAndServe(fmt.Sprintf(":%d", rateLimiterPort), handler)
}
