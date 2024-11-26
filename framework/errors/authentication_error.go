package errors

import "fmt"

type AuthenticationError struct {
	*BaseError
}

// NewAuthenticationError creates a new AuthenticationError instance using the generic authentication error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewAuthenticationError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericAuthError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &AuthenticationError{
		BaseError: baseErr,
	}
}

type UnauthorizedError struct {
	*BaseError
}

// NewUnauthorizedError creates a new UnauthorizedError instance using the generic unauthorized error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewUnauthorizedError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericUnauthorizedError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &UnauthorizedError{
		BaseError: baseErr,
	}
}

type ForbiddenError struct {
	*BaseError
}

// NewForbiddenError creates a new ForbiddenError instance using the generic forbidden error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewForbiddenError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericForbiddenError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &ForbiddenError{
		BaseError: baseErr,
	}
}

// Additional error types can be added here following the same pattern.
