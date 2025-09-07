package middleware

import (
	"github.com/gin-gonic/gin"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
	"go.opentelemetry.io/otel/trace"
)

// DefaultMiddlewareConfig holds configuration options for the default middleware chain.
type DefaultMiddlewareConfig struct {
	// Logger provides custom logging for recovery and request logging middleware.
	// If nil, middleware will use default logger implementations.
	Logger common_logger.Logger

	// TracerProvider enables distributed tracing for request instrumentation.
	// If nil, the global OpenTelemetry trace provider is used.
	TracerProvider trace.TracerProvider
}

// ConfigureDefaultMiddlewares returns a standardized middleware chain for common HTTP concerns.
// The chain includes recovery, circuit breaker, tracing, request ID, logging, and metrics
// in an order optimized for proper request handling and observability.
//
// Middleware Chain:
//   - Recovery: Recovers from panics and returns a 500 Internal Server Error response.
//   - CircuitBreaker: Monitors request failures and may trip to protect the service.
//   - Trace: Instruments incoming requests with OpenTelemetry spans if a TracerProvider is given.
//   - RequestID: Ensures each request has a unique identifier (injected in the header and context).
//   - RequestLogger: Logs incoming requests with metadata such as method, route, and status code.
//   - Prometheus: Exposes metrics for monitoring and alerting via Prometheus.
//
// The behavior of certain middlewares (e.g., tracing, logging) is influenced by
// the fields in DefaultMiddlewareConfig. If the TracerProvider is nil, the global
// trace provider is used. If the Logger is nil, a default logger is used.
//
// Example usage:
//
//	cfg := DefaultMiddlewareConfig{
//	  Logger:         myLogger,
//	  TracerProvider: myTracerProvider,
//	}
//
//	router := gin.New()
//	router.Use(ConfigureDefaultMiddlewares(cfg)...)
func ConfigureDefaultMiddlewares(config DefaultMiddlewareConfig) []gin.HandlerFunc {
	middlewares := []gin.HandlerFunc{
		Recovery(WithRecoveryLogger(config.Logger)),
		Trace(WithTracerProvider(config.TracerProvider)),
		RequestID(),
		RequestLogger(WithRequestLogger(config.Logger)),
		CircuitBreaker(),
		Prometheus(""),
	}
	return middlewares
}
