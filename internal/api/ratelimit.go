package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter supports Redis-backed rate limiting with an automatic
// in-memory fallback when Redis is unavailable (rdb == nil).
type RateLimiter struct {
	rdb    *redis.Client
	rate   int
	window time.Duration
	// in-memory fallback
	mu       sync.Mutex
	visitors map[string]*visitor
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter. When rdb is non-nil, Redis is
// used for distributed counting. When rdb is nil, an in-memory map with
// a background cleanup goroutine is used instead.
func NewRateLimiter(rdb *redis.Client, rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		rdb:      rdb,
		rate:     rate,
		window:   window,
		visitors: make(map[string]*visitor),
	}
	if rdb == nil {
		go rl.cleanup()
	}
	return rl
}

// Allow checks whether the given key is within the rate limit.
func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	if rl.rdb != nil {
		return rl.allowRedis(ctx, key)
	}
	return rl.allowMemory(key)
}

func (rl *RateLimiter) allowRedis(ctx context.Context, key string) bool {
	redisKey := fmt.Sprintf("rl:%s", key)
	pipe := rl.rdb.Pipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, rl.window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		// Redis error — fail open for availability
		return true
	}
	return incr.Val() <= int64(rl.rate)
}

func (rl *RateLimiter) allowMemory(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists || time.Since(v.lastSeen) > rl.window {
		rl.visitors[key] = &visitor{count: 1, lastSeen: time.Now()}
		return true
	}
	v.lastSeen = time.Now()
	v.count++
	return v.count <= rl.rate
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns HTTP middleware that enforces the given rate limiter
// per client IP (X-Forwarded-For with RemoteAddr fallback).
func RateLimit(rl *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.RemoteAddr
			}
			if !rl.Allow(r.Context(), ip) {
				WriteError(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests, please try again later")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
