package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/router"
	"github.com/hessayon/ya_practicum_go/internal/storage"
)

func main() {
	err := config.InitServiceConfig()
	if err != nil {
		log.Fatalf("Error in InitServiceConfig: %s", err.Error())
	}
	router.InitServiceRouter()

	storage.URLs = make(map[string]string)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.ServiceConfig.Host, config.ServiceConfig.Port), router.Router))
}
