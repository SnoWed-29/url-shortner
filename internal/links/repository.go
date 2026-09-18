package links

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateLink(
	ctx context.Context,
	db *pgxpool.Pool,
	userID int64,
	shortCode string,
	longURL string,
	expiresAt *time.Time,
) (Link, error) {
	var link Link

	err := db.QueryRow(
		ctx,
		`
		INSERT INTO links (
			user_id,
			short_code,
			long_url,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			user_id,
			short_code,
			long_url,
			created_at,
			expires_at,
			is_active
		`,
		userID,
		shortCode,
		longURL,
		expiresAt,
	).Scan(
		&link.ID,
		&link.UserID,
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

func GetLinksByUserID(
	ctx context.Context,
	db *pgxpool.Pool,
	userID int64,
) ([]Link, error) {
	rows, err := db.Query(
		ctx,
		`
		SELECT
			id,
			user_id,
			short_code,
			long_url,
			created_at,
			expires_at,
			is_active
		FROM links
		WHERE user_id = $1
		ORDER BY created_at DESC
		`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get links by user: %w", err)
	}
	defer rows.Close()

	links := make([]Link, 0)

	for rows.Next() {
		var link Link

		err := rows.Scan(
			&link.ID,
			&link.UserID,
			&link.ShortCode,
			&link.LongURL,
			&link.CreatedAt,
			&link.ExpiresAt,
			&link.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate links: %w", err)
	}

	return links, nil
}

func DisableLink(
	ctx context.Context,
	db *pgxpool.Pool,
	userID int64,
	shortCode string,
) error {
	result, err := db.Exec(
		ctx,
		`
		UPDATE links
		SET is_active = FALSE
		WHERE user_id = $1
		  AND short_code = $2
		  AND is_active = TRUE
		`,
		userID,
		shortCode,
	)
	if err != nil {
		return fmt.Errorf("disable link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func GetLinkByUserIDAndShortCode(
	ctx context.Context,
	db *pgxpool.Pool,
	userID int64,
	shortCode string,
) (Link, error) {
	var link Link

	err := db.QueryRow(
		ctx,
		`
        SELECT
            id,
            user_id,
            short_code,
            long_url,
            created_at,
            expires_at,
            is_active
        FROM links
        WHERE user_id = $1
          AND short_code = $2
        `,
		userID,
		shortCode,
	).Scan(
		&link.ID,
		&link.UserID,
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

func UpdateLinkExpiration(
	ctx context.Context,
	db *pgxpool.Pool,
	userID int64,
	shortCode string,
	expiresAt *time.Time,
) error {
	result, err := db.Exec(
		ctx,
		`
		UPDATE links
		SET expires_at = $1
		WHERE user_id = $2
		  AND short_code = $3
		`,
		expiresAt,
		userID,
		shortCode,
	)
	if err != nil {
		return fmt.Errorf("update link expiration: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
