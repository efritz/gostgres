package expressions

import (
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type namedExpression struct {
	field shared.Field
}

var _ types.Expression = namedExpression{}

func NewNamed(field shared.Field) types.Expression {
	return namedExpression{
		field: field,
	}
}

func (e namedExpression) String() string {
	return e.field.String()
}

func (e namedExpression) Equal(other types.Expression) bool {
	if o, ok := other.(namedExpression); ok {
		return e.field.Name() == o.field.Name() && (e.field.RelationName() == o.field.RelationName() || e.field.RelationName() == "" || o.field.RelationName() == "")
	}

	return false
}

func (e namedExpression) Name() string {
	return e.field.Name()
}

func (e namedExpression) Field() shared.Field {
	return e.field
}

func (e namedExpression) Fold() types.Expression {
	return e
}

func (e namedExpression) Map(f func(types.Expression) types.Expression) types.Expression {
	return f(e)
}

func (e namedExpression) ValueFrom(ctx types.Context, row shared.Row) (any, error) {
	index, err := shared.FindMatchingFieldIndex(e.field, row.Fields)
	if err != nil {
		return nil, err
	}

	return row.Values[index], nil
}
