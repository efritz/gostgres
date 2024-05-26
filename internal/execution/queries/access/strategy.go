package access

import (
	"github.com/efritz/gostgres/internal/catalog/table/indexes"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/types"
)

type accessStrategy interface {
	Serialize(w serialization.IndentWriter)
	Filter() types.Expression
	Ordering() types.OrderExpression
	Scanner(ctx types.Context) (scan.Scanner, error)
}

func selectAccessStrategy(
	table types.Table,
	filterExpression types.Expression,
	orderExpression types.OrderExpression,
) accessStrategy {
	var candidates []accessStrategy
	for _, index := range table.Indexes() {
		if index, opts, ok := indexes.CanSelectHashIndex(index, filterExpression); ok {
			candidates = append(candidates, NewIndexAccessStrategy(table, index, opts))
		}

		if index, opts, ok := indexes.CanSelectBtreeIndex(index, filterExpression, orderExpression); ok {
			candidates = append(candidates, NewIndexAccessStrategy(table, index, opts))
		}
	}

	maxScore := 0
	bestStrategy := NewTableAccessStrategy(table)

	for _, index := range candidates {
		indexCost := 0
		if expressions.SubsumesOrder(orderExpression, index.Ordering()) {
			indexCost += 100
		}

		if filterExpression != nil {
			oldLen := len(expressions.Conjunctions(filterExpression))
			newLen := 0
			if remainingFilter := expressions.FilterDifference(filterExpression, index.Filter()); remainingFilter != nil {
				newLen = len(expressions.Conjunctions(remainingFilter))
			}

			indexCost += (oldLen - newLen) * 10
		}

		if indexCost > maxScore {
			bestStrategy = index
			maxScore = indexCost
		}
	}

	return bestStrategy
}
