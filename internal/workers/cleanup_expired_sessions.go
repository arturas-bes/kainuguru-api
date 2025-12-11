package workers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/monitoring"
	"github.com/redis/go-redis/v9"
)

// CleanupExpiredSessionsWorker removes expired wizard sessions from Redis
// Runs hourly to delete wizard:session:* keys where expires_at < NOW
//
// This worker helps:
// - Free up Redis memory by removing stale sessions
// - Ensure expired sessions don't linger indefinitely
// - Track session expiration metrics for monitoring
type CleanupExpiredSessionsWorker struct {
	redis     *redis.Client
	logger    *slog.Logger
	batchSize int
}

// NewCleanupExpiredSessionsWorker creates a new worker instance
func NewCleanupExpiredSessionsWorker(redisClient *redis.Client) *CleanupExpiredSessionsWorker {
	return &CleanupExpiredSessionsWorker{
		redis:     redisClient,
		logger:    slog.Default().With("worker", "cleanup_expired_sessions"),
		batchSize: 100, // Scan 100 keys at a time
	}
}

// Run executes the cleanup job
// This is the main entry point called by the scheduler
func (w *CleanupExpiredSessionsWorker) Run(ctx context.Context) error {
	w.logger.Info("starting expired wizard sessions cleanup job")
	startTime := time.Now()

	// Track worker execution metrics
	defer func() {
		if r := recover(); r != nil {
			monitoring.WizardWorkerRunsTotal.WithLabelValues("cleanup_expired_sessions", "error").Inc()
			w.logger.Error("worker panicked", "panic", r)
		}
	}()

	// Scan Redis for all wizard session keys
	var cursor uint64
	var totalScanned, totalDeleted int
	pattern := "wizard:session:*"

	for {
		// SCAN command returns keys in batches
		var keys []string
		var err error
		keys, cursor, err = w.redis.Scan(ctx, cursor, pattern, int64(w.batchSize)).Result()
		if err != nil {
			w.logger.Error("redis scan failed",
				"pattern", pattern,
				"error", err)
			monitoring.WizardWorkerRunsTotal.WithLabelValues("cleanup_expired_sessions", "error").Inc()
			return fmt.Errorf("failed to scan redis keys: %w", err)
		}

		totalScanned += len(keys)

		// Check each key for expiration
		for _, key := range keys {
			// Note: Redis TTL automatically cleans up sessions after 30 minutes
			// This worker is a safety net for sessions that might have inconsistent TTL

			// Check Redis TTL first (fast path)
			ttl, err := w.redis.TTL(ctx, key).Result()
			if err != nil {
				w.logger.Warn("failed to get TTL",
					"key", key,
					"error", err)
				continue
			}

			// If TTL is expired (< 0) or key doesn't exist (-2), delete it
			if ttl < 0 {
				w.logger.Info("deleting session with expired/missing TTL",
					"key", key,
					"ttl", ttl)

				if err := w.redis.Del(ctx, key).Err(); err != nil {
					w.logger.Error("failed to delete session",
						"key", key,
						"error", err)
				} else {
					totalDeleted++
				}
				continue
			}

			// For sessions with valid TTL, check expires_at timestamp as secondary validation
			var session struct {
				ExpiresAt time.Time `json:"expires_at"`
			}
			if err := w.redis.Get(ctx, key).Scan(&session); err != nil {
				// Can't parse session - rely on TTL cleanup
				continue
			}

			// If session timestamp is expired but TTL isn't, force delete
			if time.Now().After(session.ExpiresAt) {
				w.logger.Info("deleting session with stale expires_at",
					"key", key,
					"expires_at", session.ExpiresAt,
					"ttl_remaining", ttl)

				if err := w.redis.Del(ctx, key).Err(); err != nil {
					w.logger.Error("failed to delete stale session",
						"key", key,
						"error", err)
				} else {
					totalDeleted++
				}
			}
		}

		// If cursor is 0, we've scanned all keys
		if cursor == 0 {
			break
		}
	}

	// Track completion metrics
	duration := time.Since(startTime)
	monitoring.WizardWorkerRunsTotal.WithLabelValues("cleanup_expired_sessions", "success").Inc()
	monitoring.WizardWorkerDurationSeconds.WithLabelValues("cleanup_expired_sessions").Observe(duration.Seconds())

	w.logger.Info("completed expired sessions cleanup",
		"total_scanned", totalScanned,
		"total_deleted", totalDeleted,
		"duration_seconds", duration.Seconds())

	return nil
}

// SessionID extracts session UUID from Redis key
// Pattern: wizard:session:{uuid}
func extractSessionID(key string) (uuid.UUID, error) {
	// Remove "wizard:session:" prefix (16 characters)
	if len(key) < 16+36 {
		return uuid.Nil, fmt.Errorf("invalid key format: %s", key)
	}

	uuidStr := key[16:] // Skip "wizard:session:"
	return uuid.Parse(uuidStr)
}
