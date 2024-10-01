package httphandler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kittipat1413/go-common/framework/event"
	httphandler "github.com/kittipat1413/go-common/framework/event/http_handler"
	eventmocks "github.com/kittipat1413/go-common/framework/event/mocks" // Adjust the import path as necessary
)

// Define a sample payload type
type SamplePayload struct {
	Data string `json:"data"`
}

func TestNewGinEventHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a sample event message
	timestamp := time.Now().UTC()
	metadata := map[string]string{
		event.MetadataKeyVersion: "1.0",
		event.MetadataKeySource:  "unit_test",
	}
	payload := SamplePayload{Data: "test data"}

	msg := &event.BaseEventMessage[SamplePayload]{
		EventType: "test_event",
		Timestamp: timestamp,
		Payload:   payload,
		Metadata:  metadata,
	}

	// Serialize the message to JSON
	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock EventHandler
	mockEventHandler := eventmocks.NewMockEventHandler[SamplePayload](ctrl)

	// Set up expectations
	mockEventHandler.EXPECT().
		UnmarshalEventMessage(gomock.Any()).
		Return(msg, nil)

	mockEventHandler.EXPECT().
		BeforeHandle(gomock.Any(), msg).
		Return(nil)

	mockEventHandler.EXPECT().
		AfterHandle(gomock.Any(), msg, nil).
		Return(nil)

	// Define a test business logic handler
	testGinHandler := func(ctx *gin.Context, msg event.EventMessage[SamplePayload]) error {
		// Business logic here
		return nil // Simulate success
	}

	// Create the Gin handler
	ginHandler := httphandler.NewGinEventHandler(testGinHandler, mockEventHandler)

	// Create a test HTTP request
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(data))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create a test context and recorder
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// Set the request in the context
	ctx.Request = req

	// Call the handler
	ginHandler(ctx)

	// Assert the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "success"}`, w.Body.String())
}

func TestNewGinEventHandler_UnmarshalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Invalid JSON data
	data := []byte(`{"invalid_json":`)

	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock EventHandler
	mockEventHandler := eventmocks.NewMockEventHandler[SamplePayload](ctrl)

	// Set up expectations
	mockEventHandler.EXPECT().
		UnmarshalEventMessage(gomock.Any()).
		Return(nil, errors.New("unmarshal error"))

	// Define a test business logic handler
	testGinHandler := func(ctx *gin.Context, msg event.EventMessage[SamplePayload]) error {
		return nil
	}

	// Create the Gin handler
	ginHandler := httphandler.NewGinEventHandler(testGinHandler, mockEventHandler)

	// Create a test HTTP request
	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(data))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Create a test context and recorder
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	// Call the handler
	ginHandler(ctx)

	// Assert the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unmarshal error")
}

func TestNewGinEventHandler_BusinessLogicError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a sample event message
	timestamp := time.Now().UTC()
	metadata := map[string]string{
		event.MetadataKeyVersion: "1.0",
		event.MetadataKeySource:  "unit_test",
	}
	payload := SamplePayload{Data: "test data"}

	msg := &event.BaseEventMessage[SamplePayload]{
		EventType: "test_event",
		Timestamp: timestamp,
		Payload:   payload,
		Metadata:  metadata,
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	// Initialize gomock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock EventHandler
	mockEventHandler := eventmocks.NewMockEventHandler[SamplePayload](ctrl)

	testError := errors.New("business logic error")

	// Set up expectations
	mockEventHandler.EXPECT().
		UnmarshalEventMessage(gomock.Any()).
		Return(msg, nil)

	mockEventHandler.EXPECT().
		BeforeHandle(gomock.Any(), msg).
		Return(nil)

	mockEventHandler.EXPECT().
		AfterHandle(gomock.Any(), msg, testError).
		Return(nil)

	// Define a test business logic handler that returns an error
	testGinHandler := func(ctx *gin.Context, msg event.EventMessage[SamplePayload]) error {
		return testError
	}

	// Create the Gin handler
	ginHandler := httphandler.NewGinEventHandler(testGinHandler, mockEventHandler)

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(data))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	ginHandler(ctx)

	// Assert the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), testError.Error())
}

func TestNewGinEventHandler_BeforeHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	timestamp := time.Now().UTC()
	metadata := map[string]string{
		event.MetadataKeyVersion: "1.0",
		event.MetadataKeySource:  "unit_test",
	}
	payload := SamplePayload{Data: "test data"}

	msg := &event.BaseEventMessage[SamplePayload]{
		EventType: "test_event",
		Timestamp: timestamp,
		Payload:   payload,
		Metadata:  metadata,
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := eventmocks.NewMockEventHandler[SamplePayload](ctrl)

	beforeHandleError := errors.New("before handle error")

	mockEventHandler.EXPECT().
		UnmarshalEventMessage(gomock.Any()).
		Return(msg, nil)

	mockEventHandler.EXPECT().
		BeforeHandle(gomock.Any(), msg).
		Return(beforeHandleError)

	testGinHandler := func(ctx *gin.Context, msg event.EventMessage[SamplePayload]) error {
		return nil
	}

	ginHandler := httphandler.NewGinEventHandler(testGinHandler, mockEventHandler)

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(data))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	ginHandler(ctx)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), beforeHandleError.Error())
}

func TestNewGinEventHandler_AfterHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	timestamp := time.Now().UTC()
	metadata := map[string]string{
		event.MetadataKeyVersion: "1.0",
		event.MetadataKeySource:  "unit_test",
	}
	payload := SamplePayload{Data: "test data"}

	msg := &event.BaseEventMessage[SamplePayload]{
		EventType: "test_event",
		Timestamp: timestamp,
		Payload:   payload,
		Metadata:  metadata,
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventHandler := eventmocks.NewMockEventHandler[SamplePayload](ctrl)

	afterHandleError := errors.New("after handle error")

	mockEventHandler.EXPECT().
		UnmarshalEventMessage(gomock.Any()).
		Return(msg, nil)

	mockEventHandler.EXPECT().
		BeforeHandle(gomock.Any(), msg).
		Return(nil)

	mockEventHandler.EXPECT().
		AfterHandle(gomock.Any(), msg, nil).
		Return(afterHandleError)

	testGinHandler := func(ctx *gin.Context, msg event.EventMessage[SamplePayload]) error {
		return nil
	}

	ginHandler := httphandler.NewGinEventHandler(testGinHandler, mockEventHandler)

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(data))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	ginHandler(ctx)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), afterHandleError.Error())
}
