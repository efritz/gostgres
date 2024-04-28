package shared

import (
	"math/big"

	"golang.org/x/exp/constraints"
)

type OrderType int

const (
	OrderTypeIncomparable OrderType = iota
	OrderTypeNulls
	OrderTypeBefore
	OrderTypeEqual
	OrderTypeAfter
)

func CompareValueSlices(left, right []any) OrderType {
	for i, lVal := range left[:min(len(left), len(right))] {
		if cmp := CompareValues(lVal, right[i]); cmp != OrderTypeEqual {
			return cmp
		}
	}

	return OrderTypeEqual
}

func CompareValues(left, right any) OrderType {
	if left == nil && right == nil {
		return OrderTypeNulls
	}
	if left == nil && right != nil {
		return OrderTypeAfter
	}
	if left != nil && right == nil {
		return OrderTypeBefore
	}

	if lVal, ok := left.(string); ok {
		if rVal, ok := right.(string); ok {
			if lVal == rVal {
				return OrderTypeEqual
			}

			if lVal < rVal {
				return OrderTypeBefore
			}

			return OrderTypeAfter
		}
	}

	if a, b, err := PromoteToCommonNumericValues(left, right); err == nil {
		switch v := a.(type) {
		case int16:
			return compareNumbers(v, b.(int16))
		case int32:
			return compareNumbers(v, b.(int32))
		case int64:
			return compareNumbers(v, b.(int64))
		case float32:
			return compareNumbers(v, b.(float32))
		case float64:
			return compareNumbers(v, b.(float64))
		case *big.Float:
			cmp := v.Cmp(b.(*big.Float))
			if cmp == -1 {
				return OrderTypeBefore
			}

			if cmp == 1 {
				return OrderTypeAfter
			}

			return OrderTypeEqual
		}
	}

	return OrderTypeIncomparable
}

func compareNumbers[T constraints.Integer | constraints.Float](a, b T) OrderType {
	if a < b {
		return OrderTypeBefore
	}

	if a > b {
		return OrderTypeAfter
	}

	return OrderTypeEqual
}
