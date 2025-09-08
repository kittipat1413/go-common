package trace

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceFunc wraps function execution with OpenTelemetry tracing for automatic observability.
// Creates spans with function names, records execution status, and captures errors with stack traces.
// Uses Go generics to support any return type while maintaining type safety.
//
// Span Configuration:
//   - Span Name: Extracted function name with common suffixes trimmed
//   - Span Kind: Internal (function execution within service)
//   - Error Recording: Stack traces included for failed operations
//   - Status Codes: OK for success, Error for failures with descriptive messages
//
// Function Requirements:
//   - Accept context.Context as first parameter
//   - Return (T, error) where T is any type
//
// Tracer Fallback:
//   - If tracer is nil, uses DefaultTracer() to ensure tracing always works
//     even when not explicitly configured.
//
// Parameters:
//   - ctx: Context for span creation and function execution
//   - tracer: OpenTelemetry tracer (nil uses default tracer)
//   - f: Function to trace following (ctx context.Context) (T, error) pattern
//
// Returns:
//   - T: Function result of generic type T
//   - error: Function error, nil on success
//
// Example Usage:
//
//	user, err := trace.TraceFunc(ctx, tracer, func(ctx context.Context) (*User, error) {
//	    return userRepo.GetByID(ctx, userID)
//	})
func TraceFunc[T any](ctx context.Context, tracer trace.Tracer, f func(ctx context.Context) (T, error)) (T, error) {
	if tracer == nil {
		tracer = DefaultTracer()
	}

	fnName := getFunctionName(f)
	ctx, span := tracer.Start(ctx, fnName, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	result, err := f(ctx)

	if err != nil {
		span.RecordError(err, trace.WithStackTrace(true))
		span.SetStatus(codes.Error, "Error executing function")
	} else {
		span.SetStatus(codes.Ok, "Function executed successfully")
	}

	return result, err
}

func getFunctionName(fn interface{}) string {
	v := reflect.ValueOf(fn)
	if v.Kind() == reflect.Func {
		fnName := runtime.FuncForPC(v.Pointer()).Name()
		// Trim common suffixes like "-fm" (method bound to a receiver) and "func1" (anonymous function)
		fnName = strings.TrimSuffix(fnName, "-fm")
		fnName = strings.TrimSuffix(fnName, ".func1")
		return fnName
	}
	return "unknown"
}
