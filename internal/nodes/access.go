package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type accessNode struct {
	table    *Table
	filter   expressions.Expression
	order    OrderExpression
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
	if n.order != nil {
		io.WriteString(w, fmt.Sprintf("%sorder: %s\n", indent(indentationLevel+1), n.order))
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
	if subsumesOrder(n.order, n.strategy.Ordering()) {
		n.order = nil
	}
}

func (n *accessNode) AddFilter(filter expressions.Expression) {
	n.filter = unionFilters(n.filter, filter)
}

func (n *accessNode) AddOrder(order OrderExpression) {
	n.order = order
}

func (n *accessNode) Filter() expressions.Expression {
	if filter := n.strategy.Filter(); filter != nil {
		return filter
	}

	return n.filter
}

func (n *accessNode) Ordering() OrderExpression {
	if order := n.strategy.Ordering(); order != nil {
		return order
	}

	return n.order
}

func (n *accessNode) SupportsMarkRestore() bool {
	return false
}

func (n *accessNode) Scanner(ctx ScanContext) (Scanner, error) {
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

	if n.order != nil {
		scanner, err = NewOrderScanner(ctx, scanner, n.Fields(), n.order)
		if err != nil {
			return nil, err
		}
	}

	return scanner, nil
}
