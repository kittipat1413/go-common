package serverutils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kittipat1413/go-common/framework/logger"
)

// ShutdownOperation defines a function type for shutdown operations
type ShutdownOperation func(ctx context.Context) error

// GracefulShutdownSystem waits for OS signal or error then runs shutdown hooks
func GracefulShutdownSystem(
	ctx context.Context,
	appLogger logger.Logger, // Logger for logging
	errCh <-chan error, // Channel for internal errors
	timeout time.Duration, // Timeout for graceful shutdown
	shutdownOps map[string]ShutdownOperation, // Map of shutdown operation names to shutdown operation
) <-chan struct{} {
	if appLogger == nil {
		appLogger = logger.FromContext(ctx)
	}

	done := make(chan struct{})
	// Start a goroutine to handle shutdown
	go func() {
		defer close(done)

		// Wait for signal or error
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-quit:
			appLogger.Warn(ctx, fmt.Sprintf("[shutdown] received OS signal: %s", sig), nil)
		case err := <-errCh:
			appLogger.Warn(ctx, fmt.Sprintf("[shutdown] received error: %v", err), nil)
		case <-ctx.Done():
			appLogger.Warn(ctx, "[shutdown] context done", nil)
		}

		// Context for shutdown operations
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Run each operation
		for name, op := range shutdownOps {
			appLogger.Warn(ctx, fmt.Sprintf("[shutdown] running %s", name), nil)
			if err := op(shutdownCtx); err != nil {
				appLogger.Error(ctx, fmt.Sprintf("[shutdown] %s failed", name), err, nil)
			} else {
				appLogger.Warn(ctx, fmt.Sprintf("[shutdown] %s completed", name), nil)
			}
		}
	}()

	return done
}
