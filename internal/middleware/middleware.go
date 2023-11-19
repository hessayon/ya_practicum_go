package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hessayon/ya_practicum_go/internal/compressing"
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


// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}


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


func GzipCompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		currentWriter := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		supportsGzip := compressing.CheckSupportOfGzip(r.Header.Values("Accept-Encoding"))

		if compressing.IsGzipContentType(r.Header.Get("Content-Type")) && supportsGzip{
			compressWr := compressing.NewCompressWriter(w)
			currentWriter = compressWr
			defer compressWr.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		sendsGzip := compressing.CheckSupportOfGzip(r.Header.Values("Content-Encoding"))
		if sendsGzip {
			cr, err := compressing.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h(currentWriter, r)
	}
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
			zap.Strings("accept_encoding", r.Header.Values("Accept-Encoding")),
		)
	})
}
const SECRET_KEY = "supersecretkey"
const TOKEN_EXP = time.Hour * 24

func BuildJWTString(userUUID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims {
			RegisteredClaims: jwt.RegisteredClaims{
					// когда создан токен
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
			},
			// собственное утверждение
			UserID: userUUID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
			return "", err
	}
	return tokenString, nil
}

func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
	func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			errMsg := fmt.Sprintf("unexpected signing method: %v", t.Header["alg"])
			return nil, errors.New(errMsg)
		}
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
			return "", err
	}

	if !token.Valid {
			return "", errors.New("token is not valid")
	}

	return claims.UserID, nil  
}

func AuthenticateUser(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("UserToken")
    if cookie == nil || err != nil{
			jwtToken, err := BuildJWTString(r.RequestURI)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			newCookie := http.Cookie{
        Name:     "UserToken",
        Value:    jwtToken,
        // Path:     "/",
        MaxAge:   3600,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    	}
			http.SetCookie(w, &newCookie)
		}
		h(w, r)
	})
}