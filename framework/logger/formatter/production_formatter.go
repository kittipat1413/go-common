package formatter

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
}

type ProdFmtKey string

const (
	timestampKey  ProdFmtKey = "timestamp"
	severityKey   ProdFmtKey = "severity"
	messageKey    ProdFmtKey = "message"
	errorKey      ProdFmtKey = "error"
	traceIDKey    ProdFmtKey = "trace_id"
	spanIDKey     ProdFmtKey = "span_id"
	callerKey     ProdFmtKey = "caller"
	callerFuncKey ProdFmtKey = "function"
	callerFileKey ProdFmtKey = "file"
	stackTraceKey ProdFmtKey = "stack_trace"
)

func (k ProdFmtKey) String() string {
	return string(k)
}

var defaultSkipPackages = []string{
	"github.com/sirupsen/logrus",
	"github.com/kittipat1413/go-common/framework/logger",
}

func (f *ProductionFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 7 is the number of additional fields added by this formatter
	data := make(logrus.Fields, len(entry.Data)+7)
	// Copy all fields to the data map
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	data[timestampKey.String()] = entry.Time.Format(f.TimestampFormat)
	data[severityKey.String()] = entry.Level.String()
	data[messageKey.String()] = entry.Message

	// Include error message if present
	if err, ok := entry.Data[logrus.ErrorKey]; ok {
		switch e := err.(type) {
		case error:
			data[errorKey.String()] = e.Error()
		default:
			data[errorKey.String()] = fmt.Sprintf("%v", e)
		}
	}

	// Include trace and span IDs if available
	if entry.Context != nil {
		traceID, spanID := extractTraceIDs(entry.Context)
		if traceID != nil {
			data[traceIDKey.String()] = traceID
		}
		if spanID != nil {
			data[spanIDKey.String()] = spanID
		}
	}

	// Combine default and custom SkipPackages
	skipPackages := slice.Union(f.SkipPackages, defaultSkipPackages)
	// Caller's function name, file, and line number
	function, file, line := getCaller(skipPackages)
	if function != "" && file != "" && line != 0 {
		data[callerKey.String()] = map[string]string{
			callerFuncKey.String(): function,
			callerFileKey.String(): fmt.Sprintf("%s:%d", file, line),
		}
	}

	// Stack trace for error levels
	if entry.Level <= logrus.ErrorLevel {
		data[stackTraceKey.String()] = getStackTrace()
	}

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
	if !span.IsRecording() {
		return nil, nil // No active span
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
	maxBufSize := 32 * 1024 // 64 KB upper limit
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
