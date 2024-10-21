package errors_test

import (
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBadRequestError(t *testing.T) {
	err, createErr := domain_error.NewBadRequestError()
	require.NoError(t, createErr, "Expected no error when creating BadRequestError")
	require.NotNil(t, err, "Expected BadRequestError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericBadRequestError), err.Code(), "Expected code to be StatusCodeGenericBadRequestError")
	assert.Equal(t, http.StatusBadRequest, err.GetHTTPCode(), "Expected status code to be http.StatusBadRequest")
}

func TestNewInvalidParametersError(t *testing.T) {
	err, createErr := domain_error.NewInvalidParametersError()
	require.NoError(t, createErr, "Expected no error when creating InvalidParametersError")
	require.NotNil(t, err, "Expected InvalidParametersError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericInvalidParameters), err.Code(), "Expected code to be StatusCodeGenericInvalidParameters")
	assert.Equal(t, http.StatusBadRequest, err.GetHTTPCode(), "Expected status code to be http.StatusBadRequest")
}

func TestNewDuplicatedEntryError(t *testing.T) {
	err, createErr := domain_error.NewDuplicatedEntryError()
	require.NoError(t, createErr, "Expected no error when creating DuplicatedEntryError")
	require.NotNil(t, err, "Expected DuplicatedEntryError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericDuplicatedEntry), err.Code(), "Expected code to be StatusCodeGenericDuplicatedEntry")
	assert.Equal(t, http.StatusConflict, err.GetHTTPCode(), "Expected status code to be http.StatusConflict")
}

func TestNewNotFoundError(t *testing.T) {
	err, createErr := domain_error.NewNotFoundError()
	require.NoError(t, createErr, "Expected no error when creating NotFoundError")
	require.NotNil(t, err, "Expected NotFoundError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericNotFoundError), err.Code(), "Expected code to be StatusCodeGenericNotFoundError")
	assert.Equal(t, http.StatusNotFound, err.GetHTTPCode(), "Expected status code to be http.StatusNotFound")
}

func TestNewUnprocessableEntityError(t *testing.T) {
	err, createErr := domain_error.NewUnprocessableEntityError()
	require.NoError(t, createErr, "Expected no error when creating UnprocessableEntityError")
	require.NotNil(t, err, "Expected UnprocessableEntityError, got nil")

	assert.Equal(t, domain_error.GetFullCode(domain_error.StatusCodeGenericUnprocessableEntity), err.Code(), "Expected code to be StatusCodeGenericUnprocessableEntity")
	assert.Equal(t, http.StatusUnprocessableEntity, err.GetHTTPCode(), "Expected status code to be http.StatusUnprocessableEntity")
}
