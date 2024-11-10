package alias

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projector"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type aliasNode struct {
	queries.Node
	name   string
	fields []fields.ResolvedField
}

var _ queries.Node = &aliasNode{}

func NewAlias(node queries.Node, name string) queries.Node {
	fields := slices.Clone(node.Fields())
	for i := range fields {
		fields[i] = fields[i] // .WithRelationName(name)
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

func (n *aliasNode) Fields() []fields.ResolvedField {
	return slices.Clone(n.fields)
}

func (n *aliasNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("alias as %s", n.name)
	n.Node.Serialize(w.Indent())
}

func (n *aliasNode) AddFilter(filter impls.Expression) {
	for _, field := range n.fields {
		filter = projector.Alias(filter, field, namedFromField(field, n.Node.Name()))
	}

	n.Node.AddFilter(filter)
}

func (n *aliasNode) AddOrder(order impls.OrderExpression) {
	mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
		for _, field := range n.fields {
			expression = projector.Alias(expression, field, namedFromField(field, n.Node.Name()))
		}

		return expression, nil
	})

	n.Node.AddOrder(mapped)
}

func (n *aliasNode) Optimize() {
	n.Node.Optimize()
}

func (n *aliasNode) Filter() impls.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	for _, field := range expressions.Fields(filter) {
		filter = projector.Alias(filter, field, namedFromField(field, n.name))
	}

	return filter
}

func (n *aliasNode) Ordering() impls.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
		for _, field := range expressions.Fields(expression) {
			expression = projector.Alias(expression, field, namedFromField(field, n.name))
		}

		return expression, nil
	})

	return mapped
}

func (n *aliasNode) SupportsMarkRestore() bool {
	return false
}

func (n *aliasNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Alias scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Alias")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		return rows.NewRow(n.fields, row.Values)
	}), nil
}

func namedFromField(field fields.Field, relationName string) impls.Expression {
	return expressions.NewNamed(field.WithRelationName(relationName))
}
