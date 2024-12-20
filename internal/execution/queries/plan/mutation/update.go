package mutation

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/mutation"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalUpdateNode struct {
	plan.LogicalNode
	table          impls.Table
	aliasName      string
	setExpressions []mutation.SetExpression
	filter         impls.Expression
	returning      *projection.Projection
}

func NewUpdate(
	node plan.LogicalNode,
	table impls.Table,
	aliasName string,
	setExpressions []mutation.SetExpression,
	filter impls.Expression,
	returning *projection.Projection,
) (plan.LogicalNode, error) {
	return &logicalUpdateNode{
		LogicalNode:    node,
		table:          table,
		aliasName:      aliasName,
		setExpressions: setExpressions,
		filter:         filter,
		returning:      returning,
	}, nil
}

func (n *logicalUpdateNode) Fields() []fields.Field                                              { return n.returning.Fields() }
func (n *logicalUpdateNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalUpdateNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalUpdateNode) Filter() impls.Expression                                            { return nil }
func (n *logicalUpdateNode) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalUpdateNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalUpdateNode) Optimize(ctx impls.OptimizationContext) {
	n.returning.Optimize(ctx)

	if n.filter != nil {
		n.filter = n.filter.Fold()
		n.LogicalNode.AddFilter(ctx, n.filter)
	}

	n.LogicalNode.Optimize(ctx)

	n.filter = expressions.FilterDifference(n.filter, n.LogicalNode.Filter())
}

func (n *logicalUpdateNode) EstimateCost() plan.Cost {
	return plan.Cost{} // TODO
}

func (n *logicalUpdateNode) Build() nodes.Node {
	node := n.LogicalNode.Build()
	if n.filter != nil {
		node = nodes.NewFilter(node, n.filter)
	}

	return mutation.NewUpdate(node, n.table, n.aliasName, n.setExpressions, n.returning)
}
