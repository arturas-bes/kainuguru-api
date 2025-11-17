package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter provides rate limiting using Redis
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a new RateLimiter instance
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		client: client,
	}
}

// CheckRateLimit checks if the user has exceeded the rate limit
// Returns true if request is allowed, false if rate limit exceeded
// key: unique identifier for the rate limit (e.g., "wizard:start:{userID}")
// maxRequests: maximum number of requests allowed in the window
// window: time window for the rate limit
func (r *RateLimiter) CheckRateLimit(ctx context.Context, key string, maxRequests int, window time.Duration) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-window).UnixMilli()

	pipe := r.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count requests in current window
	countCmd := pipe.ZCount(ctx, key, fmt.Sprintf("%d", windowStart), "+inf")

	// Add current request timestamp
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiration on the key to auto-cleanup
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("rate limit check failed: %w", err)
	}

	count := countCmd.Val()

	// Allow request if count is less than max
	// Note: count includes the request we just added, so we check <= maxRequests
	return count <= int64(maxRequests), nil
}

// GetRemainingRequests returns the number of remaining requests in the current window
func (r *RateLimiter) GetRemainingRequests(ctx context.Context, key string, maxRequests int, window time.Duration) (int, error) {
	now := time.Now()
	windowStart := now.Add(-window).UnixMilli()

	// Count requests in current window
	count, err := r.client.ZCount(ctx, key, fmt.Sprintf("%d", windowStart), "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get request count: %w", err)
	}

	remaining := maxRequests - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}
