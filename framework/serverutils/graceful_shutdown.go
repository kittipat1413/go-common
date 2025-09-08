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

// ShutdownTask represents a cleanup operation to execute during graceful shutdown.
// Tasks are executed sequentially in the order they are provided to ensure
// proper dependency management and resource cleanup ordering.
type ShutdownTask struct {
	Name string            // Human-readable task name for logging and debugging
	Op   ShutdownOperation // Cleanup operation to execute
}

// ShutdownOperation defines a cleanup function executed during graceful shutdown.
// Should handle context cancellation and return quickly to avoid blocking shutdown.
//
// Parameters:
//   - ctx: Context with shutdown timeout for cancellation control
//
// Returns:
//   - error: nil on success, error details on failure (logged but doesn't stop shutdown)
//
// Example:
//
//	func closeDatabase(ctx context.Context) error {
//	    return db.Close()
//	}
type ShutdownOperation func(ctx context.Context) error

// GracefulShutdownSystem orchestrates graceful application shutdown with configurable cleanup tasks.
// Monitors for OS signals (SIGINT/SIGTERM), internal errors, or context cancellation, then executes
// shutdown tasks sequentially within the specified timeout to ensure clean resource cleanup.
//
// Shutdown Triggers:
//   - OS Signals: SIGINT (Ctrl+C), SIGTERM (container/process manager)
//   - Internal Errors: Application errors sent via errCh channel
//   - Context Cancellation: Parent context cancellation
//
// Shutdown Process:
//  1. Wait for shutdown trigger (signal, error, or context cancellation)
//  2. Log shutdown initiation with trigger details
//  3. Create timeout context for all cleanup operations
//  4. Execute shutdown tasks sequentially in provided order
//  5. Log each task's completion or failure (failures don't stop shutdown)
//  6. Close done channel to signal shutdown completion
//
// Example Usage:
//
//	// Define cleanup tasks in dependency order
//	shutdownTasks := []serverutils.ShutdownTask{
//	    {Name: "HTTP Server", Op: func(ctx context.Context) error {
//	        return server.Shutdown(ctx)
//	    }},
//	    {Name: "Database", Op: func(ctx context.Context) error {
//	        return db.Close()
//	    }},
//	    {Name: "Cache", Op: func(ctx context.Context) error {
//	        return cache.Close()
//	    }},
//	}
//
//	// Start graceful shutdown monitoring
//	errCh := make(chan error, 1)
//	done := serverutils.GracefulShutdownSystem(
//	    ctx, logger, errCh, 30*time.Second, shutdownTasks,
//	)
//
//	// Application runs here...
//
//	// Wait for shutdown completion
//	<-done
//	logger.Info(ctx, "Application shutdown completed", nil)
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
