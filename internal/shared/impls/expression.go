package impls

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type Expression interface {
	fmt.Stringer

	Resolve(ctx ExpressionResolutionContext) error
	Type() types.Type
	Equal(other Expression) bool
	Fold() Expression
	Map(f func(Expression) (Expression, error)) (Expression, error)
	ValueFrom(cts ExecutionContext, row rows.Row) (any, error)
}

type AggregateExpression interface {
	Step(ctx ExecutionContext, row rows.Row) error
	Done(ctx ExecutionContext) (any, error)
}

type OrderExpression interface {
	Expressions() []ExpressionWithDirection
	Fold() OrderExpression
	Map(f func(e Expression) (Expression, error)) (OrderExpression, error)
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
