package expressions

import (
	"testing"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/stretchr/testify/assert"
)

func TestConditionalEqual(t *testing.T) {
	a := NewNamed(shared.NewField("t", "a", shared.TypeKindText, false))
	b := NewNamed(shared.NewField("t", "b", shared.TypeKindText, false))
	c := NewNamed(shared.NewField("t", "c", shared.TypeKindText, false))
	d := NewNamed(shared.NewField("t", "d", shared.TypeKindText, false))

	e1 := NewEquals(a, b)
	e2 := NewLessThanEquals(b, c)
	e3 := NewGreaterThan(c, d)
	e4 := NewIsDistinctFrom(d, a)

	for _, testCase := range []struct {
		name    string
		factory func(Expression, Expression) Expression
	}{
		{name: "and", factory: NewAnd},
		{name: "or", factory: NewOr},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			equivalences := map[string]Expression{
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
