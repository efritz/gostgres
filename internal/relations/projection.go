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
	expressions []AliasedExpression
	fields      []shared.Field
}

var _ Relation = &projectionRelation{}

type AliasedExpression struct {
	Alias string
	expressions.Expression
}

func NewProjection(relation Relation, expressions []AliasedExpression) Relation {
	fields := make([]shared.Field, 0, len(expressions))
	for _, field := range expressions {
		fields = append(fields, shared.Field{
			RelationName: relation.Name(),
			Name:         field.Alias,
		})
	}

	return &projectionRelation{
		Relation:    relation,
		expressions: expressions,
		fields:      fields,
	}
}

func (r *projectionRelation) Fields() []shared.Field {
	return copyFields(r.fields)
}

type named interface {
	Name() string
}

func (r *projectionRelation) Serialize(w io.Writer, indentationLevel int) {
	fields := make([]string, 0, len(r.expressions))
	for _, expression := range r.expressions {
		if named, ok := expression.Expression.(named); ok && named.Name() == expression.Alias {
			fields = append(fields, fmt.Sprintf("%s", expression.Expression))
		} else {
			fields = append(fields, fmt.Sprintf("%s as %s", expression.Expression, expression.Alias))
		}
	}

	io.WriteString(w, fmt.Sprintf("%sselect (%s)\n", indent(indentationLevel), strings.Join(fields, ", ")))
	r.Relation.Serialize(w, indentationLevel+1)
}

func (r *projectionRelation) Optimize() {
	for i := range r.expressions {
		r.expressions[i].Expression = r.expressions[i].Expression.Fold()
	}

	r.Relation.Optimize()
}

func (r *projectionRelation) PushDownFilter(filter expressions.Expression) bool {
	for _, expression := range r.expressions {
		filter = filter.Alias(shared.Field{Name: expression.Alias}, expression.Expression)
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
	values := make([]interface{}, 0, len(r.expressions))
	for _, field := range r.expressions {
		value, err := field.Expression.ValueFrom(row)
		if err != nil {
			return shared.Row{}, err
		}

		values = append(values, value)
	}

	return shared.NewRow(r.fields, values), nil
}
