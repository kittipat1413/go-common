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

const (
	DefaultSJsonFmtTimestampKey  = "timestamp"
	DefaultSJsonFmtSeverityKey   = "severity"
	DefaultSJsonFmtMessageKey    = "message"
	DefaultSJsonFmtErrorKey      = "error"
	DefaultSJsonFmtTraceIDKey    = "trace_id"
	DefaultSJsonFmtSpanIDKey     = "span_id"
	DefaultSJsonFmtCallerKey     = "caller"
	DefaultSJsonFmtCallerFuncKey = "function"
	DefaultSJsonFmtCallerFileKey = "file"
	DefaultSJsonFmtStackTraceKey = "stack_trace"
)

var defaultSJsonFmtSkipPackages = []string{
	"github.com/sirupsen/logrus",
	"github.com/kittipat1413/go-common/framework/logger",
}

/*
StructuredJSONFormatter is a custom logrus formatter for structured JSON logs.
It includes the following fields:
  - timestamp: The log timestamp in the specified format.
  - severity: The log severity level (e.g., info, debug, error).
  - message: The log message.
  - error: The error message if present.
  - trace_id: The trace ID if available.
  - span_id: The span ID if available.
  - caller: The caller's function name, file, and line number.
  - stack_trace: The stack trace for error levels.
*/
type StructuredJSONFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string
	// PrettyPrint will indent all JSON logs.
	PrettyPrint bool
	// SkipPackages is a list of packages to skip when searching for the caller.
	SkipPackages []string
	// FieldKeyFormatter is a function type that allows users to customize log field keys.
	FieldKeyFormatter FieldKeyFormatter
}

/*
FieldKeyFormatter is a function type that allows users to customize the keys of log fields.

Example usage:

	customFieldKeyFormatter := func(key string) string {
		return strings.ToUpper(key)
	}
*/
type FieldKeyFormatter func(key string) string

// NoopFieldKeyFormatter is the default implementation of FieldKeyFormatter.
// It returns the key unchanged, effectively performing no operation on the key.
func NoopFieldKeyFormatter(defaultKey string) string {
	return defaultKey
}

// Format implements the logrus.Formatter interface.
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

// extractTraceIDs retrieves the trace and span IDs from the context.
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

// getStackTrace retrieves the current stack trace.
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

// getCaller retrieves the caller's function name, file, and line number,
// skipping frames from the specified packages.
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
