package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

func IsComparison(expr impls.Expression) (_ ComparisonType, left, right impls.Expression) {
	if e, ok := expr.(*binaryExpression); ok {
		if ct := ComparisonTypeFromString(e.operatorText); ct != ComparisonTypeUnknown {
			return ct, e.left, e.right
		}
	}

	return ComparisonTypeUnknown, nil, nil
}

func NewEquals(left, right impls.Expression) impls.Expression {
	return newComparison(left, right, ComparisonTypeEquals)
}

func NewIsDistinctFrom(left, right impls.Expression) impls.Expression {
	return newComparison(left, right, ComparisonTypeDistinctFrom)
}

func NewLessThanEquals(left, right impls.Expression) impls.Expression {
	return newComparison(left, right, ComparisonTypeLessThanEquals)
}

func NewLessThan(left, right impls.Expression) impls.Expression {
	return newComparison(left, right, ComparisonTypeLessThan)
}

func NewGreaterThanEquals(left, right impls.Expression) impls.Expression {
	return newComparison(left, right, ComparisonTypeGreaterThanEquals)
}

func NewGreaterThan(left, right impls.Expression) impls.Expression {
	return newComparison(left, right, ComparisonTypeGreaterThan)
}

func NewBetween(left, middle, right impls.Expression) impls.Expression {
	return NewAnd(NewLessThanEquals(middle, left), NewLessThanEquals(left, right))
}

func NewBetweenSymmetric(left, middle, right impls.Expression) impls.Expression {
	return NewOr(NewBetween(left, middle, right), NewBetween(left, right, middle))
}

func newComparison(left, right impls.Expression, comparisonType ComparisonType) impls.Expression {
	typeChecker := func(left types.Type, right types.Type) (types.Type, error) {
		if typ := left.PromoteToCommonType(right); typ != types.TypeUnknown {
			return types.TypeBool, nil
		}

		return types.TypeUnknown, fmt.Errorf("illegal operand types for comparison: %s and %s", left, right)
	}

	valueFrom := func(ctx impls.ExecutionContext, left, right impls.Expression, row rows.Row) (any, error) {
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
	}

	return newBinaryExpression(left, right, comparisonType.String(), typeChecker, valueFrom)
}
