package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
	"errors"

	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/logger"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type requestBody struct{
	URL string `json:"url"`
}

type responseBody struct{
	ShortenURL string `json:"result"`
}

type requestBatchBody struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

type responseBatchBody struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
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
		statusCode := http.StatusCreated
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				statusCode = http.StatusConflict
				var found bool
				shortenedURL, found = s.GetShortURL(urlToShort)
				if !found {
					http.Error(w, "shortened url not found", http.StatusBadRequest)
					return
				}
			} else {
				logger.Log.Error("Error in s.Save()", zap.String("error", err.Error()))
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf("%s/%s", config.ServiceConfig.BaseAddr, shortenedURL)))
	})
}

func DecodeShortURL(s storage.URLStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL := chi.URLParam(r, "id")
		originalURL, found := s.GetOriginalURL(shortenedURL)
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
		statusCode := http.StatusCreated
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				statusCode = http.StatusConflict
				var found bool
				shortenedURL, found = s.GetShortURL(reqBody.URL)
				if !found {
					http.Error(w, "shortened url not found", http.StatusBadRequest)
					return
				}
			} else {
				logger.Log.Error("Error in s.Save()", zap.String("error", err.Error()))
			}
		}
		
		respBody := responseBody{
			ShortenURL: fmt.Sprintf("%s/%s", config.ServiceConfig.BaseAddr, shortenedURL),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
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


func CreateShortURLBatch(s storage.URLStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody []requestBatchBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "error in decoding of request's body", http.StatusBadRequest)
			return
		}
		if len(reqBody) == 0 {
			w.WriteHeader(http.StatusCreated)
			return
		}
		urlsData := make([]*storage.URLData, 0, len(reqBody))
		responseData := make([]responseBatchBody, 0, len(reqBody))
		for _, data := range reqBody {
			shortenedURL := getShortURL(data.OriginalURL)
			urlsData = append(urlsData, &storage.URLData{
				UUID: r.RequestURI,
				ShortURL: shortenedURL,
				OriginalURL: data.OriginalURL,
			})
			responseData = append(responseData, responseBatchBody{
				CorrelationID: data.CorrelationID, ShortURL: fmt.Sprintf("%s/%s", config.ServiceConfig.BaseAddr, shortenedURL),
			})
		}
		err = s.SaveBatch(urlsData)
		if err != nil {
			logger.Log.Error("Error in s.SaveBatch()", zap.String("error", err.Error()))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(responseData); err != nil{
			logger.Log.Error("error in encoding response body")
			http.Error(w, "service internal error", http.StatusBadRequest)
			return
		}
	})
}