package expressions

import (
	"fmt"
	"math/big"

	"github.com/efritz/gostgres/internal/shared"
	"golang.org/x/exp/constraints"
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

func newBinaryIntExpression(left, right Expression, operatorText string, f func(a, b any) (any, error)) Expression {
	return newBinaryExpression(left, right, operatorText, func(context Context, left, right Expression, row shared.Row) (any, error) {
		lVal, err := left.ValueFrom(context, row)
		if err != nil {
			return nil, err
		}

		rVal, err := right.ValueFrom(context, row)
		if err != nil {
			return nil, err
		}

		if lVal == nil || rVal == nil {
			return nil, nil
		}

		lVal, rVal, err = shared.PromoteToCommonNumericValues(lVal, rVal)
		if err != nil {
			return nil, err
		}

		return f(lVal, rVal)
	})
}

func add(a, b any) (any, error) {
	switch v := a.(type) {
	case int16:
		return addNumbers(v, b.(int16))
	case int32:
		return addNumbers(v, b.(int32))
	case int64:
		return addNumbers(v, b.(int64))
	case float32:
		return addNumbers(v, b.(float32))
	case float64:
		return addNumbers(v, b.(float64))
	case *big.Float:
		return new(big.Float).Add(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}

func addNumbers[T constraints.Integer | constraints.Float](a, b T) (T, error) {
	return a + b, nil
}

func sub(a, b any) (any, error) {
	switch v := a.(type) {
	case int16:
		return subNumbers(v, b.(int16))
	case int32:
		return subNumbers(v, b.(int32))
	case int64:
		return subNumbers(v, b.(int64))
	case float32:
		return subNumbers(v, b.(float32))
	case float64:
		return subNumbers(v, b.(float64))
	case *big.Float:
		return new(big.Float).Sub(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}

func subNumbers[T constraints.Integer | constraints.Float](a, b T) (T, error) {
	return a - b, nil
}

func mul(a, b any) (any, error) {
	switch v := a.(type) {
	case int16:
		return mulNumbers(v, b.(int16))
	case int32:
		return mulNumbers(v, b.(int32))
	case int64:
		return mulNumbers(v, b.(int64))
	case float32:
		return mulNumbers(v, b.(float32))
	case float64:
		return mulNumbers(v, b.(float64))
	case *big.Float:
		return new(big.Float).Mul(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}

func mulNumbers[T constraints.Integer | constraints.Float](a, b T) (T, error) {
	return a * b, nil
}

func div(a, b any) (any, error) {
	switch v := a.(type) {
	case int16:
		return divNumbers(v, b.(int16))
	case int32:
		return divNumbers(v, b.(int32))
	case int64:
		return divNumbers(v, b.(int64))
	case float32:
		return divNumbers(v, b.(float32))
	case float64:
		return divNumbers(v, b.(float64))
	case *big.Float:
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}

		return new(big.Float).Quo(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}

func divNumbers[T constraints.Integer | constraints.Float](a, b T) (T, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	return a / b, nil
}
