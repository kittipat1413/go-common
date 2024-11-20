package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
)

// recoveryOptions holds the configuration for the Recovery middleware.
type recoveryOptions struct {
	logger  common_logger.Logger                  // logger is the custom logger to use. If nil, the logger will be retrieved from the context.
	handler func(c *gin.Context, err interface{}) // handler is the function to handle the recovered panic.
}

// RecoveryOptions is a function that configures recoveryOptions.
type RecoveryOption func(*recoveryOptions)

// WithLogger sets a custom logger for the Recovery middleware.
func WithRecoveryLogger(logger common_logger.Logger) RecoveryOption {
	return func(opts *recoveryOptions) {
		opts.logger = logger
	}
}

// WithRecoveryHandler sets a custom error handler for the Recovery middleware.
func WithRecoveryHandler(handler func(c *gin.Context, err interface{})) RecoveryOption {
	return func(opts *recoveryOptions) {
		opts.handler = handler
	}
}

// Recovery returns a middleware that recovers from panics and handles errors using the provided options.
func Recovery(opts ...RecoveryOption) gin.HandlerFunc {
	// Initialize default options.
	options := &recoveryOptions{
		logger:  nil,                    // Default to nil logger.
		handler: defaultRecoveryHandler, // Use default handler if none provided.
	}

	// Apply provided options to override defaults.
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gin.Context) {
		// Retrieve the logger from options or context.
		logger := options.logger
		if logger == nil {
			logger = common_logger.FromContext(c.Request.Context())
		}

		defer func() {
			// Recover from panic if one occurred.
			if err := recover(); err != nil {
				if logger != nil {
					logger.Error(c.Request.Context(), "Panic recovered", nil, common_logger.Fields{
						"panic_info": common_logger.Fields{
							"method": c.Request.Method,
							"route":  c.FullPath(),
							"error":  err,
						},
					})
				}

				// Call the custom handler to respond to the client.
				options.handler(c, err)
			}
		}()
		c.Next()
	}
}

// defaultRecoveryHandler is the default handler that returns a 500 status code with a generic error message.
func defaultRecoveryHandler(c *gin.Context, _ interface{}) {
	c.AbortWithStatusJSON(
		http.StatusInternalServerError,
		gin.H{"error": "Internal Server Error"},
	)
}
