package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func NewEquals(left, right Expression) Expression {
	return newComparison(left, right, "=", func(ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeEqual, nil
	})
}

func NewLessThan(left, right Expression) Expression {
	return newComparison(left, right, "<", func(ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeBefore, nil
	})
}

func NewLessThanEquals(left, right Expression) Expression {
	return newComparison(left, right, "<=", func(ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeBefore || ot == shared.OrderTypeEqual, nil
	})
}

func NewGreaterThan(left, right Expression) Expression {
	return newComparison(left, right, ">", func(ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeAfter, nil
	})
}

func NewGreaterThanEquals(left, right Expression) Expression {
	return newComparison(left, right, ">=", func(ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeAfter || ot == shared.OrderTypeEqual, nil
	})
}

func newComparison(left, right Expression, operatorText string, f func(ot shared.OrderType) (interface{}, error)) Expression {
	return newBinaryExpression(left, right, operatorText, func(left, right Expression, row shared.Row) (interface{}, error) {
		lVal, err := left.ValueFrom(row)
		if err != nil {
			return nil, err
		}

		rVal, err := right.ValueFrom(row)
		if err != nil {
			return nil, err
		}

		ot := shared.CompareValues(lVal, rVal)
		switch ot {
		case shared.OrderTypeIncomparable:
			return nil, fmt.Errorf("incomparable types")
		default:
			return f(ot)
		}
	})
}
