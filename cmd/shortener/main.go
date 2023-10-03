package main

import (
	"net/http"

	"github.com/hessayon/ya_practicum_go/internal/handlers"
	"github.com/hessayon/ya_practicum_go/internal/storage"
)



func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handlers.MainHandler)
	storage.URLs = make(map[string]string)
	http.ListenAndServe(`:8080`, mux)
}
