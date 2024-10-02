package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kittipat1413/go-common/framework/trace"
	"go.opentelemetry.io/otel"
)

/*
Curl command to test the endpoint:
curl -X POST http://localhost:8080/process -H "Content-Type: application/json" -d '{"message": "Hello, Gin!"}'

Run Server:
to override the default service name and add resource attributes, run the following command:
	env OTEL_SERVICE_NAME="override-gin-tracing-service" \
	OTEL_RESOURCE_ATTRIBUTES="deployment.environment=local,service.version=1.0" \
	go run framework/trace/example/gin/main.go
*/

// Define a request body structure
type RequestData struct {
	Message string `json:"message"`
}

func main() {
	// Initialize context
	ctx := context.Background()

	tracerProvider, err := trace.InitTracerProvider(ctx, "gin-tracing-service", nil, trace.ExporterStdout)
	if err != nil {
		fmt.Printf("Error initializing tracer provider: %v\n", err)
		return
	}
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}()

	// Create a new Gin engine
	r := gin.Default()

	// Define an endpoint with tracing
	r.POST("/process", myGinHandler)

	// Run the Gin server
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}

// Handler function to process the request and call another traced function
func myGinHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var reqData RequestData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Trace another function that processes the request data
	result, err := trace.TraceFunc(ctx, otel.Tracer("gin-service-tracer"), processData(reqData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the result
	c.JSON(http.StatusOK, gin.H{"message": result})
}

// Simulate a function that processes the request data with tracing
func processData(data RequestData) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return fmt.Sprintf("Processed message: %s", data.Message), nil
	}
}
