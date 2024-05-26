package impls

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/rows"
)

type Expression interface {
	fmt.Stringer

	Equal(other Expression) bool
	Fold() Expression
	Map(f func(Expression) Expression) Expression
	ValueFrom(cts Context, row rows.Row) (any, error)
}

type AggregateExpression interface {
	Step(ctx Context, row rows.Row) error
	Done(ctx Context) (any, error)
}

type OrderExpression interface {
	Expressions() []ExpressionWithDirection
	Fold() OrderExpression
	Map(f func(e Expression) Expression) OrderExpression
}

type ExpressionWithDirection struct {
	Expression Expression
	Reverse    bool
}

func (e ExpressionWithDirection) Fold() ExpressionWithDirection {
	return ExpressionWithDirection{
		Expression: e.Expression.Fold(),
		Reverse:    e.Reverse,
	}
}
