package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"net/http"
	"time"
)

// SecretKey - secret key
const SecretKey = "secret"

// CookieTokenName - cookie name
const CookieTokenName = "session_token"

// Auth - empty struct for auth
type Auth struct {
}

// NewAuth - constructor
func NewAuth() *Auth {
	return &Auth{}
}

type claims struct {
	jwt.RegisteredClaims
	ID string
}

// GetID - get userID from cookie
func (a *Auth) GetID(w http.ResponseWriter, r *http.Request) (string, error) {
	var tokenCookie *http.Cookie
	var err error
	// получаем токен
	tokenCookie, _ = r.Cookie(CookieTokenName)

	// если нет, то создаем
	if tokenCookie == nil {
		tokenCookie, err = buildCookie()
		if err != nil {
			return "", err
		}
		http.SetCookie(w, tokenCookie)
	}

	// достаем
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenCookie.Value, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})

	if err != nil {
		return "", err
	}

	// проверяем на валидность
	if !token.Valid {
		return "", fmt.Errorf("token is not valid: %s", tokenCookie)

	}

	//проверяем на наличие идентификатора
	if claims.ID == "" {
		return "", fmt.Errorf("id not found in cookie: %s", tokenCookie)
	}

	return claims.ID, nil
}

func buildCookie() (*http.Cookie, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		ID: uuid.New().String(),
	})

	signedString, err := token.SignedString([]byte(SecretKey))

	if err != nil {
		return nil, err
	}

	return &http.Cookie{
		Name:  CookieTokenName,
		Value: signedString,
		Path:  "/",
	}, nil
}
