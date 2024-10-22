package errors_test

import (
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInternalServerError(t *testing.T) {
	err, createErr := domain_error.NewInternalServerError("test data")
	require.NoError(t, createErr, "Expected no error when creating InternalServerError")
	require.NotNil(t, err, "Expected InternalServerError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericInternalError), err.Code(), "Expected code to be StatusCodeGenericInternalError")
	assert.Equal(t, http.StatusInternalServerError, err.GetHTTPCode(), "Expected status code to be http.StatusInternalServerError")
	assert.Equal(t, "test data", err.GetData().(string), "Expected data to be 'test data'")
}

func TestNewDatabaseError(t *testing.T) {
	err, createErr := domain_error.NewDatabaseError("test data")
	require.NoError(t, createErr, "Expected no error when creating DatabaseError")
	require.NotNil(t, err, "Expected DatabaseError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericDatabaseError), err.Code(), "Expected code to be StatusCodeGenericDatabaseError")
	assert.Equal(t, http.StatusInternalServerError, err.GetHTTPCode(), "Expected status code to be http.StatusInternalServerError")
	assert.Equal(t, "test data", err.GetData().(string), "Expected data to be 'test data'")
}

func TestNewThirdPartyError(t *testing.T) {
	err, createErr := domain_error.NewThirdPartyError("test data")
	require.NoError(t, createErr, "Expected no error when creating ThirdPartyError")
	require.NotNil(t, err, "Expected ThirdPartyError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericThirdPartyError), err.Code(), "Expected code to be StatusCodeGenericThirdPartyError")
	assert.Equal(t, http.StatusBadGateway, err.GetHTTPCode(), "Expected status code to be http.StatusBadGateway")
	assert.Equal(t, "test data", err.GetData().(string), "Expected data to be 'test data'")
}
