package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/kittipat1413/go-common/util/slice"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// Default JSON field keys for structured logging output.
// These constants ensure consistent field naming across all log entries.
const (
	DefaultSJsonFmtTimestampKey  = "timestamp"   // Default key for the timestamp field in logs
	DefaultSJsonFmtSeverityKey   = "severity"    // Default key for the severity field in logs
	DefaultSJsonFmtMessageKey    = "message"     // Default key for the message field in logs
	DefaultSJsonFmtErrorKey      = "error"       // Default key for the error field in logs
	DefaultSJsonFmtTraceIDKey    = "trace_id"    // Default key for the trace_id field in logs
	DefaultSJsonFmtSpanIDKey     = "span_id"     // Default key for the span_id field in logs
	DefaultSJsonFmtCallerKey     = "caller"      // Default key for the caller field in logs
	DefaultSJsonFmtCallerFuncKey = "function"    // Default key for the function field in logs
	DefaultSJsonFmtCallerFileKey = "file"        // Default key for the file field in logs
	DefaultSJsonFmtStackTraceKey = "stack_trace" // Default key for the stack_trace field in logs
)

// defaultSJsonFmtSkipPackages defines packages to exclude from caller detection.
// These are internal packages that should not appear as the "caller" in logs.
var defaultSJsonFmtSkipPackages = []string{
	"github.com/sirupsen/logrus",
	"github.com/kittipat1413/go-common/framework/logger",
}

// StructuredJSONFormatter is a custom logrus formatter for structured JSON logs.
// Produces consistent JSON output with trace correlation, caller information,
// and configurable field key formatting suitable for log aggregation systems.
//
// JSON output includes:
//   - timestamp: RFC3339 formatted time
//   - severity: log level (debug, info, warn, error, fatal)
//   - message: log message text
//   - error: error details for Error/Fatal levels
//   - trace_id: distributed trace correlation (if available in context)
//   - span_id: span correlation within traces (if available)
//   - caller: source location (function, file, line)
//   - stack_trace: call stack for Error/Fatal levels
//
// Example JSON output:
//
//	{
//	  "timestamp": "2023-01-01T12:00:00Z",
//	  "severity": "info",
//	  "message": "User logged in",
//	  "user_id": 123,
//	  "trace_id": "abc123...",
//	  "span_id": "def456...",
//	  "caller": {
//	    "function": "main.handleLogin",
//	    "file": "/app/handlers.go:45"
//	  },
//	  "stack_trace": "..." // Only for error/fatal levels
//	}
type StructuredJSONFormatter struct {
	// TimestampFormat specifies the time format for log timestamps.
	// Defaults to time.RFC3339 if not specified.
	TimestampFormat string

	// PrettyPrint enables JSON indentation for human-readable output.
	PrettyPrint bool

	// SkipPackages contains additional package prefixes to exclude from caller detection.
	// Combined with defaultSJsonFmtSkipPackages for comprehensive filtering.
	SkipPackages []string

	// FieldKeyFormatter allows customization of JSON field keys.
	// If nil, NoopFieldKeyFormatter is used (no transformation).
	FieldKeyFormatter FieldKeyFormatter
}

// FieldKeyFormatter defines a function type for customizing JSON field keys.
// Enables consistent field naming conventions across different logging contexts.
//
// Parameters:
//   - key: Original field key
//
// Returns:
//   - string: Transformed field key
//
// Example:
//
//	customFieldKeyFormatter := func(key string) string {
//		switch key {
//		case DefaultEnvironmentKey:
//			return "env"
//		case DefaultServiceNameKey:
//			return "service"
//		case DefaultSJsonFmtSeverityKey:
//			return "level"
//		default:
//			return key
//		}
//	}
type FieldKeyFormatter func(key string) string

// NoopFieldKeyFormatter is the default field key formatter that performs no transformation.
// Returns the original key unchanged, suitable when no key customization is needed.
func NoopFieldKeyFormatter(defaultKey string) string {
	return defaultKey
}

