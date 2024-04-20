package alias

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type aliasNode struct {
	queries.Node
	name   string
	fields []shared.Field
}

var _ queries.Node = &aliasNode{}

func NewAlias(node queries.Node, name string) queries.Node {
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

func (n *aliasNode) AddOrder(order expressions.OrderExpression) {
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

func (n *aliasNode) Ordering() expressions.OrderExpression {
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

func (n *aliasNode) SupportsMarkRestore() bool {
	return false
}

func (n *aliasNode) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		row, err := scanner.Scan()
		if err != nil {
			return shared.Row{}, err
		}

		return shared.NewRow(n.fields, row.Values)
	}), nil
}

func namedFromField(field shared.Field, relationName string) expressions.Expression {
	return expressions.NewNamed(shared.NewField(relationName, field.Name, field.TypeKind, field.Internal))
}

// TODO - deduplicate

func copyFields(fields []shared.Field) []shared.Field {
	c := make([]shared.Field, len(fields))
	copy(c, fields)
	return c
}

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

func updateRelationName(fields []shared.Field, relationName string) []shared.Field {
	fields = copyFields(fields)
	for i := range fields {
		fields[i].RelationName = relationName
	}

	return fields
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}
