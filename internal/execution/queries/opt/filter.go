package opt

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalFilterNode struct {
	LogicalNode
	filter impls.Expression
}

func NewFilter(node LogicalNode, filter impls.Expression) LogicalNode {
	return &logicalFilterNode{
		LogicalNode: node,
		filter:      filter,
	}
}

func (n *logicalFilterNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filter)
}

func (n *logicalFilterNode) Optimize(ctx impls.OptimizationContext) {
	if n.filter != nil {
		n.filter = n.filter.Fold()
		n.LogicalNode.AddFilter(ctx, n.filter)
	}

	n.LogicalNode.Optimize(ctx)
	n.filter = expressions.FilterDifference(n.filter, n.LogicalNode.Filter())
}

func (n *logicalFilterNode) Filter() impls.Expression {
	return expressions.UnionFilters(n.filter, n.LogicalNode.Filter())
}

func (n *logicalFilterNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalFilterNode) Build() nodes.Node {
	if n.filter == nil {
		return n.LogicalNode.Build()
	}

	return nodes.NewFilter(n.LogicalNode.Build(), n.filter)
}
