package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type aliasNode struct {
	Node
	name   string
	fields []shared.Field
}

var _ Node = &aliasNode{}

func NewAlias(node Node, name string) Node {
	return &aliasNode{
		Node:   node,
		name:   name,
		fields: updateRelationName(node.Fields(), name),
	}
}

func (n *aliasNode) Name() string {
	return n.name
}

func (n *aliasNode) Fields() []shared.Field {
	return copyFields(n.fields)
}

func (n *aliasNode) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%salias as %s\n", indent(indentationLevel), n.name))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *aliasNode) Optimize() {
	n.Node.Optimize()
}

func (n *aliasNode) AddFilter(filter expressions.Expression) {
	for _, field := range n.fields {
		filter = filter.Alias(field, namedFromField(field, n.Node.Name()))
	}

	n.Node.AddFilter(filter)
}

func (n *aliasNode) AddOrder(order OrderExpression) {
	n.Node.AddOrder(mapOrderExpressions(order, func(expression expressions.Expression) expressions.Expression {
		for _, field := range n.fields {
			expression = expression.Alias(field, namedFromField(field, n.Node.Name()))
		}

		return expression
	}))
}

func (n *aliasNode) Filter() expressions.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	for _, field := range filter.Fields() {
		filter = filter.Alias(field, namedFromField(field, n.name))
	}

	return filter
}

func (n *aliasNode) Ordering() OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	return mapOrderExpressions(ordering, func(expression expressions.Expression) expressions.Expression {
		for _, field := range expression.Fields() {
			expression = expression.Alias(field, namedFromField(field, n.name))
		}

		return expression
	})
}

func (n *aliasNode) Scan(visitor VisitorFunc) error {
	return n.Node.Scan(func(row shared.Row) (bool, error) {
		aliasedRow, err := shared.NewRow(n.fields, row.Values)
		if err != nil {
			return false, err
		}

		return visitor(aliasedRow)
	})
}

func namedFromField(field shared.Field, relationName string) expressions.Expression {
	return expressions.NewNamed(shared.NewField(relationName, field.Name, field.TypeKind, field.Internal))
}
