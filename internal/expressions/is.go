package expressions

import "github.com/efritz/gostgres/internal/shared"

func NewIsNull(expression Expression) Expression {
	return newUnaryExpression(expression, "is null", func(expression Expression, row shared.Row) (any, error) {
		val, err := expression.ValueFrom(row)
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	})
}

func NewIsTrue(expression Expression) Expression {
	return newUnaryExpression(expression, "is true", func(expression Expression, row shared.Row) (any, error) {
		val, err := shared.EnsureNullableBool(expression.ValueFrom(row))
		if err != nil {
			return nil, err
		}
		return val != nil && *val, nil
	})
}

func NewIsFalse(expression Expression) Expression {
	return newUnaryExpression(expression, "is false", func(expression Expression, row shared.Row) (any, error) {
		val, err := shared.EnsureNullableBool(expression.ValueFrom(row))
		if err != nil {
			return nil, err
		}
		return val != nil && !*val, nil
	})
}

func NewIsUnknown(expression Expression) Expression {
	return newUnaryExpression(expression, "is unknown", func(expression Expression, row shared.Row) (any, error) {
		val, err := shared.EnsureNullableBool(expression.ValueFrom(row))
		if err != nil {
			return nil, err
		}
		return val == nil, nil
	})
}
