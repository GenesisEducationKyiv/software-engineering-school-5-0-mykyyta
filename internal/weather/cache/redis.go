package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"weatherApi/internal/weather"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

var ErrCacheMiss = errors.New("cache miss")

const cachePrefix = "weather"

func NewRedisCache(client *redis.Client) RedisCache {
	return RedisCache{client: client}
}

func normalizeCity(city string) string {
	return strings.ToLower(strings.TrimSpace(city))
}

func (r RedisCache) key(city, provider string) string {
	city = normalizeCity(city)
	return fmt.Sprintf("%s:%s:%s", cachePrefix, city, provider)
}

func (r RedisCache) Set(ctx context.Context, city, provider string, report weather.Report, ttl time.Duration) error {
	key := r.key(city, provider)

	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

func (r RedisCache) Get(ctx context.Context, city, provider string) (weather.Report, error) {
	key := r.key(city, provider)

	data, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return weather.Report{}, ErrCacheMiss
	}
	if err != nil {
		return weather.Report{}, fmt.Errorf("redis get error: %w", err)
	}

	var rep weather.Report
	if err := json.Unmarshal([]byte(data), &rep); err != nil {
		return weather.Report{}, fmt.Errorf("failed to unmarshal report: %w", err)
	}
	return rep, nil
}
