package errors

import "net/http"

type InternalServerError struct {
	*BaseError
}

// NewInternalServerError creates a new InternalServerError. It uses the generic internal error code and default message.
func NewInternalServerError(data interface{}) (*InternalServerError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericInternalError,
		"", // Empty message to use the default message
		http.StatusInternalServerError,
		data,
	)
	if err != nil {
		return nil, err
	}
	return &InternalServerError{
		BaseError: baseErr,
	}, nil
}

// Additional error types can be added here following the same pattern.
