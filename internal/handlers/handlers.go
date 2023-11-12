package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
	"database/sql"
	
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/logger"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"go.uber.org/zap"
)

type requestBody struct{
	URL string `json:"url"`
}

type responseBody struct{
	ShortenURL string `json:"result"`
}

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

func CreateShortURL(s storage.URLStorage) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error in reading of request's body", http.StatusBadRequest)
			return
		}
		urlToShort := string(body)
		shortenedURL := getShortURL(urlToShort)

		err = s.Save(&storage.URLData{
			UUID: r.RequestURI,
			ShortURL: shortenedURL,
			OriginalURL: urlToShort,
		})
		if err != nil {
			logger.Log.Error("Error in s.Save()", zap.String("error", err.Error()))
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("%s/%s", config.ServiceConfig.BaseAddr, shortenedURL)))
	})
}

func DecodeShortURL(s storage.URLStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL := chi.URLParam(r, "id")
		originalURL, found := s.Get(shortenedURL)
		if !found {
			http.Error(w, "shortened url not found", http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	})
}



func CreateShortURLJSON(s storage.URLStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody requestBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "error in decoding of request's body", http.StatusBadRequest)
			return
		}

		shortenedURL := getShortURL(reqBody.URL)

		err = s.Save(&storage.URLData{
				UUID: r.RequestURI,
				ShortURL: shortenedURL,
				OriginalURL: reqBody.URL,
			})
		if err != nil {
			logger.Log.Error("Error in s.Save()", zap.String("error", err.Error()))
		}
		
		respBody := responseBody{
			ShortenURL: fmt.Sprintf("%s/%s", config.ServiceConfig.BaseAddr, shortenedURL),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(respBody); err != nil{
			logger.Log.Error("error in encoding response body", zap.String("originalURL", reqBody.URL) ,zap.String("shortenURL", respBody.ShortenURL))
			http.Error(w, "service internal error", http.StatusBadRequest)
			return
		}
	})
}


func Ping(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("pgx", config.ServiceConfig.DBDsn)
	if err != nil {
		logger.Log.Error("error in db.Open()", zap.String("db_dsn", config.ServiceConfig.DBDsn), zap.String("error", err.Error()))
		http.Error(w, "db is not connected", http.StatusInternalServerError)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logger.Log.Error("error in db.Ping()", zap.String("db_dsn", config.ServiceConfig.DBDsn), zap.String("error", err.Error()))
		http.Error(w, "db is not connected", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}