package opt

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalProjectionNode struct {
	LogicalNode
	projection *projection.Projection
}

func NewProjection(node LogicalNode, projection *projection.Projection) LogicalNode {
	return &logicalProjectionNode{
		LogicalNode: node,
		projection:  projection,
	}
}

func (n *logicalProjectionNode) Fields() []fields.Field {
	return n.projection.Fields()
}

func (n *logicalProjectionNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.LogicalNode.AddFilter(ctx, n.projection.DeprojectExpression(filter))
}

func (n *logicalProjectionNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projection.DeprojectExpression(expression), nil
	})

	n.LogicalNode.AddOrder(ctx, mapped)
}

func (n *logicalProjectionNode) Optimize(ctx impls.OptimizationContext) {
	n.projection.Optimize(ctx)
	n.LogicalNode.Optimize(ctx)
}

func (n *logicalProjectionNode) Filter() impls.Expression {
	return n.projection.ProjectExpression(n.LogicalNode.Filter())
}

func (n *logicalProjectionNode) Ordering() impls.OrderExpression {
	ordering := n.LogicalNode.Ordering()
	if ordering == nil {
		return nil
	}

	mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projection.ProjectExpression(expression), nil
	})

	return mapped
}

func (n *logicalProjectionNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalProjectionNode) Build() nodes.Node {
	return nodes.NewProjection(n.LogicalNode.Build(), n.projection)
}
