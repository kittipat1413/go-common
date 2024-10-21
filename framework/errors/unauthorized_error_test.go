package errors_test

import (
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthenticationError(t *testing.T) {
	err, createErr := domain_error.NewAuthenticationError()
	require.NoError(t, createErr, "Expected no error when creating AuthenticationError")
	require.NotNil(t, err, "Expected AuthenticationError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericAuthError), err.Code(), "Expected code to be StatusCodeGenericAuthError")
	assert.Equal(t, http.StatusUnauthorized, err.GetHTTPCode(), "Expected status code to be http.StatusUnauthorized")
}

func TestNewUnauthorizedError(t *testing.T) {
	err, createErr := domain_error.NewUnauthorizedError()
	require.NoError(t, createErr, "Expected no error when creating UnauthorizedError")
	require.NotNil(t, err, "Expected UnauthorizedError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericUnauthorized), err.Code(), "Expected code to be StatusCodeGenericUnauthorized")
	assert.Equal(t, http.StatusUnauthorized, err.GetHTTPCode(), "Expected status code to be http.StatusUnauthorized")
}

func TestNewForbiddenError(t *testing.T) {
	err, createErr := domain_error.NewForbiddenError()
	require.NoError(t, createErr, "Expected no error when creating ForbiddenError")
	require.NotNil(t, err, "Expected ForbiddenError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericForbidden), err.Code(), "Expected code to be StatusCodeGenericForbidden")
	assert.Equal(t, http.StatusForbidden, err.GetHTTPCode(), "Expected status code to be http.StatusForbidden")
}
