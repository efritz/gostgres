package expressions

import (
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
	return newBinaryExpression(left, right, comparisonType.String(), func(left, right Expression, row shared.Row) (any, error) {
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
