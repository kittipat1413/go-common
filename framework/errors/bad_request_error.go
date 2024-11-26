package errors

import "fmt"

type ClientError struct {
	*BaseError
}

// NewClientError creates a new ClientError instance using the generic client error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewClientError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericClientError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &ClientError{
		BaseError: baseErr,
	}
}

type BadRequestError struct {
	*BaseError
}

// NewBadRequestError creates a new BadRequestError instance using the generic bad request error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewBadRequestError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericBadRequestError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &BadRequestError{
		BaseError: baseErr,
	}
}

type NotFoundError struct {
	*BaseError
}

// NewNotFoundError creates a new NotFoundError instance using the generic not found error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewNotFoundError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericNotFoundError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &NotFoundError{
		BaseError: baseErr,
	}
}

type ConflictError struct {
	*BaseError
}

// NewConflictError creates a new ConflictError instance using the generic conflict error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewConflictError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericConflictError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &ConflictError{
		BaseError: baseErr,
	}
}

type UnprocessableEntityError struct {
	*BaseError
}

// NewUnprocessableEntityError creates a new UnprocessableEntityError instance using the generic unprocessable entity error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewUnprocessableEntityError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericUnprocessableEntityError,
		message,
		data,
	)
	if err != nil {
		return fmt.Errorf("BaseError creation failed: %w", err)
	}
	return &UnprocessableEntityError{
		BaseError: baseErr,
	}
}

// Additional error types can be added here following the same pattern.
