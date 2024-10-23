package errors

import (
	"fmt"
	"reflect"
)

// BaseError provides a default implementation of the DomainError interface. It can be embedded in other error types to avoid code duplication.
type BaseError struct {
	code     string
	message  string
	httpCode int
	data     interface{}
}

func (e *BaseError) Code() string {
	return GetFullCode(e.code)
}

func (e *BaseError) GetMessage() string {
	return e.message
}

func (e *BaseError) GetHTTPCode() int {
	return e.httpCode
}

func (e *BaseError) Error() string {
	return e.GetMessage()
}

func (e *BaseError) GetData() interface{} {
	return e.data
}

/*
NewBaseError creates a new BaseError instance. If message is empty, it uses the default message from getDefaultMessages().

The error code should follow the 'xyzzz' convention:
  - 'x' (first digit): main error category.
  - 'y' (second digit): subcategory.
  - 'zzz' (last three digits): specific error detail.

The function validates the 'xy' part of the error code against valid categories and the HTTP status code against the category.
*/
func NewBaseError(code, message string, httpCode int, data interface{}) (*BaseError, error) {
	// Validate the error code length
	if len(code) < 5 {
		return nil, fmt.Errorf("error creation failed: invalid error code format %s", code)
	}

	// Extract 'xy' from the error code
	xy := code[:2]

	// Validate 'xy' category
	if !IsValidCategory(xy) {
		return nil, fmt.Errorf("error creation failed: invalid category %s", xy)
	}
	// Validate the HTTP status code against the category
	if httpCode != GetCategoryHTTPStatus(xy) {
		return nil, fmt.Errorf("error creation failed: invalid HTTP status code %d for category %s", httpCode, xy)
	}
	// Use default message if none provided
	if message == "" {
		message = getDefaultMessages(code)
	}

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
