package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

type RedisClient struct {
	client *redis.Client
	config Config
}

func NewRedis(cfg Config) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: cfg.MaxRetries,
		PoolSize:   cfg.PoolSize,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Int("db", cfg.DB).
		Int("pool_size", cfg.PoolSize).
		Msg("Redis client initialized successfully")

	return &RedisClient{
		client: rdb,
		config: cfg,
	}, nil
}

func (r *RedisClient) Close() error {
	if r.client != nil {
		log.Info().Msg("Closing Redis connection")
		return r.client.Close()
	}
	return nil
}

func (r *RedisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Basic operations
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Hash operations
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// List operations
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// Set operations
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// Distributed locking
func (r *RedisClient) AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result, err := r.client.SetNX(ctx, key, "locked", expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return result, nil
}

func (r *RedisClient) ReleaseLock(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) ExtendLock(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// Rate limiting using sliding window
func (r *RedisClient) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - window.Nanoseconds()

	pipe := r.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))

	// Count current entries
	countCmd := pipe.ZCard(ctx, key)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	currentCount := countCmd.Val()
	if currentCount >= int64(limit) {
		return false, nil // Rate limit exceeded
	}

	// Add current request
	pipe = r.client.Pipeline()
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, key, window)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to record rate limit entry: %w", err)
	}

	return true, nil
}

// Session management
func (r *RedisClient) SetSession(ctx context.Context, sessionID string, data map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.HMSet(ctx, key, data).Err()
}

func (r *RedisClient) GetSession(ctx context.Context, sessionID string) (map[string]string, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.HGetAll(ctx, key).Result()
}

func (r *RedisClient) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Del(ctx, key).Err()
}

// Cache management with JSON serialization
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	return r.client.Set(ctx, key, bytes, expiration).Err()
}

func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if dest == nil {
		return fmt.Errorf("destination must be non-nil for key %s", key)
	}

	bytes, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached value for key %s: %w", key, err)
	}

	return err
}
