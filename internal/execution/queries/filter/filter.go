package filter

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
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

func (n *logicalFilterNode) Build() queries.Node {
	if n.filter == nil {
		return n.LogicalNode.Build()
	}

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
	w.WritefLine("filter by %s", n.filter)
	n.Node.Serialize(w.Indent())
}

func (n *filterNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Filter Node scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Filter")

		for {
			row, err := scanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			if ok, err := types.ValueAs[bool](queries.Evaluate(ctx, n.filter, row)); err != nil {
				return rows.Row{}, err
			} else if ok == nil || !*ok {
				continue
			}

			return row, nil
		}
	}), nil
}
