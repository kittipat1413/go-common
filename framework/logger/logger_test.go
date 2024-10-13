package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/kittipat1413/go-common/framework/logger/formatter"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultLogger(t *testing.T) {
	log := logger.NewDefaultLogger()
	assert.NotNil(t, log, "default logger should not be nil")
}

func TestLogger_LogLevels(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.DEBUG,
		Formatter: &formatter.ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	ctx := context.Background()
	fields := logger.Fields{"key": "value"}

	log.Debug(ctx, "Debug message", fields)
	log.Info(ctx, "Info message", fields)
	log.Warn(ctx, "Warn message", fields)
	log.Error(ctx, "Error message", errors.New("test error"), fields)

	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	// Remove the last empty entry if present
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 4, len(logEntries), "should have 4 log entries")

	expectedLevels := []string{"debug", "info", "warning", "error"}
	expectedMessages := []string{"Debug message", "Info message", "Warn message", "Error message"}

	for i, entry := range logEntries {
		var logEntry map[string]interface{}
		err := json.Unmarshal(entry, &logEntry)
		assert.NoError(t, err, "log entry should be valid JSON")

		assert.Equal(t, expectedLevels[i], logEntry["severity"], "log level should match")
		assert.Equal(t, expectedMessages[i], logEntry["message"], "log message should match")
		assert.Equal(t, "value", logEntry["key"], "log field 'key' should be 'value'")

		if logEntry["severity"] == "error" {
			assert.Equal(t, "test error", logEntry["error"], "error message should match")
		}
	}
}

func TestLogger_ErrorLevel(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.ERROR,
		Formatter: &formatter.ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	ctx := context.Background()
	fields := logger.Fields{"test_field": "error_value"}

	// These logs should not appear due to the log level being ERROR
	log.Debug(ctx, "Debug message", fields)
	log.Info(ctx, "Info message", fields)
	log.Warn(ctx, "Warn message", fields)

	// This log should appear
	log.Error(ctx, "Error message", errors.New("test error"), fields)

	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	// Remove the last empty entry if present
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 1, len(logEntries), "should have 1 log entry")

	var logEntry map[string]interface{}
	err = json.Unmarshal(logEntries[0], &logEntry)
	assert.NoError(t, err, "log entry should be valid JSON")

	assert.Equal(t, "error", logEntry["severity"], "log level should be 'error'")
	assert.Equal(t, "Error message", logEntry["message"], "log message should match")
	assert.Equal(t, "test error", logEntry["error"], "error message should match")
	assert.Equal(t, "error_value", logEntry["test_field"], "log field 'test_field' should be 'error_value'")
}

func TestLogger_ContextFields(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.INFO,
		Formatter: &formatter.ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output:      buffer,
		Environment: "test_env",
		ServiceName: "test_service",
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	ctx := context.Background()
	fields := logger.Fields{"key": "value"}

	log.Info(ctx, "Info message", fields)

	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	// Remove the last empty entry if present
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 1, len(logEntries), "should have 1 log entry")

	var logEntry map[string]interface{}
	err = json.Unmarshal(logEntries[0], &logEntry)
	assert.NoError(t, err, "log entry should be valid JSON")

	assert.Equal(t, "Info message", logEntry["message"], "log message should match")
	assert.Equal(t, "test_env", logEntry["environment"], "environment should match")
	assert.Equal(t, "test_service", logEntry["service_name"], "serviceName should match")
	assert.Equal(t, "value", logEntry["key"], "log field 'key' should be 'value'")
}

func TestNoopLogger(t *testing.T) {
	log := logger.NewNoopLogger()
	assert.NotNil(t, log, "noopLogger should not be nil")

	ctx := context.Background()
	fields := logger.Fields{"key": "value"}

	// Ensure that calling the NoopLogger methods does not panic
	assert.NotPanics(t, func() {
		log.Debug(ctx, "Debug message", fields)
		log.Info(ctx, "Info message", fields)
		log.Warn(ctx, "Warn message", fields)
		log.Error(ctx, "Error message", errors.New("test error"), fields)
		// Commenting out Fatal to avoid calling os.Exit in tests
		// log.Fatal(ctx, "Fatal message", errors.New("test error"), fields)
	}, "noopLogger methods should not panic")
}
