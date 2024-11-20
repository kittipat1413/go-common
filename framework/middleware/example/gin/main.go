package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kittipat1413/go-common/framework/logger"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/kittipat1413/go-common/framework/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

/*
Run Server:
to override the default service name and add resource attributes, run the following command:
	env OTEL_RESOURCE_ATTRIBUTES="deployment.environment=local,service.version=1.0" \
	go run framework/middleware/example/gin/main.go

Curl: To test the middleware, run the following curl commands:
	curl -X GET http://localhost:8080/panic
	curl -X GET http://localhost:8080/trace
	curl -v -X GET http://localhost:8080/request-id
*/

func main() {
	ctx := context.Background()
	tracerProvider, err := trace.InitTracerProvider(ctx, "gin-middleware-testing", nil, trace.ExporterStdout)
	if err != nil {
		fmt.Printf("Error initializing tracer provider: %v\n", err)
		return
	}
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}()

	logger, _ := logger.NewLogger(logger.Config{
		Level: logger.DEBUG,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
		},
	})

	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)
	// Create a Gin router
	router := gin.New()

	// Add middlewares
	var middlewares = []gin.HandlerFunc{
		middleware.RequestID(middleware.WithRequestIDHeader("X-Request-ID")),
		middleware.Trace(middleware.WithTracerProvider(tracerProvider)),
		middleware.Recovery(middleware.WithRecoveryLogger(logger)),
	}
	router.Use(middlewares...)

	// Panic handler
	router.GET("/panic", func(c *gin.Context) {
		panic("Panic!")
	})

	// Trace handler
	router.GET("/trace", func(c *gin.Context) {
		// Create a span
		_, span := trace.DefaultTracer().Start(c.Request.Context(), "traceHandler", oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
		defer span.End()
		// Respond with the result
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Request ID handler
	router.GET("/request-id", func(c *gin.Context) {
		requestID, exists := middleware.GetRequestIDFromContext(c.Request.Context())
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No Request ID"})
			return
		}
		c.String(http.StatusOK, requestID)
	})

	// Run the Gin server
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
