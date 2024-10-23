package errors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapErrorWithPrefix(t *testing.T) {
	// MockDomainError is a mock implementation of the DomainError interface for testing.
	type MockDomainError struct {
		*domain_error.BaseError
	}
	baseErr, _ := domain_error.NewBaseError("20001", "mock domain error", http.StatusBadRequest, nil)
	domainErr := &MockDomainError{BaseError: baseErr}

	tests := []struct {
		name     string
		initial  error
		prefix   string
		expected string
	}{
		{
			name:     "should wrap error with given prefix",
			initial:  errors.New("original error"),
			prefix:   "test prefix",
			expected: "test prefix: original error",
		},
		{
			name:     "should wrap domain error with given prefix",
			initial:  domainErr,
			prefix:   "test prefix",
			expected: "test prefix: mock domain error",
		},
		{
			name:     "should do nothing if error is nil",
			initial:  nil,
			prefix:   "test prefix",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the initial error.
			err := tt.initial

			// Wrap the error with the prefix.
			domain_error.WrapErrorWithPrefix(tt.prefix, &err)

			// Check the result.
			if tt.expected == "" {
				assert.Nil(t, err, "expected error to be nil")
			} else {
				require.NotNil(t, err, "expected error to be non-nil")
				assert.Equal(t, tt.expected, err.Error())
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	// MockDomainError is a mock implementation of the DomainError interface for testing.
	type MockDomainError struct {
		*domain_error.BaseError
	}
	baseErr, _ := domain_error.NewBaseError("20001", "mock domain error", http.StatusBadRequest, nil)
	domainErr1 := &MockDomainError{BaseError: baseErr}
	domainErr2 := &MockDomainError{BaseError: baseErr}

	tests := []struct {
		name        string
		err1        error
		err2        error
		expectedErr string
	}{
		{
			name:        "should return nil if both errors are nil",
			err1:        nil,
			err2:        nil,
			expectedErr: "",
		},
		{
			name:        "should return err1 if err2 is nil",
			err1:        errors.New("error 1"),
			err2:        nil,
			expectedErr: "error 1",
		},
		{
			name:        "should return err2 if err1 is nil",
			err1:        nil,
			err2:        errors.New("error 2"),
			expectedErr: "error 2",
		},
		{
			name:        "should wrap both errors when both are non-nil",
			err1:        errors.New("error 1"),
			err2:        errors.New("error 2"),
			expectedErr: "error 2: error 1",
		},
		{
			name:        "should wrap domain error with standard error",
			err1:        domainErr1,
			err2:        errors.New("error 2"),
			expectedErr: "error 2: mock domain error",
		},
		{
			name:        "should wrap standard error with domain error",
			err1:        errors.New("error 1"),
			err2:        domainErr2,
			expectedErr: "mock domain error: error 1",
		},
		{
			name:        "should wrap domain errors",
			err1:        domainErr1,
			err2:        domainErr2,
			expectedErr: "mock domain error: mock domain error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call WrapError with the two errors
			result := domain_error.WrapError(tt.err1, tt.err2)

			// Check the result
			if tt.expectedErr == "" {
				assert.Nil(t, result, "expected error to be nil")
			} else {
				require.NotNil(t, result, "expected non-nil error")
				assert.Equal(t, tt.expectedErr, result.Error())
			}
		})
	}
}

func TestUnwrapDomainError(t *testing.T) {
	// MockDomainError is a mock implementation of the DomainError interface for testing.
	type MockDomainError struct {
		*domain_error.BaseError
	}

	tests := []struct {
		name          string
		prepareErr    func() error
		expectedMsg   string
		expectedFound bool
	}{
		{
			name: "should return DomainError when it is the error",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("20001", "mock domain error", http.StatusBadRequest, nil)
				domainErr := &MockDomainError{BaseError: baseErr}
				return domainErr
			},
			expectedMsg:   "mock domain error",
			expectedFound: true,
		},
		{
			name: "should return nil when no DomainError is found",
			prepareErr: func() error {
				return errors.New("not a DomainError")
			},
			expectedMsg:   "",
			expectedFound: false,
		},
		{
			name: "should return DomainError when it exists in the error chain",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("21001", "mock domain error", http.StatusBadRequest, nil)
				domainErr := &MockDomainError{BaseError: baseErr}
				return fmt.Errorf("wrapped error: %w", domainErr)
			},
			expectedMsg:   "mock domain error",
			expectedFound: true,
		},
		{
			name: "should unwrap multiple layers to find DomainError",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("22001", "deeply wrapped domain error", http.StatusConflict, nil)
				domainErr := &MockDomainError{BaseError: baseErr}
				wrappedErr := fmt.Errorf("level 1: %w", domainErr)
				wrappedErr = fmt.Errorf("level 2: %w", wrappedErr)
				wrappedErr = fmt.Errorf("level 3: %w", wrappedErr)
				return wrappedErr
			},
			expectedMsg:   "deeply wrapped domain error",
			expectedFound: true,
		},
		{
			name: "should return DomainError when it is wrapped with WrapError",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("21001", "mock domain error", http.StatusBadRequest, nil)
				domainErr := &MockDomainError{BaseError: baseErr}
				return domain_error.WrapError(errors.New("wrapped error"), domainErr)
			},
			expectedMsg:   "mock domain error",
			expectedFound: true,
		},
		{
			name: "should unwrap multiple layers to find DomainError when it is wrapped with WrapError",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("22001", "deeply wrapped domain error", http.StatusConflict, nil)
				domainErr := &MockDomainError{BaseError: baseErr}
				wrappedErr := domain_error.WrapError(errors.New("level 1"), domainErr)
				wrappedErr = domain_error.WrapError(errors.New("level 2"), wrappedErr)
				wrappedErr = domain_error.WrapError(errors.New("level 3"), wrappedErr)
				return wrappedErr
			},
			expectedMsg:   "deeply wrapped domain error",
			expectedFound: true,
		},
		{
			name: "should return nil when DomainError does not have BaseError",
			prepareErr: func() error {
				return &MockDomainError{}
			},
			expectedMsg:   "",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unwrap the error and check if the DomainError is found.
			result := domain_error.UnwrapDomainError(tt.prepareErr())

			// Check the result.
			if tt.expectedFound {
				require.NotNil(t, result, "expected to find DomainError, got nil")
				assert.Equal(t, tt.expectedMsg, result.GetMessage())
			} else {
				assert.Nil(t, result, "expected no DomainError, got a result")
			}
		})
	}
}
