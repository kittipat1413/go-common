package event

import (
	"net/http"
	"time"
)

/*
	TODO if client is nil; use default http.Client from another common package
*/

// EventHandler is a struct that holds shared dependencies like http.Client
type EventHandler struct {
	httpClient     *http.Client
	callbackConfig callbackConfig
}

type callbackConfig struct {
	maxRetries      int
	retryInterval   time.Duration
	callbackTimeout time.Duration
}

// NewEventHandler creates a new EventHandler with the provided options
func NewEventHandler(opts ...Option) *EventHandler {
	// Set default values
	eh := &EventHandler{
		httpClient: http.DefaultClient,
		callbackConfig: callbackConfig{
			maxRetries:      defaultMaxRetries,
			retryInterval:   defaultRetryInterval,
			callbackTimeout: defaultCallbackTimeout,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(eh)
	}

	return eh
}

// Option is a function that configures an EventHandler
type Option func(*EventHandler)

// WithHTTPClient sets a custom http.Client
func WithHTTPClient(client *http.Client) Option {
	return func(eh *EventHandler) {
		if client != nil {
			eh.httpClient = client
		}
	}
}

// WithCallbackConfig sets the callback configuration
func WithCallbackConfig(maxRetries int, retryInterval, callbackTimeout time.Duration) Option {
	return func(eh *EventHandler) {
		eh.callbackConfig = callbackConfig{
			maxRetries:      maxRetries,
			retryInterval:   retryInterval,
			callbackTimeout: callbackTimeout,
		}
	}
}
