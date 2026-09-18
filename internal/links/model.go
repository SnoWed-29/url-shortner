package links

import "time"

type link struct {
	ID        int64      `json:"id"`
	UserID    *int64     `json:"user_id,omitempty"`
	ShortCode string     `json:"short_code"`
	LongURL   string     `json:"long_url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `json:"is_active"`
}

type Link struct {
	ID        int64
	UserID    *int64
	ShortCode string
	LongURL   string
	CreatedAt time.Time
	ExpiresAt *time.Time
	IsActive  bool
}

type CreateLinkRequest struct {
	URL         string     `json:"URL"`
	CustomAlias string     `json:"custom_alias,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}
