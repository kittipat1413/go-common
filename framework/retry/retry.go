package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

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

// Retrier handles retry logic.
type Retrier struct {
	config Config
}

// NewRetrier creates a new Retrier with the given configuration.
func NewRetrier(config Config) (*Retrier, error) {
	if err := config.Validate(); err != nil {
		return nil, err // Already wrapped with ErrInvalidConfig
	}
	return &Retrier{config: config}, nil
}

// RetryFunc is the function signature for retryable operations.
type RetryFunc func(ctx context.Context) error

// RetryOnFunc is the function signature for retryable error checks.
type RetryOnFunc func(attempt int, err error) bool

// ExecuteWithRetry runs the given function with retry logic.
func (r *Retrier) ExecuteWithRetry(ctx context.Context, fn RetryFunc, retryOn RetryOnFunc) error {
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
