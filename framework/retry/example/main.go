package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kittipat1413/go-common/framework/retry"
)

var (
	ErrNetworkIssue = errors.New("network issue")
	ErrDatabaseDown = errors.New("database down")
)

func main() {
	ctx := context.Background()

	// Example 1: Fixed Backoff
	fmt.Println("\n🚀 Example 1: Fixed Backoff")
	runExample(ctx, retry.Config{
		MaxAttempts: 3,
		Backoff:     &retry.FixedBackoff{Interval: 1 * time.Second},
	})

	// Example 2: Jitter Backoff
	fmt.Println("\n🚀 Example 2: Jitter Backoff")
	runExample(ctx, retry.Config{
		MaxAttempts: 3,
		Backoff:     &retry.JitterBackoff{BaseDelay: 1 * time.Second, MaxJitter: 1 * time.Second},
	})

	// Example 3: Exponential Backoff
	fmt.Println("\n🚀 Example 3: Exponential Backoff")
	runExample(ctx, retry.Config{
		MaxAttempts: 4,
		Backoff:     &retry.ExponentialBackoff{BaseDelay: 500 * time.Millisecond, Factor: 2.0, MaxDelay: 5 * time.Second},
	})

	// Example 4: Context Timeout Handling
	fmt.Println("\n🚀 Example 4: Context Timeout Handling")
	runWithTimeoutExample()

	// Example 5: Conditional Retries (Retry Only on Network Issues)
	fmt.Println("\n🚀 Example 5: Conditional Retries")
	runConditionalRetryExample(ctx)
}

// runExample executes a retry operation with the given configuration.
func runExample(ctx context.Context, config retry.Config) {
	retrier, err := retry.NewRetrier(config)
	if err != nil {
		log.Fatalf("Failed to create retrier: %v", err)
	}

	start := time.Now()

	err = retrier.ExecuteWithRetry(ctx, failingOperation, func(attempt int, err error) bool {
		fmt.Printf("[%s] Attempt %d failed: %v\n", time.Since(start).Round(time.Millisecond), attempt, err)
		return true // Always retry
	})

	if err != nil {
		fmt.Printf("[%s] Final failure with error: %v\n", time.Since(start).Round(time.Millisecond), err)
	} else {
		fmt.Printf("[%s] Operation succeeded!\n", time.Since(start).Round(time.Millisecond))
	}
}

// failingOperation simulates an API call that always fails.
func failingOperation(ctx context.Context) error {
	fmt.Println("🔄 Attempting operation...")
	return ErrNetworkIssue
}

// runWithTimeoutExample demonstrates retry handling with a context timeout.
func runWithTimeoutExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	retrier, _ := retry.NewRetrier(retry.Config{
		MaxAttempts: 5,
		Backoff:     &retry.FixedBackoff{Interval: 2 * time.Second},
	})

	start := time.Now()

	err := retrier.ExecuteWithRetry(ctx, failingOperation, func(attempt int, err error) bool {
		fmt.Printf("[%s] Attempt %d failed: %v\n", time.Since(start).Round(time.Millisecond), attempt, err)
		return true // Always retry
	})

	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("[%s] ❌ Retries stopped due to context timeout\n", time.Since(start).Round(time.Millisecond))
	} else {
		fmt.Printf("[%s] Final failure with error: %v\n", time.Since(start).Round(time.Millisecond), err)
	}
}

// runConditionalRetryExample retries only on specific errors.
func runConditionalRetryExample(ctx context.Context) {
	retrier, _ := retry.NewRetrier(retry.Config{
		MaxAttempts: 5,
		Backoff:     &retry.FixedBackoff{Interval: 1 * time.Second},
	})

	start := time.Now()

	err := retrier.ExecuteWithRetry(ctx, func(ctx context.Context) error {
		fmt.Printf("[%s] 🔄 Trying a database operation...\n", time.Since(start).Round(time.Millisecond))
		return ErrDatabaseDown // Simulate different error
	}, func(attempt int, err error) bool {
		if errors.Is(err, ErrNetworkIssue) { // Retry only on network issues
			return true
		}
		fmt.Printf("❌ Stopping retries due to non-retryable error: %v\n", err)
		return false
	})

	if err != nil {
		fmt.Printf("[%s] Final failure with error: %v\n", time.Since(start).Round(time.Millisecond), err)
	} else {
		fmt.Printf("[%s] Operation succeeded!\n", time.Since(start).Round(time.Millisecond))
	}
}
