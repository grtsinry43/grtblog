package router

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/redis/go-redis/v9"
)

// redisFiberStorage adapts go-redis to fiber.Storage for cross-instance rate limiting.
type redisFiberStorage struct {
	client *redis.Client
	prefix string
}

func newRedisFiberStorage(client *redis.Client, prefix string) fiber.Storage {
	if client == nil {
		return nil
	}
	return &redisFiberStorage{client: client, prefix: prefix}
}

func (s *redisFiberStorage) key(k string) string {
	return s.prefix + "fiber:limiter:" + k
}

func (s *redisFiberStorage) Get(key string) ([]byte, error) {
	val, err := s.client.Get(context.Background(), s.key(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (s *redisFiberStorage) Set(key string, val []byte, exp time.Duration) error {
	if key == "" || len(val) == 0 {
		return nil
	}
	ctx := context.Background()
	fullKey := s.key(key)
	if exp > 0 {
		return s.client.Set(ctx, fullKey, val, exp).Err()
	}
	return s.client.Set(ctx, fullKey, val, 0).Err()
}

func (s *redisFiberStorage) Delete(key string) error {
	return s.client.Del(context.Background(), s.key(key)).Err()
}

func (s *redisFiberStorage) Reset() error {
	ctx := context.Background()
	iter := s.client.Scan(ctx, 0, s.prefix+"fiber:limiter:*", 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= 100 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (s *redisFiberStorage) Close() error {
	return nil
}

func limiterStorage(deps Dependencies) fiber.Storage {
	if deps.Redis == nil {
		return nil
	}
	return newRedisFiberStorage(deps.Redis, deps.Config.Redis.Prefix)
}

func newRateLimiter(deps Dependencies, cfg limiter.Config) fiber.Handler {
	cfg.Storage = limiterStorage(deps)
	return limiter.New(cfg)
}
