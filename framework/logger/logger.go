package logger

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=./logger.go -destination=./mocks/logger.go -package=logger_mocks
type Logger interface {
	WithFields(fields Fields) Logger
	Debug(ctx context.Context, msg string, fields Fields)
	Info(ctx context.Context, msg string, fields Fields)
	Warn(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, err error, fields Fields)
	Fatal(ctx context.Context, msg string, err error, fields Fields)
}

var (
	// Default logger configuration.
	defaultLoggerConfig = Config{
		Level: INFO,
		Formatter: &ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: os.Stdout,
	}
	// Mutex for protecting the default logger configuration.
	defaultLoggerMutex sync.RWMutex
)

// SetDefaultLoggerConfig allows users to set a custom configuration for the default logger.
// This function can be called multiple times to update the configuration.
func SetDefaultLoggerConfig(config Config) {
	defaultLoggerMutex.Lock()
	defer defaultLoggerMutex.Unlock()
	defaultLoggerConfig = config
}

/*
NewDefaultLogger returns a logger instance with default or user-defined configuration by calling SetDefaultLoggerConfig.
The logger uses the ProductionFormatter, which outputs logs in JSON format
with the following fields:

  - timestamp: formatted in RFC3339 format.
  - severity: the severity level of the log (e.g., info, debug, error).
  - message: the log message.
  - error: the error message for logs with error-level severity or higher.
  - trace_id: the trace identifier for correlating logs with distributed
    traces (if available).
  - span_id: the span identifier for correlating logs within specific spans
    of a trace (if available).
  - caller: the function, file, and line number where the log was generated.
  - stack_trace: included for logs with error-level severity or higher,
    providing additional debugging context.
*/
func NewDefaultLogger() Logger {
	defaultLoggerMutex.RLock()
	config := defaultLoggerConfig
	defaultLoggerMutex.RUnlock()

	defaultLog, _ := NewLogger(config)
	return defaultLog
}

// logger is the implementation of the Logger interface.
type logger struct {
	baselogger *logrus.Logger
	logLevel   LogLevel
	fields     Fields
}

// Config holds the logger configuration.
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

// NewLogger creates a new logger instance with the provided configuration.
func NewLogger(config Config) (Logger, error) {
	logrusLogger := logrus.New()

	// Set custom formatter if provided, otherwise use ProductionFormatter.
	if config.Formatter != nil {
		logrusLogger.SetFormatter(config.Formatter)
	} else {
		logrusLogger.SetFormatter(&ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		})
	}

	// Set log level.
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

// clone creates a deep copy of the logger.
func (l *logger) clone() *logger {
	c := *l
	// Deep copy the fields map.
	c.fields = make(Fields, len(l.fields))
	for k, v := range l.fields {
		c.fields[k] = v
	}
	return &c
}

// Fields represents a key-value pair for structured logging.
type Fields map[string]interface{}

// WithFields returns a new logger that includes the provided fields.
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

// Fatal logs a message at the Fatal level and exits the application.
func (l *logger) Fatal(ctx context.Context, msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	if err != nil {
		fields[DefaultErrorKey] = err
	}
	l.logWithContext(ctx, logrus.FatalLevel, msg, fields)
}

// logWithContext logs a message with the provided context and fields.
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

type noopLogger struct{}

// NewNoopLogger returns a no-op logger that discards all log messages.
func NewNoopLogger() Logger {
	return &noopLogger{}
}
func (n *noopLogger) WithFields(fields Fields) Logger                                 { return n }
func (n *noopLogger) Debug(ctx context.Context, msg string, fields Fields)            {}
func (n *noopLogger) Info(ctx context.Context, msg string, fields Fields)             {}
func (n *noopLogger) Warn(ctx context.Context, msg string, fields Fields)             {}
func (n *noopLogger) Error(ctx context.Context, msg string, err error, fields Fields) {}
func (n *noopLogger) Fatal(ctx context.Context, msg string, err error, fields Fields) {}
