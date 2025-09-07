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

// GetRequestIDFromContext retrieves the request ID from the provided context.
// Used by downstream handlers and middleware to access the request correlation ID.
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

// requestIDGeneratorMode tracks which ID generator is configured.
// Internal enum to handle different generator types consistently.
type requestIDGeneratorMode int

const (
	generatorModeNone        requestIDGeneratorMode = iota // No generator set
	generatorModeNoContext                                 // Simple generator without context
	generatorModeWithContext                               // Context-aware generator
)

// RequestIDGenerator generates unique request IDs without context dependencies.
// Simple function type for stateless ID generation.
//
// Returns:
//   - string: Unique request identifier
//
// Example:
//
//	func customGenerator() string {
//	    return fmt.Sprintf("req-%d-%s", time.Now().Unix(), uuid.NewString())
//	}
type RequestIDGenerator func() string

// RequestIDContextGenerator generates unique request IDs using Gin context information.
// Enables context-aware ID generation based on request properties like headers, IP, etc.
//
// Parameters:
//   - c: Gin context with request information
//
// Returns:
//   - string: Unique request identifier
//
// Example:
//
//	func contextAwareGenerator(c *gin.Context) string {
//	    userID := c.GetHeader("User-ID")
//	    return fmt.Sprintf("user-%s-req-%s", userID, xid.New().String())
//	}
type RequestIDContextGenerator func(c *gin.Context) string

// RequestIDOption is a function that configures the requestIDOptions.
type RequestIDOption func(*requestIDOptions)

// WithRequestIDHeader sets a custom HTTP header name for request ID extraction and injection.
// Useful when integrating with systems that use different header conventions.
func WithRequestIDHeader(headerName string) RequestIDOption {
	return func(opts *requestIDOptions) {
		if headerName != "" {
			opts.headerName = headerName
		}
	}
}

// WithRequestIDGenerator sets a simple ID generator function without context dependencies.
// Suitable for stateless ID generation using external libraries or custom algorithms.
func WithRequestIDGenerator(gen RequestIDGenerator) RequestIDOption {
	return func(opts *requestIDOptions) {
		if gen != nil {
			opts.generator = gen
			opts.generatorMode = generatorModeNoContext
		}
	}
}

// WithRequestIDContextGenerator sets a context-aware ID generator function.
// Enables ID generation based on request properties like headers, user info, or route data.
func WithRequestIDContextGenerator(gen RequestIDContextGenerator) RequestIDOption {
	return func(opts *requestIDOptions) {
		if gen != nil {
			opts.contextGenerator = gen
			opts.generatorMode = generatorModeWithContext
		}
	}
}

// RequestID returns Gin middleware that manages unique request identifiers for HTTP requests.
// Extracts existing request IDs from headers or generates new ones, then injects them into
// the request context and response headers for distributed tracing and correlation.
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

// defaultRequestIDGenerator creates unique request IDs using the XID library.
// Generates compact, URL-safe, globally unique identifiers suitable for request correlation.
func defaultRequestIDGenerator() string {
	return xid.New().String()
}
