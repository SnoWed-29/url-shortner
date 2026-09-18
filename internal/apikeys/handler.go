package apikeys

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/SnoWed-29/url-shortener/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

type CreateRequest struct {
	Name string `json:"name"`
}

type CreateResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if len(req.Name) > 100 {
		http.Error(w, "name is too long", http.StatusBadRequest)
		return
	}

	key, keyHash, err := GenerateKey()
	if err != nil {
		http.Error(w, "failed to generate api key", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	apiKey, err := CreateAPIKey(
		ctx,
		h.DB,
		userID,
		req.Name,
		keyHash,
	)
	if err != nil {
		http.Error(w, "failed to create api key", http.StatusInternalServerError)
		return
	}

	response := CreateResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       key,
		CreatedAt: apiKey.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
