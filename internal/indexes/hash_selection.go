package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/table"
)

type hashExpressioner interface {
	HashExpression() expressions.Expression
}

func (i *hashIndex) HashExpression() expressions.Expression {
	return i.expression
}

func CanSelectHashIndex(
	table *table.Table,
	index table.Index,
	filterExpression expressions.Expression,
) (_ Index[HashIndexScanOptions], opts HashIndexScanOptions, _ bool) {
	if !matchesPartial(index, filterExpression) {
		return nil, opts, false
	}

	if filterExpression == nil {
		return nil, opts, false
	}

	hashIndex, ok := index.(Index[HashIndexScanOptions])
	if !ok {
		return nil, opts, false
	}
	hashExpressioner, ok := hashIndex.Unwrap().(hashExpressioner)
	if !ok {
		return nil, opts, false
	}
	hashExpression := hashExpressioner.HashExpression()

	for _, conjunction := range filterExpression.Conjunctions() {
		if comparisonType, left, right := expressions.IsComparison(conjunction); comparisonType == expressions.ComparisonTypeEquals {
			if left.Equal(hashExpression) {
				return hashIndex, HashIndexScanOptions{expression: right}, true
			}

			if right.Equal(hashExpression) {
				return hashIndex, HashIndexScanOptions{expression: left}, true
			}
		}
	}

	return nil, opts, false
}
