package handlers

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"github.com/hessayon/ya_practicum_go/internal/storage"
)


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


func CreateShortURLHandler(w http.ResponseWriter, r *http.Request) error{
	body, err := io.ReadAll(r.Body)
	if err != nil{
		return errors.New("Error in reading of request's body")
	}
	urlToShort := string(body)
	shortenedURL := getShortURL(urlToShort)
	
	storage.URLs[shortenedURL] = urlToShort
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", serverAddr, shortenedURL)))
	return nil
}

func DecodeShortURLHandler(w http.ResponseWriter, r *http.Request) error{
	if strings.Count(r.URL.Path, "/") > 1 {
		return errors.New("Unsupported URL")
	}
	splittedURLPath := strings.Split(r.URL.Path, "/")
	if len(splittedURLPath) < 2 {
		return errors.New("Unsupported URL")
	}
	shortenedURL := splittedURLPath[1]
	originalURL, found := storage.URLs[shortenedURL]
	if !found {
		return errors.New("Shortened url not found")
	}
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	return nil
}


func MainHandler(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodPost {
		err := CreateShortURLHandler(w, r)
		if err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}	
	} else if r.Method == http.MethodGet {
		err :=DecodeShortURLHandler(w, r)
		if err != nil{
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Method is not allowed", http.StatusBadRequest)
		return
	}
}
