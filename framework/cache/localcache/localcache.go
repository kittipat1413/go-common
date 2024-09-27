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
	NoExpireDuration       time.Duration = -1
	defaultExpireDuration  time.Duration = NoExpireDuration
	defaultCleanupInterval time.Duration = 5 * time.Minute
)

type item[T any] struct {
	data    T
	expires *time.Time
}

type config struct {
	defaultExpireDuration time.Duration
	cleanupInterval       time.Duration
	stopCleanupChannel    chan struct{}
}

type Option func(*config)

// WithDefaultExpiration sets the default expiration duration for cache items.
func WithDefaultExpiration(expiration time.Duration) Option {
	return func(c *config) {
		c.defaultExpireDuration = expiration
	}
}

// WithCleanupInterval sets the interval for automatically cleaning up expired items.
func WithCleanupInterval(interval time.Duration) Option {
	return func(c *config) {
		c.cleanupInterval = interval
	}
}

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

type localcache[T any] struct {
	mutex sync.RWMutex
	group singleflight.Group
	items map[string]item[T]
	config
}

// New creates a new localcache instance with optional configurations.
// It applies defaults if no options are provided.
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

// Get retrieves a value from the cache. If the key is missing and an initializer
// is provided, it uses the initializer to obtain the value.
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
// If duration is nil, the default expiration is used.
// If duration is NoExpireDuration, the item does not expire.
func (c *localcache[T]) Set(ctx context.Context, key string, value T, duration *time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var expiration *time.Time
	if duration != nil && pointer.GetValue(duration) != NoExpireDuration { // set expiration with input duration if it's not NoExpireDuration
		expTime := time.Now().Add(pointer.GetValue(duration))
		expiration = pointer.ToPointer(expTime)
	} else if c.defaultExpireDuration != NoExpireDuration { // set expiration with defaultExpireDuration if it's not NoExpireDuration
		expTime := time.Now().Add(c.defaultExpireDuration)
		expiration = pointer.ToPointer(expTime)
	}

	c.items[key] = item[T]{
		data:    value,
		expires: expiration,
	}
}

func (c *localcache[T]) Invalidate(ctx context.Context, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	return nil
}

func (c *localcache[T]) InvalidateAll(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]item[T])
	return nil
}

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

// StopCleanup stops the background cleanup process.
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
