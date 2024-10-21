package errors

/*
Error code constants following the 'xyzzz' convention.
  - 'x' - Main category,
  - 'y' - Subcategory,
  - 'zzz' - Specific error code.
*/
const (
	// Success code
	StatusCodeSuccess = "10000"

	// Client Errors (20xxx)
	StatusCodeGenericBadRequestError     = "20000"
	StatusCodeGenericInvalidParameters   = "21000"
	StatusCodeGenericDuplicatedEntry     = "22000"
	StatusCodeGenericNotFoundError       = "23000"
	StatusCodeGenericUnprocessableEntity = "24000"

	// Internal Errors (50xxx)
	StatusCodeGenericInternalError   = "50000"
	StatusCodeGenericDatabaseError   = "51000"
	StatusCodeGenericThirdPartyError = "52000"

	// Security Errors (90xxx)
	StatusCodeGenericAuthError    = "90000"
	StatusCodeGenericUnauthorized = "91000"
	StatusCodeGenericForbidden    = "92000"
)

// errorCodeToMessages maps error codes to their default messages. If a specific message is not provided when creating an error, the default message from this map will be used.
var errorCodeToMessages = map[string]string{
	StatusCodeSuccess:                    "success",
	StatusCodeGenericBadRequestError:     "bad request",
	StatusCodeGenericInvalidParameters:   "invalid parameters",
	StatusCodeGenericDuplicatedEntry:     "duplicated entry",
	StatusCodeGenericNotFoundError:       "not found",
	StatusCodeGenericUnprocessableEntity: "unprocessable entity",
	StatusCodeGenericInternalError:       "internal error",
	StatusCodeGenericDatabaseError:       "database error",
	StatusCodeGenericThirdPartyError:     "third-party error",
	StatusCodeGenericAuthError:           "authentication error",
	StatusCodeGenericUnauthorized:        "unauthorized",
	StatusCodeGenericForbidden:           "forbidden",
}

// getDefaultMessage returns the default message for the given error code. If the code is not found in the map, it returns a generic error message.
func getDefaultMessages(code string) string {
	if defaultMsg, exists := errorCodeToMessages[code]; exists {
		return defaultMsg
	} else {
		return "an error occurred"
	}
}
