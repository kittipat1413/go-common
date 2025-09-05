[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Lock Manager Package
The Lock Manager package provides a unified and extensible distributed locking interface for Go applications. It offers a consistent way to implement distributed locks across different services, making it easy to coordinate access to shared resources in distributed systems.

## Features
- **Unified Interface:** Consistent API for different lock implementations
- **[Redis Implementation](./redsync/):** Built-in support for Redis-based distributed locks
  - **Customizable Lock Options:** Configure lock behavior with options like TTL and retry parameters
  - **Context Support:** All operations respect context cancellation and deadlines
- **[LocalLock Implementation](./locallock/):** Simple in-memory lock for local use cases
  - **Custom Token Generation:** Optionally provide custom token generation logic
  - **Auto Cleanup:** Locks are automatically cleaned up when ttl expires

## Installation
```bash
go get github.com/kittipat1413/go-common
```

## Usage
Locking Interface
The core of the package is the `LockManager` interface:
```go
type LockManager interface {
	// Acquire attempts to acquire a lock with given key and TTL.
	// Optionally accepts a token; otherwise, one is auto-generated.
	Acquire(ctx context.Context, key string, ttl time.Duration, token ...string) (string, error)

	// Release attempts to release a lock with the given key and token.
	Release(ctx context.Context, key string, token string) error
}
```

### 🚀 Getting Started with Redis
Create Redis LockManager
```go
import (
	"github.com/redis/go-redis/v9"
	redsyncLocker "github.com/kittipat1413/go-common/framework/lockmanager/redsync"
)

func main() {
	// Setup Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Initialize LockManager
	manager := redsyncLocker.NewRedsyncLockManager(client)
}
```
🔐 Acquire Lock
```go
ctx := context.Background()
token, err := lockManager.Acquire(ctx, "email-job-lock", 5*time.Second)
if err != nil {
	if errors.Is(err, locker.ErrLockAlreadyTaken) {
		// Handle lock contention
	}
	log.Fatalf("failed to acquire lock: %v", err)
}
```
🔓 Release Lock
```go
err = lockManager.Release(ctx, "email-job-lock", token)
if err != nil {
	if errors.Is(err, locker.ErrUnlockNotPermitted) {
		// You don't own this lock
	}
	log.Fatalf("failed to release lock: %v", err)
}
```
⚙️ Advanced Configuration
```go
manager := redsyncLocker.NewRedsyncLockManager(
		client,
		redsyncLocker.WithTokenGenerator(func(key string) string {
			return "lock-token-for:" + key
		}),
		redsyncLocker.WithRedsyncOptions(
			redsync.WithTries(5),                        // Retry up to 5 times
			redsync.WithRetryDelay(50*time.Millisecond), // Wait 50ms between retries
		),
	)
```

## Example
You can find a complete working example in the repository under [framework/lockmanager/example](example/).