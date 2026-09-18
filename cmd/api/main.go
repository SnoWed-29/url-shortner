package main

import (
	"context"
	"fmt"
	"github.com/SnoWed-29/url-shortener/internal"
	"github.com/SnoWed-29/url-shortener/internal/auth"
	"github.com/SnoWed-29/url-shortener/internal/cache"
	"github.com/SnoWed-29/url-shortener/internal/config"
	"github.com/SnoWed-29/url-shortener/internal/database"
	"log"
	"net/http"
	"time"
)

func main() {

	config := config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.NewPostgresPool(ctx, config.PostgresDSN)

	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	log.Println("connected to PostgreSQL")

	redisClient, err := cache.NewRedisClient(config.RedisAddr)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	log.Println("connected to Redis")

	handler := internal.NewHandler(db, redisClient, config.JWTSecret)
	authMiddleware := auth.AuthMiddleware(config.JWTSecret)
	// Routes
	http.Handle(
		"POST /api/v1/links",
		authMiddleware(http.HandlerFunc(handler.CreateLink)),
	)

	http.Handle(
		"GET /api/v1/links",
		authMiddleware(http.HandlerFunc(handler.ListLinks)),
	)

	http.Handle(
		"DELETE /api/v1/links/",
		authMiddleware(http.HandlerFunc(handler.DeleteLink)),
	)

	http.Handle(
		"GET /api/v1/links/",
		authMiddleware(http.HandlerFunc(handler.GetLink)),
	)
	http.Handle(
		"PATCH /api/v1/links/",
		authMiddleware(http.HandlerFunc(handler.UpdateLink)),
	)
	http.HandleFunc("/", handler.Redirect)
	http.HandleFunc("/api/v1/auth/register", handler.Register)
	http.HandleFunc("/api/v1/auth/login", handler.Login)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	addr := ":" + config.AppPort

	log.Printf("API Server Listening on %s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
