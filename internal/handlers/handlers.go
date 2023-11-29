package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/logger"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"github.com/hessayon/ya_practicum_go/internal/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type requestBody struct {
	URL string `json:"url"`
}

type responseBody struct {
	ShortenURL string `json:"result"`
}

type requestBatchBody struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type responseBatchBody struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
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
			UUID:        middleware.UserIDFromContext(r.Context()),
			ShortURL:    shortenedURL,
			OriginalURL: urlToShort,
		})
		statusCode := http.StatusCreated
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				statusCode = http.StatusConflict
				shortenedURL, err = s.GetShortURL(urlToShort)
				if err != nil {
					http.Error(w, "shortened url not found", http.StatusBadRequest)
					return
				}
			} else {
				logger.Log.Error("Error in s.GetShortURL()", zap.String("error", err.Error()))
			}
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf("%s/%s", config.Config.BaseAddr, shortenedURL)))
	})
}

func DecodeShortURL(s storage.URLStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortenedURL := chi.URLParam(r, "id")
		originalURL, err := s.GetOriginalURL(shortenedURL)
		if err == storage.ErrAlreadyDeleted {
			w.WriteHeader(http.StatusGone)
			return
		} else if err != nil {
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
			UUID:        middleware.UserIDFromContext(r.Context()),
			ShortURL:    shortenedURL,
			OriginalURL: reqBody.URL,
		})
		statusCode := http.StatusCreated
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				statusCode = http.StatusConflict
				shortenedURL, err = s.GetShortURL(reqBody.URL)
				if err != nil {
					http.Error(w, "shortened url not found", http.StatusBadRequest)
					return
				}
			} else {
				logger.Log.Error("Error in s.Save()", zap.String("error", err.Error()))
			}
		}

		respBody := responseBody{
			ShortenURL: fmt.Sprintf("%s/%s", config.Config.BaseAddr, shortenedURL),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(respBody); err != nil {
			logger.Log.Error("error in encoding response body", zap.String("originalURL", reqBody.URL), zap.String("shortenURL", respBody.ShortenURL))
			http.Error(w, "service internal error", http.StatusBadRequest)
			return
		}
	})
}

func Ping(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("pgx", config.Config.DBDsn)
	if err != nil {
		logger.Log.Error("error in db.Open()", zap.String("db_dsn", config.Config.DBDsn), zap.String("error", err.Error()))
		http.Error(w, "db is not connected", http.StatusInternalServerError)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logger.Log.Error("error in db.Ping()", zap.String("db_dsn", config.Config.DBDsn), zap.String("error", err.Error()))
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
				UUID:        middleware.UserIDFromContext(r.Context()),
				ShortURL:    shortenedURL,
				OriginalURL: data.OriginalURL,
			})
			responseData = append(responseData, responseBatchBody{
				CorrelationID: data.CorrelationID, ShortURL: fmt.Sprintf("%s/%s", config.Config.BaseAddr, shortenedURL),
			})
		}
		err = s.SaveBatch(urlsData)
		if err != nil {
			logger.Log.Error("Error in s.SaveBatch()", zap.String("error", err.Error()))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(responseData); err != nil {
			logger.Log.Error("error in encoding response body", zap.String("error", err.Error()))
			http.Error(w, "service internal error", http.StatusBadRequest)
			return
		}
	})
}


func GetURLsByUser(s storage.URLStorage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usersURLs, err := s.GetURLsByUserID(middleware.UserIDFromContext(r.Context()))
		if err != nil {
			logger.Log.Error("error in GetURLsByUserID", zap.String("error", err.Error()))
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resList := make([]storage.URLData, 0, len(usersURLs))
		for _, elem := range usersURLs {
			shortURL := fmt.Sprintf("%s/%s", config.Config.BaseAddr, elem.ShortURL)
			elem.ShortURL = shortURL
			resList = append(resList, elem)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resList); err != nil {
			logger.Log.Error("error in encoding response body", zap.String("error", err.Error()))
			http.Error(w, "service internal error", http.StatusBadRequest)
			return
		}

	})
}

func DeleteURLs(s storage.URLStorage) http.HandlerFunc {
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error("error in decoding request body")
			http.Error(w, "service internal error", http.StatusBadRequest)
			return
		}
		var shortURLs []string
		err = json.Unmarshal(body, &shortURLs)
		if err != nil {
			logger.Log.Error("error in decoding unmarshal request body")
			http.Error(w, "service internal error", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		err = s.DeleteURLs(userID, shortURLs...)
		if err != nil {
			logger.Log.Error("error in DeleteURLs()", zap.String("userID", userID), zap.Strings("shortURLs", shortURLs), zap.String("error", err.Error()))
			return
		} 
		logger.Log.Info("URLs are deleted", zap.String("userID", userID), zap.Strings("shortURLs", shortURLs))
	})
}