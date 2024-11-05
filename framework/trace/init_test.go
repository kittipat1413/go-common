package trace_test

import (
	"context"
	"testing"

	"github.com/kittipat1413/go-common/framework/trace"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestInitTracerProvider_StdoutExporter(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service-stdout"
	exporterType := trace.ExporterStdout

	tracerProvider, err := trace.InitTracerProvider(ctx, serviceName, nil, exporterType)
	defer func() {
		if tracerProvider != nil {
			_ = tracerProvider.Shutdown(ctx)
		}
	}()

	assert.NoError(t, err)
	assert.NotNil(t, tracerProvider)
	assert.Equal(t, tracerProvider, otel.GetTracerProvider())
}

func TestInitTracerProvider_GRPCExporter(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service-grpc"
	exporterType := trace.ExporterGRPC
	endpoint := "localhost:4317"

	tracerProvider, err := trace.InitTracerProvider(ctx, serviceName, &endpoint, exporterType)
	defer func() {
		if tracerProvider != nil {
			_ = tracerProvider.Shutdown(ctx)
		}
	}()

	assert.NoError(t, err)
	assert.NotNil(t, tracerProvider)
	assert.Equal(t, tracerProvider, otel.GetTracerProvider())
}

func TestInitTracerProvider_InvalidExporter(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service-invalid"
	exporterType := trace.ExporterType("invalid-exporter")

	tracerProvider, err := trace.InitTracerProvider(ctx, serviceName, nil, exporterType)

	assert.Error(t, err)
	assert.Nil(t, tracerProvider)
	assert.Contains(t, err.Error(), "unsupported exporter type")
}
