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

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *InternalServerError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **InternalServerError:
		*t = e
		return true
	case *InternalServerError:
		*t = *e
		return true
	default:
		return false
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

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *DatabaseError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **DatabaseError:
		*t = e
		return true
	case *DatabaseError:
		*t = *e
		return true
	default:
		return false
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

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *ThirdPartyError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **ThirdPartyError:
		*t = e
		return true
	case *ThirdPartyError:
		*t = *e
		return true
	default:
		return false
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

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *ServiceUnavailableError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **ServiceUnavailableError:
		*t = e
		return true
	case *ServiceUnavailableError:
		*t = *e
		return true
	default:
		return false
	}
}

// Additional error types can be added here following the same pattern.
