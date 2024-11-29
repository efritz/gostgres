package order

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalOrderNode struct {
	queries.LogicalNode
	order impls.OrderExpression
}

var _ queries.LogicalNode = &logicalOrderNode{}

func NewOrder(node queries.LogicalNode, order impls.OrderExpression) queries.LogicalNode {
	return &logicalOrderNode{
		LogicalNode: node,
		order:       order,
	}
}

func (n *logicalOrderNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.LogicalNode.AddFilter(ctx, filter)
}

func (n *logicalOrderNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	// We are nested in a parent sort and un-separated by an ordering boundary
	// (such as limit or offset). We'll ignore our old sort criteria and adopt
	// our parent since the ordering of rows at this point in the query should
	// not have an effect on the result.
	n.order = order
}

func (n *logicalOrderNode) Optimize(ctx impls.OptimizationContext) {
	if n.order != nil {
		n.order = n.order.Fold()
		n.LogicalNode.AddOrder(ctx, n.order)
	}

	n.LogicalNode.Optimize(ctx)

	if expressions.SubsumesOrder(n.order, n.LogicalNode.Ordering()) {
		n.order = nil
	}
}

func (n *logicalOrderNode) Ordering() impls.OrderExpression {
	if n.order == nil {
		return n.LogicalNode.Ordering()
	}

	return n.order
}

func (n *logicalOrderNode) SupportsMarkRestore() bool {
	return true
}

func (n *logicalOrderNode) Build() queries.Node {
	return &orderNode{
		Node:   n.LogicalNode.Build(),
		order:  n.order,
		fields: n.Fields(),
	}
}

//
//

type orderNode struct {
	queries.Node
	order  impls.OrderExpression
	fields []fields.Field
}

var _ queries.Node = &orderNode{}

func (n *orderNode) Serialize(w serialization.IndentWriter) {
	if n.order == nil {
		n.Node.Serialize(w)
	} else {
		w.WritefLine("order by %s", n.order)
		n.Node.Serialize(w.Indent())
	}
}

func (n *orderNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Order Node scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	// TODO - commented out to support mark-restore
	// if n.order == nil {
	// 	return scanner, nil
	// }

	return NewOrderScanner(ctx, scanner, n.fields, n.order)
}
