package main

import (
	"fmt"
	"net/http"
)

const (
	homeMsg  string = "Welcome to home"
	aboutMsg string = "About"
	pagesMsg string = "Welcome to pages"
)

func innerPageMsg(pageno int) string {
	return fmt.Sprintf("Hello from page %d", pageno)
}

func home(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, homeMsg)
}

func about(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, aboutMsg)
}

func pages(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, pagesMsg)
}

func innerPage(pageno int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, innerPageMsg(pageno))
	}
}
