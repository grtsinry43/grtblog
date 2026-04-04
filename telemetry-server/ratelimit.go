package main

import (
	"context"
	"sync"
	"time"
)

// RateLimiter enforces per-instance-id rate limiting in memory.
// Allows at most 1 request per hour per instance_id.
type RateLimiter struct {
	mu       sync.Mutex
	seen     map[string]time.Time
	interval time.Duration
}

func NewRateLimiter(ctx context.Context) *RateLimiter {
	rl := &RateLimiter{
		seen:     make(map[string]time.Time),
		interval: time.Hour,
	}
	go rl.cleanup(ctx)
	return rl
}

// Allow returns true if the instance_id is allowed to proceed.
func (rl *RateLimiter) Allow(instanceID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if last, ok := rl.seen[instanceID]; ok {
		if time.Since(last) < rl.interval {
			return false
		}
	}
	rl.seen[instanceID] = time.Now()
	return true
}

// cleanup removes stale entries periodically, stops on context cancellation.
func (rl *RateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-rl.interval)
			for k, v := range rl.seen {
				if v.Before(cutoff) {
					delete(rl.seen, k)
				}
			}
			rl.mu.Unlock()
		}
	}
}
