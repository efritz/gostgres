package nodes

import (
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/scan"
)

type accessStrategy interface {
	Serialize(w io.Writer, indentationLevel int)
	Filter() expressions.Expression
	Ordering() expressions.OrderExpression
	Scanner(ctx scan.ScanContext) (scan.Scanner, error)
}

func selectAccessStrategy(
	table *Table,
	filter expressions.Expression,
	order expressions.OrderExpression,
) accessStrategy {
	var candidates []accessStrategy
	for _, index := range table.indexes {
		if index, opts, ok := indexes.CanSelectHashIndex(table, index, filter); ok {
			candidates = append(candidates, NewIndexAccessStrategy(table, index, opts))
		}

		if index, opts, ok := indexes.CanSelectBtreeIndex(table, index, filter, order); ok {
			candidates = append(candidates, NewIndexAccessStrategy(table, index, opts))
		}
	}

	maxScore := 0
	bestStrategy := NewTableAccessStrategy(table)

	for _, index := range candidates {
		indexCost := 0
		if subsumesOrder(order, index.Ordering()) {
			indexCost += 100
		}

		if filter != nil {
			oldLen := len(filter.Conjunctions())
			newLen := 0
			if remainingFilter := filterDifference(filter, index.Filter()); remainingFilter != nil {
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
