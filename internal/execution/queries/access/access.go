package access

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/filter"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type accessNode struct {
	table    impls.Table
	filter   impls.Expression
	order    impls.OrderExpression
	strategy accessStrategy
}

var _ queries.Node = &accessNode{}

func NewAccess(table impls.Table) queries.Node {
	return &accessNode{
		table: table,
	}
}

func (n *accessNode) Name() string {
	return n.table.Name()
}

func (n *accessNode) Fields() []fields.Field {
	var fields []fields.Field
	for _, field := range n.table.Fields() {
		// TODO - should never not be the case?
		field := field.WithRelationName(n.table.Name())
		fields = append(fields, field.Field)
	}

	return fields
}

func (n *accessNode) Serialize(w serialization.IndentWriter) {
	n.strategy.Serialize(w)

	if n.filter != nil {
		w.Indent().WritefLine("filter: %s", n.filter)
	}
}

func (n *accessNode) AddFilter(filterExpression impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *accessNode) AddOrder(order impls.OrderExpression) {
	n.order = order
}

func (n *accessNode) Optimize() {
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

func (n *accessNode) Filter() impls.Expression {
	if filterExpression := n.strategy.Filter(); filterExpression != nil {
		return expressions.UnionFilters(n.filter, filterExpression)
	}

	return n.filter
}

func (n *accessNode) Ordering() impls.OrderExpression {
	return n.strategy.Ordering()
}

func (n *accessNode) SupportsMarkRestore() bool {
	return false
}

func (n *accessNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Access Node scanner")

	scanner, err := n.strategy.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	if n.filter != nil {
		scanner, err = filter.NewFilterScanner(ctx, scanner, n.filter)
		if err != nil {
			return nil, err
		}
	}

	return scanner, nil
}
