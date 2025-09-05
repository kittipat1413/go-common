package cache

import (
	"context"
	"errors"
	"time"
)

//go:generate mockgen -source=./cache.go -destination=./mocks/cache.go -package=cache_mocks

// ErrCacheMiss is returned when a requested cache key is not found.
var ErrCacheMiss = errors.New("cache miss")

// Initializer is a function type that provides lazy initialization for cache values.
// It returns the value to be cached, an optional expiration duration, and any error that occurred.
// If the duration pointer is nil, the cache implementation's default expiration will be used.
//
// Example:
//
//	initializer := func() (string, *time.Duration, error) {
//		value := fetchFromDatabase()
//		ttl := 5 * time.Minute
//		return value, &ttl, nil
//	}
type Initializer[T any] func() (T, *time.Duration, error)

// Cache provides a generic interface for caching operations with type safety.
// It supports lazy initialization, custom expiration times, and context-aware operations.
type Cache[T any] interface {
	// Get retrieves a value from the cache by key. If the key is not found,
	// the initializer function is called to provide the value, which is then
	// stored in the cache before being returned.
	//
	// The initializer function is called only when the key is not found in the cache
	// (cache miss). If the initializer returns an error, the value is not cached
	// and the error is returned to the caller.
	Get(ctx context.Context, key string, initializer Initializer[T]) (T, error)

	// Set stores a value in the cache with the specified key and expiration duration.
	// If duration is nil, the cache implementation's default expiration is used.
	// If duration is zero or negative, the value may be stored without expiration
	// (implementation-dependent).
	Set(ctx context.Context, key string, value T, duration *time.Duration)

	// Invalidate removes a specific key from the cache.
	// If the key does not exist, this operation should not return an error.
	Invalidate(ctx context.Context, key string) error

	// InvalidateAll removes all entries from the cache.
	// This operation clears the entire cache, regardless of expiration times.
	InvalidateAll(ctx context.Context) error
}
