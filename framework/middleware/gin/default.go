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
