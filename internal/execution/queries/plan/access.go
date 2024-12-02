package plan

import (
	"github.com/efritz/gostgres/internal/catalog/table/indexes"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/access"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalAccessNode struct {
	table    impls.Table
	filter   impls.Expression
	order    impls.OrderExpression
	strategy nodes.AccessStrategy
}

func NewAccess(table impls.Table) LogicalNode {
	return &logicalAccessNode{
		table: table,
	}
}

func (n *logicalAccessNode) Name() string {
	return n.table.Name()
}

func (n *logicalAccessNode) Fields() []fields.Field {
	var fields []fields.Field
	for _, field := range n.table.Fields() {
		// TODO - should never not be the case?
		field := field.WithRelationName(n.table.Name())
		fields = append(fields, field.Field)
	}

	return fields
}

func (n *logicalAccessNode) AddFilter(ctx impls.OptimizationContext, filterExpression impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *logicalAccessNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	n.order = order
}

func (n *logicalAccessNode) Optimize(ctx impls.OptimizationContext) {
	if n.filter != nil {
		n.filter = n.filter.Fold()
	}

	if n.order != nil {
		n.order = n.order.Fold()
	}

	n.strategy = selectAccessStrategy(n.table, n.filter, n.order)
	n.filter = expressions.FilterDifference(n.filter, n.strategy.Filter())
	n.order = nil
}

func (n *logicalAccessNode) Filter() impls.Expression {
	if filterExpression := n.strategy.Filter(); filterExpression != nil {
		return expressions.UnionFilters(n.filter, filterExpression)
	}

	return n.filter
}

func (n *logicalAccessNode) Ordering() impls.OrderExpression {
	return n.strategy.Ordering()
}

func (n *logicalAccessNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalAccessNode) Build() nodes.Node {
	if f := n.filter; f != nil {
		n.filter = nil
		return NewFilter(n, f).Build()
	}

	return nodes.NewAccess(n.strategy)
}

//
//

func selectAccessStrategy(
	table impls.Table,
	filterExpression impls.Expression,
	orderExpression impls.OrderExpression,
) nodes.AccessStrategy {
	var candidates []nodes.AccessStrategy
	for _, index := range table.Indexes() {
		if index, opts, ok := indexes.CanSelectHashIndex(index, filterExpression); ok {
			candidates = append(candidates, access.NewIndexAccessStrategy(table, index, opts))
		}

		if index, opts, ok := indexes.CanSelectBtreeIndex(index, filterExpression, orderExpression); ok {
			candidates = append(candidates, access.NewIndexAccessStrategy(table, index, opts))
		}
	}

	maxScore := 0
	bestStrategy := access.NewTableAccessStrategy(table)

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
