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
		filter = filter.Alias(field, expressions.NewNamed(shared.NewField(n.Node.Name(), field.Name, field.TypeKind, field.Internal)))
	}

	n.Node.AddFilter(filter)
}

func (n *aliasNode) AddOrder(order OrderExpression) {
	n.Node.AddOrder(mapOrderExpressions(order, func(expression expressions.Expression) expressions.Expression {
		for _, field := range n.fields {
			expression = expression.Alias(field, expressions.NewNamed(shared.NewField(n.Node.Name(), field.Name, field.TypeKind, field.Internal)))
		}

		return expression
	}))
}

func (n *aliasNode) Ordering() OrderExpression {
	panic("alias.Ordering unimplemented")
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
