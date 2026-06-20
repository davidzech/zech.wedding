package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

const target = "https://www.theknot.com/us/anne-nguyen-and-david-zech-2027-06-05"

func main() {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Rewrite the Host header so theknot.com accepts the request.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		req.URL.Path = targetURL.Path
		req.URL.RawQuery = targetURL.RawQuery
	}

	addr := ":" + port()
	log.Printf("Proxying %s -> %s", addr, target)
	log.Fatal(http.ListenAndServe(addr, proxy))
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
