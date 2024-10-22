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

func TestWrapError(t *testing.T) {
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

			// Wrap the error.
			domain_error.WrapError(tt.prefix, &err)

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
			name: "should return DomainError when it exists in the error chain",
			prepareErr: func() error {
				baseErr, _ := domain_error.NewBaseError("21001", "mock domain error", http.StatusBadRequest, nil)
				domainErr := &MockDomainError{BaseError: baseErr}
				return domainErr.Wrap(errors.New("wrapped error"))
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
			name: "should return nil when DomainError not has BaseError",
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
