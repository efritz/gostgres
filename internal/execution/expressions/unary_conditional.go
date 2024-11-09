package expressions

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewNot(expression impls.Expression) impls.Expression {
	return newUnaryExpression(expression, "not", func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, nil
		}
		return !*val, nil
	})
}
