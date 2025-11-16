package worker

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

type DistributedLock struct {
	redis     *redis.Client
	key       string
	value     string
	ttl       time.Duration
	renewable bool
	renewCh   chan struct{}
	stopCh    chan struct{}
}

type LockManager struct {
	redis     *redis.Client
	keyPrefix string
}

func NewLockManager(redis *redis.Client, keyPrefix string) *LockManager {
	if keyPrefix == "" {
		keyPrefix = "lock:"
	}
	return &LockManager{
		redis:     redis,
		keyPrefix: keyPrefix,
	}
}

func (lm *LockManager) AcquireLock(ctx context.Context, resource string, ttl time.Duration) (*DistributedLock, error) {
	key := lm.keyPrefix + resource
	value := uuid.New().String()

	acquired, err := lm.redis.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to acquire lock")
	}

	if !acquired {
		return nil, apperrors.Conflict("lock already held by another process")
	}

	lock := &DistributedLock{
		redis:     lm.redis,
		key:       key,
		value:     value,
		ttl:       ttl,
		renewable: false,
		renewCh:   make(chan struct{}),
		stopCh:    make(chan struct{}),
	}

	return lock, nil
}

func (lm *LockManager) AcquireRenewableLock(ctx context.Context, resource string, ttl time.Duration) (*DistributedLock, error) {
	lock, err := lm.AcquireLock(ctx, resource, ttl)
	if err != nil {
		return nil, err
	}

	lock.renewable = true

	// Start renewal goroutine
	go lock.renewalWorker(ctx)

	return lock, nil
}

func (lm *LockManager) TryAcquireLock(ctx context.Context, resource string, ttl time.Duration, maxWait time.Duration) (*DistributedLock, error) {
	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		lock, err := lm.AcquireLock(ctx, resource, ttl)
		if err == nil {
			return lock, nil
		}

		// Wait a bit before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}

	return nil, apperrors.Wrapf(nil, apperrors.ErrorTypeInternal, "timeout waiting for lock on resource %s", resource)
}

func (dl *DistributedLock) Release(ctx context.Context) error {
	// Stop renewal if it's running
	if dl.renewable {
		close(dl.stopCh)
	}

	// Use Lua script to ensure we only delete the lock if we own it
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.redis.Eval(ctx, script, []string{dl.key}, dl.value).Result()
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to release lock")
	}

	deleted := result.(int64)
	if deleted == 0 {
		return apperrors.Authentication("lock was not owned by this process")
	}

	return nil
}

func (dl *DistributedLock) Extend(ctx context.Context, newTTL time.Duration) error {
	// Use Lua script to extend TTL only if we own the lock
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := dl.redis.Eval(ctx, script, []string{dl.key}, dl.value, int(newTTL.Seconds())).Result()
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to extend lock")
	}

	extended := result.(int64)
	if extended == 0 {
		return apperrors.Authentication("lock was not owned by this process")
	}

	dl.ttl = newTTL
	return nil
}

func (dl *DistributedLock) IsHeld(ctx context.Context) (bool, error) {
	value, err := dl.redis.Get(ctx, dl.key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check lock")
	}

	return value == dl.value, nil
}

func (dl *DistributedLock) GetTTL(ctx context.Context) (time.Duration, error) {
	ttl, err := dl.redis.TTL(ctx, dl.key).Result()
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get lock TTL")
	}

	if ttl == -1 {
		return 0, apperrors.Validation("lock exists but has no TTL")
	}

	if ttl == -2 {
		return 0, apperrors.Validation("lock does not exist")
	}

	return ttl, nil
}

func (dl *DistributedLock) renewalWorker(ctx context.Context) {
	// Renew the lock at 1/3 of its TTL
	renewInterval := dl.ttl / 3
	ticker := time.NewTicker(renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dl.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := dl.Extend(ctx, dl.ttl)
			if err != nil {
				// Lock renewal failed, stop trying
				return
			}
		}
	}
}

// LockGuard provides a convenient way to use locks with defer
type LockGuard struct {
	lock *DistributedLock
	ctx  context.Context
}

func (lm *LockManager) WithLock(ctx context.Context, resource string, ttl time.Duration, fn func() error) error {
	lock, err := lm.AcquireLock(ctx, resource, ttl)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	return fn()
}

func (lm *LockManager) WithLockTimeout(ctx context.Context, resource string, ttl time.Duration, maxWait time.Duration, fn func() error) error {
	lock, err := lm.TryAcquireLock(ctx, resource, ttl, maxWait)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	return fn()
}

func (lm *LockManager) WithRenewableLock(ctx context.Context, resource string, ttl time.Duration, fn func() error) error {
	lock, err := lm.AcquireRenewableLock(ctx, resource, ttl)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	return fn()
}
