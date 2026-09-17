package main 

import (
	"fmt"
	"log"
	"net/http"
	"context"
	"time"
	"github.com/SnoWed-29/url-shortener/internal"
)

func main() {

	config := internal.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := internal.NewPostgresPool(ctx, config.PostgresDSN)

	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	log.Println("connected to PostgreSQL")
	
	handler := internal.NewHandler(db)

	http.HandleFunc("/api/v1/links", handler.CreateLink)
	http.HandleFunc("/", handler.Redirect)
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