package logger

import (
	"net/http"

	"go.uber.org/zap"
)

// глобальный логгер
var Log *zap.Logger = zap.NewNop()

type (
	ResponseData struct {
		Status int
		Size   int
	}

	// добавляем реализацию http.ResponseWriter
	LoggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		ResponseData        *ResponseData
	}
)

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size // захватываем размер
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode // захватываем код статуса
}

func NewServiceLogger(level string) (*zap.Logger, error) {
	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = logLevel
	zlogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zlogger, nil
}
