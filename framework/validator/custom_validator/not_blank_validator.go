package customval

import (
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	nonstandard "github.com/go-playground/validator/v10/non-standard/validators"
	v "github.com/kittipat1413/go-common/framework/validator"
)

// references:
// - https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Non_standard_validators

// Ensure NotBlankValidator implements the NotBlank interface.
var _ v.CustomValidator = (*NotBlankValidator)(nil)

// NotBlankValidator implements the CustomValidator interface for data validation.
type NotBlankValidator struct{}

// Tag returns the tag identifier for the the validator.
func (*NotBlankValidator) Tag() string {
	return "notblank"
}

// Func simply returns the built-in NotBlank validation provided by validator/v10.
func (*NotBlankValidator) Func() validator.Func {
	return nonstandard.NotBlank
}

// Translation returns the translation text and an custom translation function for the custom validator.
func (*NotBlankValidator) Translation() (string, validator.TranslationFunc) {
	translationText := "{0} cannot be blank"

	customTransFunc := func(ut ut.Translator, fe validator.FieldError) string {
		// {0} will be replaced with fe.Field()
		t, _ := ut.T(fe.Tag(), fe.Field())
		return t
	}

	return translationText, customTransFunc
}
