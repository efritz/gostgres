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
	Scanner() (Scanner, error)
}

func selectAccessStrategy(
	table *Table,
	filter expressions.Expression,
	order OrderExpression,
) accessStrategy {
	if filter != nil {
		for _, index := range table.indexes {
			if ix, ok := index.(Index[hashIndexScanOptions]); ok {
				if hi, ok := ix.(*hashIndex); ok {
					// TODO - partial indexes
					if values, expr, ok := extractEquality(filter, hi.expressions); ok {
						return NewIndexAccessStrategy(table, hi, hashIndexScanOptions{
							values: values,
							expr:   expr,
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

					lowerBound, upperBound, expr := extractBounds(filter, bt.expressions)

					return NewIndexAccessStrategy(table, bt, btreeIndexScanOptions{
						lowerBound: lowerBound,
						upperBound: upperBound,
						expr:       expr,
					})
				}
			}
		}
	}

	return NewTableAccessStrategy(table)
}

func extractEquality(filter expressions.Expression, indexedExprs []expressions.Expression) (values []interface{}, expr expressions.Expression, ok bool) {
	// TODO - support multi-column
	if filter == nil || len(indexedExprs) > 1 {
		return nil, nil, false
	}

	for _, indexedExpr := range indexedExprs {
		if comparisonType, left, right := expressions.IsComparison(filter); comparisonType == expressions.ComparisonTypeEquals {
			// TODO - need a more stable way to compare expression equality
			if reflect.DeepEqual(left, indexedExpr) {
				value, err := right.ValueFrom(shared.Row{})
				if err != nil {
					return nil, nil, false
				}

				return []interface{}{value}, filter, true
			} else if reflect.DeepEqual(right, indexedExpr) {
				value, err := left.ValueFrom(shared.Row{})
				if err != nil {
					return nil, nil, false
				}

				return []interface{}{value}, filter, true
			}
		}
	}

	return nil, nil, false

}

func extractBounds(filter expressions.Expression, indexedExprs []ExpressionWithDirection) (lowerBound, upperBound *scanBound, expr expressions.Expression) {
	if filter == nil {
		return nil, nil, nil
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
			// TODO - need to keep as expression
			value, err := right.ValueFrom(shared.Row{})
			if err != nil {
				return nil, nil, nil
			}
			values := []interface{}{value}

			switch comparisonType {
			case expressions.ComparisonTypeEquals:
				lowerBound = &scanBound{values: values, inclusive: true}
				upperBound = &scanBound{values: values, inclusive: true}
			case expressions.ComparisonTypeLessThan:
				upperBound = &scanBound{values: values, inclusive: false}
			case expressions.ComparisonTypeLessThanEquals:
				upperBound = &scanBound{values: values, inclusive: true}
			case expressions.ComparisonTypeGreaterThan:
				lowerBound = &scanBound{values: values, inclusive: false}
			case expressions.ComparisonTypeGreaterThanEquals:
				lowerBound = &scanBound{values: values, inclusive: true}
			}

			return lowerBound, upperBound, filter
		}
	}

	return nil, nil, nil
}
