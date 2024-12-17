package mutation

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/execution/queries/nodes/mutation"
	"github.com/efritz/gostgres/internal/execution/queries/plan"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
)

type logicalInsertNode struct {
	plan.LogicalNode
	table       impls.Table
	columnNames []string
	returning   *projection.Projection
}

func NewInsert(
	node plan.LogicalNode,
	table impls.Table,
	columnNames []string,
	returning *projection.Projection,
) (plan.LogicalNode, error) {
	return &logicalInsertNode{
		LogicalNode: node,
		table:       table,
		columnNames: columnNames,
		returning:   returning,
	}, nil
}

func (n *logicalInsertNode) Fields() []fields.Field                                              { return n.returning.Fields() }
func (n *logicalInsertNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression)    {}
func (n *logicalInsertNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {}
func (n *logicalInsertNode) Filter() impls.Expression                                            { return nil }
func (n *logicalInsertNode) Ordering() impls.OrderExpression                                     { return nil }
func (n *logicalInsertNode) SupportsMarkRestore() bool                                           { return false }

func (n *logicalInsertNode) Opitmize(ctx impls.OptimizationContext) {
	n.returning.Optimize(ctx)
	n.LogicalNode.Optimize(ctx)
}

func (n *logicalInsertNode) EstimateCost() plan.Cost {
	return plan.Cost{} // TODO
}

func (n *logicalInsertNode) Build() nodes.Node {
	node := n.LogicalNode.Build()
	node = mutation.NewInsert(node, n.table, n.columnNames)
	node = nodes.NewProjection(node, n.returning)
	return node
}
