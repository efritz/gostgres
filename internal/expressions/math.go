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

		lVal, rVal, err = promoteToCommonNumericValues(lVal, rVal)
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

//
// Move this to shared types

var promoters = [][]func(value any) any{
	{
		func(value any) any { panic("impossible promotion") },               // smallint -> smallint
		func(value any) any { return int32(value.(int16)) },                 // smallint -> integer
		func(value any) any { return int64(value.(int16)) },                 // smallint -> bigint
		func(value any) any { return float32(value.(int16)) },               // smallint -> real
		func(value any) any { return float64(value.(int16)) },               // smallint -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int16))) }, // smallint -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },               // integer -> smallint
		func(value any) any { panic("impossible promotion") },               // integer -> integer
		func(value any) any { return int64(value.(int32)) },                 // integer -> bigint
		func(value any) any { return float32(value.(int32)) },               // integer -> real
		func(value any) any { return float64(value.(int32)) },               // integer -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int32))) }, // integer -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },               // bigint -> smallint
		func(value any) any { panic("impossible promotion") },               // bigint -> integer
		func(value any) any { panic("impossible promotion") },               // bigint -> bigint
		func(value any) any { return float32(value.(int64)) },               // bigint -> real (as double precision)
		func(value any) any { return float64(value.(int64)) },               // bigint -> double precision
		func(value any) any { return big.NewFloat(float64(value.(int64))) }, // bigint -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },                 // real -> smallint
		func(value any) any { panic("impossible promotion") },                 // real -> integer
		func(value any) any { panic("impossible promotion") },                 // real -> bigint
		func(value any) any { panic("impossible promotion") },                 // real -> real
		func(value any) any { return float64(value.(float32)) },               // real -> double precision
		func(value any) any { return big.NewFloat(float64(value.(float32))) }, // real -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") },                 // double precision -> smallint
		func(value any) any { panic("impossible promotion") },                 // double precision -> integer
		func(value any) any { panic("impossible promotion") },                 // double precision -> bigint
		func(value any) any { panic("impossible promotion") },                 // double precision -> real
		func(value any) any { panic("impossible promotion") },                 // double precision -> double precision
		func(value any) any { return big.NewFloat(float64(value.(float64))) }, // double precision -> numeric
	},
	{
		func(value any) any { panic("impossible promotion") }, // numeric -> smallint
		func(value any) any { panic("impossible promotion") }, // smallint -> integer
		func(value any) any { panic("impossible promotion") }, // smallint -> bigint
		func(value any) any { panic("impossible promotion") }, // numeric -> real
		func(value any) any { panic("impossible promotion") }, // numeric -> double precision
		func(value any) any { panic("impossible promotion") }, // numeric -> numeric
	},
}

func promoteToCommonNumericValues(left, right any) (any, any, error) {
	lIndex := numericTypeIndex(left)
	if lIndex < 0 {
		return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", left)
	}

	rIndex := numericTypeIndex(right)
	if rIndex < 0 {
		return nil, nil, fmt.Errorf("unexpected type (wanted numeric type, have %v)", right)
	}

	if lIndex < rIndex {
		return promoters[lIndex][rIndex](left), right, nil
	}

	if rIndex < lIndex {
		return left, promoters[rIndex][lIndex](right), nil
	}

	return left, right, nil
}

func numericTypeIndex(value any) int {
	switch value.(type) {
	case int16:
		return 0
	case int32:
		return 1
	case int64:
		return 2
	case float32:
		return 3
	case float64:
		return 4
	case *big.Float:
		return 5
	}

	return -1
}
