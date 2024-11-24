package alias

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
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
	fields []fields.Field
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

func (n *aliasNode) Fields() []fields.Field {
	return slices.Clone(n.fields)
}

func (n *aliasNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("alias as %s", n.name)
	n.Node.Serialize(w.Indent())
}

func (n *aliasNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.Node.AddFilter(ctx, n.deprojectExpression(filter))
}

func (n *aliasNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.deprojectExpression(expression), nil
	})

	n.Node.AddOrder(ctx, mapped)
}

func (n *aliasNode) Optimize(ctx impls.OptimizationContext) {
	n.Node.Optimize(ctx)
}

func (n *aliasNode) Filter() impls.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	return n.projectExpression(filter)
}

func (n *aliasNode) Ordering() impls.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projectExpression(expression), nil
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

//
//

func (n *aliasNode) projectExpression(expression impls.Expression) impls.Expression {
	for _, field := range expressions.Fields(expression) {
		expression = projection.Alias(expression, field, expressions.NewNamed(field.WithRelationName(n.name)))
	}

	return expression
}

func (n *aliasNode) deprojectExpression(expression impls.Expression) impls.Expression {
	for _, field := range n.fields {
		expression = projection.Dealias(expression, field, expressions.NewNamed(field.WithRelationName(n.Node.Name())))
	}

	return expression
}
