package relations

import (
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
	return r.fields
}

func (r *projectionRelation) Scan(scanContext ScanContext, visitor VisitorFunc) error {
	return r.Relation.Scan(scanContext, r.decorateVisitor(scanContext, visitor))
}

func (r *projectionRelation) decorateVisitor(scanContext ScanContext, visitor VisitorFunc) VisitorFunc {
	fields := r.Relation.Fields()

	return func(scanContext ScanContext, values []interface{}) (bool, error) {
		rowValues := make([]interface{}, 0, len(r.expressions))
		for _, field := range r.expressions {
			value, err := field.Expression.ValueFrom(shared.Row{Fields: fields, Values: values})
			if err != nil {
				return false, err
			}

			rowValues = append(rowValues, value)
		}

		return visitor(scanContext, rowValues)
	}
}
