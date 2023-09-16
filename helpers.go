package main

import (
	"log"
	"net/http"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Connection",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

var deleteSpecificHeaders = []string{
	"Accept-encoding",
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func deleteHopHeadersHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delHopHeaders(r.Header)
		r.RequestURI = ""

		log.Println("Delete headers completed!")

		next.ServeHTTP(w, r)
	})
}

func removeEncoding(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header
		for _, b := range deleteSpecificHeaders {
			header.Del(b)
		}

		next.ServeHTTP(w, r)
	})
}

type Preparation interface {
	Prepare(http.Handler) http.Handler
}

type PreparationForHttp struct {
}

func (p *PreparationForHttp) Prepare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next2 := deleteHopHeadersHandler(next)
		log.Print("Delete complete")
		next3 := removeEncoding(next2)
		log.Print("Delete remove complete")
		next3.ServeHTTP(w, r)
		// deleteHopHeadersHandler(removeEncoding(next)).ServeHTTP(w, r)
	})
}
