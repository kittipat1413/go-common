package callbackhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kittipat1413/go-common/framework/event"
	callbackhandler "github.com/kittipat1413/go-common/framework/event/custom_handler/callback"
)

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestAfterHandle_SuccessCallback(t *testing.T) {
	// Define a sample payload
	type SamplePayload struct {
		Data string `json:"data"`
	}

	// Create a sample event message with callback info
	msg := &callbackhandler.CallbackEventMessage[SamplePayload]{
		BaseEventMessage: event.BaseEventMessage[SamplePayload]{
			EventType: "test_event",
			Timestamp: time.Now(),
			Payload:   SamplePayload{Data: "test data"},
			Metadata:  map[string]string{"version": "1.0"},
		},
		Callback: &callbackhandler.CallbackInfo{
			SuccessURL: "http://example.com/success",
			FailURL:    "http://example.com/fail",
		},
	}

	// Create a custom HTTP client with a RoundTripper
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test that the request URL is the SuccessURL
		assert.Equal(t, msg.Callback.SuccessURL, req.URL.String())
		assert.Equal(t, "GET", req.Method)

		// Simulate a successful response
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("OK")),
			Header:     make(http.Header),
		}
	})

	// Create the handler with the custom HTTP client
	handler := callbackhandler.NewEventHandler(
		callbackhandler.WithHTTPClient[SamplePayload](client),
	)

	// Call AfterHandle with no error to simulate success
	err := handler.AfterHandle(context.Background(), msg, nil)
	require.NoError(t, err)

	// Wait briefly to allow the goroutine to run
	time.Sleep(100 * time.Millisecond)
}

func TestAfterHandle_FailureCallback(t *testing.T) {
	// Define a sample payload
	type SamplePayload struct {
		Data string `json:"data"`
	}

	// Create a sample event message with callback info
	msg := &callbackhandler.CallbackEventMessage[SamplePayload]{
		BaseEventMessage: event.BaseEventMessage[SamplePayload]{
			EventType: "test_event",
			Timestamp: time.Now(),
			Payload:   SamplePayload{Data: "test data"},
			Metadata:  map[string]string{"version": "1.0"},
		},
		Callback: &callbackhandler.CallbackInfo{
			SuccessURL: "http://example.com/success",
			FailURL:    "http://example.com/fail",
		},
	}

	// Create a custom HTTP client with a RoundTripper
	client := NewTestClient(func(req *http.Request) *http.Response {
		// Test that the request URL is the FailURL
		assert.Equal(t, msg.Callback.FailURL, req.URL.String())
		assert.Equal(t, "GET", req.Method)

		// Simulate a successful response
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("OK")),
			Header:     make(http.Header),
		}
	})

	// Create the handler with the custom HTTP client
	handler := callbackhandler.NewEventHandler(
		callbackhandler.WithHTTPClient[SamplePayload](client),
	)

	// Call AfterHandle with an error to simulate failure
	err := handler.AfterHandle(context.Background(), msg, errors.New("processing error"))
	require.NoError(t, err)

	// Wait briefly to allow the goroutine to run
	time.Sleep(100 * time.Millisecond)
}

func TestAfterHandle_RetriesOnServerError(t *testing.T) {
	// Counter to keep track of attempts
	var attempt int32

	// Number of retries
	maxRetries := 3

	// Create a custom HTTP client with a RoundTripper
	client := NewTestClient(func(req *http.Request) *http.Response {
		atomic.AddInt32(&attempt, 1)
		// Simulate a server error response for each attempt
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("Server Error")),
			Header:     make(http.Header),
		}
	})

	// Define a sample payload
	type SamplePayload struct {
		Data string `json:"data"`
	}

	// Create a sample event message with callback info
	msg := &callbackhandler.CallbackEventMessage[SamplePayload]{
		BaseEventMessage: event.BaseEventMessage[SamplePayload]{
			EventType: "test_event",
			Timestamp: time.Now(),
			Payload:   SamplePayload{Data: "test data"},
			Metadata:  map[string]string{"version": "1.0"},
		},
		Callback: &callbackhandler.CallbackInfo{
			SuccessURL: "http://example.com/success",
			FailURL:    "http://example.com/fail",
		},
	}

	// Create the handler with the custom HTTP client
	handler := callbackhandler.NewEventHandler(
		callbackhandler.WithHTTPClient[SamplePayload](client),
		callbackhandler.WithCallbackConfig[SamplePayload](maxRetries, 1*time.Millisecond, 1*time.Minute),
	)

	// Call AfterHandle with an error to simulate failure, which triggers the FailURL callback
	err := handler.AfterHandle(context.Background(), msg, errors.New("processing error"))
	require.NoError(t, err)

	// Wait for the retries to complete
	time.Sleep(500 * time.Millisecond)

	// Atomically load the attempt counter
	totalAttempts := atomic.LoadInt32(&attempt)
	// Assert that the number of attempts is maxRetries + 1 (initial attempt + retries)
	require.Equal(t, int32(maxRetries+1), totalAttempts)
}

