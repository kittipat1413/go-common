[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Cache Package
This package provides a flexible and extensible caching interface designed for Go applications. It supports multiple types of caches, allowing you to choose the best caching strategy for your needs. Currently, the package includes an in-memory cache implementation called localcache, with plans to support additional cache types like Redis and Memcached in the future.

## Introduction
The Cache Package provides a unified caching interface (`Cache[T]`) with support for generic types, enabling type-safe caching of any data type. The package is designed with extensibility in mind, allowing for multiple cache implementations that conform to the same interface. This design enables you to switch between different cache backends (e.g., in-memory, Redis, Memcached) without changing your application logic.

## Installation
- **Local Cache**
    ```bash
    go get github.com/kittipat1413/go-common/framework/cache/localcache
    ```

## Documentation
[![Go Reference](https://pkg.go.dev/badge/github.com/kittipat1413/go-common/framework/cache.svg)](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/cache)

For detailed API documentation, examples, and usage patterns, visit the [Go Package Documentation](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/cache).

## Usage
### Cache Interface
The core of the cache package is the `Cache[T]` interface, which defines the methods that all cache implementations must provide:
```go
type Initializer[T any] func() (T, *time.Duration, error)

type Cache[T any] interface {
    Get(ctx context.Context, key string, initializer Initializer[T]) (T, error)
    Set(ctx context.Context, key string, value T, duration *time.Duration)
    Invalidate(ctx context.Context, key string) error
    InvalidateAll(ctx context.Context) error
}
```
- `Get`: Retrieves a value from the cache. If the key is missing or expired, it uses the provided `Initializer` function to load the value.
- `Set`: Manually sets a value in the cache with a specific expiration duration.
- `Invalidate`: Removes a specific key from the cache.
- `InvalidateAll`: Clears all items from the cache.

## Local Cache Implementation
The `localcache` package provides an in-memory cache implementation of the `Cache[T]` interface. It stores items in memory with optional expiration times and supports automatic cleanup of expired items.

### Creating a Cache Instance
```go
import (
    "github.com/kittipat1413/go-common/framework/cache/localcache"
)

// Create a new local cache with default settings
c := localcache.New[string]()
```

### Customizing Cache Options
You can customize the cache by providing options:
```go
import (
    "github.com/kittipat1413/go-common/framework/cache/localcache"
)

c := localcache.New[string](
    localcache.WithDefaultExpiration(10 * time.Minute),
    localcache.WithCleanupInterval(5 * time.Minute),
)
```
- `WithDefaultExpiration`: Sets the default expiration duration for cache items.
- `WithCleanupInterval`: Sets the interval for automatically cleaning up expired items.

### Using the Cache
```go
ctx := context.Background()
key := "greeting"

// Define an initializer function
initializer := func() (string, *time.Duration, error) {
    value := "Hello, World!"
    duration := 5 * time.Minute
    return value, &duration, nil
}

// Get value from cache (will initialize if not present)
value, err := c.Get(ctx, key, initializer)
if err != nil {
    // Handle error
}
fmt.Println("Value:", value)

// Set a value with a specific expiration
customDuration := 1 * time.Hour
c.Set(ctx, "customKey", "Custom Value", &customDuration)

// Invalidate a specific key
c.Invalidate(ctx, key)

// Invalidate all keys
c.InvalidateAll(ctx)
```

### Handling Items Expiration
When adding an item to the cache, you can control its expiration behavior using the Set method:
- `Custom Duration`: You can pass a specific duration for the item to expire.
- `No Expiration`: Use `NoExpireDuration` to make the item persist indefinitely.
- `Default Expiration`: If you pass `nil` for the duration, the cache will use the default expiration set during cache initialization.
```go
import "github.com/kittipat1413/go-common/framework/cache/localcache"

ctx := context.Background()
key := "session"

// Set item with default expiration (defined during cache initialization)
c.Set(ctx, key, "Session Data", nil)

// Set item with a custom expiration
customDuration := 1 * time.Hour
c.Set(ctx, "customKey", "Custom Value", &customDuration)

// Set item with no expiration (it persists indefinitely)
noExpiration := localcache.NoExpireDuration
c.Set(ctx, "persistentKey", "Persistent Value", &noExpiration)

```

## Example
You can find a complete working example in the repository under [framework/cache/example](example/).

## Extensibility
The cache package is designed to be extensible. You can implement additional cache types by creating new packages that conform to the `Cache[T]` interface.

### Implementing a New Cache Type
To implement a new cache type:
- Create a New Package: For example, `redis_cache`.
- Implement the `Cache[T]` Interface:
```go
package rediscache

import (
    "context"
    "time"

    "github.com/kittipat1413/go-common/framework/cache"
)

type redisCache[T any] struct {
    // Redis client and other fields
}

func New[T any](/* parameters */) cache.Cache[T] {
    return &redisCache[T]{ /* initialization */ }
}

func (r *redisCache[T]) Get(ctx context.Context, key string, initializer cache.Initializer[T]) (T, error) {
    // Implementation
}

func (r *redisCache[T]) Set(ctx context.Context, key string, value T, duration *time.Duration) {
    // Implementation
}

func (r *redisCache[T]) Invalidate(ctx context.Context, key string) error {
    // Implementation
}

func (r *redisCache[T]) InvalidateAll(ctx context.Context) error {
    // Implementation
}
```

## Error Handling
The cache package defines a common error for cache misses:

```go
var ErrCacheMiss = errors.New("cache miss")
```
When a key is not found or has expired, the Get method returns ErrCacheMiss. This allows you to handle cache misses distinctly from other errors.

**Example:**

```go
value, err := cache.Get(ctx, "missingKey", nil)
if err != nil {
    if errors.Is(err, cache.ErrCacheMiss) {
        // Handle cache miss (e.g., load the value from another source)
    } else {
        // Handle other errors
    }
}
```