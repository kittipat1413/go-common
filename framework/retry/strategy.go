package retry

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Strategy defines the interface for backoff delay calculation between retry attempts.
// Implementations provide different algorithms for calculating wait times based on retry count.
type Strategy interface {
	// Validate checks if the strategy configuration is valid and properly set up.
	Validate() error

	// Next calculates the delay duration before the next retry attempt.
	Next(retryCount int) time.Duration
}

// FixedBackoff applies a constant delay between all retry attempts.
// Simple strategy that waits the same duration regardless of retry count.
type FixedBackoff struct {
	Interval time.Duration // Fixed delay between retries (must be > 0)
}

// NewFixedBackoffStrategy creates a new fixed backoff strategy with validation.
// Applies the same delay duration for all retry attempts.
//
// Parameters:
//   - interval: Fixed delay between retries (must be positive)
//
// Returns:
//   - Strategy: Configured fixed backoff strategy
//   - error: Validation error if interval is invalid
func NewFixedBackoffStrategy(interval time.Duration) (Strategy, error) {
	fixedBackoff := &FixedBackoff{
		Interval: interval,
	}
	if err := fixedBackoff.Validate(); err != nil {
		return nil, err
	}
	return fixedBackoff, nil
}

// Validate ensures the interval is positive.
func (f *FixedBackoff) Validate() error {
	if f.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	return nil
}

// Next returns the fixed interval regardless of retry count.
func (f *FixedBackoff) Next(retryCount int) time.Duration {
	return f.Interval
}

// JitterBackoff adds randomness to base delay to avoid thundering herd problems.
// Prevents multiple clients from retrying simultaneously by adding random delays.
type JitterBackoff struct {
	BaseDelay time.Duration // Base delay before adding jitter (must be > 0)
	MaxJitter time.Duration // Maximum random delay to add (must be >= 0)
	randMu    sync.Mutex    // Protects random number generation
}

// NewJitterBackoffStrategy creates a new jitter backoff strategy with validation.
// Combines base delay with random jitter to prevent synchronized retry attempts.
//
// Parameters:
//   - baseDelay: Minimum delay before adding randomness (must be positive)
//   - maxJitter: Maximum random delay to add (must be non-negative)
//
// Returns:
//   - Strategy: Configured jitter backoff strategy
//   - error: Validation error if parameters are invalid
func NewJitterBackoffStrategy(baseDelay time.Duration, maxJitter time.Duration) (Strategy, error) {
	jitterBackoff := &JitterBackoff{
		BaseDelay: baseDelay,
		MaxJitter: maxJitter,
	}
	if err := jitterBackoff.Validate(); err != nil {
		return nil, err
	}
	return jitterBackoff, nil
}

// Validate ensures baseDelay is positive and maxJitter is non-negative.
func (j *JitterBackoff) Validate() error {
	if j.BaseDelay <= 0 {
		return errors.New("baseDelay must be greater than 0")
	}
	if j.MaxJitter < 0 {
		return errors.New("maxJitter cannot be negative")
	}
	return nil
}

// Next returns base delay plus random jitter up to maxJitter.
func (j *JitterBackoff) Next(retryCount int) time.Duration {
	j.randMu.Lock()
	jitter := time.Duration(rand.Int63n(int64(j.MaxJitter))) // #nosec G404
	j.randMu.Unlock()

	return j.BaseDelay + jitter
}

// ExponentialBackoff increases delay exponentially with each retry attempt.
// Starts with baseDelay and multiplies by factor for each retry, capped at maxDelay.
type ExponentialBackoff struct {
	BaseDelay time.Duration // Initial delay for first retry (must be > 0)
	Factor    float64       // Exponential growth factor (must be > 1.0, typically 2.0)
	MaxDelay  time.Duration // Upper limit to prevent excessive delays (must be >= baseDelay)
}

// NewExponentialBackoffStrategy creates a new exponential backoff strategy with validation.
// Delay grows exponentially: baseDelay * factor^retryCount, capped at maxDelay.
//
// Parameters:
//   - baseDelay: Initial delay duration (must be positive)
//   - factor: Exponential multiplier per retry (must be > 1.0, common values: 1.5, 2.0)
//   - maxDelay: Maximum delay cap to prevent infinite growth (must be >= baseDelay)
//
// Returns:
//   - Strategy: Configured exponential backoff strategy
//   - error: Validation error if parameters are invalid
func NewExponentialBackoffStrategy(baseDelay time.Duration, factor float64, maxDelay time.Duration) (Strategy, error) {
	exponentialBackoff := &ExponentialBackoff{
		BaseDelay: baseDelay,
		Factor:    factor,
		MaxDelay:  maxDelay,
	}
	if err := exponentialBackoff.Validate(); err != nil {
		return nil, err
	}
	return exponentialBackoff, nil
}

// Validate ensures baseDelay is positive, factor enables growth, and maxDelay is reasonable.
func (e *ExponentialBackoff) Validate() error {
	if e.BaseDelay <= 0 {
		return errors.New("baseDelay must be greater than 0")
	}
	if e.Factor <= 1.0 {
		return errors.New("factor must be greater than 1.0 for exponential growth")
	}
	if e.MaxDelay < e.BaseDelay {
		return errors.New("maxDelay must be greater than or equal to baseDelay")
	}
	return nil
}

// Next calculates exponential delay: baseDelay * factor^retryCount, capped at maxDelay.
func (e *ExponentialBackoff) Next(retryCount int) time.Duration {
	delay := time.Duration(float64(e.BaseDelay) * math.Pow(e.Factor, float64(retryCount)))
	if delay > e.MaxDelay {
		return e.MaxDelay
	}
	return delay
}
