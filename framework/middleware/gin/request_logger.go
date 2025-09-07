package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
)

// requestLoggerOptions holds configuration options for the RequestLogger middleware.
type requestLoggerOptions struct {
	logger  common_logger.Logger  // Custom logger instance (nil uses default)
	filters []RequestLoggerFilter // Request filters for selective logging
}

// RequestLoggerOption is a function that configures requestLoggerOptions.
type RequestLoggerOption func(*requestLoggerOptions)

// RequestLoggerFilter determines whether a request should be logged.
// Returns true to log the request, false to skip logging entirely.
type RequestLoggerFilter func(*http.Request) bool

// WithRequestLogger sets a custom logger for the request logger middleware.
// If not provided, the middleware uses the default logger from the logger package.
func WithRequestLogger(logger common_logger.Logger) RequestLoggerOption {
	return func(opts *requestLoggerOptions) {
		if logger != nil {
			opts.logger = logger
		}
	}
}

// WithRequestLoggerFilter adds request filters for selective logging.
// All filters must return true for a request to be logged.
func WithRequestLoggerFilter(filters ...RequestLoggerFilter) RequestLoggerOption {
	return func(opts *requestLoggerOptions) {
		opts.filters = append(opts.filters, filters...)
	}
}

// RequestLogger returns Gin middleware that logs detailed HTTP request and response information.
// Captures comprehensive request metadata, measures response times, and injects an enhanced
// logger with request context into the request for downstream handlers to use.
//
// Functionality:
//   - Logs request details, such as method, route, query parameters, client IP, and user agent.
//   - Measures and logs the request latency and response status code.
//   - Allows filtering of requests to determine whether they should be logged.
//   - Injects an augmented logger with request-specific fields into the request context for downstream use.
//
// Key Features:
//   - Custom Logger: Use `WithRequestLogger` to provide a custom logger. If not provided, a default logger is used.
//   - Request Filters: Use `WithRequestLoggerFilter` to specify one or more filters. Requests that do not pass the filters will not be logged.
//   - Request Context Integration: The middleware adds an augmented logger to the request context, allowing downstream handlers to use it for logging.
//
// Example Usage:
//
//	router.Use(
//		RequestLogger(
//			WithRequestLogger(customLogger), // Use a custom logger.
//			WithRequestLoggerFilter(func(req *http.Request) bool {
//				// Skip logging for health check routes.
//				return req.URL.Path != "/health"
//			}),
//		),
//	)
func RequestLogger(opts ...RequestLoggerOption) gin.HandlerFunc {
	// Set default options.
	options := &requestLoggerOptions{
		logger: common_logger.NewDefaultLogger(),
	}

	// Apply any user-provided options.
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gin.Context) {
		// Skip logging based on the filter function.
		for _, filter := range options.filters {
			if !filter(c.Request) {
				c.Next()
				return
			}
		}

		// Start timer.
		startTime := time.Now()

		// Create a logger with request-specific fields.
		requestID, _ := GetRequestIDFromContext(c.Request.Context())
		loggerWithFields := options.logger.WithFields(common_logger.Fields{
			"request": common_logger.Fields{
				"method":      c.Request.Method,
				"route":       c.FullPath(),
				"path":        c.Request.URL.Path,
				"query":       c.Request.URL.RawQuery,
				"request_uri": c.Request.RequestURI,
				"client_ip":   c.ClientIP(),
				"user_agent":  c.Request.UserAgent(),
				"request_id":  requestID,
			},
		})

		// Store the augmented logger in the context for downstream use.
		ctx := common_logger.NewContext(c.Request.Context(), loggerWithFields)
		c.Request = c.Request.WithContext(ctx)

		// Process the request.
		c.Next()

		// Calculate latency.
		latency := time.Since(startTime)
		// Get the status code of the response.
		statusCode := c.Writer.Status()
		// Log the request information.
		loggerWithFields.Info(ctx, "Request information", common_logger.Fields{
			"response": common_logger.Fields{
				"status_code": statusCode,
				"latency_ms":  latency.Milliseconds(),
				"latency_s":   latency.Seconds(),
			},
		})
	}
}
