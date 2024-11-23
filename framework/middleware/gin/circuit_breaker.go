package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// circuitBreakerOptions holds configuration options for the CircuitBreaker middleware.
type circuitBreakerOptions struct {
	settings        gobreaker.Settings     // Settings for the circuit breaker.
	statusThreshold int                    // Status code threshold to consider as errors.
	filters         []CircuitBreakerFilter // Filters to determine whether to apply the circuit breaker.
	onError         func(c *gin.Context)   // Custom error handler for circuit breaker failures.
}

// CircuitBreakerOption is a function that configures circuitBreakerOptions.
type CircuitBreakerOption func(*circuitBreakerOptions)

// CircuitBreakerFilter is a function that determines whether a request should be wrapped by the circuit breaker.
type CircuitBreakerFilter func(*http.Request) bool

// WithCircuitBreakerSettings allows customizing the circuit breaker settings.
func WithCircuitBreakerSettings(settings gobreaker.Settings) CircuitBreakerOption {
	return func(opts *circuitBreakerOptions) {
		opts.settings = settings
	}
}

// WithCircuitBreakerStatusThreshold sets the status code threshold for error detection.
func WithCircuitBreakerStatusThreshold(threshold int) CircuitBreakerOption {
	return func(opts *circuitBreakerOptions) {
		opts.statusThreshold = threshold
	}
}

// WithCircuitBreakerFilter adds one or more filters to determine whether the circuit breaker applies to specific requests.
func WithCircuitBreakerFilter(filters ...CircuitBreakerFilter) CircuitBreakerOption {
	return func(opts *circuitBreakerOptions) {
		opts.filters = append(opts.filters, filters...)
	}
}

// WithCircuitBreakerErrorHandler sets a custom error handler for circuit breaker failures.
func WithCircuitBreakerErrorHandler(handler func(c *gin.Context)) CircuitBreakerOption {
	return func(opts *circuitBreakerOptions) {
		opts.onError = handler
	}
}

// CircuitBreaker creates a Gin middleware that wraps route handlers with a circuit breaker.
//
// This middleware monitors request failures and automatically trips the circuit breaker when failures exceed
// a configured threshold. Once tripped, requests are blocked and a fallback error response is returned
// until the circuit breaker recovers.
//
// Features:
//   - Failure Detection: Automatically detects failures based on HTTP status codes or custom logic.
//   - Customizable Behavior: Configure the circuit breaker settings, error thresholds, and fallback handlers.
//   - Selective Application: Use filters to apply the circuit breaker only to specific routes or requests.
//
// Example Usage:
//
//	CircuitBreaker(
//		WithCircuitBreakerSettings(gobreaker.Settings{
//			Name: "CustomCircuitBreaker",
//			ReadyToTrip: func(counts gobreaker.Counts) bool {
//				return counts.ConsecutiveFailures > 3 // Trip after 3 consecutive failures.
//			},
//		}),
//		WithCircuitBreakerStatusThreshold(http.StatusBadRequest), // Treat >= 400 as errors.
//		WithCircuitBreakerErrorHandler(func(c *gin.Context) {
//			// Custom error handler for circuit breaker failures.
//			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
//				"error": "Custom error: Circuit breaker activated.",
//			})
//		}),
//		WithCircuitBreakerFilter(func(req *http.Request) bool {
//			return req.URL.Path != "/health" // Skip circuit breaker for health checks.
//		}),
//	),
func CircuitBreaker(opts ...CircuitBreakerOption) gin.HandlerFunc {
	// Set default options.
	options := &circuitBreakerOptions{
		settings: gobreaker.Settings{
			Name: "CircuitBreaker Middleware",
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				// Default: trip the circuit after 5 consecutive failures.
				return counts.ConsecutiveFailures >= 5
			},
		},
		statusThreshold: http.StatusInternalServerError,    // Default error status code threshold (500).
		onError:         defaultCircuitBreakerErrorHandler, // Default error handler.
	}

	// Apply user-provided options.
	for _, opt := range opts {
		opt(options)
	}

	// Create the circuit breaker.
	cb := gobreaker.NewCircuitBreaker(options.settings)

	return func(c *gin.Context) {
		// Filters are evaluated sequentially, and the circuit breaker is applied only if all filters return true.
		for _, filter := range options.filters {
			if !filter(c.Request) {
				c.Next()
				return
			}
		}

		// Execute the handler within the circuit breaker.
		_, err := cb.Execute(func() (interface{}, error) {
			c.Next()

			// Detect errors based on status code threshold.
			if c.Writer.Status() >= options.statusThreshold {
				return nil, errors.New("status code indicates failure")
			}
			return nil, nil
		})

		// Handle circuit breaker errors.
		if err != nil && !c.Writer.Written() {
			options.onError(c) // Write an error response only if no response has been written.
		}
	}
}

// defaultCircuitBreakerErrorHandler writes a default error response for circuit breaker failures.
func defaultCircuitBreakerErrorHandler(c *gin.Context) {
	c.AbortWithStatusJSON(
		http.StatusServiceUnavailable,
		gin.H{"error": "Service is temporarily unavailable due to high failure rate. Please try again later."},
	)
}
