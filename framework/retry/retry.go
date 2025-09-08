package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

//go:generate mockgen -source=./retry.go -destination=./mocks/retry.go -package=retry_mocks

// ErrInvalidConfig is returned when the Retrier configuration is invalid.
var ErrInvalidConfig = errors.New("invalid retry configuration")

// Config holds retry configuration parameters including attempt limits and backoff behavior.
type Config struct {
	MaxAttempts int      // Maximum number of retry attempts (must be >= 1)
	Backoff     Strategy // Backoff strategy for calculating delays between retries
}

// Validate checks if the retry configuration is valid and properly configured.
func (c *Config) Validate() error {
	if c.MaxAttempts < 1 {
		return fmt.Errorf("%w: maxAttempts must be at least 1", ErrInvalidConfig)
	}
	if c.Backoff == nil {
		return fmt.Errorf("%w: backoff strategy must be provided", ErrInvalidConfig)
	}
	// Validate the backoff strategy
	if err := c.Backoff.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	return nil
}

// Retrier defines the interface for executing operations with retry logic.
type Retrier interface {
	// ExecuteWithRetry attempts to execute a function with retry logic and context support.
	ExecuteWithRetry(ctx context.Context, fn RetryFunc, retryOn RetryOnFunc) error
}

// retrier implements the Retrier interface with configurable retry behavior.
type retrier struct {
	config Config // Retry configuration including max attempts and backoff strategy
}

// NewRetrier creates a new Retrier with the specified configuration.
// Validates configuration before creating the retrier instance to ensure proper setup.
//
// Parameters:
//   - config: Retry configuration with max attempts and backoff strategy
//
// Returns:
//   - Retrier: Configured retry instance
//   - error: ErrInvalidConfig if configuration is invalid
//
// Example:
//
//	backoff, _ := NewFixedBackoffStrategy(2*time.Second)
//	retrier, err := NewRetrier(Config{
//	    MaxAttempts: 5,
//	    Backoff:     backoff,
//	})
//	if err != nil {
//	    log.Fatal("Failed to create retrier:", err)
//	}
func NewRetrier(config Config) (Retrier, error) {
	if err := config.Validate(); err != nil {
		return nil, err // Already wrapped with ErrInvalidConfig
	}
	return &retrier{config: config}, nil
}

// RetryFunc represents a function that can be retried on failure.
// Should return nil on success or an error on failure that may trigger retry.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - error: nil on success, error on failure
type RetryFunc func(ctx context.Context) error

// RetryOnFunc determines whether a retry should be attempted based on the attempt and error.
// Provides flexible control over retry conditions, allowing custom error classification.
//
// Parameters:
//   - attempt: Current attempt number
//   - err: Error from the failed operation
//
// Returns:
//   - bool: true to retry the operation, false to stop and return the error
type RetryOnFunc func(attempt int, err error) bool

// ExecuteWithRetry attempts to execute the given function `fn` with retry logic.
// It retries up to `MaxAttempts` using the configured backoff strategy, and stops
// retrying based on the `retryOn` function or if the context is canceled.
//
// If the function succeeds (returns `nil`), the retry loop exits immediately.
// If the function fails (returns an error), it will retry based on `retryOn`.
//
// Parameters:
//   - ctx: A context that allows for request cancellation. If `ctx` is canceled,
//     the function stops retrying and returns `ctx.Err()`.
//   - fn: The function to be executed with retry logic. It must accept a `context.Context`
//     and return an `error`. Returning `nil` indicates success.
//   - retryOn: A function that determines if a retry should occur based on the current
//     attempt number and the returned error. It should return `true` to retry and `false`
//     to stop retrying.
//
// Returns:
//   - `nil` if the function `fn` succeeds within the allowed retries.
//   - The last encountered `error` if all retry attempts fail.
//   - `ctx.Err()` if the context is canceled before completion.
func (r *retrier) ExecuteWithRetry(ctx context.Context, fn RetryFunc, retryOn RetryOnFunc) error {
	var err error
	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		if err = fn(ctx); err == nil {
			return nil // Success, stop retrying
		}

		// Check if this error should trigger a retry
		if retryOn != nil && !retryOn(attempt+1, err) {
			return err // Immediate failure
		}

		// Check if context is canceled before sleeping
		select {
		case <-ctx.Done():
			return ctx.Err() // Stop retries if context is canceled
		case <-time.After(r.config.Backoff.Next(attempt)):
			// Continue to next retry
		}
	}
	return err
}
