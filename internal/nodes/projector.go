package nodes

import (
	"fmt"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type projector struct {
	aliases         []aliasProjection
	fields          []shared.Field
	projectedFields []shared.Field
}

func newProjector(name string, fields []shared.Field, expressions []ProjectionExpression) (*projector, error) {
	aliases, err := expandProjection(fields, expressions)
	if err != nil {
		return nil, err
	}

	return &projector{
		aliases:         aliases,
		fields:          fields,
		projectedFields: fieldsFromProjection(name, aliases),
	}, nil
}

func (p *projector) String() string {
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

func (p *projector) optimize() {
	for i := range p.aliases {
		p.aliases[i].expression = p.aliases[i].expression.Fold()
	}
}

func (p *projector) projectRow(row shared.Row) (shared.Row, error) {
	values := make([]interface{}, 0, len(p.aliases))
	for _, field := range p.aliases {
		value, err := field.expression.ValueFrom(row)
		if err != nil {
			return shared.Row{}, err
		}

		values = append(values, value)
	}

	return shared.NewRow(p.projectedFields, values)
}

func (p *projector) projectExpression(expression expressions.Expression) expressions.Expression {
	for _, alias := range p.aliases {
		expression = expression.Alias(shared.NewField("", alias.alias, shared.TypeKindAny, false), alias.expression)
	}

	return expression
}

func expandProjection(fields []shared.Field, expressions []ProjectionExpression) ([]aliasProjection, error) {
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

func fieldsFromProjection(relationName string, aliases []aliasProjection) []shared.Field {
	fields := make([]shared.Field, 0, len(aliases))
	for _, field := range aliases {
		fields = append(fields, shared.NewField(relationName, field.alias, shared.TypeKindAny, false))
	}

	return fields
}

type ProjectionExpression interface {
	Dealias(name string, fields []shared.Field, alias string) ProjectionExpression
	Expand(fields []shared.Field) ([]aliasProjection, error)
}

type aliasProjection struct {
	expression expressions.Expression
	alias      string
}

func NewAliasProjectionExpression(expression expressions.Expression, alias string) ProjectionExpression {
	return aliasProjection{
		expression: expression,
		alias:      alias,
	}
}

func (p aliasProjection) String() string {
	return fmt.Sprintf("%s as %s", p.expression, p.alias)
}

func (p aliasProjection) Dealias(name string, fields []shared.Field, alias string) ProjectionExpression {
	expression := p.expression
	for _, field := range fields {
		expression = expression.Alias(shared.NewField(alias, field.Name, field.TypeKind, field.Internal), expressions.NewNamed(field))
	}

	return aliasProjection{
		expression: expression,
		alias:      p.alias,
	}
}

func (p aliasProjection) Expand(fields []shared.Field) ([]aliasProjection, error) {
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

func (p tableWildcardProjection) Dealias(name string, fields []shared.Field, alias string) ProjectionExpression {
	if p.relationName == alias {
		return tableWildcardProjection{
			relationName: name,
		}
	}

	return p
}

func (p tableWildcardProjection) Expand(fields []shared.Field) (projections []aliasProjection, _ error) {
	matched := false
	for _, field := range fields {
		if field.Internal || field.RelationName != p.relationName {
			continue
		}

		matched = true
		projections = append(projections, aliasProjection{
			alias:      field.Name,
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

func (p wildcardProjection) Dealias(name string, fields []shared.Field, alias string) ProjectionExpression {
	return p
}

func (p wildcardProjection) Expand(fields []shared.Field) (projections []aliasProjection, _ error) {
	for _, field := range fields {
		if field.Internal {
			continue
		}

		projections = append(projections, aliasProjection{
			alias:      field.Name,
			expression: expressions.NewNamed(field),
		})
	}

	return projections, nil
}
