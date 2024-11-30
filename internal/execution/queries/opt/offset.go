package opt

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalOffsetNode struct {
	LogicalNode
	offset int
}

func NewOffset(node LogicalNode, offset int) LogicalNode {
	return &logicalOffsetNode{
		LogicalNode: node,
		offset:      offset,
	}
}

func (n *logicalOffsetNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {} // boundary
func (n *logicalOffsetNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {} // boundary
func (n *logicalOffsetNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalOffsetNode) Build() nodes.Node {
	return nodes.NewOffset(n.LogicalNode.Build(), n.offset)
}
