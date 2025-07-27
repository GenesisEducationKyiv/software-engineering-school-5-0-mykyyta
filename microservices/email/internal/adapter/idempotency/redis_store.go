package idempotency

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Status string

const (
	statusProcessing Status = "processing"
	statusDone       Status = "done"
)

type RedisStore struct {
	client    *redis.Client
	ttl       time.Duration
	namespace string
	logger    *log.Logger
}

func NewRedisStore(client *redis.Client, ttl time.Duration, logger *log.Logger) *RedisStore {
	return &RedisStore{
		client:    client,
		ttl:       ttl,
		namespace: "idemp:",
		logger:    logger,
	}
}

func (r *RedisStore) key(id string) string {
	return r.namespace + id
}

func (r *RedisStore) IsProcessed(ctx context.Context, messageID string) (bool, error) {
	val, err := r.client.Get(ctx, r.key(messageID)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		r.logger.Printf("[WARN] Redis unavailable in IsProcessed: %v", err)
		return false, nil
	}
	return Status(val) == statusDone, nil
}

func (r *RedisStore) MarkAsProcessing(ctx context.Context, messageID string) (bool, error) {
	ok, err := r.client.SetNX(ctx, r.key(messageID), string(statusProcessing), r.ttl).Result()
	if err != nil {
		r.logger.Printf("[WARN] Redis unavailable in MarkAsProcessing: %v", err)
		return true, nil
	}
	return ok, nil
}

func (r *RedisStore) MarkAsProcessed(ctx context.Context, messageID string) error {
	if err := r.client.Set(ctx, r.key(messageID), string(statusDone), r.ttl).Err(); err != nil {
		r.logger.Printf("[WARN] Redis unavailable in MarkAsProcessed: %v", err)
		return nil
	}
	return nil
}

func (r *RedisStore) ClearProcessing(ctx context.Context, messageID string) error {
	if err := r.client.Del(ctx, r.key(messageID)).Err(); err != nil {
		r.logger.Printf("[WARN] Redis unavailable in ClearProcessing: %v", err)
		return nil
	}
	return nil
}

func (r *RedisStore) GetStatus(ctx context.Context, messageID string) (Status, error) {
	val, err := r.client.Get(ctx, r.key(messageID)).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		r.logger.Printf("[WARN] Redis unavailable in GetStatus: %v", err)
		return "", nil
	}
	return Status(val), nil
}
