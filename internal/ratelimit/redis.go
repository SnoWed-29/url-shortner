package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	client *redis.Client
}

func NewLimiter(client *redis.Client) *Limiter {
	return &Limiter{
		client: client,
	}
}

func (l *Limiter) Allow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (bool, error) {
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("increment rate limit: %w", err)
	}

	if count == 1 {
		if err := l.client.Expire(ctx, key, window).Err(); err != nil {
			return false, fmt.Errorf("set rate limit expiration: %w", err)
		}
	}

	return count <= int64(limit), nil
}
