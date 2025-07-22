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

// ShutdownTask represents a task to be executed during graceful shutdown
type ShutdownTask struct {
	Name string
	Op   ShutdownOperation
}

// ShutdownOperation defines a function type for shutdown operations
type ShutdownOperation func(ctx context.Context) error

// GracefulShutdownSystem waits for OS signal or error then runs provided shutdown tasks in the given order
// within the specified timeout. It returns a channel that's closed once all tasks complete or the timeout elapses.
func GracefulShutdownSystem(
	ctx context.Context,
	appLogger logger.Logger, // Logger for logging
	errCh <-chan error, // Channel for internal errors
	timeout time.Duration, // Timeout for graceful shutdown
	shutdownTasks []ShutdownTask, // Ordered list of shutdown tasks
) <-chan struct{} {
	if appLogger == nil {
		appLogger = logger.FromContext(ctx)
	}

	// done channel to signal completion
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

		// Execute operations in defined order
		for _, task := range shutdownTasks {
			appLogger.Warn(ctx, fmt.Sprintf("[shutdown] running %s", task.Name), nil)
			if err := task.Op(shutdownCtx); err != nil {
				appLogger.Error(ctx, fmt.Sprintf("[shutdown] %s failed", task.Name), err, nil)
			} else {
				appLogger.Warn(ctx, fmt.Sprintf("[shutdown] %s completed", task.Name), nil)
			}
		}
	}()

	return done
}
