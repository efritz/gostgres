package alias

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/queries/projection"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type aliasNode struct {
	queries.Node
	name   string
	fields []shared.Field
}

var _ queries.Node = &aliasNode{}

func NewAlias(node queries.Node, name string) queries.Node {
	fields := slices.Clone(node.Fields())
	for i := range fields {
		fields[i] = fields[i].WithRelationName(name)
	}

	return &aliasNode{
		Node:   node,
		name:   name,
		fields: fields,
	}
}

func (n *aliasNode) Name() string {
	return n.name
}

func (n *aliasNode) Fields() []shared.Field {
	return slices.Clone(n.fields)
}

func (n *aliasNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("alias as %s", n.name)
	n.Node.Serialize(w.Indent())
}

func (n *aliasNode) AddFilter(filter expressions.Expression) {
	for _, field := range n.fields {
		filter = projection.Alias(filter, field, namedFromField(field, n.Node.Name()))
	}

	n.Node.AddFilter(filter)
}

func (n *aliasNode) AddOrder(order expressions.OrderExpression) {
	n.Node.AddOrder(order.Map(func(expression expressions.Expression) expressions.Expression {
		for _, field := range n.fields {
			expression = projection.Alias(expression, field, namedFromField(field, n.Node.Name()))
		}

		return expression
	}))
}

func (n *aliasNode) Optimize() {
	n.Node.Optimize()
}

func (n *aliasNode) Filter() expressions.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	for _, field := range expressions.Fields(filter) {
		filter = projection.Alias(filter, field, namedFromField(field, n.name))
	}

	return filter
}

func (n *aliasNode) Ordering() expressions.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	return ordering.Map(func(expression expressions.Expression) expressions.Expression {
		for _, field := range expressions.Fields(expression) {
			expression = projection.Alias(expression, field, namedFromField(field, n.name))
		}

		return expression
	})
}

func (n *aliasNode) SupportsMarkRestore() bool {
	return false
}

func (n *aliasNode) Scanner(ctx execution.Context) (scan.Scanner, error) {
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
	return expressions.NewNamed(field.WithRelationName(relationName))
}
