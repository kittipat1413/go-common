package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrier_ValidationFailure(t *testing.T) {
	t.Run("Invalid MaxAttempts", func(t *testing.T) {
		fixedBackoff, err := retry.NewFixedBackoffStrategy(time.Second)
		require.NoError(t, err)
		_, err = retry.NewRetrier(retry.Config{
			MaxAttempts: 0, // Invalid
			Backoff:     fixedBackoff,
		})
		assert.ErrorIs(t, err, retry.ErrInvalidConfig)
	})
	t.Run("Invalid Backoff", func(t *testing.T) {
		_, err := retry.NewRetrier(retry.Config{
			MaxAttempts: 3,
			Backoff:     nil, // Invalid
		})
		assert.ErrorIs(t, err, retry.ErrInvalidConfig)
	})
	t.Run("Invalid Backoff Strategy", func(t *testing.T) {
		_, err := retry.NewRetrier(retry.Config{
			MaxAttempts: 3,
			Backoff:     &retry.FixedBackoff{Interval: 0}, // directly passing in invalid backoff
		})
		assert.ErrorIs(t, err, retry.ErrInvalidConfig)
	})
}

func TestRetrier_SuccessWithoutRetry(t *testing.T) {
	// Create a fixed backoff strategy with a 10ms interval
	fixedBackoff, err := retry.NewFixedBackoffStrategy(10 * time.Millisecond)
	require.NoError(t, err)
	// Create a retrier with the fixed backoff strategy
	config := retry.Config{
		MaxAttempts: 3,
		Backoff:     fixedBackoff,
	}
	retrier, err := retry.NewRetrier(config)
	require.NoError(t, err)

	callCount := 0
	err = retrier.ExecuteWithRetry(context.Background(), func(ctx context.Context) error {
		callCount++
		return nil // No error, should not retry
	}, func(attempt int, err error) bool {
		return true
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "Should execute only once without retry")
}

func TestRetrier_RetriesOnFailure(t *testing.T) {
	// Create a fixed backoff strategy with a 10ms interval
	fixedBackoff, err := retry.NewFixedBackoffStrategy(10 * time.Millisecond)
	require.NoError(t, err)
	// Create a retrier with the fixed backoff strategy
	config := retry.Config{
		MaxAttempts: 3,
		Backoff:     fixedBackoff,
	}
	retrier, err := retry.NewRetrier(config)
	require.NoError(t, err)

	callCount := 0
	err = retrier.ExecuteWithRetry(context.Background(), func(ctx context.Context) error {
		callCount++
		return errors.New("error")
	}, func(attempt int, err error) bool {
		return true // Always retry
	})

	assert.Error(t, err)
	assert.Equal(t, 3, callCount, "Should retry exactly maxAttempts times")
}

func TestRetrier_StopsOnSuccess(t *testing.T) {
	// Create a fixed backoff strategy with a 10ms interval
	fixedBackoff, err := retry.NewFixedBackoffStrategy(10 * time.Millisecond)
	require.NoError(t, err)
	// Create a retrier with the fixed backoff strategy
	config := retry.Config{
		MaxAttempts: 5,
		Backoff:     fixedBackoff,
	}
	retrier, err := retry.NewRetrier(config)
	require.NoError(t, err)

	callCount := 0
	err = retrier.ExecuteWithRetry(context.Background(), func(ctx context.Context) error {
		callCount++
		if callCount == 3 {
			return nil // Succeed on 3rd attempt
		}
		return errors.New("error")
	}, func(attempt int, err error) bool {
		return true
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount, "Should stop retrying after success")
}

func TestRetrier_OnlyRetriesOnSpecificErrors(t *testing.T) {
	var retryableError = errors.New("retryable error")

	// Create a fixed backoff strategy with a 10ms interval
	fixedBackoff, err := retry.NewFixedBackoffStrategy(10 * time.Millisecond)
	require.NoError(t, err)
	// Create a retrier with the fixed backoff strategy
	config := retry.Config{
		MaxAttempts: 5,
		Backoff:     fixedBackoff,
	}
	retrier, err := retry.NewRetrier(config)
	require.NoError(t, err)

	callCount := 0
	err = retrier.ExecuteWithRetry(context.Background(), func(ctx context.Context) error {
		callCount++
		if callCount == 2 {
			return errors.New("fatal error") // Should stop immediately
		}
		return retryableError
	}, func(attempt int, err error) bool {
		return errors.Is(err, retryableError)
	})

	assert.Error(t, err)
	assert.Equal(t, 2, callCount, "Should stop retrying on fatal error")
}

func TestRetrier_ContextCancellation(t *testing.T) {
	// Create a fixed backoff strategy with a 10ms interval
	fixedBackoff, err := retry.NewFixedBackoffStrategy(10 * time.Millisecond)
	require.NoError(t, err)
	// Create a retrier with the fixed backoff strategy
	config := retry.Config{
		MaxAttempts: 5,
		Backoff:     fixedBackoff,
	}
	retrier, err := retry.NewRetrier(config)
	require.NoError(t, err)

	callCount := 0
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	err = retrier.ExecuteWithRetry(ctx, func(ctx context.Context) error {
		callCount++
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return errors.New("error")
		}
	}, func(attempt int, err error) bool {
		return true
	})

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, callCount, 5, "Should stop retries due to context timeout")
}
