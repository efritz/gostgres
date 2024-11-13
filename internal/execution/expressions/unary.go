package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
)

type unaryExpression struct {
	expression   impls.Expression
	operatorText string
	typeChecker  unaryTypeCheckerFunc
	typ          types.Type
	valueFrom    unaryValueFromFunc
}

var _ impls.Expression = &unaryExpression{}

type unaryTypeCheckerFunc func(expression types.Type) (types.Type, error)
type unaryValueFromFunc func(ctx impls.ExecutionContext, expression impls.Expression, row rows.Row) (any, error)

func newUnaryExpression(expression impls.Expression, operatorText string, typeChecker unaryTypeCheckerFunc, valueFrom unaryValueFromFunc) impls.Expression {
	return &unaryExpression{
		expression:   expression,
		operatorText: operatorText,
		typeChecker:  typeChecker,
		valueFrom:    valueFrom,
	}
}

func (e unaryExpression) String() string {
	return fmt.Sprintf("%s %s", e.operatorText, e.expression)
}

func (e *unaryExpression) Resolve(ctx impls.ExpressionResolutionContext) error {
	if err := e.expression.Resolve(ctx); err != nil {
		return err
	}

	typ, err := e.typeChecker(e.expression.Type())
	e.typ = typ
	return err
}

func (e unaryExpression) Type() types.Type {
	return e.typ
}

func (e unaryExpression) Equal(other impls.Expression) bool {
	if o, ok := other.(*unaryExpression); ok {
		return e.operatorText == o.operatorText && e.expression.Equal(o.expression)
	}

	return false
}

func (e unaryExpression) Children() []impls.Expression {
	return []impls.Expression{e.expression}
}

func (e unaryExpression) Fold() impls.Expression {
	return tryEvaluate(newUnaryExpression(e.expression.Fold(), e.operatorText, e.typeChecker, e.valueFrom))
}

func (e unaryExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	inner, err := e.expression.Map(f)
	if err != nil {
		return nil, err
	}

	return f(newUnaryExpression(inner, e.operatorText, e.typeChecker, e.valueFrom))
}

func (e unaryExpression) ValueFrom(ctx impls.ExecutionContext, row rows.Row) (any, error) {
	return e.valueFrom(ctx, e.expression, row)
}
