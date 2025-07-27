package idempotency

import (
	"context"
	"errors"
	"time"

	"email/pkg/logger"

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
}

func NewRedisStore(client *redis.Client, ttl time.Duration) *RedisStore {
	return &RedisStore{
		client:    client,
		ttl:       ttl,
		namespace: "idemp:",
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
		logger.From(ctx).Errorw("Redis error checking if message is processed", "err", err)
		return false, err
	}
	return Status(val) == statusDone, nil
}

func (r *RedisStore) MarkAsProcessing(ctx context.Context, messageID string) (bool, error) {
	ok, err := r.client.SetNX(ctx, r.key(messageID), string(statusProcessing), r.ttl).Result()
	if err != nil {
		logger.From(ctx).Errorw("Redis error marking message as processing", "err", err)
		return true, err
	}
	return ok, nil
}

func (r *RedisStore) MarkAsProcessed(ctx context.Context, messageID string) error {
	if err := r.client.Set(ctx, r.key(messageID), string(statusDone), r.ttl).Err(); err != nil {
		logger.From(ctx).Errorw("Redis error marking message as processed", "err", err)
		return err
	}
	return nil
}

func (r *RedisStore) ClearProcessing(ctx context.Context, messageID string) error {
	if err := r.client.Del(ctx, r.key(messageID)).Err(); err != nil {
		logger.From(ctx).Errorw("Redis error clearing processing lock", "err", err)
		return err
	}
	return nil
}

func (r *RedisStore) GetStatus(ctx context.Context, messageID string) (Status, error) {
	val, err := r.client.Get(ctx, r.key(messageID)).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", nil
	}
	return Status(val), nil
}
