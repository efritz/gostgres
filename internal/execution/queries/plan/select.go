package plan

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalSelectNode struct {
	LogicalNode
}

func NewSelect(node LogicalNode) LogicalNode {
	return &logicalSelectNode{
		LogicalNode: node,
	}
}

func (n *logicalSelectNode) Name() string {
	return n.LogicalNode.Name()
}

func (n *logicalSelectNode) Fields() []fields.Field {
	return n.LogicalNode.Fields()
}

func (n *logicalSelectNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.LogicalNode.AddFilter(ctx, filter)
}

func (n *logicalSelectNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	n.LogicalNode.AddOrder(ctx, order)
}

func (n *logicalSelectNode) Optimize(ctx impls.OptimizationContext) {
	n.LogicalNode.Optimize(ctx)
}

func (n *logicalSelectNode) Filter() impls.Expression {
	return n.LogicalNode.Filter()
}

func (n *logicalSelectNode) Ordering() impls.OrderExpression {
	return n.LogicalNode.Ordering()
}

func (n *logicalSelectNode) SupportsMarkRestore() bool {
	return n.LogicalNode.SupportsMarkRestore()
}

func (n *logicalSelectNode) Build() nodes.Node {
	return n.LogicalNode.Build()
}
