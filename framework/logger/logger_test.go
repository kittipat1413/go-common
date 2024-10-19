package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultLogger(t *testing.T) {
	log := logger.NewDefaultLogger()
	assert.NotNil(t, log, "default logger should not be nil")
}

func TestNewLogger_InvalidLogLevel(t *testing.T) {
	// Attempt to create a logger with an invalid log level
	invalidLevel := logger.LogLevel("invalid_level")
	_, err := logger.NewLogger(logger.Config{
		Level: invalidLevel,
	})
	assert.Error(t, err, "should return an error for invalid log level")
	assert.Equal(t, logger.ErrInvalidLogLevel, err, "error should be ErrInvalidLogLevel")
}

func TestLogger_LogLevels(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.DEBUG,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	ctx := context.Background()

	log.Debug(ctx, "Debug message", nil)
	log.Info(ctx, "Info message", nil)
	log.Warn(ctx, "Warn message", nil)
	log.Error(ctx, "Error message", errors.New("test error"), nil)

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

		if logEntry["severity"] == "error" {
			assert.Equal(t, "test error", logEntry["error"], "error message should match")
		}
	}
}

func TestLogger_ErrorLevel(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.ERROR,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	ctx := context.Background()

	// These logs should not appear due to the log level being ERROR
	log.Debug(ctx, "Debug message", nil)
	log.Info(ctx, "Info message", nil)
	log.Warn(ctx, "Warn message", nil)

	// This log should appear
	log.Error(ctx, "Error message", errors.New("test error"), nil)

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
}

func TestLogger_NilFields(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.INFO,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	ctx := context.Background()
	// Log with nil fields
	log.Info(ctx, "Info message with nil fields", nil)

	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 1, len(logEntries), "should have 1 log entry")

	var logEntry map[string]interface{}
	err = json.Unmarshal(logEntries[0], &logEntry)
	assert.NoError(t, err, "log entry should be valid JSON")

	// Check standard fields
	assert.Equal(t, "Info message with nil fields", logEntry["message"], "message should match")
	assert.Equal(t, "info", logEntry["severity"], "severity should match")
}

func TestLogger_ContextFields(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.INFO,
		Formatter: &logger.StructuredJSONFormatter{
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
	fields := logger.Fields{"key1": "value1", "key2": "value2"}

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
	assert.Equal(t, "value1", logEntry["key1"], "log field 'key1' should be 'value1'")
	assert.Equal(t, "value2", logEntry["key2"], "log field 'key2' should be 'value2'")
}

func TestSetDefaultLoggerConfig(t *testing.T) {
	// Backup the original default config and restore it after the test
	originalConfig := logger.Config{
		Level: logger.INFO,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: os.Stdout,
	}
	defer func() {
		err := logger.SetDefaultLoggerConfig(originalConfig)
		assert.NoError(t, err, "should not return an error")
	}()

	// Create a buffer to capture the logs
	buffer := &bytes.Buffer{}

	// Define a new custom config
	customConfig := logger.Config{
		Level: logger.DEBUG,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			PrettyPrint:     false,
		},
		Environment: "test_env",
		ServiceName: "test_service",
		Output:      buffer,
	}

	// Set the custom config as the default logger config
	err := logger.SetDefaultLoggerConfig(customConfig)
	assert.NoError(t, err, "should not return an error")

	// Get the default logger (it should now use the custom config)
	log := logger.NewDefaultLogger()

	// Log a debug message
	log.Debug(context.Background(), "Debug message", nil)

	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 1, len(logEntries), "should have 1 log entry")

	var logEntry map[string]interface{}
	err = json.Unmarshal(logEntries[0], &logEntry)
	assert.NoError(t, err, "log entry should be valid JSON")

	assert.Equal(t, "test_service", logEntry["service_name"], "serviceName should match")
	assert.Equal(t, "test_env", logEntry["environment"], "environment should match")
	assert.Equal(t, "debug", logEntry["severity"], "severity should match")
	assert.Equal(t, "Debug message", logEntry["message"], "message should match")
}

func TestSetDefaultLoggerConfig_InvalidLogLevel(t *testing.T) {
	// Backup the original default config
	originalConfig := logger.Config{
		Level: logger.INFO,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: os.Stdout,
	}
	defer func() {
		_ = logger.SetDefaultLoggerConfig(originalConfig)
	}()

	// Attempt to set an invalid log level in the default logger config
	invalidConfig := logger.Config{
		Level: logger.LogLevel("invalid_level"),
	}
	err := logger.SetDefaultLoggerConfig(invalidConfig)
	assert.Error(t, err, "should return an error for invalid log level")
	assert.Equal(t, logger.ErrInvalidLogLevel, err, "error should be ErrInvalidLogLevel")
}

func TestLogger_WithFields(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.INFO,
		Formatter: &logger.StructuredJSONFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Environment: "test_env",
		Output:      buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	// Create a logger with additional persistent fields
	logWithFields1 := log.WithFields(logger.Fields{
		"persistent_key1": "persistent_value1",
	})
	logWithFields2 := logWithFields1.WithFields(logger.Fields{
		"persistent_key2": "persistent_value2",
	})

	ctx := context.Background()
	fields := logger.Fields{"transient_key1": "transient_value1", "transient_key2": "transient_value2"}

	// Log a message using the logger with persistent fields
	logWithFields2.Info(ctx, "Info message with fields", fields)

	// Verify that both persistent and transient fields are present in the log entry
	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	// Remove the last empty entry if present
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 1, len(logEntries), "should have 1 log entry")

	var logEntry map[string]interface{}
	err = json.Unmarshal(logEntries[0], &logEntry)
	assert.NoError(t, err, "log entry should be valid JSON")

	// Check that persistent fields are present
	assert.Contains(t, logEntry, "environment", "log should contain 'environment'")
	assert.Equal(t, "test_env", logEntry["environment"], "value of 'environment' should be 'test_env'")
	assert.Contains(t, logEntry, "persistent_key1", "log should contain 'persistent_key'")
	assert.Equal(t, "persistent_value1", logEntry["persistent_key1"], "value of 'persistent_key1' should be 'persistent_value1'")
	assert.Contains(t, logEntry, "persistent_key2", "log should contain 'persistent_key'")
	assert.Equal(t, "persistent_value2", logEntry["persistent_key2"], "value of 'persistent_key2' should be 'persistent_value2'")

	// Check that transient fields are present
	assert.Contains(t, logEntry, "transient_key1", "log should contain 'transient_key1'")
	assert.Equal(t, "transient_value1", logEntry["transient_key1"], "value of 'transient_key1' should be 'transient_value1'")
	assert.Contains(t, logEntry, "transient_key2", "log should contain 'transient_key2'")
	assert.Equal(t, "transient_value2", logEntry["transient_key2"], "value of 'transient_key2' should be 'transient_value2'")

	// Check standard fields
	assert.Equal(t, "Info message with fields", logEntry["message"], "message should match")
	assert.Equal(t, "info", logEntry["severity"], "severity should match")
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
