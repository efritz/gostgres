package access

import (
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/table"
)

type accessStrategy interface {
	Serialize(w io.Writer, indentationLevel int)
	Filter() expressions.Expression
	Ordering() expressions.OrderExpression
	Scanner(ctx queries.Context) (scan.Scanner, error)
}

func selectAccessStrategy(
	table *table.Table,
	filterExpression expressions.Expression,
	orderExpression expressions.OrderExpression,
) accessStrategy {
	var candidates []accessStrategy
	for _, index := range table.Indexes() {
		if index, opts, ok := indexes.CanSelectHashIndex(table, index, filterExpression); ok {
			candidates = append(candidates, NewIndexAccessStrategy(table, index, opts))
		}

		if index, opts, ok := indexes.CanSelectBtreeIndex(table, index, filterExpression, orderExpression); ok {
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
			oldLen := len(filterExpression.Conjunctions())
			newLen := 0
			if remainingFilter := expressions.FilterDifference(filterExpression, index.Filter()); remainingFilter != nil {
				newLen = len(remainingFilter.Conjunctions())
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
