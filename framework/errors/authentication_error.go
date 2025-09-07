package errors

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
		return err
	}
	return &AuthenticationError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// Supports both pointer (**AuthenticationError) and value (*AuthenticationError) targets
// for compatibility with Go's errors.As function.
func (e *AuthenticationError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **AuthenticationError:
		*t = e
		return true
	case *AuthenticationError:
		*t = *e
		return true
	default:
		return false
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
		return err
	}
	return &UnauthorizedError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// Supports both pointer (**UnauthorizedError) and value (*UnauthorizedError) targets
// for compatibility with Go's errors.As function.
func (e *UnauthorizedError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **UnauthorizedError:
		*t = e
		return true
	case *UnauthorizedError:
		*t = *e
		return true
	default:
		return false
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
		return err
	}
	return &ForbiddenError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// Supports both pointer (**ForbiddenError) and value (*ForbiddenError) targets
// for compatibility with Go's errors.As function.
func (e *ForbiddenError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **ForbiddenError:
		*t = e
		return true
	case *ForbiddenError:
		*t = *e
		return true
	default:
		return false
	}
}

// Additional error types can be added here following the same pattern.
