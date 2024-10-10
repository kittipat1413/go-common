package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/kittipat1413/go-common/framework/logger/formatter"
)

/*
To run this example with Gin, execute the following commands:
curl -X GET http://localhost:8080/log
*/

// Example function to log messages at different levels
func logMessages(log logger.Logger) {
	ctx := context.Background()

	// Log a debug message
	log.Debug(ctx, "This is a debug message", logger.Fields{"example_field": "debug_value"})

	// Log an info message
	log.Info(ctx, "This is an info message", logger.Fields{"example_field": "info_value"})

	// Log a warning message
	log.Warn(ctx, "This is a warning message", logger.Fields{"example_field": "warn_value"})

	// Log an error message
	log.Error(ctx, "This is an error message", fmt.Errorf("example error"), logger.Fields{"example_field": "error_value"})
}

// Example Gin HTTP handler using logger
func handlerWithLogger(c *gin.Context) {
	// Retrieve the logger from the Gin context
	log := logger.FromContext(c.Request.Context())

	// Log a message for the incoming request
	log.Info(c.Request.Context(), "Received HTTP request", logger.Fields{
		"request": logger.Fields{
			"method": c.Request.Method,
			"url":    c.Request.URL.Path,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logged HTTP request"})
}

func main() {
	// Initialize logger with JSON formatter and info level
	logConfig := logger.Config{
		Level: logger.DEBUG,
		Formatter: &formatter.ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
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
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Middleware to attach logger to the Gin context
	r.Use(func(c *gin.Context) {
		// Attach logger to the request context
		ctx := logger.NewContext(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	// Example route
	r.GET("/log", handlerWithLogger)

	// Start the Gin HTTP server
	log.Info(context.Background(), "Starting HTTP server on :8080", nil)
	if err := r.Run(":8080"); err != nil {
		log.Fatal(context.Background(), "Failed to start server", err, nil)
	}
}
