package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"weatherApi/internal/domain"

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

func makeKey(prefix, keyType, city, provider string) string {
	return fmt.Sprintf("%s:%s:%s:%s", prefix, keyType, normalizeCity(city), provider)
}

func (r RedisCache) key(city, provider string) string {
	return makeKey(cachePrefix, "report", city, provider)
}

func (r RedisCache) notFoundKey(city, provider string) string {
	return makeKey(cachePrefix, "notfound", city, provider)
}

func (r RedisCache) Set(ctx context.Context, city, provider string, report domain.Report, ttl time.Duration) error {
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

func (r RedisCache) Get(ctx context.Context, city, provider string) (domain.Report, error) {
	key := r.key(city, provider)

	data, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return domain.Report{}, ErrCacheMiss
	}
	if err != nil {
		return domain.Report{}, fmt.Errorf("redis get error: %w", err)
	}

	var rep domain.Report
	if err := json.Unmarshal([]byte(data), &rep); err != nil {
		return domain.Report{}, fmt.Errorf("failed to unmarshal report: %w", err)
	}
	return rep, nil
}

func (r RedisCache) SetCityNotFound(ctx context.Context, city, provider string, ttl time.Duration) error {
	key := r.notFoundKey(city, provider)
	return r.client.Set(ctx, key, "1", ttl).Err()
}

func (r RedisCache) GetCityNotFound(ctx context.Context, city, provider string) (bool, error) {
	key := r.notFoundKey(city, provider)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}
