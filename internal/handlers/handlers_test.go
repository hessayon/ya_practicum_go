package handlers

import (
	"github.com/hessayon/ya_practicum_go/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
			storage.URLs = map[string]string{}
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.requestBody))
			w := httptest.NewRecorder()
			CreateShortURLHandler(w, request)
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
			name: "negative test#1",
			storage: map[string]string{
				"EwHXdJfB": "https://practicum.yandex.ru/",
			},
			requestURL: "/EwHXdJfB/yhfjOHdb",
			want: want{
				code:                400,
				locationHeaderValue: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage.URLs = test.storage
			request := httptest.NewRequest(http.MethodGet, "/EwHXdJfB", nil)
			w := httptest.NewRecorder()
			DecodeShortURLHandler(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.locationHeaderValue, res.Header.Get("Location"))
			res.Body.Close()
		})
	}
}
