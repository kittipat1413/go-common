package logger

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=./logger.go -destination=./mocks/logger.go -package=logger_mocks

// Logger defines the interface for structured logging with context support.
type Logger interface {
	// WithFields returns a new logger instance with additional structured fields.
	// Fields are merged with existing fields, with new fields taking precedence.
	WithFields(fields Fields) Logger

	// Debug logs debug-level messages for detailed troubleshooting.
	Debug(ctx context.Context, msg string, fields Fields)

	// Info logs informational messages about normal application flow.
	Info(ctx context.Context, msg string, fields Fields)

	// Warn logs warning messages for potentially harmful situations.
	Warn(ctx context.Context, msg string, fields Fields)

	// Error logs error messages for failure conditions.
	Error(ctx context.Context, msg string, err error, fields Fields)

	// Fatal logs critical errors and terminates the application.
	Fatal(ctx context.Context, msg string, err error, fields Fields)
}

// Package-level errors
var (
	// ErrInvalidLogLevel is returned when an invalid log level is specified.
	ErrInvalidLogLevel = errors.New("invalid log level")
)

// Default logger configuration and synchronization
var (
	// Default logger configuration with JSON formatting and Info level.
	defaultLoggerConfig = Config{
		Level: INFO,
		Formatter: &StructuredJSONFormatter{
			TimestampFormat:   time.RFC3339,
			PrettyPrint:       false,
			FieldKeyFormatter: NoopFieldKeyFormatter,
		},
		Output: os.Stdout,
	}
	// Mutex for protecting the default logger configuration.
	defaultLoggerMutex sync.RWMutex
)

// SetDefaultLoggerConfig sets a custom configuration for the default logger.
// Validates the configuration by attempting to create a logger with it.
// If validation fails, the existing default configuration remains unchanged.
//
// Parameters:
//   - config: New default configuration to apply
//
// Returns:
//   - error: ErrInvalidLogLevel or other configuration errors
//
// Example:
//
//	config := Config{Level: DEBUG, Output: os.Stderr}
//	err := logger.SetDefaultLoggerConfig(config)
func SetDefaultLoggerConfig(config Config) error {
	// Lock the mutex to protect the defaultLoggerConfig.
	defaultLoggerMutex.Lock()
	defer defaultLoggerMutex.Unlock()

	// Try to create a logger with the new configuration.
	_, err := NewLogger(config)
	if err != nil {
		// If there is an error, keep the original configuration unchanged.
		return err
	}
	// If logger creation is successful, update the default configuration.
	defaultLoggerConfig = config
	return nil
}

// NewDefaultLogger returns a logger instance with the current default configuration.
// Uses user-defined configuration if SetDefaultLoggerConfig was called, otherwise uses package defaults.
//
// Default JSON output includes:
//   - timestamp: RFC3339 formatted time
//   - severity: log level (debug, info, warn, error, fatal)
//   - message: log message text
//   - error: error details for Error/Fatal levels
//   - trace_id: distributed trace correlation (if available in context)
//   - span_id: span correlation within traces (if available)
//   - caller: source location (function, file, line)
//   - stack_trace: call stack for Error/Fatal levels
//
// Returns:
//   - Logger: Configured logger instance ready for use
//
// Example:
//
//	logger := logger.NewDefaultLogger()
//	logger.Info(ctx, "Application started", nil)
func NewDefaultLogger() Logger {
	defaultLoggerMutex.RLock()
	config := defaultLoggerConfig
	defaultLoggerMutex.RUnlock()

	defaultLog, _ := NewLogger(config)
	return defaultLog
}

// logger is the concrete implementation of the Logger interface using logrus.
type logger struct {
	baselogger *logrus.Logger // Underlying logrus instance
	logLevel   LogLevel       // Current log level
	fields     Fields         // Persistent fields for this logger instance
}

// Config holds configuration parameters for logger creation.
// Provides options for level, formatting, output destination, and metadata.
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

// NewLogger creates a new logger instance with the provided configuration.
// Validates configuration and sets up logrus with specified formatting and output.
//
// Parameters:
//   - config: Logger configuration including level, formatter, and metadata
//
// Returns:
//   - Logger: Configured logger instance
//   - error: ErrInvalidLogLevel if log level is invalid
//
// Example:
//
//	config := Config{
//	    Level:       DEBUG,
//	    Environment: "development",
//	    ServiceName: "user-service",
//	}
//	logger, err := logger.NewLogger(config)
func NewLogger(config Config) (Logger, error) {
	logrusLogger := logrus.New()

	// Set custom formatter if provided, otherwise use StructuredJSONFormatter.
	if config.Formatter != nil {
		logrusLogger.SetFormatter(config.Formatter)
	} else {
		logrusLogger.SetFormatter(&StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		})
	}

	// Set log level.
	if !config.Level.IsValid() {
		return nil, ErrInvalidLogLevel
	}
	logrusLogger.SetLevel(config.Level.ToLogrusLevel())

	// Set output to the provided output or default to stdout.
	if config.Output != nil {
		logrusLogger.SetOutput(config.Output)
	} else {
		logrusLogger.SetOutput(os.Stdout)
	}

	// Add environment and service name fields to the logger.
	fields := make(Fields)
	if config.Environment != "" {
		fields[DefaultEnvironmentKey] = config.Environment
	}
	if config.ServiceName != "" {
		fields[DefaultServiceNameKey] = config.ServiceName
	}

	return &logger{
		baselogger: logrusLogger,
		logLevel:   config.Level,
		fields:     fields,
	}, nil
}

