package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/handlers"
	"github.com/hessayon/ya_practicum_go/internal/middleware"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"go.uber.org/zap"
)



func NewServiceRouter(log *zap.Logger, s storage.URLStorage) *chi.Mux {
	newRouter := chi.NewRouter()
	newRouter.Post("/", middleware.AuthenticateUser(false, middleware.RequestLogger(log, middleware.GzipCompress(handlers.CreateShortURL(s)))))
	newRouter.Get("/{id}", middleware.AuthenticateUser(false, middleware.RequestLogger(log, middleware.GzipCompress(handlers.DecodeShortURL(s)))))
	newRouter.Post("/api/shorten", middleware.AuthenticateUser(false, middleware.RequestLogger(log, middleware.GzipCompress(handlers.CreateShortURLJSON(s)))))
	newRouter.Get("/ping", middleware.AuthenticateUser(false, middleware.RequestLogger(log, middleware.GzipCompress(handlers.Ping))))
	newRouter.Post("/api/shorten/batch", middleware.AuthenticateUser(false, middleware.RequestLogger(log, middleware.GzipCompress(handlers.CreateShortURLBatch(s)))))
	newRouter.Post("/api/user/urls", middleware.AuthenticateUser(true, middleware.RequestLogger(log, middleware.GzipCompress(handlers.GetURLsByUser(s)))))
	return newRouter
}
