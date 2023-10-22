package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/logger"
	"github.com/hessayon/ya_practicum_go/internal/router"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"go.uber.org/zap"
)

func main() {
	err := config.InitServiceConfig()
	if err != nil {
		log.Fatalf("Error in InitServiceConfig: %s", err.Error())
	}
	router.InitServiceRouter()

	err = logger.InitServiceLogger("INFO")
	if err != nil {
		log.Fatalf("Error in InitServiceLogger: %s", err.Error())
	}

	storage.URLs = make(map[string]string)
	logger.Log.Info("Start URL Shortener service", zap.String("host", config.ServiceConfig.Host), zap.Int("port", config.ServiceConfig.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.ServiceConfig.Host, config.ServiceConfig.Port), router.Router))
}
