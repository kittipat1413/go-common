package logger

import (
	"context"
	"net/http"
)

// contextKey is an unexported type used as a key for storing logger in context.
type contextKey struct{}

// loggerKey is the singleton key used for logger storage in context.
var loggerKey = &contextKey{}

// FromContext retrieves a Logger from the context.
// Returns a default logger if no logger is found in the context,
// ensuring that logging is always available even without explicit setup.
//
// Parameters:
//   - ctx: Context that may contain a logger instance
//
// Returns:
//   - Logger: Logger from context or default logger if not found
//
// Example:
//
//	logger := logger.FromContext(ctx)
//	logger.Info(ctx, "Processing request", nil)
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return NewDefaultLogger()
}

// FromRequest retrieves a Logger from the HTTP request's context.
// Convenience function for HTTP handlers that need access to the logger
// without manually extracting it from the request context.
//
// Parameters:
//   - r: HTTP request containing context with potential logger
//
// Returns:
//   - Logger: Logger from request context or default logger if not found
//
// Example:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    logger := logger.FromRequest(r)
//	    logger.Info(r.Context(), "Handling request", map[string]interface{}{
//	        "method": r.Method,
//	        "path":   r.URL.Path,
//	    })
//	}
func FromRequest(r *http.Request) Logger {
	return FromContext(r.Context())
}

// NewContext returns a new context that carries the provided logger.
// The logger can later be retrieved using FromContext() for consistent
// logging throughout the request lifecycle.
//
// Parameters:
//   - ctx: Parent context
//   - logger: Logger instance to store in the new context
//
// Returns:
//   - context.Context: New context containing the logger
//
// Example:
//
//	logger := logger.NewLoggerWithConfig(config)
//	ctx = logger.NewContext(ctx, logger)
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// NewRequest returns a new HTTP request that carries the provided logger in its context.
// Useful for middleware that wants to inject a configured logger into
// incoming requests for downstream handlers to use.
//
// Parameters:
//   - r: Original HTTP request
//   - logger: Logger instance to attach to the request
//
// Returns:
//   - *http.Request: New request with logger attached to its context
//
// Example:
//
//	// In middleware
//	func LoggerMiddleware(logger Logger) func(http.Handler) http.Handler {
//	    return func(next http.Handler) http.Handler {
//	        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	            r = logger.NewRequest(r, logger)
//	            next.ServeHTTP(w, r)
//	        })
//	    }
//	}
func NewRequest(r *http.Request, logger Logger) *http.Request {
	return r.WithContext(NewContext(r.Context(), logger))
}
