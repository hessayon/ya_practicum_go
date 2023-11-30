package middleware


import (
	"net/http"
	"time"

	"go.uber.org/zap"
)


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


// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(log *zap.Logger, h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}

		h(&lw, r) // обслуживание оригинального запроса
		duration := time.Since(start)
		log.Info("got incoming HTTP request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.String("duration", duration.String()),
			zap.Strings("content_encoding", r.Header.Values("Content-Encoding")),
			zap.Strings("accept_encoding", r.Header.Values("Accept-Encoding")),
		)
		log.Info("response to incoming HTTP request",
			zap.Int("status", responseData.Status),
			zap.Int("size", responseData.Size),
		)
	})
}