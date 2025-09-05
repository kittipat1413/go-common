package localcache

import (
	"context"
	"sync"
	"time"

	cache "github.com/kittipat1413/go-common/framework/cache"
	"github.com/kittipat1413/go-common/util/pointer"
	"golang.org/x/sync/singleflight"
)

const (
	// NoExpireDuration indicates that cache items should not expire.
	// Use this value to store items permanently in the cache.
	NoExpireDuration time.Duration = -1

	// defaultExpireDuration is the default expiration duration for cache items.
	// Set to NoExpireDuration, meaning items don't expire by default.
	defaultExpireDuration time.Duration = NoExpireDuration

	// defaultCleanupInterval is the default interval for cleaning up expired items.
	// The cleanup process runs every 5 minutes by default.
	defaultCleanupInterval time.Duration = 5 * time.Minute
)

// item represents a cached item with its data and optional expiration time.
type item[T any] struct {
	data    T          // The cached data
	expires *time.Time // Expiration time (nil means no expiration)
}

// config holds the configuration options for the local cache.
type config struct {
	defaultExpireDuration time.Duration // Default expiration duration for items
	cleanupInterval       time.Duration // Interval for automatic cleanup of expired items
	stopCleanupChannel    chan struct{} // Channel to signal cleanup goroutine to stop
}

// Option is a function type for configuring the local cache.
type Option func(*config)

// WithDefaultExpiration sets the default expiration duration for cache items.
// If not specified, items will not expire by default.
func WithDefaultExpiration(expiration time.Duration) Option {
	return func(c *config) {
		c.defaultExpireDuration = expiration
	}
}

// WithCleanupInterval sets the interval for automatically cleaning up expired items.
// A background goroutine will run periodically to remove expired items from memory.
func WithCleanupInterval(interval time.Duration) Option {
	return func(c *config) {
		c.cleanupInterval = interval
	}
}

// newConfig creates a new configuration with default values and applies the provided options.
func newConfig(opts ...Option) *config {
	c := &config{
		defaultExpireDuration: defaultExpireDuration,
		cleanupInterval:       defaultCleanupInterval,
		stopCleanupChannel:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// localcache is an in-memory cache implementation that provides thread-safe operations
// with automatic expiration and cleanup of items.
type localcache[T any] struct {
	mutex sync.RWMutex
	group singleflight.Group
	items map[string]item[T]
	config
}

// New creates a new localcache instance with optional configurations.
// It applies defaults if no options are provided and starts the cleanup
// goroutine if a cleanup interval is configured.
//
// The cache is thread-safe and can be used concurrently from multiple goroutines.
// A background cleanup process will automatically remove expired items based on
// the configured cleanup interval.
//
// Parameters:
//   - opts: Optional configuration functions
//
// Returns:
//   - A new cache.Cache[T] instance
//
// Example:
//
//	// Create cache with default settings
//	cache := New[string]()
//
//	// Create cache with custom expiration and cleanup interval
//	cache := New[User](
//		WithDefaultExpiration(30 * time.Minute),
//		WithCleanupInterval(5 * time.Minute),
//	)
func New[T any](opts ...Option) cache.Cache[T] {
	cfg := newConfig(opts...)
	c := &localcache[T]{
		items:  make(map[string]item[T]),
		config: pointer.GetValue(cfg),
	}

	// Start the cleanup process if a valid interval is provided
	if c.cleanupInterval > 0 {
		go c.startCleanup()
	}

	return c
}

// Get retrieves a value from the cache by key. If the key is missing and an initializer
// is provided, it uses the initializer to obtain the value using singleflight to prevent
// cache stampedes.
//
// The method first checks if the key exists and is not expired. If found, returns the cached value.
// If not found and an initializer is provided, it calls the initializer exactly once even if
// multiple goroutines request the same key simultaneously.
func (c *localcache[T]) Get(ctx context.Context, key string, initializer cache.Initializer[T]) (T, error) {
	if data, ok := c.get(key); ok {
		return data, nil
	} else {
		if initializer == nil {
			var zero T
			return zero, cache.ErrCacheMiss
		}
		return c.initialize(ctx, key, initializer)
	}
}

// Set adds an item to the cache with the specified key and duration.
// The method handles various expiration scenarios based on the duration parameter.
func (c *localcache[T]) Set(ctx context.Context, key string, value T, duration *time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var expiration *time.Time
	if duration != nil && pointer.GetValue(duration) != NoExpireDuration { // set expiration with input duration if it's not NoExpireDuration
		expTime := time.Now().Add(pointer.GetValue(duration))
		expiration = pointer.ToPointer(expTime)
	} else if duration == nil && c.defaultExpireDuration != NoExpireDuration { // set expiration with defaultExpireDuration if it's not NoExpireDuration
		expTime := time.Now().Add(c.defaultExpireDuration)
		expiration = pointer.ToPointer(expTime)
	}

	c.items[key] = item[T]{
		data:    value,
		expires: expiration,
	}
}

// Invalidate removes a specific key from the cache.
// If the key does not exist, this operation succeeds without error.
func (c *localcache[T]) Invalidate(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	return nil
}

// InvalidateAll removes all entries from the cache.
// This operation clears the entire cache, regardless of expiration times.
// Memory is freed by creating a new map.
func (c *localcache[T]) InvalidateAll(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]item[T])
	return nil
}

// get is an internal method that retrieves a value from the cache without initialization.
// It checks expiration and returns false for expired items.
func (c *localcache[T]) get(key string) (result T, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if itm, found := c.items[key]; found {
		// If expiration is nil, the item never expires
		if itm.expires == nil || time.Now().Before(pointer.GetValue(itm.expires)) {
			result, ok = itm.data, true
			return
		}
	}
	return
}

// initialize is an internal method that handles cache initialization using singleflight.
// It ensures that only one goroutine calls the initializer for each key, preventing
// cache stampede scenarios.
func (c *localcache[T]) initialize(ctx context.Context, key string, initializer cache.Initializer[T]) (T, error) {
	v, err, _ := c.group.Do(key, func() (interface{}, error) {
		// Double-check if the item was initialized by another goroutine
		if data, ok := c.get(key); ok {
			return data, nil
		}

		result, duration, err := initializer()
		if err != nil {
			var zero T
			return zero, err
		}

		// Set the item in the cache
		c.Set(ctx, key, result, duration)

		return result, nil
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return v.(T), nil
}

// startCleanup runs a background goroutine to periodically remove expired items.
// This method is automatically called when creating a new cache instance if
// a cleanup interval is configured.
//
// The cleanup process will continue until StopCleanup() is called or the
// stopCleanupChannel is closed.
func (c *localcache[T]) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCleanupChannel:
			return
		}
	}
}

// deleteExpired removes all expired items from the cache.
// This method is called periodically by the cleanup goroutine and can also
// be called manually if needed.
func (c *localcache[T]) deleteExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, itm := range c.items {
		if itm.expires != nil && itm.expires.Before(now) {
			delete(c.items, key)
		}
	}
}

// StopCleanup stops the background cleanup process gracefully.
// This method should be called when the cache is no longer needed to
// prevent goroutine leaks.
func (c *localcache[T]) StopCleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	select {
	case <-c.stopCleanupChannel:
		// Channel already closed, do nothing
	default:
		close(c.stopCleanupChannel)
	}
}
