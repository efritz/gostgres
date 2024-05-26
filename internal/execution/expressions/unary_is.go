package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

func NewIsNull(expression types.Expression) types.Expression {
	return newUnaryExpression(expression, "is null", func(ctx types.Context, expression types.Expression, row shared.Row) (any, error) {
		val, err := expression.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	})
}

func NewIsTrue(expression types.Expression) types.Expression {
	return newUnaryExpression(expression, "is true", func(ctx types.Context, expression types.Expression, row shared.Row) (any, error) {
		val, err := shared.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val != nil && *val, nil
	})
}

func NewIsFalse(expression types.Expression) types.Expression {
	return newUnaryExpression(expression, "is false", func(ctx types.Context, expression types.Expression, row shared.Row) (any, error) {
		val, err := shared.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val != nil && !*val, nil
	})
}

func NewIsUnknown(expression types.Expression) types.Expression {
	return newUnaryExpression(expression, "is unknown", func(ctx types.Context, expression types.Expression, row shared.Row) (any, error) {
		val, err := shared.ValueAs[bool](expression.ValueFrom(ctx, row))
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	})
}
