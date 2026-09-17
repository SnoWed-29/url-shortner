package internal

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func CacheLink(
	ctx context.Context,
	redisClient *redis.Client,
	link Link,
) error {
	ttl := 24 * time.Hour

	if link.ExpiresAt != nil {
		ttl = time.Until(*link.ExpiresAt)

		if ttl <= 0 {
			return nil
		}
	}

	return redisClient.Set(
		ctx,
		link.ShortCode,
		link.LongURL,
		ttl,
	).Err()
}

func DeleteCachedLink(
	ctx context.Context,
	redisClient *redis.Client,
	shortCode string,
) error {
	return redisClient.Del(ctx, shortCode).Err()
}

func LogCacheError(operation string, shortCode string, err error) {
	if err != nil {
		log.Printf(
			"redis %s failed for short_code=%s: %v",
			operation,
			shortCode,
			err,
		)
	}
}
