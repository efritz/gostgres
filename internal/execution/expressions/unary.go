package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type unaryExpression struct {
	expression   impls.Expression
	operatorText string
	valueFrom    unaryValueFromFunc
}

type unaryValueFromFunc func(ctx impls.Context, expression impls.Expression, row rows.Row) (any, error)

func newUnaryExpression(expression impls.Expression, operatorText string, valueFrom unaryValueFromFunc) impls.Expression {
	return unaryExpression{
		expression:   expression,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e unaryExpression) String() string {
	return fmt.Sprintf("%s %s", e.operatorText, e.expression)
}

func (e unaryExpression) Equal(other impls.Expression) bool {
	if o, ok := other.(unaryExpression); ok {
		return e.operatorText == o.operatorText && e.expression.Equal(o.expression)
	}

	return false
}

func (e unaryExpression) Children() []impls.Expression {
	return []impls.Expression{e.expression}
}

func (e unaryExpression) Fold() impls.Expression {
	return tryEvaluate(newUnaryExpression(e.expression.Fold(), e.operatorText, e.valueFrom))
}

func (e unaryExpression) Map(f func(impls.Expression) (impls.Expression, error)) (impls.Expression, error) {
	inner, err := e.expression.Map(f)
	if err != nil {
		return nil, err
	}

	return f(newUnaryExpression(inner, e.operatorText, e.valueFrom))
}

func (e unaryExpression) ValueFrom(ctx impls.Context, row rows.Row) (any, error) {
	return e.valueFrom(ctx, e.expression, row)
}
