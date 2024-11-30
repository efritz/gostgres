package opt

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalOrderNode struct {
	LogicalNode
	order impls.OrderExpression
}

func NewOrder(node LogicalNode, order impls.OrderExpression) LogicalNode {
	return &logicalOrderNode{
		LogicalNode: node,
		order:       order,
	}
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

func (n *logicalOrderNode) Build() nodes.Node {
	if n.order == nil {
		return n.LogicalNode.Build()
	}

	return nodes.NewOrder(n.LogicalNode.Build(), n.order, n.Fields())
}
