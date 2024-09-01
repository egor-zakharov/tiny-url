package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

// SecretKey - secret key
const SecretKey = "secret"

// TokenName - cookie name
const TokenName = "session-token"

// UserIDKey - metadata key
const UserIDKey = "user-id"

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
	tokenCookie, _ = r.Cookie(TokenName)

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
		Name:  TokenName,
		Value: signedString,
		Path:  "/",
	}, nil
}

// Interceptor check token
func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("metadata is not provided")
	}

	tokens := md.Get(TokenName)
	if len(tokens) == 0 {
		return "", fmt.Errorf("token not found in metadata")
	}

	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokens[0], claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	if claims.ID == "" {
		return "", fmt.Errorf("id not found in token")
	}

	md.Set(UserIDKey, claims.ID)
	ctx = metadata.NewIncomingContext(ctx, md)

	return handler(ctx, req)
}

// BuildToken - create token
func BuildToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		ID: uuid.New().String(),
	})

	return token.SignedString([]byte(SecretKey))
}

// GetIDGrpc - get userId for grpc calls
func (a *Auth) GetIDGrpc(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("metadata is not provided")
	}

	tokens := md.Get(UserIDKey)
	if len(tokens) == 0 {
		return "", fmt.Errorf("token not found in metadata")
	}

	return tokens[0], nil
}

// ExcludeMethodsInterceptor - exclude method
func ExcludeMethodsInterceptor(excludeMethods []string, interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		for _, method := range excludeMethods {
			// Если текущий метод - это исключение, пропускаем интерцептор
			if info.FullMethod == method {
				return handler(ctx, req)
			}
		}

		// В противном случае применяем интерцептор
		return interceptor(ctx, req, info, handler)
	}
}
