package retry

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

// Strategy defines a backoff strategy.
type Strategy interface {
	Validate() error
	Next(retryCount int) time.Duration
}

// FixedBackoff applies a fixed delay between retries.
type FixedBackoff struct {
	Interval time.Duration // Delay between retries
}

func (f *FixedBackoff) Validate() error {
	if f.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	return nil
}

func (f *FixedBackoff) Next(retryCount int) time.Duration {
	return f.Interval
}

// JitterBackoff adds randomness to avoid thundering herd.
type JitterBackoff struct {
	BaseDelay time.Duration // Base delay between retries
	MaxJitter time.Duration // Maximum random delay to add
}

func (j *JitterBackoff) Validate() error {
	if j.BaseDelay <= 0 {
		return errors.New("baseDelay must be greater than 0")
	}
	if j.MaxJitter < 0 {
		return errors.New("maxJitter cannot be negative")
	}
	return nil
}

func (j *JitterBackoff) Next(retryCount int) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(j.MaxJitter))) // #nosec G404
	return j.BaseDelay + jitter
}

// ExponentialBackoff increases the delay exponentially.
type ExponentialBackoff struct {
	BaseDelay time.Duration // Initial delay
	Factor    float64       // Growth factor (e.g., 2.0 means double delay each time)
	MaxDelay  time.Duration // Upper limit for delay
}

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

func (e *ExponentialBackoff) Next(retryCount int) time.Duration {
	delay := time.Duration(float64(e.BaseDelay) * math.Pow(e.Factor, float64(retryCount)))
	if delay > e.MaxDelay {
		return e.MaxDelay
	}
	return delay
}
