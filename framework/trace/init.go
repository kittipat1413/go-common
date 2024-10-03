package trace

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

func DefaultTracer() trace.Tracer {
	return otel.Tracer("go-common default")
}

type ExporterType string

const (
	ExporterStdout ExporterType = "stdout"
	ExporterGRPC   ExporterType = "grpc"
)

// InitTracerProvider initializes an OpenTelemetry TracerProvider with the specified service name, exporter type, and gRPC endpoint (if needed).
// It supports both stdout and gRPC exporters and sets global tracing and propagation configurations.
//
// Params:
// - ctx: Context for initialization, used for trace exporter creation and resource detection.
// - serviceName: The name of the service being traced. Can be overridden by the environment variable `OTEL_SERVICE_NAME`.
// - endpoint: The gRPC endpoint for trace exporters (only applicable for the gRPC exporter).
// - exporterType: The type of exporter to use (either "stdout" or "grpc").
//
// Returns:
// - *sdktrace.TracerProvider: A new tracer provider to manage tracing.
// - error: An error if the initialization fails.
func InitTracerProvider(ctx context.Context, serviceName string, endpoint *string, exporterType ExporterType) (*sdktrace.TracerProvider, error) {
	if envServiceName := os.Getenv("OTEL_SERVICE_NAME"); envServiceName != "" {
		serviceName = envServiceName
	}

	var (
		exporter sdktrace.SpanExporter
		err      error
	)
	switch exporterType {
	case ExporterGRPC:
		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint(*endpoint))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC trace exporter: %w", err)
		}
	case ExporterStdout:
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to initialize stdout trace exporter: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s", exporterType)
	}

	// Create a resource that describes the service for the trace
	systemResource, err := resource.New(ctx,
		resource.WithOS(),                        // Discover and provide OS information.
		resource.WithProcessRuntimeName(),        // Discover and provide process information.
		resource.WithProcessRuntimeVersion(),     // Discover and provide process information.
		resource.WithProcessRuntimeDescription(), // Discover and provide process information.
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName), // Set the service name as an attribute
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create system resource: %w", err)
	}

	// https://opentelemetry.io/docs/languages/go/resources/
	// Merge system resource with any resources automatically detected (e.g., from the environment)
	resource, err := resource.Merge(
		resource.Default(), // Use the default resource detection (e.g., environment variables)
		systemResource,     // Add the system resource to the default resource
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create the TracerProvider with the exporter and resource
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	// Register the TracerProvider as the global provider
	otel.SetTracerProvider(tracerProvider)

	// Register the W3C trace context and baggage propagators so data is propagated across services/processes
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return tracerProvider, nil
}
