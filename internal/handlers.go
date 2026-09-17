package internal

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"context"
	
	"github.com/jackc/pgx/v5/pgxpool"
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