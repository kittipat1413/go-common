package errors

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrBaseErrorCreationFailed is returned when BaseError creation fails due to invalid parameters.
// This typically occurs when the error code doesn't follow the required format or contains
// invalid category codes.
var ErrBaseErrorCreationFailed = errors.New("BaseError creation failed")

// BaseError provides a default implementation of the DomainError interface.
// It serves as the foundational error type that can be embedded in other error types
// to provide consistent error handling behavior across the application.
//
// BaseError implements the following interfaces:
//   - error (standard Go error interface)
//   - DomainError (custom domain error interface)
//
// The error code follows a structured format 'xyyzzz' where:
//   - 'x' (1st digit): Main error category
//   - 'yy' (2nd-3rd digits): Subcategory within the main category
//   - 'zzz' (4th-6th digits): Specific error detail identifier
type BaseError struct {
	code     string
	message  string
	httpCode int
	data     interface{}
}

// GetHTTPCode returns the HTTP status code associated with this error.
// The HTTP code is automatically determined based on the error category
// during BaseError creation.
func (e *BaseError) GetHTTPCode() int {
	return e.httpCode
}

// Code returns the full error code with any configured prefix.
// This method calls GetFullCode() to ensure consistent code formatting
// across the application.
//
// Returns:
//   - string: The complete error code with prefix (e.g., "APP-400001")
//
// Example:
//
//	err := NewBaseError("400001", "Invalid input", nil)
//	code := err.Code() // Returns "APP-400001" (assuming "APP-" prefix)
func (e *BaseError) Code() string {
	return GetFullCode(e.code)
}

// GetMessage returns the human-readable error message.
// If no custom message was provided during creation, this returns
// the default message associated with the error code.
//
// Returns:
//   - string: The error message for display to users or logging
//
// Example:
//
//	err := NewBaseError("400001", "Custom message", nil)
//	message := err.GetMessage() // Returns "Custom message"
func (e *BaseError) GetMessage() string {
	return e.message
}

// GetData returns any additional data associated with this error.
// This can be used to provide context-specific information such as
// validation details, request parameters, or debugging information.
//
// Returns:
//   - interface{}: The additional data, or nil if no data was provided
//
// Example:
//
//	validationData := map[string]string{"field": "email", "reason": "invalid format"}
//	err := NewBaseError("400001", "Validation failed", validationData)
//	data := err.GetData() // Returns the validation data map
func (e *BaseError) GetData() interface{} {
	return e.data
}

// Error implements the standard Go error interface.
// It returns the error message, making BaseError compatible with
// all Go error handling patterns.
//
// Returns:
//   - string: The error message (same as GetMessage())
//
// Example:
//
//	err := NewBaseError("500001", "Database connection failed", nil)
//	fmt.Println(err.Error()) // Prints: "Database connection failed"
func (e *BaseError) Error() string {
	return e.GetMessage()
}

// NewBaseError creates a new BaseError instance with validation and automatic defaults.
// If the message parameter is empty, it uses the default message from the error code
// configuration. The HTTP status code is automatically determined based on the error category.
//
// Error Code Format:
// The error code must follow the 'xyyzzz' convention (exactly 6 characters):
//   - 'x' (1st digit): Main error category (must be valid as defined in validCategories)
//   - 'yy' (2nd-3rd digits): Subcategory within the main category
//   - 'zzz' (4th-6th digits): Specific error detail identifier
//
// Parameters:
//   - code: The 6-character error code following the 'xyyzzz' format
//   - message: Human-readable error message (empty string uses default message)
//   - data: Optional additional data to associate with the error
//
// Returns:
//   - *BaseError: The created BaseError instance
//   - error: ErrBaseErrorCreationFailed if validation fails
//
// Validation Rules:
//   - Code must be exactly 6 characters long
//   - First 3 characters ('xyy') must match a valid category
//   - Category must be defined in the system's category configuration
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

// ExtractBaseError attempts to extract a BaseError from any error type that embeds it.
// This function uses reflection to search for an embedded *BaseError field in the error's
// struct definition, supporting both pointer and non-pointer error types.
//
// The function performs the following checks:
//  1. Direct type assertion for *BaseError
//  2. Reflection-based search for embedded *BaseError fields (one layer deep)
//  3. Support for both pointer and value receivers
//
// Parameters:
//   - err: The error to extract BaseError from (can be nil)
//
// Returns:
//   - *BaseError: The extracted BaseError if found, nil otherwise
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
