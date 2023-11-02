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

	var err error
	config.ServiceConfig, err = config.NewServiceConfig()
	if err != nil {
		log.Fatalf("Error in NewServiceConfig: %s", err.Error())
	}
	err = storage.InitURLStorage(config.ServiceConfig.Filename)
	if err != nil {
		log.Fatalf("Error in InitURLStorage: %s", err.Error())
	}

	err = storage.InitStorageSaver(config.ServiceConfig.Filename)
	if err != nil {
		log.Fatalf("Error in InitStorageSaver: %s", err.Error())
	}
	defer storage.StorageSaver.Close()

	logger.Log, err = logger.NewServiceLogger("INFO")
	if err != nil {
		log.Fatalf("Error in NewServiceLogger: %s", err.Error())
	}

	serviceRouter := router.NewServiceRouter(logger.Log)

	logger.Log.Info("Start URL Shortener service", zap.String("host", config.ServiceConfig.Host), zap.Int("port", config.ServiceConfig.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.ServiceConfig.Host, config.ServiceConfig.Port), serviceRouter))
}
