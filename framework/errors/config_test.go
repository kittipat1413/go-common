package errors_test

import (
	"testing"

	"github.com/kittipat1413/go-common/framework/errors"
	"github.com/stretchr/testify/assert"
)

func TestServicePrefix(t *testing.T) {
	// Store original prefix to restore after tests
	originalPrefix := errors.GetServicePrefix()
	defer errors.SetServicePrefix(originalPrefix)

	tests := []struct {
		name           string
		inputPrefix    string
		expectedPrefix string
	}{
		{
			name:           "empty prefix should use default",
			inputPrefix:    "",
			expectedPrefix: errors.DefaultServicePrefix,
		},
		{
			name:           "should convert prefix to uppercase",
			inputPrefix:    "test-svc",
			expectedPrefix: "TEST-SVC",
		},
		{
			name:           "already uppercase prefix should remain unchanged",
			inputPrefix:    "USER-API",
			expectedPrefix: "USER-API",
		},
		{
			name:           "mixed case prefix should convert to uppercase",
			inputPrefix:    "Auth-Service",
			expectedPrefix: "AUTH-SERVICE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors.SetServicePrefix(tt.inputPrefix)
			assert.Equal(t, tt.expectedPrefix, errors.GetServicePrefix())
		})
	}
}

func TestGetServicePrefix_Default(t *testing.T) {
	// Store original prefix to restore after test
	originalPrefix := errors.GetServicePrefix()
	defer errors.SetServicePrefix(originalPrefix)

	// Reset to default
	errors.SetServicePrefix("")

	assert.Equal(t, errors.DefaultServicePrefix, errors.GetServicePrefix())
}
