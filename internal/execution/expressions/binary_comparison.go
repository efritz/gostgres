package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

func IsComparison(expr types.Expression) (_ ComparisonType, left, right types.Expression) {
	if e, ok := expr.(binaryExpression); ok {
		if ct := ComparisonTypeFromString(e.operatorText); ct != ComparisonTypeUnknown {
			return ct, e.left, e.right
		}
	}

	return ComparisonTypeUnknown, nil, nil
}

func NewEquals(left, right types.Expression) types.Expression {
	return newComparison(left, right, ComparisonTypeEquals)
}

func NewIsDistinctFrom(left, right types.Expression) types.Expression {
	return newComparison(left, right, ComparisonTypeDistinctFrom)
}

func NewLessThanEquals(left, right types.Expression) types.Expression {
	return newComparison(left, right, ComparisonTypeLessThanEquals)
}

func NewLessThan(left, right types.Expression) types.Expression {
	return newComparison(left, right, ComparisonTypeLessThan)
}

func NewGreaterThanEquals(left, right types.Expression) types.Expression {
	return newComparison(left, right, ComparisonTypeGreaterThanEquals)
}

func NewGreaterThan(left, right types.Expression) types.Expression {
	return newComparison(left, right, ComparisonTypeGreaterThan)
}

func NewBetween(left, middle, right types.Expression) types.Expression {
	return NewAnd(NewLessThanEquals(middle, left), NewLessThanEquals(left, right))
}

func NewBetweenSymmetric(left, middle, right types.Expression) types.Expression {
	return NewOr(NewBetween(left, middle, right), NewBetween(left, right, middle))
}

func newComparison(left, right types.Expression, comparisonType ComparisonType) types.Expression {
	return newBinaryExpression(left, right, comparisonType.String(), func(ctx types.Context, left, right types.Expression, row shared.Row) (any, error) {
		lVal, err := left.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}

		rVal, err := right.ValueFrom(ctx, row)
		if err != nil {
			return nil, err
		}

		if lVal == nil || rVal == nil {
			return nil, nil
		}

		return comparisonType.MatchesOrderType(lVal, rVal)
	})
}
