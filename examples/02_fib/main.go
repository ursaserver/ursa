// This program puts the ursa rate limiter in front of a fibonacci series
// server that computes nth number in the sequence using the naievee recursion.
//
// It is used to simuate situations in which the upstream server is performing
// a heavy computation, potetionally leading to keep many connections from ursa
// and upstream alive for a long period of time.
//
// Note that because this package contains multiple files, if you're running via
// go run without building, you'll have to do be at the directory where the files
// are and run using the command:
// go run .

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ursaserver/ursa"
)

func main() {
	upstream, err := url.Parse("http://localhost:8000")
	if err != nil {
		log.Fatalf("error parsing url")
	}
	conf := ursaconfig(upstream)
	handler := ursa.New(conf)

	// Start fibonacci server at port 8000
	go fibServer(8000)

	// Start th rate limiter
	addr := ":3000"
	fmt.Println("listening at", addr)
	http.ListenAndServe(addr, handler)
}

func fibServer(port int) {
	addr := fmt.Sprintf(":%d", port)
	f := func(w http.ResponseWriter, r *http.Request) {
		// Read the number from the query param
		numstr := r.URL.Query().Get("n")
		log.Printf("got request for n=%v", numstr)
		num, err := strconv.Atoi(numstr)
		if err != nil {
			num = 0
		}
		fmt.Fprintf(w, "fin(%d)=%d", num, fib(num))
	}
	handler := http.HandlerFunc(f)
	http.ListenAndServe(addr, handler)
}

func fib(n int) int {
	if n < 2 {
		return 1
	}
	return fib(n-1) + fib(n-2)
}
