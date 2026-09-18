package cache

import (
	"context"
	"log"
	"time"

	"github.com/SnoWed-29/url-shortener/internal/links"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func CacheLink(
	ctx context.Context,
	redisClient *redis.Client,
	link links.Link,
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
