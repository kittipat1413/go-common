package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kittipat1413/go-common/framework/logger"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

/*
To run this example with Gin, execute the following commands:
1. go run main.go
2. curl -X GET http://localhost:8080/log
*/

// Initialize OpenTelemetry Tracer Provider with Noop Exporter
func initTracer() func() {
	// Create a Noop tracer provider
	tp := sdktrace.NewTracerProvider()

	// Set the Noop tracer provider as the global tracer provider
	otel.SetTracerProvider(tp)

	// Return a no-op function for shutdown
	return func() {
		// Noop shutdown function
	}
}

// Key to store the request ID in the context
type requestIdKey struct{}

// Middleware to attach logger and tracing to the Gin context
func loggerAndTracingMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate a unique request ID
		requestID := uuid.New().String()

		// Start a new span for the request
		tracer := otel.Tracer("server-tracer")
		ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
		// Store the span in the context for later use
		c.Request = c.Request.WithContext(ctx)

		// Add request ID to context
		ctx = context.WithValue(ctx, requestIdKey{}, requestID)

		// Attach logger to the context with request ID as a persistent field
		requestLogger := log.WithFields(logger.Fields{"request_id": requestID})
		ctx = logger.NewContext(ctx, requestLogger)

		// Replace the request with the new context
		c.Request = c.Request.WithContext(ctx)

		// Proceed to the next handler
		c.Next()

		// After request handling, end the span
		span.End()
	}
}

// Example function to log messages at different levels
func logMessages(log logger.Logger) {
	ctx := context.Background()
	log.Debug(ctx, "This is a debug message", logger.Fields{"example_field": "debug_value"})
	log.Info(ctx, "This is an info message", logger.Fields{"example_field": "info_value"})
	log.Warn(ctx, "This is a warning message", logger.Fields{"example_field": "warn_value"})
	log.Error(ctx, "This is an error message", fmt.Errorf("example error"), logger.Fields{"example_field": "error_value"})
}

// Example Gin HTTP handler using logger
func handlerWithLogger(c *gin.Context) {
	// Retrieve the logger from the Gin context
	log := logger.FromContext(c.Request.Context())

	// Retrieve the request ID from the context
	requestID := c.Request.Context().Value(requestIdKey{}).(string)

	// Start timer
	startTime := time.Now()

	// Proceed with handling the request
	c.JSON(http.StatusOK, gin.H{"message": "Logged HTTP request", "request_id": requestID})

	// Calculate response time
	duration := time.Since(startTime)

	// Log a message for the incoming request
	log.Info(c.Request.Context(), "Handled HTTP request", logger.Fields{
		"request": logger.Fields{
			"method": c.Request.Method,
			"url":    c.Request.URL.Path,
		},
		"status":        c.Writer.Status(),
		"response_time": duration.Seconds(),
	})
}

func main() {
	// Initialize context
	ctx := context.Background()
	// Initialize OpenTelemetry tracer
	shutdownTracer := initTracer()
	defer shutdownTracer()
	// Initialize logger with custom configuration
	logConfig := logger.Config{
		Level: logger.DEBUG,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
			FieldKeyFormatter: func(key string) string {
				switch key {
				case logger.DefaultEnvironmentKey:
					return "env"
				case logger.DefaultServiceNameKey:
					return "service"
				case logger.DefaultSJsonFmtSeverityKey:
					return "level"
				case logger.DefaultSJsonFmtMessageKey:
					return "msg"
				case logger.DefaultSJsonFmtErrorKey:
					return "err"
				default:
					return key
				}
			},
		},
		Environment: "development",
		ServiceName: "logger-example",
	}
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	// Example of logging messages at different levels
	logMessages(log)

	// Initialize Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Use the logger and tracing middleware
	r.Use(loggerAndTracingMiddleware(log))

	// Example route
	r.GET("/log", handlerWithLogger)

	// Start the Gin HTTP server in a goroutine
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Channel to listen for errors
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Info(ctx, "Starting HTTP server on :8080", nil)
		serverErrors <- server.ListenAndServe()
	}()

	// Listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Error(ctx, "Server error", err, nil)
	case sig := <-quit:
		log.Info(ctx, "Shutting down server", logger.Fields{"signal": sig.String()})
	}

	// Graceful shutdown
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Error(ctx, "Server shutdown error", err, nil)
	} else {
		log.Info(ctx, "Server exited properly", nil)
	}
}
