package plan

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalDeleteNode struct {
	LogicalNode
	table     impls.Table
	aliasName string
	filter    impls.Expression
	returning *projection.Projection
}

func NewDelete(
	node LogicalNode,
	table impls.Table,
	aliasName string,
	filter impls.Expression,
	returning *projection.Projection,
) (LogicalNode, error) {
	return &logicalDeleteNode{
		LogicalNode: node,
		table:       table,
		aliasName:   aliasName,
		filter:      filter,
		returning:   returning,
	}, nil
}

func (n *logicalDeleteNode) Fields() []fields.Field                                              { return n.returning.Fields() }
func (n *logicalDeleteNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalDeleteNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalDeleteNode) Filter() impls.Expression                                            { return nil }
func (n *logicalDeleteNode) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalDeleteNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalDeleteNode) Optimize(ctx impls.OptimizationContext) {
	n.returning.Optimize(ctx)

	if n.filter != nil {
		n.filter = n.filter.Fold()
		n.LogicalNode.AddFilter(ctx, n.filter)
	}

	n.LogicalNode.Optimize(ctx)

	n.filter = expressions.FilterDifference(n.filter, n.LogicalNode.Filter())
}

func (n *logicalDeleteNode) Build() nodes.Node {
	node := n.LogicalNode.Build()
	if n.filter != nil {
		node = nodes.NewFilter(node, n.filter)
	}

	node = nodes.NewDelete(node, n.table, n.aliasName)
	node = nodes.NewProjection(node, n.returning)
	return node
}
