package localcache_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kittipat1413/go-common/framework/cache"
	"github.com/kittipat1413/go-common/framework/cache/localcache"
)

func TestLocalCache_SetAndGet(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "testKey"
	value := "testValue"
	duration := time.Minute

	// Set a value
	c.Set(ctx, key, value, &duration)

	// Get the value
	gotValue, err := c.Get(ctx, key, nil)
	require.NoError(t, err)
	require.Equal(t, value, gotValue, "The value retrieved from cache should match the expected value")
}

func TestLocalCache_GetWithInitializer(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[int]()

	key := "number"
	expectedValue := 42
	duration := time.Minute

	var initializerCalled int
	initializer := func() (int, *time.Duration, error) {
		initializerCalled++
		return expectedValue, &duration, nil
	}

	// Get value with initializer
	value, err := c.Get(ctx, key, initializer)
	require.NoError(t, err)
	require.Equal(t, expectedValue, value, "The value retrieved should match the expected value")
	require.Equal(t, 1, initializerCalled, "Initializer should have been called once")

	// Get value again, initializer should not be called
	initializerCalled = 0
	value, err = c.Get(ctx, key, initializer)
	require.NoError(t, err)
	require.Equal(t, expectedValue, value, "The value retrieved should match the expected value")
	require.Equal(t, 0, initializerCalled, "Initializer should not have been called again")
}

func TestLocalCache_Get_CacheMiss(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "missingKey"

	// Try to get a key that doesn't exist without an initializer
	_, err := c.Get(ctx, key, nil)
	require.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss when getting a missing key without initializer")
}

func TestLocalCache_Get_InitializerError(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "initErrorKey"
	expectedErr := errors.New("initializer error")

	initializer := func() (string, *time.Duration, error) {
		return "", nil, expectedErr
	}

	_, err := c.Get(ctx, key, initializer)
	require.ErrorIs(t, err, expectedErr, "Expected error from initializer")
}

func TestLocalCache_Expiration(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "tempKey"
	value := "tempValue"
	duration := 100 * time.Millisecond

	c.Set(ctx, key, value, &duration)

	// Get the value before expiration
	gotValue, err := c.Get(ctx, key, nil)
	require.NoError(t, err)
	require.Equal(t, value, gotValue, "Value should be retrievable before expiration")

	// Wait for the item to expire
	time.Sleep(duration + 50*time.Millisecond)

	// Try to get the value after expiration
	_, err = c.Get(ctx, key, nil)
	require.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss after expiration")
}

func TestLocalCache_DefaultExpiration(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string](localcache.WithDefaultExpiration(20 * time.Millisecond))

	key := "defaultExpiringKey"
	value := "defaultExpiringValue"
	// No duration specified, should use default expiration
	c.Set(ctx, key, value, nil)

	// Retrieve the item before it expires
	gotValue, err := c.Get(ctx, key, nil)
	require.NoError(t, err)
	require.Equal(t, value, gotValue, "Expected to retrieve the item before expiration")

	// Wait for the default expiration time
	time.Sleep(30 * time.Millisecond)

	// Attempt to retrieve the expired item
	_, err = c.Get(ctx, key, nil)
	require.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss after default expiration")
}

func TestLocalCache_NoExpiration(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string](localcache.WithDefaultExpiration(20 * time.Millisecond))

	key := "noExpireKey"
	value := "noExpireValue"
	duration := localcache.NoExpireDuration
	c.Set(ctx, key, value, &duration)

	// Wait for some time to ensure no expiration
	time.Sleep(30 * time.Millisecond)

	// Item should still be present
	gotValue, err := c.Get(ctx, key, nil)
	require.NoError(t, err)
	require.Equal(t, value, gotValue, "Expected item to persist without expiration")
}

func TestLocalCache_CleanupExpiredItems(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string](localcache.WithCleanupInterval(10 * time.Millisecond))

	key := "expiringKey"
	value := "expiringValue"
	duration := 20 * time.Millisecond
	c.Set(ctx, key, value, &duration)

	// Give some time for the cleanup goroutine to remove expired items
	time.Sleep(30 * time.Millisecond)

	// Assert that the cache instance has a StopCleanup method.
	lc, ok := c.(interface{ StopCleanup() })
	if !ok {
		t.Fatalf("Expected localcache to have a StopCleanup method")
	}
	// Stop the cleanup process
	lc.StopCleanup()

	// Use reflection to access the private items field
	val := reflect.ValueOf(c).Elem().FieldByName("items")
	if !val.IsValid() || val.Kind() != reflect.Map {
		t.Fatalf("Expected items field to be a valid map")
	}

	require.Equal(t, 0, val.Len(), "Expected items to be empty after cleanup")
}

func TestLocalCache_StopCleanup(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string](localcache.WithCleanupInterval(10 * time.Millisecond))

	// Assert that the cache instance has a StopCleanup method.
	lc, ok := c.(interface{ StopCleanup() })
	if !ok {
		t.Fatalf("Expected localcache to have a StopCleanup method")
	}
	// Stop the cleanup process
	lc.StopCleanup()

	key := "expiringKey"
	value := "expiringValue"
	duration := 20 * time.Millisecond
	c.Set(ctx, key, value, &duration)

	// Wait for more than the cleanup interval to check if cleanup is stopped
	time.Sleep(30 * time.Millisecond)

	// Use reflection to access the private items field
	val := reflect.ValueOf(c).Elem().FieldByName("items")
	if !val.IsValid() || val.Kind() != reflect.Map {
		t.Fatalf("Expected items field to be a valid map")
	}

	require.Equal(t, 1, val.Len(), "Expected 1 items in cache")
}

func TestLocalCache_Invalidate(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "testKey"
	value := "testValue"
	duration := time.Minute

	c.Set(ctx, key, value, &duration)

	// Invalidate the key
	err := c.Invalidate(ctx, key)
	require.NoError(t, err)

	// Try to get the invalidated key
	_, err = c.Get(ctx, key, nil)
	require.ErrorIs(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss after invalidation")
}

func TestLocalCache_InvalidateAll(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[string]()

	keys := []string{"key1", "key2", "key3"}
	value := "testValue"
	duration := time.Minute

	for _, key := range keys {
		c.Set(ctx, key, value, &duration)
	}

	// Invalidate all keys
	err := c.InvalidateAll(ctx)
	require.NoError(t, err)

	// Try to get the keys
	for _, key := range keys {
		_, err := c.Get(ctx, key, nil)
		require.ErrorIsf(t, err, cache.ErrCacheMiss, "Expected ErrCacheMiss for key %q after InvalidateAll", key)
	}
}

func TestLocalCache_Concurrency(t *testing.T) {
	ctx := context.Background()
	c := localcache.New[int]()

	key := "concurrentKey"
	duration := time.Minute
	var initializerCallCount int
	var mu sync.Mutex

	initializer := func() (int, *time.Duration, error) {
		mu.Lock()
		initializerCallCount++
		mu.Unlock()
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		return 42, &duration, nil
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := c.Get(ctx, key, initializer)
			require.NoError(t, err)
			require.Equal(t, 42, value, "Value retrieved should be 42")
		}()
	}

	wg.Wait()

	require.Equal(t, 1, initializerCallCount, "Initializer should have been called exactly once")
}
