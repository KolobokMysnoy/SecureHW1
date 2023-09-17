package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Create an HTTP server that listens on port 8080
	bd := MongoDB{}
	p := ProxyHTTP{}
	p.SaveReqAndResp(bd.SaveResponseRequest)
	handler := p

	http.Handle("/", handler)
	fmt.Println("Proxy server listening on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
