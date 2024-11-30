package opt

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalLimitNode struct {
	LogicalNode
	limit int
}

func NewLimit(node LogicalNode, limit int) LogicalNode {
	return &logicalLimitNode{
		LogicalNode: node,
		limit:       limit,
	}
}

func (n *logicalLimitNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {} // boundary
func (n *logicalLimitNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {} // boundary
func (n *logicalLimitNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalLimitNode) Build() nodes.Node {
	return nodes.NewLimit(n.LogicalNode.Build(), n.limit)
}
