package expressions

import (
	"testing"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateNot(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		expr     types.Expression
		expected any
	}{
		{
			name:     "true",
			expr:     NewConstant(true),
			expected: false,
		},
		{
			name:     "false",
			expr:     NewConstant(false),
			expected: true,
		},
		{
			name:     "nil",
			expr:     NewConstant(nil),
			expected: nil,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			e := NewNot(testCase.expr)
			val, err := e.ValueFrom(types.EmptyContext, shared.Row{})
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, val)
		})
	}
}

func TestEvaluateAnd(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		left     types.Expression
		right    types.Expression
		expected any
	}{
		{
			name:     "true",
			left:     NewConstant(true),
			right:    NewConstant(true),
			expected: true,
		},
		{
			name:     "false",
			left:     NewConstant(true),
			right:    NewConstant(false),
			expected: false,
		},
		{
			name:     "nil",
			left:     NewConstant(nil),
			right:    NewConstant(true),
			expected: nil,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			e := NewAnd(testCase.left, testCase.right)
			val, err := e.ValueFrom(types.EmptyContext, shared.Row{})
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, val)
		})
	}
}

func TestEvaluateOr(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		left     types.Expression
		right    types.Expression
		expected any
	}{
		{
			name:     "true",
			left:     NewConstant(true),
			right:    NewConstant(false),
			expected: true,
		},
		{
			name:     "false",
			left:     NewConstant(false),
			right:    NewConstant(false),
			expected: false,
		},
		{
			name:     "semi-nil (true)",
			left:     NewConstant(nil),
			right:    NewConstant(true),
			expected: true,
		},
		{
			name:     "semi-nil (false)",
			left:     NewConstant(nil),
			right:    NewConstant(false),
			expected: nil,
		},
		{
			name:     "nil",
			left:     NewConstant(nil),
			right:    NewConstant(nil),
			expected: nil,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			e := NewOr(testCase.left, testCase.right)
			val, err := e.ValueFrom(types.EmptyContext, shared.Row{})
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, val)
		})
	}
}

func TestConditionalEqual(t *testing.T) {
	a := NewNamed(shared.NewField("t", "a", shared.TypeText))
	b := NewNamed(shared.NewField("t", "b", shared.TypeText))
	c := NewNamed(shared.NewField("t", "c", shared.TypeText))
	d := NewNamed(shared.NewField("t", "d", shared.TypeText))

	e1 := NewEquals(a, b)
	e2 := NewLessThanEquals(b, c)
	e3 := NewGreaterThan(c, d)
	e4 := NewIsDistinctFrom(d, a)

	for _, testCase := range []struct {
		name    string
		factory func(types.Expression, types.Expression) types.Expression
	}{
		{name: "and", factory: NewAnd},
		{name: "or", factory: NewOr},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			equivalences := map[string]types.Expression{
				"chain":     testCase.factory(e1, testCase.factory(e2, testCase.factory(e3, e4))), // e1 . (e2 . (e3 . e4))
				"reordered": testCase.factory(e4, testCase.factory(testCase.factory(e3, e2), e1)), // e4 . ((e3 . e2) . e1)
				"balanced":  testCase.factory(testCase.factory(e1, e2), testCase.factory(e3, e4)), // (e1 . e2) . (e3 . e4)
			}

			for aName, a := range equivalences {
				for bName, b := range equivalences {
					assert.True(t, a.Equal(b), "%s != %s", aName, bName)
				}
			}

			smaller := testCase.factory(e1, testCase.factory(e2, e3))
			larger := testCase.factory(testCase.factory(e1, e3), testCase.factory(e2, e4))
			assert.False(t, smaller.Equal(larger))
		})
	}
}
