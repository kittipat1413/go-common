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

type DatabaseError struct {
	*BaseError
}

// NewDatabaseError creates a new DatabaseError. It uses the generic database error code and default message.
func NewDatabaseError(data interface{}) (*DatabaseError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericDatabaseError,
		"", // Empty message to use the default message
		http.StatusInternalServerError,
		data,
	)
	if err != nil {
		return nil, err
	}
	return &DatabaseError{
		BaseError: baseErr,
	}, nil
}

type ThirdPartyError struct {
	*BaseError
}

// NewThirdPartyError creates a new ThirdPartyError. It uses the generic third-party error code and default message.
func NewThirdPartyError(data interface{}) (*ThirdPartyError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericThirdPartyError,
		"", // Empty message to use the default message
		http.StatusBadGateway,
		data,
	)
	if err != nil {
		return nil, err
	}
	return &ThirdPartyError{
		BaseError: baseErr,
	}, nil
}

// Additional error types can be added here following the same pattern.
