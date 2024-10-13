package logger_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/stretchr/testify/assert"
)

func TestFromContextWithNoLogger(t *testing.T) {
	// Create a context without a logger
	ctx := context.Background()

	// Retrieve the logger from context and check it's not nil
	retrievedLogger := logger.FromContext(ctx)
	assert.NotNil(t, retrievedLogger)
}

func TestFromContextWithLogger(t *testing.T) {
	// Create a context with a specific logger
	ctx := context.Background()
	specificLogger := logger.NewDefaultLogger()
	ctxWithLogger := logger.NewContext(ctx, specificLogger)

	// Retrieve the logger from context and check it's the correct one
	retrievedLogger := logger.FromContext(ctxWithLogger)
	assert.NotNil(t, retrievedLogger)
	assert.Equal(t, specificLogger, retrievedLogger)
}

func TestFromRequest(t *testing.T) {
	// Create a new HTTP request
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// Attach the DefaultLogger to the request's context
	defaultLogger := logger.NewDefaultLogger()
	reqWithLogger := logger.NewRequest(req, defaultLogger)

	// Retrieve the logger from the request's context and verify it
	retrievedLogger := logger.FromRequest(reqWithLogger)
	assert.NotNil(t, retrievedLogger)
	assert.Equal(t, defaultLogger, retrievedLogger)
}
