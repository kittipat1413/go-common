package errors_test

import (
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
