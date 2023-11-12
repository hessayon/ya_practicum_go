package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortURLHandler(t *testing.T) {

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name        string
		requestBody string
		want        want
	}{
		{
			name:        "positive test#1",
			requestBody: "https://practicum.yandex.ru/",
			want: want{
				code:        201,
				contentType: "text/plain",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config.ServiceConfig = config.NewDefaultServiceConfig()
			storage.Storage, _ = storage.NewURLStorage("")
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.requestBody))
			router := chi.NewRouter()
			router.Get("/{id}", DecodeShortURL)
			router.Post("/", CreateShortURL)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			res.Body.Close()
		})
	}
}

func TestDecodeShortURLHandler(t *testing.T) {
	type want struct {
		code                int
		locationHeaderValue string
	}
	tests := []struct {
		name       string
		storage    map[string]string
		requestURL string
		want       want
	}{
		{
			name: "positive test#1",
			storage: map[string]string{
				"EwHXdJfB": "https://practicum.yandex.ru/",
			},
			requestURL: "/EwHXdJfB",
			want: want{
				code:                307,
				locationHeaderValue: "https://practicum.yandex.ru/",
			},
		},
		{
			name: "negative test#1",
			storage: map[string]string{
				"EwHXdJfB": "https://practicum.yandex.ru/",
			},
			requestURL: "/yhfjOHdb",
			want: want{
				code:                400,
				locationHeaderValue: "",
			},
		},
		{
			name: "negative test#2",
			storage: map[string]string{
				"EwHXdJfB": "https://practicum.yandex.ru/",
			},
			requestURL: "/EwHXdJfB/yhfjOHdb",
			want: want{
				code:                404,
				locationHeaderValue: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config.ServiceConfig = config.NewDefaultServiceConfig()
			storage.Storage, _ = storage.NewURLStorage("")
			storage.Storage.Store = test.storage
			request := httptest.NewRequest(http.MethodGet, test.requestURL, nil)
			router := chi.NewRouter()
			router.Get("/{id}", DecodeShortURL)
			router.Post("/", CreateShortURL)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.locationHeaderValue, res.Header.Get("Location"))
			res.Body.Close()
		})
	}
}


func TestCreateShortURLJSONHandler(t *testing.T) {

	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name        string
		requestBody string
		want        want
	}{
		{
			name:        "positive test#1",
			requestBody: "{\"url\": \"https://practicum.yandex.ru\"}",
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name:        "negative test#1: emptyBody",
			requestBody: "",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config.ServiceConfig = config.NewDefaultServiceConfig()
			storage.Storage, _ = storage.NewURLStorage("")
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(test.requestBody))
			router := chi.NewRouter()
			router.Get("/{id}", DecodeShortURL)
			router.Post("/", CreateShortURL)
			router.Post("/api/shorten", CreateShortURLJSON)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			res.Body.Close()
		})
	}
}
