package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
)

type namedExpression struct {
	field shared.Field
}

var _ Expression = &namedExpression{}

func NewNamed(field shared.Field) Expression {
	return namedExpression{
		field: field,
	}
}

func (e namedExpression) String() string {
	return e.field.String()
}

func (e namedExpression) Equal(other Expression) bool {
	if o, ok := other.(namedExpression); ok {
		return e.field.Name() == o.field.Name() && (e.field.RelationName() == o.field.RelationName() || e.field.RelationName() == "" || o.field.RelationName() == "")
	}

	return false
}

func (e namedExpression) Name() string {
	return e.field.Name()
}

func (e namedExpression) Fields() []shared.Field {
	return []shared.Field{e.field}
}

func (e namedExpression) Named() (shared.Field, bool) {
	return e.field, true
}

func (e namedExpression) Fold() Expression {
	return e
}

func (e namedExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e namedExpression) Map(f func(Expression) Expression) Expression {
	return f(e)
}

func (e namedExpression) ValueFrom(context ExpressionContext, row shared.Row) (any, error) {
	index, err := shared.FindMatchingFieldIndex(e.field, row.Fields)
	if err != nil {
		return nil, err
	}

	return row.Values[index], nil
}
