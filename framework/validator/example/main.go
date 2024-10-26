package main

import (
	"fmt"

	ut "github.com/go-playground/universal-translator"
	validatorV10 "github.com/go-playground/validator/v10"
	"github.com/kittipat1413/go-common/framework/validator"
)

var _ validator.CustomValidator = (*MyValidator)(nil)

type MyValidator struct{}

// Tag returns the tag identifier used in struct field validation tags.
func (*MyValidator) Tag() string {
	return "mytag"
}

// Func returns the validator.Func that performs the validation logic.
func (*MyValidator) Func() validatorV10.Func {
	return func(fl validatorV10.FieldLevel) bool {
		// Custom validation logic
		return false
	}
}

// Translation returns the translation text and an custom translation function for the custom validator.
func (*MyValidator) Translation() (string, validatorV10.TranslationFunc) {
	translationText := "{0} failed custom validation"

	customTransFunc := func(ut ut.Translator, fe validatorV10.FieldError) string {
		// {0} will be replaced with fe.Field()
		t, _ := ut.T(fe.Tag(), fe.Field())
		return t
	}

	return translationText, customTransFunc
}

type Data struct {
	Field1 string `validate:"mytag"`
	Field2 string `validate:"required"`
	Field3 int    `validate:"gte=0,lte=130"`
}

func main() {
	v, err := validator.NewValidator(
		validator.WithCustomValidator(new(MyValidator)),
	)
	if err != nil {
		fmt.Println("Error initializing validator:", err)
		return
	}

	data := Data{
		Field1: "test",
		Field3: 200,
	}

	err = v.ValidateStruct(data)
	if err != nil {
		fmt.Println("Validation failed:", err)
	} else {
		fmt.Println("Validation passed")
	}
}
