package internal

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	DB *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		DB: db,
	}
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request CreateLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	request.URL = strings.TrimSpace(request.URL)

	if request.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	shortCode, err := GenerateShortCode(6)
	if err != nil {
		http.Error(w, "failed to generate short code", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	link, err := CreateLink(
		ctx,
		h.DB,
		shortCode,
		request.URL,
	)

	if err != nil {
		http.Error(w, "failed to create link", http.StatusInternalServerError)
		return
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
		http.NotFound(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	link, err := GetLinkByShortCode(ctx, h.DB, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "failed to resolve link", http.StatusInternalServerError)
		return
	}

	if !link.IsActive {
		http.Error(w, "link is disabled", http.StatusGone)
		return
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		http.Error(w, "link has expired", http.StatusGone)
		return
	}

	http.Redirect(w, r, link.LongURL, http.StatusFound)
}