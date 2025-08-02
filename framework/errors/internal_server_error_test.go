package errors_test

import (
	"errors"
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInternalServerError(t *testing.T) {
	t.Run("should create InternalServerError successfully with custom message and data", func(t *testing.T) {
		message := "Custom internal server error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewInternalServerError(message, data)
		require.NotNil(t, err, "Expected InternalServerError, got nil")

		internalErr, ok := err.(*domain_error.InternalServerError)
		require.True(t, ok, "Expected error to be of type InternalServerError")

		assert.Equal(t, http.StatusInternalServerError, internalErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericInternalServerError), internalErr.Code(), "Unexpected error code")
		assert.Equal(t, message, internalErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, internalErr.GetData(), "Unexpected data")
	})

	t.Run("should create InternalServerError successfully with default message", func(t *testing.T) {
		err := domain_error.NewInternalServerError("", nil)
		require.NotNil(t, err, "Expected InternalServerError, got nil")

		internalErr, ok := err.(*domain_error.InternalServerError)
		require.True(t, ok, "Expected error to be of type InternalServerError")

		assert.Equal(t, http.StatusInternalServerError, internalErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericInternalServerError), internalErr.Code(), "Unexpected error code")
	})
}

func TestNewDatabaseError(t *testing.T) {
	t.Run("should create DatabaseError successfully with custom message and data", func(t *testing.T) {
		message := "Custom database error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewDatabaseError(message, data)
		require.NotNil(t, err, "Expected DatabaseError, got nil")

		dbErr, ok := err.(*domain_error.DatabaseError)
		require.True(t, ok, "Expected error to be of type DatabaseError")

		assert.Equal(t, http.StatusInternalServerError, dbErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericDatabaseError), dbErr.Code(), "Unexpected error code")
		assert.Equal(t, message, dbErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, dbErr.GetData(), "Unexpected data")
	})

	t.Run("should create DatabaseError successfully with default message", func(t *testing.T) {
		err := domain_error.NewDatabaseError("", nil)
		require.NotNil(t, err, "Expected DatabaseError, got nil")

		dbErr, ok := err.(*domain_error.DatabaseError)
		require.True(t, ok, "Expected error to be of type DatabaseError")

		assert.Equal(t, http.StatusInternalServerError, dbErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericDatabaseError), dbErr.Code(), "Unexpected error code")
	})
}

func TestNewThirdPartyError(t *testing.T) {
	t.Run("should create ThirdPartyError successfully with custom message and data", func(t *testing.T) {
		message := "Custom third party error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewThirdPartyError(message, data)
		require.NotNil(t, err, "Expected ThirdPartyError, got nil")

		thirdPartyErr, ok := err.(*domain_error.ThirdPartyError)
		require.True(t, ok, "Expected error to be of type ThirdPartyError")

		assert.Equal(t, http.StatusBadGateway, thirdPartyErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericThirdPartyError), thirdPartyErr.Code(), "Unexpected error code")
		assert.Equal(t, message, thirdPartyErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, thirdPartyErr.GetData(), "Unexpected data")
	})

	t.Run("should create ThirdPartyError successfully with default message", func(t *testing.T) {
		err := domain_error.NewThirdPartyError("", nil)
		require.NotNil(t, err, "Expected ThirdPartyError, got nil")

		thirdPartyErr, ok := err.(*domain_error.ThirdPartyError)
		require.True(t, ok, "Expected error to be of type ThirdPartyError")

		assert.Equal(t, http.StatusBadGateway, thirdPartyErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericThirdPartyError), thirdPartyErr.Code(), "Unexpected error code")
	})
}

func TestNewServiceUnavailableError(t *testing.T) {
	t.Run("should create ServiceUnavailableError successfully with custom message and data", func(t *testing.T) {
		message := "Custom service unavailable error message"
		data := map[string]string{"key": "value"}

		err := domain_error.NewServiceUnavailableError(message, data)
		require.NotNil(t, err, "Expected ServiceUnavailableError, got nil")

		serviceUnavailableErr, ok := err.(*domain_error.ServiceUnavailableError)
		require.True(t, ok, "Expected error to be of type ServiceUnavailableError")

		assert.Equal(t, http.StatusServiceUnavailable, serviceUnavailableErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericServiceUnavailableError), serviceUnavailableErr.Code(), "Unexpected error code")
		assert.Equal(t, message, serviceUnavailableErr.GetMessage(), "Unexpected error message")
		assert.Equal(t, data, serviceUnavailableErr.GetData(), "Unexpected data")
	})

	t.Run("should create ServiceUnavailableError successfully with default message", func(t *testing.T) {
		err := domain_error.NewServiceUnavailableError("", nil)
		require.NotNil(t, err, "Expected ServiceUnavailableError, got nil")

		serviceUnavailableErr, ok := err.(*domain_error.ServiceUnavailableError)
		require.True(t, ok, "Expected error to be of type ServiceUnavailableError")

		assert.Equal(t, http.StatusServiceUnavailable, serviceUnavailableErr.GetHTTPCode(), "Unexpected HTTP code")
		assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericServiceUnavailableError), serviceUnavailableErr.Code(), "Unexpected error code")
	})
}

