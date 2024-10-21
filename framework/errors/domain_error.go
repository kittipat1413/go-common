package errors

/*
DomainError is the interface that all custom errors in the framework implement. It provides methods to retrieve the error code, message, HTTP status code, additional data, and to wrap underlying errors.

Error Code Convention: The error code follows the format 'xyzzz'
  - 'x' (first digit) represents the main error category (e.g., '2' for Client Error).
  - 'y' (second digit) represents the subcategory (e.g., '1' for Invalid Parameters).
  - 'zzz' (last three digits) provide specific details about the error.

This convention helps in categorizing errors consistently across services.
*/
type DomainError interface {
	// Code returns the full error code, including the service prefix (e.g., 'USER-SVC-21001').
	Code() string

	// GetMessage returns the error message.
	GetMessage() string

	// GetHTTPCode returns the HTTP status code associated with the error.
	GetHTTPCode() int

	// Error implements the standard error interface.
	Error() string

	// GetData returns any additional data associated with the error.
	GetData() interface{}

	// Wrap allows wrapping an underlying error with the current error.
	Wrap(err error) error
}
