package errors_test

import (
	"errors"
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseError(t *testing.T) {
	// Test cases for creating a new BaseError
	tests := []struct {
		name             string
		code             string
		message          string
		data             interface{}
		expectedError    bool
		expectedHttpCode int
		expectedMsg      string
	}{
		{
			name:             "valid BaseError creation",
			code:             "400001",
			message:          "valid error",
			data:             nil,
			expectedError:    false,
			expectedHttpCode: http.StatusBadRequest,
			expectedMsg:      "valid error",
		},
		{
			name:             "BaseError with empty message, should use default",
			code:             "500000",
			message:          "",
			data:             nil,
			expectedError:    false,
			expectedHttpCode: http.StatusInternalServerError,
			expectedMsg:      "An internal server error occurred. Please try again later.",
		},
		{
			name:             "BaseError with empty message and no default, should use generic message",
			code:             "500001",
			message:          "",
			data:             nil,
			expectedError:    false,
			expectedHttpCode: http.StatusInternalServerError,
			expectedMsg:      "An unexpected error occurred. Please try again later.",
		},
		{
			name:             "invalid code length",
			code:             "123",
			message:          "short code error",
			data:             nil,
			expectedError:    true,
			expectedHttpCode: http.StatusBadRequest,
			expectedMsg:      "",
		},
		{
			name:          "invalid category",
			code:          "799001",
			message:       "unknown category error",
			data:          nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to create a new BaseError
			baseErr, err := domain_error.NewBaseError(tt.code, tt.message, tt.data)

			// Verify the results
			if tt.expectedError {
				require.Error(t, err, "expected an error but got none")
				assert.Nil(t, baseErr, "expected nil BaseError on error")
			} else {
				require.NoError(t, err, "expected no error but got one")
				require.NotNil(t, baseErr, "expected a valid BaseError")

				// Verify the BaseError fields
				assert.Equal(t, domain_error.GetFullCode(tt.code), baseErr.Code(), "unexpected error code")
				assert.Equal(t, tt.expectedMsg, baseErr.GetMessage(), "unexpected error message")
				assert.Equal(t, tt.expectedHttpCode, baseErr.GetHTTPCode(), "unexpected HTTP code")
				assert.Equal(t, tt.data, baseErr.GetData(), "unexpected data")
			}
		})
	}
}

func TestBaseErrorMethods(t *testing.T) {
	// Create a sample BaseError for testing
	domain_error.SetServicePrefix("TEST")
	baseErr, err := domain_error.NewBaseError("400001", "sample error", "extra data")
	require.NoError(t, err, "expected no error when creating BaseError")

	t.Run("Test Code() method", func(t *testing.T) {
		assert.Equal(t, "TEST-400001", baseErr.Code(), "unexpected Code() output")
	})

	t.Run("Test GetMessage() method", func(t *testing.T) {
		assert.Equal(t, "sample error", baseErr.GetMessage(), "unexpected GetMessage() output")
	})

	t.Run("Test GetHTTPCode() method", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, baseErr.GetHTTPCode(), "unexpected GetHTTPCode() output")
	})

	t.Run("Test Error() method", func(t *testing.T) {
		assert.Equal(t, "sample error", baseErr.Error(), "unexpected Error() output")
	})

	t.Run("Test GetData() method", func(t *testing.T) {
		assert.Equal(t, "extra data", baseErr.GetData(), "unexpected GetData() output")
	})
}

func TestExtractBaseError(t *testing.T) {
	// Mock error type that embeds BaseError
	type MockDomainError struct {
		*domain_error.BaseError
	}

	// Test cases
	tests := []struct {
		name          string
		prepareErr    func() error
		expectedFound bool
		expectedMsg   string
	}{
		{
			name: "should return BaseError when input is BaseError",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("400001", "mock domain error", nil)
				return baseErr
			},
			expectedFound: true,
			expectedMsg:   "mock domain error",
		},
		{
			name: "should extract BaseError when directly embedded in a pointer struct",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("400001", "mock domain error", nil)
				return &MockDomainError{BaseError: baseErr}
			},
			expectedFound: true,
			expectedMsg:   "mock domain error",
		},
		{
			name: "should extract BaseError when directly embedded in a non-pointer struct",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("400001", "mock domain error", nil)
				return MockDomainError{BaseError: baseErr}
			},
			expectedFound: true,
			expectedMsg:   "mock domain error",
		},
		{
			name: "should not find BaseError when not embedded",
			prepareErr: func() error {
				return errors.New("standard error")
			},
			expectedFound: false,
			expectedMsg:   "",
		},
		{
			name: "should not find BaseError when embedded is nil",
			prepareErr: func() error {
				return &MockDomainError{BaseError: nil}
			},
			expectedFound: false,
			expectedMsg:   "",
		},
		{
			name: "should not find BaseError when error is nil",
			prepareErr: func() error {
				return nil
			},
			expectedFound: false,
			expectedMsg:   "",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prepareErr()
			result := domain_error.ExtractBaseError(err)

			if tt.expectedFound {
				require.NotNil(t, result, "expected to find BaseError, got nil")
				assert.Equal(t, tt.expectedMsg, result.GetMessage())
			} else {
				assert.Nil(t, result, "expected to not find BaseError, but found one")
			}
		})
	}
}
