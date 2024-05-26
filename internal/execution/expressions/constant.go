package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type constantExpression struct {
	value any
}

var _ types.Expression = constantExpression{}

func NewConstant(value any) types.Expression {
	return constantExpression{
		value: value,
	}
}

func (e constantExpression) Equal(other types.Expression) bool {
	if o, ok := other.(constantExpression); ok {
		return e.value == o.value
	}

	return false
}

func (e constantExpression) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e constantExpression) Fold() types.Expression {
	return e
}

func (e constantExpression) Map(f func(types.Expression) types.Expression) types.Expression {
	return f(e)
}

func (e constantExpression) ValueFrom(ctx types.Context, row shared.Row) (any, error) {
	return e.value, nil
}
