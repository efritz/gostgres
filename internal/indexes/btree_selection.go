package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/table"
)

type btreeExpressioner interface {
	Expressions() []expressions.ExpressionWithDirection
}

func (i *btreeIndex) Expressions() []expressions.ExpressionWithDirection {
	return i.expressions
}

func CanSelectBtreeIndex(
	table *table.Table,
	index table.Index,
	filterExpression expressions.Expression,
	order expressions.OrderExpression,
) (_ Index[BtreeIndexScanOptions], opts BtreeIndexScanOptions, _ bool) {
	if !matchesPartial(index, filterExpression) {
		return nil, opts, false
	}

	btreeIndex, ok := index.(Index[BtreeIndexScanOptions])
	if !ok {
		return nil, opts, false
	}
	btreeExpressioner, ok := btreeIndex.Unwrap().(btreeExpressioner)
	if !ok {
		return nil, opts, false
	}
	btreeExpressions := btreeExpressioner.Expressions()

	scanDirection := scanDirection(order, btreeExpressions)
	if scanDirection == ScanDirectionUnknown {
		scanDirection = ScanDirectionForward
	}

	lowerBounds, upperBounds := extractBounds(filterExpression, btreeExpressions)

	opts = BtreeIndexScanOptions{
		scanDirection: scanDirection,
		lowerBounds:   lowerBounds,
		upperBounds:   upperBounds,
	}
	return btreeIndex, opts, true
}

func scanDirection(order expressions.OrderExpression, indexDirections []expressions.ExpressionWithDirection) ScanDirection {
	if order == nil {
		return ScanDirectionUnknown
	}
	orderExpressions := order.Expressions()

	if len(orderExpressions) > len(indexDirections) {
		return ScanDirectionUnknown
	}

	var forward bool
	for i, orderExpr := range orderExpressions {
		if !orderExpr.Expression.Equal(indexDirections[i].Expression) {
			return ScanDirectionUnknown
		}

		matchesDirection := orderExpr.Reverse == indexDirections[i].Reverse

		if i == 0 {
			forward = matchesDirection
		} else if forward != matchesDirection {
			return ScanDirectionUnknown
		}
	}

	if !forward {
		return ScanDirectionBackward
	}

	return ScanDirectionForward
}

func extractBounds(filter expressions.Expression, indexedExprs []expressions.ExpressionWithDirection) (lowerBounds, upperBounds [][]scanBound) {
	if filter == nil {
		return nil, nil
	}

	conjunctions := expressions.Conjunctions(filter)

	for _, target := range indexedExprs {
		var (
			exprLowerBounds []scanBound
			exprUpperBounds []scanBound
		)

		for _, conjunction := range conjunctions {
			if comparisonType, left, right := expressions.IsComparison(conjunction); comparisonType != expressions.ComparisonTypeUnknown {
				if right.Equal(target.Expression) {
					left, right = right, left
					comparisonType = comparisonType.Flip()
				}

				if left.Equal(target.Expression) {
					switch comparisonType {
					case expressions.ComparisonTypeEquals:
						exprLowerBounds = append(exprLowerBounds, scanBound{expression: right, inclusive: true})
						exprUpperBounds = append(exprUpperBounds, scanBound{expression: right, inclusive: true})
					case expressions.ComparisonTypeLessThan:
						exprUpperBounds = append(exprUpperBounds, scanBound{expression: right, inclusive: false})
					case expressions.ComparisonTypeLessThanEquals:
						exprUpperBounds = append(exprUpperBounds, scanBound{expression: right, inclusive: true})
					case expressions.ComparisonTypeGreaterThan:
						exprLowerBounds = append(exprLowerBounds, scanBound{expression: right, inclusive: false})
					case expressions.ComparisonTypeGreaterThanEquals:
						exprLowerBounds = append(exprLowerBounds, scanBound{expression: right, inclusive: true})
					}
				}
			}
		}

		lowerBounds = append(lowerBounds, exprLowerBounds)
		upperBounds = append(upperBounds, exprUpperBounds)
	}

	prunedLowerBounds := lowerBounds[:0]
	for _, bound := range lowerBounds {
		if len(bound) == 0 {
			break
		}

		prunedLowerBounds = append(prunedLowerBounds, bound)
	}

	prunedUpperBounds := upperBounds[:0]
	for _, bound := range upperBounds {
		if len(bound) == 0 {
			break
		}

		prunedUpperBounds = append(prunedUpperBounds, bound)
	}

	return prunedLowerBounds, prunedUpperBounds
}
