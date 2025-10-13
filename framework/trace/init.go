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
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

// DefaultTracer returns a tracer instance with this package's instrumentation scope.
// Uses a consistent naming convention for spans created by this common library.
//
// Returns:
//   - trace.Tracer: Tracer with package-specific instrumentation scope
//
// Example:
//
//	tracer := trace.DefaultTracer()
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
func DefaultTracer() trace.Tracer {
	return otel.Tracer("github.com/kittipat1413/go-common/framework/trace")
}

// GetTracer returns a named tracer or the default tracer if name is empty.
// Enables custom instrumentation scopes for different components or services.
//
// Parameters:
//   - name: Custom instrumentation scope name (empty returns default tracer)
//
// Returns:
//   - trace.Tracer: Named tracer instance
//
// Example:
//
//	// Custom tracer for specific component
//	dbTracer := trace.GetTracer("database-operations")
//
//	// Default tracer when name is empty
//	defaultTracer := trace.GetTracer("")
func GetTracer(name string) trace.Tracer {
	if name == "" {
		return DefaultTracer()
	}
	return otel.Tracer(name)
}

// ExporterType defines supported trace exporter backends.
// Determines where trace data is sent for collection and analysis.
type ExporterType string

const (
	ExporterStdout ExporterType = "stdout" // Console output for development and debugging
	ExporterGRPC   ExporterType = "grpc"   // OTLP gRPC for production observability platforms
)

// InitTracerProvider initializes OpenTelemetry tracing with configurable exporters and resource detection.
// Sets up global tracer provider and propagators for distributed tracing across services.
//
// Initialization Process:
//  1. Override serviceName with OTEL_SERVICE_NAME environment variable if present
//  2. Create exporter based on exporterType
//  3. Auto-detect system resources (OS, runtime, process information)
//  4. Merge with default OpenTelemetry resource detection
//  5. Configure tracer provider with batcher and resource attribution
//  6. Set global tracer provider and W3C trace context propagation
//
// Resource Detection:
//   - OS information (name, version, architecture)
//   - Runtime details (Go version, process info)
//   - Environment variables (OTEL_* configuration)
//   - Service name and version identification
//
// Propagation:
//   - W3C Trace Context for cross-service trace correlation
//   - W3C Baggage for additional metadata propagation
//
// Parameters:
//   - ctx: Context for exporter initialization and resource detection
//   - serviceName: Service identifier for traces (overridden by OTEL_SERVICE_NAME)
//   - endpoint: gRPC endpoint for OTLP exporter in "host:port" format (nil for stdout)
//   - exporterType: Exporter backend type (stdout for dev, grpc for production)
//
// Returns:
//   - *sdktrace.TracerProvider: Configured tracer provider for cleanup
//   - error: Configuration or initialization error
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
		if endpoint == nil {
			return nil, fmt.Errorf("endpoint must be provided for gRPC exporter")
		}
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
