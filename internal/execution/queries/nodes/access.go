package nodes

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalAccessNode struct {
	table    impls.Table
	filter   impls.Expression
	order    impls.OrderExpression
	strategy accessStrategy
}

var _ LogicalNode = &logicalAccessNode{}

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

func (n *logicalAccessNode) Build() Node {
	if f := n.filter; f != nil {
		n.filter = nil
		return NewFilter(n, f).Build()
	}

	return &accessNode{
		table:    n.table,
		strategy: n.strategy,
	}
}

//
//

type accessNode struct {
	table    impls.Table
	strategy accessStrategy
}

var _ Node = &accessNode{}

func (n *accessNode) Serialize(w serialization.IndentWriter) {
	n.strategy.Serialize(w)
}

func (n *accessNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	return n.strategy.Scanner(ctx)
}
