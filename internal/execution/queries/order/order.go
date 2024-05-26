package order

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/types"
)

type orderNode struct {
	queries.Node
	order expressions.OrderExpression
}

var _ queries.Node = &orderNode{}

func NewOrder(node queries.Node, order expressions.OrderExpression) queries.Node {
	return &orderNode{
		Node:  node,
		order: order,
	}
}

func (n *orderNode) Serialize(w serialization.IndentWriter) {
	if n.order == nil {
		n.Node.Serialize(w)
	} else {
		w.WritefLine("order by %s", n.order)
		n.Node.Serialize(w.Indent())
	}
}

func (n *orderNode) AddFilter(filter types.Expression) {
	n.Node.AddFilter(filter)
}

func (n *orderNode) AddOrder(order expressions.OrderExpression) {
	// We are nested in a parent sort and un-separated by an ordering boundary
	// (such as limit or offset). We'll ignore our old sort criteria and adopt
	// our parent since the ordering of rows at this point in the query should
	// not have an effect on the result.
	n.order = order
}

func (n *orderNode) Optimize() {
	if n.order != nil {
		n.order = n.order.Fold()
		n.Node.AddOrder(n.order)
	}

	n.Node.Optimize()

	if expressions.SubsumesOrder(n.order, n.Node.Ordering()) {
		n.order = nil
	}
}

func (n *orderNode) Ordering() expressions.OrderExpression {
	if n.order == nil {
		return n.Node.Ordering()
	}

	return n.order
}

func (n *orderNode) SupportsMarkRestore() bool {
	return true
}

func (n *orderNode) Scanner(ctx types.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	// TODO - commented out to support mark-restore
	// if n.order == nil {
	// 	return scanner, nil
	// }

	return NewOrderScanner(ctx, scanner, n.Fields(), n.order)
}
