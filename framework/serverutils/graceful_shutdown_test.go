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

	// Define shutdown tasks
	shutdownTasks := []serverutils.ShutdownTask{
		{
			Name: "db",
			Op: func(ctx context.Context) error {
				shutdownExecuted["db"] = true
				return nil
			},
		},
		{
			Name: "cache",
			Op: func(ctx context.Context) error {
				shutdownExecuted["cache"] = true
				return nil
			},
		},
	}

	// Start the graceful shutdown system
	done := serverutils.GracefulShutdownSystem(ctx, appLogger, errCh, 5*time.Second, shutdownTasks)

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

func TestGracefulShutdownSystem_Order(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appLogger := logger.NewNoopLogger()

	// Internal error channel
	errCh := make(chan error, 1)

	// Create flags to confirm shutdown operations were executed in order
	shutdownOrder := []string{}

	// Define shutdown tasks with specific order
	shutdownTasks := []serverutils.ShutdownTask{
		{
			Name: "first",
			Op: func(ctx context.Context) error {
				shutdownOrder = append(shutdownOrder, "first")
				return nil
			},
		},
		{
			Name: "second",
			Op: func(ctx context.Context) error {
				shutdownOrder = append(shutdownOrder, "second")
				return nil
			},
		},
		{
			Name: "third",
			Op: func(ctx context.Context) error {
				shutdownOrder = append(shutdownOrder, "third")
				return nil
			},
		},
	}

	// Start the graceful shutdown system
	done := serverutils.GracefulShutdownSystem(ctx, appLogger, errCh, 5*time.Second, shutdownTasks)

	// Simulate sending an interrupt signal
	go func() {
		time.Sleep(100 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGINT)
	}()

	select {
	case <-done:
		// success
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for graceful shutdown to complete")
	}

	assert.Equal(t, []string{"first", "second", "third"}, shutdownOrder, "shutdown operations should be executed in the defined order")
}

func TestGracefulShutdownSystem_ErrorCase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appLogger := logger.NewNoopLogger()

	// Internal error channel
	errCh := make(chan error, 1)

	// Create flags to confirm shutdown operations were executed
	shutdownExecuted := make(map[string]bool)

	// Define shutdown tasks
	shutdownTasks := []serverutils.ShutdownTask{
		{
			Name: "service",
			Op: func(ctx context.Context) error {
				shutdownExecuted["service"] = true
				return errors.New("shutdown failed")
			},
		},
	}

	// Start the graceful shutdown system
	done := serverutils.GracefulShutdownSystem(ctx, appLogger, errCh, 5*time.Second, shutdownTasks)

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

	// Define shutdown tasks
	shutdownTasks := []serverutils.ShutdownTask{
		{
			Name: "worker",
			Op: func(ctx context.Context) error {
				shutdownExecuted["worker"] = true
				return nil
			},
		},
	}

	// Start the graceful shutdown system
	done := serverutils.GracefulShutdownSystem(ctx, appLogger, errCh, 5*time.Second, shutdownTasks)

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