func TestAfterHandle_NoRetriesOnBadRequest(t *testing.T) {
	// Counter to keep track of attempts
	var attempt int32

	// Number of retries
	maxRetries := 3

	// Create a custom HTTP client with a RoundTripper
	client := NewTestClient(func(req *http.Request) *http.Response {
		atomic.AddInt32(&attempt, 1)
		// Simulate a server error response for each attempt
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString("bad request")),
			Header:     make(http.Header),
		}
	})

	// Define a sample payload
	type SamplePayload struct {
		Data string `json:"data"`
	}

	// Create a sample event message with callback info
	msg := &callbackhandler.CallbackEventMessage[SamplePayload]{
		BaseEventMessage: event.BaseEventMessage[SamplePayload]{
			EventType: "test_event",
			Timestamp: time.Now(),
			Payload:   SamplePayload{Data: "test data"},
			Metadata:  map[string]string{"version": "1.0"},
		},
		Callback: &callbackhandler.CallbackInfo{
			SuccessURL: "http://example.com/success",
			FailURL:    "http://example.com/fail",
		},
	}

	// Create the handler with the custom HTTP client
	handler := callbackhandler.NewEventHandler(
		callbackhandler.WithHTTPClient[SamplePayload](client),
		callbackhandler.WithCallbackConfig[SamplePayload](maxRetries, 1*time.Millisecond, 1*time.Minute),
	)

	// Call AfterHandle with an error to simulate failure, which triggers the FailURL callback
	err := handler.AfterHandle(context.Background(), msg, errors.New("processing error"))
	require.NoError(t, err)

	// Wait for the retries to complete
	time.Sleep(500 * time.Millisecond)

	// Atomically load the attempt counter
	totalAttempts := atomic.LoadInt32(&attempt)
	// Assert that the number of attempts is 1 (initial attempt only)
	require.Equal(t, int32(1), totalAttempts)
}

func TestUnmarshalEventMessage_Success(t *testing.T) {
	type SamplePayload struct {
		Data string `json:"data"`
	}

	msg := callbackhandler.CallbackEventMessage[SamplePayload]{
		BaseEventMessage: event.BaseEventMessage[SamplePayload]{
			EventType: "test_event",
			Timestamp: time.Now().UTC(),
			Payload:   SamplePayload{Data: "test data"},
			Metadata:  map[string]string{"version": "1.0"},
		},
		Callback: &callbackhandler.CallbackInfo{
			SuccessURL: "http://example.com/success",
			FailURL:    "http://example.com/fail",
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	handler := callbackhandler.NewEventHandler[SamplePayload]()

	unmarshalledMsg, err := handler.UnmarshalEventMessage(data)
	require.NoError(t, err)
	require.NotNil(t, unmarshalledMsg)

	assert.Equal(t, msg.GetVersion(), unmarshalledMsg.GetVersion())
	assert.Equal(t, msg.GetEventType(), unmarshalledMsg.GetEventType())
	assert.Equal(t, msg.GetTimestamp(), unmarshalledMsg.GetTimestamp())
	assert.Equal(t, msg.GetPayload(), unmarshalledMsg.GetPayload())
	assert.Equal(t, msg.GetMetadata(), unmarshalledMsg.GetMetadata())
}

func TestUnmarshalEventMessage_Error(t *testing.T) {
	data := []byte(`{"invalid_json":`)

	handler := callbackhandler.NewEventHandler[interface{}]()

	_, err := handler.UnmarshalEventMessage(data)
	require.Error(t, err)
}
