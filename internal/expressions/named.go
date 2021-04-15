package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type namedExpression struct {
	field shared.Field
}

var _ Expression = &namedExpression{}

func NewNamed(field shared.Field) Expression {
	return &namedExpression{
		field: field,
	}
}

func (e namedExpression) String() string {
	if e.field.RelationName != "" {
		return fmt.Sprintf("%s.%s", e.field.RelationName, e.field.Name)
	}

	return e.field.Name
}

func (e namedExpression) Name() string {
	return e.field.Name
}

func (e namedExpression) Fields() []shared.Field {
	return []shared.Field{e.field}
}

func (e namedExpression) Fold() Expression {
	return e
}

func (e namedExpression) Alias(from, to string) Expression {
	if e.field.RelationName == from {
		return namedExpression{
			field: shared.Field{
				RelationName: to,
				Name:         e.field.Name,
			},
		}
	}

	return e
}

func (e namedExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e namedExpression) ValueFrom(row shared.Row) (interface{}, error) {
	index, err := shared.FindMatchingFieldIndex(e.field, row.Fields)
	if err != nil {
		return nil, err
	}

	return row.Values[index], nil
}
