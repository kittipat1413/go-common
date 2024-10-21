package errors

import "net/http"

type AuthenticationError struct {
	*BaseError
}

// NewAuthenticationError creates a new AuthenticationError. It uses the generic authentication error code and default message.
func NewAuthenticationError() (*AuthenticationError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericAuthError,
		"", // Empty message to use the default message
		http.StatusUnauthorized,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &AuthenticationError{
		BaseError: baseErr,
	}, nil
}

type UnauthorizedError struct {
	*BaseError
}

// NewUnauthorizedError creates a new UnauthorizedError. It uses the generic unauthorized error code and default message.
func NewUnauthorizedError() (*UnauthorizedError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericUnauthorized,
		"", // Empty message to use the default message
		http.StatusUnauthorized,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &UnauthorizedError{
		BaseError: baseErr,
	}, nil
}

type ForbiddenError struct {
	*BaseError
}

// NewForbiddenError creates a new ForbiddenError. It uses the generic forbidden error code and default message.
func NewForbiddenError() (*ForbiddenError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericForbidden,
		"", // Empty message to use the default message
		http.StatusForbidden,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &ForbiddenError{
		BaseError: baseErr,
	}, nil
}

// Additional error types can be added here following the same pattern.
