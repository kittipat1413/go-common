package callbackhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/kittipat1413/go-common/framework/event"
	common_logger "github.com/kittipat1413/go-common/framework/logger"
)

/*
	[]TODO if client is nil; use default http.Client from another common package
*/

const (
	defaultCallbackTimeout = 60 * time.Second
	defaultMaxRetries      = 3 // initial attempt + 3 retries
	defaultRetryInterval   = 2 * time.Second
)

// callbackConfig contains the configuration for sending callbacks
type callbackConfig struct {
	maxRetries      int
	retryInterval   time.Duration
	callbackTimeout time.Duration
}

// callbackEventHandler is an EventHandler that sends callbacks based on the success or failure of an event
type callbackEventHandler[T any] struct {
	httpClient     *http.Client
	callbackConfig callbackConfig
	logger         common_logger.Logger
}

/*
NewEventHandler creates a new event.EventHandler that sends callbacks based on the success or failure of an event.
  - This handler supports internal logging; if a custom logger is provided via WithLogger, it will be used; otherwise, it will try to extract a logger from the context, and if none is found, a default logger will be used.
*/
func NewEventHandler[T any](opts ...Option[T]) event.EventHandler[T] {
	// Set default values
	handler := &callbackEventHandler[T]{
		httpClient: http.DefaultClient,
		callbackConfig: callbackConfig{
			maxRetries:      defaultMaxRetries,
			retryInterval:   defaultRetryInterval,
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

// Option is a function that configures an EventHandler
type Option[T any] func(*callbackEventHandler[T])

// WithHTTPClient sets a custom http.Client
func WithHTTPClient[T any](client *http.Client) Option[T] {
	return func(eh *callbackEventHandler[T]) {
		if client != nil {
			eh.httpClient = client
		}
	}
}

// WithCallbackConfig sets the callback configuration
func WithCallbackConfig[T any](maxRetries int, retryInterval, callbackTimeout time.Duration) Option[T] {
	return func(eh *callbackEventHandler[T]) {
		eh.callbackConfig = callbackConfig{
			maxRetries:      maxRetries,
			retryInterval:   retryInterval,
			callbackTimeout: callbackTimeout,
		}
	}
}

// WithLogger sets a custom logger.Logger implementation for the EventHandler. If not provided, the logger will be extracted from the context or a default will be used.
func WithLogger[T any](customLogger common_logger.Logger) Option[T] {
	return func(eh *callbackEventHandler[T]) {
		if customLogger != nil {
			eh.logger = customLogger
		}
	}
}

// UnmarshalEventMessage unmarshals the provided JSON data into a CallbackEventMessage.
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

// AfterHandle performs post-processing after the event has been handled, including triggering
// callbacks based on the result of the event handling.
func (eh *callbackEventHandler[T]) AfterHandle(ctx context.Context, msg event.EventMessage[T], eventResult error) error {
	// Check if the event message includes callback information and handle the callback accordingly
	if callbackMsg, ok := msg.(interface{ GetCallback() *CallbackInfo }); ok {
		eh.handleCallback(ctx, eventResult, callbackMsg.GetCallback())
	}
	return nil
}

// HandleCallback handles the success and failure callback logic using the EventHandler's http.Client
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

// sendCallback sends a callback using the EventHandler's http.Client with retry logic
func (eh *callbackEventHandler[T]) sendCallback(ctx context.Context, url string) {
	var resp *http.Response
	var err error

	logger := eh.logger
	if logger == nil {
		logger = common_logger.FromContext(ctx)
	}
	for attempt := 0; attempt <= eh.callbackConfig.maxRetries; attempt++ {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			logger.Error(ctx, fmt.Sprintf("Failed to create request for callback to: %s", url), reqErr, nil)
			return
		}

		resp, err = eh.httpClient.Do(req)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Attempt %d: Failed to send callback to: %s", attempt+1, url), err, nil)
		} else {
			// Read and discard the response body
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("Attempt %d: Failed to read response body for callback to: %s", attempt+1, url), err, nil)
				return
			}
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				logger.Info(ctx, fmt.Sprintf("Callback succeeded with status: %s", resp.Status), nil)
				return // success, exit the function
			} else if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				// Server error, can retry
				logger.Info(ctx, fmt.Sprintf("Attempt %d: Server error for callback to %s: %s", attempt+1, url, resp.Status), nil)
			} else {
				// Client error or other non-retryable status
				logger.Info(ctx, fmt.Sprintf("Callback failed with status: %s", resp.Status), nil)
				return
			}
		}

		if attempt < eh.callbackConfig.maxRetries {
			// Exponential backoff
			sleepDuration := eh.callbackConfig.retryInterval * (1 << attempt)
			jitter := time.Duration(float64(sleepDuration) * 0.1 * (0.5 - rand.Float64())) // #nosec G404
			sleepDuration += jitter

			select {
			case <-ctx.Done():
				logger.Info(ctx, fmt.Sprintf("Context canceled, aborting retries for callback to: %s", url), nil)
				return
			case <-time.After(sleepDuration):
				// Proceed to next attempt
			}
		}
	}

	logger.Info(ctx, fmt.Sprintf("All retries failed for callback to: %s", url), nil)
}
