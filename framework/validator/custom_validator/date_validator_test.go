package customval_test

import (
	"testing"

	"github.com/kittipat1413/go-common/framework/validator"
	custom_validator "github.com/kittipat1413/go-common/framework/validator/custom_validator"
	"github.com/kittipat1413/go-common/util/pointer"
	"github.com/stretchr/testify/assert"
)

func TestValidateDateInitialization(t *testing.T) {
	v, err := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestValidateInvalidTag(t *testing.T) {
	v, _ := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)

	type InvalidTagStruct struct {
		Date string `validate:"date=invalid"`
	}

	// Test case with an unsupported "invalid" tag
	input := InvalidTagStruct{Date: "2023-10-25"}
	err := v.ValidateStruct(input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Date must be a valid date in 'invalid' format")
}

func TestValidateDateOnly(t *testing.T) {
	v, _ := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)

	type DateOnlyStruct struct {
		Date    string  `validate:"date=dateonly"`
		DatePtr *string `validate:"date=dateonly"`
	}

	testCases := []struct {
		name    string
		input   DateOnlyStruct
		wantErr bool
		wantMsg string
	}{
		{
			name:    "Valid dateonly format",
			input:   DateOnlyStruct{Date: "2023-10-25", DatePtr: pointer.ToPointer("2023-10-25")},
			wantErr: false,
		},
		{
			name:    "Invalid dateonly format",
			input:   DateOnlyStruct{Date: "10-25-2023", DatePtr: pointer.ToPointer("10/25/2023")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'dateonly' format, DatePtr must be a valid date in 'dateonly' format",
		},
		{
			name:    "Empty date string",
			input:   DateOnlyStruct{Date: "", DatePtr: pointer.ToPointer("2023-10-25")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'dateonly' format",
		},
		{
			name:    "Nil pointer for dateonly",
			input:   DateOnlyStruct{Date: "2023-10-25", DatePtr: nil},
			wantErr: true,
			wantMsg: "DatePtr must be a valid date in 'dateonly' format",
		},
		{
			name:    "Invalid leap year date",
			input:   DateOnlyStruct{Date: "2023-02-29", DatePtr: pointer.ToPointer("2023-02-29")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'dateonly' format, DatePtr must be a valid date in 'dateonly' format",
		},
		{
			name:    "Valid leap year date",
			input:   DateOnlyStruct{Date: "2024-02-29", DatePtr: pointer.ToPointer("2024-02-29")},
			wantErr: false,
			wantMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateStruct(tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, tc.wantMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDateTime(t *testing.T) {
	v, _ := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)

	type DateTimeStruct struct {
		Date    string  `validate:"date=datetime"`
		DatePtr *string `validate:"date=datetime"`
	}

	testCases := []struct {
		name    string
		input   DateTimeStruct
		wantErr bool
		wantMsg string
	}{
		{
			name:    "Valid datetime format",
			input:   DateTimeStruct{Date: "2023-10-25 14:30:00", DatePtr: pointer.ToPointer("2023-10-25 14:30:00")},
			wantErr: false,
		},
		{
			name:    "Invalid datetime format",
			input:   DateTimeStruct{Date: "2023/10/25 14:30:00", DatePtr: pointer.ToPointer("10-25-2023 14:30")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'datetime' format, DatePtr must be a valid date in 'datetime' format",
		},
		{
			name:    "Empty date string",
			input:   DateTimeStruct{Date: "", DatePtr: pointer.ToPointer("2023-10-25 14:30:00")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'datetime' format",
		},
		{
			name:    "Nil pointer for datetime",
			input:   DateTimeStruct{Date: "2023-10-25 14:30:00", DatePtr: nil},
			wantErr: true,
			wantMsg: "DatePtr must be a valid date in 'datetime' format",
		},
		{
			name:    "Invalid leap year datetime",
			input:   DateTimeStruct{Date: "2023-02-29 14:30:00", DatePtr: pointer.ToPointer("2023-02-29 14:30:00")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'datetime' format, DatePtr must be a valid date in 'datetime' format",
		},
		{
			name:    "Valid leap year datetime",
			input:   DateTimeStruct{Date: "2024-02-29 14:30:00", DatePtr: pointer.ToPointer("2024-02-29 14:30:00")},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateStruct(tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, tc.wantMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRFC3339(t *testing.T) {
	v, _ := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)

	type RFC3339Struct struct {
		Date    string  `validate:"date=rfc3339"`
		DatePtr *string `validate:"date=rfc3339"`
	}

	testCases := []struct {
		name    string
		input   RFC3339Struct
		wantErr bool
		wantMsg string
	}{
		{
			name:    "Valid RFC3339 format UTC",
			input:   RFC3339Struct{Date: "2023-10-25T14:30:00Z", DatePtr: pointer.ToPointer("2023-10-25T14:30:00Z")},
			wantErr: false,
		},
		{
			name:    "Valid RFC3339 format with timezone +02:00",
			input:   RFC3339Struct{Date: "2023-10-25T14:30:00+02:00", DatePtr: pointer.ToPointer("2023-10-25T14:30:00+02:00")},
			wantErr: false,
		},
		{
			name:    "Valid RFC3339 format with timezone -07:00",
			input:   RFC3339Struct{Date: "2023-10-25T14:30:00-07:00", DatePtr: pointer.ToPointer("2023-10-25T14:30:00-07:00")},
			wantErr: false,
		},
		{
			name:    "Invalid RFC3339 format with incorrect separator",
			input:   RFC3339Struct{Date: "2023-10-25 14:30:00", DatePtr: pointer.ToPointer("2023/10/25T14:30:00")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'rfc3339' format, DatePtr must be a valid date in 'rfc3339' format",
		},
		{
			name:    "Invalid RFC3339 format without timezone",
			input:   RFC3339Struct{Date: "2023-10-25T14:30:00", DatePtr: pointer.ToPointer("2023-10-25T14:30:00")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'rfc3339' format, DatePtr must be a valid date in 'rfc3339' format",
		},
		{
			name:    "Invalid RFC3339 format with incorrect timezone",
			input:   RFC3339Struct{Date: "2023-10-25T14:30:00+2:00", DatePtr: pointer.ToPointer("2023-10-25T14:30:00+2:00")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'rfc3339' format, DatePtr must be a valid date in 'rfc3339' format",
		},
		{
			name:    "Empty date string",
			input:   RFC3339Struct{Date: "", DatePtr: pointer.ToPointer("2023-10-25T14:30:00Z")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'rfc3339' format",
		},
		{
			name:    "Nil pointer for RFC3339",
			input:   RFC3339Struct{Date: "2023-10-25T14:30:00Z", DatePtr: nil},
			wantErr: true,
			wantMsg: "DatePtr must be a valid date in 'rfc3339' format",
		},
		{
			name:    "Invalid leap year RFC3339",
			input:   RFC3339Struct{Date: "2023-02-29T14:30:00Z", DatePtr: pointer.ToPointer("2023-02-29T14:30:00Z")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'rfc3339' format, DatePtr must be a valid date in 'rfc3339' format",
		},
		{
			name:    "Valid leap year RFC3339",
			input:   RFC3339Struct{Date: "2024-02-29T14:30:00Z", DatePtr: pointer.ToPointer("2024-02-29T14:30:00Z")},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateStruct(tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, tc.wantMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimeOnly(t *testing.T) {
	v, _ := validator.NewValidator(
		validator.WithCustomValidator(new(custom_validator.DateValidator)),
	)

	type TimeOnlyStruct struct {
		Date    string  `validate:"date=timeonly"`
		DatePtr *string `validate:"date=timeonly"`
	}

	testCases := []struct {
		name    string
		input   TimeOnlyStruct
		wantErr bool
		wantMsg string
	}{
		{
			name:    "Valid timeonly format",
			input:   TimeOnlyStruct{Date: "14:30:00", DatePtr: pointer.ToPointer("14:30:00")},
			wantErr: false,
		},
		{
			name:    "Invalid timeonly format",
			input:   TimeOnlyStruct{Date: "14:30", DatePtr: pointer.ToPointer("14:30:00:00")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'timeonly' format, DatePtr must be a valid date in 'timeonly' format",
		},
		{
			name:    "Empty date string",
			input:   TimeOnlyStruct{Date: "", DatePtr: pointer.ToPointer("14:30:00")},
			wantErr: true,
			wantMsg: "Date must be a valid date in 'timeonly' format",
		},
		{
			name:    "Nil pointer for timeonly",
			input:   TimeOnlyStruct{Date: "14:30:00", DatePtr: nil},
			wantErr: true,
			wantMsg: "DatePtr must be a valid date in 'timeonly' format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.ValidateStruct(tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, tc.wantMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
