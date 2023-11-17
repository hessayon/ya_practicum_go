package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/hessayon/ya_practicum_go/internal/config"
	"github.com/hessayon/ya_practicum_go/internal/mocks"
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
			config.Config = config.NewDefaultServiceConfig()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockURLStorage(ctrl)
			m.EXPECT().Save(gomock.Any()).AnyTimes()
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.requestBody))
			router := chi.NewRouter()
			router.Get("/{id}", DecodeShortURL(m))
			router.Post("/", CreateShortURL(m))
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
		name          string
		correctReq    bool
		getCallKey    string
		getCallValue  string
		getCallStatus bool
		requestURL    string
		want          want
	}{
		{
			name:          "positive test#1",
			correctReq:    true,
			getCallKey:    "EwHXdJfB",
			getCallValue:  "https://practicum.yandex.ru/",
			getCallStatus: true,
			requestURL:    "/EwHXdJfB",
			want: want{
				code:                307,
				locationHeaderValue: "https://practicum.yandex.ru/",
			},
		},
		{
			name:          "negative test#1",
			correctReq:    true,
			getCallKey:    "yhfjOHdb",
			getCallValue:  "",
			getCallStatus: false,
			requestURL:    "/yhfjOHdb",
			want: want{
				code:                400,
				locationHeaderValue: "",
			},
		},
		{
			name:       "negative test#2",
			correctReq: false,
			requestURL: "/EwHXdJfB/yhfjOHdb",
			want: want{
				code:                404,
				locationHeaderValue: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config.Config = config.NewDefaultServiceConfig()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockURLStorage(ctrl)
			if test.correctReq {

				m.EXPECT().GetOriginalURL(test.getCallKey).Return(test.getCallValue, test.getCallStatus)
			}

			request := httptest.NewRequest(http.MethodGet, test.requestURL, nil)
			router := chi.NewRouter()
			router.Get("/{id}", DecodeShortURL(m))
			router.Post("/", CreateShortURL(m))
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
			config.Config = config.NewDefaultServiceConfig()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockURLStorage(ctrl)
			m.EXPECT().Save(gomock.Any()).AnyTimes()
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(test.requestBody))
			router := chi.NewRouter()
			router.Get("/{id}", DecodeShortURL(m))
			router.Post("/", CreateShortURL(m))
			router.Post("/api/shorten", CreateShortURLJSON(m))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			res.Body.Close()
		})
	}
}
