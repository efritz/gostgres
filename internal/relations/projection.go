package relations

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type projectionRelation struct {
	Relation
	aliases []aliasProjection
	fields  []shared.Field
}

var _ Relation = &projectionRelation{}

func NewProjection(relation Relation, expressions []ProjectionExpression) (Relation, error) {
	var aliases []aliasProjection
	for _, expression := range expressions {
		as, err := expression.Expand(relation.Fields())
		if err != nil {
			return nil, err
		}

		aliases = append(aliases, as...)
	}

	fields := make([]shared.Field, 0, len(aliases))
	for _, field := range aliases {
		fields = append(fields, shared.Field{
			RelationName: relation.Name(),
			Name:         field.alias,
		})
	}

	return &projectionRelation{
		Relation: relation,
		aliases:  aliases,
		fields:   fields,
	}, nil
}

func (r *projectionRelation) Fields() []shared.Field {
	return copyFields(r.fields)
}

type named interface {
	Name() string
}

func (r *projectionRelation) Serialize(w io.Writer, indentationLevel int) {
	fields := make([]string, 0, len(r.aliases))
	for _, expression := range r.aliases {
		if named, ok := expression.expression.(named); ok && named.Name() == expression.alias {
			fields = append(fields, fmt.Sprintf("%s", expression.expression))
		} else {
			fields = append(fields, fmt.Sprintf("%s as %s", expression.expression, expression.alias))
		}
	}

	io.WriteString(w, fmt.Sprintf("%sselect (%s)\n", indent(indentationLevel), strings.Join(fields, ", ")))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *projectionRelation) Optimize() {
	for i := range r.aliases {
		r.aliases[i].expression = r.aliases[i].expression.Fold()
	}

	r.Relation.Optimize()
}

func (r *projectionRelation) PushDownFilter(filter expressions.Expression) bool {
	for _, expression := range r.aliases {
		filter = filter.Alias(shared.Field{Name: expression.alias}, expression.expression)
	}

	return r.Relation.PushDownFilter(filter)
}

func (r *projectionRelation) Scan(visitor VisitorFunc) error {
	return r.Relation.Scan(r.decorateVisitor(visitor))
}

func (r *projectionRelation) decorateVisitor(visitor VisitorFunc) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		row, err := r.project(row)
		if err != nil {
			return false, err
		}

		return visitor(row)
	}
}

func (r *projectionRelation) project(row shared.Row) (shared.Row, error) {
	values := make([]interface{}, 0, len(r.aliases))
	for _, field := range r.aliases {
		value, err := field.expression.ValueFrom(row)
		if err != nil {
			return shared.Row{}, err
		}

		values = append(values, value)
	}

	return shared.NewRow(r.fields, values), nil
}

type ProjectionExpression interface {
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

func (p aliasProjection) Expand(fields []shared.Field) ([]aliasProjection, error) {
	return []aliasProjection{p}, nil
}

type wildcardProjection struct {
	relationName string
}

func NewWildcardProjectionExpression(relationName string) ProjectionExpression {
	return wildcardProjection{
		relationName: relationName,
	}
}

func (p wildcardProjection) Expand(fields []shared.Field) (projections []aliasProjection, _ error) {
	matched := false
	for _, field := range fields {
		if field.RelationName != p.relationName {
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
