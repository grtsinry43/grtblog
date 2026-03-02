package federation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type RedisRateLimiter struct {
	client *redis.Client
	prefix string
}

func NewRedisRateLimiter(client *redis.Client, prefix string) *RedisRateLimiter {
	return &RedisRateLimiter{client: client, prefix: prefix}
}

func (l *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if l == nil || l.client == nil {
		return false, nil
	}
	if limit <= 0 {
		return true, nil
	}
	if window <= 0 {
		window = time.Minute
	}
	nowWindow := time.Now().UTC().Format("200601021504")
	redisKey := fmt.Sprintf("%sfed:rl:%s:%s", l.prefix, key, nowWindow)
	pipe := l.client.TxPipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window+10*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return incr.Val() <= int64(limit), nil
}

// InMemoryRateLimiter provides a simple in-memory fallback when Redis is unavailable.
type InMemoryRateLimiter struct {
	mu      sync.Mutex
	windows map[string]*memWindow
}

type memWindow struct {
	count   int
	expires time.Time
}

func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	return &InMemoryRateLimiter{windows: make(map[string]*memWindow)}
}

func (l *InMemoryRateLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	if window <= 0 {
		window = time.Minute
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	// Lazy cleanup: remove expired entries periodically.
	if len(l.windows) > 1000 {
		for k, w := range l.windows {
			if now.After(w.expires) {
				delete(l.windows, k)
			}
		}
	}

	w, ok := l.windows[key]
	if !ok || now.After(w.expires) {
		l.windows[key] = &memWindow{count: 1, expires: now.Add(window)}
		return true, nil
	}
	w.count++
	return w.count <= limit, nil
}
