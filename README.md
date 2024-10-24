# Common Backend Framework
Welcome to the **Common Backend Framework** repository! This project provides a collection of utilities, libraries, and frameworks designed to standardize and streamline the development of backend API services within our organization.

## Introduction
This repository serves as a central codebase for common functionalities and utilities used across various backend API projects. By consolidating shared code, we aim to:
- Promote consistency and code standardization.
- Reduce duplication of effort.
- Facilitate easier maintenance and updates.
- Accelerate development by providing ready-to-use components.

## Installation
Include the framework in your project by adding it to your go.mod file:
```bash
go get github.com/kittipat1413/go-common
```

## Getting Started
1. Import the Framework
```golang
import (
    "github.com/kittipat1413/go-common/framework/logger"
    "github.com/kittipat1413/go-common/framework/event"
    // ... other imports
)
```
2. Initialize Components
Set up the components you need, such as the logger and event handler.
```golang
func main() {
    // Initialize the logger
    logger := logger.NewLogger()

    // Initialize the event handler
    eventHandler := event.NewEventHandler(
        // Custom configurations if needed
    )

    // ... rest of your application setup
}
```

## Modules and Packages
### [Logger](/framework/logger/)
Provides a structured, context-aware logging interface using logrus. Designed for both development and production environments.
- Features:
  - Configurable log levels.
  - Structured logging with fields.
  - Context propagation for tracing (trace_id, span_id).
  - Flexible output destinations (stdout, files, etc.).
  - No-op logger for testing.

### [Tracer](/framework/trace/)
Implements distributed tracing capabilities to monitor and debug microservices.
- Features:
  - Integrates with tracing systems like OpenTelemetry.
  - Captures spans and context propagation.
  - Minimal performance overhead.

### [Error](/framework/errors/) and Response Handler
Standardizes error handling and HTTP responses.
- Features:
  - Consistent error formatting and wrapping.
  - Unified response structure.
  - Integration with Gin middleware.

### [Event Package](/framework//event/)
Handles event-driven workflows, including message parsing and callback mechanisms.
- Features:
  - Integration with HTTP frameworks (like Gin).
  - Defining and processing event messages with flexible, user-defined payload types.
  - Generic payload support with Go generics.

### Utilities
A collection of helper functions and common utilities.
- Features:
  - Date and time parsing.
  - String manipulation.
  - Configuration loading (e.g., from environment variables, config files).