package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kittipat1413/go-common/framework/trace"
	"go.opentelemetry.io/otel"
)

/*
to override the default service name and add resource attributes, run the following command:
	env OTEL_SERVICE_NAME="my-override-service" \
	OTEL_RESOURCE_ATTRIBUTES="deployment.environment=local,service.namespace=namespace,service.version=1.0" \
	go run framework/trace/example/trace_func/main.go
*/

func main() {
	// Initialize context
	ctx := context.Background()

	// Initialize the tracer provider with Stdout exporter
	tracerProvider, err := trace.InitTracerProvider(ctx, "my-service", nil, trace.ExporterStdout)
	if err != nil {
		fmt.Printf("Error initializing tracer provider: %v\n", err)
		return
	}
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}()

	// Trace the function execution using TraceFunc
	result, err := trace.TraceFunc(ctx, otel.Tracer("my-service-tracer"), myFunc)
	if err != nil {
		fmt.Printf("Error during function execution: %v\n", err)
		return
	}

	fmt.Println("Function result:", result)
}

func myFunc(ctx context.Context) (string, error) {
	// Simulate data fetching or computation
	time.Sleep(100 * time.Millisecond)

	return "Hello, World!", nil
}
