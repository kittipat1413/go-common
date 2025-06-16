package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"

	redsyncLocker "github.com/kittipat1413/go-common/framework/lockmanager/redsync"
)

func main() {
	// Setup Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Initialize LockManager
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

	ctx := context.Background()
	key := "example-lock-key"
	key2 := "example-lock-key-2"
	ttl := 2 * time.Second

	// Acquire lock
	lockToken, err := manager.Acquire(ctx, key, ttl)
	if err != nil {
		log.Fatalf("Failed to acquire lock: %v", err)
	}
	fmt.Printf("Lock acquired with token: %s\n", lockToken)

	// Acquire lock again with the same key but different token (should fail)
	_, err = manager.Acquire(ctx, key, ttl, "lock-again")
	if err != nil {
		fmt.Printf("Failed to acquire lock again (as expected): %v\n", err)
	}

	// Acquire lock on a different key (key2)
	lockToken2, err := manager.Acquire(ctx, key2, ttl)
	if err != nil {
		log.Fatalf("Failed to acquire lock on second key: %v", err)
	}
	fmt.Printf("Lock acquired on second key with token: %s\n", lockToken2)

	// Release lock
	if err := manager.Release(ctx, key, lockToken); err != nil {
		log.Fatalf("Failed to release lock: %v", err)
	}
	fmt.Printf("Lock released successfully.\n")

	// Release lock again (should succeed since the lock was released)
	err = manager.Release(ctx, key, lockToken)
	if err == nil {
		fmt.Printf("Lock released again successfully (should not happen, but no error returned).\n")
	}

	// Simulate work
	time.Sleep(2 * time.Second)

	// Release lock on the expired key2 (should succeed)
	if err := manager.Release(ctx, key2, lockToken2); err != nil {
		log.Fatalf("Failed to release lock on second key: %v", err)
	}
	fmt.Printf("Lock on second key released successfully.\n")

	// Cleanup Redis client
	if err := client.Close(); err != nil {
		log.Fatalf("Failed to close Redis client: %v", err)
	}
	fmt.Printf("Redis client closed successfully.\n")
}
