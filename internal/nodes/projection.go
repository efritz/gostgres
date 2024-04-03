package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type projectionNode struct {
	Node
	projector *projector
}

var _ Node = &projectionNode{}

func NewProjection(node Node, expressions []ProjectionExpression) (Node, error) {
	projector, err := newProjector(node.Name(), node.Fields(), expressions)
	if err != nil {
		return nil, err
	}

	return &projectionNode{
		Node:      node,
		projector: projector,
	}, nil
}

func (n *projectionNode) Fields() []shared.Field {
	return copyFields(n.projector.projectedFields)
}

func (n *projectionNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sselect (%s)\n", indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *projectionNode) Optimize() {
	n.projector.optimize()
	n.Node.Optimize()
}

func (n *projectionNode) AddFilter(filter expressions.Expression) {
	n.Node.AddFilter(n.projector.projectExpression(filter))
}

func (n *projectionNode) AddOrder(order OrderExpression) {
	n.Node.AddOrder(mapOrderExpressions(order, func(expression expressions.Expression) expressions.Expression {
		return n.projector.projectExpression(expression)
	}))
}

func (n *projectionNode) Filter() expressions.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	return n.projector.deprojectExtension(filter)
}

func (n *projectionNode) Ordering() OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	return mapOrderExpressions(ordering, func(expression expressions.Expression) expressions.Expression {
		return n.projector.deprojectExtension(expression)
	})
}

func (n *projectionNode) Scanner() (Scanner, error) {
	scanner, err := n.Node.Scanner()
	if err != nil {
		return nil, err
	}

	return ScannerFunc(func() (shared.Row, error) {
		row, err := scanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		return n.projector.projectRow(row)
	}), nil
}
