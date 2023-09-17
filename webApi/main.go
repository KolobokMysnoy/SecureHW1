package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	router := chi.NewRouter()
	router.Get("/requests")
	router.Get("/requests/{id}", func(writer http.ResponseWriter, request *http.Request) {
		username := chi.URLParam(request, "username") // 👈 getting path param
		_, err := writer.Write([]byte("Hello " + username))
		if err != nil {
			log.Println(err)
		}
	})
	router.Get("/repeat/{id}")
	router.Get("/scan/{id}")

}

func repeat(writer http.ResponseWriter, request *http.Request) {
	username := chi.URLParam(request, "username") // 👈 getting path param
	_, err := writer.Write([]byte("Hello " + username))
	if err != nil {
		log.Println(err)
	}
}
