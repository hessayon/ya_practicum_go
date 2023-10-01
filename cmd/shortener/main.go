package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var urls map[string]string

const serverAddr = "http://localhost:8080"

func getShortURL(url string) (string) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const urlLength = 8
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shortURL := make([]byte, urlLength)
	for i := range shortURL {
		shortURL[i] = charset[r.Intn(len(charset))]
	}
	return string(shortURL)
}


func mainHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil{
			http.Error(w, "Error in reading of request's body", http.StatusBadRequest)
			return
		}
		urlToShort := string(body)
		shortenedURL := getShortURL(urlToShort)
		urls[shortenedURL] = urlToShort
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("%s/%s", serverAddr, shortenedURL)))
	
	} else if r.Method == http.MethodGet {

		if strings.Count(r.URL.Path, "/") > 1 {
			http.Error(w, "Unsupported URL", http.StatusBadRequest)
			return
		}
		splittedURLPath := strings.Split(r.URL.Path, "/")
		if len(splittedURLPath) < 2 {
			http.Error(w, "Unsupported URL",  http.StatusBadRequest)
			return
		}
		shortenedURL := splittedURLPath[1]
		originalURL, found := urls[shortenedURL]
		if !found {
			http.Error(w, "Shortened url not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "Method is not allowed", http.StatusBadRequest)
		return
	}
}


func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainHandler)
	urls = make(map[string]string)
	http.ListenAndServe(`:8080`, mux)
}
