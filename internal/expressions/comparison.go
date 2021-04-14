package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func NewEquals(left, right Expression) Expression {
	return newComparison(left, right, "=", func(relation shared.OrderType) (interface{}, error) {
		return relation == shared.OrderTypeEqual, nil
	})
}

func NewLessThan(left, right Expression) Expression {
	return newComparison(left, right, "<", func(relation shared.OrderType) (interface{}, error) {
		return relation == shared.OrderTypeBefore, nil
	})
}

func NewLessThanEquals(left, right Expression) Expression {
	return newComparison(left, right, "<=", func(relation shared.OrderType) (interface{}, error) {
		return relation == shared.OrderTypeBefore || relation == shared.OrderTypeEqual, nil
	})
}

func NewGreaterThan(left, right Expression) Expression {
	return newComparison(left, right, ">", func(relation shared.OrderType) (interface{}, error) {
		return relation == shared.OrderTypeAfter, nil
	})
}

func NewGreaterThanEquals(left, right Expression) Expression {
	return newComparison(left, right, ">=", func(relation shared.OrderType) (interface{}, error) {
		return relation == shared.OrderTypeAfter || relation == shared.OrderTypeEqual, nil
	})
}

func newComparison(left, right Expression, operatorText string, f func(relation shared.OrderType) (interface{}, error)) Expression {
	return newBinaryExpression(left, right, operatorText, func(row shared.Row) (interface{}, error) {
		lVal, err := left.ValueFrom(row)
		if err != nil {
			return nil, err
		}

		rVal, err := right.ValueFrom(row)
		if err != nil {
			return nil, err
		}

		relation := shared.CompareValues(lVal, rVal)

		if relation == shared.OrderTypeIncomparable {
			return nil, fmt.Errorf("incomparable types")
		}

		return f(relation)
	})
}
