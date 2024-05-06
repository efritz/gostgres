package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type unaryExpression struct {
	expression   Expression
	operatorText string
	valueFrom    unaryValueFromFunc
}

type unaryValueFromFunc func(context Context, expression Expression, row shared.Row) (any, error)

func newUnaryExpression(expression Expression, operatorText string, valueFrom unaryValueFromFunc) Expression {
	return unaryExpression{
		expression:   expression,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e unaryExpression) String() string {
	return fmt.Sprintf("%s %s", e.operatorText, e.expression)
}

func (e unaryExpression) Equal(other Expression) bool {
	if o, ok := other.(unaryExpression); ok {
		return e.operatorText == o.operatorText && e.expression.Equal(o.expression)
	}

	return false
}

func (e unaryExpression) Fields() []shared.Field {
	return e.expression.Fields()
}

func (e unaryExpression) Named() (shared.Field, bool) {
	return shared.Field{}, false
}

func (e unaryExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e unaryExpression) Fold() Expression {
	return tryEvaluate(newUnaryExpression(e.expression.Fold(), e.operatorText, e.valueFrom))
}

func (e unaryExpression) Map(f func(Expression) Expression) Expression {
	return f(newUnaryExpression(e.expression.Map(f), e.operatorText, e.valueFrom))
}

func (e unaryExpression) ValueFrom(context Context, row shared.Row) (any, error) {
	return e.valueFrom(context, e.expression, row)
}
