package expressions

import "github.com/efritz/gostgres/internal/shared"

func NewNot(expression Expression) Expression {
	return newUnaryExpression(expression, "not", func(row shared.Row) (interface{}, error) {
		val, err := EnsureBool(expression.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		return !val, nil
	})
}

func NewAnd(left, right Expression) Expression {
	return newBinaryBoolExpression(left, right, " and ", func(a, b bool) (interface{}, error) {
		return a && b, nil
	})
}

func NewOr(left, right Expression) Expression {
	return newBinaryBoolExpression(left, right, " or ", func(a, b bool) (interface{}, error) {
		return a || b, nil
	})
}

func newBinaryBoolExpression(left, right Expression, operatorText string, f func(a, b bool) (interface{}, error)) Expression {
	return newBinaryExpression(left, right, operatorText, func(row shared.Row) (interface{}, error) {
		lVal, err := EnsureBool(left.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		rVal, err := EnsureBool(right.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		return f(lVal, rVal)
	})
}
