package errors_test

import (
	"errors"
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

func TestAuthenticationErrorAs(t *testing.T) {
	authErr := domain_error.NewAuthenticationError("test error", nil).(*domain_error.AuthenticationError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		result := authErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, authErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.AuthenticationError
		result := authErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *authErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.UnauthorizedError
		result := authErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := authErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		assert.Nil(t, target, "Target should be nil initially")

		result := authErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, authErr, target, "Target should reference the same instance")
		assert.Equal(t, authErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.AuthenticationError
		assert.Equal(t, domain_error.AuthenticationError{}, target, "Target should be zero value initially")

		result := authErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.AuthenticationError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *authErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, authErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		result := errors.As(authErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, authErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.AuthenticationError
		result := errors.As(authErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *authErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.UnauthorizedError
		result := errors.As(authErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(authErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, authErr, target, "Target should reference the same instance")
		assert.Equal(t, authErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.AuthenticationError
		assert.Equal(t, domain_error.AuthenticationError{}, target, "Target should be zero value initially")

		result := errors.As(authErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.AuthenticationError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *authErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, authErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestUnauthorizedErrorAs(t *testing.T) {
	unauthorizedErr := domain_error.NewUnauthorizedError("test error", nil).(*domain_error.UnauthorizedError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.UnauthorizedError
		result := unauthorizedErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, unauthorizedErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.UnauthorizedError
		result := unauthorizedErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *unauthorizedErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		result := unauthorizedErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := unauthorizedErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.UnauthorizedError
		assert.Nil(t, target, "Target should be nil initially")

		result := unauthorizedErr.As(&target)
		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, unauthorizedErr, target, "Target should reference the same instance")
		assert.Equal(t, unauthorizedErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.UnauthorizedError
		assert.Equal(t, domain_error.UnauthorizedError{}, target, "Target should be zero value initially")

		result := unauthorizedErr.As(&target)
		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.UnauthorizedError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *unauthorizedErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, unauthorizedErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.UnauthorizedError
		result := errors.As(unauthorizedErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, unauthorizedErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.UnauthorizedError
		result := errors.As(unauthorizedErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *unauthorizedErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		result := errors.As(unauthorizedErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.UnauthorizedError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(unauthorizedErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, unauthorizedErr, target, "Target should reference the same instance")
		assert.Equal(t, unauthorizedErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.UnauthorizedError
		assert.Equal(t, domain_error.UnauthorizedError{}, target, "Target should be zero value initially")

		result := errors.As(unauthorizedErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.UnauthorizedError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *unauthorizedErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, unauthorizedErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestForbiddenErrorAs(t *testing.T) {
	forbiddenErr := domain_error.NewForbiddenError("test error", nil).(*domain_error.ForbiddenError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ForbiddenError
		result := forbiddenErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, forbiddenErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.ForbiddenError
		result := forbiddenErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *forbiddenErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		result := forbiddenErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := forbiddenErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ForbiddenError
		assert.Nil(t, target, "Target should be nil initially")

		result := forbiddenErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, forbiddenErr, target, "Target should reference the same instance")
		assert.Equal(t, forbiddenErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ForbiddenError
		assert.Equal(t, domain_error.ForbiddenError{}, target, "Target should be zero value initially")

		result := forbiddenErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.ForbiddenError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *forbiddenErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, forbiddenErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ForbiddenError
		result := errors.As(forbiddenErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, forbiddenErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.ForbiddenError
		result := errors.As(forbiddenErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *forbiddenErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.AuthenticationError
		result := errors.As(forbiddenErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ForbiddenError
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(forbiddenErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, forbiddenErr, target, "Target should reference the same instance")
		assert.Equal(t, forbiddenErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ForbiddenError
		assert.Equal(t, domain_error.ForbiddenError{}, target, "Target should be zero value initially")

		result := errors.As(forbiddenErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.ForbiddenError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *forbiddenErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, forbiddenErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}
