package apikeys

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Authenticate(
	ctx context.Context,
	db *pgxpool.Pool,
	rawKey string,
) (int64, error) {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	apiKey, err := GetAPIKeyByHash(ctx, db, keyHash)
	if err != nil {
		return 0, err
	}

	if err := UpdateLastUsed(ctx, db, apiKey.ID); err != nil {
		return 0, err
	}

	return apiKey.UserID, nil
}
