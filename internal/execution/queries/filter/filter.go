package filter

import (
	"github.com/efritz/gostgres/internal/execution"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
)

type filterNode struct {
	queries.Node
	filter expressions.Expression
}

var _ queries.Node = &filterNode{}

func NewFilter(node queries.Node, filter expressions.Expression) queries.Node {
	return &filterNode{
		Node:   node,
		filter: filter,
	}
}

func (n *filterNode) Serialize(w serialization.IndentWriter) {
	if n.filter == nil {
		n.Node.Serialize(w)
	} else {
		w.WritefLine("filter by %s", n.filter)
		n.Node.Serialize(w.Indent())
	}
}

func (n *filterNode) AddFilter(filter expressions.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filter)
}

func (n *filterNode) AddOrder(order expressions.OrderExpression) {
	n.Node.AddOrder(order)
}

func (n *filterNode) Optimize() {
	if n.filter != nil {
		n.filter = n.filter.Fold()
		n.Node.AddFilter(n.filter)
	}

	n.Node.Optimize()
	n.filter = expressions.FilterDifference(n.filter, n.Node.Filter())
}

func (n *filterNode) Filter() expressions.Expression {
	return expressions.UnionFilters(n.filter, n.Node.Filter())
}

func (n *filterNode) Ordering() expressions.OrderExpression {
	return n.Node.Ordering()
}

func (n *filterNode) SupportsMarkRestore() bool {
	return false
}

func (n *filterNode) Scanner(ctx execution.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return NewFilterScanner(ctx, scanner, n.filter)
}
