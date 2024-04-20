package projection

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
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
	return copyFields(n.projector.projectedFields)
}

func (n *projectionNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%sselect (%s)\n", indent(indentationLevel), n.projector))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *projectionNode) Optimize() {
	n.projector.Optimize()
	n.Node.Optimize()
}

func (n *projectionNode) AddFilter(filter expressions.Expression) {
	n.Node.AddFilter(n.projector.projectExpression(filter))
}

func (n *projectionNode) AddOrder(order expressions.OrderExpression) {
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

func (n *projectionNode) Ordering() expressions.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	return mapOrderExpressions(ordering, func(expression expressions.Expression) expressions.Expression {
		return n.projector.deprojectExtension(expression)
	})
}

func (n *projectionNode) SupportsMarkRestore() bool {
	return false
}

func (n *projectionNode) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
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

// TODO - deduplicate

func mapOrderExpressions(order expressions.OrderExpression, f func(expressions.Expression) expressions.Expression) expressions.OrderExpression {
	orderExpressions := order.Expressions()
	aliasedExpressions := make([]expressions.ExpressionWithDirection, 0, len(orderExpressions))

	for _, expression := range orderExpressions {
		aliasedExpressions = append(aliasedExpressions, expressions.ExpressionWithDirection{
			Expression: f(expression.Expression),
			Reverse:    expression.Reverse,
		})
	}

	return expressions.NewOrderExpression(aliasedExpressions)
}

func copyFields(fields []shared.Field) []shared.Field {
	c := make([]shared.Field, len(fields))
	copy(c, fields)
	return c
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}
