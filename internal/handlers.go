package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

func NewHandler(db *pgxpool.Pool, redisClient *redis.Client) *Handler {
	return &Handler{
		DB:    db,
		Redis: redisClient,
	}
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request CreateLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	request.URL = strings.TrimSpace(request.URL)

	if err := ValidateURL(request.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.CustomAlias != "" {
		if err := ValidateCustomAlias(request.CustomAlias); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if request.ExpiresAt != nil && request.ExpiresAt.Before(time.Now()) {
		http.Error(w, "expiration time must be in the future", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var link Link
	var err error

	if request.CustomAlias != "" {
		link, err = CreateLink(
			ctx,
			h.DB,
			request.CustomAlias,
			request.URL,
			request.ExpiresAt,
		)

		if err != nil {
			http.Error(w, "custom alias is already in use", http.StatusConflict)
			return
		}
	} else {
		for attempt := 0; attempt < 3; attempt++ {
			shortCode, err := GenerateShortCode(6)
			if err != nil {
				http.Error(w, "failed to generate short code", http.StatusInternalServerError)
				return
			}

			link, err = CreateLink(
				ctx,
				h.DB,
				shortCode,
				request.URL,
				request.ExpiresAt,
			)

			if err == nil {
				break
			}

			if attempt == 2 {
				http.Error(w, "failed to create unique short code", http.StatusInternalServerError)
				return
			}
		}
	}

	response := map[string]string{
		"short_code": link.ShortCode,
		"short_url":  "http://localhost:8080/" + link.ShortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/")

	if shortCode == "" {
		var err error

		shortCode, err = GenerateShortCode(6)
		if err != nil {
			http.Error(w, "failed to generate short code", http.StatusInternalServerError)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// 1. Try Redis first
	longURL, err := h.Redis.Get(ctx, shortCode).Result()

	if err == nil {
		// Cache hit
		http.Redirect(w, r, longURL, http.StatusFound)
		return
	}

	if err != redis.Nil {
		// Redis is unavailable, but we can still continue with PostgreSQL.
		log.Printf("redis get failed for %s: %v", shortCode, err)
	}

	// 2. Redis miss -> PostgreSQL
	link, err := GetLinkByShortCode(ctx, h.DB, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "failed to resolve link", http.StatusInternalServerError)
		return
	}

	// 3. Check link status
	if !link.IsActive {
		http.Error(w, "link is disabled", http.StatusGone)
		return
	}

	// 4. Check expiration
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		http.Error(w, "link has expired", http.StatusGone)
		return
	}

	// 5. Cache the URL
	if err := CacheLink(ctx, h.Redis, link); err != nil {
		LogCacheError("set", shortCode, err)
	}

	// 6. Redirect
	http.Redirect(w, r, link.LongURL, http.StatusFound)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	email := NormalizeEmail(request.Email)

	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	passwordHash, err := HashPassword(request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	user, err := CreateUser(
		ctx,
		h.DB,
		email,
		passwordHash,
	)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
}