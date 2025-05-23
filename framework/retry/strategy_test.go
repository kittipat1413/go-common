package retry_test

import (
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFixedBackoffStrategy(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		strategy, err := retry.NewFixedBackoffStrategy(500 * time.Millisecond)
		require.NoError(t, err)
		require.NotNil(t, strategy)

		assert.Equal(t, 500*time.Millisecond, strategy.Next(0))
	})

	t.Run("invalid config", func(t *testing.T) {
		strategy, err := retry.NewFixedBackoffStrategy(0)
		require.Error(t, err)
		assert.Nil(t, strategy)
	})
}

func TestFixedBackoff_Validate(t *testing.T) {
	tests := []struct {
		name      string
		interval  time.Duration
		expectErr bool
	}{
		{"Valid interval", 100 * time.Millisecond, false},
		{"Invalid zero interval", 0, true},
		{"Invalid negative interval", -10 * time.Millisecond, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := retry.FixedBackoff{Interval: tt.interval}
			err := b.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFixedBackoff_Next(t *testing.T) {
	b := retry.FixedBackoff{Interval: 100 * time.Millisecond}
	require.NoError(t, b.Validate())

	for i := 0; i < 5; i++ {
		assert.Equal(t, 100*time.Millisecond, b.Next(i), "FixedBackoff should return the same interval")
	}
}

func TestNewJitterBackoffStrategy(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		strategy, err := retry.NewJitterBackoffStrategy(100*time.Millisecond, 50*time.Millisecond)
		require.NoError(t, err)
		require.NotNil(t, strategy)

		for i := 0; i < 10; i++ {
			delay := strategy.Next(i)
			assert.GreaterOrEqual(t, delay, 100*time.Millisecond)
			assert.LessOrEqual(t, delay, 150*time.Millisecond)
		}
	})

	t.Run("invalid baseDelay", func(t *testing.T) {
		strategy, err := retry.NewJitterBackoffStrategy(0, 50*time.Millisecond)
		require.Error(t, err)
		assert.Nil(t, strategy)
	})

	t.Run("invalid maxJitter", func(t *testing.T) {
		strategy, err := retry.NewJitterBackoffStrategy(100*time.Millisecond, -1*time.Millisecond)
		require.Error(t, err)
		assert.Nil(t, strategy)
	})
}

func TestJitterBackoff_Validate(t *testing.T) {
	tests := []struct {
		name      string
		baseDelay time.Duration
		maxJitter time.Duration
		expectErr bool
	}{
		{"Valid jitter", 100 * time.Millisecond, 50 * time.Millisecond, false},
		{"Invalid baseDelay", 0, 50 * time.Millisecond, true},
		{"Invalid negative jitter", 100 * time.Millisecond, -10 * time.Millisecond, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := retry.JitterBackoff{BaseDelay: tt.baseDelay, MaxJitter: tt.maxJitter}
			err := b.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJitterBackoff_Next(t *testing.T) {
	b := retry.JitterBackoff{BaseDelay: 100 * time.Millisecond, MaxJitter: 50 * time.Millisecond}
	require.NoError(t, b.Validate())

	for i := 0; i < 10; i++ {
		delay := b.Next(i)
		assert.GreaterOrEqual(t, delay, 100*time.Millisecond, "JitterBackoff should not be less than BaseDelay")
		assert.LessOrEqual(t, delay, 150*time.Millisecond, "JitterBackoff should not exceed BaseDelay + MaxJitter")
	}
}

func TestNewExponentialBackoffStrategy(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		strategy, err := retry.NewExponentialBackoffStrategy(100*time.Millisecond, 2.0, 2*time.Second)
		require.NoError(t, err)
		require.NotNil(t, strategy)

		expected := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			400 * time.Millisecond,
			800 * time.Millisecond,
			1600 * time.Millisecond,
			2000 * time.Millisecond, // capped at MaxDelay
		}

		for i, want := range expected {
			got := strategy.Next(i)
			assert.Equal(t, want, got)
		}
	})

	t.Run("invalid baseDelay", func(t *testing.T) {
		strategy, err := retry.NewExponentialBackoffStrategy(0, 2.0, 2*time.Second)
		require.Error(t, err)
		assert.Nil(t, strategy)
	})

	t.Run("invalid factor", func(t *testing.T) {
		strategy, err := retry.NewExponentialBackoffStrategy(100*time.Millisecond, 1.0, 2*time.Second)
		require.Error(t, err)
		assert.Nil(t, strategy)
	})

	t.Run("invalid maxDelay", func(t *testing.T) {
		strategy, err := retry.NewExponentialBackoffStrategy(1*time.Second, 2.0, 500*time.Millisecond)
		require.Error(t, err)
		assert.Nil(t, strategy)
	})
}

func TestExponentialBackoff_Validate(t *testing.T) {
	tests := []struct {
		name      string
		baseDelay time.Duration
		factor    float64
		maxDelay  time.Duration
		expectErr bool
	}{
		{"Valid exponential", 100 * time.Millisecond, 2.0, 5 * time.Second, false},
		{"Invalid baseDelay", 0, 2.0, 5 * time.Second, true},
		{"Invalid factor", 100 * time.Millisecond, 1.0, 5 * time.Second, true},
		{"Invalid maxDelay", 100 * time.Millisecond, 2.0, 50 * time.Millisecond, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := retry.ExponentialBackoff{BaseDelay: tt.baseDelay, Factor: tt.factor, MaxDelay: tt.maxDelay}
			err := b.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExponentialBackoff_Next(t *testing.T) {
	b := retry.ExponentialBackoff{BaseDelay: 100 * time.Millisecond, Factor: 2.0, MaxDelay: 5 * time.Second}
	require.NoError(t, b.Validate())

	expectedDelays := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
		1600 * time.Millisecond,
		3200 * time.Millisecond,
		5000 * time.Millisecond, // Should cap at MaxDelay
	}

	for i, expected := range expectedDelays {
		actual := b.Next(i)
		assert.Equal(t, expected, actual, "ExponentialBackoff should match expected delay")
	}
}
