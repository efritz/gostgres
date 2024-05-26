package expressions

import (
	"fmt"
	"math/big"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
	"golang.org/x/exp/constraints"
)

func IsArithmetic(expr impls.Expression) (_ ArithmeticType, left, right impls.Expression) {
	if e, ok := expr.(binaryExpression); ok {
		if ct := ArithmeticTypeFromString(e.operatorText); ct != ArithmeticTypeUnknown {
			return ct, e.left, e.right
		}
	}

	return ArithmeticTypeUnknown, nil, nil
}

func NewAddition(left, right impls.Expression) impls.Expression {
	return newBinaryIntExpression(left, right, "+", add)
}

func NewSubtraction(left, right impls.Expression) impls.Expression {
	return newBinaryIntExpression(left, right, "-", sub)
}

func NewMultiplication(left, right impls.Expression) impls.Expression {
	return newBinaryIntExpression(left, right, "*", mul)
}

func NewDivision(left, right impls.Expression) impls.Expression {
	return newBinaryIntExpression(left, right, "/", div)
}

func NewUnaryPlus(expression impls.Expression) impls.Expression {
	return NewAddition(NewConstant(0), expression)
}

func NewUnaryMinus(expression impls.Expression) impls.Expression {
	return NewSubtraction(NewConstant(0), expression)
}

func newBinaryIntExpression(left, right impls.Expression, operatorText string, f func(a, b any) (any, error)) impls.Expression {
	return newBinaryExpression(left, right, operatorText, func(ctx impls.Context, left, right impls.Expression, row rows.Row) (any, error) {
		lVal, err := left.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}

		rVal, err := right.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}

		if lVal == nil || rVal == nil {
			return nil, nil
		}

		lVal, rVal, err = types.PromoteToCommonNumericValues(lVal, rVal)
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
		if b.(*big.Float).Cmp(big.NewFloat(0)) == 0 {
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
