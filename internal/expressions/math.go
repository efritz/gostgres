package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func IsArithmetic(expr Expression) (_ ArithmeticType, left, right Expression) {
	if e, ok := expr.(binaryExpression); ok {
		if ct := ArithmeticTypeFromString(e.operatorText); ct != ArithmeticTypeUnknown {
			return ct, e.left, e.right
		}
	}

	return ArithmeticTypeUnknown, nil, nil
}

func NewAddition(left, right Expression) Expression {
	return newBinaryIntExpression(left, right, "+", add)
}

func NewSubtraction(left, right Expression) Expression {
	return newBinaryIntExpression(left, right, "-", sub)
}

func NewMultiplication(left, right Expression) Expression {
	return newBinaryIntExpression(left, right, "*", mul)
}

func NewDivision(left, right Expression) Expression {
	return newBinaryIntExpression(left, right, "/", div)
}

func NewUnaryPlus(expression Expression) Expression {
	return NewAddition(NewConstant(0), expression)
}

func NewUnaryMinus(expression Expression) Expression {
	return NewSubtraction(NewConstant(0), expression)
}

func newBinaryIntExpression(left, right Expression, operatorText string, f func(a, b int) (any, error)) Expression {
	return newBinaryExpression(left, right, operatorText, func(left, right Expression, row shared.Row) (any, error) {
		lVal, err := shared.EnsureInt(left.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		rVal, err := shared.EnsureInt(right.ValueFrom(row))
		if err != nil {
			return nil, err
		}

		return f(lVal, rVal)
	})
}

func add(a, b int) (any, error) { return a + b, nil }
func sub(a, b int) (any, error) { return a - b, nil }
func mul(a, b int) (any, error) { return a * b, nil }
func div(a, b int) (any, error) {
	if b == 0 {
		return nil, fmt.Errorf("division by zero")
	}

	return a / b, nil
}
