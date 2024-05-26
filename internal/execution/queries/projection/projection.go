package projection

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type projectionNode struct {
	queries.Node
	projector *Projector
}

var _ queries.Node = &projectionNode{}

func NewProjection(node queries.Node, expressions []ProjectionExpression) (queries.Node, error) {
	projector, err := NewProjector(node.Name(), node.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &projectionNode{
		Node:      node,
		projector: projector,
	}, nil
}

func (n *projectionNode) Fields() []shared.Field {
	return slices.Clone(n.projector.projectedFields)
}

func (n *projectionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("select (%s)", n.projector)
	n.Node.Serialize(w.Indent())
}

func (n *projectionNode) AddFilter(filter types.Expression) {
	n.Node.AddFilter(n.projector.projectExpression(filter))
}

func (n *projectionNode) AddOrder(order expressions.OrderExpression) {
	n.Node.AddOrder(order.Map(func(expression types.Expression) types.Expression {
		return n.projector.projectExpression(expression)
	}))
}

func (n *projectionNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *projectionNode) Filter() types.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	return n.projector.deprojectExpression(filter)
}

func (n *projectionNode) Ordering() expressions.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	return ordering.Map(func(expression types.Expression) types.Expression {
		return n.projector.deprojectExpression(expression)
	})
}

func (n *projectionNode) SupportsMarkRestore() bool {
	return false
}

func (n *projectionNode) Scanner(ctx types.Context) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		row, err := scanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		return n.projector.ProjectRow(ctx, row)
	}), nil
}
