package errors

import "net/http"

type BadRequestError struct {
	*BaseError
}

// NewBadRequestError creates a new BadRequestError. It uses the generic unauthorized error code and default message.
func NewBadRequestError() (*BadRequestError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericBadRequestError,
		"", // Empty message to use the default message
		http.StatusBadRequest,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &BadRequestError{
		BaseError: baseErr,
	}, nil
}

type InvalidParametersError struct {
	*BaseError
}

// NewInvalidParametersError creates a new InvalidParametersError. It uses the generic invalid parameters error code and default message.
func NewInvalidParametersError() (*InvalidParametersError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericInvalidParameters,
		"", // Empty message to use the default message
		http.StatusBadRequest,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &InvalidParametersError{
		BaseError: baseErr,
	}, nil
}

type DuplicatedEntryError struct {
	*BaseError
}

// NewDuplicatedEntryError creates a new DuplicatedEntryError. It uses the generic duplicated entry error code and default message.
func NewDuplicatedEntryError() (*DuplicatedEntryError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericDuplicatedEntry,
		"", // Empty message to use the default message
		http.StatusConflict,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &DuplicatedEntryError{
		BaseError: baseErr,
	}, nil
}

type NotFoundError struct {
	*BaseError
}

// NewNotFoundError creates a new NotFoundError. It uses the generic not found error code and default message.
func NewNotFoundError() (*NotFoundError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericNotFoundError,
		"", // Empty message to use the default message
		http.StatusNotFound,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &NotFoundError{
		BaseError: baseErr,
	}, nil
}

type UnprocessableEntityError struct {
	*BaseError
}

// NewUnprocessableEntityError creates a new UnprocessableEntityError. It uses the generic unprocessable entity error code and default message.
func NewUnprocessableEntityError() (*UnprocessableEntityError, error) {
	baseErr, err := NewBaseError(
		StatusCodeGenericUnprocessableEntity,
		"", // Empty message to use the default message
		http.StatusUnprocessableEntity,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &UnprocessableEntityError{
		BaseError: baseErr,
	}, nil
}

// Additional error types can be added here following the same pattern.
