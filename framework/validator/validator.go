package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

// Validator holds the validator and translator instances used for struct validation.
type Validator struct {
	validate   *validator.Validate
	translator ut.Translator
}

// ValidatorOption defines a functional option for configuring the validator instance.
type ValidatorOption func(*validator.Validate, ut.Translator) error

// NewValidator creates and returns a new Validator instance with the provided options.
// It initializes the validator, applies custom options, sets up the translator, and registers default translations.
//
// Example:
//
//	v, err := NewValidator(
//	    WithCustomValidator(DateValidator{}),
//	)
//	if err != nil {
//	    // handle error
//	}
func NewValidator(opts ...ValidatorOption) (*Validator, error) {
	v := validator.New()

	// Initialize the English locale and the universal translator.
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	// Get the translator for English.
	translator, found := uni.GetTranslator("en")
	if !found {
		return nil, fmt.Errorf("translator not found for locale 'en'")
	}
	// Register English translations.
	if err := enTranslations.RegisterDefaultTranslations(v, translator); err != nil {
		return nil, err
	}

	// Apply any custom validator options provided
	for _, opt := range opts {
		if err := opt(v, translator); err != nil {
			return nil, err
		}
	}

	return &Validator{
		validate:   v,
		translator: translator,
	}, nil
}

// WithCustomValidator registers a custom validator along with its translation.
// It uses the CustomValidator interface to get the tag, function, and translation details.
func WithCustomValidator(cv CustomValidator) ValidatorOption {
	return func(v *validator.Validate, translator ut.Translator) error {
		// Register the custom validation function
		if err := v.RegisterValidation(cv.Tag(), cv.Func()); err != nil {
			return err
		}

		// Get the translation text and custom translation function
		translationText, customTransFunc := cv.Translation()

		// Register the translation only if both translationText and customTransFunc are provided
		if translationText == "" || customTransFunc == nil {
			return nil // Skip registration if either component is missing
		}

		// Register function for adding the translation
		registerFn := func(ut ut.Translator) error {
			return ut.Add(cv.Tag(), translationText, true)
		}

		// Register the translation with the custom function
		return v.RegisterTranslation(cv.Tag(), translator, registerFn, customTransFunc)
	}
}

// ValidateStruct validates the provided struct using the validator instance.
// It returns an error containing all validation errors with messages translated using the translator.
//
// Example:
//
//	err := v.ValidateStruct(myStruct)
//	if err != nil {
//	    // Handle validation errors
//	}
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			errMsgs := make([]string, len(ve))
			for i, fe := range ve {
				errMsgs[i] = fe.Translate(v.translator)
			}
			return errors.New(strings.Join(errMsgs, ", "))
		}
		return err
	}
	return nil
}
