# Common Backend Framework
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
![coverage](https://img.shields.io/badge/coverage-90.8%25-brightgreen)
[![Test](https://github.com/kittipat1413/go-common/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/kittipat1413/go-common/actions/workflows/test.yaml)
[![Lint](https://github.com/kittipat1413/go-common/actions/workflows/lint.yaml/badge.svg?branch=main)](https://github.com/kittipat1413/go-common/actions/workflows/lint.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kittipat1413/go-common)](https://goreportcard.com/report/github.com/kittipat1413/go-common)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)
---
Welcome to the **Common Backend Framework** repository! This project provides a collection of utilities, libraries, and frameworks designed to standardize and streamline the development of backend API services within our organization.

- **Gitbook**: [https://kittipat1413.gitbook.io/go-common](https://kittipat1413.gitbook.io/go-common)
- **Go Documentation**: [https://pkg.go.dev/github.com/kittipat1413/go-common](https://pkg.go.dev/github.com/kittipat1413/go-common)
- **Github Repo**: [https://github.com/kittipat1413/go-common](https://github.com/kittipat1413/go-common)

## Introduction
This repository serves as a central codebase for common functionalities and utilities used across various backend API projects. By consolidating shared code, we aim to:
- Promote consistency and code standardization.
- Reduce duplication of effort.
- Facilitate easier maintenance and updates.
- Accelerate development by providing ready-to-use components.

## Installation
Include the framework in your project by adding it to your `go.mod` file:
```bash
go get github.com/kittipat1413/go-common
```

## Getting Started
1. Import the Framework
```go
import (
    "github.com/kittipat1413/go-common/framework/logger"
    "github.com/kittipat1413/go-common/framework/trace"
    // ... other imports
)
```
2. Initialize Components
Set up the components you need, such as the logger and tracer, in your application's entry point.
```go
func main() {
    // Initialize the logger
    logger := logger.NewDefaultLogger()

    // Initialize the tracer
    endpoint := "localhost:4317"
    tracerProvider, err := trace.InitTracerProvider(ctx, "my-service", &endpoint, trace.ExporterGRPC)
    if err != nil {
      // Handle error
      return
    }
    defer func() {
      if err := tracerProvider.Shutdown(ctx); err != nil {
        // Handle error
      }
    }()

    // ... rest of your application setup
}
```

## Example Project

Looking for a real-world implementation?
- Check out the [🎟️ Ticket Reservation System](https://github.com/kittipat1413/ticket-reservation), a clean-architecture-based backend service that demonstrates how to integrate and use the `go-common` framework across all layers — config, logging, tracing, error handling, HTTP handlers, use cases, and repositories.

## Modules and Packages

#### [Cache](/framework/cache/) 
Provides a unified caching interface with multiple backend implementations for high-performance data access.
- **LocalCache** - In-memory caching with TTL support and automatic cleanup
- Features:
  - Generic type-safe interface
  - Thread-safe operations
  - Configurable TTL and cleanup intervals
  - Context-aware operations

#### [Config](/framework/config/) 
Manages application configuration with support for multiple sources and context propagation.
- Features:
  - Environment variable loading
  - YAML/JSON configuration files
  - Type-safe configuration access
  - Context-based configuration retrieval

#### [Errors](/framework/errors/)
Standardizes error handling and response formatting with categorized error types.
- Features:
  - Predefined error categories (authentication, bad request, internal server, etc.)
  - Consistent error formatting and wrapping
  - Error codes and messages
  - HTTP status code mapping
  - Error response generation for APIs

#### [Lock Manager](/framework/lockmanager/)
Unified distributed locking interface for coordinating access to shared resources.
- **LocalLock** - In-memory locks for single-instance applications
- **Redsync** - Redis-based distributed locks for multi-instance deployments
- Features:
  - Consistent locking API across implementations
  - Configurable lock TTL and retry policies
  - Context support with cancellation
  - Custom token generation

#### [Logger](/framework/logger/)
Structured, context-aware logging built on logrus for production-ready applications.
- Features:
  - Configurable log levels (DEBUG, INFO, WARN, ERROR, FATAL)
  - Structured JSON logging
  - Context propagation for tracing (`trace_id`, `span_id`)
  - Flexible output destinations
  - Custom formatters support
  - No-op logger for testing

#### [Middleware](/framework/middleware/)
Collection of HTTP middleware for common web application concerns.
- **Gin Middleware** - Ready-to-use middleware for Gin framework
- Features:
  - Request logging
  - Error handling
  - Recovery from panics
  - Request ID generation

#### [Retry](/framework/retry/)
Flexible retry mechanism with multiple backoff strategies for resilient operations.
- Features:
  - Exponential backoff
  - Linear backoff
  - Configurable retry policies
  - Context-aware execution
  - Custom retry conditions

#### [Server Utils](/framework/serverutils/)
Server lifecycle utilities for graceful startup and shutdown.
- Features:
  - Graceful shutdown handling
  - Signal handling (SIGTERM, SIGINT)
  - Configurable shutdown timeout
  - Multiple server support

#### [SFTP](/framework/sftp/)
Production-ready SFTP client with connection pooling and advanced file transfer capabilities.
- Features:
  - Connection pooling with automatic management
  - Password and private key authentication
  - Upload/Download with progress tracking
  - Smart overwrite policies
  - Automatic retry with exponential backoff
  - Directory operations (create, list, remove)

#### [Trace](/framework/trace/)
Distributed tracing implementation using OpenTelemetry for observability.
- Features:
  - OpenTelemetry integration
  - Span creation and context propagation
  - Multiple exporter support (GRPC, HTTP)
  - Minimal performance overhead
  - Trace function utilities

#### [Validator](/framework/validator/)
Request validation with custom validation rules and integration with validator/v10.
- Features:
  - Struct validation
  - Custom validation rules
  - Field-level validation
  - Integration with web frameworks
  - Detailed error messages

#### [JWT](/util/jwt/)
JWT token creation and validation with support for multiple signing algorithms.
- Features:
  - HS256 (HMAC) signing
  - RS256 (RSA) signing
  - Custom claims support
  - Token validation and parsing
  - Context-aware operations

## Package Index

| Package | Description |
|---------|-------------|
| [cache](/framework/cache/) | Unified caching interface with local implementation |
| [config](/framework/config/) | Application configuration management |
| [errors](/framework/errors/) | Standardized error handling |
| [lockmanager](/framework/lockmanager/) | Distributed locking interface |
| [logger](/framework/logger/) | Structured logging |
| [middleware](/framework/middleware/) | HTTP middleware collection |
| [retry](/framework/retry/) | Flexible retry mechanism |
| [serverutils](/framework/serverutils/) | Server lifecycle utilities |
| [sftp](/framework/sftp/) | SFTP client with connection pooling |
| [trace](/framework/trace/) | Distributed tracing |
| [validator](/framework/validator/) | Request validation |
| [jwt](/util/jwt/) | JWT token management |

## Contributing
Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License
This project is licensed under the [MIT License](LICENSE).