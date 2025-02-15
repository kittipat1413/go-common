package errors

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrBaseErrorCreationFailed = errors.New("BaseError creation failed")

// BaseError provides a default implementation of the DomainError interface. It can be embedded in other error types to avoid code duplication.
type BaseError struct {
	code     string
	message  string
	httpCode int
	data     interface{}
}

func (e *BaseError) GetHTTPCode() int {
	return e.httpCode
}

func (e *BaseError) Code() string {
	return GetFullCode(e.code)
}

func (e *BaseError) GetMessage() string {
	return e.message
}

func (e *BaseError) GetData() interface{} {
	return e.data
}

func (e *BaseError) Error() string {
	return e.GetMessage()
}

/*
NewBaseError creates a new BaseError instance. If the message is empty, it uses the default message
from `getDefaultMessages()` based on the error code.

The error code should follow the 'xyyzzz' convention:
  - 'x' (first digit): main error category.
  - 'yy' (second digit): subcategory.
  - 'zzz' (last three digits): specific error detail.

**Note:** The 'xyy' prefix of the code must match a valid category defined in `validCategories`.
*/
func NewBaseError(code, message string, data interface{}) (*BaseError, error) {
	// Validate the error code length
	const codeLength = 6
	if len(code) != codeLength {
		return nil, fmt.Errorf("%w: error code '%s' must be exactly %d characters", ErrBaseErrorCreationFailed, code, codeLength)
	}

	// Extract the category 'xyy' from the error code
	xyy := code[:3]

	// Validate the extracted category
	if !IsValidCategory(xyy) {
		return nil, fmt.Errorf("%w: invalid category '%s' in code '%s'", ErrBaseErrorCreationFailed, xyy, code)
	}

	// Determine the HTTP status code for the category
	httpCode := GetCategoryHTTPStatus(xyy)

	// Assign default message if no custom message is provided
	if message == "" {
		message = getDefaultMessages(code)
	}

	// Create and return the BaseError instance
	return &BaseError{
		code:     code,
		message:  message,
		httpCode: httpCode,
		data:     data,
	}, nil
}

// ExtractBaseError attempts to extract the BaseError from the error's concrete type.
// It supports both pointer and non-pointer types, checking if the error directly embeds a *BaseError (one layer deep).
func ExtractBaseError(err error) *BaseError {
	if err == nil {
		return nil
	}

	// Check if err is a *BaseError directly
	if baseErr, ok := err.(*BaseError); ok {
		return baseErr
	}

	// Get the concrete value of the error
	errValue := reflect.ValueOf(err)
	if errValue.Kind() == reflect.Ptr {
		if errValue.IsNil() {
			return nil
		}
		// Dereference the pointer to get the underlying struct
		errValue = errValue.Elem()
	}

	// Ensure the underlying type is a struct
	if errValue.Kind() != reflect.Struct {
		return nil
	}

	// Iterate over the fields of the struct
	for i := 0; i < errValue.NumField(); i++ {
		field := errValue.Field(i)
		fieldType := errValue.Type().Field(i)

		// Check if the field is embedded (anonymous) and of type *BaseError
		if fieldType.Anonymous && field.Type() == reflect.TypeOf((*BaseError)(nil)) {
			// Extract the *BaseError value
			if baseErr, ok := field.Interface().(*BaseError); ok {
				return baseErr
			}
		}
	}

	return nil
}
