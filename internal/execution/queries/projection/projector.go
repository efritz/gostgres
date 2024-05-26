package projection

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type Projector struct {
	aliases         []aliasProjection
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
	return p.projectedFields
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

func (p *Projector) ProjectRow(ctx impls.Context, row rows.Row) (rows.Row, error) {
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

func (p *Projector) projectExpression(expression impls.Expression) impls.Expression {
	for _, alias := range p.aliases {
		expression = Alias(expression, fields.NewField("", alias.alias, types.TypeAny), alias.expression)
	}

	return expression
}

func (p *Projector) deprojectExpression(expression impls.Expression) impls.Expression {
	for i, alias := range p.aliases {
		if named, ok := alias.expression.(expressions.NamedExpression); ok {
			expression = Alias(expression, named.Field(), expressions.NewNamed(p.projectedFields[i]))
		}
	}

	return expression

}

func expandProjection(fields []fields.Field, expressions []ProjectionExpression) ([]aliasProjection, error) {
	aliases := make([]aliasProjection, 0, len(fields))
	for _, expression := range expressions {
		as, err := expression.Expand(fields)
		if err != nil {
			return nil, err
		}

		aliases = append(aliases, as...)
	}

	return aliases, nil
}

func fieldsFromProjection(relationName string, aliases []aliasProjection) []fields.Field {
	projectedFields := make([]fields.Field, 0, len(aliases))
	for _, field := range aliases {
		projectedFields = append(projectedFields, fields.NewField(relationName, field.alias, types.TypeAny))
	}

	return projectedFields
}

type ProjectionExpression interface {
	Dealias(name string, fields []fields.Field, alias string) ProjectionExpression
	Expand(fields []fields.Field) ([]aliasProjection, error)
}

type aliasProjection struct {
	expression impls.Expression
	alias      string
}

func NewAliasProjectionExpression(expression impls.Expression, alias string) ProjectionExpression {
	return aliasProjection{
		expression: expression,
		alias:      alias,
	}
}

func UnwrapAlias(e ProjectionExpression) (impls.Expression, string, bool) {
	if alias, ok := e.(aliasProjection); ok {
		return alias.expression, alias.alias, true
	}

	return nil, "", false
}

func (p aliasProjection) String() string {
	return fmt.Sprintf("%s as %s", p.expression, p.alias)
}

func (p aliasProjection) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	expression := p.expression
	for _, field := range fields {
		expression = Alias(expression, field.WithRelationName(name), expressions.NewNamed(field))
	}

	return aliasProjection{
		expression: expression,
		alias:      p.alias,
	}
}

func (p aliasProjection) Expand(fields []fields.Field) ([]aliasProjection, error) {
	return []aliasProjection{p}, nil
}

type tableWildcardProjection struct {
	relationName string
}

func NewTableWildcardProjectionExpression(relationName string) ProjectionExpression {
	return tableWildcardProjection{
		relationName: relationName,
	}
}

func (p tableWildcardProjection) String() string {
	return fmt.Sprintf("%s.*", p.relationName)
}

func (p tableWildcardProjection) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	if p.relationName == alias {
		return tableWildcardProjection{
			relationName: name,
		}
	}

	return p
}

func (p tableWildcardProjection) Expand(fields []fields.Field) (projections []aliasProjection, _ error) {
	matched := false
	for _, field := range fields {
		if field.Internal() || field.RelationName() != p.relationName {
			continue
		}

		matched = true
		projections = append(projections, aliasProjection{
			alias:      field.Name(),
			expression: expressions.NewNamed(field),
		})
	}

	if !matched {
		return nil, fmt.Errorf("unknown relation %s", p.relationName)
	}

	return projections, nil
}

type wildcardProjection struct{}

func NewWildcardProjectionExpression() ProjectionExpression {
	return wildcardProjection{}
}

func (p wildcardProjection) String() string {
	return "*"
}

func (p wildcardProjection) Dealias(name string, fields []fields.Field, alias string) ProjectionExpression {
	return p
}

func (p wildcardProjection) Expand(fields []fields.Field) (projections []aliasProjection, _ error) {
	for _, field := range fields {
		if field.Internal() {
			continue
		}

		projections = append(projections, aliasProjection{
			alias:      field.Name(),
			expression: expressions.NewNamed(field),
		})
	}

	return projections, nil
}

func Alias(e impls.Expression, field fields.Field, target impls.Expression) impls.Expression {
	return e.Map(func(e impls.Expression) impls.Expression {
		if named, ok := e.(expressions.NamedExpression); ok {
			if field.RelationName() == "" || named.Field().RelationName() == field.RelationName() {
				if named.Field().Name() == field.Name() {
					return target
				}
			}
		}

		return e
	})
}
