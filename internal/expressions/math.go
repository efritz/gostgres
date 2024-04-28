package expressions

import (
	"fmt"
	"math/big"

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
		return v + b.(int16), nil
	case int32:
		return v + b.(int32), nil
	case int64:
		return v + b.(int64), nil
	case float32:
		return v + b.(float32), nil
	case float64:
		return v + b.(float64), nil
	case *big.Float:
		return new(big.Float).Add(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}

func sub(a, b any) (any, error) {
	switch v := a.(type) {
	case int16:
		return v - b.(int16), nil
	case int32:
		return v - b.(int32), nil
	case int64:
		return v - b.(int64), nil
	case float32:
		return v - b.(float32), nil
	case float64:
		return v - b.(float64), nil
	case *big.Float:
		return new(big.Float).Sub(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}

func mul(a, b any) (any, error) {
	switch v := a.(type) {
	case int16:
		return v * b.(int16), nil
	case int32:
		return v * b.(int32), nil
	case int64:
		return v * b.(int64), nil
	case float32:
		return v * b.(float32), nil
	case float64:
		return v * b.(float64), nil
	case *big.Float:
		return new(big.Float).Mul(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}

func div(a, b any) (any, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	switch v := a.(type) {
	case int16:
		return v / b.(int16), nil
	case int32:
		return v / b.(int32), nil
	case int64:
		return v / b.(int64), nil
	case float32:
		return v / b.(float32), nil
	case float64:
		return v / b.(float64), nil
	case *big.Float:
		return new(big.Float).Quo(v, b.(*big.Float)), nil
	}

	panic("unreachable after promotion")
}
