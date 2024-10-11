package trace

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceFunc traces the start and end of a function and returns its result.
// It works with any function that returns a result and an error.
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
