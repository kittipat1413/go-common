package errors

import "net/http"

// errorCategory represents a complete definition of an error category including
// its code, human-readable description, and corresponding HTTP status code.
// This structure enables automatic HTTP status mapping and error categorization.
type errorCategory struct {
	// CategoryCode is the 3-character category identifier ('xyy' from 'xyyzzz' format)
	CategoryCode string

	// Description is a human-readable name for the category
	Description string

	// HTTPStatus is the HTTP status code that should be returned for errors in this category
	HTTPStatus int
}

// validCategories defines all recognized error categories in the system.
// Each category maps a 3-character code to its metadata including description and HTTP status.
//
// Category Structure:
//   - Key: 3-character category code (first 3 digits of error codes)
//   - Value: errorCategory with full metadata
//
// Adding New Categories:
// To add a new category, define the error code constant in error_code.go,
// then add an entry to this map with appropriate HTTP status mapping.
var validCategories = map[string]errorCategory{
	// Success Categories - HTTP 200 family
	StatusCodeSuccess[:3]:        {CategoryCode: StatusCodeSuccess[:3], Description: "Success", HTTPStatus: http.StatusOK},
	StatusCodePartialSuccess[:3]: {CategoryCode: StatusCodePartialSuccess[:3], Description: "Partial Success", HTTPStatus: http.StatusOK},
	StatusCodeAccepted[:3]:       {CategoryCode: StatusCodeAccepted[:3], Description: "Accepted", HTTPStatus: http.StatusAccepted},

	// Client Error Categories - HTTP 400 family
	StatusCodeGenericClientError[:3]:              {CategoryCode: StatusCodeGenericClientError[:3], Description: "Client Error", HTTPStatus: http.StatusBadRequest},
	StatusCodeGenericBadRequestError[:3]:          {CategoryCode: StatusCodeGenericBadRequestError[:3], Description: "Bad Request", HTTPStatus: http.StatusBadRequest},
	StatusCodeGenericNotFoundError[:3]:            {CategoryCode: StatusCodeGenericNotFoundError[:3], Description: "Not Found", HTTPStatus: http.StatusNotFound},
	StatusCodeGenericConflictError[:3]:            {CategoryCode: StatusCodeGenericConflictError[:3], Description: "Conflict", HTTPStatus: http.StatusConflict},
	StatusCodeGenericUnprocessableEntityError[:3]: {CategoryCode: StatusCodeGenericUnprocessableEntityError[:3], Description: "Unprocessable Entity", HTTPStatus: http.StatusUnprocessableEntity},

	// Server Error Categories - HTTP 500 family
	StatusCodeGenericInternalServerError[:3]:     {CategoryCode: StatusCodeGenericInternalServerError[:3], Description: "Internal Error", HTTPStatus: http.StatusInternalServerError},
	StatusCodeGenericDatabaseError[:3]:           {CategoryCode: StatusCodeGenericDatabaseError[:3], Description: "Database Error", HTTPStatus: http.StatusInternalServerError},
	StatusCodeGenericThirdPartyError[:3]:         {CategoryCode: StatusCodeGenericThirdPartyError[:3], Description: "Third-party Error", HTTPStatus: http.StatusBadGateway},
	StatusCodeGenericServiceUnavailableError[:3]: {CategoryCode: StatusCodeGenericServiceUnavailableError[:3], Description: "Service Unavailable", HTTPStatus: http.StatusServiceUnavailable},

	// Authentication/Authorization Categories - HTTP 401/403
	StatusCodeGenericAuthError[:3]:         {CategoryCode: StatusCodeGenericAuthError[:3], Description: "Security Error", HTTPStatus: http.StatusUnauthorized},
	StatusCodeGenericUnauthorizedError[:3]: {CategoryCode: StatusCodeGenericUnauthorizedError[:3], Description: "Unauthorized", HTTPStatus: http.StatusUnauthorized},
	StatusCodeGenericForbiddenError[:3]:    {CategoryCode: StatusCodeGenericForbiddenError[:3], Description: "Forbidden", HTTPStatus: http.StatusForbidden},
}

// IsValidCategory validates whether a 3-character category code ('xyy') is recognized by the system.
// This function is used during BaseError creation to ensure error codes follow valid conventions.
//
// The function performs a simple map lookup to determine if the category exists in the
// predefined validCategories map. This validation helps catch typos and ensures
// consistent error categorization across the application.
//
// Parameters:
//   - xyy: The 3-character category code to validate (e.g., "400", "501", "902")
//
// Returns:
//   - bool: true if the category is valid and recognized, false otherwise
func IsValidCategory(xyy string) bool {
	_, exists := validCategories[xyy]
	return exists
}

// GetCategoryDescription returns a human-readable description for a given category code.
// This function is primarily used for logging, debugging, and error reporting purposes
// where a descriptive category name is more useful than the numeric code.
//
// Parameters:
//   - xyy: The 3-character category code to look up (e.g., "400", "501", "902")
//
// Returns:
//   - string: Human-readable description of the category, or "Unknown Category" if not found
func GetCategoryDescription(xyy string) string {
	if desc, exists := validCategories[xyy]; exists {
		return desc.Description
	}
	return "Unknown Category"
}

// GetCategoryHTTPStatus returns the appropriate HTTP status code for a given category code.
// This function provides the automatic HTTP status code mapping that enables consistent
// API responses across services without manual status code management.
//
// If the category code is not recognized, the function returns HTTP 500 (Internal Server Error)
// as a safe fallback, indicating that something unexpected occurred in the error handling system.
//
// Parameters:
//   - xyy: The 3-character category code to look up (e.g., "400", "501", "902")
//
// Returns:
//   - int: HTTP status code corresponding to the category, or 500 if category is unknown
func GetCategoryHTTPStatus(xyy string) int {
	if desc, exists := validCategories[xyy]; exists {
		return desc.HTTPStatus
	}
	return http.StatusInternalServerError
}
