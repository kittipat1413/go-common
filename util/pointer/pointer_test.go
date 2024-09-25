package pointer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Internal_ToPointer(t *testing.T) {
	var tests = []struct {
		name string
		have any
		want any
	}{
		{
			name: "integer input",
			have: 10,
			want: 10,
		},
		{
			name: "string input",
			have: "test-string",
			want: "test-string",
		},
		{
			name: "boolean input",
			have: false,
			want: false,
		},
		{
			name: "struct input",
			have: time.Date(2014, 6, 25, 12, 24, 40, 0, time.UTC),
			want: time.Date(2014, 6, 25, 12, 24, 40, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ToPointer(test.have)
			assert.Equal(t, test.want, *got)
		})

	}
}

func Test_Internal_GetValue(t *testing.T) {
	var (
		nilInt    *int
		nilString *string
		nilBool   *bool
		nilTime   *time.Time
	)
	var tests = []struct {
		name string
		have any
		want any
	}{
		{
			name: "integer input",
			have: 10,
			want: 10,
		},
		{
			name: "string input",
			have: "test-string",
			want: "test-string",
		},
		{
			name: "boolean input",
			have: false,
			want: false,
		},
		{
			name: "struct input",
			have: time.Date(2014, 6, 25, 12, 24, 40, 0, time.UTC),
			want: time.Date(2014, 6, 25, 12, 24, 40, 0, time.UTC),
		},
		{
			name: "zero value integer input",
			have: nilInt,
			want: 0,
		},
		{
			name: "zero value string input",
			have: nilString,
			want: "",
		},
		{
			name: "zero value boolean input",
			have: nilBool,
			want: false,
		},
		{
			name: "zero value struct input",
			have: nilTime,
			want: time.Time{},
		},
		{
			name: "nil input",
			have: nil,
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got any
			switch v := test.have.(type) {
			case *int:
				got = GetValue(v)
			case *string:
				got = GetValue(v)
			case *bool:
				got = GetValue(v)
			case *time.Time:
				got = GetValue(v)
			default:
				got = GetValue(&test.have)
			}
			assert.Equal(t, test.want, got)
		})
	}
}
