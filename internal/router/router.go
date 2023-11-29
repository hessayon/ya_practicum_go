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
	newRouter.Post("/", middleware.RequestLogger(log, middleware.AuthenticateUser(false, middleware.GzipCompress(handlers.CreateShortURL(s)))))
	newRouter.Get("/{id}", middleware.RequestLogger(log, middleware.AuthenticateUser(false, middleware.GzipCompress(handlers.DecodeShortURL(s)))))
	newRouter.Post("/api/shorten", middleware.RequestLogger(log, middleware.AuthenticateUser(false, middleware.GzipCompress(handlers.CreateShortURLJSON(s)))))
	newRouter.Get("/ping", middleware.RequestLogger(log, middleware.AuthenticateUser(false, middleware.GzipCompress(handlers.Ping))))
	newRouter.Post("/api/shorten/batch", middleware.RequestLogger(log, middleware.AuthenticateUser(false, middleware.GzipCompress(handlers.CreateShortURLBatch(s)))))
	newRouter.Get("/api/user/urls", middleware.RequestLogger(log, middleware.AuthenticateUser(true, middleware.GzipCompress(handlers.GetURLsByUser(s)))))
	newRouter.Delete("/api/user/urls", middleware.RequestLogger(log, middleware.AuthenticateUser(false, middleware.GzipCompress(handlers.DeleteURLs(s)))))
	return newRouter
}
