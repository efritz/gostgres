package indexes

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/types"
)

type hashExpressioner interface {
	HashExpression() types.Expression
}

func (i *hashIndex) HashExpression() types.Expression {
	return i.expression
}

func CanSelectHashIndex(
	index types.BaseIndex,
	filterExpression types.Expression,
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

	for _, conjunction := range expressions.Conjunctions(filterExpression) {
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
