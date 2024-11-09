package expressions

import (
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func NewIsNull(expression impls.Expression) impls.Expression {
	return newUnaryExpression(expression, "is null", func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := expression.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	})
}

func NewIsTrue(expression impls.Expression) impls.Expression {
	return newUnaryExpression(expression, "is true", func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val != nil && *val, nil
	})
}

func NewIsFalse(expression impls.Expression) impls.Expression {
	return newUnaryExpression(expression, "is false", func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val != nil && !*val, nil
	})
}

func NewIsUnknown(expression impls.Expression) impls.Expression {
	return newUnaryExpression(expression, "is unknown", func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error) {
		val, err := types.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	})
}
