package nodes

import (
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalProjectionNode struct {
	LogicalNode
	projection *projection.Projection
}

var _ LogicalNode = &logicalProjectionNode{}

func NewProjection(node LogicalNode, projection *projection.Projection) LogicalNode {
	return &logicalProjectionNode{
		LogicalNode: node,
		projection:  projection,
	}
}

func (n *logicalProjectionNode) Fields() []fields.Field {
	return n.projection.Fields()
}

func (n *logicalProjectionNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.LogicalNode.AddFilter(ctx, n.projection.DeprojectExpression(filter))
}

func (n *logicalProjectionNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projection.DeprojectExpression(expression), nil
	})

	n.LogicalNode.AddOrder(ctx, mapped)
}

func (n *logicalProjectionNode) Optimize(ctx impls.OptimizationContext) {
	n.projection.Optimize(ctx)
	n.LogicalNode.Optimize(ctx)
}

func (n *logicalProjectionNode) Filter() impls.Expression {
	return n.projection.ProjectExpression(n.LogicalNode.Filter())
}

func (n *logicalProjectionNode) Ordering() impls.OrderExpression {
	ordering := n.LogicalNode.Ordering()
	if ordering == nil {
		return nil
	}

	mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projection.ProjectExpression(expression), nil
	})

	return mapped
}

func (n *logicalProjectionNode) SupportsMarkRestore() bool {
	return false
}

func (n *logicalProjectionNode) Build() Node {
	return &projectionNode{
		Node:       n.LogicalNode.Build(),
		projection: n.projection,
	}
}

//
//

type projectionNode struct {
	Node
	projection *projection.Projection
}

var _ Node = &projectionNode{}

func (n *projectionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("project %s", n.projection)
	n.Node.Serialize(w.Indent())
}

func (n *projectionNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Projection scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	aliases := n.projection.Aliases()

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Projection")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		values := make([]any, 0, len(aliases))
		for _, field := range aliases {
			value, err := queries.Evaluate(ctx, field.Expression, row)
			if err != nil {
				return rows.Row{}, err
			}

			values = append(values, value)
		}

		return rows.NewRow(n.projection.Fields(), values)
	}), nil
}
