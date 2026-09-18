package ratelimit

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

func Middleware(
	limiter *Limiter,
	limit int,
	window time.Duration,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			key := "rate-limit:" + r.URL.Path + ":" + ip

			allowed, err := limiter.Allow(
				r.Context(),
				key,
				limit,
				window,
			)
			if err != nil {
				log.Printf("rate limiter error: %v", err)

				// Fail open: Redis failure should not take down the API.
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", "60")
				http.Error(
					w,
					"rate limit exceeded",
					http.StatusTooManyRequests,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}
