[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Middleware Package
The middleware package provides reusable middleware components for Go web applications. These middlewares are designed to handle common application requirements, including request tracking, error recovery, logging, tracing, and failure handling.

## Features
- **RequestID Middleware**: Generates and propagates a unique request ID for each HTTP request.
    - Adds the request ID to the context and response headers.
    - Supports custom header names and ID generators.
- **Recovery Middleware**: Recovers from panics and ensures the application continues running.
    - Logs the panic information (including HTTP method and route) using the provided logger or retrieves one from the context.
    - Calls a custom error handler to generate a response, or defaults to a 500 Internal Server Error response if no custom handler is provided.
- **RequestLogger Middleware**: Logs incoming HTTP requests and their corresponding responses.
    - Logs request details, such as method, route, query parameters, client IP, and user agent.
    - Allows filtering of requests to determine whether they should be logged.
    - Injects an augmented logger with request-specific fields into the request context for downstream use.
- **Trace Middleware**: Enables distributed tracing for HTTP requests using OpenTelemetry.
    - Supports custom tracer providers and span name formatters.
    - Allows filtering of routes for selective tracing.
    - Automatically injects the active span into the request context, enabling downstream handlers to log and report spans.
- **CircuitBreaker Middleware**: Protects routes from excessive failures by introducing a circuit breaker mechanism.
	- Monitors request failures and trips the circuit breaker based on configurable thresholds.
	- Supports custom error handlers and route-specific filters.
- **Prometheus Middleware**: Exposes HTTP metrics to Prometheus.
    - Tracks request counts, durations, and sizes.
    - Configurable namespace for metrics.
    - Provides a `/metrics` endpoint for Prometheus scraping.

## Installation
- **Gin**
    ```bash
    go get github.com/kittipat1413/go-common/framework/middleware/gin
    ```

## Documentation
- **Gin** <br/>

    [![Go Reference](https://pkg.go.dev/badge/github.com/kittipat1413/go-common/framework/middleware/gin.svg)](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/middleware/gin)

    For detailed API documentation, examples, and usage patterns, visit the [Go Package Documentation](https://pkg.go.dev/github.com/kittipat1413/go-common/framework/middleware/gin).

## Examples
- You can find a complete working example in the repository under [framework/middleware/example](example/).