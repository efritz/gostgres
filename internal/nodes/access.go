package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type accessNode struct {
	table    *Table
	filter   expressions.Expression
	order    expressions.OrderExpression
	strategy accessStrategy
}

var _ Node = &accessNode{}

func NewData(table *Table) Node {
	return &accessNode{
		table: table,
	}
}

func (n *accessNode) Name() string {
	return n.table.name
}

func (n *accessNode) Fields() []shared.Field {
	return updateRelationName(n.table.Fields(), n.table.name)
}

func (n *accessNode) Serialize(w io.Writer, indentationLevel int) {
	n.strategy.Serialize(w, indentationLevel)

	if n.filter != nil {
		io.WriteString(w, fmt.Sprintf("%sfilter: %s\n", indent(indentationLevel+1), n.filter))
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
	n.filter = filterDifference(n.filter, n.strategy.Filter())
	n.order = nil
}

func (n *accessNode) AddFilter(filter expressions.Expression) {
	n.filter = unionFilters(n.filter, filter)
}

func (n *accessNode) AddOrder(order expressions.OrderExpression) {
	n.order = order
}

func (n *accessNode) Filter() expressions.Expression {
	if filter := n.strategy.Filter(); filter != nil {
		return unionFilters(n.filter, filter)
	}

	return n.filter
}

func (n *accessNode) Ordering() expressions.OrderExpression {
	return n.strategy.Ordering()
}

func (n *accessNode) SupportsMarkRestore() bool {
	return false
}

func (n *accessNode) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	scanner, err := n.strategy.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	if n.filter != nil {
		scanner, err = NewFilterScanner(ctx, scanner, n.filter)
		if err != nil {
			return nil, err
		}
	}

	return scanner, nil
}
