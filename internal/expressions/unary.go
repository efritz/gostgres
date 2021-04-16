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

type unaryValueFromFunc func(expression Expression, row shared.Row) (interface{}, error)

func newUnaryExpression(expression Expression, operatorText string, valueFrom unaryValueFromFunc) Expression {
	return unaryExpression{
		expression:   expression,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e unaryExpression) String() string             { return fmt.Sprintf("%s %s", e.operatorText, e.expression) }
func (e unaryExpression) Fields() []shared.Field     { return e.expression.Fields() }
func (e unaryExpression) Conjunctions() []Expression { return []Expression{e} }

func (e unaryExpression) Fold() Expression {
	return tryEvaluate(newUnaryExpression(e.expression.Fold(), e.operatorText, e.valueFrom))
}

func (e unaryExpression) Alias(field shared.Field, expression Expression) Expression {
	return newUnaryExpression(e.expression.Alias(field, expression), e.operatorText, e.valueFrom)
}

func (e unaryExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.valueFrom(e.expression, row)
}
