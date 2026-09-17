package tests

import (
	"github.com/SnoWed-29/url-shortener/internal"
	"testing"
	"time"
)

func TestLinkCacheTTL(t *testing.T) {
	t.Run("non-expiring link", func(t *testing.T) {
		link := internal.Link{}

		ttl := 24 * time.Hour

		if link.ExpiresAt != nil {
			ttl = time.Until(*link.ExpiresAt)
		}

		if ttl != 24*time.Hour {
			t.Fatalf("expected 24h TTL, got %v", ttl)
		}
	})

	t.Run("expiring link", func(t *testing.T) {
		expiresAt := time.Now().Add(10 * time.Minute)

		link := internal.Link{
			ExpiresAt: &expiresAt,
		}

		ttl := time.Until(*link.ExpiresAt)

		if ttl <= 0 {
			t.Fatal("expected positive TTL")
		}

		if ttl > 10*time.Minute {
			t.Fatalf("unexpected TTL: %v", ttl)
		}
	})
}
