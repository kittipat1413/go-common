package errors

type InternalServerError struct {
	*BaseError
}

// NewInternalServerError creates a new InternalServerError instance using the generic internal error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewInternalServerError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericInternalServerError,
		message,
		data,
	)
	if err != nil {
		return err
	}
	return &InternalServerError{
		BaseError: baseErr,
	}
}

type DatabaseError struct {
	*BaseError
}

// NewDatabaseError creates a new DatabaseError instance using the generic database error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewDatabaseError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericDatabaseError,
		message,
		data,
	)
	if err != nil {
		return err
	}
	return &DatabaseError{
		BaseError: baseErr,
	}
}

type ThirdPartyError struct {
	*BaseError
}

// NewThirdPartyError creates a new ThirdPartyError instance using the generic third-party error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewThirdPartyError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericThirdPartyError,
		message,
		data,
	)
	if err != nil {
		return err
	}
	return &ThirdPartyError{
		BaseError: baseErr,
	}
}

type ServiceUnavailableError struct {
	*BaseError
}

// NewServiceUnavailableError creates a new ServiceUnavailableError instance using the generic service unavailable error code.
// If the `message` parameter is an empty string (""), the default message for the error code will be used.
func NewServiceUnavailableError(message string, data interface{}) error {
	baseErr, err := NewBaseError(
		StatusCodeGenericServiceUnavailableError,
		message,
		data,
	)
	if err != nil {
		return err
	}
	return &ServiceUnavailableError{
		BaseError: baseErr,
	}
}

// Additional error types can be added here following the same pattern.
