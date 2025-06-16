package lockmanager

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrLockAlreadyTaken indicates that the lock is already held by another process.
	ErrLockAlreadyTaken = errors.New("lock already taken")
	// ErrUnlockNotPermitted indicates that the unlock operation is not permitted, likely due to a token mismatch.
	ErrUnlockNotPermitted = errors.New("unlock not permitted")
)

//go:generate mockgen -source=./locker.go -destination=./mocks/locker.go -package=locker_mocks
type LockManager interface {
	// Acquire attempts to acquire a lock for the given key with the specified TTL.
	//
	// If a token is provided, it will be used for the lock identity. If no token is provided,
	// a new unique token will be generated and returned.
	//
	// Returns ErrLockAlreadyTaken if the lock is currently held by another process.
	Acquire(ctx context.Context, key string, ttl time.Duration, token ...string) (string, error)

	// Release attempts to release the lock identified by key and token.
	//
	// Returns ErrUnlockNotPermitted if the token does not match the currently held lock.
	Release(ctx context.Context, key string, token string) error
}
