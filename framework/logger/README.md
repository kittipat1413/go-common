[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/kittipat1413/go-common/issues)
[![Release](https://img.shields.io/github/release/kittipat1413/go-common.svg?style=flat)](https://github.com/kittipat1413/go-common/releases/latest)

# Logger Package
The logger package provides a structured, context-aware logging solution for Go applications. It is built on top of the [logrus](https://github.com/sirupsen/logrus) library and is designed to facilitate easy integration with your projects, offering features like:
- JSON-formatted logs suitable for production environments.
- Support for multiple log levels (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`).
- Context propagation to include tracing information (e.g., `trace_id`, `span_id`).
- Customizable log formatters and output destinations.
- Integration with web frameworks like Gin.

## Installation
```bash
go get github.com/kittipat1413/go-common/framework/logger
```

## Features
- **Structured Logging**: Outputs logs in JSON format, making them easy to parse and analyze.
- **Context-Aware**: Supports logging with `context.Context`, allowing you to include tracing information automatically.
- **Customizable Formatter**: Use the default `StructuredJSONFormatter` or provide your own formatter to customize the log output.
- **Environment and Service Name**: Optionally include environment and service name in your logs for better traceability.
- **No-Op Logger**: Provides a no-operation logger for testing purposes, which discards all log messages.

## Usage

### Creating a Logger
You can create a logger using the `NewLogger` function, providing a `Config` struct to customize its behavior:
```go
package main

import (
    "context"
    "time"
    "github.com/kittipat1413/go-common/framework/logger"
)

func main() {
    logConfig := logger.Config{
        Level: logger.INFO,
        Formatter: &logger.StructuredJSONFormatter{
            TimestampFormat: time.RFC3339,
            PrettyPrint:     false,
        },
        Environment: "production",
        ServiceName: "my-service",
    }
    
    log, err := logger.NewLogger(logConfig)
    if err != nil {
        panic(err)
    }

    log.Info(context.Background(), "Logger initialized successfully", nil)
}
```

Alternatively, you can use the default logger:
```go
log := logger.NewDefaultLogger()
```
- The `NewDefaultLogger` returns a logger instance with the default or user-defined configuration.
- If `SetDefaultLoggerConfig` has been called, it uses the user-defined configuration; otherwise, it uses the package's default configuration.

### Updating the Default Logger Configuration
You can update the default logger configuration using SetDefaultLoggerConfig:
```go
err := logger.SetDefaultLoggerConfig(logConfig)
if err != nil {
    // Handle error
    panic(err)
}
```

## Configuration
The Config struct allows you to customize the logger:
```go
type Config struct {
	// Level determines the minimum log level that will be processed by the logger.
	// Logs with a level lower than this will be ignored.
	Level LogLevel
	// Formatter is an optional field for specifying a custom logrus formatter.
	// If not provided, the logger will use the StructuredJSONFormatter by default.
	Formatter logrus.Formatter
	// Environment is an optional field for specifying the running environment (e.g., "production", "staging").
	// This field is used for adding environment-specific fields to logs.
	Environment string
	// ServiceName is an optional field for specifying the name of the service.
	// This field is used for adding service-specific fields to logs.
	ServiceName string
	// Output is an optional field for specifying the output destination for logs (e.g., os.Stdout, file).
	// If not provided, logs will be written to stdout by default.
	Output io.Writer
}
```

## Logging Messages
The logger provides methods for different log levels:
```go
type Logger interface {
    WithFields(fields Fields) Logger
	Debug(ctx context.Context, msg string, fields Fields)
	Info(ctx context.Context, msg string, fields Fields)
	Warn(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, err error, fields Fields)
	Fatal(ctx context.Context, msg string, err error, fields Fields)
}
```
Example:
```go
ctx := context.Background()
fields := logger.Fields{"user_id": 12345}

log.Info(ctx, "User logged in", fields)
```
### Including Errors
For error and fatal logs, you can include an error object:
```go
err := errors.New("something went wrong")
log.Error(ctx, "Failed to process request", err, fields)
```
### Adding Persistent Fields
You can add persistent fields to the logger using WithFields, which returns a new logger instance:
```go
logWithFields := log.WithFields(logger.Fields{
    "component": "authentication",
})
logWithFields.Info(ctx, "Authentication successful", nil)

```
You can find a complete working example in the repository under [framework/logger/example](example/).

---

## StructuredJSONFormatter
The `StructuredJSONFormatter` is a custom `logrus.Formatter` designed to include contextual information in logs. It outputs logs in JSON format with a standardized structure, making it suitable for log aggregation and analysis tools.
### Features
- **Timestamp**: Includes a timestamp formatted according to `TimestampFormat`.
- **Severity**: The log level (`debug`, `info`, `warning`, `error`, `fatal`).
- **Message**: The log message.
- **Error Handling**: Automatically includes error messages if an `error` is provided.
- **Tracing Information**: Extracts `trace_id` and `span_id` from the context if available (e.g., when using OpenTelemetry).
- **Caller Information**: Adds information about the function, file, and line number where the log was generated.
- **Stack Trace**: Includes a stack trace for logs at the `error` level or higher.
- **Custom Fields**: Supports additional fields provided via `logger.Fields`.
- **Field Key Customization**: Allows custom formatting of field keys via `FieldKeyFormatter`.

### Configuration
You can customize the `StructuredJSONFormatter` when initializing the logger:
```go
import (
    "github.com/kittipat1413/go-common/framework/logger"
    "time"
)

formatter := &logger.StructuredJSONFormatter{
    TimestampFormat: time.RFC3339, // Customize timestamp format
    PrettyPrint:     true,         // Indent JSON output
    FieldKeyFormatter: func(key string) string {
        // Customize field keys
        switch key {
        case logger.DefaultEnvironmentKey:
            return "env"
        case logger.DefaultServiceNameKey:
            return "service"
        case logger.DefaultSJsonFmtSeverityKey:
            return "level"
        case logger.DefaultSJsonFmtMessageKey:
            return "msg"
        default:
            return key
        }
    },
}

logConfig := logger.Config{
    Level:     logger.INFO,
    Formatter: formatter,
}

```

Example Log Entry (default `FieldKeyFormatter`)
```json
{
  "caller": {
    "file": "/go-common/framework/logger/example/gin_with_logger/main.go:99",
    "function": "main.handlerWithLogger"
  },
  "environment": "development",
  "message": "Handled HTTP request",
  "request": {
    "method": "GET",
    "url": "/log"
  },
  "request_id": "afa241ba-cb59-4053-a1f5-d82e6193790c",
  "response_time": 0.000026542,
  "service_name": "logger-example",
  "severity": "info",
  "span_id": "94e92f0e1b8532e6",
  "status": 200,
  "timestamp": "2024-10-20T02:01:57+07:00",
  "trace_id": "f891fd44c417fc7efa297e6a18ddf0cf"
}
```
```json
{
  "caller": {
    "file": "/go-common/framework/logger/example/gin_with_logger/main.go:78",
    "function": "main.logMessages"
  },
  "environment": "development",
  "error": "example error",
  "example_field": "error_value",
  "message": "This is an error message",
  "service_name": "logger-example",
  "severity": "error",
  "stack_trace": "goroutine 1 [running]:\ngithub.com/kittipat1413/go-common/framework/logger.getStackTrace()\n\t/Users/kittipat/go-github-repo/go-common/framework/logger/structured_json_formatter.go:181 .....\n",
  "timestamp": "2024-10-20T02:01:55+07:00"
}
```

### Tracing Integration
The `StructuredJSONFormatter` can extract tracing information (`trace_id` and `span_id`) from the `context.Context` if you are using a tracing system like OpenTelemetry. Ensure that spans are started and the context is propagated correctly.

### Caller and Stack Trace
- **Caller Information**: The formatter includes the function name, file, and line number where the log was generated, aiding in debugging.
- **Stack Trace**: For logs at the `error` level or higher, a stack trace is included. This can be useful for diagnosing issues in production.

---

## Custom Formatter
If you need a different format or additional customization, you can implement your own formatter by satisfying the `logrus.Formatter` interface and providing it to the logger configuration.
```go
type MyCustomFormatter struct {
    // Custom fields...
}

func (f *MyCustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    // Custom formatting logic...
}
```
Usage:
```go
logConfig := logger.Config{
    Formatter: &MyCustomFormatter{},
}
```

## No-Op Logger
For testing purposes, you can use the no-operation logger, which implements the `Logger` interface but discards all log messages:
```go
log := logger.NewNoopLogger()
```
This can be useful to avoid cluttering test output with logs or when you need a logger that does nothing.

## Example
You can find a complete working example in the repository under [framework/logger/example](example/).