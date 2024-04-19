package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareValues(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		left     any
		right    any
		expected OrderType
	}{
		{
			name:     "numbers =",
			left:     200,
			right:    200,
			expected: OrderTypeEqual,
		},
		{
			name:     "numbers <",
			left:     100,
			right:    200,
			expected: OrderTypeBefore,
		},
		{
			name:     "numbers >",
			left:     200,
			right:    100,
			expected: OrderTypeAfter,
		},
		{
			name:     "text =",
			left:     "foo",
			right:    "foo",
			expected: OrderTypeEqual,
		},
		{
			name:     "text <",
			left:     "bar",
			right:    "foo",
			expected: OrderTypeBefore,
		},
		{
			name:     "text >",
			left:     "foo",
			right:    "bar",
			expected: OrderTypeAfter,
		},
		{
			name:     "incomparable",
			left:     "foo",
			right:    200,
			expected: OrderTypeIncomparable,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, CompareValues(testCase.left, testCase.right))
		})
	}
}

func TestCompareValueSlices(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		left     []any
		right    []any
		expected OrderType
	}{
		{
			name:     "single column =",
			left:     []any{200},
			right:    []any{200},
			expected: OrderTypeEqual,
		},
		{
			name:     "single column <",
			left:     []any{100},
			right:    []any{200},
			expected: OrderTypeBefore,
		},
		{
			name:     "single column >",
			left:     []any{200},
			right:    []any{100},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, columns =",
			left:     []any{"foo", 200},
			right:    []any{"foo", 200},
			expected: OrderTypeEqual,
		},
		{
			name:     "multi-column, first column <",
			left:     []any{"bar", 100},
			right:    []any{"foo", 200},
			expected: OrderTypeBefore,
		},
		{
			name:     "multi-column, first column >",
			left:     []any{"foo", 200},
			right:    []any{"bar", 100},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, second column <",
			left:     []any{"foo", 100},
			right:    []any{"foo", 200},
			expected: OrderTypeBefore,
		},
		{
			name:     "multi-column, second column >",
			left:     []any{"foo", 200},
			right:    []any{"foo", 100},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, left larger",
			left:     []any{"foo", 200, "bar"},
			right:    []any{"foo", 200},
			expected: OrderTypeEqual,
		},
		{
			name:     "multi-column, right larger",
			left:     []any{"foo", 200},
			right:    []any{"foo", 200, "bar"},
			expected: OrderTypeEqual,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, CompareValueSlices(testCase.left, testCase.right))
		})
	}
}
