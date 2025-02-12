package middleware

import (
	"github.com/gin-gonic/gin"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
	"go.opentelemetry.io/otel/trace"
)

type DefaultMiddlewareConfig struct {
	Logger         common_logger.Logger
	TracerProvider trace.TracerProvider
}

// ConfigureDefaultMiddlewares returns a slice of Gin handlers that constitute
// a standardized "default" middleware chain for common features. These include:
//
//   - Recovery: Recovers from panics and returns a 500 Internal Server Error response.
//   - CircuitBreaker: Monitors request failures and may trip to protect the service.
//   - Trace: Instruments incoming requests with OpenTelemetry spans if a TracerProvider is given.
//   - RequestID: Ensures each request has a unique identifier (injected in the header and context).
//   - RequestLogger: Logs incoming requests with metadata such as method, route, and status code.
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
		Recovery(),
		CircuitBreaker(),
		Trace(WithTracerProvider(config.TracerProvider)),
		RequestID(),
		RequestLogger(WithRequestLogger(config.Logger)),
	}
	return middlewares
}
