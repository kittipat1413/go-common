package errors

import "net/http"

// errorCategory represents a category of error codes. It contains the category code, description, and HTTP status code.
type errorCategory struct {
	CategoryCode string
	Description  string
	HTTPStatus   int
}

// validCategories is a map of valid error categories. It contains the category code, description, and HTTP status code.
var validCategories = map[string]errorCategory{
	StatusCodeSuccess[:3]:                         {CategoryCode: StatusCodeSuccess[:3], Description: "Success", HTTPStatus: 200},
	StatusCodePartialSuccess[:3]:                  {CategoryCode: StatusCodePartialSuccess[:3], Description: "Partial Success", HTTPStatus: 200},
	StatusCodeAccepted[:3]:                        {CategoryCode: StatusCodeAccepted[:3], Description: "Accepted", HTTPStatus: 202},
	StatusCodeGenericClientError[:3]:              {CategoryCode: StatusCodeGenericClientError[:3], Description: "Client Error", HTTPStatus: 400},
	StatusCodeGenericBadRequestError[:3]:          {CategoryCode: StatusCodeGenericBadRequestError[:3], Description: "Bad Request", HTTPStatus: 400},
	StatusCodeGenericNotFoundError[:3]:            {CategoryCode: StatusCodeGenericNotFoundError[:3], Description: "Not Found", HTTPStatus: 404},
	StatusCodeGenericConflictError[:3]:            {CategoryCode: StatusCodeGenericConflictError[:3], Description: "Conflict", HTTPStatus: 409},
	StatusCodeGenericUnprocessableEntityError[:3]: {CategoryCode: StatusCodeGenericUnprocessableEntityError[:3], Description: "Unprocessable Entity", HTTPStatus: 422},
	StatusCodeGenericInternalServerError[:3]:      {CategoryCode: StatusCodeGenericInternalServerError[:3], Description: "Internal Error", HTTPStatus: 500},
	StatusCodeGenericDatabaseError[:3]:            {CategoryCode: StatusCodeGenericDatabaseError[:3], Description: "Database Error", HTTPStatus: 500},
	StatusCodeGenericThirdPartyError[:3]:          {CategoryCode: StatusCodeGenericThirdPartyError[:3], Description: "Third-party Error", HTTPStatus: 502},
	StatusCodeGenericServiceUnavailableError[:3]:  {CategoryCode: StatusCodeGenericServiceUnavailableError[:3], Description: "Service Unavailable", HTTPStatus: 503},
	StatusCodeGenericAuthError[:3]:                {CategoryCode: StatusCodeGenericAuthError[:3], Description: "Security Error", HTTPStatus: 401},
	StatusCodeGenericUnauthorizedError[:3]:        {CategoryCode: StatusCodeGenericUnauthorizedError[:3], Description: "Unauthorized", HTTPStatus: 401},
	StatusCodeGenericForbiddenError[:3]:           {CategoryCode: StatusCodeGenericForbiddenError[:3], Description: "Forbidden", HTTPStatus: 403},
}

// IsValidCategory validates the 'xyy' part of an error code. It returns true if the category exists, and false otherwise.
func IsValidCategory(xyy string) bool {
	_, exists := validCategories[xyy]
	return exists
}

// GetCategoryDescription returns the description of the 'xyy' category. If the category does not exist, it returns "Unknown Category".
func GetCategoryDescription(xyy string) string {
	if desc, exists := validCategories[xyy]; exists {
		return desc.Description
	}
	return "Unknown Category"
}

// GetCategoryHTTPStatus returns the HTTP status code of the 'xyy' category. If the category does not exist, it returns 500.
func GetCategoryHTTPStatus(xyy string) int {
	if desc, exists := validCategories[xyy]; exists {
		return desc.HTTPStatus
	}
	return http.StatusInternalServerError
}
