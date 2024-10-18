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

/*
ProductionFormatter is a custom logrus formatter for production use.
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
type ProductionFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string
	// PrettyPrint will indent all JSON logs.
	PrettyPrint bool
	// SkipPackages is a list of packages to skip when searching for the caller.
	SkipPackages []string
	// FieldKeyFormatter is a function type that allows users to customize log field keys.
	FieldKeyFormatter FieldKeyFormatter
}

const (
	DefaultProdFmtTimestampKey  = "timestamp"
	DefaultProdFmtSeverityKey   = "severity"
	DefaultProdFmtMessageKey    = "message"
	DefaultProdFmtErrorKey      = "error"
	DefaultProdFmtTraceIDKey    = "trace_id"
	DefaultProdFmtSpanIDKey     = "span_id"
	DefaultProdFmtCallerKey     = "caller"
	DefaultProdFmtCallerFuncKey = "function"
	DefaultProdFmtCallerFileKey = "file"
	DefaultProdFmtStackTraceKey = "stack_trace"
)

var defaultProdFmtSkipPackages = []string{
	"github.com/sirupsen/logrus",
	"github.com/kittipat1413/go-common/framework/logger",
}

// Format implements the logrus.Formatter interface.
func (f *ProductionFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Use the default field key formatter if not provided.
	if f.FieldKeyFormatter == nil {
		f.FieldKeyFormatter = NoopFieldKeyFormatter
	}

	// Prepare the data map for JSON serialization.
	data := make(logrus.Fields, len(entry.Data)+7)

	// Apply FieldKeyFormatter to keys in entry.Data and copy them to data.
	for key, value := range entry.Data {
		formattedKey := f.FieldKeyFormatter(key)
		switch v := value.(type) {
		case error:
			data[formattedKey] = v.Error()
		default:
			data[formattedKey] = v
		}
	}

	// Add predefined keys with formatted keys.
	data[f.FieldKeyFormatter(DefaultProdFmtTimestampKey)] = entry.Time.Format(f.TimestampFormat)
	data[f.FieldKeyFormatter(DefaultProdFmtSeverityKey)] = entry.Level.String()
	data[f.FieldKeyFormatter(DefaultProdFmtMessageKey)] = entry.Message

	// Include error message if present.
	if err, ok := entry.Data[DefaultErrorKey]; ok {
		formattedErrorKey := f.FieldKeyFormatter(DefaultProdFmtErrorKey)
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
			data[f.FieldKeyFormatter(DefaultProdFmtTraceIDKey)] = *traceID
		}
		if spanID != nil {
			data[f.FieldKeyFormatter(DefaultProdFmtSpanIDKey)] = *spanID
		}
	}

	// Combine default and custom SkipPackages.
	skipPackages := slice.Union(f.SkipPackages, defaultProdFmtSkipPackages)

	// Caller's function name, file, and line number.
	function, file, line := getCaller(skipPackages)
	if function != "" && file != "" && line != 0 {
		callerInfo := map[string]string{
			f.FieldKeyFormatter(DefaultProdFmtCallerFuncKey): function,
			f.FieldKeyFormatter(DefaultProdFmtCallerFileKey): fmt.Sprintf("%s:%d", file, line),
		}
		data[f.FieldKeyFormatter(DefaultProdFmtCallerKey)] = callerInfo
	}

	// Stack trace for error levels.
	if entry.Level <= logrus.ErrorLevel {
		data[f.FieldKeyFormatter(DefaultProdFmtStackTraceKey)] = getStackTrace()
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
