package nodes

import (
	"io"

	"github.com/efritz/gostgres/internal/expressions"
)

type accessStrategy interface {
	Serialize(w io.Writer, indentationLevel int)
	Filter() expressions.Expression
	Ordering() OrderExpression
	Scanner(ctx ScanContext) (Scanner, error)
}

func selectAccessStrategy(
	table *Table,
	filter expressions.Expression,
	order OrderExpression,
) accessStrategy {
	// TODO - partial indexes

	if filter != nil {
		// TODO - use btree for filters as well

		for _, index := range table.indexes {
			if ix, ok := index.(Index[hashIndexScanOptions]); ok {
				if hi, ok := ix.(*hashIndex); ok {
					if expr, ok := extractEquality(filter, hi.expression); ok {
						return NewIndexAccessStrategy(table, hi, hashIndexScanOptions{
							expression: expr,
						})
					}
				}
			}
		}
	}

	if order != nil {
	outer:
		for _, index := range table.indexes {
			if ix, ok := index.(Index[btreeIndexScanOptions]); ok {
				if bt, ok := ix.(*btreeIndex); ok {
					indexExpressions := bt.expressions
					orderExpressions := order.Expressions()

					scanDirection, ok := scanDirection(orderExpressions, indexExpressions)
					if !ok {
						continue outer
					}

					lowerBounds, upperBounds := extractBounds(filter, bt.expressions)

					return NewIndexAccessStrategy(table, bt, btreeIndexScanOptions{
						scanDirection: scanDirection,
						lowerBounds:   lowerBounds,
						upperBounds:   upperBounds,
					})
				}
			}
		}
	}

	return NewTableAccessStrategy(table)
}

func scanDirection(orderExpressions, indexDirections []ExpressionWithDirection) (ScanDirection, bool) {
	if len(orderExpressions) > len(indexDirections) {
		return ScanDirectionUnknown, false
	}

	var forward bool
	for i, orderExpr := range orderExpressions {
		if !orderExpr.Expression.Equal(indexDirections[i].Expression) {
			return ScanDirectionUnknown, false
		}

		matchesDirection := orderExpr.Reverse == indexDirections[i].Reverse

		if i == 0 {
			forward = matchesDirection
		} else if forward != matchesDirection {
			return ScanDirectionUnknown, false
		}
	}

	if !forward {
		return ScanDirectionBackward, true
	}

	return ScanDirectionForward, true
}

func extractEquality(filter expressions.Expression, indexedExpr expressions.Expression) (_ expressions.Expression, ok bool) {
	if filter == nil {
		return nil, false
	}

	if comparisonType, left, right := expressions.IsComparison(filter); comparisonType == expressions.ComparisonTypeEquals {
		if left.Equal(indexedExpr) {
			return right, true
		}

		if right.Equal(indexedExpr) {
			return left, true
		}
	}

	return nil, false
}

func extractBounds(filter expressions.Expression, indexedExprs []ExpressionWithDirection) (lowerBounds, upperBounds [][]scanBound) {
	if filter == nil {
		return nil, nil
	}

	conjunctions := filter.Conjunctions()

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
