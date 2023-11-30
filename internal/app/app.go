package app

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"github.com/hessayon/ya_practicum_go/internal/taskpool"
	"go.uber.org/zap"
)


type App struct {
	Router *chi.Mux
	Storage storage.URLStorage
	SrvcConfig *config.ServiceConfig
	Logger *zap.Logger
	TaskPool *taskpool.TaskPool
}

func NewAppInstance(r *chi.Mux, s storage.URLStorage, l *zap.Logger, c *config.ServiceConfig, tp *taskpool.TaskPool) *App {
	return &App{
		Router: r,
		Storage: s,
		SrvcConfig: c,
		Logger: l,
		TaskPool: tp,
	}
}


func (app *App) Run() error {
	defer app.Storage.Close()
	defer app.TaskPool.Stop()

	app.Logger.Info("Start URL Shortener service", zap.String("host", app.SrvcConfig.Host), zap.Int("port", app.SrvcConfig.Port))
	app.TaskPool.Start()
	return http.ListenAndServe(fmt.Sprintf("%s:%d", app.SrvcConfig.Host, app.SrvcConfig.Port), app.Router)
}