package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/hessayon/ya_practicum_go/internal/logger"
	"go.uber.org/zap"
)

// для передачи userID через контекст
type uuidKey struct{ }

// claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type claims struct {
	jwt.RegisteredClaims
	UserID string
}

const secretKey = "9VAwVrKwAJNUYBySuhPVQWHnwkpuogGj"
const tokenExp = time.Hour * 24
var ErrInvalidValue = errors.New("invalid cookie value")


func buildJWTString(userUUID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims {
			RegisteredClaims: jwt.RegisteredClaims{
					// когда создан токен
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
			},
			// собственное утверждение
			UserID: userUUID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
			return "", err
	}
	return tokenString, nil
}


func getDecryptedCookie(name string, encryptedCookie []byte, secretKey []byte) (string, error) {

	block, err := aes.NewCipher(secretKey)
	if err != nil {
			return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
			return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce := encryptedCookie[:nonceSize]
	ciphertext := encryptedCookie[nonceSize:]

	plaintext, err := aesGCM.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
			return "", ErrInvalidValue
	}

	expectedName, value, ok := strings.Cut(string(plaintext), ":")
	if !ok {
			return "", ErrInvalidValue
	}

	if expectedName != name {
			return "", ErrInvalidValue
	}

	return value, nil
}


func getUserID(cookie *http.Cookie) (string, error) {
	cookieValue, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return "", err
	}
	decryptedCookie, err := getDecryptedCookie(cookie.Name, cookieValue, []byte(secretKey))
	if err != nil {
		return "", err
	}
	claims := &claims{}
	token, err := jwt.ParseWithClaims(decryptedCookie, claims,
	func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			errMsg := fmt.Sprintf("unexpected signing method: %v", t.Header["alg"])
			return nil, errors.New(errMsg)
		}
		return []byte(secretKey), nil
	})
	if err != nil {
			return "", err
	}

	if !token.Valid {
			return "", errors.New("token is not valid")
	}

	return claims.UserID, nil  
}

func getEncryptedCookie(cookie http.Cookie, secretKey []byte) (*http.Cookie, error) {

	block, err := aes.NewCipher(secretKey)
	if err != nil {
			return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
			return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
			return nil, err
	}

	plaintext := fmt.Sprintf("%s:%s", cookie.Name, cookie.Value)
	encryptedValue := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	cookie.Value = hex.EncodeToString(encryptedValue)

	return &cookie, nil
}


func setUserTokenCookie(w http.ResponseWriter, userID string) error {
	
	jwtToken, err := buildJWTString(userID)
	if err != nil {
		return err
	}
	newCookie := http.Cookie{
		Name:     "UserToken",
		Value:    jwtToken,
	}
	encryptedCookie, err := getEncryptedCookie(newCookie, []byte(secretKey))
	if err != nil {
		return err
	}
	http.SetCookie(w, encryptedCookie)

	return nil
}



func AuthenticateUser(authRequired bool, h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("UserToken")
		var userID string
    if cookie == nil || err != nil {

			logger.Log.Warn("error in r.Cookie():", zap.String("error", err.Error()))
			if authRequired {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID = uuid.Must(uuid.NewRandom()).String()
			err := setUserTokenCookie(w, userID)
			if err != nil {
				logger.Log.Error("error in setUserTokenCookie()", zap.String("error", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			userID, err = getUserID(cookie)
		
			if err != nil {
				logger.Log.Warn("error in getUserID():", zap.String("error", err.Error()))
				if authRequired {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				userID = uuid.Must(uuid.NewRandom()).String()
				err = setUserTokenCookie(w, userID)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		h(w, r.WithContext(
			context.WithValue(r.Context(), uuidKey{}, userID),
		))
	})
}

func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(uuidKey{}).(string)
	return userID
}