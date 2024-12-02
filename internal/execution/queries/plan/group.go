package plan

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalGroupNode struct {
	LogicalNode
	groupExpressions []impls.Expression
	projection       *projection.Projection
}

func NewHashAggregate(
	node LogicalNode,
	groupExpressions []impls.Expression,
	projection *projection.Projection,
) LogicalNode {
	return &logicalGroupNode{
		LogicalNode:      node,
		groupExpressions: groupExpressions,
		projection:       projection,
	}
}

func (n *logicalGroupNode) Name() string {
	return ""
}

func (n *logicalGroupNode) Fields() []fields.Field {
	return n.projection.Fields()
}

func (n *logicalGroupNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	for _, expr := range expressions.Conjunctions(filter) {
		expr := n.projection.DeprojectExpression(expr)

		if _, _, containsAggregate, _ := expressions.PartitionAggregatedFieldReferences(ctx, []impls.Expression{expr}, nil); !containsAggregate {
			n.LogicalNode.AddFilter(ctx, expr)
		}
	}
}

func (n *logicalGroupNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	// No-op
}

func (n *logicalGroupNode) Optimize(ctx impls.OptimizationContext) {
	n.projection.Optimize(ctx)
	n.LogicalNode.Optimize(ctx)
}

func (n *logicalGroupNode) Filter() impls.Expression {
	return n.projection.ProjectExpression(n.LogicalNode.Filter())
}

func (n *logicalGroupNode) Ordering() impls.OrderExpression {
	return nil
}

func (n *logicalGroupNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalGroupNode) Build() nodes.Node {
	return nodes.NewGroup(n.LogicalNode.Build(), n.groupExpressions, n.projection)
}
