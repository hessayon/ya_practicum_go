package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/compressing"
	"github.com/hessayon/ya_practicum_go/internal/handlers"
	"github.com/hessayon/ya_practicum_go/internal/logger"
)

var Router *chi.Mux

func InitServiceRouter() {
	Router = chi.NewRouter()
	Router.Post("/", logger.RequestLogger(compressing.GzipCompress(handlers.CreateShortURL)))
	Router.Get("/{id}", logger.RequestLogger(compressing.GzipCompress(handlers.DecodeShortURL)))
	Router.Post("/api/shorten", logger.RequestLogger(compressing.GzipCompress(handlers.CreateShortURLJSON)))
}