// clone creates a deep copy of the logger for safe field modification.
// Used internally by WithFields to avoid modifying the original logger.
func (l *logger) clone() *logger {
	c := *l
	// Deep copy the fields map.
	c.fields = make(Fields, len(l.fields))
	for k, v := range l.fields {
		c.fields[k] = v
	}
	return &c
}

// Fields represents structured key-value pairs for rich logging context.
// Used to add metadata, request IDs, user information, or other contextual data.
type Fields map[string]interface{}

// WithFields returns a new logger instance with additional structured fields.
// Original logger remains unchanged. New fields override existing ones with same keys.
//
// Parameters:
//   - fields: Key-value pairs to add to the logger context
//
// Returns:
//   - Logger: New logger instance with merged fields
//
// Example:
//
//	userLogger := logger.WithFields(Fields{
//	    "user_id": 123,
//	    "request_id": "req-abc-123",
//	})
//	userLogger.Info(ctx, "User logged in", nil)
func (l *logger) WithFields(fields Fields) Logger {
	clone := l.clone()
	// Add new fields to the cloned logger's fields.
	for key, value := range fields {
		clone.fields[key] = value
	}
	return clone
}

// Debug logs a message at the Debug level.
func (l *logger) Debug(ctx context.Context, msg string, fields Fields) {
	l.logWithContext(ctx, logrus.DebugLevel, msg, fields)
}

// Info logs a message at the Info level.
func (l *logger) Info(ctx context.Context, msg string, fields Fields) {
	l.logWithContext(ctx, logrus.InfoLevel, msg, fields)
}

// Warn logs a message at the Warn level.
func (l *logger) Warn(ctx context.Context, msg string, fields Fields) {
	l.logWithContext(ctx, logrus.WarnLevel, msg, fields)
}

// Error logs a message at the Error level.
func (l *logger) Error(ctx context.Context, msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	if err != nil {
		fields[DefaultErrorKey] = err
	}
	l.logWithContext(ctx, logrus.ErrorLevel, msg, fields)
}

// Fatal logs a message at the Fatal level and terminates the application.
func (l *logger) Fatal(ctx context.Context, msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	if err != nil {
		fields[DefaultErrorKey] = err
	}
	l.logWithContext(ctx, logrus.FatalLevel, msg, fields)
}

// logWithContext handles the core logging logic with context integration and field merging.
// Merges logger fields with provided fields and delegates to appropriate logrus methods.
func (l *logger) logWithContext(ctx context.Context, level logrus.Level, msg string, fields Fields) {
	entry := l.baselogger.WithContext(ctx)

	// Merge logger's fields with input fields.
	mergedFields := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}
	entry = entry.WithFields(logrus.Fields(mergedFields))

	// Log the message at the specified level.
	switch level {
	case logrus.DebugLevel:
		entry.Debug(msg)
	case logrus.InfoLevel:
		entry.Info(msg)
	case logrus.WarnLevel:
		entry.Warn(msg)
	case logrus.ErrorLevel:
		entry.Error(msg)
	case logrus.FatalLevel:
		entry.Fatal(msg)
	}
}

// noopLogger is a logger implementation that discards all log messages.
// Useful for testing or when logging needs to be completely disabled.
type noopLogger struct{}

// NewNoopLogger returns a no-operation logger that silently discards all messages.
// Useful for testing scenarios or when logging needs to be disabled entirely.
//
// Returns:
//   - Logger: No-op logger that ignores all logging calls
//
// Example:
//
//	logger := logger.NewNoopLogger()
//	logger.Info(ctx, "This message is discarded", nil) // No output
func NewNoopLogger() Logger {
	return &noopLogger{}
}
func (n *noopLogger) WithFields(fields Fields) Logger                                 { return n }
func (n *noopLogger) Debug(ctx context.Context, msg string, fields Fields)            {}
func (n *noopLogger) Info(ctx context.Context, msg string, fields Fields)             {}
func (n *noopLogger) Warn(ctx context.Context, msg string, fields Fields)             {}
func (n *noopLogger) Error(ctx context.Context, msg string, err error, fields Fields) {}
func (n *noopLogger) Fatal(ctx context.Context, msg string, err error, fields Fields) {}
