package expressions

import (
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type namedExpression struct {
	field fields.Field
}

var _ impls.Expression = &namedExpression{}

func NewNamed(field fields.Field) impls.Expression {
	return &namedExpression{
		field: field,
	}
}

func (e namedExpression) String() string {
	return e.field.String()
}

func (e *namedExpression) Resolve(ctx impls.ExpressionResolutionContext) error {
	return nil
}

func (e namedExpression) Type() types.Type {
	return e.field.Type()
}

func (e namedExpression) Equal(other impls.Expression) bool {
	if o, ok := other.(NamedExpression); ok {
		return e.field.Name() == o.Field().Name() && (e.field.RelationName() == o.Field().RelationName() || e.field.RelationName() == "" || o.Field().RelationName() == "")
	}

	return false
}

func (e namedExpression) Name() string {
	return e.field.Name()
}

func (e namedExpression) Field() fields.Field {
	return e.field
}

func (e namedExpression) Children() []impls.Expression {
	return nil
}

func (e namedExpression) Fold() impls.Expression {
	return &e
}

func (e namedExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	return f(&e)
}

func (e namedExpression) ValueFrom(ctx impls.ExecutionContext, row rows.Row) (any, error) {
	index, err := fields.FindMatchingFieldIndex(e.field, row.Fields)
	if err != nil {
		return nil, err
	}

	return row.Values[index], nil
}
