package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// requestIDKey is an unexported type for context keys defined in this package.
type requestIDKey struct{}

// requestIDContextKey is the key for request ID values in context.
var requestIDContextKey = &requestIDKey{}

// GetRequestIDFromContext retrieves the request ID from the context.
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDContextKey).(string)
	return requestID, ok
}

// DefaultRequestIDHeader is the default header name where the request ID is stored.
const DefaultRequestIDHeader = "X-Request-ID"

// requestIDOptions holds configuration options for the RequestID middleware.
type requestIDOptions struct {
	headerName       string                    // The header name to use for the request ID.
	generatorMode    requestIDGeneratorMode    // The mode to use for the request ID generator.
	generator        RequestIDGenerator        // The function used to generate a new request ID.
	contextGenerator RequestIDContextGenerator // The function used to generate a new request ID based on the Gin context.
}

// requestIDGeneratorMode is an internal enum to track which generator we use.
type requestIDGeneratorMode int

const (
	generatorModeNone requestIDGeneratorMode = iota
	generatorModeNoContext
	generatorModeWithContext
)

// RequestIDGenerator is a function type that generates a unique ID.
type RequestIDGenerator func() string

// RequestIDContextGenerator is a function type that generates a unique ID based on the Gin context.
type RequestIDContextGenerator func(c *gin.Context) string

// RequestIDOption is a function that configures the requestIDOptions.
type RequestIDOption func(*requestIDOptions)

// WithRequestIDHeader allows setting a custom header name for the request ID.
func WithRequestIDHeader(headerName string) RequestIDOption {
	return func(opts *requestIDOptions) {
		if headerName != "" {
			opts.headerName = headerName
		}
	}
}

// WithRequestIDGenerator allows setting a custom ID generator function.
func WithRequestIDGenerator(gen RequestIDGenerator) RequestIDOption {
	return func(opts *requestIDOptions) {
		if gen != nil {
			opts.generator = gen
			opts.generatorMode = generatorModeNoContext
		}
	}
}

// WithRequestIDContextGenerator allows setting a custom ID generator function that uses the Gin context.
func WithRequestIDContextGenerator(gen RequestIDContextGenerator) RequestIDOption {
	return func(opts *requestIDOptions) {
		if gen != nil {
			opts.contextGenerator = gen
			opts.generatorMode = generatorModeWithContext
		}
	}
}

// RequestID returns a Gin middleware that injects a unique request ID into each HTTP request's context.
//
// The middleware performs the following tasks:
//  1. Extracts the request ID from the incoming request headers using the specified header name (default: "X-Request-ID").
//  2. Validates the request ID to ensure it is not empty and does not exceed 64 characters. If invalid or missing, it generates a new request ID using the provided or default generator function.
//  3. Sets the request ID in the response headers so that the client knows which request ID was assigned.
//  4. Stores the request ID in the request context, making it accessible to downstream middlewares and handlers.
//
// Key Features:
//   - Custom Header Name: Use `WithRequestIDHeader` to specify a custom header name for the request ID.
//   - Custom ID Generator: Use `WithRequestIDGenerator` or `WithRequestIDContextGenerator` to provide a custom generator function for creating request IDs.
//   - Default Generator: By default, the middleware uses the `xid` package to generate compact and globally unique request IDs.
//   - Request Context Integration: The request ID is injected into the context, enabling downstream handlers to retrieve it using `GetRequestIDFromContext`.
//
// Example Usage:
//
//	router.Use(
//		RequestID(
//	    	WithRequestIDHeader("X-Custom-Request-ID"), // Use a custom header name.
//	    	WithRequestIDGenerator(func() string {     // Use a custom generator function.
//	        	return "custom-" + xid.New().String()
//	    	}),
//		),
//	)
func RequestID(opts ...RequestIDOption) gin.HandlerFunc {
	// Set default options.
	options := &requestIDOptions{
		headerName:    DefaultRequestIDHeader,    // Use X-Request-ID as the default header name.
		generatorMode: generatorModeNoContext,    // Use the default generator mode.
		generator:     defaultRequestIDGenerator, // Use xid as the default request ID generator.
	}

	// Apply any user-provided options.
	for _, opt := range opts {
		opt(options)
	}

	return func(c *gin.Context) {
		// Retrieve the request ID from the incoming request headers.
		requestID := c.GetHeader(options.headerName)

		// Limit the length of incoming request IDs to prevent abuse.
		if len(requestID) > 64 {
			requestID = ""
		}

		// If no valid incoming request ID, generate one.
		if requestID == "" {
			switch options.generatorMode {
			case generatorModeWithContext:
				if options.contextGenerator != nil {
					requestID = options.contextGenerator(c)
					break
				}
				fallthrough // Fallback to default if somehow contextGenerator is nil
			case generatorModeNoContext:
				if options.generator != nil {
					requestID = options.generator()
					break
				}
				fallthrough // Fallback if generator is also somehow nil
			default:
				requestID = defaultRequestIDGenerator()
			}
		}

		// Set the request ID in the response headers so the client knows which request ID was used.
		c.Writer.Header().Set(options.headerName, requestID)

		// Store the request ID in the context for downstream handlers.
		ctx := context.WithValue(c.Request.Context(), requestIDContextKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Continue processing the request.
		c.Next()
	}
}

// defaultRequestIDGenerator generates a unique request ID using the xid package.
func defaultRequestIDGenerator() string {
	return xid.New().String()
}
