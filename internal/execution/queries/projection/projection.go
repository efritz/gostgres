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
	w.WritefLine("project %s", n.projection)
	n.Node.Serialize(w.Indent())
}

func (n *projectionNode) AddFilter(ctx impls.OptimizationContext, filter impls.Expression) {
	n.Node.AddFilter(ctx, n.deprojectExpression(filter))
}

func (n *projectionNode) AddOrder(ctx impls.OptimizationContext, order impls.OrderExpression) {
	mapped, _ := order.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.deprojectExpression(expression), nil
	})

	n.Node.AddOrder(ctx, mapped)
}

func (n *projectionNode) Optimize(ctx impls.OptimizationContext) {
	n.projection.Optimize(ctx)
	n.Node.Optimize(ctx)
}

func (n *projectionNode) Filter() impls.Expression {
	filter := n.Node.Filter()
	if filter == nil {
		return nil
	}

	return n.projectExpression(filter)
}

func (n *projectionNode) Ordering() impls.OrderExpression {
	ordering := n.Node.Ordering()
	if ordering == nil {
		return nil
	}

	mapped, _ := ordering.Map(func(expression impls.Expression) (impls.Expression, error) {
		return n.projectExpression(expression), nil
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

	aliases := n.projection.Aliases()

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Projection")

		row, err := scanner.Scan()
		if err != nil {
			return rows.Row{}, err
		}

		values := make([]any, 0, len(aliases))
		for _, field := range aliases {
			value, err := queries.Evaluate(ctx, field.Expression, row)
			if err != nil {
				return rows.Row{}, err
			}

			values = append(values, value)
		}

		return rows.NewRow(n.projection.Fields(), values)
	}), nil
}

//
//

func (p *projectionNode) projectExpression(expression impls.Expression) impls.Expression {
	aliases := p.projection.Aliases()
	projectionFields := p.projection.Fields()

	for i, alias := range aliases {
		if named, ok := alias.Expression.(expressions.NamedExpression); ok {
			expression = projection.Alias(expression, named.Field(), expressions.NewNamed(projectionFields[i]))
		}
	}

	return expression
}

func (p *projectionNode) deprojectExpression(expression impls.Expression) impls.Expression {
	aliases := p.projection.Aliases()
	projectionFields := p.projection.Fields()

	for i, alias := range aliases {
		field := fields.NewField(projectionFields[i].RelationName(), alias.Alias, types.TypeAny, fields.NonInternalField)
		expression = projection.Dealias(expression, field, alias.Expression)
	}

	return expression
}
