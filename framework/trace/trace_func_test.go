package trace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kittipat1413/go-common/framework/trace"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTraceFunc_Success(t *testing.T) {
	ctx := context.Background()

	// Set up the SpanRecorder and TracerProvider
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	tracer := tracerProvider.Tracer("test-tracer")

	// Function that returns a successful result
	testFunc := func(ctx context.Context) (string, error) {
		return "success", nil
	}

	result, err := trace.TraceFunc(ctx, tracer, testFunc)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// Retrieve recorded spans and validate
	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, codes.Ok, span.Status().Code, "Expected span status code to be Ok")
}

func TestTraceFunc_Error(t *testing.T) {
	ctx := context.Background()

	// Set up the SpanRecorder and TracerProvider
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	tracer := tracerProvider.Tracer("test-tracer")

	// Function that returns an error
	testFunc := func(ctx context.Context) (string, error) {
		return "", errors.New("test error")
	}

	result, err := trace.TraceFunc(ctx, tracer, testFunc)

	assert.Error(t, err)
	assert.Equal(t, "", result)

	// Retrieve recorded spans and validate
	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, codes.Error, span.Status().Code, "Expected span status code to be Error")
	assert.Equal(t, "Error executing function", span.Status().Description)
	assert.NotEmpty(t, span.Events(), "Expected error events to be recorded")
}

func TestTraceFunc_DefaultTracer(t *testing.T) {
	ctx := context.Background()

	// Function that returns a successful result
	testFunc := func(ctx context.Context) (string, error) {
		return "default-tracer-success", nil
	}

	// Call TraceFunc without specifying a tracer, so it defaults
	result, err := trace.TraceFunc(ctx, nil, testFunc)

	assert.NoError(t, err)
	assert.Equal(t, "default-tracer-success", result)
}
