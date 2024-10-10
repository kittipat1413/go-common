package slice_test

import (
	"testing"

	"github.com/kittipat1413/go-common/util/slice"
	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicate(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "No duplicates",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "With duplicates",
			input:    []int{1, 2, 2, 3, 4, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "All duplicates",
			input:    []int{1, 1, 1, 1},
			expected: []int{1},
		},
		{
			name:     "Empty slice",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := slice.RemoveDuplicate(test.input)
			assert.Equal(t, test.expected, output, "Test %s failed", test.name)
		})
	}
}

func TestUnion(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []int
		slice2   []int
		expected []int
	}{
		{
			name:     "No overlap",
			slice1:   []int{1, 2, 3},
			slice2:   []int{4, 5, 6},
			expected: []int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "With overlap",
			slice1:   []int{1, 2, 3},
			slice2:   []int{3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "All duplicates",
			slice1:   []int{1, 1, 1},
			slice2:   []int{1, 1, 1},
			expected: []int{1},
		},
		{
			name:     "One empty slice",
			slice1:   []int{1, 2, 3},
			slice2:   []int{},
			expected: []int{1, 2, 3},
		},
		{
			name:     "Both empty slices",
			slice1:   []int{},
			slice2:   []int{},
			expected: []int{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := slice.Union(test.slice1, test.slice2)
			assert.ElementsMatch(t, test.expected, output, "Test %s failed", test.name)
		})
	}
}
