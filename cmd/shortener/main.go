package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/handlers"
	"github.com/hessayon/ya_practicum_go/internal/storage"
)

func main() {
	err := config.InitServiceConfig()
	if err != nil {
		log.Fatalf("Error in InitSerInitServiceConfig: %s", err.Error())
	}
	r := chi.NewRouter()
	r.Post("/", handlers.CreateShortURLHandler)
	r.Get("/{id}", handlers.DecodeShortURLHandler)
	storage.URLs = make(map[string]string)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.ServiceConfig.Host, config.ServiceConfig.Port), r))
}
