package expressions

import (
	"testing"

	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/types"
	"github.com/stretchr/testify/assert"
)

func TestBinaryEqual(t *testing.T) {
	a := NewNamed(fields.NewField("t", "a", types.TypeText))
	b := NewNamed(fields.NewField("t", "b", types.TypeText))

	for _, testCase := range []struct {
		name     string
		left     impls.Expression
		right    impls.Expression
		expected bool
	}{
		{
			name:     "equal",
			left:     NewAddition(a, b),
			right:    NewAddition(a, b),
			expected: true,
		},
		{
			name:     "symmetric addition",
			left:     NewAddition(a, b),
			right:    NewAddition(b, a),
			expected: true,
		},
		{
			name:     "symmetric multiplication",
			left:     NewMultiplication(a, b),
			right:    NewMultiplication(b, a),
			expected: true,
		},
		{
			name:     "asymmetric subtraction",
			left:     NewSubtraction(a, b),
			right:    NewSubtraction(b, a),
			expected: false,
		},
		{
			name:     "asymmetric division",
			left:     NewDivision(a, b),
			right:    NewDivision(b, a),
			expected: false,
		},
		{
			name:     "unequal comparison",
			left:     NewGreaterThan(a, b),
			right:    NewGreaterThanEquals(b, a),
			expected: false,
		},
		{
			name:     "symmetric equality",
			left:     NewEquals(a, b),
			right:    NewEquals(b, a),
			expected: true,
		},
		{
			name:     "symmetric distinct from",
			left:     NewIsDistinctFrom(a, b),
			right:    NewIsDistinctFrom(b, a),
			expected: true,
		},
		{
			name:     "flipped greater than",
			left:     NewGreaterThan(a, b),
			right:    NewLessThan(b, a),
			expected: true,
		},
		{
			name:     "flipped less than",
			left:     NewLessThan(a, b),
			right:    NewGreaterThan(b, a),
			expected: true,
		},
		{
			name:     "flipped greater than or equals",
			left:     NewGreaterThanEquals(a, b),
			right:    NewLessThanEquals(b, a),
			expected: true,
		},
		{
			name:     "flipped less than or equals",
			left:     NewLessThanEquals(a, b),
			right:    NewGreaterThanEquals(b, a),
			expected: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.left.Equal(testCase.right) != testCase.expected {
				assert.Equal(t, testCase.expected, testCase.left.Equal(testCase.right))
			}
		})
	}
}
