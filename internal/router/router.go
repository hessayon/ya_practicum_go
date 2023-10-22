package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/handlers"
	"github.com/hessayon/ya_practicum_go/internal/logger"
)

var Router *chi.Mux

func InitServiceRouter() {
	Router = chi.NewRouter()
	Router.Post("/", logger.RequestLogger(handlers.CreateShortURL))
	Router.Get("/{id}", logger.RequestLogger(handlers.DecodeShortURL))
	Router.Post("/api/shorten", logger.RequestLogger(handlers.CreateShortURLJSON))
}
