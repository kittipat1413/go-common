package logger

import (
	"context"
	"os"
	"time"

	"github.com/kittipat1413/go-common/framework/logger/formatter"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -source=./logger.go -destination=./mocks/logger.go -package=logger_mocks
type Logger interface {
	Debug(ctx context.Context, msg string, fields Fields)
	Info(ctx context.Context, msg string, fields Fields)
	Warn(ctx context.Context, msg string, fields Fields)
	Error(ctx context.Context, msg string, err error, fields Fields)
	Fatal(ctx context.Context, msg string, err error, fields Fields)
}

/*
DefaultLogger returns a logger instance with default configuration.
The logger uses the ProductionFormatter, which outputs logs in JSON format
with the following fields:

  - timestamp: formatted in RFC3339 format.
  - severity: the severity level of the log (e.g., info, debug, error).
  - trace_id: the trace identifier for correlating logs with distributed
    traces (if available).
  - span_id: the span identifier for correlating logs within specific spans
    of a trace (if available).
  - caller: the function, file, and line number where the log was generated.
  - stack_trace: included for logs with error-level severity or higher,
    providing additional debugging context.
*/
func NewDefaultLogger() Logger {
	defaultLog, _ := NewLogger(Config{
		Level: INFO,
		Formatter: &formatter.ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
	})
	return defaultLog
}

// logger is implementation of Logger interface.
type logger struct {
	baselogger  *logrus.Logger
	logLevel    LogLevel
	environment string
	serviceName string
}

// Config holds the logger configuration.
type Config struct {
	// Level determines the minimum log level that will be processed by the logger.
	// Logs with a level lower than this will be ignored.
	Level LogLevel
	// Formatter is as optional field for specifying a custom logrus formatter.
	// If not provided, the logger will use the ProductionFormatter by default.
	Formatter logrus.Formatter
	// Environment is an optional field for specifying the running environment (e.g., "production", "staging").
	// This field is used for adding environment-specific fields to logs.
	Environment string
	// ServiceName is an optional field for specifying the name of the service.
	// This field is used for adding service-specific fields to logs.
	ServiceName string
}

// NewLogger creates a new logger instance with the provided configuration.
func NewLogger(config Config) (Logger, error) {
	logrusLogger := logrus.New()

	// Set custom formatter if provided, otherwise use ProductionFormatter
	if config.Formatter != nil {
		logrusLogger.SetFormatter(config.Formatter)
	} else {
		logrusLogger.SetFormatter(&formatter.ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		})
	}

	// Set log level
	logrusLogger.SetLevel(config.Level.ToLogrusLevel())

	// Set output to stdout
	logrusLogger.SetOutput(os.Stdout)

	return &logger{
		baselogger:  logrusLogger,
		logLevel:    config.Level,
		environment: config.Environment,
		serviceName: config.ServiceName,
	}, nil
}

// Fields represents a key-value pair for structured logging.
type Fields map[string]any

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
	fields[logrus.ErrorKey] = err
	l.logWithContext(ctx, logrus.ErrorLevel, msg, fields)
}

// Fatal logs a message at the Fatal level and exits the application.
func (l *logger) Fatal(ctx context.Context, msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	fields[logrus.ErrorKey] = err
	l.logWithContext(ctx, logrus.FatalLevel, msg, fields)
}

// logWithContext logs a message with the provided context and fields.
func (l *logger) logWithContext(ctx context.Context, level logrus.Level, msg string, fields Fields) {
	entry := l.baselogger.WithContext(ctx).WithFields(logrus.Fields(fields))
	// Add optional environment and service_name fields if provided
	if l.environment != "" {
		entry = entry.WithField(environmentKey, l.environment)
	}
	if l.serviceName != "" {
		entry = entry.WithField(serviceNameKey, l.serviceName)
	}
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

// NoopLogger returns a no-op logger for tests.
type noopLogger struct{}

func NoopLogger() Logger {
	return &noopLogger{}
}
func (n *noopLogger) Debug(ctx context.Context, msg string, fields Fields)            {}
func (n *noopLogger) Info(ctx context.Context, msg string, fields Fields)             {}
func (n *noopLogger) Warn(ctx context.Context, msg string, fields Fields)             {}
func (n *noopLogger) Error(ctx context.Context, msg string, err error, fields Fields) {}
func (n *noopLogger) Fatal(ctx context.Context, msg string, err error, fields Fields) {}
