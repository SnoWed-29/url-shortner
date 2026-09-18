package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SnoWed-29/url-shortener/internal/auth"
	"github.com/SnoWed-29/url-shortener/internal/cache"
	"github.com/SnoWed-29/url-shortener/internal/links"
	"github.com/SnoWed-29/url-shortener/internal/users"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	DB        *pgxpool.Pool
	Redis     *redis.Client
	JWTSecret string
}

func NewHandler(
	db *pgxpool.Pool,
	redisClient *redis.Client,
	jwtSecret string,
) *Handler {
	return &Handler{
		DB:        db,
		Redis:     redisClient,
		JWTSecret: jwtSecret,
	}
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserID(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request links.CreateLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	request.URL = strings.TrimSpace(request.URL)

	if err := links.ValidateURL(request.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.CustomAlias != "" {
		if err := links.ValidateCustomAlias(request.CustomAlias); err != nil {
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

	var link links.Link
	var err error

	if request.CustomAlias != "" {
		link, err = links.CreateLink(
			ctx,
			h.DB,
			userID,
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
			shortCode, err := links.GenerateShortCode(6)
			if err != nil {
				http.Error(w, "failed to generate short code", http.StatusInternalServerError)
				return
			}

			link, err = links.CreateLink(
				ctx,
				h.DB,
				userID,
				shortCode,
				request.URL,
				request.ExpiresAt,
			)

			if err == nil {
				break
			}

			log.Printf("failed to create link with short_code=%s: %v", shortCode, err)

			if attempt == 2 {
				http.Error(w, "failed to create link", http.StatusInternalServerError)
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

		shortCode, err = links.GenerateShortCode(6)
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
	link, err := links.GetLinkByShortCode(ctx, h.DB, shortCode)
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
	if err := cache.CacheLink(ctx, h.Redis, link); err != nil {
		cache.LogCacheError("set", shortCode, err)
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

	var request users.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	email := auth.NormalizeEmail(request.Email)

	if email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.HashPassword(request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	user, err := users.CreateUser(
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
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request users.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	email := auth.NormalizeEmail(request.Email)

	if email == "" || request.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	user, err := users.GetUserByEmail(ctx, h.DB, email)
	if err != nil {
		// Don't reveal whether the email exists.
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(request.Password, user.PasswordHash) {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(user.ID, h.JWTSecret)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) ListLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserID(r.Context())

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	links, err := links.GetLinksByUserID(ctx, h.DB, userID)
	if err != nil {
		log.Printf("failed to get links for user %d: %v", userID, err)
		http.Error(w, "failed to get links", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(links); err != nil {
		log.Printf("failed to encode links response: %v", err)
	}
}
