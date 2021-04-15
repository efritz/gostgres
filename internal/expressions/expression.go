package expressions

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type Expression interface {
	Fields() []shared.Field
	Fold() Expression
	Alias(field shared.Field, expression Expression) Expression
	Conjunctions() []Expression
	ValueFrom(row shared.Row) (interface{}, error)
}

type unaryExpression struct {
	expression   Expression
	operatorText string
	valueFrom    unaryValueFromFunc
}

type unaryValueFromFunc func(expression Expression, row shared.Row) (interface{}, error)

func newUnaryExpression(expression Expression, operatorText string, valueFrom unaryValueFromFunc) unaryExpression {
	return unaryExpression{
		expression:   expression,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e unaryExpression) String() string {
	return fmt.Sprintf("%s %s", e.operatorText, e.expression)
}

func (e unaryExpression) Fields() []shared.Field {
	return e.expression.Fields()
}

func (e unaryExpression) Fold() Expression {
	e = newUnaryExpression(e.expression.Fold(), e.operatorText, e.valueFrom)

	value, err := e.valueFrom(e.expression, shared.Row{})
	if err == nil {
		return NewConstant(value)
	}

	return e
}

func (e unaryExpression) Alias(field shared.Field, expression Expression) Expression {
	return newUnaryExpression(e.expression.Alias(field, expression), e.operatorText, e.valueFrom)
}

func (e unaryExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e unaryExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.valueFrom(e.expression, row)
}

type binaryExpression struct {
	left         Expression
	right        Expression
	operatorText string
	valueFrom    binaryValueFromFunc
}

type binaryValueFromFunc func(left, right Expression, row shared.Row) (interface{}, error)

func newBinaryExpression(left, right Expression, operatorText string, valueFrom binaryValueFromFunc) binaryExpression {
	return binaryExpression{
		left:         left,
		right:        right,
		operatorText: operatorText,
		valueFrom:    valueFrom,
	}
}

func (e binaryExpression) String() string {
	return fmt.Sprintf("%s %s %s", e.left, e.operatorText, e.right)
}

func (e binaryExpression) Fields() []shared.Field {
	return append(e.left.Fields(), e.right.Fields()...)
}

func (e binaryExpression) Fold() Expression {
	e = newBinaryExpression(e.left.Fold(), e.right.Fold(), e.operatorText, e.valueFrom)

	value, err := e.valueFrom(e.left, e.right, shared.Row{})
	if err == nil {
		return NewConstant(value)
	}

	return e
}

func (e binaryExpression) Alias(field shared.Field, expression Expression) Expression {
	return newBinaryExpression(e.left.Alias(field, expression), e.right.Alias(field, expression), e.operatorText, e.valueFrom)
}

func (e binaryExpression) Conjunctions() []Expression {
	return []Expression{e}
}

func (e binaryExpression) ValueFrom(row shared.Row) (interface{}, error) {
	return e.valueFrom(e.left, e.right, row)
}
