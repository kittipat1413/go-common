package redsync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kittipat1413/go-common/framework/lockmanager"
	"github.com/rs/xid"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

// redsyncLockManager implements lockmanager.LockManager using go-redsync for distributed locking.
// Provides Redis-based distributed mutex functionality with configurable retry behavior,
// custom token generation, and automatic lock expiration.
type redsyncLockManager struct {
	rsync         *redsync.Redsync
	genTokenFunc  func(key string) string
	mutexDefaults []redsync.Option
}

// Option represents a functional option for configuring redsyncLockManager.
type Option func(*redsyncLockManager)

// WithTokenGenerator sets a custom function for generating lock tokens based on the lock key.
// Tokens are used to identify lock ownership and ensure only the lock holder can release it.
// If not provided, a default XID-based generator will be used.
//
// Parameters:
//   - f: Function that takes a lock key and returns a unique token
//
// Returns:
//   - Option: Configuration option for the lock manager
//
// Example:
//
//	WithTokenGenerator(func(key string) string {
//	    return fmt.Sprintf("service-a-%s-%d", key, time.Now().Unix())
//	})
func WithTokenGenerator(f func(key string) string) Option {
	return func(r *redsyncLockManager) {
		r.genTokenFunc = f
	}
}

// WithRedsyncOptions sets default options for all mutexes created by this LockManager.
// These options apply to both Acquire and Release calls, enabling consistent
// behavior across all locks managed by this instance.
//
// Parameters:
//   - opts: Redsync options to apply to all mutexes
//
// Returns:
//   - Option: Configuration option for the lock manager
//
// Example:
//
//	WithRedsyncOptions(
//	    redsync.WithTries(5),
//	    redsync.WithRetryDelay(200*time.Millisecond),
//	)
func WithRedsyncOptions(opts ...redsync.Option) Option {
	return func(r *redsyncLockManager) {
		r.mutexDefaults = append(r.mutexDefaults, opts...)
	}
}

// NewRedsyncLockManager constructs a new distributed LockManager using go-redis as backend.
// The lock manager provides distributed mutex functionality across multiple Redis instances
// with configurable retry behavior and token generation.
//
// Parameters:
//   - redisClient: Redis client (single instance, cluster, or sentinel)
//   - opts: Optional configuration for token generation and mutex behavior
//
// Returns:
//   - lockmanager.LockManager: Configured distributed lock manager
//
// Example usage:
//
//	locker := redsync.NewRedsyncLockManager(
//	    redisClient,
//	    redsync.WithTokenGenerator(func(key string) string {
//	        return "custom-token-for:" + key
//	    }),
//	    redsync.WithRedsyncOptions(
//	        redsync.WithTries(10),
//	        redsync.WithRetryDelay(100*time.Millisecond),
//	    ),
//	)
//
//	token, err := locker.Acquire(ctx, "resource-key", 5*time.Second)
//	if err != nil {
//	    log.Fatalf("failed to acquire lock: %v", err)
//	}
//
//	defer locker.Release(ctx, "resource-key", token)
func NewRedsyncLockManager(redisClient redis.UniversalClient, opts ...Option) lockmanager.LockManager {
	pool := goredis.NewPool(redisClient)
	rs := redsync.New(pool)

	manager := &redsyncLockManager{
		rsync:         rs,
		genTokenFunc:  defaultTokenGenerator,
		mutexDefaults: []redsync.Option{},
	}

	for _, opt := range opts {
		opt(manager)
	}

	return manager
}

// Acquire attempts to obtain a distributed lock for the given key with a specified TTL.
// The lock will automatically expire after the TTL duration to prevent deadlocks.
// If a token is provided, it will be used as the lock identifier; otherwise, a token is auto-generated.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - key: Unique identifier for the resource to lock
//   - ttl: Time-to-live for the lock (automatic expiration)
//   - token: Optional custom token (auto-generated if not provided)
//
// Returns:
//   - string: Lock token for releasing the lock
//   - error: ErrLockAlreadyTaken if lock is held by another process, or other acquisition errors
func (r *redsyncLockManager) Acquire(ctx context.Context, key string, ttl time.Duration, token ...string) (string, error) {
	var value string
	if len(token) > 0 && token[0] != "" {
		value = token[0]
	} else {
		value = r.genTokenFunc(key)
	}

	opts := append([]redsync.Option{}, r.mutexDefaults...)
	opts = append(opts,
		redsync.WithExpiry(ttl),
		redsync.WithGenValueFunc(func() (string, error) {
			return value, nil
		}),
	)

	mutex := r.rsync.NewMutex(key, opts...)

	if err := mutex.LockContext(ctx); err != nil {
		var taken *redsync.ErrTaken
		if errors.As(err, &taken) {
			return "", lockmanager.ErrLockAlreadyTaken
		}
		return "", fmt.Errorf("redsync lock failed: %w", err)
	}

	return value, nil
}

// Release releases the lock for the given key using the provided token.
// Only the process that holds the lock (matching token) can successfully release it.
// If the lock has already expired, the operation is treated as successful.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - key: Unique identifier for the resource to unlock
//   - token: Lock token returned from Acquire operation
//
// Returns:
//   - error: ErrUnlockNotPermitted if token doesn't match lock ownership, nil on success
func (r *redsyncLockManager) Release(ctx context.Context, key string, token string) error {
	opts := append([]redsync.Option{}, r.mutexDefaults...)
	opts = append(opts,
		redsync.WithValue(token),
	)

	mutex := r.rsync.NewMutex(key, opts...)

	ok, err := mutex.UnlockContext(ctx)
	switch {
	case errors.Is(err, redsync.ErrLockAlreadyExpired):
		// lock is already expired – treat as success
		return nil

	case errors.As(err, new(*redsync.ErrTaken)):
		// someone else holds the lock – not permitted
		return lockmanager.ErrUnlockNotPermitted

	case !ok:
		// unexpected error during unlock
		return fmt.Errorf("redsync unlock failed: %w", err)
	}

	return nil
}

// defaultTokenGenerator creates a globally unique lock token using xid.
func defaultTokenGenerator(key string) string {
	return xid.New().String()
}
