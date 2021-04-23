package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func NewEquals(left, right Expression) Expression {
	return newComparison(left, right, "=", func(lVal, rVal interface{}, ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeEqual, nil
	})
}

func NewIsDistinctFrom(left, right Expression) Expression {
	return newComparison(left, right, "is distinct from", func(lVal, rVal interface{}, ot shared.OrderType) (interface{}, error) {
		if lVal == nil && rVal == nil {
			return false, nil
		}
		if lVal == nil || rVal == nil {
			return true, nil
		}
		return ot != shared.OrderTypeEqual, nil
	})
}

func NewBetween(left, middle, right Expression) Expression {
	return NewAnd(NewLessThanEquals(middle, left), NewLessThanEquals(left, right))
}

func NewBetweenSymmetric(left, middle, right Expression) Expression {
	return NewOr(NewBetween(left, middle, right), NewBetween(left, right, middle))
}

func NewLessThan(left, right Expression) Expression {
	return newComparison(left, right, "<", func(lVal, rVal interface{}, ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeBefore, nil
	})
}

func NewLessThanEquals(left, right Expression) Expression {
	return newComparison(left, right, "<=", func(lVal, rVal interface{}, ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeBefore || ot == shared.OrderTypeEqual, nil
	})
}

func NewGreaterThan(left, right Expression) Expression {
	return newComparison(left, right, ">", func(lVal, rVal interface{}, ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeAfter, nil
	})
}

func NewGreaterThanEquals(left, right Expression) Expression {
	return newComparison(left, right, ">=", func(lVal, rVal interface{}, ot shared.OrderType) (interface{}, error) {
		return ot == shared.OrderTypeAfter || ot == shared.OrderTypeEqual, nil
	})
}

func newComparison(left, right Expression, operatorText string, f func(lVal, rVal interface{}, ot shared.OrderType) (interface{}, error)) Expression {
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
		if ot == shared.OrderTypeIncomparable {
			return nil, fmt.Errorf("incomparable types")
		}

		return f(lVal, rVal, ot)
	})
}
