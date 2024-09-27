# Cache Package
This package provides a flexible and extensible caching interface designed for Go applications. It supports multiple types of caches, allowing you to choose the best caching strategy for your needs. Currently, the package includes an in-memory cache implementation called localcache, with plans to support additional cache types like Redis and Memcached in the future.

## Introduction
The Cache Package provides a unified caching interface (`Cache[T]`) with support for generic types, enabling type-safe caching of any data type. The package is designed with extensibility in mind, allowing for multiple cache implementations that conform to the same interface. This design enables you to switch between different cache backends (e.g., in-memory, Redis, Memcached) without changing your application logic.

## Usage
### Cache Interface
The core of the cache package is the `Cache[T]` interface, which defines the methods that all cache implementations must provide:
```golang
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

## Example
You can find a complete working example in the repository under [framework/cache/example](example/).

## Extensibility
The cache package is designed to be extensible. You can implement additional cache types by creating new packages that conform to the `Cache[T]` interface.

### Implementing a New Cache Type
To implement a new cache type:
- Create a New Package: For example, `redis_cache`.
- Implement the `Cache[T]` Interface:
```golang
package redis_cache

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