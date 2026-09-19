package main

import (
	"context"
	"fmt"
	"github.com/SnoWed-29/url-shortener/internal/apikeys"
	"github.com/SnoWed-29/url-shortener/internal/auth"
	"github.com/SnoWed-29/url-shortener/internal/cache"
	"github.com/SnoWed-29/url-shortener/internal/config"
	"github.com/SnoWed-29/url-shortener/internal/database"
	"github.com/SnoWed-29/url-shortener/internal/links"
	"github.com/SnoWed-29/url-shortener/internal/ratelimit"
	"github.com/SnoWed-29/url-shortener/internal/users"

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

	rateLimiter := ratelimit.NewLimiter(redisClient)
	linkHandler := links.NewHandler(db, redisClient, config.JWTSecret)
	userHandler := users.Handler{DB: db, JWTSecret: config.JWTSecret}
	authMiddleware := auth.AuthMiddleware(config.JWTSecret)
	apiKeyMiddleware := apikeys.Middleware(db)

	authRateLimit := ratelimit.Middleware(
		rateLimiter,
		5,
		time.Minute,
	)

	apiKeyHandler := &apikeys.Handler{
		DB: db,
	}

	// Routes
	http.Handle(
		"POST /api/v1/links",
		authMiddleware(http.HandlerFunc(linkHandler.CreateLink)),
	)

	http.Handle(
		"POST /api/v1/links/api-key",
		apiKeyMiddleware(http.HandlerFunc(linkHandler.CreateLink)),
	)

	http.Handle(
		"GET /api/v1/links",
		authMiddleware(http.HandlerFunc(linkHandler.ListLinks)),
	)

	http.Handle(
		"DELETE /api/v1/links/",
		authMiddleware(http.HandlerFunc(linkHandler.DeleteLink)),
	)

	http.Handle(
		"GET /api/v1/links/",
		authMiddleware(http.HandlerFunc(linkHandler.GetLink)),
	)
	http.Handle(
		"PATCH /api/v1/links/",
		authMiddleware(http.HandlerFunc(linkHandler.UpdateLink)),
	)
	http.HandleFunc("/", linkHandler.Redirect)
	http.Handle("/api/v1/auth/register", authRateLimit(http.HandlerFunc(userHandler.Register)))
	http.Handle("/api/v1/auth/login", authRateLimit(http.HandlerFunc(userHandler.Login)))

	http.Handle(
		"POST /api/v1/api-keys",
		authMiddleware(http.HandlerFunc(apiKeyHandler.Create)),
	)

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
