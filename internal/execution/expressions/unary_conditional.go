package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

func NewNot(expression types.Expression) types.Expression {
	return newUnaryExpression(expression, "not", func(ctx types.Context, expression types.Expression, row shared.Row) (any, error) {
		val, err := shared.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, nil
		}
		return !*val, nil
	})
}
