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
		middleware.Trace(middleware.WithTracerProvider(tracerProvider)),
		middleware.Recovery(middleware.WithRecoveryLogger(logger)),
	}
	router.Use(middlewares...)

	// Panic handler
	router.GET("/panic", func(c *gin.Context) {
		panic("Panic!")
	})

	// Trace handler
	router.GET("/trace", traceHandler)

	// Run the Gin server
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}

// traceHandler is a handler that creates a span and responds with a message.
func traceHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Create a span
	_, span := trace.DefaultTracer().Start(ctx, "traceHandler", oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	defer span.End()

	// Respond with the result
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
