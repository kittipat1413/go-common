package errors

import "net/http"

// errorCategory represents a category of error codes. It contains a description and an HTTP status code.
type errorCategory struct {
	Description string
	HTTPStatus  int
}

// validCategories maps 'xy' codes to their category descriptions. This is used to validate the main and subcategories of error codes.
var validCategories = map[string]errorCategory{
	StatusCodeSuccess[:2]:                    {Description: "Success", HTTPStatus: 200},
	StatusCodeGenericBadRequestError[:2]:     {Description: "Client Error", HTTPStatus: 400},
	StatusCodeGenericInvalidParameters[:2]:   {Description: "Invalid Parameters", HTTPStatus: 400},
	StatusCodeGenericDuplicatedEntry[:2]:     {Description: "Duplicated Entry", HTTPStatus: 409},
	StatusCodeGenericNotFoundError[:2]:       {Description: "Not Found", HTTPStatus: 404},
	StatusCodeGenericUnprocessableEntity[:2]: {Description: "Unprocessable Entity", HTTPStatus: 422},
	StatusCodeGenericInternalError[:2]:       {Description: "Internal Error", HTTPStatus: 500},
	StatusCodeGenericDatabaseError[:2]:       {Description: "Database Error", HTTPStatus: 500},
	StatusCodeGenericThirdPartyError[:2]:     {Description: "Third-party Error", HTTPStatus: 502},
	StatusCodeGenericAuthError[:2]:           {Description: "Security Error", HTTPStatus: 401},
	StatusCodeGenericUnauthorized[:2]:        {Description: "Unauthorized", HTTPStatus: 401},
	StatusCodeGenericForbidden[:2]:           {Description: "Forbidden", HTTPStatus: 403},
}

// IsValidCategory validates the 'xy' part of the error code. Returns true if the category exists, false otherwise.
func IsValidCategory(xy string) bool {
	_, exists := validCategories[xy]
	return exists
}

// GetCategoryDescription returns the description of the 'xy' category. If the category does not exist, it returns "Unknown Category".
func GetCategoryDescription(xy string) string {
	if desc, exists := validCategories[xy]; exists {
		return desc.Description
	}
	return "Unknown Category"
}

// GetCategoryHTTPStatus returns the HTTP status code of the 'xy' category. If the category does not exist, it returns 500.
func GetCategoryHTTPStatus(xy string) int {
	if desc, exists := validCategories[xy]; exists {
		return desc.HTTPStatus
	}
	return http.StatusInternalServerError
}
