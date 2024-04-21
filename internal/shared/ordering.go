package shared

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

	if lVal, ok := left.(int); ok {
		if rVal, ok := right.(int); ok {
			if lVal == rVal {
				return OrderTypeEqual
			}

			if lVal < rVal {
				return OrderTypeBefore
			}

			return OrderTypeAfter
		}
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

	return OrderTypeIncomparable
}
