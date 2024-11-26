package errors_test

import (
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientError(t *testing.T) {
	t.Run("should create ClientError successfully with custom message and data", func(t *testing.T) {
		message := "Custom client error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewClientError(message, data)
		require.NotNil(t, err, "Expected ClientError, got nil")

		// Assert that the error is of type ClientError
		clientErr, ok := err.(*domain_error.ClientError)
		require.True(t, ok, "Expected error to be of type ClientError")

		// Assert the BaseError fields
		assert.Equal(t, http.StatusBadRequest, clientErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericClientError), clientErr.Code(), "Unexpected error code")
		assert.Equal(t, message, clientErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, clientErr.GetData(), "Unexpected data")
	})

	t.Run("should create ClientError successfully with default message", func(t *testing.T) {
		err := domain_error.NewClientError("", nil)
		require.NotNil(t, err, "Expected ClientError, got nil")

		// Assert that the error is of type ClientError
		clientErr, ok := err.(*domain_error.ClientError)
		require.True(t, ok, "Expected error to be of type ClientError")

		// Assert the BaseError fields
		assert.Equal(t, http.StatusBadRequest, clientErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericClientError), clientErr.Code(), "Unexpected error code")
	})
}

func TestNewBadRequestError(t *testing.T) {
	t.Run("should create BadRequestError successfully with custom message and data", func(t *testing.T) {
		message := "Custom bad request error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewBadRequestError(message, data)
		require.NotNil(t, err, "Expected BadRequestError, got nil")

		badRequestErr, ok := err.(*domain_error.BadRequestError)
		require.True(t, ok, "Expected error to be of type BadRequestError")

		assert.Equal(t, http.StatusBadRequest, badRequestErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericBadRequestError), badRequestErr.Code(), "Unexpected error code")
		assert.Equal(t, message, badRequestErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, badRequestErr.GetData(), "Unexpected data")
	})

	t.Run("should create BadRequestError successfully with default message", func(t *testing.T) {
		err := domain_error.NewBadRequestError("", nil)
		require.NotNil(t, err, "Expected BadRequestError, got nil")

		badRequestErr, ok := err.(*domain_error.BadRequestError)
		require.True(t, ok, "Expected error to be of type BadRequestError")

		assert.Equal(t, http.StatusBadRequest, badRequestErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericBadRequestError), badRequestErr.Code(), "Unexpected error code")
	})
}

func TestNewNotFoundError(t *testing.T) {
	t.Run("should create NotFoundError successfully with custom message and data", func(t *testing.T) {
		message := "Custom not found error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewNotFoundError(message, data)
		require.NotNil(t, err, "Expected NotFoundError, got nil")

		notFoundErr, ok := err.(*domain_error.NotFoundError)
		require.True(t, ok, "Expected error to be of type NotFoundError")

		assert.Equal(t, http.StatusNotFound, notFoundErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericNotFoundError), notFoundErr.Code(), "Unexpected error code")
		assert.Equal(t, message, notFoundErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, notFoundErr.GetData(), "Unexpected data")
	})

	t.Run("should create NotFoundError successfully with default message", func(t *testing.T) {
		err := domain_error.NewNotFoundError("", nil)
		require.NotNil(t, err, "Expected NotFoundError, got nil")

		notFoundErr, ok := err.(*domain_error.NotFoundError)
		require.True(t, ok, "Expected error to be of type NotFoundError")

		assert.Equal(t, http.StatusNotFound, notFoundErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericNotFoundError), notFoundErr.Code(), "Unexpected error code")
	})
}

func TestNewConflictError(t *testing.T) {
	t.Run("should create ConflictError successfully with custom message and data", func(t *testing.T) {
		message := "Custom conflict error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewConflictError(message, data)
		require.NotNil(t, err, "Expected ConflictError, got nil")

		conflictErr, ok := err.(*domain_error.ConflictError)
		require.True(t, ok, "Expected error to be of type ConflictError")

		assert.Equal(t, http.StatusConflict, conflictErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericConflictError), conflictErr.Code(), "Unexpected error code")
		assert.Equal(t, message, conflictErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, conflictErr.GetData(), "Unexpected data")
	})

	t.Run("should create ConflictError successfully with default message", func(t *testing.T) {
		err := domain_error.NewConflictError("", nil)
		require.NotNil(t, err, "Expected ConflictError, got nil")

		conflictErr, ok := err.(*domain_error.ConflictError)
		require.True(t, ok, "Expected error to be of type ConflictError")

		assert.Equal(t, http.StatusConflict, conflictErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericConflictError), conflictErr.Code(), "Unexpected error code")
	})
}

func TestNewUnprocessableEntityError(t *testing.T) {
	t.Run("should create UnprocessableEntityError successfully with custom message and data", func(t *testing.T) {
		message := "Custom unprocessable entity error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewUnprocessableEntityError(message, data)
		require.NotNil(t, err, "Expected UnprocessableEntityError, got nil")

		unprocessableErr, ok := err.(*domain_error.UnprocessableEntityError)
		require.True(t, ok, "Expected error to be of type UnprocessableEntityError")

		assert.Equal(t, http.StatusUnprocessableEntity, unprocessableErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericUnprocessableEntityError), unprocessableErr.Code(), "Unexpected error code")
		assert.Equal(t, message, unprocessableErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, unprocessableErr.GetData(), "Unexpected data")
	})

	t.Run("should create UnprocessableEntityError successfully with default message", func(t *testing.T) {
		err := domain_error.NewUnprocessableEntityError("", nil)
		require.NotNil(t, err, "Expected UnprocessableEntityError, got nil")

		unprocessableErr, ok := err.(*domain_error.UnprocessableEntityError)
		require.True(t, ok, "Expected error to be of type UnprocessableEntityError")

		assert.Equal(t, http.StatusUnprocessableEntity, unprocessableErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericUnprocessableEntityError), unprocessableErr.Code(), "Unexpected error code")
	})
}
