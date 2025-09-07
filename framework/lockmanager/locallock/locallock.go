package local

import (
	"context"
	"sync"
	"time"

	"github.com/kittipat1413/go-common/framework/lockmanager"
	"github.com/rs/xid"
)

const (
	// defaultCleanupInterval defines how frequently expired locks are cleaned up.
	defaultCleanupInterval time.Duration = 5 * time.Minute
)

// localLock represents a lock with an associated token and expiration time.
type localLock struct {
	token     string    // Unique identifier for lock ownership
	expiresAt time.Time // When the lock automatically expires
}

// config contains configuration parameters for the localLockManager.
type config struct {
	cleanupInterval time.Duration           // How often to clean expired locks
	tokenGenerator  func(key string) string // Function to generate lock tokens
}

// Option defines a functional option for customizing the localLockManager.
type Option func(*config)

// WithCleanupInterval sets the interval for cleaning up expired locks.
//
// Parameters:
//   - interval: How often to run cleanup
//
// Returns:
//   - Option: Configuration option for the lock manager
//
// Example:
//
//	WithCleanupInterval(10*time.Minute) // Clean up every 10 minutes
func WithCleanupInterval(interval time.Duration) Option {
	return func(c *config) {
		c.cleanupInterval = interval
	}
}

// WithTokenGenerator sets a custom token generator for lock acquisition.
// Used when no explicit token is provided during Acquire calls.
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
//	    return fmt.Sprintf("process-%d-%s", os.Getpid(), key)
//	})
func WithTokenGenerator(f func(key string) string) Option {
	return func(c *config) {
		c.tokenGenerator = f
	}
}

// localLockManager provides an in-memory, single-node implementation of the LockManager interface.
// It supports TTL-based locking and periodic cleanup of expired locks.
type localLockManager struct {
	mu    sync.Mutex           // Protects the locks map
	locks map[string]localLock // In-memory storage for active locks
	cfg   config               // Configuration parameters
	done  chan struct{}        // Signal channel for cleanup goroutine
}

// NewLocalLockManager creates a new local in-memory lock manager with optional configuration.
// Starts a background goroutine for periodic cleanup of expired locks.
//
// Parameters:
//   - opts: Optional configuration functions
//
// Returns:
//   - lockmanager.LockManager: Configured local lock manager
//
// Example usage:
//
//	locker := local.NewLocalLockManager(
//	    local.WithCleanupInterval(10*time.Minute),
//	    local.WithTokenGenerator(func(key string) string {
//	        return "custom-token-for:" + key
//	    }),
//	)
func NewLocalLockManager(opts ...Option) lockmanager.LockManager {
	cfg := config{
		cleanupInterval: defaultCleanupInterval,
		tokenGenerator:  defaultTokenGenerator,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	m := &localLockManager{
		locks: make(map[string]localLock),
		cfg:   cfg,
		done:  make(chan struct{}),
	}

	go m.cleanupLoop()
	return m
}

// Acquire tries to acquire a lock for the specified key with the given TTL.
// If a token is provided, it is used as-is; otherwise, a token is generated using the configured generator.
// Returns the token associated with the lock or ErrLockAlreadyTaken if the key is already locked.
func (m *localLockManager) Acquire(ctx context.Context, key string, ttl time.Duration, token ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if lock, exists := m.locks[key]; exists && lock.expiresAt.After(now) {
		// Allow re-acquire with same token to extend TTL
		if len(token) == 0 || token[0] != lock.token {
			return "", lockmanager.ErrLockAlreadyTaken
		}
	}

	var lockToken string
	if len(token) > 0 && token[0] != "" {
		lockToken = token[0]
	} else {
		lockToken = m.cfg.tokenGenerator(key)
	}

	m.locks[key] = localLock{
		token:     lockToken,
		expiresAt: now.Add(ttl),
	}
	return lockToken, nil
}

// Release releases a previously acquired lock for the specified key.
// If the token does not match the existing lock, ErrUnlockNotPermitted is returned.
func (m *localLockManager) Release(ctx context.Context, key string, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locks[key]
	if !exists {
		// Lock already expired or not found — treat as successful release.
		return nil
	}
	if lock.token != token {
		return lockmanager.ErrUnlockNotPermitted
	}

	delete(m.locks, key)
	return nil
}

// Stop terminates the background cleanup goroutine and prevents memory leaks.
// Should be called during application shutdown to clean up resources.
func (m *localLockManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-m.done:
		// Already closed.
	default:
		close(m.done)
	}
}

// cleanupLoop runs a periodic task that deletes expired locks at the configured interval.
// Runs in a separate goroutine to avoid blocking lock operations.
func (m *localLockManager) cleanupLoop() {
	ticker := time.NewTicker(m.cfg.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpiredLocks()
		case <-m.done:
			return
		}
	}
}

// cleanupExpiredLocks removes all expired locks from the map to prevent memory leaks.
// Called periodically by the cleanup loop.
func (m *localLockManager) cleanupExpiredLocks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, lock := range m.locks {
		if lock.expiresAt.Before(now) {
			delete(m.locks, key)
		}
	}
}

// defaultTokenGenerator generates a unique lock token using xid.
// Used as the fallback generator if WithTokenGenerator is not set.
func defaultTokenGenerator(key string) string {
	return xid.New().String()
}
