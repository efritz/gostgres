package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type ComparisonType int

const (
	ComparisonTypeUnknown ComparisonType = iota
	ComparisonTypeEquals
	ComparisonTypeDistinctFrom
	ComparisonTypeLessThanEquals
	ComparisonTypeLessThan
	ComparisonTypeGreaterThanEquals
	ComparisonTypeGreaterThan
)

var comparisonTypeOperators = map[ComparisonType]string{
	ComparisonTypeEquals:            "=",
	ComparisonTypeDistinctFrom:      "is distinct from",
	ComparisonTypeLessThanEquals:    "<=",
	ComparisonTypeLessThan:          "<",
	ComparisonTypeGreaterThanEquals: ">=",
	ComparisonTypeGreaterThan:       ">",
}

func ComparisonTypeFromString(operator string) ComparisonType {
	for ct, op := range comparisonTypeOperators {
		if op == operator {
			return ct
		}
	}

	return ComparisonTypeUnknown
}

func (ct ComparisonType) String() string {
	if operator, ok := comparisonTypeOperators[ct]; ok {
		return operator
	}

	return "unknown"
}

func (ct ComparisonType) Flip() ComparisonType {
	switch ct {
	case ComparisonTypeEquals:
		return ComparisonTypeEquals
	case ComparisonTypeDistinctFrom:
		return ComparisonTypeDistinctFrom
	case ComparisonTypeLessThanEquals:
		return ComparisonTypeGreaterThanEquals
	case ComparisonTypeLessThan:
		return ComparisonTypeGreaterThan
	case ComparisonTypeGreaterThanEquals:
		return ComparisonTypeLessThanEquals
	case ComparisonTypeGreaterThan:
		return ComparisonTypeLessThan
	}

	return ComparisonTypeUnknown
}

func (ct ComparisonType) MatchesOrderType(lVal, rVal interface{}) (bool, error) {
	ot := shared.CompareValues(lVal, rVal)

	switch ct {
	case ComparisonTypeEquals:
		return ot == shared.OrderTypeEqual, nil
	case ComparisonTypeDistinctFrom:
		if lVal == nil && rVal == nil {
			return false, nil
		}
		if lVal == nil || rVal == nil {
			return true, nil
		}
		return ot != shared.OrderTypeEqual, nil
	case ComparisonTypeLessThanEquals:
		return ot == shared.OrderTypeEqual || ot == shared.OrderTypeBefore, nil
	case ComparisonTypeLessThan:
		return ot == shared.OrderTypeBefore, nil
	case ComparisonTypeGreaterThanEquals:
		return ot == shared.OrderTypeEqual || ot == shared.OrderTypeAfter, nil
	case ComparisonTypeGreaterThan:
		return ot == shared.OrderTypeAfter, nil
	}

	return false, fmt.Errorf("incomparable types")
}
