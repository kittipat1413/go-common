package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestProductionFormatter__WithCustomFieldKeyFormatter(t *testing.T) {
	buffer := &bytes.Buffer{}

	customFieldKeyFormatter := func(key string) string {
		return strings.ToUpper(key)
	}

	log, err := logger.NewLogger(logger.Config{
		Level: logger.INFO,
		Formatter: &logger.ProductionFormatter{
			TimestampFormat:   time.RFC3339,
			PrettyPrint:       false,
			FieldKeyFormatter: customFieldKeyFormatter,
		},
		Output: buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	ctx := context.Background()
	fields := logger.Fields{"custom_key": "custom_value"}

	log.Info(ctx, "Info message", fields)

	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 1, len(logEntries), "should have 1 log entry")

	var logEntry map[string]interface{}
	err = json.Unmarshal(logEntries[0], &logEntry)
	assert.NoError(t, err, "log entry should be valid JSON")

	// Custom fields should remain unchanged
	assert.Contains(t, logEntry, "CUSTOM_KEY", "log should contain 'custom_key'")
	assert.Equal(t, "custom_value", logEntry["CUSTOM_KEY"], "value of 'custom_key' should be 'custom_value'")

	// Standard fields should be uppercase
	assert.Contains(t, logEntry, "MESSAGE", "log should contain 'MESSAGE'")
	assert.Equal(t, "Info message", logEntry["MESSAGE"], "message should match")

	assert.Contains(t, logEntry, "SEVERITY", "log should contain 'SEVERITY'")
	assert.Equal(t, "info", logEntry["SEVERITY"], "severity should match")
}

func TestLogger_WithTraceAndSpanIDs(t *testing.T) {
	buffer := &bytes.Buffer{}
	log, err := logger.NewLogger(logger.Config{
		Level: logger.INFO,
		Formatter: &logger.ProductionFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		},
		Output: buffer,
	})
	assert.NoError(t, err)
	assert.NotNil(t, log)

	// Set up OpenTelemetry Tracer
	tracerProvider := sdktrace.NewTracerProvider()
	defer func() { _ = tracerProvider.Shutdown(context.Background()) }()
	tracer := tracerProvider.Tracer("test-tracer")

	// Start a new span
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	// Log a message within the context that includes the span
	log.Info(ctx, "Info message with trace and span IDs", nil)

	// Parse the log output
	logEntries := bytes.Split(buffer.Bytes(), []byte("\n"))
	// Remove the last empty entry if present
	if len(logEntries) > 0 && len(logEntries[len(logEntries)-1]) == 0 {
		logEntries = logEntries[:len(logEntries)-1]
	}

	assert.Equal(t, 1, len(logEntries), "should have 1 log entry")

	var logEntry map[string]interface{}
	err = json.Unmarshal(logEntries[0], &logEntry)
	assert.NoError(t, err, "log entry should be valid JSON")

	// Check that trace_id and span_id are present
	assert.Contains(t, logEntry, "trace_id", "log should contain 'trace_id'")
	assert.Contains(t, logEntry, "span_id", "log should contain 'span_id'")

	// Check that the IDs match the span context
	traceID := logEntry["trace_id"].(string)
	spanID := logEntry["span_id"].(string)
	expectedTraceID := span.SpanContext().TraceID().String()
	expectedSpanID := span.SpanContext().SpanID().String()
	assert.Equal(t, expectedTraceID, traceID, "trace_id should match")
	assert.Equal(t, expectedSpanID, spanID, "span_id should match")

	// Check standard fields
	assert.Equal(t, "Info message with trace and span IDs", logEntry["message"], "message should match")
	assert.Equal(t, "info", logEntry["severity"], "severity should match")
}
