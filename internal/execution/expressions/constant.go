package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type constantExpression struct {
	value any
}

var _ impls.Expression = &constantExpression{}

func NewConstant(value any) impls.Expression {
	return &constantExpression{
		value: value,
	}
}

func (e *constantExpression) Resolve(ctx impls.ExpressionResolutionContext) error {
	return nil
}

func (e constantExpression) Type() types.Type {
	// TODO - should consider explicit `::type` casts
	return types.TypeKindFromValue(e.value)
}

func (e constantExpression) Equal(other impls.Expression) bool {
	if o, ok := other.(*constantExpression); ok {
		return e.value == o.value
	}

	return false
}

func (e constantExpression) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e constantExpression) Children() []impls.Expression {
	return nil
}

func (e constantExpression) Fold() impls.Expression {
	return &e
}

func (e constantExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	return f(&e)
}

func (e constantExpression) ValueFrom(ctx impls.ExecutionContext, row rows.Row) (any, error) {
	return e.value, nil
}

//
//

type ConstantPlaceholder interface {
	impls.Expression
	SetValue(value any)
}

type mutableConstant struct {
	constantExpression
}

func NewMutableConstant() ConstantPlaceholder {
	return &mutableConstant{
		constantExpression: constantExpression{},
	}
}

func (e *mutableConstant) SetValue(value any) {
	e.value = value
}
