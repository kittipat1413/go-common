package callbackhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kittipat1413/go-common/framework/event"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
	"github.com/kittipat1413/go-common/framework/retry"
)

/*
	[]TODO if client is nil; use default http.Client from another common package
*/

// Default configuration values for callback behavior
const (
	defaultCallbackTimeout = 60 * time.Second // Maximum time for callback request
	defaultMaxRetries      = 3                // Initial attempt + 3 retries
	defaultRetryInterval   = 2 * time.Second  // Delay between retry attempts
)

// callbackConfig holds configuration for HTTP callback behavior including
// retry strategy and timeout settings.
type callbackConfig struct {
	retrier         retry.Retrier // Retry strategy for failed callbacks
	callbackTimeout time.Duration // Timeout for individual callback requests
}

// callbackEventHandler implements EventHandler[T] with HTTP callback functionality.
// Sends success or failure callbacks based on event processing results with
// configurable retry behavior and timeout handling.
type callbackEventHandler[T any] struct {
	httpClient     *http.Client         // HTTP client for callback requests
	callbackConfig callbackConfig       // Callback retry and timeout configuration
	logger         common_logger.Logger // Logger for callback operations
}

// NewEventHandler creates a callback-enabled event handler that sends HTTP callbacks
// based on event processing results.
//
// The handler supports custom logger configuration; if none provided, it extracts
// logger from context or uses default logging.
//
// Parameters:
//   - opts: Configuration options to customize behavior
//
// Returns:
//   - event.EventHandler[T]: Configured callback event handler
//
// Example:
//
//	handler := NewEventHandler[UserData](
//	    WithHTTPClient(customClient),
//	    WithCallbackConfig(5, 3*time.Second, 30*time.Second),
//	    WithLogger(customLogger),
//	)
func NewEventHandler[T any](opts ...Option[T]) event.EventHandler[T] {
	// Create default fixed backoff strategy for callbacks
	backoffStrategy, _ := retry.NewFixedBackoffStrategy(
		defaultRetryInterval,
	)
	// Create default retrier
	defaultRetrier, _ := retry.NewRetrier(retry.Config{
		MaxAttempts: defaultMaxRetries,
		Backoff:     backoffStrategy,
	})

	// Set default values
	handler := &callbackEventHandler[T]{
		httpClient: http.DefaultClient,
		callbackConfig: callbackConfig{
			retrier:         defaultRetrier,
			callbackTimeout: defaultCallbackTimeout,
		},
		logger: nil,
	}

	// Apply options
	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

// Option is a function type for configuring the callback event handler.
type Option[T any] func(*callbackEventHandler[T])

// WithHTTPClient configures a custom HTTP client for callback requests.
// Use to customize timeout, transport, or authentication settings.
func WithHTTPClient[T any](client *http.Client) Option[T] {
	return func(eh *callbackEventHandler[T]) {
		if client != nil {
			eh.httpClient = client
		}
	}
}

// WithCallbackConfig configures retry behavior and timeouts for callbacks.
// Allows customization of retry attempts, intervals, and request timeouts.
func WithCallbackConfig[T any](maxRetries int, retryInterval time.Duration, callbackTimeout time.Duration) Option[T] {
	return func(eh *callbackEventHandler[T]) {
		// Create fixed backoff strategy
		backoffStrategy, err := retry.NewFixedBackoffStrategy(
			retryInterval,
		)
		if err != nil {
			return // Fallback to default if creation fails
		}

		// Create retrier with custom config
		retrier, err := retry.NewRetrier(retry.Config{
			MaxAttempts: maxRetries,
			Backoff:     backoffStrategy,
		})
		if err != nil {
			return // Fallback to default if creation fails
		}

		eh.callbackConfig = callbackConfig{
			retrier:         retrier,
			callbackTimeout: callbackTimeout,
		}
	}
}

// WithLogger configures a custom logger for the event handler.
// If not provided, logger will be extracted from context or use default.
func WithLogger[T any](customLogger common_logger.Logger) Option[T] {
	return func(eh *callbackEventHandler[T]) {
		if customLogger != nil {
			eh.logger = customLogger
		}
	}
}

// UnmarshalEventMessage unmarshals JSON data into a CallbackEventMessage.
// Handles conversion from raw JSON to typed event message with callback configuration.
//
// Expected JSON format includes callback URLs and event data:
//
//	{
//	    "event_type": "user.created",
//	    "payload": {...},
//	    "callback": {
//	        "success_url": "https://api.example.com/success",
//	        "fail_url": "https://api.example.com/failure"
//	    }
//	}
//
// Parameters:
//   - data: Raw JSON event data
//
// Returns:
//   - event.EventMessage[T]: Typed event message with callback info
//   - error: JSON unmarshaling error if data invalid
func (eh *callbackEventHandler[T]) UnmarshalEventMessage(data []byte) (event.EventMessage[T], error) {
	var msg CallbackEventMessage[T]
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event message: %w", err)
	}
	return &msg, nil
}

