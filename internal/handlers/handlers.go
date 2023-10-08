package handlers

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/storage"
)

const serverAddr = "http://localhost:8080"

func getShortURL(url string) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const urlLength = 8
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shortURL := make([]byte, urlLength)
	for i := range shortURL {
		shortURL[i] = charset[r.Intn(len(charset))]
	}
	return string(shortURL)
}

func CreateShortURLHandler(w http.ResponseWriter, r *http.Request){
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error in reading of request's body",  http.StatusBadRequest)
		return
	}
	urlToShort := string(body)
	shortenedURL := getShortURL(urlToShort)

	storage.URLs[shortenedURL] = urlToShort
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", serverAddr, shortenedURL)))
}

func DecodeShortURLHandler(w http.ResponseWriter, r *http.Request){
	
	shortenedURL := chi.URLParam(r, "id")
	originalURL, found := storage.URLs[shortenedURL]
	if !found {
		http.Error(w, "shortened url not found",  http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

