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

func (n *projectionNode) Ordering() OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	return mapOrderExpressions(ordering, func(expression expressions.Expression) expressions.Expression {
		// TODO - move this into projector
		for i, alias := range n.projector.aliases {
			if field, ok := alias.expression.Named(); ok {
				expression = expression.Alias(field, expressions.NewNamed(n.projector.projectedFields[i]))
			}
		}
		return expression
	})
}

func (n *projectionNode) Scan(visitor VisitorFunc) error {
	return n.Node.Scan(n.decorateVisitor(visitor))
}

func (n *projectionNode) decorateVisitor(visitor VisitorFunc) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		row, err := n.projector.projectRow(row)
		if err != nil {
			return false, err
		}

		return visitor(row)
	}
}
