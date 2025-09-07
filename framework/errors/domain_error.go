package errors

// DomainError is the interface that all custom errors in the framework implement.
// It provides a standardized contract for retrieving error information including
// structured error codes, human-readable messages, HTTP status codes, and additional context data.
//
// Error Code Convention:
// The error code follows the format 'xyyzzz' where:
//   - 'x' (1st digit): Main error category
//   - 'yy' (2nd-3rd digits): Subcategory within the main category
//   - 'zzz' (4th-6th digits): Specific error identifier within the subcategory
//
// Implementation Requirements:
// All implementations should embed *BaseError to ensure consistent behavior
// and enable automatic error extraction via ExtractBaseError().
type DomainError interface {
	// GetHTTPCode returns the HTTP status code associated with the error.
	// The status code is automatically determined based on the error category
	// and follows standard HTTP status code conventions.
	GetHTTPCode() int

	// Code returns the full error code including the configured service prefix.
	// The format is typically "{PREFIX}-{ERRORCODE}" where PREFIX is set via
	// SetServicePrefix() and ERRORCODE follows the 'xyyzzz' convention.
	Code() string

	// GetMessage returns a human-readable error message suitable for logging
	// or displaying to users. If no custom message was provided during error
	// creation, this returns the default message associated with the error code.
	GetMessage() string

	// GetData returns any additional structured data associated with the error.
	// This can include validation details, request context, debugging information,
	// or any other relevant metadata that might be useful for error handling,
	// logging, or client responses.
	GetData() interface{}

	// Error implements the standard Go error interface, making DomainError
	// compatible with all Go error handling patterns and libraries.
	Error() string
}
