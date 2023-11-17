package main

import (
	"log"

	"github.com/hessayon/ya_practicum_go/internal/app"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/logger"
	"github.com/hessayon/ya_practicum_go/internal/router"
	"github.com/hessayon/ya_practicum_go/internal/storage"
)

func main() {

	var err error
	config.Config, err = config.NewServiceConfig()
	if err != nil {
		log.Fatalf("Error in NewServiceConfig: %s", err.Error())
	}

	logger.Log, err = logger.NewServiceLogger("INFO")
	if err != nil {
		log.Fatalf("Error in NewServiceLogger: %s", err.Error())
	}

	var urlStorage storage.URLStorage
	urlStorage, err = storage.NewURLStorage(config.Config)
	if err != nil {
		log.Fatalf("Error in NewURLStorage: %s", err.Error())
	}

	serviceRouter := router.NewServiceRouter(logger.Log, urlStorage)

	application := app.NewAppInstance(serviceRouter, urlStorage, logger.Log, config.Config)
	log.Fatal(application.Run())
}
