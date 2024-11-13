package projector

import (
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type Projector struct {
	projection *Projection
}

func NewProjector(relationName string, aliases []ProjectedExpression) *Projector {
	return &Projector{
		projection: NewProjection(relationName, aliases),
	}
}

func (p *Projector) Fields() []fields.Field {
	return p.projection.Fields()
}

func (p *Projector) String() string {
	return p.projection.String()
}

func (p *Projector) Optimize() {
	p.projection.Optimize()
}

func (p *Projector) ProjectRow(ctx impls.ExecutionContext, row rows.Row) (rows.Row, error) {
	values := make([]any, 0, len(p.projection.aliases))
	for _, field := range p.projection.aliases {
		value, err := queries.Evaluate(ctx, field.Expression, row)
		if err != nil {
			return rows.Row{}, err
		}

		values = append(values, value)
	}

	return rows.NewRow(p.projection.projectedFields, values)
}

func (p *Projector) ProjectExpression(expression impls.Expression) impls.Expression {
	for _, alias := range p.projection.aliases {
		expression = Alias(expression, fields.NewField("", alias.Alias, types.TypeAny, fields.NonInternalField), alias.Expression)
	}

	return expression
}

func (p *Projector) DeprojectExpression(expression impls.Expression) impls.Expression {
	for i, alias := range p.projection.aliases {
		if named, ok := alias.Expression.(expressions.NamedExpression); ok {
			expression = Alias(expression, named.Field(), expressions.NewNamed(p.projection.projectedFields[i]))
		}
	}

	return expression
}
