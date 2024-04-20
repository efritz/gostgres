package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
)

type hashExpressioner interface {
	HashExpression() expressions.Expression
}

func (i *hashIndex) HashExpression() expressions.Expression {
	return i.expression
}

func CanSelectHashIndex(
	table TableIndexer,
	index BaseIndex,
	filterExpression expressions.Expression,
) (_ Index[hashIndexScanOptions], opts hashIndexScanOptions, _ bool) {
	if !matchesPartial(index, filterExpression) {
		return nil, opts, false
	}

	if filterExpression == nil {
		return nil, opts, false
	}

	hashIndex, ok := index.(Index[hashIndexScanOptions])
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
				return hashIndex, hashIndexScanOptions{expression: right}, true
			}

			if right.Equal(hashExpression) {
				return hashIndex, hashIndexScanOptions{expression: left}, true
			}
		}
	}

	return nil, opts, false
}
