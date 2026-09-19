package apikeys

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/SnoWed-29/url-shortener/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Middleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := strings.TrimSpace(r.Header.Get("X-API-Key"))

			if rawKey == "" {
				http.Error(w, "api key required", http.StatusUnauthorized)
				return
			}

			hash := sha256.Sum256([]byte(rawKey))
			keyHash := hex.EncodeToString(hash[:])

			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()

			apiKey, err := GetAPIKeyByHash(
				ctx,
				db,
				keyHash,
			)

			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					http.Error(w, "invalid api key", http.StatusUnauthorized)
					return
				}

				http.Error(w, "authentication failed", http.StatusInternalServerError)
				return
			}

			if err := UpdateLastUsed(ctx, db, apiKey.ID); err != nil {
				http.Error(w, "authentication failed", http.StatusInternalServerError)
				return
			}

			authContext := auth.SetUserID(r.Context(), apiKey.UserID)

			next.ServeHTTP(w, r.WithContext(authContext))
		})
	}
}