// BeforeHandle performs any necessary pre-processing before the event is handled.
// This can be used to validate or transform the event message, or prepare any resources needed.
func (eh *callbackEventHandler[T]) BeforeHandle(ctx context.Context, msg event.EventMessage[T]) error {
	return nil
}

// AfterHandle performs post-processing including sending success/failure callbacks.
// Runs after main event processing and sends appropriate HTTP callbacks based on
// processing results. Callbacks are sent asynchronously with retry logic.
func (eh *callbackEventHandler[T]) AfterHandle(ctx context.Context, msg event.EventMessage[T], eventResult error) error {
	// Check if the event message includes callback information and handle the callback accordingly
	if callbackMsg, ok := msg.(interface{ GetCallback() *CallbackInfo }); ok {
		eh.handleCallback(ctx, eventResult, callbackMsg.GetCallback())
	}
	return nil
}

// handleCallback manages callback delivery logic based on event processing results.
// Determines which callback URL to use and initiates asynchronous delivery.
func (eh *callbackEventHandler[T]) handleCallback(ctx context.Context, err error, callback *CallbackInfo) {
	if callback == nil {
		return
	}

	copyCtx := context.WithoutCancel(ctx)
	if err != nil {
		if callback.FailURL != "" {
			go func() {
				ctx, cancel := context.WithTimeout(copyCtx, eh.callbackConfig.callbackTimeout)
				defer cancel()
				eh.sendCallback(ctx, callback.FailURL)
			}()
		}
	} else {
		if callback.SuccessURL != "" {
			go func() {
				ctx, cancel := context.WithTimeout(copyCtx, eh.callbackConfig.callbackTimeout)
				defer cancel()
				eh.sendCallback(ctx, callback.SuccessURL)
			}()
		}
	}
}

// sendCallback delivers HTTP callback with retry logic and error handling.
// Uses configured retry strategy with proper error classification for retry decisions.
func (eh *callbackEventHandler[T]) sendCallback(ctx context.Context, url string) {
	logger := eh.logger
	if logger == nil {
		logger = common_logger.FromContext(ctx)
	}

	err := eh.callbackConfig.retrier.ExecuteWithRetry(ctx,
		func(ctx context.Context) error {
			return eh.makeCallbackRequest(ctx, url, logger)
		},
		func(attempt int, err error) bool {
			if httpErr, ok := err.(*CallbackHTTPError); ok {
				isRetryable := httpErr.StatusCode >= 500 && httpErr.StatusCode < 600
				if !isRetryable {
					logger.Info(ctx, fmt.Sprintf("Non-retryable error for callback to %s: %s", url, httpErr.Error()), nil)
				} else {
					logger.Info(ctx, fmt.Sprintf("Attempt %d: Server error for callback to %s: %s", attempt, url, httpErr.Error()), nil)
				}
				return isRetryable
			}
			// Non-retryable if it's not an HTTP error
			logger.Error(ctx, fmt.Sprintf("Failed to send callback to %s: %s", url, err.Error()), err, nil)
			return false
		},
	)

	if err != nil {
		logger.Info(ctx, fmt.Sprintf("All retries failed for callback to: %s", url), nil)
	}
}

// CallbackHTTPError represents HTTP-specific errors from callback requests.
// Enables proper error classification for retry decisions.
type CallbackHTTPError struct {
	StatusCode int    // HTTP status code from response
	Status     string // HTTP status text
	URL        string // Target URL that failed
}

// Error implements the error interface for CallbackHTTPError.
func (e *CallbackHTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s for URL %s", e.StatusCode, e.Status, e.URL)
}

// makeCallbackRequest performs a single HTTP GET callback request.
// Handles HTTP communication with proper error handling and response consumption.
func (eh *callbackEventHandler[T]) makeCallbackRequest(ctx context.Context, url string, logger common_logger.Logger) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := eh.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and discard the response body
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Info(ctx, fmt.Sprintf("Callback succeeded with status: %s", resp.Status), nil)
		return nil // Success
	}

	// Return HTTP error for non-2xx responses
	return &CallbackHTTPError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		URL:        url,
	}
}
