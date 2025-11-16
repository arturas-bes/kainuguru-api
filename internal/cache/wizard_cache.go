package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

const (
	// WizardSessionTTL is the time-to-live for wizard sessions (30 minutes per constitution)
	WizardSessionTTL = 30 * time.Minute

	// WizardIdempotencyTTL is the time-to-live for idempotency keys (24 hours per data-model.md)
	WizardIdempotencyTTL = 24 * time.Hour
)

// WizardCache handles Redis operations for wizard sessions
type WizardCache struct {
	redis *RedisClient
}

// NewWizardCache creates a new wizard cache instance
func NewWizardCache(redis *RedisClient) *WizardCache {
	return &WizardCache{
		redis: redis,
	}
}

// SaveSession stores a wizard session in Redis with 30-minute TTL
// Key pattern: wizard:session:{session_id}
func (c *WizardCache) SaveSession(ctx context.Context, session *models.WizardSession) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	key := c.sessionKey(session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := c.redis.Set(ctx, key, data, WizardSessionTTL); err != nil {
		return fmt.Errorf("failed to save session to Redis: %w", err)
	}

	return nil
}

// GetSession retrieves a wizard session from Redis
// Returns nil if session doesn't exist or has expired
func (c *WizardCache) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.WizardSession, error) {
	key := c.sessionKey(sessionID)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		// Key not found or expired
		return nil, nil
	}

	var session models.WizardSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	// Check if session has expired (defensive check)
	if session.IsExpired() {
		// Clean up expired session
		_ = c.DeleteSession(ctx, sessionID)
		return nil, nil
	}

	return &session, nil
}

// DeleteSession removes a wizard session from Redis
func (c *WizardCache) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	key := c.sessionKey(sessionID)
	if err := c.redis.Del(ctx, key); err != nil {
		return fmt.Errorf("failed to delete session from Redis: %w", err)
	}
	return nil
}

// ExtendSessionTTL resets the session TTL to 30 minutes (used when user makes progress)
func (c *WizardCache) ExtendSessionTTL(ctx context.Context, sessionID uuid.UUID) error {
	key := c.sessionKey(sessionID)
	exists, err := c.redis.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("session not found")
	}

	if err := c.redis.Expire(ctx, key, WizardSessionTTL); err != nil {
		return fmt.Errorf("failed to extend session TTL: %w", err)
	}

	return nil
}

// SaveIdempotencyKey stores an idempotency key result with 24-hour TTL
// Key pattern: wizard:idempotency:{key}
// Value: session_id or result payload
func (c *WizardCache) SaveIdempotencyKey(ctx context.Context, key string, sessionID uuid.UUID) error {
	redisKey := c.idempotencyKey(key)
	if err := c.redis.Set(ctx, redisKey, sessionID.String(), WizardIdempotencyTTL); err != nil {
		return fmt.Errorf("failed to save idempotency key: %w", err)
	}
	return nil
}

// GetIdempotencyKey retrieves the session ID associated with an idempotency key
// Returns empty UUID if key doesn't exist
func (c *WizardCache) GetIdempotencyKey(ctx context.Context, key string) (uuid.UUID, error) {
	redisKey := c.idempotencyKey(key)
	data, err := c.redis.Get(ctx, redisKey)
	if err != nil {
		// Key not found or expired
		return uuid.Nil, nil
	}

	sessionID, err := uuid.Parse(data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse session ID from idempotency key: %w", err)
	}

	return sessionID, nil
}

// sessionKey generates the Redis key for a wizard session
func (c *WizardCache) sessionKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("wizard:session:%s", sessionID.String())
}

// idempotencyKey generates the Redis key for an idempotency key
func (c *WizardCache) idempotencyKey(key string) string {
	return fmt.Sprintf("wizard:idempotency:%s", key)
}
