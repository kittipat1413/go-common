package errors_test

import (
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthenticationError(t *testing.T) {
	t.Run("should create AuthenticationError successfully with custom message and data", func(t *testing.T) {
		message := "Custom authentication error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewAuthenticationError(message, data)
		require.NotNil(t, err, "Expected AuthenticationError, got nil")

		authErr, ok := err.(*domain_error.AuthenticationError)
		require.True(t, ok, "Expected error to be of type AuthenticationError")

		assert.Equal(t, http.StatusUnauthorized, authErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericAuthError), authErr.Code(), "Unexpected error code")
		assert.Equal(t, message, authErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, authErr.GetData(), "Unexpected data")
	})

	t.Run("should create AuthenticationError successfully with default message", func(t *testing.T) {
		err := domain_error.NewAuthenticationError("", nil)
		require.NotNil(t, err, "Expected AuthenticationError, got nil")

		authErr, ok := err.(*domain_error.AuthenticationError)
		require.True(t, ok, "Expected error to be of type AuthenticationError")

		assert.Equal(t, http.StatusUnauthorized, authErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericAuthError), authErr.Code(), "Unexpected error code")
	})
}

func TestNewUnauthorizedError(t *testing.T) {
	t.Run("should create UnauthorizedError successfully with custom message and data", func(t *testing.T) {
		message := "Custom unauthorized error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewUnauthorizedError(message, data)
		require.NotNil(t, err, "Expected UnauthorizedError, got nil")

		unauthorizedErr, ok := err.(*domain_error.UnauthorizedError)
		require.True(t, ok, "Expected error to be of type UnauthorizedError")

		assert.Equal(t, http.StatusUnauthorized, unauthorizedErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericUnauthorizedError), unauthorizedErr.Code(), "Unexpected error code")
		assert.Equal(t, message, unauthorizedErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, unauthorizedErr.GetData(), "Unexpected data")
	})

	t.Run("should create UnauthorizedError successfully with default message", func(t *testing.T) {
		err := domain_error.NewUnauthorizedError("", nil)
		require.NotNil(t, err, "Expected UnauthorizedError, got nil")

		unauthorizedErr, ok := err.(*domain_error.UnauthorizedError)
		require.True(t, ok, "Expected error to be of type UnauthorizedError")

		assert.Equal(t, http.StatusUnauthorized, unauthorizedErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericUnauthorizedError), unauthorizedErr.Code(), "Unexpected error code")
	})
}

func TestNewForbiddenError(t *testing.T) {
	t.Run("should create ForbiddenError successfully with custom message and data", func(t *testing.T) {
		message := "Custom forbidden error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewForbiddenError(message, data)
		require.NotNil(t, err, "Expected ForbiddenError, got nil")

		forbiddenErr, ok := err.(*domain_error.ForbiddenError)
		require.True(t, ok, "Expected error to be of type ForbiddenError")

		assert.Equal(t, http.StatusForbidden, forbiddenErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericForbiddenError), forbiddenErr.Code(), "Unexpected error code")
		assert.Equal(t, message, forbiddenErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, forbiddenErr.GetData(), "Unexpected data")
	})

	t.Run("should create ForbiddenError successfully with default message", func(t *testing.T) {
		err := domain_error.NewForbiddenError("", nil)
		require.NotNil(t, err, "Expected ForbiddenError, got nil")

		forbiddenErr, ok := err.(*domain_error.ForbiddenError)
		require.True(t, ok, "Expected error to be of type ForbiddenError")

		assert.Equal(t, http.StatusForbidden, forbiddenErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericForbiddenError), forbiddenErr.Code(), "Unexpected error code")
	})
}
