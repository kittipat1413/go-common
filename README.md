# Common Backend Framework
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
![coverage](https://img.shields.io/badge/coverage-91.3%25-brightgreen)
[![Test](https://github.com/kittipat1413/go-common/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/kittipat1413/go-common/actions/workflows/test.yaml)
[![Lint](https://github.com/kittipat1413/go-common/actions/workflows/lint.yaml/badge.svg?branch=main)](https://github.com/kittipat1413/go-common/actions/workflows/lint.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kittipat1413/go-common)](https://goreportcard.com/report/github.com/kittipat1413/go-common)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)
---
Welcome to the **Common Backend Framework** repository! This project provides a collection of utilities, libraries, and frameworks designed to standardize and streamline the development of backend API services within our organization.

- **Gitbook**: [https://kittipat1413.gitbook.io/go-common](https://kittipat1413.gitbook.io/go-common)
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
### [Logger](/framework/logger/)
Provides a structured, context-aware logging interface using logrus. Designed for both development and production environments.
- Features:
  - Configurable log levels.
  - Structured logging with fields.
  - Context propagation for tracing (`trace_id`, `span_id`).
  - Flexible output destinations (`stdout`, `files`, etc.).
  - No-op logger for testing.

### [Tracer](/framework/trace/)
Implements distributed tracing capabilities to monitor and debug microservices.
- Features:
  - Integrates with tracing systems like OpenTelemetry.
  - Captures spans and context propagation.
  - Minimal performance overhead.

### [Error](/framework/errors/)
Standardizes error handling and response formatting.
- Features:
  - Consistent error formatting and wrapping.
  - Error codes and messages.
  - HTTP status code mapping.
  - Error response generation.

### [Event Package](/framework/event/)
Handles event-driven workflows, including message parsing and callback mechanisms.
- Features:
  - Integration with HTTP frameworks (like Gin).
  - Defining and processing event messages with flexible, user-defined payload types.
  - Generic payload support with Go generics.

### [Utilities](/util/)
A collection of helper functions and common utilities.
- Features:
  - Date and time parsing.
  - String manipulation.
  - Configuration loading (e.g., from environment variables, config files).
  - etc.

## License
This project is licensed under the [MIT License](LICENSE).