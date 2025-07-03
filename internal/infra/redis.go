package infra

import (
	"context"
	"fmt"
	"time"

	"weatherApi/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	if !cfg.Cache.Enabled {
		return nil, nil
	}

	opts, err := redis.ParseURL(cfg.Cache.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("redis connection failed: %w, also failed to close client: %w", err, closeErr)
		}
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return client, nil
}
