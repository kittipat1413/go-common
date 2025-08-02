package errors

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
		return err
	}
	return &ClientError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *ClientError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **ClientError:
		*t = e
		return true
	case *ClientError:
		*t = *e
		return true
	default:
		return false
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
		return err
	}
	return &BadRequestError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *BadRequestError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **BadRequestError:
		*t = e
		return true
	case *BadRequestError:
		*t = *e
		return true
	default:
		return false
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
		return err
	}
	return &NotFoundError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *NotFoundError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **NotFoundError:
		*t = e
		return true
	case *NotFoundError:
		*t = *e
		return true
	default:
		return false
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
		return err
	}
	return &ConflictError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *ConflictError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **ConflictError:
		*t = e
		return true
	case *ConflictError:
		*t = *e
		return true
	default:
		return false
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
		return err
	}
	return &UnprocessableEntityError{
		BaseError: baseErr,
	}
}

// As checks if the error can be assigned to the target interface.
// It supports both pointer and non-pointer types for the target.
func (e *UnprocessableEntityError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	switch t := target.(type) {
	case **UnprocessableEntityError:
		*t = e
		return true
	case *UnprocessableEntityError:
		*t = *e
		return true
	default:
		return false
	}
}

// Additional error types can be added here following the same pattern.
