package app

import (
	"net/http"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"go.uber.org/zap"
)


type App struct {
	Router *chi.Mux
	Storage storage.URLStorage
	SrvcConfig *config.ServiceConfig
	Logger *zap.Logger
}

func NewAppInstance(r *chi.Mux, s storage.URLStorage, l *zap.Logger, c *config.ServiceConfig) *App {
	return &App{
		Router: r,
		Storage: s,
		SrvcConfig: c,
		Logger: l,
	}
}


func (app *App) Run() error {
	defer app.Storage.Close()

	app.Logger.Info("Start URL Shortener service", zap.String("host", app.SrvcConfig.Host), zap.Int("port", app.SrvcConfig.Port))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", app.SrvcConfig.Host, app.SrvcConfig.Port), app.Router)
}