// Format implements the logrus.Formatter interface for JSON log formatting.
// Produces structured JSON output with all configured fields including
// timestamps, trace IDs, caller information, and custom fields.
//
// Parameters:
//   - entry: logrus.Entry containing log data and context
//
// Returns:
//   - []byte: Formatted JSON log entry with newline
//   - error: JSON marshaling error if formatting fails
//
// Example Output:
//
//	{"timestamp":"2023-01-01T12:00:00Z","severity":"info","message":"test","trace_id":"abc123"}
func (f *StructuredJSONFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Use the default field key formatter if not provided.
	if f.FieldKeyFormatter == nil {
		f.FieldKeyFormatter = NoopFieldKeyFormatter
	}

	// Prepare the data map for JSON serialization.
	data := make(logrus.Fields, len(entry.Data)+7)

	// Apply FieldKeyFormatter to keys in entry.Data and copy them to data.
	for key, value := range entry.Data {
		if key == DefaultErrorKey {
			continue // Skip the default error key
		}
		formattedKey := f.FieldKeyFormatter(key)
		switch v := value.(type) {
		case error:
			data[formattedKey] = v.Error()
		default:
			data[formattedKey] = v
		}
	}

	// Add predefined keys with formatted keys.
	data[f.FieldKeyFormatter(DefaultSJsonFmtTimestampKey)] = entry.Time.Format(f.TimestampFormat)
	data[f.FieldKeyFormatter(DefaultSJsonFmtSeverityKey)] = entry.Level.String()
	data[f.FieldKeyFormatter(DefaultSJsonFmtMessageKey)] = entry.Message

	// Include error message if present.
	if err, ok := entry.Data[DefaultErrorKey]; ok {
		formattedErrorKey := f.FieldKeyFormatter(DefaultSJsonFmtErrorKey)
		switch e := err.(type) {
		case error:
			data[formattedErrorKey] = e.Error()
		default:
			data[formattedErrorKey] = fmt.Sprintf("%v", e)
		}
	}

	// Include trace and span IDs if available.
	if entry.Context != nil {
		traceID, spanID := extractTraceIDs(entry.Context)
		if traceID != nil {
			data[f.FieldKeyFormatter(DefaultSJsonFmtTraceIDKey)] = *traceID
		}
		if spanID != nil {
			data[f.FieldKeyFormatter(DefaultSJsonFmtSpanIDKey)] = *spanID
		}
	}

	// Combine default and custom SkipPackages.
	skipPackages := slice.Union(f.SkipPackages, defaultSJsonFmtSkipPackages)

	// Caller's function name, file, and line number.
	function, file, line := getCaller(skipPackages)
	if function != "" && file != "" && line != 0 {
		callerInfo := map[string]string{
			f.FieldKeyFormatter(DefaultSJsonFmtCallerFuncKey): function,
			f.FieldKeyFormatter(DefaultSJsonFmtCallerFileKey): fmt.Sprintf("%s:%d", file, line),
		}
		data[f.FieldKeyFormatter(DefaultSJsonFmtCallerKey)] = callerInfo
	}

	// Stack trace for error levels.
	if entry.Level <= logrus.ErrorLevel {
		data[f.FieldKeyFormatter(DefaultSJsonFmtStackTraceKey)] = getStackTrace()
	}

	// Serialize the data to JSON.
	var serialized []byte
	var err error
	if f.PrettyPrint {
		serialized, err = json.MarshalIndent(data, "", "  ")
	} else {
		serialized, err = json.Marshal(data)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON: %v", err)
	}
	return append(serialized, '\n'), nil
}

// extractTraceIDs retrieves OpenTelemetry trace and span IDs from context.
// Returns nil values if no valid span is found in the context.
func extractTraceIDs(ctx context.Context) (*string, *string) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return nil, nil // No valid span
	}

	var traceID, spanID string
	spanCtx := span.SpanContext()
	if spanCtx.HasTraceID() {
		traceID = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		spanID = spanCtx.SpanID().String()
	}

	return &traceID, &spanID
}

// getStackTrace captures the current goroutine's stack trace with dynamic buffer sizing.
// Uses progressively larger buffers to handle stack traces of varying sizes efficiently.
func getStackTrace() string {
	bufSize := 1024
	maxBufSize := 32 * 1024 // 32 KB upper limit
	for bufSize <= maxBufSize {
		buf := make([]byte, bufSize)
		n := runtime.Stack(buf, false)
		if n < bufSize {
			// The buffer was large enough
			return string(buf[:n])
		}
		// Buffer was too small, increase the size and try again
		bufSize *= 2
	}
	// If all else fails, return what we have
	buf := make([]byte, bufSize)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// getCaller identifies the source location that triggered the log entry.
// Skips internal logging packages to find the actual application caller.
func getCaller(skipPackages []string) (function string, file string, line int) {
	const maxDepth = 25
	pcs := make([]uintptr, maxDepth)
	depth := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for {
		frame, more := frames.Next()

		if frame.Function == "" {
			if !more {
				break
			}
			continue
		}

		skip := false
		for _, pkg := range skipPackages {
			if strings.HasPrefix(frame.Function, pkg) {
				skip = true
				break
			}
		}

		if !skip {
			function = frame.Function
			file = frame.File
			line = frame.Line
			return
		}

		if !more {
			break
		}
	}
	return
}
