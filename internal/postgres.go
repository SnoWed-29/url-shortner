package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Link struct {
	ID        int64
	ShortCode string
	LongURL   string
	CreatedAt time.Time
	ExpiresAt *time.Time
	IsActive  bool
}

func NewPostgresPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func CreateLink(
	ctx context.Context,
	db *pgxpool.Pool,
	shortCode string,
	longURL string,
	expiresAt *time.Time,
) (Link, error) {
	var link Link

	err := db.QueryRow(
		ctx,
		`
		INSERT INTO links (
			short_code,
			long_url,
			expires_at
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			short_code,
			long_url,
			created_at,
			expires_at,
			is_active
		`,
		shortCode,
		longURL,
		expiresAt,
	).Scan(
		&link.ID,
		&link.ShortCode,
		&link.LongURL,
		&link.CreatedAt,
		&link.ExpiresAt,
		&link.IsActive,
	)

	if err != nil {
		return Link{}, fmt.Errorf("create link: %w", err)
	}

	return link, nil
}

func GetLinkByShortCode(
	ctx context.Context,
	db *pgxpool.Pool,
	shortCode string,
) (Link, error) {
	var link Link

	err := db.QueryRow(
		ctx,
		`
		SELECT id, short_code, long_url, created_at, expires_at, is_active
		FROM links
		WHERE short_code = $1
		`,
		shortCode,
	).Scan(
		&link.ID,
		&link.ShortCode,
		&link.LongURL,
		&link.CreatedAt,
		&link.ExpiresAt,
		&link.IsActive,
	)

	if err != nil {
		return Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

func CreateUser(
	ctx context.Context,
	db *pgxpool.Pool,
	email string,
	passwordHash string,
) (User, error) {
	var user User

	err := db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			email,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at
		`,
		email,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func GetUserByEmail(
	ctx context.Context,
	db *pgxpool.Pool,
	email string,
) (User, error) {
	var user User

	err := db.QueryRow(
		ctx,
		`
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}
