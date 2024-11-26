package errors

/*
DomainError is the interface that all custom errors in the framework implement. It provides methods to retrieve the error code, message, HTTP status code and additional data.

Error Code Convention: The error code follows the format 'xyyzzz' ([x][yy][zzz])
  - 'x' (first digit) represents the main error category (e.g., '4' for Client Error).
  - 'yy' (second digit) represents the subcategory (e.g., '01' for Bad Request).
  - 'zzz' (last three digits) provide specific details about the error.

This convention helps in categorizing errors consistently across services.
*/
type DomainError interface {
	// GetHTTPCode returns the HTTP status code associated with the error.
	GetHTTPCode() int

	// Code returns the full error code, including the service prefix (e.g., 'SVC-401000').
	Code() string

	// GetMessage returns the error message.
	GetMessage() string

	// GetData returns any additional data associated with the error.
	GetData() interface{}

	// Error implements the standard error interface.
	Error() string
}
