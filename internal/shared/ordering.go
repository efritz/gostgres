package shared

type OrderType int

const (
	OrderTypeIncomparable OrderType = iota
	OrderTypeBefore
	OrderTypeEqual
	OrderTypeAfter
)

func CompareValues(left, right interface{}) OrderType {
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
