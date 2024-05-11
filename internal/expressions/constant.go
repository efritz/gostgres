package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type constantExpression struct {
	value any
}

var _ Expression = &constantExpression{}

func NewConstant(value any) Expression {
	return constantExpression{
		value: value,
	}
}

func (e constantExpression) Equal(other Expression) bool {
	if o, ok := other.(constantExpression); ok {
		return e.value == o.value
	}

	return false
}

func (e constantExpression) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e constantExpression) Fold() Expression {
	return e
}

func (e constantExpression) Map(f func(Expression) Expression) Expression {
	return f(e)
}

func (e constantExpression) ValueFrom(context ExpressionContext, row shared.Row) (any, error) {
	return e.value, nil
}
