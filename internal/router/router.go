package router


import (
	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/handlers"
)


var Router *chi.Mux

func InitServiceRouter() {
	Router = chi.NewRouter()
	Router.Post("/", handlers.CreateShortURLHandler)
	Router.Get("/{id}", handlers.DecodeShortURLHandler)
}