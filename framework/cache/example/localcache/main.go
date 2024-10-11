package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kittipat1413/go-common/framework/cache/localcache"
)

func main() {
	ctx := context.Background()
	// Create a new cache with default expiration of 10 minutes and cleanup interval of 5 minutes
	c := localcache.New[string](localcache.WithDefaultExpiration(10*time.Minute), localcache.WithCleanupInterval(5*time.Minute))

	// Initializer function
	initializer := func() (string, *time.Duration, error) {
		// Simulate data fetching or computation
		time.Sleep(100 * time.Millisecond)

		expire := 100 * time.Millisecond
		return "Hello, World!", &expire, nil
	}

	key := "greeting"
	// Get value from cache (will initialize if not present)
	value, err := c.Get(ctx, key, initializer)
	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return
	}
	fmt.Printf("Value for key '%s': %s\n", key, value)

	// Get value from cache after 100 milliseconds (should be expired)
	time.Sleep(100 * time.Millisecond)
	_, err = c.Get(ctx, key, nil)
	if err != nil {
		fmt.Printf("Error getting for key '%s': value: %v\n", key, err)
	}

	// Set value manually with no expiration
	noExpiration := localcache.NoExpireDuration
	c.Set(ctx, "farewell", "Goodbye!", &noExpiration)

	// Get the manually set value
	value, err = c.Get(ctx, "farewell", nil)
	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return
	}
	fmt.Printf("Value for key 'farewell': %s\n", value)

	// Invalidate a key
	err = c.InvalidateAll(ctx)
	if err != nil {
		fmt.Printf("Error invalidating cache: %v\n", err)
		return
	}

	// Try to get the invalidated value (will re-initialize)
	value, err = c.Get(ctx, key, initializer)
	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return
	}
	fmt.Printf("Value for key '%s' after invalidation: %s\n", key, value)

	// Try to get the invalidated value
	_, err = c.Get(ctx, "farewell", nil)
	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return
	}
}
