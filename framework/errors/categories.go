package errors

// validCategories maps 'xy' codes to their category descriptions. This is used to validate the main and subcategories of error codes.
var validCategories = map[string]string{
	"10": "Success",
	"20": "Client Error",
	"21": "Invalid Parameters",
	"22": "Duplicated Entry",
	"23": "Not Found",
	"24": "Unprocessable Entity",
	"50": "Internal Error",
	"51": "Database Error",
	"52": "Third-party Error",
	"90": "Security Error",
	"91": "Unauthorized",
	"92": "Forbidden",
}

// IsValidCategory validates the 'xy' part of the error code. Returns true if the category exists, false otherwise.
func IsValidCategory(xy string) bool {
	_, exists := validCategories[xy]
	return exists
}

// GetCategoryDescription returns the description of the 'xy' category. If the category does not exist, it returns "Unknown Category".
func GetCategoryDescription(xy string) string {
	if desc, exists := validCategories[xy]; exists {
		return desc
	}
	return "Unknown Category"
}
