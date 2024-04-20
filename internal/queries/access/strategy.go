package access

import (
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/table"
)

type accessStrategy interface {
	Serialize(w io.Writer, indentationLevel int)
	Filter() expressions.Expression
	Ordering() expressions.OrderExpression
	Scanner(ctx scan.ScanContext) (scan.Scanner, error)
}

func selectAccessStrategy(
	table *table.Table,
	filter expressions.Expression,
	order expressions.OrderExpression,
) accessStrategy {
	var candidates []accessStrategy
	for _, index := range table.Indexes() {
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

// TODO - deduplicate

func filterDifference(filter, childFilter expressions.Expression) expressions.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []expressions.Expression) {
		for i, f1 := range conjunctions {
			for _, f2 := range childConjunctions {
				if f1.Equal(f2) {
					conjunctions[i] = nil
					break
				}
			}
		}
	})
}

func combineFilters(filter, childFilter expressions.Expression, filterConjunctions func(conjunctions, childConjunctions []expressions.Expression)) expressions.Expression {
	if filter == nil {
		return nil
	}
	if childFilter == nil {
		return filter
	}

	conjunctions := filter.Conjunctions()
	filterConjunctions(conjunctions, childFilter.Conjunctions())
	return unionFilters(conjunctions...)
}

func unionFilters(filters ...expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filters {
		if expression == nil {
			continue
		}

		conjunctions = append(conjunctions, expression.Conjunctions()...)
	}
	if len(conjunctions) == 0 {
		return nil
	}

	for i, c1 := range conjunctions {
		for j, c2 := range conjunctions {
			if c1 == nil || c2 == nil || j <= i {
				continue
			}

			if c1.Equal(c2) {
				conjunctions[j] = nil
			}
		}
	}

	filter := conjunctions[0]
	for _, conjunction := range conjunctions[1:] {
		if conjunction == nil {
			continue
		}

		filter = expressions.NewAnd(filter, conjunction)
	}

	return filter
}

func subsumesOrder(a, b expressions.OrderExpression) bool {
	if a == nil || b == nil {
		return false
	}

	aExpressions := a.Expressions()
	bExpressions := b.Expressions()
	if len(bExpressions) < len(aExpressions) {
		return false
	}

	for i, expression := range aExpressions {
		if expression.Reverse != bExpressions[i].Reverse {
			return false
		}

		if !expression.Expression.Equal(bExpressions[i].Expression) {
			return false
		}
	}

	return true
}
