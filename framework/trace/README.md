[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Trace Package
The trace package provides an easy-to-use utility for adding distributed tracing to Go applications using OpenTelemetry. With support for multiple trace exporters, function-level tracing, and integration with web frameworks like Gin, this package simplifies instrumentation and performance monitoring across services.

## Key Features:
- **Tracer Provider Initialization:** Easily initialize and configure OpenTelemetry's Tracer Provider with support for multiple exporters, including:
  `gRPC Exporter`: Export traces to remote tracing backends.
  `Stdout Exporter`: Print traces to the console for local development and debugging.
- **Function-Level Tracing (TraceFunc)**: Automatically trace the execution of any Go function, including capturing function start/end times, execution status, and errors.

## Usage
### 1. Initialize Tracer Provider
To initialize the Tracer Provider, call `InitTracerProvider` with the service name and desired exporter:
```go
ctx := context.Background()
tracerProvider, err := trace.InitTracerProvider(ctx, "my-service", nil, trace.ExporterStdout)
if err != nil {
    log.Fatalf("Failed to initialize tracer provider: %v", err)
}
defer tracerProvider.Shutdown(ctx)
```
> You can also set the `OTEL_SERVICE_NAME` environment variable to override the service name dynamically. Additionally, you can set the `OTEL_RESOURCE_ATTRIBUTES` environment variable to specify additional resource attributes.
### 2. Function-Level Tracing (`TraceFunc`)
Wrap any function in TraceFunc to automatically trace its execution:
```go
result, err := trace.TraceFunc(ctx, otel.Tracer("my-tracer"), func(ctx context.Context) (string, error) {
    // Simulate some processing
    time.Sleep(100 * time.Millisecond)
    return "Hello, World!", nil
})

if err != nil {
    log.Fatalf("Error executing traced function: %v", err)
}
fmt.Println("Result:", result)
```

## Example
You can find a complete working example in the repository under [framework/trace/example](example/).