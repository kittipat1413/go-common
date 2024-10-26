package validator

import (
	validator "github.com/go-playground/validator/v10"
)

// CustomValidator defines the interface that custom validators must implement. It requires methods to return the validation tag, function, and translation details.
type CustomValidator interface {
	// Tag returns the tag identifier used in struct field validation tags (e.g., `validate:"tag"`).
	Tag() string
	// Func returns the validator.Func that performs the validation logic.
	Func() validator.Func
	// Translation returns the translation text and an custom translation function for the custom validator.
	// To use the default translation, return an empty string and nil.
	Translation() (translation string, customFunc validator.TranslationFunc)
}
