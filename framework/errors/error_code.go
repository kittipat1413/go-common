package errors

// Error code constants following the 'xyyzzz' ([x][yy][zzz]) convention.
// Each constant represents a specific error scenario with automatic HTTP status mapping.
//
// Code Structure:
//   - 'x' (1st digit): Main category
//   - 'yy' (2nd-3rd digits): Subcategory within main category
//   - 'zzz' (4th-6th digits): Specific error identifier
//
// Usage:
// These constants should be used when creating BaseError instances to ensure
// consistency across services and automatic HTTP status code mapping.
const (
	// Success (2yyzzz)
	StatusCodeSuccess        = "200000" // General Success
	StatusCodePartialSuccess = "201000" // Partial Success (e.g., batch processing)
	StatusCodeAccepted       = "202000" // Accepted (e.g., long-running task queued)

	// Client Errors (4yyzzz)
	StatusCodeGenericClientError              = "400000" // General Client Error
	StatusCodeGenericBadRequestError          = "401000" // Bad Request (e.g., missing or invalid parameters)
	StatusCodeGenericNotFoundError            = "402000" // Not Found (e.g., resource not found)
	StatusCodeGenericConflictError            = "403000" // Conflict (e.g., resource already exists)
	StatusCodeGenericUnprocessableEntityError = "404000" // Unprocessable Entity (e.g., validation error)

	// Server Errors (5yyzzz)
	StatusCodeGenericInternalServerError     = "500000" // General Internal Server Error
	StatusCodeGenericDatabaseError           = "501000" // Database Error
	StatusCodeGenericThirdPartyError         = "502000" // Third-party Error
	StatusCodeGenericServiceUnavailableError = "503000" // Service Unavailable (e.g., maintenance mode)

	// Authentication and Authorization Errors (9yyzzz)
	StatusCodeGenericAuthError         = "900000" // General Authentication Error
	StatusCodeGenericUnauthorizedError = "901000" // Unauthorized (e.g., missing or invalid token)
	StatusCodeGenericForbiddenError    = "902000" // Forbidden (e.g., insufficient permissions)
)

// GetFullCode constructs the complete error code by combining the configured
// service prefix with the base error code using a hyphen separator.
//
// The service prefix is set globally via SetServicePrefix() and defaults to "ERR".
// This function is used internally by the BaseError implementation to generate
// the final error code returned by the Code() method.
//
// Returns:
//   - string: Complete error code in format "{PREFIX}-{CODE}"
//
// Examples:
//   - GetFullCode("400001") → "ERR-400001" (default prefix)
//   - GetFullCode("500001") → "USER-SVC-500001" (custom prefix "USER-SVC")
func GetFullCode(code string) string {
	return servicePrefix + "-" + code
}

// errorCodeToMessages maps error codes to their default human-readable messages.
// These messages are used when no custom message is provided during error creation.
//
// Customization:
// Applications can override these defaults by providing custom messages
// when creating BaseError instances via NewBaseError().
var errorCodeToMessages = map[string]string{
	// Success
	StatusCodeSuccess:        "Operation completed successfully.",
	StatusCodePartialSuccess: "Operation partially completed.",
	StatusCodeAccepted:       "Request accepted. Processing is ongoing.",
	// Client Errors
	StatusCodeGenericClientError:              "An error occurred while processing the request.",
	StatusCodeGenericBadRequestError:          "The request was invalid or cannot be served.",
	StatusCodeGenericNotFoundError:            "The requested resource could not be found.",
	StatusCodeGenericConflictError:            "The request could not be completed due to a conflict with the current state of the resource.",
	StatusCodeGenericUnprocessableEntityError: "The request could not be processed due to semantic errors.",
	// Internal Errors
	StatusCodeGenericInternalServerError:     "An internal server error occurred. Please try again later.",
	StatusCodeGenericDatabaseError:           "A database error occurred while processing the request.",
	StatusCodeGenericThirdPartyError:         "An error occurred while communicating with an external service.",
	StatusCodeGenericServiceUnavailableError: "The service is currently unavailable. Please try again later.",
	// Security Errors
	StatusCodeGenericAuthError:         "Authentication failed. Please verify your credentials.",
	StatusCodeGenericUnauthorizedError: "Access denied. You are not authorized to perform this action.",
	StatusCodeGenericForbiddenError:    "Access to this resource is forbidden.",
}

// getDefaultMessages retrieves the default message for a given error code.
// This function is used internally by NewBaseError when no custom message is provided.
//
// If the error code is not found in the predefined messages map, it returns a
// generic fallback message to ensure all errors have meaningful messages.
//
// Parameters:
//   - code: The error code to look up (format: 'xyyzzz')
//
// Returns:
//   - string: Default message for the code, or generic message if code not found
func getDefaultMessages(code string) string {
	if defaultMsg, exists := errorCodeToMessages[code]; exists {
		return defaultMsg
	} else {
		return "An unexpected error occurred. Please try again later."
	}
}
