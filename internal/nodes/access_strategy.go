package nodes

import (
	"io"
	"reflect"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
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
	if filter != nil {
		// TODO - use btree for filters as well

		for _, index := range table.indexes {
			if ix, ok := index.(Index[hashIndexScanOptions]); ok {
				if hi, ok := ix.(*hashIndex); ok {
					// TODO - partial indexes

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
					// TODO - partial indexes
					indexExpressions := bt.expressions
					orderExpressions := order.Expressions()

					// TODO - check backwards scan
					if len(indexExpressions) < len(orderExpressions) || !reflect.DeepEqual(indexExpressions[:len(orderExpressions)], orderExpressions) {
						continue outer
					}

					lowerBound, upperBound := extractBounds(filter, bt.expressions)

					return NewIndexAccessStrategy(table, bt, btreeIndexScanOptions{
						lowerBound: lowerBound,
						upperBound: upperBound,
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

func extractBounds(filter expressions.Expression, indexedExprs []ExpressionWithDirection) (lowerBound, upperBound *scanBound) {
	if filter == nil {
		return nil, nil
	}

	target := indexedExprs[0].Expression

	// TODO - decompose multiple comparison (e.g., a < b and c < d where (a, c) is indexed)
	if comparisonType, left, right := expressions.IsComparison(filter); comparisonType != expressions.ComparisonTypeUnknown {
		bindsAllFields := func(fields []shared.Field, expr expressions.Expression) bool {
			for _, field := range expr.Fields() {
				if _, err := shared.FindMatchingFieldIndex(field, fields); err != nil {
					return false
				}
			}

			return true
		}

		if bindsAllFields(target.Fields(), left) {
			switch comparisonType {
			case expressions.ComparisonTypeEquals:
				lowerBound = &scanBound{expressions: []expressions.Expression{right}, inclusive: true}
				upperBound = &scanBound{expressions: []expressions.Expression{right}, inclusive: true}
			case expressions.ComparisonTypeLessThan:
				upperBound = &scanBound{expressions: []expressions.Expression{right}, inclusive: false}
			case expressions.ComparisonTypeLessThanEquals:
				upperBound = &scanBound{expressions: []expressions.Expression{right}, inclusive: true}
			case expressions.ComparisonTypeGreaterThan:
				lowerBound = &scanBound{expressions: []expressions.Expression{right}, inclusive: false}
			case expressions.ComparisonTypeGreaterThanEquals:
				lowerBound = &scanBound{expressions: []expressions.Expression{right}, inclusive: true}
			}

			return lowerBound, upperBound
		}
	}

	return nil, nil
}
