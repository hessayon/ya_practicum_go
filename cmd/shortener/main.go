package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/handlers"
	"github.com/hessayon/ya_practicum_go/internal/storage"
)

func main() {
	r := chi.NewRouter()
	r.Post("/", handlers.CreateShortURLHandler)
	r.Get("/{id}", handlers.DecodeShortURLHandler)
	storage.URLs = make(map[string]string)
	http.ListenAndServe(`:8080`, r)
}
