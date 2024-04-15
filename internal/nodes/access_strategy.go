package nodes

import (
	"io"
	"reflect"

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
					if values, ok := extractEquality(filter, hi.expressions); ok {
						return NewIndexAccessStrategy(table, hi, hashIndexScanOptions{
							expressions: values,
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

					// TODO - check backwards scan
					if len(indexExpressions) < len(orderExpressions) || !reflect.DeepEqual(indexExpressions[:len(orderExpressions)], orderExpressions) {
						continue outer
					}

					lowerBounds, upperBounds := extractBounds(filter, bt.expressions)

					return NewIndexAccessStrategy(table, bt, btreeIndexScanOptions{
						lowerBounds: lowerBounds,
						upperBounds: upperBounds,
					})
				}
			}
		}
	}

	return NewTableAccessStrategy(table)
}

func extractEquality(filter expressions.Expression, indexedExprs []expressions.Expression) (_ []expressions.Expression, ok bool) {
	// TODO - support multi-column
	if filter == nil || len(indexedExprs) > 1 {
		return nil, false
	}

	for _, indexedExpr := range indexedExprs {
		if comparisonType, left, right := expressions.IsComparison(filter); comparisonType == expressions.ComparisonTypeEquals {
			// TODO - need a more stable way to compare expression equality
			if reflect.DeepEqual(left, indexedExpr) {
				return []expressions.Expression{right}, true
			} else if reflect.DeepEqual(right, indexedExpr) {
				return []expressions.Expression{left}, true
			}
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
