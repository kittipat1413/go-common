package validator_test

import (
	"testing"

	ut "github.com/go-playground/universal-translator"
	validatorV10 "github.com/go-playground/validator/v10"
	"github.com/kittipat1413/go-common/framework/validator"
	"github.com/stretchr/testify/assert"
)

type MockCustomValidator struct{}

func (*MockCustomValidator) Tag() string {
	return "mock"
}

func (*MockCustomValidator) Func() validatorV10.Func {
	return func(validatorV10.FieldLevel) bool {
		return true
	}
}

func (*MockCustomValidator) Translation() (string, validatorV10.TranslationFunc) {
	return "", nil
}

type MockCustomValidatorWithTranslation struct{}

func (*MockCustomValidatorWithTranslation) Tag() string {
	return "mock"
}

func (*MockCustomValidatorWithTranslation) Func() validatorV10.Func {
	return func(validatorV10.FieldLevel) bool {
		return true
	}
}

func (*MockCustomValidatorWithTranslation) Translation() (string, validatorV10.TranslationFunc) {
	customTransFunc := func(ut ut.Translator, fe validatorV10.FieldError) string {
		t, _ := ut.T(fe.Tag())
		return t
	}
	return "Mock validation failed", customTransFunc
}

func TestValidatorInitialization(t *testing.T) {
	v, err := validator.NewValidator()
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestCustomValidatorRegistration(t *testing.T) {
	v, err := validator.NewValidator(
		validator.WithCustomValidator(new(MockCustomValidator)),
		validator.WithCustomValidator(new(MockCustomValidatorWithTranslation)),
	)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestValidateStruct(t *testing.T) {
	v, _ := validator.NewValidator()

	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required"`
		Age   int    `validate:"gte=0,lte=130"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
		wantMsg string
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john.doe@example.com",
				Age:   30,
			},
			wantErr: false,
			wantMsg: "",
		},
		{
			name: "missing required field (name)",
			input: TestStruct{
				Email: "john.doe@example.com",
				Age:   30,
			},
			wantErr: true,
			wantMsg: "Name is a required field",
		},
		{
			name: "out of range field (age)",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john.doe@example.com",
				Age:   150,
			},
			wantErr: true,
			wantMsg: "Age must be 130 or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.input)

			// Check if we expect an error
			if tt.wantErr {
				assert.Error(t, err)
				// If an error is returned, check that it contains the expected message
				if err != nil {
					assert.Contains(t, err.Error(), tt.wantMsg, "Error message mismatch")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWithTagNameFunc(t *testing.T) {
	v, err := validator.NewValidator(
		validator.WithTagNameFunc(validator.JSONTagNameFunc),
	)
	assert.NoError(t, err)
	assert.NotNil(t, v)

	type TestStruct struct {
		FullName string `json:"full_name" validate:"required"`
		Email    string `json:"-" validate:"required,email"`
		Age      int    `json:"age" validate:"gte=0,lte=130"`
	}

	tests := []struct {
		name    string
		input   TestStruct
		wantErr bool
		wantMsg string
	}{
		{
			name: "valid struct",
			input: TestStruct{
				FullName: "John Doe",
				Email:    "test@example.com",
				Age:      30,
			},
			wantErr: false,
			wantMsg: "",
		},
		{
			name: "missing required field (full_name)",
			input: TestStruct{
				Age: 30,
			},
			wantErr: true,
			wantMsg: "full_name is a required field",
		},
		{
			name: "field with json:\"-\" tag",
			input: TestStruct{
				FullName: "John Doe",
				Email:    "",
				Age:      30,
			},
			wantErr: true,
			wantMsg: "Email is a required field",
		},
		{
			name: "out of range field (age)",
			input: TestStruct{
				FullName: "John Doe",
				Age:      150,
			},
			wantErr: true,
			wantMsg: "age must be 130 or less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.input)

			// Check if we expect an error
			if tt.wantErr {
				assert.Error(t, err)
				// If an error is returned, check that it contains the expected message
				if err != nil {
					assert.Contains(t, err.Error(), tt.wantMsg, "Error message mismatch")
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
