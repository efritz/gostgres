package projection

import (
	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type projectionNode struct {
	queries.Node
	projector *projector.Projector
}

var _ queries.Node = &projectionNode{}

func NewProjection(node queries.Node, expressions []projector.ProjectionExpression) (queries.Node, error) {
	projector, err := projector.NewProjector(node.Name(), node.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &projectionNode{
		Node:      node,
		projector: projector,
	}, nil
}

func (n *projectionNode) Fields() []fields.Field {
	return n.projector.Fields()
}

func (n *projectionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("select (%s)", n.projector)
	n.Node.Serialize(w.Indent())
}

func (n *projectionNode) AddFilter(filter impls.Expression) {
	n.Node.AddFilter(n.projector.ProjectExpression(filter))
}

func (n *projectionNode) AddOrder(order impls.OrderExpression) {
	mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projector.ProjectExpression(expression), nil
	})

	n.Node.AddOrder(mapped)
}

func (n *projectionNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *projectionNode) Filter() impls.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	return n.projector.DeprojectExpression(filter)
}

func (n *projectionNode) Ordering() impls.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projector.DeprojectExpression(expression), nil
	})

	return mapped
}

func (n *projectionNode) SupportsMarkRestore() bool {
	return false
}

func (n *projectionNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Projection scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Projection")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		return n.projector.ProjectRow(ctx, row)
	}), nil
}