func TestInternalServerErrorAs(t *testing.T) {
	internalErr := domain_error.NewInternalServerError("test error", nil).(*domain_error.InternalServerError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := internalErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, internalErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.InternalServerError
		result := internalErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *internalErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.DatabaseError
		result := internalErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := internalErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.InternalServerError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := internalErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, internalErr, target, "Target should reference the same instance")
		assert.Equal(t, internalErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.InternalServerError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.InternalServerError{}, target, "Target should be zero value initially")

		result := internalErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.InternalServerError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *internalErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, internalErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := errors.As(internalErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, internalErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.InternalServerError
		result := errors.As(internalErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *internalErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.DatabaseError
		result := errors.As(internalErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.InternalServerError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(internalErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, internalErr, target, "Target should reference the same instance")
		assert.Equal(t, internalErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.InternalServerError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.InternalServerError{}, target, "Target should be zero value initially")

		result := errors.As(internalErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.InternalServerError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *internalErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, internalErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestDatabaseErrorAs(t *testing.T) {
	dbErr := domain_error.NewDatabaseError("test error", nil).(*domain_error.DatabaseError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.DatabaseError
		result := dbErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, dbErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.DatabaseError
		result := dbErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *dbErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := dbErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := dbErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.DatabaseError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := dbErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, dbErr, target, "Target should reference the same instance")
		assert.Equal(t, dbErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.DatabaseError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.DatabaseError{}, target, "Target should be zero value initially")

		result := dbErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.DatabaseError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *dbErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, dbErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.DatabaseError
		result := errors.As(dbErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, dbErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.DatabaseError
		result := errors.As(dbErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *dbErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := errors.As(dbErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.DatabaseError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(dbErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, dbErr, target, "Target should reference the same instance")
		assert.Equal(t, dbErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.DatabaseError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.DatabaseError{}, target, "Target should be zero value initially")

		result := errors.As(dbErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.DatabaseError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *dbErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, dbErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestThirdPartyErrorAs(t *testing.T) {
	thirdPartyErr := domain_error.NewThirdPartyError("test error", nil).(*domain_error.ThirdPartyError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ThirdPartyError
		result := thirdPartyErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, thirdPartyErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.ThirdPartyError
		result := thirdPartyErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *thirdPartyErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := thirdPartyErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := thirdPartyErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ThirdPartyError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := thirdPartyErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, thirdPartyErr, target, "Target should reference the same instance")
		assert.Equal(t, thirdPartyErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ThirdPartyError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.ThirdPartyError{}, target, "Target should be zero value initially")

		result := thirdPartyErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.ThirdPartyError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *thirdPartyErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, thirdPartyErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ThirdPartyError
		result := errors.As(thirdPartyErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, thirdPartyErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.ThirdPartyError
		result := errors.As(thirdPartyErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *thirdPartyErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := errors.As(thirdPartyErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ThirdPartyError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(thirdPartyErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, thirdPartyErr, target, "Target should reference the same instance")
		assert.Equal(t, thirdPartyErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ThirdPartyError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.ThirdPartyError{}, target, "Target should be zero value initially")

		result := errors.As(thirdPartyErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.ThirdPartyError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *thirdPartyErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, thirdPartyErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}

func TestServiceUnavailableErrorAs(t *testing.T) {
	serviceErr := domain_error.NewServiceUnavailableError("test error", nil).(*domain_error.ServiceUnavailableError)

	t.Run("should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ServiceUnavailableError
		result := serviceErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, serviceErr, target, "Target should be assigned correctly")
	})

	t.Run("should work with pointer target", func(t *testing.T) {
		var target domain_error.ServiceUnavailableError
		result := serviceErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.Equal(t, *serviceErr, target, "Target should be assigned correctly")
	})

	t.Run("should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := serviceErr.As(&target)
		assert.False(t, result, "As should return false for incompatible target")
	})

	t.Run("should return false for nil target", func(t *testing.T) {
		result := serviceErr.As(nil)
		assert.False(t, result, "As should return false for nil target")
	})

	t.Run("should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ServiceUnavailableError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := serviceErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, serviceErr, target, "Target should reference the same instance")
		assert.Equal(t, serviceErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ServiceUnavailableError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.ServiceUnavailableError{}, target, "Target should be zero value initially")

		result := serviceErr.As(&target)

		assert.True(t, result, "As should return true")
		assert.NotEqual(t, domain_error.ServiceUnavailableError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *serviceErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, serviceErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should work with pointer to pointer target", func(t *testing.T) {
		var target *domain_error.ServiceUnavailableError
		result := errors.As(serviceErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, serviceErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should work with pointer target", func(t *testing.T) {
		var target domain_error.ServiceUnavailableError
		result := errors.As(serviceErr, &target)
		assert.True(t, result, "errors.As should return true")
		assert.Equal(t, *serviceErr, target, "Target should be assigned correctly")
	})

	t.Run("errors.As should return false for incompatible target", func(t *testing.T) {
		var target *domain_error.InternalServerError
		result := errors.As(serviceErr, &target)
		assert.False(t, result, "errors.As should return false for incompatible target")
	})

	t.Run("errors.As should verify target assignment for pointer to pointer", func(t *testing.T) {
		var target *domain_error.ServiceUnavailableError
		// Ensure target is initially nil
		assert.Nil(t, target, "Target should be nil initially")

		result := errors.As(serviceErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotNil(t, target, "Target should not be nil after assignment")
		assert.Same(t, serviceErr, target, "Target should reference the same instance")
		assert.Equal(t, serviceErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})

	t.Run("errors.As should verify target assignment for pointer", func(t *testing.T) {
		var target domain_error.ServiceUnavailableError
		// Ensure target has zero value initially
		assert.Equal(t, domain_error.ServiceUnavailableError{}, target, "Target should be zero value initially")

		result := errors.As(serviceErr, &target)

		assert.True(t, result, "errors.As should return true")
		assert.NotEqual(t, domain_error.ServiceUnavailableError{}, target, "Target should not be zero value after assignment")
		assert.Equal(t, *serviceErr, target, "Target should have same value as dereferenced error")
		assert.Equal(t, serviceErr.GetMessage(), target.GetMessage(), "Target should have same message")
	})
}
