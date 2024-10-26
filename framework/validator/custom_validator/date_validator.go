package customval

import (
	"time"

	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	v "github.com/kittipat1413/go-common/framework/validator"
)

// Ensure DateValidator implements the CustomValidator interface.
var _ v.CustomValidator = (*DateValidator)(nil)

const (
	// DateValidatorTag is the tag identifier for date validation (`validate:"date=..."`).
	DateValidatorTag = "date"

	// DateOnly represents the 'dateonly' format (YYYY-MM-DD) (`validate:"date=dateonly"`).
	DateOnly = "dateonly"
)

// DateValidator implements the CustomValidator interface for date validation.
type DateValidator struct{}

// Tag returns the tag identifier for the date validator.
func (*DateValidator) Tag() string {
	return DateValidatorTag
}

// Func returns the validation function for date validation.
func (*DateValidator) Func() validator.Func {
	return validateDate
}

// Translation returns the translation text and custom translation function for the date validator.
func (*DateValidator) Translation() (string, validator.TranslationFunc) {
	translationText := "{0} must be a valid date in '{1}' format"

	// Custom translation function to handle parameters
	customTransFunc := func(ut ut.Translator, fe validator.FieldError) string {
		//{0} will be replaced with fe.Field(), {1} with fe.Param()
		t, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
		return t
	}

	return translationText, customTransFunc
}

// validateDate validates a date field based on the specified format.
// Supported formats:
//   - "dateonly": Validates dates in "YYYY-MM-DD" format.
//
// Returns false if the format is unrecognized or if the date doesn't match the specified format.
func validateDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	if dateStr == "" {
		return false
	}

	format := fl.Param()
	var layout string

	switch format {
	case DateOnly:
		layout = "2006-01-02"
	default:
		// If an unrecognized format is provided, validation should fail.
		return false
	}

	_, err := time.Parse(layout, dateStr)
	return err == nil
}
