package main

import (
	"context"
	"fmt"
	"time"

	local "github.com/kittipat1413/go-common/framework/lockmanager/locallock"
)

func main() {
	ctx := context.Background()

	// Create a LocalLockManager with a cleanup interval of 1 second
	manager := local.NewLocalLockManager(
		local.WithCleanupInterval(1 * time.Second),
	)

	key := "example-lock"
	ttl := 3 * time.Second

	// Acquire a lock
	token, err := manager.Acquire(ctx, key, ttl)
	if err != nil {
		fmt.Println("Failed to acquire lock:", err)
		return
	}
	fmt.Printf("Lock acquired for key '%s' with token: %s\n", key, token)

	// Try to acquire the same lock with a different token (should fail)
	_, err = manager.Acquire(ctx, key, ttl, "different-token")
	if err != nil {
		fmt.Println("Could not acquire lock with different token (as expected):", err)
	}

	// Extend the lock with the same token
	_, err = manager.Acquire(ctx, key, ttl, token)
	if err != nil {
		fmt.Println("Failed to reacquire lock with same token:", err)
		return
	}
	fmt.Println("Lock reacquired to extend TTL")

	// Release the lock
	err = manager.Release(ctx, key, token)
	if err != nil {
		fmt.Println("Failed to release lock:", err)
		return
	}
	fmt.Println("Lock released successfully")

	// Try to release again (noop)
	err = manager.Release(ctx, key, token)
	if err != nil {
		fmt.Println("Second release failed:", err)
	} else {
		fmt.Println("Second release was a no-op")
	}

	// Shutdown cleanup goroutine before exit
	llm, ok := manager.(interface{ Stop() })
	if ok {
		llm.Stop()
	}
	fmt.Println("Cleanup stopped. Done.")
}
