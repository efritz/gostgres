package filter

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalFilterNode struct {
	queries.LogicalNode
	filter impls.Expression
}

var _ queries.LogicalNode = &logicalFilterNode{}

func NewFilter(node queries.LogicalNode, filter impls.Expression) queries.LogicalNode {
	return &logicalFilterNode{
		LogicalNode: node,
		filter:      filter,
	}
}

func (n *logicalFilterNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.filter = expressions.UnionFilters(n.filter, filter)
}

func (n *logicalFilterNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	n.LogicalNode.AddOrder(ctx, order)
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

func (n *logicalFilterNode) Ordering() impls.OrderExpression {
	return n.LogicalNode.Ordering()
}

func (n *logicalFilterNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalFilterNode) Build() queries.Node {
	return &filterNode{
		Node:   n.LogicalNode.Build(),
		filter: n.filter,
	}
}

//
//

type filterNode struct {
	queries.Node
	filter impls.Expression
}

var _ queries.Node = &filterNode{}

func (n *filterNode) Serialize(w serialization.IndentWriter) {
	if n.filter == nil {
		n.Node.Serialize(w)
	} else {
		w.WritefLine("filter by %s", n.filter)
		n.Node.Serialize(w.Indent())
	}
}

func (n *filterNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Filter Node scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return NewFilterScanner(ctx, scanner, n.filter)
}
