package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

func IsComparison(expr Expression) (_ ComparisonType, left, right Expression) {
	if e, ok := expr.(binaryExpression); ok {
		if ct := ComparisonTypeFromString(e.operatorText); ct != ComparisonTypeUnknown {
			return ct, e.left, e.right
		}
	}

	return ComparisonTypeUnknown, nil, nil
}

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

func NewEquals(left, right Expression) Expression {
	return newComparison(left, right, ComparisonTypeEquals)
}

func NewIsDistinctFrom(left, right Expression) Expression {
	return newComparison(left, right, ComparisonTypeDistinctFrom)
}

func NewLessThanEquals(left, right Expression) Expression {
	return newComparison(left, right, ComparisonTypeLessThanEquals)
}

func NewLessThan(left, right Expression) Expression {
	return newComparison(left, right, ComparisonTypeLessThan)
}

func NewGreaterThanEquals(left, right Expression) Expression {
	return newComparison(left, right, ComparisonTypeGreaterThanEquals)
}

func NewGreaterThan(left, right Expression) Expression {
	return newComparison(left, right, ComparisonTypeGreaterThan)
}

func NewBetween(left, middle, right Expression) Expression {
	return NewAnd(NewLessThanEquals(middle, left), NewLessThanEquals(left, right))
}

func NewBetweenSymmetric(left, middle, right Expression) Expression {
	return NewOr(NewBetween(left, middle, right), NewBetween(left, right, middle))
}

func newComparison(left, right Expression, comparisonType ComparisonType) Expression {
	return newBinaryExpression(left, right, comparisonType.String(), func(left, right Expression, row shared.Row) (interface{}, error) {
		lVal, err := left.ValueFrom(row)
		if err != nil {
			return nil, err
		}

		rVal, err := right.ValueFrom(row)
		if err != nil {
			return nil, err
		}

		return comparisonType.MatchesOrderType(lVal, rVal)
	})
}
