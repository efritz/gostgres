package projector

import (
	"slices"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type Projector struct {
	aliases         []ProjectedExpression
	fields          []fields.Field
	projectedFields []fields.Field
}

func NewProjector(name string, fields []fields.Field, expressions []ProjectionExpression) (*Projector, error) {
	aliases, err := expandProjection(fields, expressions)
	if err != nil {
		return nil, err
	}

	return &Projector{
		aliases:         aliases,
		fields:          fields,
		projectedFields: fieldsFromProjection(name, aliases),
	}, nil
}

func (p *Projector) Fields() []fields.Field {
	return slices.Clone(p.projectedFields)
}

func (p *Projector) String() string {
	type named interface {
		Name() string
	}

	fields := make([]string, 0, len(p.aliases))
	for _, expression := range p.aliases {
		if named, ok := expression.expression.(named); ok && named.Name() == expression.alias {
			fields = append(fields, expression.alias)
		} else {
			fields = append(fields, expression.String())
		}
	}

	return strings.Join(fields, ", ")
}

func (p *Projector) Optimize() {
	for i := range p.aliases {
		p.aliases[i].expression = p.aliases[i].expression.Fold()
	}
}

func (p *Projector) ProjectRow(ctx impls.ExecutionContext, row rows.Row) (rows.Row, error) {
	values := make([]any, 0, len(p.aliases))
	for _, field := range p.aliases {
		value, err := queries.Evaluate(ctx, field.expression, row)
		if err != nil {
			return rows.Row{}, err
		}

		values = append(values, value)
	}

	return rows.NewRow(p.projectedFields, values)
}

func (p *Projector) ProjectExpression(expression impls.Expression) impls.Expression {
	for _, alias := range p.aliases {
		expression = Alias(expression, fields.NewField("", alias.alias, types.TypeAny, fields.NonInternalField), alias.expression)
	}

	return expression
}

func (p *Projector) DeprojectExpression(expression impls.Expression) impls.Expression {
	for i, alias := range p.aliases {
		if named, ok := alias.expression.(expressions.NamedExpression); ok {
			expression = Alias(expression, named.Field(), expressions.NewNamed(p.projectedFields[i]))
		}
	}

	return expression

}

func fieldsFromProjection(relationName string, aliases []ProjectedExpression) []fields.Field {
	projectedFields := make([]fields.Field, 0, len(aliases))
	for _, field := range aliases {
		projectedFields = append(projectedFields, fields.NewField(relationName, field.alias, types.TypeAny, fields.NonInternalField))
	}

	return projectedFields
}
