package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

//go:generate mockgen -source=./retry.go -destination=./mocks/retry.go -package=retry_mocks

// ErrInvalidConfig is returned when the Retrier is misconfigured.
var ErrInvalidConfig = errors.New("invalid retry configuration")

// Config holds the retry settings.
type Config struct {
	MaxAttempts int
	Backoff     Strategy
}

// Validate checks if the config is valid.
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

// Retrier defines the interface for retrying operations.
type Retrier interface {
	ExecuteWithRetry(ctx context.Context, fn RetryFunc, retryOn RetryOnFunc) error
}

// retrier implements the Retrier interface.
type retrier struct {
	config Config
}

// NewRetrier creates a new Retrier with the given configuration.
func NewRetrier(config Config) (Retrier, error) {
	if err := config.Validate(); err != nil {
		return nil, err // Already wrapped with ErrInvalidConfig
	}
	return &retrier{config: config}, nil
}

// RetryFunc is the function signature for retryable operations.
type RetryFunc func(ctx context.Context) error

// RetryOnFunc is the function signature for retryable error checks.
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
