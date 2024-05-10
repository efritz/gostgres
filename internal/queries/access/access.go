package access

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/queries/filter"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type accessNode struct {
	table    *table.Table
	filter   expressions.Expression
	order    expressions.OrderExpression
	strategy accessStrategy
}

var _ queries.Node = &accessNode{}

func NewAccess(table *table.Table) queries.Node {
	return &accessNode{
		table: table,
	}
}

func (n *accessNode) Name() string {
	return n.table.Name()
}

func (n *accessNode) Fields() []shared.Field {
	var fields []shared.Field
	for _, field := range n.table.Fields() {
		// TODO - should never not be the case?
		field := field.WithRelationName(n.table.Name())
		fields = append(fields, field.Field)
	}

	return fields
}

func (n *accessNode) Serialize(w io.Writer, indentationLevel int) {
	n.strategy.Serialize(w, indentationLevel)

	if n.filter != nil {
		io.WriteString(w, fmt.Sprintf("%sfilter: %s\n", serialization.Indent(indentationLevel+1), n.filter))
	}
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

func (n *accessNode) AddFilter(filterExpression expressions.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filterExpression)
}

func (n *accessNode) AddOrder(order expressions.OrderExpression) {
	n.order = order
}

func (n *accessNode) Filter() expressions.Expression {
	if filterExpression := n.strategy.Filter(); filterExpression != nil {
		return expressions.UnionFilters(n.filter, filterExpression)
	}

	return n.filter
}

func (n *accessNode) Ordering() expressions.OrderExpression {
	return n.strategy.Ordering()
}

func (n *accessNode) SupportsMarkRestore() bool {
	return false
}

func (n *accessNode) Scanner(ctx queries.Context) (scan.Scanner, error) {
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
