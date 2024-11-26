package errors

/*
Error code constants following the 'xyyzzz' ([x][yy][zzz]) convention.
  - 'x' - Main category,
  - 'yy' - Subcategory,
  - 'zzz' - Specific error code within the subcategory.
*/
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

	// Internal Errors (5yyzzz)
	StatusCodeGenericInternalServerError = "500000" // General Internal Error
	StatusCodeGenericDatabaseError       = "501000" // Database Error
	StatusCodeGenericThirdPartyError     = "502000" // Third-party Error

	// Authentication and Authorization Errors (9yyzzz)
	StatusCodeGenericAuthError         = "900000" // General Authentication Error
	StatusCodeGenericUnauthorizedError = "901000" // Unauthorized (e.g., missing or invalid token)
	StatusCodeGenericForbiddenError    = "902000" // Forbidden (e.g., insufficient permissions)
)

// GetFullCode constructs the full error code with the service prefix.
// If servicePrefix is "SVC" and code is "200000", it returns "SVC-200000".
func GetFullCode(code string) string {
	return servicePrefix + "-" + code
}

// errorCodeToMessages maps error codes to their default messages. If a specific message is not provided when creating an error, the default message from this map will be used.
var errorCodeToMessages = map[string]string{
	// Success
	StatusCodeSuccess:        "Operation completed successfully.",
	StatusCodePartialSuccess: "Operation partially completed.",
	StatusCodeAccepted:       "Request accepted. Processing is ongoing.",
	// Client Errors
	StatusCodeGenericClientError:              "An error occurred while processing the request.",
	StatusCodeGenericBadRequestError:          "The request was invalid or cannot be served.",
	StatusCodeGenericConflictError:            "The request could not be completed due to a conflict with the current state of the resource.",
	StatusCodeGenericNotFoundError:            "The requested resource could not be found.",
	StatusCodeGenericUnprocessableEntityError: "The request could not be processed due to semantic errors.",
	// Internal Errors
	StatusCodeGenericInternalServerError: "An internal server error occurred. Please try again later.",
	StatusCodeGenericDatabaseError:       "A database error occurred while processing the request.",
	StatusCodeGenericThirdPartyError:     "An error occurred while communicating with an external service.",
	// Security Errors
	StatusCodeGenericAuthError:         "Authentication failed. Please verify your credentials.",
	StatusCodeGenericUnauthorizedError: "Access denied. You are not authorized to perform this action.",
	StatusCodeGenericForbiddenError:    "Access to this resource is forbidden.",
}

// getDefaultMessage returns the default message for the given error code. If the code is not found in the map, it returns a generic error message.
func getDefaultMessages(code string) string {
	if defaultMsg, exists := errorCodeToMessages[code]; exists {
		return defaultMsg
	} else {
		return "An unexpected error occurred. Please try again later."
	}
}
