package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type unaryExpression struct {
	expression   types.Expression
	operatorText string
	valueFrom    unaryValueFromFunc
}

type unaryValueFromFunc func(ctx types.Context, expression types.Expression, row shared.Row) (any, error)

func newUnaryExpression(expression types.Expression, operatorText string, valueFrom unaryValueFromFunc) types.Expression {
	return unaryExpression{
		expression:   expression,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e unaryExpression) String() string {
	return fmt.Sprintf("%s %s", e.operatorText, e.expression)
}

func (e unaryExpression) Equal(other types.Expression) bool {
	if o, ok := other.(unaryExpression); ok {
		return e.operatorText == o.operatorText && e.expression.Equal(o.expression)
	}

	return false
}

func (e unaryExpression) Children() []types.Expression {
	return []types.Expression{e.expression}
}

func (e unaryExpression) Fold() types.Expression {
	return tryEvaluate(newUnaryExpression(e.expression.Fold(), e.operatorText, e.valueFrom))
}

func (e unaryExpression) Map(f func(types.Expression) types.Expression) types.Expression {
	return f(newUnaryExpression(e.expression.Map(f), e.operatorText, e.valueFrom))
}

func (e unaryExpression) ValueFrom(ctx types.Context, row shared.Row) (any, error) {
	return e.valueFrom(ctx, e.expression, row)
}
