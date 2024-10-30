package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type constantExpression struct {
	value any
}

var _ impls.Expression = constantExpression{}

func NewConstant(value any) impls.Expression {
	return constantExpression{
		value: value,
	}
}

func (e constantExpression) Equal(other impls.Expression) bool {
	if o, ok := other.(constantExpression); ok {
		return e.value == o.value
	}

	return false
}

func (e constantExpression) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e constantExpression) Fold() impls.Expression {
	return e
}

func (e constantExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	return f(e)
}

func (e constantExpression) ValueFrom(ctx impls.Context, row rows.Row) (any, error) {
	return e.value, nil
}
