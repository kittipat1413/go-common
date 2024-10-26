package customval_test

import (
	"testing"

	"github.com/kittipat1413/go-common/framework/validator"
	custom_validator "github.com/kittipat1413/go-common/framework/validator/custom_validator"
	"github.com/stretchr/testify/assert"
)

func TestValidateDateInitialization(t *testing.T) {
	v, err := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestValidateDate(t *testing.T) {
	v, _ := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)

	type TestStruct struct {
		Date string `validate:"date=dateonly"`
	}
	type TestInvalidTagStruct struct {
		Date string `validate:"date=invalid"`
	}

	testCases := []struct {
		name    string
		input   interface{}
		wantErr bool
		wantMsg string
	}{
		{
			name:    "Valid date",
			input:   TestStruct{Date: "2023-10-25"},
			wantErr: false,
			wantMsg: "",
		},
		{
			name:    "Invalid date format",
			input:   TestStruct{Date: "25-10-2023"},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'dateonly' format",
		},
		{
			name:    "Empty date string",
			input:   TestStruct{Date: ""},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'dateonly' format",
		},
		{
			name:    "Invalid tag",
			input:   TestInvalidTagStruct{Date: "2023-10-25"},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'invalid' format",
		},
		{
			name:    "Invalid leap year date",
			input:   TestStruct{Date: "2023-02-29"},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'dateonly' format",
		},
		{
			name:    "Valid leap year date",
			input:   TestStruct{Date: "2024-02-29"},
			wantErr: false,
			wantMsg: "",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateStruct(tc.input)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				// Check if the error message contains the expected message
				if err != nil {
					assert.Contains(t, err.Error(), tc.wantMsg, "Error message mismatch")
				}
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}
