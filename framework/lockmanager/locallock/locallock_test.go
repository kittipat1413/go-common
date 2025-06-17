package local_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/lockmanager"
	local "github.com/kittipat1413/go-common/framework/lockmanager/locallock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalLockManager_AcquireAndRelease(t *testing.T) {
	manager := local.NewLocalLockManager()
	ctx := context.Background()

	key := "test-lock"
	ttl := 100 * time.Millisecond

	token, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	err = manager.Release(ctx, key, token)
	assert.NoError(t, err)
}

func TestLocalLockManager_WithTokenGenerator(t *testing.T) {
	customToken := "custom-token"
	manager := local.NewLocalLockManager(
		local.WithTokenGenerator(func(key string) string {
			return customToken
		}),
	)
	ctx := context.Background()

	key := "custom-token-lock"
	ttl := 100 * time.Millisecond

	token, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)
	assert.Equal(t, customToken, token)

	err = manager.Release(ctx, key, token)
	assert.NoError(t, err)
}

func TestLocalLockManager_WithProvidedToken(t *testing.T) {
	manager := local.NewLocalLockManager()
	ctx := context.Background()

	key := "manual-token-lock"
	tokenValue := "my-custom-token"
	ttl := 100 * time.Millisecond

	token, err := manager.Acquire(ctx, key, ttl, tokenValue)
	assert.NoError(t, err)
	assert.Equal(t, tokenValue, token)

	err = manager.Release(ctx, key, token)
	assert.NoError(t, err)
}

func TestLocalLockManager_AcquireWithSameToken_ExtendsTTL(t *testing.T) {
	manager := local.NewLocalLockManager()
	ctx := context.Background()

	key := "reacquire-lock"
	token := "shared-token"
	ttl := 100 * time.Millisecond

	// First acquire
	tok, err := manager.Acquire(ctx, key, ttl, token)
	assert.NoError(t, err)
	assert.Equal(t, token, tok)

	time.Sleep(50 * time.Millisecond)

	// Re-acquire with same token to extend TTL
	tok, err = manager.Acquire(ctx, key, ttl, token)
	assert.NoError(t, err)
	assert.Equal(t, token, tok)

	// Wait less than new TTL to ensure it's not expired
	time.Sleep(60 * time.Millisecond)

	// Should still be able to re-acquire
	tok, err = manager.Acquire(ctx, key, ttl, token)
	assert.NoError(t, err)
	assert.Equal(t, token, tok)
}

func TestLocalLockManager_AcquireWithDifferentToken_Fails(t *testing.T) {
	manager := local.NewLocalLockManager()
	ctx := context.Background()

	key := "different-token-lock"
	ttl := 100 * time.Millisecond

	// First acquire with one token
	token1 := "first-token"
	token, err := manager.Acquire(ctx, key, ttl, token1)
	assert.NoError(t, err)
	assert.Equal(t, token1, token)

	// Try to acquire with a different token
	token2 := "different-token"
	_, err = manager.Acquire(ctx, key, ttl, token2)
	assert.ErrorIs(t, err, lockmanager.ErrLockAlreadyTaken)
}

func TestLocalLockManager_AcquireSucceeds_AfterTTLExpires(t *testing.T) {
	manager := local.NewLocalLockManager(local.WithCleanupInterval(50 * time.Millisecond))
	ctx := context.Background()

	key := "ttl-lock"
	ttl := 60 * time.Millisecond

	token, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)

	time.Sleep(70 * time.Millisecond)

	// Try again after TTL should have expired
	newToken, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)
	assert.NotEqual(t, token, newToken)
}

func TestLocalLockManager_ReleaseFails_WhenTokenDoesNotMatch(t *testing.T) {
	manager := local.NewLocalLockManager()
	ctx := context.Background()

	key := "wrong-token-lock"
	ttl := 100 * time.Millisecond

	_, err := manager.Acquire(ctx, key, ttl)
	assert.NoError(t, err)

	// Try to release with incorrect token
	err = manager.Release(ctx, key, "invalid-token")
	assert.ErrorIs(t, err, lockmanager.ErrUnlockNotPermitted)
}

func TestLocalLockManager_ReleaseNoop_WhenLockNotFound(t *testing.T) {
	manager := local.NewLocalLockManager()
	ctx := context.Background()

	// Should not error even though lock doesn't exist
	err := manager.Release(ctx, "non-existent-key", "any-token")
	assert.NoError(t, err)
}

func TestLocalLockManager_StopTerminatesCleanup(t *testing.T) {
	manager := local.NewLocalLockManager(
		local.WithCleanupInterval(10 * time.Millisecond),
	)
	ctx := context.Background()

	// Use reflection to ensure Stop method exists (sanity check)
	llm, ok := manager.(interface{ Stop() })
	require.True(t, ok, "Expected local lock manager to have a Stop method")

	// Stop cleanup loop
	llm.Stop()

	// Acquire a lock with short TTL (should have expired if cleanup were running)
	key := "lock-key"
	ttl := 15 * time.Millisecond
	_, err := manager.Acquire(ctx, key, ttl)
	require.NoError(t, err)

	// Wait beyond TTL and cleanup interval
	time.Sleep(30 * time.Millisecond)

	// Try acquiring again — should succeed since lock should have expired even without cleanup
	_, err = manager.Acquire(ctx, key, ttl)
	require.NoError(t, err)
}

func TestLocalLockManager_CleanupExpiredLocks(t *testing.T) {
	ctx := context.Background()

	// Use short cleanup interval for test speed
	manager := local.NewLocalLockManager(local.WithCleanupInterval(10 * time.Millisecond))

	key := "expiring-lock"
	ttl := 10 * time.Millisecond

	// Acquire a short-lived lock
	_, err := manager.Acquire(ctx, key, ttl)
	require.NoError(t, err)

	// Wait for lock to expire and be cleaned
	time.Sleep(50 * time.Millisecond)

	// Use reflection to ensure Stop method exists (sanity check)
	llm, ok := manager.(interface{ Stop() })
	require.True(t, ok, "Expected local lock manager to have a Stop method")

	// Stop cleanup loop
	llm.Stop()

	// Use reflection to check internal locks map is empty
	val := reflect.ValueOf(manager).Elem().FieldByName("locks")
	if !val.IsValid() || val.Kind() != reflect.Map {
		t.Fatalf("Expected localLockManager to have a 'locks' field of type map")
	}

	require.Equal(t, 0, val.Len(), "Expected expired lock to be cleaned up")
}
