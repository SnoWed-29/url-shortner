package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

type APIKeyAuthenticator func(context.Context, string) (int64, error)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			userID, err := UserIDFromJWT(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := SetUserID(r.Context(), userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UnifiedMiddleware rejects requests that include both credential types so an
// invalid or malformed JWT can never fall back to an API key.
func UnifiedMiddleware(
	jwtSecret string,
	apiKeyAuthenticator APIKeyAuthenticator,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			rawAPIKey := strings.TrimSpace(r.Header.Get("X-API-Key"))

			if authHeader != "" && rawAPIKey != "" {
				http.Error(w, "multiple authentication credentials provided", http.StatusUnauthorized)
				return
			}

			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != "Bearer" {
					http.Error(w, "invalid or expired token", http.StatusUnauthorized)
					return
				}

				userID, err := UserIDFromJWT(parts[1], jwtSecret)
				if err != nil {
					http.Error(w, "invalid or expired token", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r.WithContext(SetUserID(r.Context(), userID)))
				return
			}

			if rawAPIKey == "" {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()

			userID, err := apiKeyAuthenticator(ctx, rawAPIKey)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					http.Error(w, "invalid api key", http.StatusUnauthorized)
					return
				}

				http.Error(w, "authentication failed", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r.WithContext(SetUserID(r.Context(), userID)))
		})
	}
}

func UserIDFromJWT(tokenString string, jwtSecret string) (int64, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(jwtSecret), nil
		},
	)
	if err != nil || !token.Valid {
		return 0, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user id")
	}

	return int64(userIDFloat), nil
}

func SetUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}
