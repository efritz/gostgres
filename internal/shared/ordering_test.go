package shared

import (
	"math/big"
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
			name:     "integers =",
			left:     int32(200),
			right:    int32(200),
			expected: OrderTypeEqual,
		},
		{
			name:     "integers <",
			left:     int32(100),
			right:    int32(200),
			expected: OrderTypeBefore,
		},
		{
			name:     "integers >",
			left:     int32(200),
			right:    int32(100),
			expected: OrderTypeAfter,
		},

		{
			name:     "promotable numbers =",
			left:     int16(200),
			right:    float32(200),
			expected: OrderTypeEqual,
		},
		{
			name:     "promotable numbers <",
			left:     int64(100),
			right:    float32(200),
			expected: OrderTypeBefore,
		},
		{
			name:     "promotable numbers >",
			left:     big.NewFloat(200),
			right:    int32(100),
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
			left:     []any{int32(200)},
			right:    []any{int32(200)},
			expected: OrderTypeEqual,
		},
		{
			name:     "single column <",
			left:     []any{int32(100)},
			right:    []any{int32(200)},
			expected: OrderTypeBefore,
		},
		{
			name:     "single column >",
			left:     []any{int32(200)},
			right:    []any{int32(100)},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, columns =",
			left:     []any{"foo", int32(200)},
			right:    []any{"foo", int32(200)},
			expected: OrderTypeEqual,
		},
		{
			name:     "multi-column, first column <",
			left:     []any{"bar", int32(100)},
			right:    []any{"foo", int32(200)},
			expected: OrderTypeBefore,
		},
		{
			name:     "multi-column, first column >",
			left:     []any{"foo", int32(200)},
			right:    []any{"bar", int32(100)},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, second column <",
			left:     []any{"foo", int32(100)},
			right:    []any{"foo", int32(200)},
			expected: OrderTypeBefore,
		},
		{
			name:     "multi-column, second column >",
			left:     []any{"foo", int32(200)},
			right:    []any{"foo", int32(100)},
			expected: OrderTypeAfter,
		},
		{
			name:     "multi-column, left larger",
			left:     []any{"foo", int32(200), "bar"},
			right:    []any{"foo", int32(200)},
			expected: OrderTypeEqual,
		},
		{
			name:     "multi-column, right larger",
			left:     []any{"foo", int32(200)},
			right:    []any{"foo", int32(200), "bar"},
			expected: OrderTypeEqual,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, CompareValueSlices(testCase.left, testCase.right))
		})
	}
}
