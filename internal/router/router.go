package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/handlers"
	"github.com/hessayon/ya_practicum_go/internal/middleware"
	"go.uber.org/zap"
)



func NewServiceRouter(log *zap.Logger) *chi.Mux {
	newRouter := chi.NewRouter()
	newRouter.Post("/", middleware.RequestLogger(log, middleware.GzipCompress(handlers.CreateShortURL)))
	newRouter.Get("/{id}", middleware.RequestLogger(log, middleware.GzipCompress(handlers.DecodeShortURL)))
	newRouter.Post("/api/shorten", middleware.RequestLogger(log, middleware.GzipCompress(handlers.CreateShortURLJSON)))
	return newRouter
}
