# Logger Package
The logger package provides a structured, context-aware logging solution for Go applications. It is built on top of the [logrus](https://github.com/sirupsen/logrus) library and is designed to facilitate easy integration with your projects, offering features like:
- JSON-formatted logs suitable for production environments.
- Support for multiple log levels (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`).
- Context propagation to include tracing information (e.g., `trace_id`, `span_id`).
- Customizable log formatters and output destinations.
- Integration with web frameworks like Gin.

## Features
- **Structured Logging**: Outputs logs in JSON format, making them easy to parse and analyze.
- **Context-Aware**: Supports logging with `context.Context`, allowing you to include tracing information automatically.
- **Customizable Formatter**: Use the default `ProductionFormatter` or provide your own formatter to customize the log output.
- **Environment and Service Name**: Optionally include environment and service name in your logs for better traceability.
- **No-Op Logger**: Provides a no-operation logger for testing purposes, which discards all log messages.

## Usage

### Creating a Logger
You can create a logger using the `NewLogger` function, providing a `Config` struct to customize its behavior:
```golang
import (
    "github.com/kittipat1413/go-common/framework/logger"
    "github.com/kittipat1413/go-common/framework/logger/formatter"
    "time"
)

logConfig := logger.Config{
    Level: logger.INFO,
    Formatter: &formatter.ProductionFormatter{
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
```

Alternatively, you can use the default logger:
```golang
log := logger.NewDefaultLogger()
```

## Configuration
The Config struct allows you to customize the logger:
```golang
type Config struct {
	// Level determines the minimum log level that will be processed by the logger.
	// Logs with a level lower than this will be ignored.
	Level LogLevel
	// Formatter is an optional field for specifying a custom logrus formatter.
	// If not provided, the logger will use the ProductionFormatter by default.
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
```golang
type Logger interface {
	Debug(ctx context.Context, msg string, fields Fields)
	Info(ctx context.Context, msg string, fields Fields)
	Warn(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, err error, fields Fields)
	Fatal(ctx context.Context, msg string, err error, fields Fields)
}
```
Example:
```golang
ctx := context.Background()
fields := logger.Fields{"user_id": 12345}

log.Info(ctx, "User logged in", fields)
```
### Including Errors
For error and fatal logs, you can include an error object:
```golang
err := errors.New("something went wrong")
log.Error(ctx, "Failed to process request", err, fields)
```

You can find a complete working example in the repository under [framework/logger/example](example/).

---

## ProductionFormatter
The `ProductionFormatter` is a custom `logrus.Formatter` designed for production environments. It outputs logs in JSON format with a standardized structure, making it suitable for log aggregation and analysis tools.
### Features
- **Timestamp**: Includes a timestamp formatted according to `TimestampFormat`.
- **Severity**: The log level (`debug`, `info`, `warning`, `error`, `fatal`).
- **Message**: The log message.
- **Error** Handling: Automatically includes error messages if an error is provided.
- **Tracing Information**: Extracts `trace_id` and `span_id` from the context if available (e.g., when using OpenTelemetry).
- **Caller Information**: Adds information about the function, file, and line number where the log was generated.
- **Stack Trace**: Includes a stack trace for logs at the `error` level or higher.
- **Custom Fields**: Supports additional fields provided via `logger.Fields`.

### Configuration
You can customize the `ProductionFormatter` when initializing the logger:
```golang
import (
    "github.com/kittipat1413/go-common/framework/logger/formatter"
    "time"
)

formatter := &formatter.ProductionFormatter{
    TimestampFormat: time.RFC3339, // Customize timestamp format
    PrettyPrint:     true,         // Indent JSON output
    SkipPackages: []string{
        "myproject/internal/pkg1", // Packages to skip when determining the caller
    },
}

logConfig := logger.Config{
    Level:     logger.INFO,
    Formatter: formatter,
}
```

Example Log Entry
```json
{
  "service_name": "logger-example",
  "environment": "development",
  "timestamp": "2024-10-15T20:03:43+07:00",
  "severity": "error",
  "message": "Failed to process request",
  "error": "database connection failed",
  "trace_id": "4d1e00c0e6c0a0c3",
  "span_id": "6d1e00c0e6c0a0c4",
  "caller": {
    "function": "service.ProcessRequest",
    "file": "/path/to/service/handler.go:42"
  },
  "stack_trace": "goroutine 1 [running]:\nmain.main()\n\t/path/to/main.go:12 +0x25",
  "user_id": 12345
}
```

### Tracing Integration
The `ProductionFormatter` can extract tracing information (`trace_id` and `span_id`) from the `context.Context` if you are using a tracing system like OpenTelemetry. Ensure that spans are started and the context is propagated correctly.

### Caller and Stack Trace
- **Caller Information**: The formatter includes the function name, file, and line number where the log was generated, aiding in debugging.
- **Stack Trace**: For logs at the `error` level or higher, a stack trace is included. This can be useful for diagnosing issues in production.

---

## Custom Formatter
If you need a different format or additional customization, you can implement your own formatter by satisfying the `logrus.Formatter` interface and providing it to the logger configuration.
```golang
type MyCustomFormatter struct {
    // Custom fields...
}

func (f *MyCustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
    // Custom formatting logic...
}
```
Usage:
```golang
logConfig := logger.Config{
    Formatter: &MyCustomFormatter{},
}
```

## No-Op Logger
For testing purposes, you can use the no-operation logger, which implements the Logger interface but discards all log messages:
```golang
log := logger.NewNoopLogger()
```
This can be useful to avoid cluttering test output with logs or when you need a logger that does nothing.

## Example
You can find a complete working example in the repository under [framework/logger/example](example/).