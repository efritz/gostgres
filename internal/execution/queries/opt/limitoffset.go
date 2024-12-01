package opt

import (
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalLimitNode struct {
	LogicalNode
	limit  *int
	offset *int
}

func NewLimitOffset(node LogicalNode, limit *int, offset *int) LogicalNode {
	return &logicalLimitNode{
		LogicalNode: node,
		limit:       limit,
		offset:      offset,
	}
}

func (n *logicalLimitNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {} // boundary
func (n *logicalLimitNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {} // boundary
func (n *logicalLimitNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalLimitNode) Build() nodes.Node {
	node := n.LogicalNode.Build()
	if n.offset != nil {
		node = nodes.NewOffset(node, *n.offset)
	}
	if n.limit != nil {
		node = nodes.NewLimit(node, *n.limit)
	}

	return node
}
