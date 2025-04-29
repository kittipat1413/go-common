package serverutils_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/kittipat1413/go-common/framework/logger"
	"github.com/kittipat1413/go-common/framework/serverutils"
	"github.com/stretchr/testify/assert"
)

func TestGracefulShutdownSystem(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appLogger := logger.NewNoopLogger()

	// Internal error channel
	errCh := make(chan error, 1)

	// Create flags to confirm shutdown operations were executed
	shutdownExecuted := make(map[string]bool)

	shutdownOps := map[string]serverutils.ShutdownOperation{
		"db": func(ctx context.Context) error {
			shutdownExecuted["db"] = true
			return nil
		},
		"cache": func(ctx context.Context) error {
			shutdownExecuted["cache"] = true
			return nil
		},
	}

	// Start the graceful shutdown system
	done := serverutils.GracefulShutdownSystem(ctx, appLogger, errCh, 5*time.Second, shutdownOps)

	// Simulate sending an interrupt signal
	go func() {
		time.Sleep(100 * time.Millisecond) // small delay to ensure the goroutine is listening
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
	}()

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for graceful shutdown to complete")
	}

	// Verify all shutdown operations were called
	assert.True(t, shutdownExecuted["db"], "db shutdown operation should be executed")
	assert.True(t, shutdownExecuted["cache"], "cache shutdown operation should be executed")
}

func TestGracefulShutdownSystem_ErrorCase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appLogger := logger.NewNoopLogger()

	// Internal error channel
	errCh := make(chan error, 1)

	// Create flags to confirm shutdown operations were executed
	shutdownExecuted := make(map[string]bool)

	shutdownOps := map[string]serverutils.ShutdownOperation{
		"service": func(ctx context.Context) error {
			shutdownExecuted["service"] = true
			return errors.New("shutdown failed")
		},
	}

	// Start the graceful shutdown system
	done := serverutils.GracefulShutdownSystem(ctx, appLogger, errCh, 5*time.Second, shutdownOps)

	// Simulate sending an error instead of OS signal
	go func() {
		time.Sleep(100 * time.Millisecond)
		errCh <- errors.New("internal error")
	}()

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for graceful shutdown to complete")
	}

	assert.True(t, shutdownExecuted["service"], "service shutdown operation should be executed even if it fails")
}

func TestGracefulShutdownSystem_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appLogger := logger.NewNoopLogger()

	// Internal error channel
	errCh := make(chan error, 1)

	// Create flags to confirm shutdown operations were executed
	shutdownExecuted := make(map[string]bool)

	shutdownOps := map[string]serverutils.ShutdownOperation{
		"worker": func(ctx context.Context) error {
			shutdownExecuted["worker"] = true
			return nil
		},
	}

	// Start the graceful shutdown system
	done := serverutils.GracefulShutdownSystem(ctx, appLogger, errCh, 5*time.Second, shutdownOps)

	// Simulate a manual cancellation of the context
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for graceful shutdown to complete")
	}

	assert.True(t, shutdownExecuted["worker"], "worker shutdown operation should be executed on context cancel")
}
