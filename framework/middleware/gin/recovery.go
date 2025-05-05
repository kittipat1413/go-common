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

// Recovery returns a Gin middleware that recovers from panics during request handling and logs the error.
//
// The middleware performs the following tasks:
//  1. Recovers from any panic that occurs in the middleware chain or route handlers.
//  2. Logs the panic information (including HTTP method and route) using the provided logger or retrieves one from the context.
//  3. Calls a custom error handler to generate a response, or defaults to a 500 Internal Server Error response if no custom handler is provided.
//
// Key Features:
//   - Custom Logger: Use `WithRecoveryLogger` to specify a logger for capturing panic details. If no logger is provided, the middleware attempts to retrieve one from the context.
//   - Custom Error Handler: Use `WithRecoveryHandler` to define a custom function for handling the panic and responding to the client.
//   - Default Behavior: If no logger or custom handler is specified, the middleware logs the panic (using the context logger) and returns a 500 Internal Server Error response.
//
// Example Usage:
//
//	router.Use(
//		Recovery(
//	    	WithRecoveryLogger(logger), // Use a custom logger for panic recovery.
//	    	WithRecoveryHandler(func(c *gin.Context, err interface{}) {
//	        	// Custom error handling logic (e.g., custom JSON response).
//	        	c.AbortWithStatusJSON(
//					http.StatusInternalServerError,
//					gin.H{"message": "Something went wrong. Please contact support.", "details": err}
//				)
//	    	}),
//		),
//	)
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
				if !c.Writer.Written() {
					options.handler(c, err) // Write an error response only if no response has been written.
				}
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
