package projection

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/projection"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
)

type projectionNode struct {
	queries.Node
	projection *projection.Projection
}

var _ queries.Node = &projectionNode{}

func NewProjection(node queries.Node, projection *projection.Projection) queries.Node {
	return &projectionNode{
		Node:       node,
		projection: projection,
	}
}

func (n *projectionNode) Fields() []fields.Field {
	return n.projection.Fields()
}

func (n *projectionNode) Serialize(w serialization.IndentWriter) {
	w.WritefLine("select (%s)", n.projection)
	n.Node.Serialize(w.Indent())
}

func (n *projectionNode) AddFilter(filter impls.Expression) {
	n.Node.AddFilter(n.projectExpression(filter))
}

func (n *projectionNode) AddOrder(order impls.OrderExpression) {
	mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projectExpression(expression), nil
	})

	n.Node.AddOrder(mapped)
}

func (n *projectionNode) Optimize() {
	n.projection.Optimize()
	n.Node.Optimize()
}

func (n *projectionNode) Filter() impls.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	return n.deprojectExpression(filter)
}

func (n *projectionNode) Ordering() impls.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.deprojectExpression(expression), nil
	})

	return mapped
}

func (n *projectionNode) SupportsMarkRestore() bool {
	return false
}

func (n *projectionNode) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Projection scanner")

	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Projection")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		return n.projectRow(ctx, row)
	}), nil
}

//
//

func (p *projectionNode) projectRow(ctx impls.ExecutionContext, row rows.Row) (rows.Row, error) {
	aliases := p.projection.Aliases()

	values := make([]any, 0, len(aliases))
	for _, field := range aliases {
		value, err := queries.Evaluate(ctx, field.Expression, row)
		if err != nil {
			return rows.Row{}, err
		}

		values = append(values, value)
	}

	return rows.NewRow(p.projection.Fields(), values)
}

func (p *projectionNode) projectExpression(expression impls.Expression) impls.Expression {
	aliases := p.projection.Aliases()

	for _, alias := range aliases {
		expression = projection.Alias(expression, fields.NewField("", alias.Alias, types.TypeAny, fields.NonInternalField), alias.Expression)
	}

	return expression
}

func (p *projectionNode) deprojectExpression(expression impls.Expression) impls.Expression {
	aliases := p.projection.Aliases()
	fields := p.projection.Fields()

	for i, alias := range aliases {
		if named, ok := alias.Expression.(expressions.NamedExpression); ok {
			expression = projection.Alias(expression, named.Field(), expressions.NewNamed(fields[i]))
		}
	}

	return expression
}
