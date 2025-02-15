package errors_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	domain_error "github.com/kittipat1413/go-common/framework/errors"
)

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		expected bool
	}{
		{
			name:     "Valid category - BadRequest",
			category: domain_error.StatusCodeGenericBadRequestError[:3],
			expected: true,
		},
		{
			name:     "Valid category - Internal Server Error",
			category: domain_error.StatusCodeGenericInternalServerError[:3],
			expected: true,
		},
		{
			name:     "Valid category - Auth Error",
			category: domain_error.StatusCodeGenericAuthError[:3],
			expected: true,
		},
		{
			name:     "Invalid category - Random code",
			category: "999",
			expected: false,
		},
		{
			name:     "Invalid category - Empty string",
			category: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain_error.IsValidCategory(tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCategoryDescription(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		expectedOutput string
	}{
		{
			name:           "Valid category - BadRequest",
			category:       domain_error.StatusCodeGenericBadRequestError[:3],
			expectedOutput: "Bad Request",
		},
		{
			name:           "Valid category - Database Error",
			category:       domain_error.StatusCodeGenericDatabaseError[:3],
			expectedOutput: "Database Error",
		},
		{
			name:           "Valid category - Third-party Error",
			category:       domain_error.StatusCodeGenericThirdPartyError[:3],
			expectedOutput: "Third-party Error",
		},
		{
			name:           "Invalid category - Nonexistent",
			category:       "999",
			expectedOutput: "Unknown Category",
		},
		{
			name:           "Invalid category - Empty String",
			category:       "",
			expectedOutput: "Unknown Category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain_error.GetCategoryDescription(tt.category)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

func TestGetCategoryHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		expectedStatus int
	}{
		{
			name:           "Valid category - BadRequest",
			category:       domain_error.StatusCodeGenericBadRequestError[:3],
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid category - Not Found",
			category:       domain_error.StatusCodeGenericNotFoundError[:3],
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Valid category - Forbidden",
			category:       domain_error.StatusCodeGenericForbiddenError[:3],
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Valid category - Internal Server Error",
			category:       domain_error.StatusCodeGenericInternalServerError[:3],
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid category - Nonexistent",
			category:       "999",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid category - Empty String",
			category:       "",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain_error.GetCategoryHTTPStatus(tt.category)
			assert.Equal(t, tt.expectedStatus, result)
		})
	}
}
