package plan

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
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
	node := n.LogicalNode.Build()
	if n.order != nil {
		node = nodes.NewOrder(node, n.order, n.Fields())
	}

	return node
}

//
//

func lowerOrder(ctx impls.OptimizationContext, order impls.OrderExpression, nodes ...LogicalNode) {
	orderExpressions := order.Expressions()

	for _, node := range nodes {
		filteredExpressions := make([]impls.ExpressionWithDirection, 0, len(orderExpressions))
	exprLoop:
		for _, expression := range orderExpressions {
			for _, field := range expressions.Fields(expression.Expression) {
				if _, err := fields.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					continue exprLoop
				}
			}

			filteredExpressions = append(filteredExpressions, expression)
		}

		if len(filteredExpressions) != 0 {
			node.AddOrder(ctx, expressions.NewOrderExpression(filteredExpressions))
		}
	}
}
