package apikeys

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateAPIKey(
	ctx context.Context,
	db *pgxpool.Pool,
	userID int64,
	name string,
	keyHash string,
) (APIKey, error) {
	var key APIKey

	err := db.QueryRow(
		ctx,
		`
		INSERT INTO api_keys (
			user_id,
			name,
			key_hash
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			user_id,
			name,
			created_at,
			last_used_at
		`,
		userID,
		name,
		keyHash,
	).Scan(
		&key.ID,
		&key.UserID,
		&key.Name,
		&key.CreatedAt,
		&key.LastUsedAt,
	)

	if err != nil {
		return APIKey{}, fmt.Errorf("create api key: %w", err)
	}

	return key, nil
}

func GetAPIKeyByHash(
	ctx context.Context,
	db *pgxpool.Pool,
	keyHash string,
) (APIKey, error) {
	var key APIKey

	err := db.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			name,
			created_at,
			last_used_at
		FROM api_keys
		WHERE key_hash = $1
		`,
		keyHash,
	).Scan(
		&key.ID,
		&key.UserID,
		&key.Name,
		&key.CreatedAt,
		&key.LastUsedAt,
	)

	if err != nil {
		return APIKey{}, fmt.Errorf("get api key: %w", err)
	}

	return key, nil
}

func UpdateLastUsed(
	ctx context.Context,
	db *pgxpool.Pool,
	id int64,
) error {
	_, err := db.Exec(
		ctx,
		`
		UPDATE api_keys
		SET last_used_at = NOW()
		WHERE id = $1
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf("update api key usage: %w", err)
	}

	return nil
}
