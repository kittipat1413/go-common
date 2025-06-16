package redsync_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/go-redsync/redsync/v4"
	"github.com/kittipat1413/go-common/framework/lockmanager"
	redsyncLocker "github.com/kittipat1413/go-common/framework/lockmanager/redsync"
	"github.com/stretchr/testify/assert"
)

func TestRedsyncLockManager_AcquireAndRelease(t *testing.T) {
	ctx := context.Background()

	// Create mock Redis pool
	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(client)

	key := "test-lock"
	ttl := 2 * time.Second

	// Acquire lock
	mock.Regexp().ExpectSetNX(key, `[a-z]+`, ttl).SetVal(true)
	token, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Release lock
	mock.Regexp().ExpectEvalSha(`[a-z]+`, []string{key}, token).SetVal(int64(1))
	err = manager.Release(ctx, key, token)
	assert.NoError(t, err)
}

func TestRedsyncLockManager_WithTokenGenerator(t *testing.T) {
	ctx := context.Background()

	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(
		client,
		redsyncLocker.WithTokenGenerator(func(key string) string {
			return "custom-token-for:" + key
		}),
	)

	key := "custom-token-lock"
	ttl := 2 * time.Second

	// Acquire lock with custom token generator
	mock.ExpectSetNX(key, "custom-token-for:"+key, ttl).SetVal(true)
	token, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)
	assert.Equal(t, "custom-token-for:"+key, token)

	// Release lock
	mock.Regexp().ExpectEvalSha(`[a-z]+`, []string{key}, token).SetVal(int64(1))
	err = manager.Release(ctx, key, token)
	assert.NoError(t, err)
}

func TestRedsyncLockManager_WithTokenOption(t *testing.T) {
	ctx := context.Background()

	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(client)

	key := "custom-token-lock"
	tokenValue := "my-custom-token"
	ttl := 2 * time.Second

	// Acquire lock with custom token
	mock.ExpectSetNX(key, tokenValue, ttl).SetVal(true)
	token, err := manager.Acquire(ctx, key, ttl, tokenValue)
	assert.NoError(t, err)
	assert.Equal(t, tokenValue, token)

	// Release lock
	mock.Regexp().ExpectEvalSha(`[a-z]+`, []string{key}, tokenValue).SetVal(int64(1))
	err = manager.Release(ctx, key, token)
	assert.NoError(t, err)
}

func TestRedsyncLockManager_AcquireFails_WhenLockAlreadyExists(t *testing.T) {
	ctx := context.Background()

	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(
		client,
		redsyncLocker.WithRedsyncOptions(
			redsync.WithTries(1),
			redsync.WithFailFast(true),
		),
	)

	key := "lock-fail"
	ttl := 2 * time.Second

	// Acquire lock
	mock.Regexp().ExpectSetNX(key, `[a-z]+`, ttl).SetErr(&redsync.ErrTaken{})
	token, err := manager.Acquire(ctx, key, ttl)
	assert.ErrorIs(t, err, lockmanager.ErrLockAlreadyTaken)
	assert.Empty(t, token)
}

func TestRedsyncLockManager_AcquireFails_WhenRedisError(t *testing.T) {
	ctx := context.Background()

	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(
		client,
		redsyncLocker.WithRedsyncOptions(
			redsync.WithTries(1),
			redsync.WithFailFast(true),
		),
	)

	key := "lock-error"
	ttl := 2 * time.Second

	// Simulate Redis error
	mock.Regexp().ExpectSetNX(key, `[a-z]+`, ttl).SetErr(assert.AnError)
	token, err := manager.Acquire(ctx, key, ttl)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestRedsyncLockManager_ReleaseSucceeds_WhenLockAlreadyExpired(t *testing.T) {
	ctx := context.Background()

	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(
		client,
		redsyncLocker.WithRedsyncOptions(
			redsync.WithTries(1),
			redsync.WithFailFast(true),
		),
	)
	key := "release-fail"
	token := "some-token"

	// Release lock
	mock.Regexp().ExpectEvalSha(`[a-z]+`, []string{key}, token).SetVal(int64(-1))
	err := manager.Release(ctx, key, token)
	assert.NoError(t, err)
}

func TestRedsyncLockManager_ReleaseFails_WhenTokenDoesNotMatch(t *testing.T) {
	ctx := context.Background()

	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(
		client,
		redsyncLocker.WithRedsyncOptions(
			redsync.WithTries(1),
			redsync.WithFailFast(true),
		),
	)

	key := "release-mismatch"
	ttl := 2 * time.Second

	// Acquire lock
	mock.Regexp().ExpectSetNX(key, `[a-z]+`, ttl).SetVal(true)
	token, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	mismatchedToken := "mismatched-token"
	// Release lock
	mock.Regexp().ExpectEvalSha(`[a-z]+`, []string{key}, mismatchedToken).SetErr(&redsync.ErrTaken{})
	err = manager.Release(ctx, key, mismatchedToken)
	assert.ErrorIs(t, err, lockmanager.ErrUnlockNotPermitted)
}

func TestRedsyncLockManager_ReleaseFails_WhenRedisError(t *testing.T) {
	ctx := context.Background()

	client, mock := redismock.NewClientMock()
	manager := redsyncLocker.NewRedsyncLockManager(
		client,
		redsyncLocker.WithRedsyncOptions(
			redsync.WithTries(1),
			redsync.WithFailFast(true),
		),
	)

	key := "release-error"
	token := "some-token"

	// Simulate Redis error
	mock.Regexp().ExpectEvalSha(`[a-z]+`, []string{key}, token).SetErr(assert.AnError)
	err := manager.Release(ctx, key, token)
	assert.Error(t, err)
}
