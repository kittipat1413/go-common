package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kittipat1413/go-common/framework/cache/localcache"
)

func main() {
	ctx := context.Background()
	c := localcache.New[string]()

	key := "greeting"

	// Initializer function
	initializer := func() (string, time.Duration, error) {
		// Simulate data fetching or computation
		time.Sleep(100 * time.Millisecond)
		return "Hello, World!", 5 * time.Minute, nil
	}

	// Get value from cache (will initialize if not present)
	value, err := c.Get(ctx, key, initializer)
	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return
	}
	fmt.Printf("Value for key '%s': %s\n", key, value)

	// Set value manually
	c.Set(ctx, "farewell", "Goodbye!", 10*time.Minute)

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
