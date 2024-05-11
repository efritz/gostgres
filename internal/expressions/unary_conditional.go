package expressions

import "github.com/efritz/gostgres/internal/shared"

func NewNot(expression Expression) Expression {
	return newUnaryExpression(expression, "not", func(context ExpressionContext, expression Expression, row shared.Row) (any, error) {
		val, err := shared.ValueAs[bool](expression.ValueFrom(context, row))
		if err != nil {
			return nil, err
		}
		if val == nil {
			return nil, nil
		}
		return !*val, nil
	})
}
