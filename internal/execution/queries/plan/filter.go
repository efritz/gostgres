package plan

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
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
	node := n.LogicalNode.Build()
	if n.filter != nil {
		node = nodes.NewFilter(node, n.filter)
	}

	return node
}

//
//

func lowerFilter(ctx impls.OptimizationContext, filter impls.Expression, nodes ...LogicalNode) {
	for _, expression := range expressions.Conjunctions(filter) {
		for _, node := range nodes {
			lowerFilterToNode(ctx, expression, node)
		}
	}
}

func lowerFilterToNode(ctx impls.OptimizationContext, expression impls.Expression, node LogicalNode) {
	expressionFields := expressions.Fields(expression)
	nodeFields := node.Fields()
	outerFields := ctx.OuterFields()
	combinedNodeFields := append(slices.Clone(nodeFields), outerFields...)

	// Ensure that we don't push down filters that references unknown fields
	// Also skip filters that aren't _specific_ to the node and only reference outer fields
	if matchesAllFields(expressionFields, combinedNodeFields) && !matchesAllFields(expressionFields, outerFields) {
		node.AddFilter(ctx, expression)
	}
}

func matchesAllFields(haystack, candidates []fields.Field) bool {
	for _, f := range haystack {
		if _, err := fields.FindMatchingFieldIndex(f, candidates); err != nil {
			return false
		}
	}

	return true
}
