package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
	middleware "github.com/kittipat1413/go-common/framework/middleware/gin"
	"github.com/kittipat1413/go-common/framework/trace"
	"github.com/sony/gobreaker"
	oteltrace "go.opentelemetry.io/otel/trace"
)

/*
Run Server:
to override the default service name and add resource attributes, run the following command:
	env OTEL_RESOURCE_ATTRIBUTES="deployment.environment=local,service.version=1.0" \
	go run framework/middleware/example/gin/main.go

Curl: To test the middleware, run the following curl commands:
	curl -X GET http://localhost:8080/health
	curl -X GET http://localhost:8080/unstable/down
	curl -X GET http://localhost:8080/panic
	curl -X GET http://localhost:8080/trace
	curl -v -X GET http://localhost:8080/request-id
	curl -X GET http://localhost:8080/request-logger
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

	logger, _ := common_logger.NewLogger(common_logger.Config{
		Level: common_logger.DEBUG,
		Formatter: &common_logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
		},
		Environment: "local",
		ServiceName: "gin-middleware-testing",
	})

	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)
	// Create a Gin router
	router := gin.New()

	// Add middlewares to the Gin router.
	// The order of middlewares is important to ensure they function correctly:
	var middlewares = []gin.HandlerFunc{
		middleware.Recovery(),
		middleware.Trace(middleware.WithTracerProvider(tracerProvider)),
		middleware.RequestID(middleware.WithRequestIDHeader("X-Request-ID")),
		middleware.RequestLogger(
			middleware.WithRequestLogger(logger),
			middleware.WithRequestLoggerFilter(func(req *http.Request) bool {
				return req.URL.Path != "/health"
			})),
		middleware.CircuitBreaker(middleware.WithCircuitBreakerSettings(gobreaker.Settings{
			Name: "TestCircuitBreaker",
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				// Default: trip the circuit after 3 consecutive failures.
				return counts.ConsecutiveFailures >= 3
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				logger.Warn(context.Background(), fmt.Sprintf("Circuit breaker %s changed state from %s to %s", name, from, to), nil)
			},
			Timeout: 10 * time.Second,
		})),
		middleware.Prometheus("gin-middleware-testing"),
	}
	router.Use(middlewares...)

	// Metrics handler
	router.GET("/metrics", middleware.MetricsHandler())

	// Health check handler
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Unstable handler
	router.GET("/unstable/:status", func(c *gin.Context) {
		status := c.Param("status")
		if status == "down" {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "DOWN"})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "UP"})
		}
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

	// Request logger handler
	router.GET("/request-logger", func(c *gin.Context) {
		logger := common_logger.FromContext(c.Request.Context())
		logger.Info(c.Request.Context(), "Request logger handler", nil)
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})

	// Panic handler
	router.GET("/panic", func(c *gin.Context) {
		panic("Panic!")
	})

	// Run the Gin server
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
