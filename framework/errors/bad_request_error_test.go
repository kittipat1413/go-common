package errors_test

import (
	"errors"
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

func TestClientErrorAs(t *testing.T) {
	clientErr := domain_error.NewClientError("test error", nil).(*domain_error.ClientError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ClientError
		result := clientErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, clientErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.ClientError
		result := clientErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *clientErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := clientErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := clientErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ClientError
		assert.Nil(t, target, "Target should be nil initially")

		result := clientErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, clientErr, target, "Target should reference the same instance")
		assert.Equal(t, clientErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ClientError
		assert.Equal(t, domain_error.ClientError{}, target, "Target should be empty initially")

		result := clientErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.ThirdPartyError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *clientErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, clientErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ClientError
		result := errors.As(clientErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, clientErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.ClientError
		result := errors.As(clientErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *clientErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := errors.As(clientErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
		assert.Nil(t, target, "Target should remain nil")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ClientError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(clientErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, clientErr, target, "Target should reference the same instance")
		assert.Equal(t, clientErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ClientError
		assert.Equal(t, domain_error.ClientError{}, target, "Target should be empty initially")

		result := errors.As(clientErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.ClientError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *clientErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, clientErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestBadRequestErrorAs(t *testing.T) {
	badRequestErr := domain_error.NewBadRequestError("test error", nil).(*domain_error.BadRequestError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := badRequestErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, badRequestErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.BadRequestError
		result := badRequestErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *badRequestErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.ClientError
		result := badRequestErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := badRequestErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.BadRequestError
		assert.Nil(t, target, "Target should be nil initially")

		result := badRequestErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, badRequestErr, target, "Target should reference the same instance")
		assert.Equal(t, badRequestErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.BadRequestError
		assert.Equal(t, domain_error.BadRequestError{}, target, "Target should be empty initially")

		result := badRequestErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.BadRequestError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *badRequestErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, badRequestErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := errors.As(badRequestErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, badRequestErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.BadRequestError
		result := errors.As(badRequestErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *badRequestErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.ClientError
		result := errors.As(badRequestErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
		assert.Nil(t, target, "Target should remain nil")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.BadRequestError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(badRequestErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, badRequestErr, target, "Target should reference the same instance")
		assert.Equal(t, badRequestErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.BadRequestError
		assert.Equal(t, domain_error.BadRequestError{}, target, "Target should be empty initially")

		result := errors.As(badRequestErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.BadRequestError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *badRequestErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, badRequestErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestNotFoundErrorAs(t *testing.T) {
	notFoundErr := domain_error.NewNotFoundError("test error", nil).(*domain_error.NotFoundError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.NotFoundError
		result := notFoundErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, notFoundErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.NotFoundError
		result := notFoundErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *notFoundErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := notFoundErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := notFoundErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.NotFoundError
		assert.Nil(t, target, "Target should be nil initially")

		result := notFoundErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, notFoundErr, target, "Target should reference the same instance")
		assert.Equal(t, notFoundErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.NotFoundError
		assert.Equal(t, domain_error.NotFoundError{}, target, "Target should be empty initially")

		result := notFoundErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.NotFoundError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *notFoundErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, notFoundErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.NotFoundError
		result := errors.As(notFoundErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, notFoundErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.NotFoundError
		result := errors.As(notFoundErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *notFoundErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := errors.As(notFoundErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
		assert.Nil(t, target, "Target should remain nil")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.NotFoundError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(notFoundErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, notFoundErr, target, "Target should reference the same instance")
		assert.Equal(t, notFoundErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.NotFoundError
		assert.Equal(t, domain_error.NotFoundError{}, target, "Target should be empty initially")

		result := errors.As(notFoundErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.NotFoundError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *notFoundErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, notFoundErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestConflictErrorAs(t *testing.T) {
	conflictErr := domain_error.NewConflictError("test error", nil).(*domain_error.ConflictError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ConflictError
		result := conflictErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, conflictErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.ConflictError
		result := conflictErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *conflictErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := conflictErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := conflictErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ConflictError
		assert.Nil(t, target, "Target should be nil initially")

		result := conflictErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, conflictErr, target, "Target should reference the same instance")
		assert.Equal(t, conflictErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ConflictError
		assert.Equal(t, domain_error.ConflictError{}, target, "Target should be empty initially")

		result := conflictErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.ConflictError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *conflictErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, conflictErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ConflictError
		result := errors.As(conflictErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, conflictErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.ConflictError
		result := errors.As(conflictErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *conflictErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := errors.As(conflictErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
		assert.Nil(t, target, "Target should remain nil")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ConflictError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(conflictErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, conflictErr, target, "Target should reference the same instance")
		assert.Equal(t, conflictErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ConflictError
		assert.Equal(t, domain_error.ConflictError{}, target, "Target should be empty initially")

		result := errors.As(conflictErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.ConflictError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *conflictErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, conflictErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestUnprocessableEntityErrorAs(t *testing.T) {
	unprocessableErr := domain_error.NewUnprocessableEntityError("test error", nil).(*domain_error.UnprocessableEntityError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.UnprocessableEntityError
		result := unprocessableErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, unprocessableErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.UnprocessableEntityError
		result := unprocessableErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *unprocessableErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := unprocessableErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := unprocessableErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.UnprocessableEntityError
		assert.Nil(t, target, "Target should be nil initially")

		result := unprocessableErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, unprocessableErr, target, "Target should reference the same instance")
		assert.Equal(t, unprocessableErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.UnprocessableEntityError
		assert.Equal(t, domain_error.UnprocessableEntityError{}, target, "Target should be empty initially")

		result := unprocessableErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.UnprocessableEntityError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *unprocessableErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, unprocessableErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.UnprocessableEntityError
		result := errors.As(unprocessableErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, unprocessableErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.UnprocessableEntityError
		result := errors.As(unprocessableErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *unprocessableErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.BadRequestError
		result := errors.As(unprocessableErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
		assert.Nil(t, target, "Target should remain nil")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.UnprocessableEntityError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(unprocessableErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, unprocessableErr, target, "Target should reference the same instance")
		assert.Equal(t, unprocessableErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.UnprocessableEntityError
		assert.Equal(t, domain_error.UnprocessableEntityError{}, target, "Target should be empty initially")

		result := errors.As(unprocessableErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.UnprocessableEntityError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *unprocessableErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, unprocessableErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}
