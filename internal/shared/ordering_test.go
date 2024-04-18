package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareValues(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		left     interface{}
		right    interface{}
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
		left     []interface{}
		right    []interface{}
		expected OrderType
	}{
		{
			name:     "single column =",
			left:     []interface{}{200},
			right:    []interface{}{200},
			expected: OrderTypeEqual,
		},
		{
			name:     "single column <",
			left:     []interface{}{100},
			right:    []interface{}{200},
			expected: OrderTypeBefore,
		},
		{
			name:     "single column >",
			left:     []interface{}{200},
			right:    []interface{}{100},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, columns =",
			left:     []interface{}{"foo", 200},
			right:    []interface{}{"foo", 200},
			expected: OrderTypeEqual,
		},
		{
			name:     "multi-column, first column <",
			left:     []interface{}{"bar", 100},
			right:    []interface{}{"foo", 200},
			expected: OrderTypeBefore,
		},
		{
			name:     "multi-column, first column >",
			left:     []interface{}{"foo", 200},
			right:    []interface{}{"bar", 100},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, second column <",
			left:     []interface{}{"foo", 100},
			right:    []interface{}{"foo", 200},
			expected: OrderTypeBefore,
		},
		{
			name:     "multi-column, second column >",
			left:     []interface{}{"foo", 200},
			right:    []interface{}{"foo", 100},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, left larger",
			left:     []interface{}{"foo", 200, "bar"},
			right:    []interface{}{"foo", 200},
			expected: OrderTypeEqual,
		},
		{
			name:     "multi-column, right larger",
			left:     []interface{}{"foo", 200},
			right:    []interface{}{"foo", 200, "bar"},
			expected: OrderTypeEqual,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, CompareValueSlices(testCase.left, testCase.right))
		})
	}
}
