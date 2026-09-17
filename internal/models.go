package internal

import "time"

type link struct {
	ID        int64      `json:"id"`
	ShortCode string     `json:"short_code"`
	LongURL   string     `json:"long_url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
}

type CreateLinkRequest struct {
	URL string `json:"URL"`
}